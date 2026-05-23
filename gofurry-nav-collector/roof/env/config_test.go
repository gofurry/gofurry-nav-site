package env

import (
	"testing"
	"time"
)

func TestProbeBudgetDefaults(t *testing.T) {
	cfg := ProbeBudgetConfig{}

	if got := cfg.RedisTimeout(); got != 2*time.Second {
		t.Fatalf("RedisTimeout() = %s, want 2s", got)
	}
	if got := cfg.HTTPTimeout(); got != 25*time.Second {
		t.Fatalf("HTTPTimeout() = %s, want 25s", got)
	}
	if got := cfg.MaxHTTPRedirects(); got != 5 {
		t.Fatalf("MaxHTTPRedirects() = %d, want 5", got)
	}
	if got := cfg.TLSHandshakeTimeout(); got != 5*time.Second {
		t.Fatalf("TLSHandshakeTimeout() = %s, want 5s", got)
	}
	if got := cfg.DNSTimeout(); got != 5*time.Second {
		t.Fatalf("DNSTimeout() = %s, want 5s", got)
	}
	if got := cfg.PTRTimeout(); got != 2*time.Second {
		t.Fatalf("PTRTimeout() = %s, want 2s", got)
	}
	if got := cfg.MaxDNSRecords(); got != 64 {
		t.Fatalf("MaxDNSRecords() = %d, want 64", got)
	}
	if got := cfg.MaxHTTPResponseBytes(); got != 1024*1024 {
		t.Fatalf("MaxHTTPResponseBytes() = %d, want 1048576", got)
	}
}

func TestProbeBudgetOverrides(t *testing.T) {
	cfg := ProbeBudgetConfig{
		RedisTimeoutSeconds:        3,
		HTTPTimeoutSeconds:         9,
		HTTPRedirects:              4,
		TLSHandshakeTimeoutSeconds: 6,
		DNSTimeoutSeconds:          7,
		PTRTimeoutSeconds:          8,
		MaxDNSRecordsPerQuery:      16,
		MaxResponseBytes:           512,
	}

	if got := cfg.RedisTimeout(); got != 3*time.Second {
		t.Fatalf("RedisTimeout() = %s, want 3s", got)
	}
	if got := cfg.HTTPTimeout(); got != 9*time.Second {
		t.Fatalf("HTTPTimeout() = %s, want 9s", got)
	}
	if got := cfg.MaxHTTPRedirects(); got != 4 {
		t.Fatalf("MaxHTTPRedirects() = %d, want 4", got)
	}
	if got := cfg.TLSHandshakeTimeout(); got != 6*time.Second {
		t.Fatalf("TLSHandshakeTimeout() = %s, want 6s", got)
	}
	if got := cfg.DNSTimeout(); got != 7*time.Second {
		t.Fatalf("DNSTimeout() = %s, want 7s", got)
	}
	if got := cfg.PTRTimeout(); got != 8*time.Second {
		t.Fatalf("PTRTimeout() = %s, want 8s", got)
	}
	if got := cfg.MaxDNSRecords(); got != 16 {
		t.Fatalf("MaxDNSRecords() = %d, want 16", got)
	}
	if got := cfg.MaxHTTPResponseBytes(); got != 512 {
		t.Fatalf("MaxHTTPResponseBytes() = %d, want 512", got)
	}
}

func TestCollectorV2DefaultsDisabled(t *testing.T) {
	cfg := CollectorV2Config{}
	for _, protocol := range []string{"ping", "http", "dns"} {
		if cfg.ProtocolEnabled(protocol) {
			t.Fatalf("ProtocolEnabled(%q) should be false by default", protocol)
		}
		if cfg.ObservationEnabled(protocol) {
			t.Fatalf("ObservationEnabled(%q) should be false by default", protocol)
		}
		if cfg.LatestRedisEnabled(protocol) {
			t.Fatalf("LatestRedisEnabled(%q) should be false by default", protocol)
		}
	}
}

func TestCollectorV2ProtocolSwitches(t *testing.T) {
	cfg := CollectorV2Config{
		Enabled:       true,
		ObservationDB: true,
		LatestRedis:   true,
		Protocols: CollectorProtocols{
			Ping: true,
			HTTP: false,
			DNS:  true,
		},
	}
	if !cfg.ProtocolEnabled("ping") || !cfg.ObservationEnabled("ping") || !cfg.LatestRedisEnabled("ping") {
		t.Fatal("ping should be fully enabled")
	}
	if cfg.ProtocolEnabled("http") || cfg.ObservationEnabled("http") || cfg.LatestRedisEnabled("http") {
		t.Fatal("http should stay disabled by protocol switch")
	}
	if !cfg.ProtocolEnabled("dns") || !cfg.ObservationEnabled("dns") || !cfg.LatestRedisEnabled("dns") {
		t.Fatal("dns should be fully enabled")
	}
	if cfg.ProtocolEnabled("unknown") {
		t.Fatal("unknown protocol should be disabled")
	}
}
