package env

import "testing"

func TestNavV2RouteSwitchesKeepSummaryEnabledCompatibility(t *testing.T) {
	cfg := NavV2Config{SummaryEnabled: true}
	if !cfg.AnyRouteEnabled() || !cfg.SummaryRoutesEnabled() || !cfg.DetailRoutesEnabled() || !cfg.ReadModelRoutesEnabled() {
		t.Fatalf("summary_enabled compatibility gates failed: %+v", cfg)
	}
}

func TestNavV2RouteSwitchesAllowIndependentDisable(t *testing.T) {
	disabled := false
	cfg := NavV2Config{
		Enabled:          boolPtr(true),
		SummaryEnabled:   true,
		DetailEnabled:    &disabled,
		ReadModelEnabled: &disabled,
	}
	if !cfg.SummaryRoutesEnabled() {
		t.Fatalf("summary route should stay enabled: %+v", cfg)
	}
	if cfg.DetailRoutesEnabled() || cfg.ReadModelRoutesEnabled() {
		t.Fatalf("detail/read model routes should be disabled: %+v", cfg)
	}
}

func TestNavV2PayloadDefaults(t *testing.T) {
	cfg := NavV2Config{}
	if cfg.FullPayloadEnabled {
		t.Fatal("full payload should be disabled by default")
	}
	if cfg.PayloadResponseMaxBytesOrDefault() != 2*1024*1024 {
		t.Fatalf("payload response max = %d", cfg.PayloadResponseMaxBytesOrDefault())
	}
}

func TestRedisTimeoutDefault(t *testing.T) {
	if (RedisConfig{}).Timeout().Seconds() != 2 {
		t.Fatalf("redis timeout = %s", (RedisConfig{}).Timeout())
	}
	if (RedisConfig{TimeoutSeconds: 5}).Timeout().Seconds() != 5 {
		t.Fatalf("configured redis timeout = %s", (RedisConfig{TimeoutSeconds: 5}).Timeout())
	}
}

func boolPtr(value bool) *bool {
	return &value
}
