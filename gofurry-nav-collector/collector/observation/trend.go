package observation

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/log"
	cs "github.com/gofurry/gofurry-nav-collector/common/service"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
)

const (
	defaultTrendLookback = 7 * 24 * time.Hour
	defaultTrendTimeout  = 2 * time.Second
	defaultTrendMaxRows  = 3000
)

var (
	derivedRunMu sync.Mutex
	derivedRunAt = map[string]time.Time{}
)

func TargetTrendKey(siteID int64, target string) string {
	return fmt.Sprintf("collector:v2:trend:target:%d:%s", siteID, target)
}

func UpdateTrendIfEnabled(siteID int64, target string, now time.Time) common.GFError {
	if siteID <= 0 || target == "" {
		return nil
	}
	cfg := env.GetServerConfig().Collector.V2
	if !cfg.Enabled || !cfg.ObservationDB || !cfg.LatestRedis {
		return nil
	}
	if !cfg.Derived.TrendEnabledOrDefault() {
		return nil
	}
	if !reserveDerivedRun("trend", siteID, target, now, cfg.Derived.MinInterval()) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Derived.QueryTimeout())
	defer cancel()
	rows, err := GetObservationDao().ListTrendRows(ctx, siteID, target, now.Add(-defaultTrendLookback), cfg.Derived.TrendRows())
	if err != nil {
		log.WarnFields(map[string]interface{}{
			"event":   "v2_trend_query_failed",
			"site_id": siteID,
			"target":  target,
		}, "v2 observation 趋势查询失败: "+err.GetMsg())
		return err
	}
	doc := BuildTargetTrend(siteID, target, rows, now)
	bytes, marshalErr := sonic.Marshal(doc)
	if marshalErr != nil {
		log.WarnFields(map[string]interface{}{
			"event":   "v2_trend_encode_failed",
			"site_id": siteID,
			"target":  target,
		}, "v2 observation 趋势 JSON 编码失败: "+marshalErr.Error())
		return common.NewServiceError("v2 trend 编码失败")
	}
	if err := cs.Set(TargetTrendKey(siteID, target), string(bytes)); err != nil {
		log.WarnFields(map[string]interface{}{
			"event":     "v2_trend_redis_write_failed",
			"redis_key": TargetTrendKey(siteID, target),
			"site_id":   siteID,
			"target":    target,
		}, "v2 observation 趋势写入 Redis 失败: "+err.GetMsg())
		return err
	}
	return nil
}

func reserveDerivedRun(kind string, siteID int64, target string, now time.Time, minInterval time.Duration) bool {
	if kind == "" || siteID <= 0 || target == "" || minInterval <= 0 {
		return true
	}
	key := fmt.Sprintf("%s:%d:%s", kind, siteID, target)
	derivedRunMu.Lock()
	defer derivedRunMu.Unlock()
	last, ok := derivedRunAt[key]
	if ok && now.Sub(last) < minInterval {
		return false
	}
	derivedRunAt[key] = now
	return true
}

func resetDerivedRunGateForTest() {
	derivedRunMu.Lock()
	defer derivedRunMu.Unlock()
	derivedRunAt = map[string]time.Time{}
}

func BuildTargetTrend(siteID int64, target string, rows []ObservationTrendRow, now time.Time) TargetTrendDocument {
	windows := map[string]TrendWindow{
		"24h": buildTrendWindow(rows, now.Add(-24*time.Hour), now),
		"7d":  buildTrendWindow(rows, now.Add(-7*24*time.Hour), now),
	}
	return TargetTrendDocument{
		SiteID:        siteID,
		Target:        target,
		Windows:       windows,
		GeneratedAt:   now,
		SchemaVersion: schemaVersion,
	}
}

func buildTrendWindow(rows []ObservationTrendRow, since time.Time, now time.Time) TrendWindow {
	grouped := map[string][]trendObservation{}
	for _, row := range rows {
		if row.ObservedAt.Before(since) || row.ObservedAt.After(now) {
			continue
		}
		if row.Protocol != ProtocolPing && row.Protocol != ProtocolHTTP && row.Protocol != ProtocolDNS {
			continue
		}
		grouped[row.Protocol] = append(grouped[row.Protocol], decodeTrendObservation(row))
	}
	protocols := map[string]ProtocolTrend{}
	for _, protocol := range []string{ProtocolHTTP, ProtocolPing, ProtocolDNS} {
		values := grouped[protocol]
		if len(values) == 0 {
			continue
		}
		sort.Slice(values, func(i, j int) bool {
			return values[i].ObservedAt.After(values[j].ObservedAt)
		})
		protocols[protocol] = buildProtocolTrend(protocol, values, now)
	}
	return TrendWindow{
		Since:     since,
		Until:     now,
		Protocols: protocols,
	}
}

type trendObservation struct {
	Protocol   string
	Status     string
	ObservedAt time.Time
	DurationMS int64
	ErrorCode  string
	Payload    map[string]any
}

