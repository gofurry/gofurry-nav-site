package observation

import (
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-collector/roof/env"
)

func TestSummaryKeys(t *testing.T) {
	if got := TargetLatestKey(ProtocolHTTP, 123, "www.example.com"); got != "collector:v2:latest:http:123:www.example.com" {
		t.Fatalf("TargetLatestKey() = %q", got)
	}
	if got := TargetSummaryKey(123, "www.example.com"); got != "collector:v2:summary:target:123:www.example.com" {
		t.Fatalf("TargetSummaryKey() = %q", got)
	}
	if got := SiteSummaryKey(123); got != "collector:v2:summary:site:123" {
		t.Fatalf("SiteSummaryKey() = %q", got)
	}
	if got := SiteSummaryTargetsKey(123); got != "collector:v2:summary:site_targets:123" {
		t.Fatalf("SiteSummaryTargetsKey() = %q", got)
	}
}

func TestBuildTargetSummaryHTTPHealthyPingFailureWarning(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	docs := map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"tls_handshake": "not_tls",
		}),
		ProtocolPing: latestDoc(ProtocolPing, StatusFailure, now.Add(-time.Minute), nil),
	}

	summary := BuildTargetSummary(1, "example.com", docs, now)
	if summary.Status != StatusWarning {
		t.Fatalf("Status = %q, want warning, reasons=%v", summary.Status, summary.ReasonCodes)
	}
	if !contains(summary.ReasonCodes, "ping_failed_but_http_ok") {
		t.Fatalf("missing ping warning reason: %v", summary.ReasonCodes)
	}
}

func TestBuildTargetSummaryHTTPOnlyHealthyWhenOtherProtocolsDisabled(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	oldV2 := env.GetServerConfig().Collector.V2
	env.GetServerConfig().Collector.V2 = env.CollectorV2Config{}
	t.Cleanup(func() {
		env.GetServerConfig().Collector.V2 = oldV2
	})

	summary := BuildTargetSummary(1, "example.com", map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"tls_handshake": "not_tls",
		}),
	}, now)
	if summary.Status != StatusHealthy {
		t.Fatalf("Status = %q, want healthy, reasons=%v", summary.Status, summary.ReasonCodes)
	}
}

func TestBuildTargetSummaryHTTPAndDNSFailureDown(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	docs := map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusFailure, now.Add(-time.Minute), nil),
		ProtocolDNS:  latestDoc(ProtocolDNS, StatusFailure, now.Add(-time.Minute), nil),
	}

	summary := BuildTargetSummary(1, "example.com", docs, now)
	if summary.Status != StatusDown {
		t.Fatalf("Status = %q, want down, reasons=%v", summary.Status, summary.ReasonCodes)
	}
	if !contains(summary.ReasonCodes, "http_failed") || !contains(summary.ReasonCodes, "dns_failed") {
		t.Fatalf("missing down reasons: %v", summary.ReasonCodes)
	}
}

func TestBuildTargetSummaryStaleHTTPUnknown(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	oldHTTPInterval := env.GetServerConfig().Collector.Request.RequestInterval
	env.GetServerConfig().Collector.Request.RequestInterval = 1
	t.Cleanup(func() {
		env.GetServerConfig().Collector.Request.RequestInterval = oldHTTPInterval
	})
	docs := map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-3*time.Hour), nil),
	}

	summary := BuildTargetSummary(1, "example.com", docs, now)
	if summary.Status != StatusUnknown {
		t.Fatalf("Status = %q, want unknown, reasons=%v", summary.Status, summary.ReasonCodes)
	}
	if !summary.Protocols[ProtocolHTTP].Stale {
		t.Fatal("HTTP protocol summary should be stale")
	}
}

func TestBuildTargetSummaryTLSExpiringWarning(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	docs := map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"tls_handshake":  "collected",
			"cert_verified":  true,
			"cert_not_after": now.Add(15 * 24 * time.Hour).Format(time.RFC3339),
		}),
	}

	summary := BuildTargetSummary(1, "example.com", docs, now)
	if summary.Status != StatusWarning {
		t.Fatalf("Status = %q, want warning, reasons=%v", summary.Status, summary.ReasonCodes)
	}
	if !contains(summary.ReasonCodes, "tls_cert_expiring_soon") {
		t.Fatalf("missing TLS expiring reason: %v", summary.ReasonCodes)
	}
}

func TestBuildTargetSummaryTLSVerifyFailureDegraded(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	docs := map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"tls_handshake":         "collected",
			"cert_verified":         false,
			"verify_error_category": "hostname_mismatch",
		}),
	}

	summary := BuildTargetSummary(1, "example.com", docs, now)
	if summary.Status != StatusDegraded {
		t.Fatalf("Status = %q, want degraded, reasons=%v", summary.Status, summary.ReasonCodes)
	}
	if !contains(summary.ReasonCodes, "tls_verify_hostname_mismatch") {
		t.Fatalf("missing TLS verify reason: %v", summary.ReasonCodes)
	}
}

func TestBuildTargetSummaryDNSRiskWarning(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	docs := map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), nil),
		ProtocolDNS: latestDoc(ProtocolDNS, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"risk_flags": []any{"private_ip"},
		}),
	}

	summary := BuildTargetSummary(1, "example.com", docs, now)
	if summary.Status != StatusWarning {
		t.Fatalf("Status = %q, want warning, reasons=%v", summary.Status, summary.ReasonCodes)
	}
	if !contains(summary.ReasonCodes, "dns_risk_private_ip") {
		t.Fatalf("missing DNS risk reason: %v", summary.ReasonCodes)
	}
}

func TestBuildSiteSummaryAggregatesTargetsConservatively(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	summary := BuildSiteSummary(1, []TargetSummaryDocument{
		{SiteID: 1, Target: "a.example.com", Status: StatusHealthy},
		{SiteID: 1, Target: "b.example.com", Status: StatusDown},
	}, now)

	if summary.Status != StatusDegraded {
		t.Fatalf("Status = %q, want degraded, reasons=%v", summary.Status, summary.ReasonCodes)
	}
	if summary.StatusCounts[StatusHealthy] != 1 || summary.StatusCounts[StatusDown] != 1 {
		t.Fatalf("status counts wrong: %+v", summary.StatusCounts)
	}

	allDown := BuildSiteSummary(1, []TargetSummaryDocument{
		{SiteID: 1, Target: "a.example.com", Status: StatusDown},
		{SiteID: 1, Target: "b.example.com", Status: StatusDown},
	}, now)
	if allDown.Status != StatusDown {
		t.Fatalf("all down status = %q, want down", allDown.Status)
	}
}

func latestDoc(protocol string, status string, observedAt time.Time, payload any) LatestDocument {
	if payload == nil {
		payload = map[string]any{}
	}
	return LatestDocument{
		SiteID:        1,
		Target:        "example.com",
		Protocol:      protocol,
		Status:        status,
		ObservedAt:    observedAt,
		DurationMS:    123,
		Payload:       payload,
		SchemaVersion: schemaVersion,
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
