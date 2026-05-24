package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-collector/collector/dns/dao"
	"github.com/gofurry/gofurry-nav-collector/collector/dns/models"
	"github.com/gofurry/gofurry-nav-collector/collector/observation"
	runstate "github.com/gofurry/gofurry-nav-collector/collector/scheduler"
	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/log"
	cs "github.com/gofurry/gofurry-nav-collector/common/service"
	"github.com/gofurry/gofurry-nav-collector/common/util"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
	"github.com/miekg/dns"
	"github.com/oschwald/geoip2-golang"
	"github.com/sourcegraph/conc/pool"
)

const (
	// 最大递归查询深度，防止 CNAME/MX/NS 无限递归
	MaxDepth = 2
	// PTR 查询并发数控制
	PTRWorkers = 5

	maxDNSObservationTextLength = 512

	dnsRiskPrivateIP          = "private_ip"
	dnsRiskLowTTL             = "low_ttl"
	dnsRiskNXDomainWithAnswer = "nxdomain_with_answer"
	dnsRiskPTREmpty           = "ptr_empty"
)

// 按域名并行查 加锁
var dnsRunning atomic.Bool

// 缓存 IP
var geoCache sync.Map                        // geoCache 缓存 IP 的 GeoIP/ASN 查询结果
var ptrCache sync.Map                        // ptrCache 缓存 IP 的反向 PTR 查询结果
var ptrSem = make(chan struct{}, PTRWorkers) // ptrSem 用于限制 PTR 查询并发

var resolver = env.GetServerConfig().Collector.Dns.Resolver
var geoDBs *GeoDBSet

type dnsResponseSummary struct {
	Rcode              string  `json:"rcode"`
	Authoritative      bool    `json:"authoritative"`
	Truncated          bool    `json:"truncated"`
	RecursionAvailable bool    `json:"recursion_available"`
	AnswerCount        int     `json:"answer_count"`
	AuthorityCount     int     `json:"authority_count"`
	AdditionalCount    int     `json:"additional_count"`
	TTLMin             uint32  `json:"ttl_min"`
	TTLMax             uint32  `json:"ttl_max"`
	TTLAvg             float64 `json:"ttl_avg"`
	DNSSECRRSIGPresent bool    `json:"dnssec_rrsig_present"`
	DNSSECAD           bool    `json:"dnssec_ad"`
}

type dnsMXPriority struct {
	Host     string `json:"host"`
	Priority uint16 `json:"priority"`
}

type dnsSOASummary struct {
	NS      string `json:"ns"`
	Mbox    string `json:"mbox"`
	Serial  uint32 `json:"serial"`
	Refresh uint32 `json:"refresh"`
	Retry   uint32 `json:"retry"`
	Expire  uint32 `json:"expire"`
	Minttl  uint32 `json:"minttl"`
}

type dnsQueryMetadata struct {
	ResponseSummary dnsResponseSummary
	MXPriorities    []dnsMXPriority
	SOA             *dnsSOASummary
}

type GeoDBSet struct {
	Country *geoip2.Reader
	City    *geoip2.Reader
	ASN     *geoip2.Reader
}

func InitGeoDB(dbPath string) *GeoDBSet {
	return &GeoDBSet{
		Country: openGeoDB(dbPath, "GeoLite2-Country.mmdb", "Country"),
		City:    openGeoDB(dbPath, "GeoLite2-City.mmdb", "City"),
		ASN:     openGeoDB(dbPath, "GeoLite2-ASN.mmdb", "ASN"),
	}
}

func openGeoDB(dbPath string, fileName string, label string) *geoip2.Reader {
	reader, err := geoip2.Open(filepath.Join(dbPath, fileName))
	if err != nil {
		log.WarnFields(map[string]interface{}{
			"db":       label,
			"event":    "geoip_open_failed",
			"path":     filepath.Join(dbPath, fileName),
			"protocol": "dns",
		}, "GeoIP 数据库不可用，DNS 地理字段将降级为 Unknown: "+err.Error())
		return nil
	}
	log.InfoFields(map[string]interface{}{
		"db":       label,
		"event":    "geoip_opened",
		"path":     filepath.Join(dbPath, fileName),
		"protocol": "dns",
	}, "GeoIP 数据库已打开")
	return reader
}

func CloseGeoDB() {
	if geoDBs == nil {
		log.InfoFields(map[string]interface{}{
			"event":    "geoip_close_skipped",
			"protocol": "dns",
			"reason":   "not_open",
		}, "GeoIP 数据库未打开，跳过关闭")
		return
	}
	geoDBs.Close()
	geoDBs = nil
	log.InfoFields(map[string]interface{}{
		"event":    "geoip_closed",
		"protocol": "dns",
	}, "GeoIP 数据库已关闭")
}

func (g *GeoDBSet) Close() {
	if g == nil {
		return
	}
	if g.Country != nil {
		_ = g.Country.Close()
	}
	if g.City != nil {
		_ = g.City.Close()
	}
	if g.ASN != nil {
		_ = g.ASN.Close()
	}
}

