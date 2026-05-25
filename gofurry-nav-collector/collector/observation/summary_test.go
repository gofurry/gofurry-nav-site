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

func TestTargetFromSummaryKey(t *testing.T) {
	key := TargetSummaryKey(123, "www.example.com")
	if got := targetFromSummaryKey(123, key); got != "www.example.com" {
		t.Fatalf("targetFromSummaryKey() = %q", got)
	}
	if got := targetFromSummaryKey(456, key); got != "" {
		t.Fatalf("targetFromSummaryKey(site mismatch) = %q, want empty", got)
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

func TestBuildTargetSummaryUnknownDNSRiskFallsBackToOther(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	docs := map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), nil),
		ProtocolDNS: latestDoc(ProtocolDNS, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"risk_flags": []any{"future_risk"},
		}),
	}

	summary := BuildTargetSummary(1, "example.com", docs, now)
	if !contains(summary.ReasonCodes, "dns_risk_other") {
		t.Fatalf("missing fallback DNS risk reason: %v", summary.ReasonCodes)
	}
}

func TestBuildTargetSummaryUnknownTLSVerifyCategoryFallsBackToOther(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	docs := map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"tls_handshake":         "collected",
			"cert_verified":         false,
			"verify_error_category": "new_tls_error",
		}),
	}

	summary := BuildTargetSummary(1, "example.com", docs, now)
	if !contains(summary.ReasonCodes, "tls_verify_other") {
		t.Fatalf("missing fallback TLS reason: %v", summary.ReasonCodes)
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

func TestReasonDefinitionsAreStableAndComplete(t *testing.T) {
	definitions := AllReasonDefinitions()
	if len(definitions) == 0 {
		t.Fatal("reason definitions should not be empty")
	}
	seen := map[string]bool{}
	for _, definition := range definitions {
		if definition.Code == "" || definition.MessageZH == "" || definition.DescriptionZH == "" {
			t.Fatalf("reason definition has empty required field: %+v", definition)
		}
		if seen[definition.Code] {
			t.Fatalf("duplicate reason definition code: %s", definition.Code)
		}
		seen[definition.Code] = true
		if !validReasonSeverity(definition.Severity) {
			t.Fatalf("invalid reason severity: %+v", definition)
		}
		if definition.Scope != ReasonScopeTarget && definition.Scope != ReasonScopeSite {
			t.Fatalf("invalid reason scope: %+v", definition)
		}
		if _, ok := ReasonDefinitionByCode(definition.Code); !ok {
			t.Fatalf("ReasonDefinitionByCode(%q) missing", definition.Code)
		}
	}
	for _, code := range []string{
		"http_missing_or_stale",
		"http_failed",
		"dns_failed",
		"dns_missing_or_stale",
		"dns_failed_but_http_ok",
		"ping_failed_but_http_ok",
		"dns_risk_private_ip",
		"dns_risk_low_ttl",
		"dns_risk_nxdomain_with_answer",
		"dns_risk_ptr_empty",
		"dns_risk_other",
		"tls_verify_expired",
		"tls_verify_not_yet_valid",
		"tls_verify_hostname_mismatch",
		"tls_verify_unknown_authority",
		"tls_verify_incompatible_usage",
		"tls_verify_other",
		"tls_cert_expired",
		"tls_cert_expiring_soon",
		"no_target_summary",
		"all_targets_down",
		"all_targets_unknown",
		"some_targets_degraded",
		"some_targets_warning",
	} {
		if _, ok := ReasonDefinitionByCode(code); !ok {
			t.Fatalf("expected reason definition %q", code)
		}
	}
}

func TestSummaryReasonCodesHaveDefinitions(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	targetSummaries := []TargetSummaryDocument{
		BuildTargetSummary(1, "missing-http.example.com", map[string]LatestDocument{}, now),
		BuildTargetSummary(1, "http-dns-failed.example.com", map[string]LatestDocument{
			ProtocolHTTP: latestDoc(ProtocolHTTP, StatusFailure, now.Add(-time.Minute), nil),
			ProtocolDNS:  latestDoc(ProtocolDNS, StatusFailure, now.Add(-time.Minute), nil),
		}, now),
		BuildTargetSummary(1, "warning.example.com", map[string]LatestDocument{
			ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
				"tls_handshake":         "collected",
				"cert_verified":         false,
				"verify_error_category": "unknown_authority",
				"cert_not_after":        now.Add(15 * 24 * time.Hour).Format(time.RFC3339),
			}),
			ProtocolDNS: latestDoc(ProtocolDNS, StatusSuccess, now.Add(-time.Minute), map[string]any{
				"risk_flags": []any{"ptr_empty"},
			}),
			ProtocolPing: latestDoc(ProtocolPing, StatusFailure, now.Add(-time.Minute), nil),
		}, now),
	}
	for _, summary := range targetSummaries {
		assertReasonCodesDefined(t, summary.ReasonCodes)
	}

	siteSummary := BuildSiteSummary(1, targetSummaries, now)
	assertReasonCodesDefined(t, siteSummary.ReasonCodes)
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

func validReasonSeverity(value string) bool {
	switch value {
	case ReasonSeverityInfo, ReasonSeverityWarning, ReasonSeverityDegraded, ReasonSeverityDown, ReasonSeverityUnknown:
		return true
	default:
		return false
	}
}

func assertReasonCodesDefined(t *testing.T, codes []string) {
	t.Helper()
	for _, code := range codes {
		if _, ok := ReasonDefinitionByCode(code); !ok {
			t.Fatalf("reason code %q has no definition", code)
		}
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
