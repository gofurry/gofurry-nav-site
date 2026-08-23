package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExampleConfigLoadsStrictly(t *testing.T) {
	cfg, err := Load("../../conf/server.example.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != "9980" || cfg.Storage.Path != "./data/uptime.db" {
		t.Fatalf("unexpected example config: %+v", cfg)
	}
	if len(cfg.Uptime.Endpoints) != 5 || cfg.Uptime.Endpoints[0].ID != "nav-web" {
		t.Fatalf("unexpected endpoints: %+v", cfg.Uptime.Endpoints)
	}
}

func TestLoadRejectsUnknownField(t *testing.T) {
	data, err := os.ReadFile("../../conf/server.example.yaml")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "server.yaml")
	if err = os.WriteFile(path, append(data, []byte("\nunknown: true\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err = Load(path); err == nil || !strings.Contains(err.Error(), "field unknown not found") {
		t.Fatalf("Load() error = %v, want unknown field rejection", err)
	}
}

func TestValidateRejectsTimezoneAndEndpoint(t *testing.T) {
	cfg, err := Load("../../conf/server.example.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Uptime.Timezone = "Mars/Olympus"
	if err = cfg.Validate(); err == nil || !strings.Contains(err.Error(), "timezone") {
		t.Fatalf("Validate() error = %v, want timezone error", err)
	}
	cfg.Uptime.Timezone = "Asia/Shanghai"
	cfg.Uptime.Endpoints[0].URL = "file:///tmp/status"
	if err = cfg.Validate(); err == nil || !strings.Contains(err.Error(), "invalid HTTP URL") {
		t.Fatalf("Validate() error = %v, want endpoint URL error", err)
	}
}