func currentGeoDBs() *GeoDBSet {
	if geoDBs != nil {
		return geoDBs
	}
	return &GeoDBSet{}
}

// ============== DNS解析 - 初始化部分 ==============

// 初始化
func InitDNSOnStart() {
	defer func() {
		if err := recover(); err != nil {
			log.ErrorFields(map[string]interface{}{
				"event":    "init_recovered",
				"protocol": "dns",
			}, err)
		}
	}()
	log.InfoFields(map[string]interface{}{
		"event":           "module_init_start",
		"interval":        time.Duration(env.GetServerConfig().Collector.Dns.DnsInterval) * time.Hour,
		"protocol":        "dns",
		"query_workers":   dnsQueryThreadLimit(env.GetServerConfig().Collector.Dns.QueryThread, len(models.RecordTypes)),
		"resolver":        resolver,
		"retention_every": time.Hour * 72,
		"workers":         env.GetServerConfig().Collector.Dns.DnsThread,
	}, "DNS 采集模块初始化开始")
	geoDBs = InitGeoDB(env.GetServerConfig().Collector.Dns.Geolite2Path)

	//初始化后执行一次 ParseDNS
	go ParseDNS()
	go Delete()
	// 定时任务执行 ParseDNS
	cs.AddCronJob(time.Duration(env.GetServerConfig().Collector.Dns.DnsInterval)*time.Hour, ParseDNS)
	cs.AddCronJob(72*time.Hour, Delete)

	log.InfoFields(map[string]interface{}{
		"event":    "module_init_complete",
		"protocol": "dns",
	}, "DNS 采集模块初始化完成")
}

// 每天清理一次日志表
func Delete() {
	defer func() {
		if err := recover(); err != nil {
			log.ErrorFields(map[string]interface{}{
				"event":    "retention_recovered",
				"protocol": "dns",
			}, err)
		}
	}()

	start := time.Now()
	keepCount := env.GetServerConfig().Collector.Dns.LogCount
	log.InfoFields(map[string]interface{}{
		"event":      "retention_start",
		"keep_count": keepCount,
		"protocol":   "dns",
	}, "DNS 历史日志保留清理开始")

	// 每个域名仅保留 500 条 DNS 记录
	count, deleteErr := dao.GetDNSDao().DeleteByNum(keepCount)
	if deleteErr != nil {
		log.ErrorFields(map[string]interface{}{
			"deleted":    count,
			"duration":   time.Since(start),
			"event":      "retention_failed",
			"keep_count": keepCount,
			"protocol":   "dns",
		}, "DNS 历史日志保留清理失败: "+deleteErr.GetMsg())
	} else {
		log.InfoFields(map[string]interface{}{
			"deleted":    count,
			"duration":   time.Since(start),
			"event":      "retention_complete",
			"keep_count": keepCount,
			"protocol":   "dns",
		}, "DNS 历史日志保留清理完成")
	}
	if env.GetServerConfig().Collector.V2.ProtocolEnabled(observation.ProtocolDNS) {
		v2Count, v2DeleteErr := observation.DeleteByProtocolLimit(observation.ProtocolDNS, keepCount)
		if v2DeleteErr != nil {
			log.ErrorFields(map[string]interface{}{
				"deleted":    v2Count,
				"duration":   time.Since(start),
				"event":      "v2_retention_failed",
				"keep_count": keepCount,
				"protocol":   "dns",
			}, "DNS v2 observation 保留清理失败: "+v2DeleteErr.GetMsg())
		} else if v2Count > 0 {
			log.InfoFields(map[string]interface{}{
				"deleted":    v2Count,
				"duration":   time.Since(start),
				"event":      "v2_retention_complete",
				"keep_count": keepCount,
				"protocol":   "dns",
			}, "DNS v2 observation 保留清理完成")
		}
	}
}

// ============== DNS解析 - 执行部分 ==============

