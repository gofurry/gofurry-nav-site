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

func boolPtr(value bool) *bool {
	return &value
}
