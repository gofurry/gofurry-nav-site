package observation

import (
	"fmt"
	"sort"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/log"
	cs "github.com/gofurry/gofurry-nav-collector/common/service"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
)

const (
	StatusHealthy  = "healthy"
	StatusWarning  = "warning"
	StatusDegraded = "degraded"
	StatusUnknown  = "unknown"
	StatusDown     = "down"
)

func TargetLatestKey(protocol string, siteID int64, target string) string {
	return fmt.Sprintf("collector:v2:latest:%s:%d:%s", protocol, siteID, target)
}

func TargetSummaryKey(siteID int64, target string) string {
	return fmt.Sprintf("collector:v2:summary:target:%d:%s", siteID, target)
}

func SiteSummaryTargetsKey(siteID int64) string {
	return fmt.Sprintf("collector:v2:summary:site_targets:%d", siteID)
}

func SiteSummaryKey(siteID int64) string {
	return fmt.Sprintf("collector:v2:summary:site:%d", siteID)
}

func UpdateSummaryIfEnabled(siteID int64, target string) common.GFError {
	if siteID <= 0 || target == "" {
		return nil
	}
	cfg := env.GetServerConfig().Collector.V2
	if !cfg.Enabled || !cfg.LatestRedis {
		return nil
	}

	now := time.Now()
	docs := latestDocumentsForTarget(siteID, target)
	targetSummary := BuildTargetSummary(siteID, target, docs, now)
	if err := writeJSON(TargetSummaryKey(siteID, target), targetSummary); err != nil {
		log.ErrorFields(map[string]interface{}{
			"event":     "v2_target_summary_write_failed",
			"redis_key": TargetSummaryKey(siteID, target),
			"site_id":   siteID,
			"target":    target,
		}, "v2 target summary Redis 写入失败: "+err.GetMsg())
		return err
	}
	if err := rememberSummaryTarget(siteID, target); err != nil {
		log.WarnFields(map[string]interface{}{
			"event":     "v2_summary_target_index_failed",
			"redis_key": SiteSummaryTargetsKey(siteID),
			"site_id":   siteID,
			"target":    target,
		}, "v2 target summary 索引写入 Redis 失败: "+err.GetMsg())
	}

	if err := UpdateSiteSummary(siteID, now); err != nil {
		log.ErrorFields(map[string]interface{}{
			"event":     "v2_site_summary_write_failed",
			"redis_key": SiteSummaryKey(siteID),
			"site_id":   siteID,
		}, "v2 site summary Redis 写入失败: "+err.GetMsg())
		return err
	}
	return nil
}

func UpdateSiteSummary(siteID int64, now time.Time) common.GFError {
	keys, err := targetSummaryKeysForSite(siteID)
	if err != nil {
		return err
	}
	summaries := make([]TargetSummaryDocument, 0, len(keys))
	for _, key := range keys {
		raw, getErr := cs.Get(key)
		if getErr != nil {
			return getErr
		}
		if raw == "" {
			forgetSummaryTargetByKey(siteID, key)
			continue
		}
		var summary TargetSummaryDocument
		if jsonErr := sonic.UnmarshalString(raw, &summary); jsonErr != nil {
			log.WarnFields(map[string]interface{}{
				"event":     "v2_target_summary_decode_failed",
				"redis_key": key,
				"site_id":   siteID,
			}, "v2 target summary JSON 解析失败: "+jsonErr.Error())
			continue
		}
		if summary.SiteID == siteID {
			summaries = append(summaries, summary)
		} else {
			forgetSummaryTargetByKey(siteID, key)
		}
	}

	siteSummary := BuildSiteSummary(siteID, summaries, now)
	return writeJSON(SiteSummaryKey(siteID), siteSummary)
}

func rememberSummaryTarget(siteID int64, target string) common.GFError {
	if siteID <= 0 || target == "" {
		return nil
	}
	return cs.SAdd(SiteSummaryTargetsKey(siteID), target)
}

func forgetSummaryTarget(siteID int64, target string) common.GFError {
	if siteID <= 0 || target == "" {
		return nil
	}
	return cs.SRem(SiteSummaryTargetsKey(siteID), target)
}

func forgetSummaryTargetByKey(siteID int64, key string) {
	target := targetFromSummaryKey(siteID, key)
	if target == "" {
		return
	}
	if err := forgetSummaryTarget(siteID, target); err != nil {
		log.WarnFields(map[string]interface{}{
			"event":     "v2_summary_target_index_cleanup_failed",
			"redis_key": SiteSummaryTargetsKey(siteID),
			"site_id":   siteID,
			"target":    target,
		}, "v2 target summary 索引清理 Redis 失败: "+err.GetMsg())
	}
}