// 执行 ParseDNS
func ParseDNS() {
	interval := time.Duration(env.GetServerConfig().Collector.Dns.DnsInterval) * time.Hour
	run := runstate.NewRun(observation.ProtocolDNS, interval)
	defer func() {
		if err := recover(); err != nil {
			run.Fail("panic", int(run.Snapshot(runstate.StatusFailed, "").TargetCount))
			fields := run.Fields()
			fields["event"] = "run_recovered"
			log.ErrorFields(fields, fmt.Sprintf("DNS 采集运行触发 panic，已恢复: %v", err))
		}
	}()
	if !dnsRunning.CompareAndSwap(false, true) {
		run.Skip("previous_run_running", 0)
		fields := run.Fields()
		fields["event"] = "run_skipped"
		fields["reason"] = "上一轮采集仍在运行"
		fields["skipped_reason"] = "previous_run_running"
		fields["status"] = "skipped"
		log.WarnFields(fields, "DNS 采集已跳过：上一轮仍在运行")
		return
	}
	defer dnsRunning.Store(false)
	if !run.AcquireLeaseOrSkip() {
		fields := run.Fields()
		fields["event"] = "run_skipped"
		fields["reason"] = "采集 lease 已被其他实例持有"
		fields["skipped_reason"] = "lease_held_by_other_collector"
		fields["status"] = "skipped"
		log.WarnFields(fields, "DNS 采集已跳过：采集 lease 已被其他实例持有")
		return
	}
	defer run.ReleaseLease()
	run.Start()

	start := time.Now()
	fields := run.Fields()
	fields["event"] = "run_start"
	fields["max_depth"] = MaxDepth
	fields["resolver"] = resolver
	fields["timeout"] = env.GetServerConfig().Collector.ProbeBudget.DNSTimeout()
	fields["ptr_timeout"] = env.GetServerConfig().Collector.ProbeBudget.PTRTimeout()
	fields["query_workers"] = dnsQueryThreadLimit(env.GetServerConfig().Collector.Dns.QueryThread, len(models.RecordTypes))
	fields["workers"] = env.GetServerConfig().Collector.Dns.DnsThread
	log.InfoFields(fields, "DNS 采集运行开始")

	requestList, err := dao.GetDNSDao().GetList()
	if err != nil {
		run.Fail("load_targets", 0)
		fields := run.Fields()
		fields["duration"] = time.Since(start)
		fields["event"] = "run_failed"
		fields["stage"] = "load_targets"
		log.ErrorFields(fields, "DNS 目标列表读取失败: "+err.GetMsg())
		return
	}
	// 判空
	if len(requestList) < 1 {
		run.Complete(0)
		fields := run.Fields()
		fields["duration"] = time.Since(start)
		fields["event"] = "run_complete"
		fields["reason"] = "目标列表为空"
		fields["targets"] = 0
		log.InfoFields(fields, "DNS 采集完成：没有需要探测的目标")
		return
	}
	run.SetTargetCount(len(requestList))

	log.InfoFields(map[string]interface{}{
		"event":         "probe_start",
		"protocol":      "dns",
		"query_workers": dnsQueryThreadLimit(env.GetServerConfig().Collector.Dns.QueryThread, len(models.RecordTypes)),
		"targets":       len(requestList),
		"workers":       env.GetServerConfig().Collector.Dns.DnsThread,
	}, "DNS 探测开始")
	dnsThread := pool.New().WithMaxGoroutines(env.GetServerConfig().Collector.Dns.DnsThread)
	// 遍历站点列表, 每个站点开一个线程执行采集
	for _, v := range requestList {
		dnsThread.Go(getDNSResult(v, run))
	}
	// 等待所有采集和解析执行完毕
	dnsThread.Wait()
	run.Complete(len(requestList))
	snapshot := run.Snapshot(runstate.StatusComplete, "")
	fields = run.Fields()
	fields["duration"] = time.Since(start)
	fields["event"] = "run_complete"
	fields["failure_count"] = snapshot.FailureCount
	fields["success_count"] = snapshot.SuccessCount
	fields["targets"] = len(requestList)
	log.InfoFields(fields, "DNS 采集运行完成")
}

