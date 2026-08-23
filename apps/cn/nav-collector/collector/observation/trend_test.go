package observation

import (
	"testing"
	"time"
)

func TestTargetTrendKey(t *testing.T) {
	got := TargetTrendKey(123, "www.example.com")
	want := "collector:v2:trend:target:123:www.example.com"
	if got != want {
		t.Fatalf("TargetTrendKey() = %q, want %q", got, want)
	}
}

func TestReserveDerivedRunDebouncesByKindSiteAndTarget(t *testing.T) {
	resetDerivedRunGateForTest()
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)
	if !reserveDerivedRun("trend", 1, "example.com", now, time.Minute) {
		t.Fatal("first derived run should be reserved")
	}
	if reserveDerivedRun("trend", 1, "example.com", now.Add(30*time.Second), time.Minute) {
		t.Fatal("second derived run within min interval should be skipped")
	}
	if !reserveDerivedRun("change", 1, "example.com", now.Add(30*time.Second), time.Minute) {
		t.Fatal("different derived kind should have an independent debounce key")
	}
	if !reserveDerivedRun("trend", 1, "example.com", now.Add(time.Minute), time.Minute) {
		t.Fatal("derived run after min interval should be reserved")
	}
}

func TestBuildTargetTrendDerivesHTTPPingDNSTLSWindows(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	rows := []ObservationTrendRow{
		trendRow(ProtocolHTTP, StatusSuccess, now.Add(-1*time.Hour), 100, `{
			"response_time_ms": 90,
			"cert_not_after": "2026-06-24T12:00:00Z",
			"cert_issuer": "Issuer B",
			"cert_fingerprint_sha256": "fp-b"
		}`),
		trendRow(ProtocolHTTP, StatusFailure, now.Add(-2*time.Hour), 300, `{
			"response_time_ms": 300,
			"cert_not_after": "2026-06-20T12:00:00Z",
			"cert_issuer": "Issuer A",
			"cert_fingerprint_sha256": "fp-a"
		}`),
		trendRow(ProtocolPing, StatusSuccess, now.Add(-30*time.Minute), 50, `{
			"avg_rtt_ms": 40,
			"loss_rate": 0,
			"jitter_ms": 3
		}`),
		trendRow(ProtocolPing, StatusSuccess, now.Add(-2*time.Hour), 70, `{
			"avg_rtt_ms": 60,
			"loss_rate": 20,
			"jitter_ms": 7
		}`),
		trendRow(ProtocolDNS, StatusSuccess, now.Add(-1*time.Hour), 30, `{
			"risk_flags": ["low_ttl", "ptr_empty"],
			"response_summary": {
				"A": {"ttl_min": 30, "ttl_max": 60, "ttl_avg": 45}
			}
		}`),
		trendRow(ProtocolDNS, StatusSuccess, now.Add(-3*time.Hour), 40, `{
			"risk_flags": ["low_ttl"],
			"response_summary": {
				"A": {"ttl_min": 120, "ttl_max": 240, "ttl_avg": 180}
			}
		}`),
		trendRow(ProtocolHTTP, StatusSuccess, now.Add(-48*time.Hour), 500, `{
			"response_time_ms": 500
		}`),
	}

	doc := BuildTargetTrend(1, "example.com", rows, now)
	window := doc.Windows["24h"]
	if len(window.Protocols) != 3 {
		t.Fatalf("24h protocols = %+v", window.Protocols)
	}

	httpTrend := window.Protocols[ProtocolHTTP]
	if httpTrend.ObservationCount != 2 || httpTrend.SuccessCount != 1 || httpTrend.FailureCount != 1 {
		t.Fatalf("http trend counts wrong: %+v", httpTrend)
	}
	if httpTrend.SuccessRate == nil || *httpTrend.SuccessRate != 0.5 {
		t.Fatalf("http success rate = %v, want 0.5", httpTrend.SuccessRate)
	}
	if httpTrend.HTTP == nil || httpTrend.HTTP.AvgResponseTimeMS == nil || *httpTrend.HTTP.AvgResponseTimeMS != 195 {
		t.Fatalf("http response trend wrong: %+v", httpTrend.HTTP)
	}
	if httpTrend.TLS == nil || !httpTrend.TLS.CertIssuerChanged || !httpTrend.TLS.CertFingerprintChanged {
		t.Fatalf("tls trend should detect issuer/fingerprint changes: %+v", httpTrend.TLS)
	}
	if httpTrend.TLS.LatestCertDaysLeft == nil || *httpTrend.TLS.LatestCertDaysLeft != 30 {
		t.Fatalf("latest cert days left = %+v, want 30", httpTrend.TLS.LatestCertDaysLeft)
	}

	pingTrend := window.Protocols[ProtocolPing]
	if pingTrend.Ping == nil || pingTrend.Ping.AvgRTTMS == nil || *pingTrend.Ping.AvgRTTMS != 50 {
		t.Fatalf("ping trend wrong: %+v", pingTrend.Ping)
	}
	if pingTrend.Ping.AvgLossRate == nil || *pingTrend.Ping.AvgLossRate != 10 {
		t.Fatalf("ping loss trend wrong: %+v", pingTrend.Ping)
	}

	dnsTrend := window.Protocols[ProtocolDNS]
	if dnsTrend.DNS == nil {
		t.Fatal("dns trend should not be nil")
	}
	if dnsTrend.DNS.RiskFlagCounts["low_ttl"] != 2 || dnsTrend.DNS.RiskFlagCounts["ptr_empty"] != 1 {
		t.Fatalf("dns risk counts wrong: %+v", dnsTrend.DNS.RiskFlagCounts)
	}
	if dnsTrend.DNS.LatestTTLMin == nil || *dnsTrend.DNS.LatestTTLMin != 30 {
		t.Fatalf("latest ttl min wrong: %+v", dnsTrend.DNS)
	}
	if dnsTrend.DNS.PreviousTTLAvg == nil || *dnsTrend.DNS.PreviousTTLAvg != 180 {
		t.Fatalf("previous ttl avg wrong: %+v", dnsTrend.DNS)
	}

	if doc.Windows["7d"].Protocols[ProtocolHTTP].ObservationCount != 3 {
		t.Fatalf("7d http count = %d, want 3", doc.Windows["7d"].Protocols[ProtocolHTTP].ObservationCount)
	}
}

func TestBuildTargetTrendIgnoresRowsOutsideWindowAndUnknownProtocols(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	doc := BuildTargetTrend(1, "example.com", []ObservationTrendRow{
		trendRow(ProtocolHTTP, StatusSuccess, now.Add(-8*24*time.Hour), 100, `{}`),
		trendRow(ProtocolPortCheck, StatusSuccess, now.Add(-time.Hour), 100, `{}`),
	}, now)
	if len(doc.Windows["24h"].Protocols) != 0 || len(doc.Windows["7d"].Protocols) != 0 {
		t.Fatalf("unexpected protocols in trend doc: %+v", doc.Windows)
	}
}

func trendRow(protocol string, status string, observedAt time.Time, durationMS int64, payload string) ObservationTrendRow {
	return ObservationTrendRow{
		Protocol:   protocol,
		Status:     status,
		ObservedAt: observedAt,
		DurationMS: durationMS,
		Payload:    payload,
	}
}
