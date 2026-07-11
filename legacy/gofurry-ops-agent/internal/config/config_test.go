package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadExpandsEnvAndDefaults(t *testing.T) {
	t.Setenv("OPS_TOKEN", "secret-token")
	dir := t.TempDir()
	path := filepath.Join(dir, "agent.yaml")
	content := []byte(`
node:
  id: cn-business-a
  region: cn
center:
  endpoint: http://127.0.0.1:8080/api/v1/agent/ingest
  token: ${OPS_TOKEN}
system:
  enabled: true
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Center.Token != "secret-token" {
		t.Fatalf("token not expanded: %q", cfg.Center.Token)
	}
	if cfg.Collect.Interval.Duration != 30*time.Second {
		t.Fatalf("unexpected interval: %s", cfg.Collect.Interval.Duration)
	}
	if len(cfg.System.DiskMounts) != 1 || cfg.System.DiskMounts[0] != "/" {
		t.Fatalf("unexpected disk defaults: %#v", cfg.System.DiskMounts)
	}
}

func TestLoadRejectsMissingToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agent.yaml")
	content := []byte(`
node:
  id: cn-business-a
center:
  endpoint: http://127.0.0.1:8080/api/v1/agent/ingest
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected validation error")
	}
}