func getDNSResult(site models.GfnCollectorDomain, run *runstate.Run) func() {
	return func() {
		defer func() {
			if err := recover(); err != nil {
				log.ErrorFields(map[string]interface{}{
					"event":    "probe_recovered",
					"protocol": "dns",
					"site":     site.TargetName(),
				}, fmt.Sprintf("DNS 单目标探测触发 panic，已恢复: %v", err))
			}
		}()

		// 执行 Request 获取结果
		probeStart := time.Now()
		geoDBSet := currentGeoDBs()
		results, metadata := performDNSQuery(site, geoDBSet.ASN, geoDBSet.City, geoDBSet.Country)

		siteName := site.TargetName()

		// Ping 结果储存回 redis
		resultKey := "dns:" + siteName
		resultMap := make(map[string]string)
		for k, result := range results {
			jsonResult, _ := sonic.Marshal(result)
			resultMap[k] = string(jsonResult)
		}
		gfError := cs.HSetMap(resultKey, resultMap)
		if gfError != nil {
			log.ErrorFields(map[string]interface{}{
				"event":     "redis_write_failed",
				"protocol":  "dns",
				"recordset": len(resultMap),
				"redis_key": resultKey,
				"site":      siteName,
			}, "DNS 探测结果写入 Redis 失败: "+gfError.GetMsg())
		}

		newRecord := models.GfnCollectorLogDn{
			ID:         util.GenerateId(),
			Name:       siteName,
			CreateTime: time.Now(),
		}
		for k, v := range results {
			marshal, jsonErr := sonic.Marshal(v)
			if jsonErr != nil {
				log.ErrorFields(map[string]interface{}{
					"event":       "result_encode_failed",
					"protocol":    "dns",
					"record_type": k,
					"site":        siteName,
				}, "DNS 探测结果 JSON 编码失败: "+jsonErr.Error())
			}
			jsonRecord := string(marshal)
			if &jsonRecord != nil {
				newRecord.Status = "success"
			}
			switch k {
			case "A":
				newRecord.A = &jsonRecord
			case "AAAA":
				newRecord.Aaaa = &jsonRecord
			case "CNAME":
				newRecord.Cname = &jsonRecord
			case "MX":
				newRecord.Mx = &jsonRecord
			case "NS":
				newRecord.Ns = &jsonRecord
			case "TXT":
				newRecord.Txt = &jsonRecord
			case "SOA":
				newRecord.Soa = &jsonRecord
			case "CAA":
				newRecord.Caa = &jsonRecord
			default:
			}
		}
		if newRecord.Status != "success" {
			newRecord.Status = "failure"
		}
		if run != nil {
			if newRecord.Status == "success" {
				run.RecordSuccess()
			} else {
				run.RecordFailure()
			}
		}

		// 存数据库
		daoErr := dao.GetDNSDao().Add(&newRecord)
		if daoErr != nil {
			log.ErrorFields(map[string]interface{}{
				"event":    "db_write_failed",
				"protocol": "dns",
				"site":     siteName,
				"status":   newRecord.Status,
			}, "DNS 探测结果写入数据库失败: "+daoErr.GetMsg())
		}
		errorCode := ""
		if newRecord.Status != "success" {
			errorCode = "dns_no_records"
		}
		collectorID, jobID := "", ""
		if run != nil {
			collectorID = run.CollectorID
			jobID = run.JobID
		}
		saveErr := observation.SaveIfEnabled(observation.Input{
			SiteID:      site.SiteID,
			Target:      siteName,
			Protocol:    observation.ProtocolDNS,
			Status:      observationStatusFromDNS(newRecord.Status),
			ObservedAt:  newRecord.CreateTime,
			DurationMS:  time.Since(probeStart).Milliseconds(),
			ErrorCode:   errorCode,
			Payload:     buildDNSObservationPayloadWithMetadata(results, metadata),
			CollectorID: collectorID,
			JobID:       jobID,
		})
		if saveErr != nil {
			log.WarnFields(map[string]interface{}{
				"event":    "v2_observation_write_failed",
				"protocol": "dns",
				"site_id":  site.SiteID,
				"site":     siteName,
			}, "DNS v2 observation 旁路写入失败: "+saveErr.GetMsg())
		}

	}
}

func observationStatusFromDNS(status string) string {
	if status == "success" {
		return observation.StatusSuccess
	}
	return observation.StatusFailure
}

func buildDNSObservationPayload(results map[string][]models.DNSRecord) map[string]any {
	return buildDNSObservationPayloadWithMetadata(results, nil)
}

func buildDNSObservationPayloadWithMetadata(results map[string][]models.DNSRecord, metadata map[string]dnsQueryMetadata) map[string]any {
	payload := make(map[string]any, len(results)+1)
	aggregateRisks := map[string]struct{}{}
	responseSummary := make(map[string]dnsResponseSummary, len(results))

	for recordType, records := range results {
		v2Records := make([]map[string]any, 0, len(records))
		for _, record := range records {
			v2Record, risks := buildDNSObservationRecord(record)
			v2Records = append(v2Records, v2Record)
			for _, risk := range risks {
				aggregateRisks[risk] = struct{}{}
			}
		}
		payload[recordType] = v2Records
		responseSummary[recordType] = buildDNSResponseSummary(recordType, records, metadata)
	}
	payload["risk_flags"] = sortedRiskFlags(aggregateRisks)
	payload["response_summary"] = responseSummary
	payload["cname_chain_depth"] = maxDNSRecordDepth(results)
	payload["mx_priorities"] = buildDNSMXPriorities(metadata)
	payload["soa"] = buildDNSSOA(metadata)
	return payload
}

func buildDNSResponseSummary(recordType string, records []models.DNSRecord, metadata map[string]dnsQueryMetadata) dnsResponseSummary {
	if metadata != nil {
		if item, ok := metadata[recordType]; ok {
			return item.ResponseSummary
		}
	}

	summary := dnsResponseSummary{
		AnswerCount: len(records),
	}
	if len(records) == 0 {
		return summary
	}

	var ttlSum uint64
	summary.TTLMin = records[0].TTL
	summary.TTLMax = records[0].TTL
	for _, record := range records {
		ttlSum += uint64(record.TTL)
		if record.TTL < summary.TTLMin {
			summary.TTLMin = record.TTL
		}
		if record.TTL > summary.TTLMax {
			summary.TTLMax = record.TTL
		}
		if record.DNSSEC {
			summary.DNSSECRRSIGPresent = true
		}
	}
	summary.TTLAvg = float64(ttlSum) / float64(len(records))
	return summary
}

