package service

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-collector/collector/dns/models"
	"github.com/miekg/dns"
)

func TestLookupGeoASNNilReadersFallsBackToUnknown(t *testing.T) {
	country, city, asn, isp := lookupGeoASN(net.ParseIP("203.0.113.10"), nil, nil, nil)

	if country != "Unknown" || city != "Unknown" || asn != "Unknown" || isp != "Unknown" {
		t.Fatalf("lookupGeoASN() = (%q, %q, %q, %q), want all Unknown", country, city, asn, isp)
	}
}

func TestDNSQueryThreadLimit(t *testing.T) {
	tests := []struct {
		name            string
		configured      int
		recordTypeCount int
		want            int
	}{
		{name: "zero config uses current record type count", configured: 0, recordTypeCount: 8, want: 8},
		{name: "negative config uses current record type count", configured: -1, recordTypeCount: 8, want: 8},
		{name: "positive config limits workers", configured: 3, recordTypeCount: 8, want: 3},
		{name: "too large config does not increase current concurrency", configured: 20, recordTypeCount: 8, want: 8},
		{name: "empty record types stays safe", configured: 0, recordTypeCount: 0, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dnsQueryThreadLimit(tt.configured, tt.recordTypeCount); got != tt.want {
				t.Fatalf("dnsQueryThreadLimit(%d, %d) = %d, want %d", tt.configured, tt.recordTypeCount, got, tt.want)
			}
		})
	}
}

func TestDetectDNSRiskFlags(t *testing.T) {
	msg := &dns.Msg{
		MsgHdr: dns.MsgHdr{Rcode: dns.RcodeNameError},
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 5},
				A:   net.ParseIP("10.0.0.1"),
			},
		},
	}

	risks := detectDNSRiskFlags(net.ParseIP("10.0.0.1"), msg, 5, "")

	assertContainsRisk(t, risks, dnsRiskPrivateIP)
	assertContainsRisk(t, risks, dnsRiskLowTTL)
	assertContainsRisk(t, risks, dnsRiskNXDomainWithAnswer)
	assertContainsRisk(t, risks, dnsRiskPTREmpty)
}

func TestDetectDNSRiskFlagsIncludesIPv6PrivateAndSpecialRanges(t *testing.T) {
	tests := []string{
		"::1",
		"fc00::1",
		"fe80::1",
	}
	for _, ip := range tests {
		t.Run(ip, func(t *testing.T) {
			risks := detectDNSRiskFlags(net.ParseIP(ip), nil, 300, "skip_ptr_check")
			assertContainsRisk(t, risks, dnsRiskPrivateIP)
		})
	}
}

func TestDNSRecordBudgetTracksExhaustion(t *testing.T) {
	budget := newDNSRecordBudget(2)
	if !budget.Take() || !budget.Take() {
		t.Fatal("前两条记录应允许进入预算")
	}
	if budget.Take() {
		t.Fatal("第三条记录应被预算拒绝")
	}
	if !budget.Exhausted() {
		t.Fatal("预算耗尽后应标记 exhausted")
	}
}