func targetFromSummaryKey(siteID int64, key string) string {
	prefix := TargetSummaryKey(siteID, "")
	if len(key) <= len(prefix) || key[:len(prefix)] != prefix {
		return ""
	}
	return key[len(prefix):]
}

func targetSummaryKeysForSite(siteID int64) ([]string, common.GFError) {
	targets, err := cs.SMembers(SiteSummaryTargetsKey(siteID))
	if err != nil {
		return nil, err
	}
	if len(targets) > 0 {
		keys := make([]string, 0, len(targets))
		for _, target := range targets {
			if target == "" {
				continue
			}
			keys = append(keys, TargetSummaryKey(siteID, target))
		}
		sort.Strings(keys)
		return keys, nil
	}
	return cs.FindByPrefix(fmt.Sprintf("collector:v2:summary:target:%d:", siteID))
}

func latestDocumentsForTarget(siteID int64, target string) map[string]LatestDocument {
	docs := map[string]LatestDocument{}
	for _, protocol := range enabledLatestProtocols() {
		key := TargetLatestKey(protocol, siteID, target)
		raw, err := cs.Get(key)
		if err != nil {
			log.WarnFields(map[string]interface{}{
				"event":     "v2_target_latest_read_failed",
				"protocol":  protocol,
				"redis_key": key,
				"site_id":   siteID,
				"target":    target,
			}, "v2 target latest Redis 读取失败: "+err.GetMsg())
			continue
		}
		if raw == "" {
			continue
		}
		var doc LatestDocument
		if jsonErr := sonic.UnmarshalString(raw, &doc); jsonErr != nil {
			log.WarnFields(map[string]interface{}{
				"event":     "v2_target_latest_decode_failed",
				"protocol":  protocol,
				"redis_key": key,
				"site_id":   siteID,
				"target":    target,
			}, "v2 target latest JSON 解析失败: "+jsonErr.Error())
			continue
		}
		docs[protocol] = doc
	}
	return docs
}

func enabledLatestProtocols() []string {
	cfg := env.GetServerConfig().Collector.V2
	protocols := make([]string, 0, 3)
	for _, protocol := range []string{ProtocolHTTP, ProtocolDNS, ProtocolPing} {
		if cfg.LatestRedisEnabled(protocol) {
			protocols = append(protocols, protocol)
		}
	}
	return protocols
}

func writeJSON(key string, value any) common.GFError {
	bytes, err := sonic.Marshal(value)
	if err != nil {
		return common.NewServiceError("v2 summary 编码失败")
	}
	return cs.Set(key, string(bytes))
}

func BuildTargetSummary(siteID int64, target string, docs map[string]LatestDocument, now time.Time) TargetSummaryDocument {
	protocols := make(map[string]ProtocolSummary, len(docs))
	healthByProtocol := make(map[string]protocolHealth, len(docs))
	var observedAt time.Time
	for protocol, doc := range docs {
		staleAfter := staleAfterForProtocol(protocol)
		stale := doc.ObservedAt.IsZero() || now.Sub(doc.ObservedAt) > staleAfter
		protocols[protocol] = ProtocolSummary{
			Protocol:          protocol,
			Status:            doc.Status,
			ObservedAt:        doc.ObservedAt,
			DurationMS:        doc.DurationMS,
			Stale:             stale,
			StaleAfterSeconds: int64(staleAfter.Seconds()),
			ErrorCode:         doc.ErrorCode,
		}
		healthByProtocol[protocol] = protocolHealth{
			doc:   doc,
			stale: stale,
		}
		if doc.ObservedAt.After(observedAt) {
			observedAt = doc.ObservedAt
		}
	}

	status, reasonCodes, reasonMessages := evaluateTargetHealth(healthByProtocol, now)
	var edgeHints []EdgeProviderHint
	if env.GetServerConfig().Collector.V2.EdgeHints.EnabledOrDefault() {
		edgeHints = BuildEdgeProviderHints(docs)
	}
	return TargetSummaryDocument{
		SiteID:            siteID,
		Target:            target,
		Status:            status,
		ReasonCodes:       reasonCodes,
		ReasonMessages:    reasonMessages,
		Protocols:         protocols,
		EdgeProviderHints: edgeHints,
		ObservedAt:        observedAt,
		GeneratedAt:       now,
		SchemaVersion:     schemaVersion,
	}
}