func buildDNSObservationRecord(record models.DNSRecord) (map[string]any, []string) {
	value, truncated := limitDNSObservationText(record.Value)
	reversePTR, _ := limitDNSObservationText(record.ReversePTR)
	riskFlags := sortedStringSlice(record.RiskFlags)
	riskSet := make(map[string]struct{}, len(riskFlags))
	for _, risk := range riskFlags {
		riskSet[risk] = struct{}{}
	}

	children := make([]map[string]any, 0, len(record.Children))
	for _, child := range record.Children {
		childRecord, childRisks := buildDNSObservationRecord(child)
		children = append(children, childRecord)
		for _, risk := range childRisks {
			riskSet[risk] = struct{}{}
		}
	}

	v2Record := map[string]any{
		"type":          record.Type,
		"value":         value,
		"ttl":           record.TTL,
		"dnssec":        record.DNSSEC,
		"asn":           record.ASN,
		"country":       record.Country,
		"city":          record.City,
		"provider_type": record.ProviderType,
		"isp":           record.ISP,
		"duration_ms":   record.Duration.Milliseconds(),
		"children":      children,
		"reverse_ptr":   reversePTR,
		"risk_flags":    sortedRiskFlags(riskSet),
	}
	if truncated {
		v2Record["value_truncated"] = true
		v2Record["value_original_length"] = len([]rune(record.Value))
		v2Record["value_sha256"] = sha256Hex(record.Value)
	}
	if textKind := dnsTextKind(record); textKind != "" {
		v2Record["text_kind"] = textKind
	}
	return v2Record, sortedRiskFlags(riskSet)
}