func decodeTrendObservation(row ObservationTrendRow) trendObservation {
	payload := map[string]any{}
	if row.Payload != "" {
		_ = sonic.UnmarshalString(row.Payload, &payload)
	}
	errorCode := ""
	if row.ErrorCode != nil {
		errorCode = *row.ErrorCode
	}
	return trendObservation{
		Protocol:   row.Protocol,
		Status:     row.Status,
		ObservedAt: row.ObservedAt,
		DurationMS: row.DurationMS,
		ErrorCode:  errorCode,
		Payload:    payload,
	}
}

func buildProtocolTrend(protocol string, values []trendObservation, now time.Time) ProtocolTrend {
	successCount := 0
	var lastFailureAt *time.Time
	durations := make([]float64, 0, len(values))
	for _, value := range values {
		if value.Status == StatusSuccess {
			successCount++
		} else if lastFailureAt == nil || value.ObservedAt.After(*lastFailureAt) {
			observedAt := value.ObservedAt
			lastFailureAt = &observedAt
		}
		if value.DurationMS > 0 {
			durations = append(durations, float64(value.DurationMS))
		}
	}
	successRate := ratio(successCount, len(values))
	latestObservedAt := values[0].ObservedAt
	trend := ProtocolTrend{
		Protocol:         protocol,
		ObservationCount: len(values),
		SuccessCount:     successCount,
		FailureCount:     len(values) - successCount,
		SuccessRate:      &successRate,
		AvgDurationMS:    avgFloat64(durations),
		P95DurationMS:    percentileFloat64(durations, 0.95),
		LastObservedAt:   &latestObservedAt,
		LastFailureAt:    lastFailureAt,
	}
	switch protocol {
	case ProtocolHTTP:
		trend.HTTP = buildHTTPTrend(values)
		trend.TLS = buildTLSTrend(values, now)
	case ProtocolPing:
		trend.Ping = buildPingTrend(values)
	case ProtocolDNS:
		trend.DNS = buildDNSTrend(values)
	}
	return trend
}

func buildHTTPTrend(values []trendObservation) *HTTPTrend {
	durations := make([]float64, 0, len(values))
	var latestFailureAt *time.Time
	for _, value := range values {
		if responseTime, ok := floatFromMap(value.Payload, "response_time_ms"); ok && responseTime > 0 {
			durations = append(durations, responseTime)
		} else if value.DurationMS > 0 {
			durations = append(durations, float64(value.DurationMS))
		}
		if value.Status != StatusSuccess && (latestFailureAt == nil || value.ObservedAt.After(*latestFailureAt)) {
			observedAt := value.ObservedAt
			latestFailureAt = &observedAt
		}
	}
	return &HTTPTrend{
		AvgResponseTimeMS: avgFloat64(durations),
		P95ResponseTimeMS: percentileFloat64(durations, 0.95),
		LatestFailureAt:   latestFailureAt,
	}
}

func buildPingTrend(values []trendObservation) *PingTrend {
	rtts := make([]float64, 0, len(values))
	lossRates := make([]float64, 0, len(values))
	jitters := make([]float64, 0, len(values))
	for _, value := range values {
		if rtt, ok := floatFromMap(value.Payload, "avg_rtt_ms"); ok {
			rtts = append(rtts, rtt)
		}
		if loss, ok := floatFromMap(value.Payload, "loss_rate"); ok {
			lossRates = append(lossRates, loss)
		}
		if jitter, ok := floatFromMap(value.Payload, "jitter_ms"); ok {
			jitters = append(jitters, jitter)
		}
	}
	return &PingTrend{
		AvgRTTMS:       avgFloat64(rtts),
		AvgLossRate:    avgFloat64(lossRates),
		AvgJitterMS:    avgFloat64(jitters),
		LatestAvgRTTMS: latestFloatFromPayload(values, "avg_rtt_ms"),
		LatestLossRate: latestFloatFromPayload(values, "loss_rate"),
		LatestJitterMS: latestFloatFromPayload(values, "jitter_ms"),
	}
}

func buildDNSTrend(values []trendObservation) *DNSTrend {
	riskCounts := map[string]int{}
	var latestRisks []string
	var latestTTL *dnsTTLStats
	var previousTTL *dnsTTLStats
	for index, value := range values {
		risks := riskFlagsFromPayload(value.Payload)
		if index == 0 {
			latestRisks = risks
		}
		for _, risk := range risks {
			riskCounts[risk]++
		}
		ttl := ttlStatsFromDNSPayload(value.Payload)
		if ttl != nil {
			if latestTTL == nil {
				latestTTL = ttl
			} else {
				previousTTL = ttl
			}
		}
	}
	successCount := 0
	for _, value := range values {
		if value.Status == StatusSuccess {
			successCount++
		}
	}
	successRate := ratio(successCount, len(values))
	trend := &DNSTrend{
		SuccessRate:     &successRate,
		RiskFlagCounts:  riskCounts,
		LatestRiskFlags: latestRisks,
	}
	if len(riskCounts) == 0 {
		trend.RiskFlagCounts = nil
	}
	if latestTTL != nil {
		trend.LatestTTLMin = latestTTL.min
		trend.LatestTTLMax = latestTTL.max
		trend.LatestTTLAvg = latestTTL.avg
	}
	if previousTTL != nil {
		trend.PreviousTTLMin = previousTTL.min
		trend.PreviousTTLMax = previousTTL.max
		trend.PreviousTTLAvg = previousTTL.avg
	}
	return trend
}

