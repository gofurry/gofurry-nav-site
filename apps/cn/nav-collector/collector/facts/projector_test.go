package facts

import (
	"encoding/json"
	"testing"
)

func TestProjectKnownStateWhitelistsHTTP(t *testing.T) {
	payload := []byte(`{"status_code":200,"response_time_ms":12,"headers":{"secret":"x"},"body":"private","tls_version":"TLS1.3","cert_verified":true}`)
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
	if _, ok := state["tls"].(map[string]any); !ok {
		t.Fatalf("tls projection = %#v", state["tls"])
	}
}

func TestProjectKnownStateNormalizesDNS(t *testing.T) {
	payload := []byte(`{"has_a":true,"has_aaaa":false,"ipv4_count":2,"A":[{"value":"203.0.113.2","ttl":60},{"value":"203.0.113.1","ttl":30}],"TXT":[{"value":"secret"}],"name_server_hosts":["ns1.example"],"cname_chain_depth":1}`)
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
	records := state["a_records"].([]any)
	if records[0] != "203.0.113.1" {
		t.Fatalf("a_records = %#v", records)
	}
}