func limitDNSObservationText(value string) (string, bool) {
	if len([]rune(value)) <= maxDNSObservationTextLength {
		return value, false
	}
	return string([]rune(value)[:maxDNSObservationTextLength]), true
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func dnsTextKind(record models.DNSRecord) string {
	switch record.Type {
	case "CAA":
		return "caa"
	case "TXT":
		normalized := strings.ToLower(strings.TrimSpace(record.Value))
		if strings.HasPrefix(normalized, "v=spf1") {
			return "spf"
		}
		if strings.HasPrefix(normalized, "v=dmarc1") {
			return "dmarc"
		}
		return "txt"
	default:
		return ""
	}
}

func sortedRiskFlags(riskSet map[string]struct{}) []string {
	risks := make([]string, 0, len(riskSet))
	for risk := range riskSet {
		risks = append(risks, risk)
	}
	sort.Strings(risks)
	return risks
}

func sortedStringSlice(values []string) []string {
	if values == nil {
		return nil
	}
	copied := append([]string(nil), values...)
	sort.Strings(copied)
	return copied
}

func performDNSQuery(site models.GfnCollectorDomain, asnDB *geoip2.Reader, cityDB *geoip2.Reader, countryDB *geoip2.Reader) (map[string][]models.DNSRecord, map[string]dnsQueryMetadata) {
	// 按记录类型并行查 加锁
	var queryMu sync.Mutex
	var queryMG sync.WaitGroup
	// 全局统计
	var globalMinTTL uint32 = 1<<32 - 1
	var globalMaxTTL uint32
	var globalTTLsum uint64
	var globalDurations []time.Duration
	var globalTotalTime time.Duration
	// 最终结果
	result := make(map[string][]models.DNSRecord)
	metadata := make(map[string]dnsQueryMetadata)

	domain := site.TargetName()

	// 并行查询每种记录类型
	queryWorkers := dnsQueryThreadLimit(env.GetServerConfig().Collector.Dns.QueryThread, len(models.RecordTypes))
	querySem := make(chan struct{}, queryWorkers)
	for _, rt := range models.RecordTypes {
		queryMG.Add(1)
		go func(rt models.RecordType) {
			defer queryMG.Done()
			querySem <- struct{}{}
			defer func() { <-querySem }()

			records, stats, queryMetadata, err := queryDNS(domain, rt.Type, resolver, countryDB, cityDB, asnDB, 0)
			if err != nil {
				log.WarnFields(map[string]interface{}{
					"domain":      domain,
					"event":       "query_failed",
					"protocol":    "dns",
					"record_type": rt.Name,
					"resolver":    resolver,
				}, "DNS 查询失败: "+err.GetMsg())
				return
			}

			// 保存结果 & 更新全局统计
			queryMu.Lock()
			for _, record := range records {
				result[record.Type] = append(result[record.Type], record)
			}
			metadata[rt.Name] = queryMetadata

			if stats.MinTTL < globalMinTTL {
				globalMinTTL = stats.MinTTL
			}
			if stats.MaxTTL > globalMaxTTL {
				globalMaxTTL = stats.MaxTTL
			}
			globalTTLsum += uint64(stats.AvgTTL * float64(len(records)))
			globalDurations = append(globalDurations, stats.MinTime, stats.MaxTime)
			globalTotalTime += stats.TotalTime
			queryMu.Unlock()
		}(rt)

	}
	queryMG.Wait()
	return result, metadata
}

// ============== DNS解析 - 采集和解析部分 ==============

func dnsQueryThreadLimit(configured int, recordTypeCount int) int {
	if recordTypeCount <= 0 {
		return 1
	}
	if configured <= 0 || configured > recordTypeCount {
		return recordTypeCount
	}
	return configured
}

func queryDNS(domain string, qtype uint16, resolver string, countryDB *geoip2.Reader, cityDB *geoip2.Reader, asnDB *geoip2.Reader, depth int) ([]models.DNSRecord, models.DNSStatistics, dnsQueryMetadata, common.GFError) {
	// 防止递归过深
	if depth > MaxDepth {
		return nil, models.DNSStatistics{}, dnsQueryMetadata{}, nil
	}

	start := time.Now()

	probeBudget := env.GetServerConfig().Collector.ProbeBudget

	// UDP DNS 查询客户端
	c := &dns.Client{Net: "udp", Timeout: probeBudget.DNSTimeout()}
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), qtype)
	m.SetEdns0(4096, true) // 支持 DNSSEC

	// 执行 DNS 查询
	in, _, err := c.Exchange(m, resolver)
	if err != nil {
		return nil, models.DNSStatistics{}, dnsQueryMetadata{}, common.NewServiceError("DNS 查询失败: " + err.Error())
	}
	totalTime := time.Since(start)

	// 检查是否有 DNSSEC
	dnssec := false
	for _, rr := range in.Answer {
		if rr.Header().Rrtype == dns.TypeRRSIG {
			dnssec = true
			break
		}
	}

	var results []models.DNSRecord
	var ttlSum uint32
	minTTL, maxTTL := uint32(1<<32-1), uint32(0)
	var durations []time.Duration
	metadata := dnsQueryMetadata{
		ResponseSummary: dnsResponseSummary{
			Rcode:              dnsRcodeName(in.Rcode),
			Authoritative:      in.Authoritative,
			Truncated:          in.Truncated,
			RecursionAvailable: in.RecursionAvailable,
			AnswerCount:        len(in.Answer),
			AuthorityCount:     len(in.Ns),
			AdditionalCount:    len(in.Extra),
			DNSSECRRSIGPresent: dnssec,
			DNSSECAD:           in.AuthenticatedData,
		},
	}

	// 遍历每条 Answer 记录
	maxRecords := probeBudget.MaxDNSRecords()
	for _, rr := range in.Answer {
		if len(results) >= maxRecords {
			break
		}
		recStart := time.Now()
		rec := models.DNSRecord{
			Type:   dns.TypeToString[rr.Header().Rrtype],
			TTL:    rr.Header().Ttl,
			DNSSEC: dnssec,
		}
		ttlSum += rr.Header().Ttl
		if rr.Header().Ttl < minTTL {
			minTTL = rr.Header().Ttl
		}
		if rr.Header().Ttl > maxTTL {
			maxTTL = rr.Header().Ttl
		}

		// 根据记录类型处理
		switch v := rr.(type) {
		case *dns.A:
			rec.Value = v.A.String()
			rec.Country, rec.City, rec.ASN, rec.ISP = lookupGeoASN(v.A, countryDB, cityDB, asnDB)
			rec.ProviderType = detectCDN(rec.ASN, v.A, domain)
			rec.ReversePTR = reversePTR(v.A)
			rec.RiskFlags = detectDNSRiskFlags(v.A, in, v.Hdr.Ttl, rec.ReversePTR)
			rec.Hijacked = detectHijack(v.A, in, v.Hdr.Ttl)
		case *dns.AAAA:
			rec.Value = v.AAAA.String()
			rec.Country, rec.City, rec.ASN, rec.ISP = lookupGeoASN(v.AAAA, countryDB, cityDB, asnDB)
			rec.ProviderType = detectCDN(rec.ASN, v.AAAA, domain)
			rec.ReversePTR = reversePTR(v.AAAA)
			rec.RiskFlags = detectDNSRiskFlags(v.AAAA, in, v.Hdr.Ttl, rec.ReversePTR)
			rec.Hijacked = detectHijack(v.AAAA, in, v.Hdr.Ttl)
		case *dns.CNAME:
			rec.Value = v.Target
			// 递归查询 CNAME 指向的 A/AAAA
			childrenA, _, _, _ := queryDNS(v.Target, dns.TypeA, resolver, countryDB, cityDB, asnDB, depth+1)
			childrenAAAA, _, _, _ := queryDNS(v.Target, dns.TypeAAAA, resolver, countryDB, cityDB, asnDB, depth+1)
			rec.Children = append(rec.Children, childrenA...)
			rec.Children = append(rec.Children, childrenAAAA...)
		case *dns.MX:
			rec.Value = fmt.Sprintf("%s (优先级 %d)", v.Mx, v.Preference)
			metadata.MXPriorities = append(metadata.MXPriorities, dnsMXPriority{
				Host:     limitDNSObservationString(v.Mx),
				Priority: v.Preference,
			})
			childrenA, _, _, _ := queryDNS(v.Mx, dns.TypeA, resolver, countryDB, cityDB, asnDB, depth+1)
			childrenAAAA, _, _, _ := queryDNS(v.Mx, dns.TypeAAAA, resolver, countryDB, cityDB, asnDB, depth+1)
			rec.Children = append(rec.Children, childrenA...)
			rec.Children = append(rec.Children, childrenAAAA...)
		case *dns.NS:
			rec.Value = v.Ns
			childrenA, _, _, _ := queryDNS(v.Ns, dns.TypeA, resolver, countryDB, cityDB, asnDB, depth+1)
			childrenAAAA, _, _, _ := queryDNS(v.Ns, dns.TypeAAAA, resolver, countryDB, cityDB, asnDB, depth+1)
			rec.Children = append(rec.Children, childrenA...)
			rec.Children = append(rec.Children, childrenAAAA...)
		case *dns.TXT:
			rec.Value = strings.Join(v.Txt, " ")
		case *dns.SOA:
			rec.Value = fmt.Sprintf("%s %s", v.Ns, v.Mbox)
			metadata.SOA = &dnsSOASummary{
				NS:      limitDNSObservationString(v.Ns),
				Mbox:    limitDNSObservationString(v.Mbox),
				Serial:  v.Serial,
				Refresh: v.Refresh,
				Retry:   v.Retry,
				Expire:  v.Expire,
				Minttl:  v.Minttl,
			}
		case *dns.CAA:
			rec.Value = fmt.Sprintf("%d %s %s", v.Flag, v.Tag, v.Value)
		default:
			rec.Value = rr.String()
		}

		rec.Duration = time.Since(recStart)
		durations = append(durations, rec.Duration)
		results = append(results, rec)
	}

	// 统计 TTL / 耗时信息
	stats := models.DNSStatistics{
		MinTTL:    minTTL,
		MaxTTL:    maxTTL,
		TotalTime: totalTime,
	}
	if len(results) > 0 {
		stats.AvgTTL = float64(ttlSum) / float64(len(results))
		metadata.ResponseSummary.TTLMin = minTTL
		metadata.ResponseSummary.TTLMax = maxTTL
		metadata.ResponseSummary.TTLAvg = stats.AvgTTL
		stats.MinTime, stats.MaxTime = durations[0], durations[0]
		var total time.Duration
		for _, d := range durations {
			total += d
			if d < stats.MinTime {
				stats.MinTime = d
			}
			if d > stats.MaxTime {
				stats.MaxTime = d
			}
		}
		stats.AvgTime = total / time.Duration(len(durations))
	}

	return results, stats, metadata, nil
}

