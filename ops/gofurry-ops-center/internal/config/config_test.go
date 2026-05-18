package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadConfigDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "center.yaml")
	err := os.WriteFile(path, []byte(`
storage:
  dsn: postgres://ops@127.0.0.1:5432/gofurry_ops?sslmode=disable
security:
  dashboard_passcode: local-access-code-123
  session_secret: 0123456789abcdef0123456789abcdef
  agent_tokens:
    - node_id: cn-business-a
      token: abcdef0123456789abcdef0123456789
`), 0o600)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CenterID != "ops-center-local" {
		t.Fatalf("unexpected center id: %s", cfg.CenterID)
	}
	if cfg.Security.SignatureWindow.Duration != 5*time.Minute {
		t.Fatalf("unexpected signature window: %s", cfg.Security.SignatureWindow.Duration)
	}
	if cfg.AgentTokenMap()["cn-business-a"] != "abcdef0123456789abcdef0123456789" {
		t.Fatal("agent token map missing token")
	}
}

func TestLoadRejectsWeakSecrets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "center.yaml")
	err := os.WriteFile(path, []byte(`
storage:
  dsn: postgres://ops@127.0.0.1:5432/gofurry_ops?sslmode=disable
security:
  dashboard_passcode: change-me-dashboard-passcode
  session_secret: change-me-long-random-session-secret
  agent_tokens:
    - node_id: cn-business-a
      token: change-me-agent-token
`), 0o600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Load(path)
	if err == nil {
		t.Fatal("expected weak secret validation error")
	}
	if !strings.Contains(err.Error(), "weak placeholder") {
		t.Fatalf("expected weak placeholder error, got %v", err)
	}
}
