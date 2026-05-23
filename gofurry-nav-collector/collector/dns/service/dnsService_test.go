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

	payload := buildDNSObservationPayload(results)

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

func assertContainsRisk(t *testing.T, risks []string, expected string) {
	t.Helper()

	for _, risk := range risks {
		if risk == expected {
			return
		}
	}
	t.Fatalf("risk_flags=%v, want %s", risks, expected)
}