func BuildSiteSummary(siteID int64, targetSummaries []TargetSummaryDocument, now time.Time) SiteSummaryDocument {
	sort.Slice(targetSummaries, func(i, j int) bool {
		return targetSummaries[i].Target < targetSummaries[j].Target
	})

	counts := map[string]int{
		StatusHealthy:  0,
		StatusWarning:  0,
		StatusDegraded: 0,
		StatusUnknown:  0,
		StatusDown:     0,
	}
	targets := make([]TargetSummaryItem, 0, len(targetSummaries))
	for _, summary := range targetSummaries {
		counts[summary.Status]++
		targets = append(targets, TargetSummaryItem{
			Target:            summary.Target,
			Status:            summary.Status,
			ReasonCodes:       summary.ReasonCodes,
			ReasonMessages:    summary.ReasonMessages,
			EdgeProviderHints: summary.EdgeProviderHints,
			ObservedAt:        summary.ObservedAt,
		})
	}

	status, reasonCodes, reasonMessages := evaluateSiteHealth(counts, len(targetSummaries))
	return SiteSummaryDocument{
		SiteID:         siteID,
		Status:         status,
		ReasonCodes:    reasonCodes,
		ReasonMessages: reasonMessages,
		TargetCount:    len(targetSummaries),
		StatusCounts:   counts,
		Targets:        targets,
		GeneratedAt:    now,
		SchemaVersion:  schemaVersion,
	}
}

type protocolHealth struct {
	doc   LatestDocument
	stale bool
}

func evaluateTargetHealth(protocols map[string]protocolHealth, now time.Time) (string, []string, []string) {
	reasons := reasonCollector{}
	httpDoc, hasHTTP := protocols[ProtocolHTTP]
	if !hasHTTP || httpDoc.stale {
		reasons.add("http_missing_or_stale", "HTTP 观测缺失或已过期，无法判断访客是否可打开")
		return StatusUnknown, reasons.codes, reasons.messages
	}

	httpSuccess := httpDoc.doc.Status == StatusSuccess
	dnsFailure := freshProtocolFailed(protocols, ProtocolDNS)
	pingFailure := freshProtocolFailed(protocols, ProtocolPing)
	dnsStale := protocolMissingOrStale(protocols, ProtocolDNS)

	if !httpSuccess {
		reasons.add("http_failed", "HTTP 访问失败")
		if dnsFailure {
			reasons.add("dns_failed", "DNS 解析也失败")
			return StatusDown, reasons.codes, reasons.messages
		}
		if dnsStale {
			reasons.add("dns_missing_or_stale", "DNS 观测缺失或已过期")
		}
		return StatusDegraded, reasons.codes, reasons.messages
	}

	status := StatusHealthy
	if dnsFailure {
		status = worseStatus(status, StatusWarning)
		reasons.add("dns_failed_but_http_ok", "DNS 失败但 HTTP 当前仍可访问")
	} else if dnsStale {
		status = worseStatus(status, StatusWarning)
		reasons.add("dns_missing_or_stale", "DNS 观测缺失或已过期")
	}
	if pingFailure {
		status = worseStatus(status, StatusWarning)
		reasons.add("ping_failed_but_http_ok", "Ping 失败但 HTTP 当前仍可访问")
	}

	dnsRisks := riskFlagsFromPayload(protocols[ProtocolDNS].doc.Payload)
	if len(dnsRisks) > 0 {
		status = worseStatus(status, StatusWarning)
		for _, risk := range dnsRisks {
			reasons.add("dns_risk_"+risk, "DNS observation 出现风险信号: "+risk)
		}
	}

	tlsStatus, tlsCodes, tlsMessages := evaluateTLSSignal(httpDoc.doc.Payload, now)
	status = worseStatus(status, tlsStatus)
	for i, code := range tlsCodes {
		message := ""
		if i < len(tlsMessages) {
			message = tlsMessages[i]
		}
		reasons.add(code, message)
	}

	return status, reasons.codes, reasons.messages
}

func evaluateSiteHealth(counts map[string]int, targetCount int) (string, []string, []string) {
	reasons := reasonCollector{}
	if targetCount == 0 {
		reasons.add("no_target_summary", "没有可用的采集目标健康摘要")
		return StatusUnknown, reasons.codes, reasons.messages
	}
	if counts[StatusDown] == targetCount {
		reasons.add("all_targets_down", "所有采集目标都判定为 down")
		return StatusDown, reasons.codes, reasons.messages
	}
	if counts[StatusUnknown] == targetCount {
		reasons.add("all_targets_unknown", "所有采集目标状态未知")
		return StatusUnknown, reasons.codes, reasons.messages
	}
	if counts[StatusDown] > 0 || counts[StatusDegraded] > 0 {
		reasons.add("some_targets_degraded", "部分采集目标不可用或降级")
		return StatusDegraded, reasons.codes, reasons.messages
	}
	if counts[StatusWarning] > 0 || counts[StatusUnknown] > 0 {
		reasons.add("some_targets_warning", "部分采集目标存在需要关注的观测信号")
		return StatusWarning, reasons.codes, reasons.messages
	}
	return StatusHealthy, reasons.codes, reasons.messages
}

