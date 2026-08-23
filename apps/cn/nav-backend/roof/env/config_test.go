package env

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestExampleConfigDecodesWithYAMLV3(t *testing.T) {
	data, err := os.ReadFile("../../conf/server.example.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var cfg serverConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("yaml.Unmarshal() error = %v", err)
	}
	if cfg.Server.Mode != "debug" || cfg.Server.Port != "9999" || cfg.DataBase.DBName != "gfn" {
		t.Fatalf("example config fields changed semantics: %+v", cfg)
	}
	if cfg.Middleware.Cors.AllowOrigins == "" || !cfg.Middleware.Limiter.IsOn {
		t.Fatalf("middleware config fields were not decoded: %+v", cfg.Middleware)
	}
}

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

func TestWafDirectivesFilesPreserveExplicitRules(t *testing.T) {
	explicit := []string{"./conf/coraza.conf", "./conf/custom-rules.conf"}
	resolved := (WafConfig{DirectivesFiles: explicit}).ResolveDirectivesFiles()
	if len(resolved) != 2 || resolved[0] != explicit[0] || resolved[1] != explicit[1] {
		t.Fatalf("resolved directives = %#v, want %#v", resolved, explicit)
	}
	resolved[0] = "changed"
	if explicit[0] == "changed" {
		t.Fatal("ResolveDirectivesFiles returned the caller's mutable slice")
	}
}

func TestWafDirectivesFilesDefaultToConfPath(t *testing.T) {
	if files := (WafConfig{}).ResolveDirectivesFiles(); len(files) != 1 || files[0] != "./conf/coraza.conf" {
		t.Fatalf("default directives = %#v", files)
	}
	if files := (WafConfig{ConfPath: "./conf/site.conf"}).ResolveDirectivesFiles(); len(files) != 1 || files[0] != "./conf/site.conf" {
		t.Fatalf("configured directives = %#v", files)
	}
}

func boolPtr(value bool) *bool {
	return &value
}

func TestExplicitViperConfigDoesNotEnableAutomaticEnv(t *testing.T) {
	previous := configuration
	t.Cleanup(func() { configuration = previous })
	t.Setenv("APP_SERVER_PORT", "1")
	if err := LoadServerConfig("../../conf/server.example.yaml"); err != nil {
		t.Fatal(err)
	}
	if configuration.Server.Port != "9999" {
		t.Fatalf("server port = %q, want file value", configuration.Server.Port)
	}
}
