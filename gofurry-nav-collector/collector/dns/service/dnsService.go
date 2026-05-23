package service

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-collector/collector/dns/dao"
	"github.com/gofurry/gofurry-nav-collector/collector/dns/models"
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
)

// 按域名并行查 加锁
var dnsRunning atomic.Bool

// 缓存 IP
var geoCache sync.Map                        // geoCache 缓存 IP 的 GeoIP/ASN 查询结果
var ptrCache sync.Map                        // ptrCache 缓存 IP 的反向 PTR 查询结果
var ptrSem = make(chan struct{}, PTRWorkers) // ptrSem 用于限制 PTR 查询并发

var resolver = env.GetServerConfig().Collector.Dns.Resolver
var geoDBs *GeoDBSet

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
}

// ============== DNS解析 - 执行部分 ==============

// 执行 ParseDNS
func ParseDNS() {
	defer func() {
		if err := recover(); err != nil {
			log.ErrorFields(map[string]interface{}{
				"event":    "run_recovered",
				"protocol": "dns",
			}, fmt.Sprintf("DNS 采集运行触发 panic，已恢复: %v", err))
		}
	}()
	if !dnsRunning.CompareAndSwap(false, true) {
		log.WarnFields(map[string]interface{}{
			"event":    "run_skipped",
			"protocol": "dns",
			"reason":   "上一轮采集仍在运行",
			"status":   "skipped",
		}, "DNS 采集已跳过：上一轮仍在运行")
		return
	}
	defer dnsRunning.Store(false)

	start := time.Now()
	log.InfoFields(map[string]interface{}{
		"event":       "run_start",
		"max_depth":   MaxDepth,
		"protocol":    "dns",
		"resolver":    resolver,
		"timeout":     env.GetServerConfig().Collector.ProbeBudget.DNSTimeout(),
		"ptr_timeout": env.GetServerConfig().Collector.ProbeBudget.PTRTimeout(),
		"workers":     env.GetServerConfig().Collector.Dns.DnsThread,
	}, "DNS 采集运行开始")

	requestList, err := dao.GetDNSDao().GetList()
	if err != nil {
		log.ErrorFields(map[string]interface{}{
			"duration": time.Since(start),
			"event":    "run_failed",
			"protocol": "dns",
			"stage":    "load_targets",
		}, "DNS 目标列表读取失败: "+err.GetMsg())
		return
	}
	// 判空
	if len(requestList) < 1 {
		log.InfoFields(map[string]interface{}{
			"duration": time.Since(start),
			"event":    "run_complete",
			"protocol": "dns",
			"reason":   "目标列表为空",
			"targets":  0,
		}, "DNS 采集完成：没有需要探测的目标")
		return
	}

	log.InfoFields(map[string]interface{}{
		"event":    "probe_start",
		"protocol": "dns",
		"targets":  len(requestList),
		"workers":  env.GetServerConfig().Collector.Dns.DnsThread,
	}, "DNS 探测开始")
	dnsThread := pool.New().WithMaxGoroutines(env.GetServerConfig().Collector.Dns.DnsThread)
	// 遍历站点列表, 每个站点开一个线程执行采集
	for _, v := range requestList {
		dnsThread.Go(getDNSResult(v))
	}
	// 等待所有采集和解析执行完毕
	dnsThread.Wait()
	log.InfoFields(map[string]interface{}{
		"duration": time.Since(start),
		"event":    "run_complete",
		"protocol": "dns",
		"targets":  len(requestList),
	}, "DNS 采集运行完成")
}

func getDNSResult(site models.GfnCollectorDomain) func() {
	return func() {
		defer func() {
			if err := recover(); err != nil {
				log.ErrorFields(map[string]interface{}{
					"event":    "probe_recovered",
					"protocol": "dns",
					"site":     site.Name,
				}, fmt.Sprintf("DNS 单目标探测触发 panic，已恢复: %v", err))
			}
		}()

		// 执行 Request 获取结果
		geoDBSet := currentGeoDBs()
		results := performDNSQuery(site, geoDBSet.ASN, geoDBSet.City, geoDBSet.Country)

		var siteName string
		if site.Prefix != nil {
			siteName = *site.Prefix + site.Name
		} else {
			siteName = site.Name
		}

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

	}
}