func freshProtocolFailed(protocols map[string]protocolHealth, protocol string) bool {
	value, ok := protocols[protocol]
	return ok && !value.stale && value.doc.Status != StatusSuccess
}

func protocolMissingOrStale(protocols map[string]protocolHealth, protocol string) bool {
	value, ok := protocols[protocol]
	if !ok {
		return env.GetServerConfig().Collector.V2.LatestRedisEnabled(protocol)
	}
	return value.stale
}

func evaluateTLSSignal(payload any, now time.Time) (string, []string, []string) {
	payloadMap, ok := payload.(map[string]any)
	if !ok {
		return StatusHealthy, nil, nil
	}
	handshake := stringFromMap(payloadMap, "tls_handshake")
	if handshake == "not_tls" || handshake == "" {
		return StatusHealthy, nil, nil
	}

	reasons := reasonCollector{}
	status := StatusHealthy
	certVerified, hasCertVerified := boolFromMap(payloadMap, "cert_verified")
	category := stringFromMap(payloadMap, "verify_error_category")
	if hasCertVerified && !certVerified {
		status = worseStatus(status, StatusDegraded)
		if category == "" {
			category = "other"
		}
		reasons.add("tls_verify_"+category, "TLS 证书校验未通过: "+category)
	}

	notAfter := stringFromMap(payloadMap, "cert_not_after")
	if notAfter != "" {
		if expiresAt, err := time.Parse(time.RFC3339, notAfter); err == nil {
			daysLeft := int(expiresAt.Sub(now).Hours() / 24)
			switch {
			case daysLeft < 0:
				status = worseStatus(status, StatusDegraded)
				reasons.add("tls_cert_expired", "TLS 证书已过期")
			case daysLeft <= 30:
				status = worseStatus(status, StatusWarning)
				reasons.add("tls_cert_expiring_soon", "TLS 证书将在 30 天内过期")
			}
		}
	}
	return status, reasons.codes, reasons.messages
}

func riskFlagsFromPayload(payload any) []string {
	payloadMap, ok := payload.(map[string]any)
	if !ok {
		return nil
	}
	raw, ok := payloadMap["risk_flags"]
	if !ok {
		return nil
	}
	flags := []string{}
	switch values := raw.(type) {
	case []any:
		for _, value := range values {
			if flag, ok := value.(string); ok && flag != "" {
				flags = append(flags, flag)
			}
		}
	case []string:
		flags = append(flags, values...)
	}
	sort.Strings(flags)
	return flags
}

func boolFromMap(values map[string]any, key string) (bool, bool) {
	raw, ok := values[key]
	if !ok {
		return false, false
	}
	value, ok := raw.(bool)
	return value, ok
}

func stringFromMap(values map[string]any, key string) string {
	raw, ok := values[key]
	if !ok {
		return ""
	}
	value, _ := raw.(string)
	return value
}

func staleAfterForProtocol(protocol string) time.Duration {
	cfg := env.GetServerConfig().Collector
	switch protocol {
	case ProtocolPing:
		seconds := cfg.Ping.PingInterval
		if seconds <= 0 {
			seconds = 600
		}
		return time.Duration(seconds*3) * time.Second
	case ProtocolHTTP:
		hours := cfg.Request.RequestInterval
		if hours <= 0 {
			hours = 6
		}
		return time.Duration(hours*2) * time.Hour
	case ProtocolDNS:
		hours := cfg.Dns.DnsInterval
		if hours <= 0 {
			hours = 24
		}
		return time.Duration(hours*2) * time.Hour
	default:
		return time.Hour
	}
}

func worseStatus(current string, candidate string) string {
	if statusSeverity(candidate) > statusSeverity(current) {
		return candidate
	}
	return current
}

func statusSeverity(status string) int {
	switch status {
	case StatusDown:
		return 4
	case StatusDegraded:
		return 3
	case StatusWarning:
		return 2
	case StatusUnknown:
		return 1
	case StatusHealthy:
		return 0
	default:
		return 1
	}
}

type reasonCollector struct {
	codes    []string
	messages []string
	seen     map[string]bool
}

func (collector *reasonCollector) add(code string, message string) {
	if code == "" {
		return
	}
	if collector.seen == nil {
		collector.seen = map[string]bool{}
	}
	if collector.seen[code] {
		return
	}
	collector.seen[code] = true
	collector.codes = append(collector.codes, code)
	if message != "" {
		collector.messages = append(collector.messages, message)
	}
}