func buildTLSTrend(values []trendObservation, now time.Time) *TLSTrend {
	issuers := map[string]struct{}{}
	fingerprints := map[string]struct{}{}
	var latestDays *int
	var previousDays *int
	var latestIssuer string
	var latestFingerprint string
	var latestNotAfter string
	var latestObservedAt *time.Time
	for _, value := range values {
		issuer := stringFromMap(value.Payload, "cert_issuer")
		if issuer != "" {
			issuers[issuer] = struct{}{}
		}
		fingerprint := stringFromMap(value.Payload, "cert_fingerprint_sha256")
		if fingerprint != "" {
			fingerprints[fingerprint] = struct{}{}
		}
		notAfter := stringFromMap(value.Payload, "cert_not_after")
		if notAfter == "" {
			continue
		}
		daysLeft, ok := certDaysLeft(notAfter, now)
		if !ok {
			continue
		}
		if latestDays == nil {
			latestDays = &daysLeft
			latestIssuer = issuer
			latestFingerprint = fingerprint
			latestNotAfter = notAfter
			observedAt := value.ObservedAt
			latestObservedAt = &observedAt
		} else {
			previousDays = &daysLeft
		}
	}
	if latestDays == nil && len(issuers) == 0 && len(fingerprints) == 0 {
		return nil
	}
	return &TLSTrend{
		LatestCertDaysLeft:     latestDays,
		PreviousCertDaysLeft:   previousDays,
		CertIssuerChanged:      len(issuers) > 1,
		CertFingerprintChanged: len(fingerprints) > 1,
		LatestCertIssuer:       latestIssuer,
		LatestCertFingerprint:  latestFingerprint,
		LatestCertNotAfter:     latestNotAfter,
		LatestCertObservedAt:   latestObservedAt,
	}
}

type dnsTTLStats struct {
	min *float64
	max *float64
	avg *float64
}

func ttlStatsFromDNSPayload(payload map[string]any) *dnsTTLStats {
	responseSummary := mapFromAny(payload["response_summary"])
	if len(responseSummary) == 0 {
		return nil
	}
	minValues := []float64{}
	maxValues := []float64{}
	avgValues := []float64{}
	for _, value := range responseSummary {
		summary := mapFromAny(value)
		if ttl, ok := floatFromMap(summary, "ttl_min"); ok {
			minValues = append(minValues, ttl)
		}
		if ttl, ok := floatFromMap(summary, "ttl_max"); ok {
			maxValues = append(maxValues, ttl)
		}
		if ttl, ok := floatFromMap(summary, "ttl_avg"); ok {
			avgValues = append(avgValues, ttl)
		}
	}
	if len(minValues) == 0 && len(maxValues) == 0 && len(avgValues) == 0 {
		return nil
	}
	return &dnsTTLStats{
		min: minFloat64(minValues),
		max: maxFloat64(maxValues),
		avg: avgFloat64(avgValues),
	}
}

func latestFloatFromPayload(values []trendObservation, key string) *float64 {
	for _, value := range values {
		if number, ok := floatFromMap(value.Payload, key); ok {
			return &number
		}
	}
	return nil
}

func floatFromMap(values map[string]any, key string) (float64, bool) {
	raw, ok := values[key]
	if !ok || raw == nil {
		return 0, false
	}
	switch value := raw.(type) {
	case float64:
		return value, true
	case float32:
		return float64(value), true
	case int:
		return float64(value), true
	case int64:
		return float64(value), true
	case uint64:
		return float64(value), true
	default:
		return 0, false
	}
}

func avgFloat64(values []float64) *float64 {
	if len(values) == 0 {
		return nil
	}
	total := 0.0
	for _, value := range values {
		total += value
	}
	avg := total / float64(len(values))
	return &avg
}

func minFloat64(values []float64) *float64 {
	if len(values) == 0 {
		return nil
	}
	min := values[0]
	for _, value := range values[1:] {
		if value < min {
			min = value
		}
	}
	return &min
}

func maxFloat64(values []float64) *float64 {
	if len(values) == 0 {
		return nil
	}
	max := values[0]
	for _, value := range values[1:] {
		if value > max {
			max = value
		}
	}
	return &max
}

func percentileFloat64(values []float64, percentile float64) *float64 {
	if len(values) == 0 {
		return nil
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	index := int(math.Ceil(percentile*float64(len(sorted)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	value := sorted[index]
	return &value
}

func ratio(part int, total int) float64 {
	if total <= 0 {
		return 0
	}
	return float64(part) / float64(total)
}

func certDaysLeft(notAfter string, now time.Time) (int, bool) {
	parsed, err := time.Parse(time.RFC3339, notAfter)
	if err != nil {
		return 0, false
	}
	return int(parsed.Sub(now).Hours() / 24), true
}
