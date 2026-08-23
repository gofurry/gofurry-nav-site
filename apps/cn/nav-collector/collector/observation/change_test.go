package observation

import (
	"testing"
	"time"

	"github.com/bytedance/sonic"
)

func TestTargetChangeKey(t *testing.T) {
	if got := TargetChangeKey(123, "www.example.com"); got != "collector:v2:change:target:123:www.example.com" {
		t.Fatalf("TargetChangeKey() = %q", got)
	}
}

func TestBuildTargetChangesDetectsHTTPAndTLSChanges(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	rows := []ObservationTrendRow{
		changeRow(t, ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"status_code": 200,
			"title":       "New title",
			"server":      "nginx",
			"server_hints": map[string]any{
				"x_powered_by": "Go",
			},
			"final_url": "https://example.com/new",
			"security_headers": map[string]any{
				"strict_transport_security": true,
				"content_security_policy":   true,
			},
			"cert_fingerprint_sha256": "new-fingerprint",
			"cert_issuer":             "New CA",
			"cert_san_count":          4,
			"cert_not_after":          "2026-12-31T00:00:00Z",
		}),
		changeRow(t, ProtocolHTTP, StatusSuccess, now.Add(-time.Hour), map[string]any{
			"status_code": 301,
			"title":       "Old title",
			"server":      "caddy",
			"server_hints": map[string]any{
				"x_powered_by": "PHP",
			},
			"final_url": "https://old.example.com",
			"security_headers": map[string]any{
				"strict_transport_security": true,
			},
			"cert_fingerprint_sha256": "old-fingerprint",
			"cert_issuer":             "Old CA",
			"cert_san_count":          2,
			"cert_not_after":          "2026-06-30T00:00:00Z",
		}),
	}

	doc := BuildTargetChanges(1, "example.com", rows, now)
	for _, field := range []string{
		"status_code",
		"title",
		"server",
		"x_powered_by",
		"final_url",
		"security_headers",
		"cert_fingerprint_sha256",
		"cert_issuer",
		"cert_san_count",
		"cert_not_after",
	} {
		assertChangeField(t, doc.Events, field)
	}
	if len(doc.Events) != 10 {
		t.Fatalf("event count = %d, want 10: %+v", len(doc.Events), doc.Events)
	}
}

func TestBuildTargetChangesDetectsDNSPortAndRDAPChanges(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	rows := []ObservationTrendRow{
		changeRow(t, ProtocolDNS, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"A":                 []any{map[string]any{"value": "203.0.113.2"}},
			"AAAA":              []any{map[string]any{"value": "2001:db8::2"}},
			"cname_terminal":    "cdn-new.example.net.",
			"mx_hosts":          []any{"mx2.example.com."},
			"name_server_hosts": []any{"ns2.example.com."},
			"soa":               map[string]any{"serial": 2026052502},
		}),
		changeRow(t, ProtocolDNS, StatusSuccess, now.Add(-time.Hour), map[string]any{
			"A":                 []any{map[string]any{"value": "203.0.113.1"}},
			"AAAA":              []any{map[string]any{"value": "2001:db8::1"}},
			"cname_terminal":    "cdn-old.example.net.",
			"mx_hosts":          []any{"mx1.example.com."},
			"name_server_hosts": []any{"ns1.example.com."},
			"soa":               map[string]any{"serial": 2026052501},
		}),
		changeRow(t, ProtocolPortCheck, StatusSuccess, now.Add(-2*time.Minute), map[string]any{
			"results": []any{
				map[string]any{"port": 443, "status": "open"},
				map[string]any{"port": 6379, "status": "timeout"},
			},
		}),
		changeRow(t, ProtocolPortCheck, StatusSuccess, now.Add(-2*time.Hour), map[string]any{
			"results": []any{
				map[string]any{"port": 443, "status": "closed"},
				map[string]any{"port": 6379, "status": "closed"},
			},
		}),
		changeRow(t, ProtocolRDAP, StatusSuccess, now.Add(-3*time.Minute), map[string]any{
			"statuses":    []any{"active", "client transfer prohibited"},
			"expires_at":  "2027-05-25T00:00:00Z",
			"nameservers": []any{"ns2.example.com"},
		}),
		changeRow(t, ProtocolRDAP, StatusSuccess, now.Add(-3*time.Hour), map[string]any{
			"statuses":    []any{"active"},
			"expires_at":  "2026-05-25T00:00:00Z",
			"nameservers": []any{"ns1.example.com"},
		}),
	}

	doc := BuildTargetChanges(1, "example.com", rows, now)
	for _, field := range []string{
		"A",
		"AAAA",
		"cname_terminal",
		"mx_hosts",
		"name_server_hosts",
		"soa_serial",
		"port_443",
		"port_6379",
		"statuses",
		"expires_at",
		"nameservers",
	} {
		assertChangeField(t, doc.Events, field)
	}
	if len(doc.Events) != 11 {
		t.Fatalf("event count = %d, want 11: %+v", len(doc.Events), doc.Events)
	}
}

func TestBuildTargetChangesIgnoresFailuresAndSingleObservation(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	rows := []ObservationTrendRow{
		changeRow(t, ProtocolHTTP, StatusFailure, now.Add(-time.Minute), map[string]any{
			"status_code": 500,
			"title":       "Transient failure",
		}),
		changeRow(t, ProtocolHTTP, StatusSuccess, now.Add(-time.Hour), map[string]any{
			"status_code": 200,
			"title":       "Stable title",
		}),
		changeRow(t, ProtocolDNS, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"A": []any{map[string]any{"value": "203.0.113.1"}},
		}),
	}

	doc := BuildTargetChanges(1, "example.com", rows, now)
	if len(doc.Events) != 0 {
		t.Fatalf("events = %+v, want empty", doc.Events)
	}
}

func changeRow(t *testing.T, protocol string, status string, observedAt time.Time, payload map[string]any) ObservationTrendRow {
	t.Helper()
	bytes, err := sonic.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return ObservationTrendRow{
		Protocol:   protocol,
		Status:     status,
		ObservedAt: observedAt,
		DurationMS: 123,
		Payload:    string(bytes),
	}
}

func assertChangeField(t *testing.T, events []ChangeEvent, field string) {
	t.Helper()
	for _, event := range events {
		if event.Field == field {
			return
		}
	}
	t.Fatalf("missing change event field %q in %+v", field, events)
}