func TestBuildDNSObservationPayloadLimitsTextAndRemovesHijacked(t *testing.T) {
	longTXT := "v=spf1 " + strings.Repeat("include:example.com ", 80)
	results := map[string][]models.DNSRecord{
		"A": {
			{
				Type:       "A",
				Value:      "10.0.0.1",
				TTL:        5,
				Duration:   12 * time.Millisecond,
				ReversePTR: "",
				Hijacked:   true,
				RiskFlags:  []string{dnsRiskPrivateIP, dnsRiskLowTTL, dnsRiskPTREmpty},
				Children: []models.DNSRecord{
					{
						Type:  "CNAME",
						Value: "edge.example.com.",
						Children: []models.DNSRecord{
							{Type: "A", Value: "203.0.113.9"},
						},
					},
				},
			},
		},
		"TXT": {
			{
				Type:      "TXT",
				Value:     longTXT,
				TTL:       300,
				Duration:  4 * time.Millisecond,
				RiskFlags: nil,
			},
		},
	}
	metadata := map[string]dnsQueryMetadata{
		"A": {
			ResponseSummary: dnsResponseSummary{
				Rcode:              "NOERROR",
				Authoritative:      true,
				Truncated:          false,
				RecursionAvailable: true,
				AnswerCount:        1,
				AuthorityCount:     0,
				AdditionalCount:    0,
				TTLMin:             5,
				TTLMax:             5,
				TTLAvg:             5,
				DNSSECRRSIGPresent: false,
				DNSSECAD:           false,
			},
		},
		"TXT": {
			ResponseSummary: dnsResponseSummary{
				Rcode:              "NOERROR",
				Authoritative:      false,
				Truncated:          false,
				RecursionAvailable: true,
				AnswerCount:        1,
				AuthorityCount:     0,
				AdditionalCount:    0,
				TTLMin:             300,
				TTLMax:             300,
				TTLAvg:             300,
				DNSSECRRSIGPresent: true,
				DNSSECAD:           true,
			},
		},
	}

	payload := buildDNSObservationPayloadWithMetadata(results, metadata)

	topRisks, ok := payload["risk_flags"].([]string)
	if !ok {
		t.Fatalf("顶层 risk_flags 类型错误: %T", payload["risk_flags"])
	}
	assertContainsRisk(t, topRisks, dnsRiskPrivateIP)
	assertContainsRisk(t, topRisks, dnsRiskLowTTL)
	assertContainsRisk(t, topRisks, dnsRiskPTREmpty)

	aRecords := payload["A"].([]map[string]any)
	if _, ok := aRecords[0]["hijacked"]; ok {
		t.Fatal("v2 DNS payload 不应输出 hijacked 字段")
	}
	aRisks := aRecords[0]["risk_flags"].([]string)
	assertContainsRisk(t, aRisks, dnsRiskPrivateIP)

	txtRecords := payload["TXT"].([]map[string]any)
	txt := txtRecords[0]
	if got := len([]rune(txt["value"].(string))); got != maxDNSObservationTextLength {
		t.Fatalf("TXT value 未限长，got %d", got)
	}
	if txt["value_truncated"] != true {
		t.Fatalf("TXT value_truncated 应为 true，got %v", txt["value_truncated"])
	}
	if txt["value_original_length"] != len([]rune(longTXT)) {
		t.Fatalf("TXT 原始长度错误，got %v", txt["value_original_length"])
	}
	if txt["value_sha256"] != sha256Hex(longTXT) {
		t.Fatalf("TXT sha256 摘要错误，got %v", txt["value_sha256"])
	}
	if txt["text_kind"] != "spf" {
		t.Fatalf("TXT SPF 类型识别错误，got %v", txt["text_kind"])
	}
	responseSummary, ok := payload["response_summary"].(map[string]dnsResponseSummary)
	if !ok {
		t.Fatalf("response_summary 类型错误: %T", payload["response_summary"])
	}
	if responseSummary["A"].Rcode != "NOERROR" || !responseSummary["A"].Authoritative {
		t.Fatalf("A response_summary 错误: %+v", responseSummary["A"])
	}
	if !responseSummary["TXT"].DNSSECRRSIGPresent || !responseSummary["TXT"].DNSSECAD {
		t.Fatalf("TXT DNSSEC 汇总错误: %+v", responseSummary["TXT"])
	}
	if payload["cname_chain_depth"] != 2 {
		t.Fatalf("cname_chain_depth 错误，got %v", payload["cname_chain_depth"])
	}
	if payload["has_a"] != true || payload["has_aaaa"] != false {
		t.Fatalf("A/AAAA 标记错误: has_a=%v has_aaaa=%v", payload["has_a"], payload["has_aaaa"])
	}
	if payload["ipv4_count"] != 2 || payload["ipv6_count"] != 0 {
		t.Fatalf("IP 数量错误: ipv4=%v ipv6=%v", payload["ipv4_count"], payload["ipv6_count"])
	}
	if payload["cname_terminal"] != "edge.example.com." {
		t.Fatalf("cname_terminal 错误，got %v", payload["cname_terminal"])
	}
	if payload["ttl_spread"] != uint32(295) {
		t.Fatalf("ttl_spread 错误，got %v", payload["ttl_spread"])
	}
	if payload["mixed_private_public_ip"] != true {
		t.Fatalf("mixed_private_public_ip 应为 true，got %v", payload["mixed_private_public_ip"])
	}
}

func TestBuildDNSObservationPayloadMarksRecordBudgetExhausted(t *testing.T) {
	payload := buildDNSObservationPayloadWithMetadata(map[string][]models.DNSRecord{
		"A": {{Type: "A", Value: "203.0.113.10"}},
	}, map[string]dnsQueryMetadata{
		"A": {RecordBudgetExhausted: true},
	})

	if payload["record_budget_exhausted"] != true {
		t.Fatalf("record_budget_exhausted = %v, want true", payload["record_budget_exhausted"])
	}
}

