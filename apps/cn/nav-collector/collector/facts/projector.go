package facts

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// projectKnownState is the code-level whitelist boundary between raw probe
// payloads and durable historical state. Unknown keys are always discarded.
func projectKnownState(protocol string, payload []byte) ([]byte, error) {
	if len(payload) == 0 {
		return nil, nil
	}
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("decode %s known-state payload: %w", protocol, err)
	}
	var projected map[string]any
	switch protocol {
	case "http":
		projected = pick(raw, "status_code", "response_time_ms", "ttfb_ms", "http_protocol", "server", "remote_ip", "final_url", "canonical_url", "security_headers")
		tls := renamePick(raw, map[string]string{
			"tls_handshake": "handshake", "tls_version": "version",
			"cipher_suite": "cipher_suite", "cert_verified": "cert_verified",
			"verify_error_category": "verify_error_category",
			"cert_not_before":       "cert_not_before", "cert_not_after": "cert_not_after",
			"cert_issuer": "cert_issuer", "fingerprint_sha256": "fingerprint_sha256",
			"spki_sha256": "spki_sha256", "cert_dns_names": "cert_dns_names",
		})
		dropInvalidRFC3339Fields(tls, "cert_not_before", "cert_not_after")
		if len(tls) > 0 {
			projected["tls"] = tls
		}
	case "dns":
		projected = renamePick(raw, map[string]string{
			"has_a": "has_a", "has_aaaa": "has_aaaa", "ipv4_count": "ipv4_count",
			"ipv6_count": "ipv6_count", "cname_terminal": "cname_terminal",
			"cname_chain_depth": "cname_depth", "name_server_hosts": "ns_hosts",
			"mx_hosts": "mx_hosts", "risk_flags": "risk_flags", "aaaa_evidence": "aaaa_evidence",
		})
		projected["a_records"] = dnsRecordValues(raw["A"])
		projected["aaaa_records"] = dnsRecordValues(raw["AAAA"])
		if minimum, maximum, ok := dnsTTLBounds(raw); ok {
			projected["min_ttl"], projected["max_ttl"] = minimum, maximum
		}
	case "ping":
		projected = pick(raw, "icmp_status", "avg_rtt_ms", "min_rtt_ms", "max_rtt_ms", "loss_rate", "jitter_ms", "selected_ip", "ip_family", "icmp_blocked_suspected")
	case "rdap":
		projected = pick(raw, "registrable_domain", "registrar", "statuses", "expires_at", "nameservers")
	case "robots":
		projected = pick(raw, "exists", "status_code", "sitemap_count", "global_disallow_all")
	case "security_txt":
		projected = pick(raw, "exists", "recognition", "validation_errors", "path_used", "status_code", "content_type", "contact", "expires", "policy", "canonical")
	case "llms_txt":
		projected = pick(raw, "exists", "status_code", "title", "heading_count", "link_count", "optional_section_present")
	case "page_assets":
		projected = map[string]any{}
		for _, key := range []string{"icon", "manifest"} {
			if nested, ok := raw[key].(map[string]any); ok {
				projected[key] = pick(nested, "exists", "status_code", "url", "content_type", "sha256", "body_truncated")
			}
		}
	case "port_check":
		projected = pick(raw, "ports_checked", "open_count")
		projected["results"] = compactObjects(raw["results"], "port", "open", "error_category")
	case "waf_canary":
		projected = pick(raw, "cases_total", "blocked_count", "unexpected_pass_count")
		projected["cases"] = compactObjects(raw["cases"], "name", "status_code", "blocked", "unexpected_pass")
	default:
		return nil, fmt.Errorf("unsupported known-state protocol %q", protocol)
	}
	return json.Marshal(projected)
}

func dropInvalidRFC3339Fields(values map[string]any, keys ...string) {
	for _, key := range keys {
		value, ok := values[key].(string)
		if !ok {
			delete(values, key)
			continue
		}
		if _, err := time.Parse(time.RFC3339, strings.TrimSpace(value)); err != nil {
			delete(values, key)
		}
	}
}

func pick(source map[string]any, keys ...string) map[string]any {
	result := make(map[string]any, len(keys))
	for _, key := range keys {
		if value, ok := source[key]; ok {
			result[key] = value
		}
	}
	return result
}

func renamePick(source map[string]any, keys map[string]string) map[string]any {
	result := make(map[string]any, len(keys))
	for sourceKey, targetKey := range keys {
		if value, ok := source[sourceKey]; ok {
			result[targetKey] = value
		}
	}
	return result
}

func compactObjects(value any, keys ...string) []map[string]any {
	items, ok := value.([]any)
	if !ok {
		return []map[string]any{}
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if object, ok := item.(map[string]any); ok {
			result = append(result, pick(object, keys...))
		}
	}
	return result
}

func dnsRecordValues(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return []string{}
	}
	seen := make(map[string]struct{})
	for _, item := range items {
		if record, ok := item.(map[string]any); ok {
			if text, ok := record["value"].(string); ok && text != "" {
				seen[text] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func dnsTTLBounds(raw map[string]any) (float64, float64, bool) {
	minimum, maximum, found := float64(0), float64(0), false
	for _, recordType := range []string{"A", "AAAA", "CNAME", "NS", "MX"} {
		items, _ := raw[recordType].([]any)
		for _, item := range items {
			record, _ := item.(map[string]any)
			ttl, ok := record["ttl"].(float64)
			if !ok || ttl <= 0 {
				continue
			}
			if !found || ttl < minimum {
				minimum = ttl
			}
			if !found || ttl > maximum {
				maximum = ttl
			}
			found = true
		}
	}
	return minimum, maximum, found
}