func dnsRcodeName(rcode int) string {
	if name, ok := dns.RcodeToString[rcode]; ok {
		return name
	}
	return fmt.Sprintf("RCODE_%d", rcode)
}

func buildDNSMXPriorities(metadata map[string]dnsQueryMetadata) []dnsMXPriority {
	if metadata == nil {
		return nil
	}
	item, ok := metadata["MX"]
	if !ok || len(item.MXPriorities) == 0 {
		return nil
	}
	return item.MXPriorities
}

func buildDNSSOA(metadata map[string]dnsQueryMetadata) any {
	if metadata == nil {
		return nil
	}
	item, ok := metadata["SOA"]
	if !ok || item.SOA == nil {
		return nil
	}
	return item.SOA
}

func maxDNSRecordDepth(results map[string][]models.DNSRecord) int {
	maxDepth := 0
	for _, records := range results {
		for _, record := range records {
			if depth := dnsRecordDepth(record); depth > maxDepth {
				maxDepth = depth
			}
		}
	}
	return maxDepth
}

func dnsRecordDepth(record models.DNSRecord) int {
	maxChildDepth := 0
	for _, child := range record.Children {
		if depth := dnsRecordDepth(child); depth > maxChildDepth {
			maxChildDepth = depth
		}
	}
	if len(record.Children) == 0 {
		return 0
	}
	return maxChildDepth + 1
}

func limitDNSObservationString(value string) string {
	limited, _ := limitDNSObservationText(value)
	return limited
}

// lookupGeoASN 查询 IP 的国家、城市、ASN 和 ISP 信息
// 优先使用缓存，减少重复查询
func lookupGeoASN(ip net.IP, countryDB, cityDB, asnDB *geoip2.Reader) (string, string, string, string) {
	if val, ok := geoCache.Load(ip.String()); ok {
		data := val.([4]string)
		return data[0], data[1], data[2], data[3]
	}

	// 默认值
	country, city, asn, isp := "Unknown", "Unknown", "Unknown", "Unknown"

	// 查询国家信息
	if countryDB != nil {
		if rec, err := countryDB.Country(ip); err == nil {
			if n, ok := rec.Country.Names["en"]; ok {
				country = n
			}
		}
	}

	// 查询城市信息
	if cityDB != nil {
		if rec, err := cityDB.City(ip); err == nil {
			if n := rec.City.Names["en"]; n != "" {
				city = n
			}
			if n, ok := rec.Country.Names["en"]; ok {
				country = n
			}
		}
	}

	// 查询 ASN / ISP 信息
	if asnDB != nil {
		if rec, err := asnDB.ASN(ip); err == nil {
			asn = fmt.Sprintf("AS%d (%s)", rec.AutonomousSystemNumber, rec.AutonomousSystemOrganization)
			isp = rec.AutonomousSystemOrganization
		}
	}

	geoCache.Store(ip.String(), [4]string{country, city, asn, isp})
	return country, city, asn, isp
}

