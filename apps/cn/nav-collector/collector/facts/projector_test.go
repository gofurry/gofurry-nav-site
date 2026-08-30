package facts

import (
	"encoding/json"
	"testing"
)

func TestProjectKnownStateWhitelistsHTTP(t *testing.T) {
	payload := []byte(`{"status_code":200,"response_time_ms":12,"headers":{"secret":"x"},"body":"private","tls_version":"TLS1.3","cert_verified":true,"cert_not_before":"","cert_not_after":"2026-09-01T00:00:00Z"}`)
	projected, err := projectKnownState("http", payload)
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]any
	if err := json.Unmarshal(projected, &state); err != nil {
		t.Fatal(err)
	}
	if _, ok := state["body"]; ok {
		t.Fatal("raw body escaped whitelist")
	}
	if _, ok := state["headers"]; ok {
		t.Fatal("raw headers escaped whitelist")
	}
	if state["status_code"] != float64(200) {
		t.Fatalf("status_code = %v", state["status_code"])
	}
	tls, ok := state["tls"].(map[string]any)
	if !ok {
		t.Fatalf("tls projection = %#v", state["tls"])
	}
	if _, ok := tls["cert_not_before"]; ok {
		t.Fatalf("empty cert_not_before escaped normalization: %#v", tls)
	}
	if tls["cert_not_after"] != "2026-09-01T00:00:00Z" {
		t.Fatalf("valid cert_not_after was not preserved: %#v", tls)
	}
}

func TestProjectKnownStateDropsMalformedHTTPTimestamps(t *testing.T) {
	projected, err := projectKnownState("http", []byte(`{"cert_not_before":"not-a-time","cert_not_after":null}`))
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]any
	if err := json.Unmarshal(projected, &state); err != nil {
		t.Fatal(err)
	}
	if _, ok := state["tls"]; ok {
		t.Fatalf("invalid-only TLS state should be omitted: %#v", state)
	}
}

func TestProjectKnownStateNormalizesDNS(t *testing.T) {
	payload := []byte(`{"has_a":true,"has_aaaa":false,"aaaa_evidence":"unavailable","ipv4_count":2,"A":[{"value":"203.0.113.2","ttl":60},{"value":"203.0.113.1","ttl":30}],"TXT":[{"value":"secret"}],"name_server_hosts":["ns1.example"],"cname_chain_depth":1}`)
	projected, err := projectKnownState("dns", payload)
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]any
	if err := json.Unmarshal(projected, &state); err != nil {
		t.Fatal(err)
	}
	if _, ok := state["TXT"]; ok {
		t.Fatal("TXT raw record escaped whitelist")
	}
	if state["min_ttl"] != float64(30) || state["max_ttl"] != float64(60) {
		t.Fatalf("TTL bounds = %#v", state)
	}
	if state["aaaa_evidence"] != "unavailable" {
		t.Fatalf("AAAA evidence = %#v", state["aaaa_evidence"])
	}
	records := state["a_records"].([]any)
	if records[0] != "203.0.113.1" {
		t.Fatalf("a_records = %#v", records)
	}
}

func TestProjectKnownStatePreservesSecurityTXTRecognition(t *testing.T) {
	projected, err := projectKnownState("security_txt", []byte(`{"exists":true,"recognition":"present_invalid","validation_errors":["content_type_not_text_plain"],"content_type":"text/html","body":"private"}`))
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]any
	if err := json.Unmarshal(projected, &state); err != nil {
		t.Fatal(err)
	}
	if state["recognition"] != "present_invalid" || state["content_type"] != "text/html" {
		t.Fatalf("security.txt state = %#v", state)
	}
	if _, ok := state["body"]; ok {
		t.Fatalf("raw security.txt body escaped whitelist: %#v", state)
	}
}