func performDNSQuery(site models.GfnCollectorDomain, asnDB *geoip2.Reader, cityDB *geoip2.Reader, countryDB *geoip2.Reader) map[string][]models.DNSRecord {
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

	var domain string
	if site.Prefix != nil {
		domain = *site.Prefix + site.Name
	} else {
		domain = site.Name
	}

	// 并行查询每种记录类型
	for _, rt := range models.RecordTypes {
		queryMG.Add(1)
		go func(rt models.RecordType) {
			defer queryMG.Done()

			records, stats, err := queryDNS(domain, rt.Type, resolver, countryDB, cityDB, asnDB, 0)
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
	return result
}

// ============== DNS解析 - 采集和解析部分 ==============

func queryDNS(domain string, qtype uint16, resolver string, asnDB *geoip2.Reader, cityDB *geoip2.Reader, countryDB *geoip2.Reader, depth int) ([]models.DNSRecord, models.DNSStatistics, common.GFError) {
	// 防止递归过深
	if depth > MaxDepth {
		return nil, models.DNSStatistics{}, nil
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
		return nil, models.DNSStatistics{}, common.NewServiceError("DNS 查询失败: " + err.Error())
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
			rec.Hijacked = detectHijack(v.A, in, v.Hdr.Ttl)
		case *dns.AAAA:
			rec.Value = v.AAAA.String()
			rec.Country, rec.City, rec.ASN, rec.ISP = lookupGeoASN(v.AAAA, countryDB, cityDB, asnDB)
			rec.ProviderType = detectCDN(rec.ASN, v.AAAA, domain)
			rec.ReversePTR = reversePTR(v.AAAA)
			rec.Hijacked = detectHijack(v.AAAA, in, v.Hdr.Ttl)
		case *dns.CNAME:
			rec.Value = v.Target
			// 递归查询 CNAME 指向的 A/AAAA
			childrenA, _, _ := queryDNS(v.Target, dns.TypeA, resolver, countryDB, cityDB, asnDB, depth+1)
			childrenAAAA, _, _ := queryDNS(v.Target, dns.TypeAAAA, resolver, countryDB, cityDB, asnDB, depth+1)
			rec.Children = append(rec.Children, childrenA...)
			rec.Children = append(rec.Children, childrenAAAA...)
		case *dns.MX:
			rec.Value = fmt.Sprintf("%s (优先级 %d)", v.Mx, v.Preference)
			childrenA, _, _ := queryDNS(v.Mx, dns.TypeA, resolver, countryDB, cityDB, asnDB, depth+1)
			childrenAAAA, _, _ := queryDNS(v.Mx, dns.TypeAAAA, resolver, countryDB, cityDB, asnDB, depth+1)
			rec.Children = append(rec.Children, childrenA...)
			rec.Children = append(rec.Children, childrenAAAA...)
		case *dns.NS:
			rec.Value = v.Ns
			childrenA, _, _ := queryDNS(v.Ns, dns.TypeA, resolver, countryDB, cityDB, asnDB, depth+1)
			childrenAAAA, _, _ := queryDNS(v.Ns, dns.TypeAAAA, resolver, countryDB, cityDB, asnDB, depth+1)
			rec.Children = append(rec.Children, childrenA...)
			rec.Children = append(rec.Children, childrenAAAA...)
		case *dns.TXT:
			rec.Value = strings.Join(v.Txt, " ")
		case *dns.SOA:
			rec.Value = fmt.Sprintf("%s %s", v.Ns, v.Mbox)
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

	return results, stats, nil
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
	privateRanges := []string{
		"0.0.0.0/8", "10.0.0.0/8", "127.0.0.0/8",
		"169.254.0.0/16", "172.16.0.0/12", "192.168.0.0/16",
		"100.64.0.0/10",
	}
	for _, cidr := range privateRanges {
		_, block, _ := net.ParseCIDR(cidr)
		if block.Contains(ip) {
			return true
		}
	}

	// RcodeNameError 且有 Answer 也可能劫持
	if msg.Rcode == dns.RcodeNameError && len(msg.Answer) > 0 {
		return true
	}

	// TTL 异常过低也认为可能劫持
	if ttl > 0 && ttl < 10 {
		return true
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
