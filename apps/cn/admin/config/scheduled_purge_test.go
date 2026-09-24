package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func validScheduledEdgeOne() EdgeOneConfig {
	return EdgeOneConfig{Enabled: true, ZoneID: "zone-test", SecretID: "test-id", SecretKey: "test-secret", MainHost: "main.example.test",
		ScheduledPurge: EdgeOneScheduledPurgeConfig{Enabled: true, Time: "05:30", Timezone: "Asia/Shanghai"}}
}

func TestEdgeOneScheduledPurgeConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name    string
		change  func(*EdgeOneConfig)
		invalid bool
	}{
		{"valid", func(*EdgeOneConfig) {}, false},
		{"disabled empty provider", func(c *EdgeOneConfig) { *c = EdgeOneConfig{} }, false},
		{"disabled malformed schedule", func(c *EdgeOneConfig) { c.ScheduledPurge = EdgeOneScheduledPurgeConfig{Time: "bad", Timezone: "bad"} }, false},
		{"hour", func(c *EdgeOneConfig) { c.ScheduledPurge.Time = "24:30" }, true},
		{"minute", func(c *EdgeOneConfig) { c.ScheduledPurge.Time = "05:60" }, true},
		{"short hour", func(c *EdgeOneConfig) { c.ScheduledPurge.Time = "5:30" }, true},
		{"short minute", func(c *EdgeOneConfig) { c.ScheduledPurge.Time = "05:3" }, true},
		{"seconds", func(c *EdgeOneConfig) { c.ScheduledPurge.Time = "05:30:00" }, true},
		{"empty time", func(c *EdgeOneConfig) { c.ScheduledPurge.Time = "" }, true},
		{"cron", func(c *EdgeOneConfig) { c.ScheduledPurge.Time = "30 5 * * *" }, true},
		{"timezone", func(c *EdgeOneConfig) { c.ScheduledPurge.Timezone = "invalid/zone" }, true},
		{"empty timezone", func(c *EdgeOneConfig) { c.ScheduledPurge.Timezone = "" }, true},
		{"host timezone", func(c *EdgeOneConfig) { c.ScheduledPurge.Timezone = "Local" }, true},
		{"empty host", func(c *EdgeOneConfig) { c.MainHost = " " }, true},
		{"wildcard host", func(c *EdgeOneConfig) { c.MainHost = "*.example.test" }, true},
		{"host URL", func(c *EdgeOneConfig) { c.MainHost = "https://main.example.test" }, true},
		{"disabled provider", func(c *EdgeOneConfig) { c.Enabled = false }, true},
		{"empty zone", func(c *EdgeOneConfig) { c.ZoneID = "" }, true},
		{"wildcard zone", func(c *EdgeOneConfig) { c.ZoneID = "*" }, true},
		{"empty ID", func(c *EdgeOneConfig) { c.SecretID = " " }, true},
		{"empty secret", func(c *EdgeOneConfig) { c.SecretKey = "" }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validScheduledEdgeOne()
			tc.change(&cfg)
			spec, location, err := cfg.ScheduledPurgeSchedule()
			if (err != nil) != tc.invalid {
				t.Fatalf("validation error = %v", err)
			}
			if err != nil && (!strings.Contains(err.Error(), "scheduled_purge") || strings.Contains(err.Error(), "test-secret")) {
				t.Fatalf("unsafe or unclear validation: %v", err)
			}
			if !tc.invalid && cfg.ScheduledPurge.Enabled && (spec != "30 5 * * *" || location.String() != "Asia/Shanghai") {
				t.Fatalf("schedule = %q / %v", spec, location)
			}
		})
	}
}

func TestScheduledPurgeYAMLDefaultsAndStartupValidation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.yaml")
	if err := os.WriteFile(path, []byte("server:\n  port: '10099'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var cfg serverConfig
	if err := InitConfig("gofurry-admin", "server.yaml", path, &cfg); err != nil {
		t.Fatal(err)
	}
	want := EdgeOneScheduledPurgeConfig{Enabled: false, Time: "05:30", Timezone: "Asia/Shanghai"}
	if cfg.ExternalServices.CloudOps.EdgeOne.ScheduledPurge != want {
		t.Fatal("omitted schedule must default to disabled 05:30 Asia/Shanghai")
	}
	if err := InitConfig("gofurry-admin", "server.yaml", "server.example.yaml", &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.ExternalServices.CloudOps.EdgeOne.ScheduledPurge != want {
		t.Fatal("example schedule must remain disabled")
	}
	cfg.normalize()
	cfg.ExternalServices.CloudOps.EdgeOne = validScheduledEdgeOne()
	if err := cfg.validate(); err != nil {
		t.Fatal(err)
	}
	cfg.ExternalServices.CloudOps.EdgeOne.ScheduledPurge.Time = "25:30"
	if err := cfg.validate(); err == nil || !strings.Contains(err.Error(), "scheduled_purge") {
		t.Fatalf("startup validation = %v", err)
	}
	if err := os.WriteFile(path, []byte("external_services:\n  cloud_ops:\n    edgeone:\n      scheduled_purge:\n        enabled: true\n        time: '06:15'\n        timezone: UTC\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := InitConfig("gofurry-admin", "server.yaml", path, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.ExternalServices.CloudOps.EdgeOne.ScheduledPurge != (EdgeOneScheduledPurgeConfig{Enabled: true, Time: "06:15", Timezone: "UTC"}) {
		t.Fatal("explicit YAML schedule was not loaded")
	}
}