// detectCDN 判断 IP 是否属于 CDN 节点
func detectCDN(asn string, ip net.IP, domain string) string {
	for _, p := range models.CdnProviders {
		if strings.Contains(asn, p) || strings.Contains(strings.ToLower(domain), strings.ToLower(p)) {
			return "CDN"
		}
	}
	return "Origin"
}

// detectHijack 检测是否存在 DNS 劫持行为
// 包括私网 IP、TTL 异常、NXDOMAIN 等
func detectHijack(ip net.IP, msg *dns.Msg, ttl uint32) bool {
	risks := detectDNSRiskFlags(ip, msg, ttl, "skip_ptr_check")
	return containsRiskFlag(risks, dnsRiskPrivateIP) ||
		containsRiskFlag(risks, dnsRiskNXDomainWithAnswer) ||
		containsRiskFlag(risks, dnsRiskLowTTL)
}

func detectDNSRiskFlags(ip net.IP, msg *dns.Msg, ttl uint32, reversePTR string) []string {
	riskSet := map[string]struct{}{}
	privateRanges := []string{
		"0.0.0.0/8", "10.0.0.0/8", "127.0.0.0/8",
		"169.254.0.0/16", "172.16.0.0/12", "192.168.0.0/16",
		"100.64.0.0/10",
	}
	for _, cidr := range privateRanges {
		_, block, _ := net.ParseCIDR(cidr)
		if block.Contains(ip) {
			riskSet[dnsRiskPrivateIP] = struct{}{}
			break
		}
	}

	// RcodeNameError 且有 Answer 也可能劫持
	if msg != nil && msg.Rcode == dns.RcodeNameError && len(msg.Answer) > 0 {
		riskSet[dnsRiskNXDomainWithAnswer] = struct{}{}
	}

	// TTL 异常过低也认为可能劫持
	if ttl > 0 && ttl < 10 {
		riskSet[dnsRiskLowTTL] = struct{}{}
	}

	if reversePTR == "" {
		riskSet[dnsRiskPTREmpty] = struct{}{}
	}

	return sortedRiskFlags(riskSet)
}

func containsRiskFlag(risks []string, target string) bool {
	for _, risk := range risks {
		if risk == target {
			return true
		}
	}
	return false
}

// reversePTR 查询 IP 的 PTR 反向解析，使用并发限制
func reversePTR(ip net.IP) string {
	if val, ok := ptrCache.Load(ip.String()); ok {
		return val.(string)
	}

	// 并发限制
	ptrSem <- struct{}{}
	defer func() { <-ptrSem }()

	ctx, cancel := context.WithTimeout(context.Background(), env.GetServerConfig().Collector.ProbeBudget.PTRTimeout())
	defer cancel()

	names, err := net.DefaultResolver.LookupAddr(ctx, ip.String())
	ptr := ""
	if err == nil && len(names) > 0 {
		ptr = strings.Join(names, ",")
	}

	ptrCache.Store(ip.String(), ptr)
	return ptr
}

// ============== DNS解析 - 存储部分 ==============

func saveDNSResult() common.GFError {

	return nil
}

// printDNSRecord 格式化打印 DNSRecord，包括递归子记录
func printDNSRecord(rec models.DNSRecord, indent int) {
	prefix := strings.Repeat("  ", indent)
	fmt.Printf("%s[%s] %s TTL=%d DNSSEC=%v", prefix, rec.Type, rec.Value, rec.TTL, rec.DNSSEC)
	if rec.ASN != "" {
		fmt.Printf(" ASN=%s", rec.ASN)
	}
	if rec.Country != "" || rec.City != "" {
		fmt.Printf(" Geo=%s/%s", rec.Country, rec.City)
	}
	if rec.ISP != "" {
		fmt.Printf(" ISP=%s", rec.ISP)
	}
	if rec.ProviderType != "" {
		fmt.Printf(" Type=%s", rec.ProviderType)
	}
	if rec.ReversePTR != "" {
		fmt.Printf(" PTR=%s", rec.ReversePTR)
	}
	if rec.Hijacked {
		fmt.Printf("劫持嫌疑")
	}
	fmt.Printf(" 耗时=%v\n", rec.Duration)
	for _, child := range rec.Children {
		printDNSRecord(child, indent+1)
	}
}
