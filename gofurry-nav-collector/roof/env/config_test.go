package env

import (
	"testing"
	"time"

	"gopkg.in/yaml.v2"
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
	for _, protocol := range []string{"ping", "http", "dns", "rdap", "robots", "security_txt", "page_assets", "port_check", "waf_canary"} {
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

func TestEdgeHintsDefaultsEnabled(t *testing.T) {
	cfg := EdgeHintsConfig{}
	if !cfg.EnabledOrDefault() {
		t.Fatal("edge_hints should be enabled by default")
	}
	disabled := false
	cfg.Enabled = &disabled
	if cfg.EnabledOrDefault() {
		t.Fatal("edge_hints explicit false should disable hints")
	}
}

func TestLightProbeDefaults(t *testing.T) {
	cfg := LightProbeConfig{}
	if cfg.RDAP.Enabled || cfg.Robots.Enabled || cfg.SecurityTXT.Enabled || cfg.PageAssets.Enabled || cfg.PortCheck.Enabled || cfg.WAFCanary.Enabled {
		t.Fatal("light probe should be disabled by default")
	}
	if cfg.RDAP.RunOnStart || cfg.Robots.RunOnStart || cfg.SecurityTXT.RunOnStart || cfg.PageAssets.RunOnStart || cfg.PortCheck.RunOnStart || cfg.WAFCanary.RunOnStart {
		t.Fatal("light probe should not run on service start by default")
	}
	if cfg.RDAP.Interval() != 168*time.Hour || cfg.Robots.Interval() != 168*time.Hour || cfg.SecurityTXT.Interval() != 168*time.Hour || cfg.PageAssets.Interval() != 168*time.Hour || cfg.PortCheck.Interval() != 168*time.Hour {
		t.Fatal("light probe interval should default to 168 hours")
	}
	if cfg.RDAP.Timeout() != 10*time.Second || cfg.Robots.Timeout() != 10*time.Second || cfg.SecurityTXT.Timeout() != 10*time.Second || cfg.PageAssets.Timeout() != 10*time.Second {
		t.Fatal("light probe timeout should default to 10 seconds")
	}
	if cfg.WAFCanary.Interval() != 720*time.Hour || cfg.WAFCanary.Timeout() != 10*time.Second || cfg.WAFCanary.Path() != "/.well-known/gofurry-waf-canary" || cfg.WAFCanary.UserAgentOrDefault() != "GoFurry-Nav-Collector-WAF-Canary/1.0" || cfg.WAFCanary.RunOnStart || cfg.WAFCanary.MaxTargets() != 0 {
		t.Fatal("waf_canary defaults are incorrect")
	}
	if got := (LightProbeWAFCanaryConfig{UserAgent: "Custom-Agent"}).UserAgentOrDefault(); got != "Custom-Agent GoFurry-Nav-Collector-WAF-Canary/1.0" {
		t.Fatalf("custom waf_canary user-agent = %q", got)
	}
	if cfg.PortCheck.Timeout() != 2*time.Second || cfg.PortCheck.WorkerCount() != 8 || cfg.PortCheck.MaxPorts() != 24 {
		t.Fatal("port_check defaults are incorrect")
	}
	if cfg.Robots.MaxResponseSize() != 64*1024 || cfg.SecurityTXT.MaxResponseSize() != 64*1024 {
		t.Fatal("light probe max response size should default to 64KiB")
	}
	if cfg.PageAssets.MaxIconSize() != 256*1024 || cfg.PageAssets.MaxManifestSize() != 64*1024 {
		t.Fatal("page_assets max response size defaults are incorrect")
	}
	if cfg.Robots.MaxSitemaps() != 20 {
		t.Fatal("robots max sitemap links should default to 20")
	}
}

func TestLightProbeYAMLConfig(t *testing.T) {
	var cfg serverConfig
	err := yaml.Unmarshal([]byte(`
collector:
  v2:
    enabled: true
    edge_hints:
      enabled: false
    light_probe:
      rdap:
        enabled: true
        interval_hours: 24
        timeout_seconds: 3
        run_on_start: true
      robots:
        enabled: true
        interval_hours: 48
        timeout_seconds: 4
        max_response_bytes: 1234
        max_sitemap_links: 7
        run_on_start: true
      security_txt:
        enabled: true
        interval_hours: 72
        timeout_seconds: 5
        max_response_bytes: 2345
        run_on_start: true
      page_assets:
        enabled: true
        interval_hours: 96
        timeout_seconds: 6
        max_icon_bytes: 3456
        max_manifest_bytes: 4567
        allowed_asset_hosts:
          - cdn.example.com
        run_on_start: true
      port_check:
        enabled: true
        interval_hours: 120
        timeout_seconds: 2
        concurrency: 3
        max_ports_per_target: 4
        ports:
          - 80
          - 443
          - 3306
        run_on_start: true
      waf_canary:
        enabled: true
        interval_hours: 720
        timeout_seconds: 8
        canary_path: "/waf-canary"
        user_agent: "Custom-WAF-Canary/1.0"
        run_on_start: true
        max_targets_per_run: 12
`), &cfg)
	if err != nil {
		t.Fatalf("yaml.Unmarshal() error = %v", err)
	}
	if !cfg.Collector.V2.ProtocolEnabled("rdap") || !cfg.Collector.V2.ProtocolEnabled("robots") || !cfg.Collector.V2.ProtocolEnabled("security_txt") || !cfg.Collector.V2.ProtocolEnabled("page_assets") || !cfg.Collector.V2.ProtocolEnabled("port_check") || !cfg.Collector.V2.ProtocolEnabled("waf_canary") {
		t.Fatalf("light probe protocol switches not loaded: %+v", cfg.Collector.V2.LightProbe)
	}
	if cfg.Collector.V2.EdgeHints.EnabledOrDefault() {
		t.Fatal("edge_hints explicit false should be loaded")
	}
	if cfg.Collector.V2.LightProbe.RDAP.Interval() != 24*time.Hour || cfg.Collector.V2.LightProbe.RDAP.Timeout() != 3*time.Second {
		t.Fatalf("rdap config not loaded: %+v", cfg.Collector.V2.LightProbe.RDAP)
	}
	if !cfg.Collector.V2.LightProbe.RDAP.RunOnStart || !cfg.Collector.V2.LightProbe.Robots.RunOnStart || !cfg.Collector.V2.LightProbe.SecurityTXT.RunOnStart || !cfg.Collector.V2.LightProbe.PageAssets.RunOnStart || !cfg.Collector.V2.LightProbe.PortCheck.RunOnStart {
		t.Fatalf("light probe run_on_start config not loaded: %+v", cfg.Collector.V2.LightProbe)
	}
	if cfg.Collector.V2.LightProbe.Robots.MaxResponseSize() != 1234 || cfg.Collector.V2.LightProbe.Robots.MaxSitemaps() != 7 {
		t.Fatalf("robots limits not loaded: %+v", cfg.Collector.V2.LightProbe.Robots)
	}
	if cfg.Collector.V2.LightProbe.SecurityTXT.MaxResponseSize() != 2345 {
		t.Fatalf("security_txt limits not loaded: %+v", cfg.Collector.V2.LightProbe.SecurityTXT)
	}
	if cfg.Collector.V2.LightProbe.PageAssets.Interval() != 96*time.Hour || cfg.Collector.V2.LightProbe.PageAssets.Timeout() != 6*time.Second {
		t.Fatalf("page_assets timing config not loaded: %+v", cfg.Collector.V2.LightProbe.PageAssets)
	}
	if cfg.Collector.V2.LightProbe.PageAssets.MaxIconSize() != 3456 || cfg.Collector.V2.LightProbe.PageAssets.MaxManifestSize() != 4567 {
		t.Fatalf("page_assets size limits not loaded: %+v", cfg.Collector.V2.LightProbe.PageAssets)
	}
	if len(cfg.Collector.V2.LightProbe.PageAssets.AllowedAssetHosts) != 1 || cfg.Collector.V2.LightProbe.PageAssets.AllowedAssetHosts[0] != "cdn.example.com" {
		t.Fatalf("page_assets allowed hosts not loaded: %+v", cfg.Collector.V2.LightProbe.PageAssets.AllowedAssetHosts)
	}
	if cfg.Collector.V2.LightProbe.PortCheck.Interval() != 120*time.Hour || cfg.Collector.V2.LightProbe.PortCheck.Timeout() != 2*time.Second {
		t.Fatalf("port_check timing config not loaded: %+v", cfg.Collector.V2.LightProbe.PortCheck)
	}
	if cfg.Collector.V2.LightProbe.PortCheck.WorkerCount() != 3 || cfg.Collector.V2.LightProbe.PortCheck.MaxPorts() != 4 {
		t.Fatalf("port_check limits not loaded: %+v", cfg.Collector.V2.LightProbe.PortCheck)
	}
	if len(cfg.Collector.V2.LightProbe.PortCheck.Ports) != 3 || cfg.Collector.V2.LightProbe.PortCheck.Ports[2] != 3306 {
		t.Fatalf("port_check ports not loaded: %+v", cfg.Collector.V2.LightProbe.PortCheck.Ports)
	}
	if cfg.Collector.V2.LightProbe.WAFCanary.Interval() != 720*time.Hour || cfg.Collector.V2.LightProbe.WAFCanary.Timeout() != 8*time.Second {
		t.Fatalf("waf_canary timing config not loaded: %+v", cfg.Collector.V2.LightProbe.WAFCanary)
	}
	if cfg.Collector.V2.LightProbe.WAFCanary.Path() != "/waf-canary" || cfg.Collector.V2.LightProbe.WAFCanary.UserAgentOrDefault() != "Custom-WAF-Canary/1.0" {
		t.Fatalf("waf_canary path/user-agent config not loaded: %+v", cfg.Collector.V2.LightProbe.WAFCanary)
	}
	if !cfg.Collector.V2.LightProbe.WAFCanary.RunOnStart || cfg.Collector.V2.LightProbe.WAFCanary.MaxTargets() != 12 {
		t.Fatalf("waf_canary run control config not loaded: %+v", cfg.Collector.V2.LightProbe.WAFCanary)
	}
}

func TestV2DerivedDefaultsAndYAMLConfig(t *testing.T) {
	cfg := V2DerivedConfig{}
	if !cfg.TrendEnabledOrDefault() || !cfg.ChangeEnabledOrDefault() {
		t.Fatal("v2 derived outputs should be enabled by default")
	}
	if cfg.MinInterval() != 300*time.Second || cfg.QueryTimeout() != 2*time.Second || cfg.TrendRows() != 3000 || cfg.ChangeRows() != 80 {
		t.Fatalf("unexpected derived defaults: %+v", cfg)
	}

	var loaded serverConfig
	err := yaml.Unmarshal([]byte(`
collector:
  v2:
    derived:
      trend_enabled: false
      change_enabled: false
      min_interval_seconds: 60
      query_timeout_seconds: 3
      trend_max_rows: 100
      change_max_rows: 20
`), &loaded)
	if err != nil {
		t.Fatalf("yaml.Unmarshal() error = %v", err)
	}
	derived := loaded.Collector.V2.Derived
	if derived.TrendEnabledOrDefault() || derived.ChangeEnabledOrDefault() {
		t.Fatal("derived explicit false switches should be loaded")
	}
	if derived.MinInterval() != time.Minute || derived.QueryTimeout() != 3*time.Second || derived.TrendRows() != 100 || derived.ChangeRows() != 20 {
		t.Fatalf("derived limits not loaded: %+v", derived)
	}
}

func TestSchedulerDefaults(t *testing.T) {
	cfg := SchedulerConfig{}
	if cfg.LeaseEnabled {
		t.Fatal("Scheduler lease should be disabled by default")
	}
	if !cfg.RunStateEnabled() {
		t.Fatal("Scheduler run state should be enabled by default")
	}
	if got := cfg.RunStateTTL(); got != 168*time.Hour {
		t.Fatalf("RunStateTTL() = %s, want 168h", got)
	}
}

func TestSchedulerYAMLConfig(t *testing.T) {
	var cfg serverConfig
	err := yaml.Unmarshal([]byte(`
collector:
  scheduler:
    collector_id: "collector-a"
    lease_enabled: true
    lease_ttl_seconds: 90
    run_state_redis: false
    run_state_ttl_hours: 24
`), &cfg)
	if err != nil {
		t.Fatalf("yaml.Unmarshal() error = %v", err)
	}
	if cfg.Collector.Scheduler.CollectorID != "collector-a" || !cfg.Collector.Scheduler.LeaseEnabled {
		t.Fatalf("Scheduler config not loaded: %+v", cfg.Collector.Scheduler)
	}
	if cfg.Collector.Scheduler.RunStateEnabled() {
		t.Fatal("RunStateEnabled() should follow explicit false")
	}
	if got := cfg.Collector.Scheduler.RunStateTTL(); got != 24*time.Hour {
		t.Fatalf("RunStateTTL() = %s, want 24h", got)
	}
}

func TestServerModeYAMLTag(t *testing.T) {
	var cfg serverConfig
	err := yaml.Unmarshal([]byte(`
server:
  app_name: "GF-Nav-Collector"
  app_version: "v1.0.0"
  mode: "debug"
  memory_limit: 1
`), &cfg)
	if err != nil {
		t.Fatalf("yaml.Unmarshal() error = %v", err)
	}
	if cfg.Server.Mode != "debug" {
		t.Fatalf("Server.Mode = %q, want debug", cfg.Server.Mode)
	}
}