func TestResetDNSLookupCachesClearsCachedValues(t *testing.T) {
	geoCache.Store("203.0.113.10", [4]string{"A", "B", "C", "D"})
	ptrCache.Store("203.0.113.10", "ptr.example.")

	resetDNSLookupCaches()

	if _, ok := geoCache.Load("203.0.113.10"); ok {
		t.Fatal("geoCache should be empty after reset")
	}
	if _, ok := ptrCache.Load("203.0.113.10"); ok {
		t.Fatal("ptrCache should be empty after reset")
	}
}

func TestBuildDNSObservationPayloadIdentifiesDMARCAndCAA(t *testing.T) {
	payload := buildDNSObservationPayload(map[string][]models.DNSRecord{
		"TXT": {
			{Type: "TXT", Value: "v=DMARC1; p=none"},
		},
		"CAA": {
			{Type: "CAA", Value: "0 issue letsencrypt.org"},
		},
	})

	txtRecords := payload["TXT"].([]map[string]any)
	if txtRecords[0]["text_kind"] != "dmarc" {
		t.Fatalf("DMARC 类型识别错误，got %v", txtRecords[0]["text_kind"])
	}
	caaRecords := payload["CAA"].([]map[string]any)
	if caaRecords[0]["text_kind"] != "caa" {
		t.Fatalf("CAA 类型识别错误，got %v", caaRecords[0]["text_kind"])
	}
}

func TestBuildDNSObservationPayloadAddsMXAndSOASummary(t *testing.T) {
	payload := buildDNSObservationPayloadWithMetadata(map[string][]models.DNSRecord{
		"MX": {
			{Type: "MX", Value: "mail.example.com. (优先级 10)"},
		},
		"SOA": {
			{Type: "SOA", Value: "ns1.example.com. hostmaster.example.com."},
		},
	}, map[string]dnsQueryMetadata{
		"MX": {
			ResponseSummary: dnsResponseSummary{Rcode: "NOERROR", AnswerCount: 1},
			MXPriorities: []dnsMXPriority{
				{Host: "mail.example.com.", Priority: 10},
			},
		},
		"SOA": {
			ResponseSummary: dnsResponseSummary{Rcode: "NOERROR", AnswerCount: 1},
			SOA: &dnsSOASummary{
				NS:      "ns1.example.com.",
				Mbox:    "hostmaster.example.com.",
				Serial:  2026052401,
				Refresh: 7200,
				Retry:   3600,
				Expire:  1209600,
				Minttl:  300,
			},
		},
	})

	mxPriorities, ok := payload["mx_priorities"].([]dnsMXPriority)
	if !ok {
		t.Fatalf("mx_priorities 类型错误: %T", payload["mx_priorities"])
	}
	if len(mxPriorities) != 1 || mxPriorities[0].Host != "mail.example.com." || mxPriorities[0].Priority != 10 {
		t.Fatalf("mx_priorities 错误: %+v", mxPriorities)
	}
	soa, ok := payload["soa"].(*dnsSOASummary)
	if !ok {
		t.Fatalf("soa 类型错误: %T", payload["soa"])
	}
	if soa.NS != "ns1.example.com." || soa.Mbox != "hostmaster.example.com." || soa.Serial != 2026052401 {
		t.Fatalf("soa 摘要错误: %+v", soa)
	}
	mxHosts := payload["mx_hosts"].([]string)
	if len(mxHosts) != 1 || mxHosts[0] != "mail.example.com." {
		t.Fatalf("mx_hosts 错误: %+v", mxHosts)
	}
}

func TestBuildDNSObservationPayloadAddsNameServerHosts(t *testing.T) {
	payload := buildDNSObservationPayload(map[string][]models.DNSRecord{
		"NS": {
			{Type: "NS", Value: "ns2.example.com.", TTL: 600},
			{Type: "NS", Value: "ns1.example.com.", TTL: 300},
		},
	})

	nameServerHosts, ok := payload["name_server_hosts"].([]string)
	if !ok {
		t.Fatalf("name_server_hosts 类型错误: %T", payload["name_server_hosts"])
	}
	if len(nameServerHosts) != 2 || nameServerHosts[0] != "ns1.example.com." || nameServerHosts[1] != "ns2.example.com." {
		t.Fatalf("name_server_hosts 错误: %+v", nameServerHosts)
	}
	if payload["ttl_spread"] != uint32(300) {
		t.Fatalf("ttl_spread 错误，got %v", payload["ttl_spread"])
	}
}

func assertContainsRisk(t *testing.T, risks []string, expected string) {
	t.Helper()

	for _, risk := range risks {
		if risk == expected {
			return
		}
	}
	t.Fatalf("risk_flags=%v, want %s", risks, expected)
}
