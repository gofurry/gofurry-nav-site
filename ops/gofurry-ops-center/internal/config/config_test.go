package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfigDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "center.yaml")
	err := os.WriteFile(path, []byte(`
storage:
  dsn: postgres://ops:pass@127.0.0.1:5432/gofurry_ops?sslmode=disable
security:
  dashboard_passcode: change-me
  session_secret: session-secret
  agent_tokens:
    - node_id: cn-business-a
      token: agent-token
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
	if cfg.AgentTokenMap()["cn-business-a"] != "agent-token" {
		t.Fatal("agent token map missing token")
	}
}
