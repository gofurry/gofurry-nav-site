package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitConfigFromFileAndDefaults(t *testing.T) {
	t.Cleanup(ResetForTest)
	dir := t.TempDir()
	path := filepath.Join(dir, "server.yaml")
	if err := os.WriteFile(path, []byte(`
server:
  port: "19090"
auth:
  console_passcode: "secret"
  jwt_secret: "jwt"
database:
  postgres:
    db_name: "ragtest"
    db_username: "rag"
    db_password: "pw"
rag:
  chunk_size: 300
`), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := MustInitServerConfig("gofurry-rag", path); err != nil {
		t.Fatal(err)
	}
	cfg := GetServerConfig()
	if cfg.Server.Port != "19090" {
		t.Fatalf("port = %q", cfg.Server.Port)
	}
	if cfg.ChunkSize != 300 {
		t.Fatalf("chunk size = %d", cfg.ChunkSize)
	}
	if !strings.Contains(cfg.DatabaseDSN, "ragtest") {
		t.Fatalf("dsn = %q", cfg.DatabaseDSN)
	}
	if cfg.Server.TrustProxy {
		t.Fatal("trust_proxy should default to false")
	}
	if cfg.Server.ProxyHeader != "X-Forwarded-For" {
		t.Fatalf("proxy header = %q", cfg.Server.ProxyHeader)
	}
}

func TestEnvOverrideUsesAppPrefix(t *testing.T) {
	t.Cleanup(ResetForTest)
	t.Setenv("APP_SERVER_PORT", "19191")
	t.Setenv("RAG_CHUNK_SIZE", "999")
	t.Setenv("APP_RAG_TENCENT_BASE_URL", "https://example.test/v1")
	t.Setenv("APP_RAG_TENCENT_MODEL", "deepseek-v4-flash")
	t.Setenv("APP_RAG_TENCENT_API_KEY", "secret-key")
	dir := t.TempDir()
	path := filepath.Join(dir, "server.yaml")
	if err := os.WriteFile(path, []byte(`
auth:
  console_passcode: "secret"
  jwt_secret: "jwt"
database:
  postgres:
    db_name: "ragtest"
rag:
  chunk_size: 300
`), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := MustInitServerConfig("gofurry-rag", path); err != nil {
		t.Fatal(err)
	}
	cfg := GetServerConfig()
	if cfg.Server.Port != "19191" {
		t.Fatalf("port = %q", cfg.Server.Port)
	}
	if cfg.ChunkSize != 300 {
		t.Fatalf("legacy env should not override chunk size, got %d", cfg.ChunkSize)
	}
	if cfg.TencentBaseURL != "https://example.test/v1" || cfg.TencentModel != "deepseek-v4-flash" || cfg.TencentAPIKey != "secret-key" {
		t.Fatalf("tencent env override failed: %+v", cfg.RAG)
	}
}

func TestUpdateConsolePasscode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.yaml")
	if err := os.WriteFile(path, []byte("auth:\n  console_passcode: old\n  jwt_secret: jwt\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := UpdateConsolePasscode(path, "new-secret"); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "console_passcode: new-secret") {
		t.Fatalf("content = %s", content)
	}
}

func TestProdConfigRejectsPlaceholderSecrets(t *testing.T) {
	t.Cleanup(ResetForTest)
	dir := t.TempDir()
	path := filepath.Join(dir, "server.yaml")
	if err := os.WriteFile(path, []byte(`
server:
  mode: "prod"
auth:
  console_passcode: "change-me"
  jwt_secret: "change-this-jwt-secret"
  cookie_secure: true
database:
  postgres:
    db_name: "ragtest"
rag:
  embed_dim: 1024
`), 0o600); err != nil {
		t.Fatal(err)
	}

	err := MustInitServerConfig("gofurry-rag", path)
	if err == nil {
		t.Fatal("expected placeholder secret validation error")
	}
	if !strings.Contains(err.Error(), "auth.console_passcode") || !strings.Contains(err.Error(), "auth.jwt_secret") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProdConfigRequiresSecureCookie(t *testing.T) {
	t.Cleanup(ResetForTest)
	dir := t.TempDir()
	path := filepath.Join(dir, "server.yaml")
	if err := os.WriteFile(path, []byte(`
server:
  mode: "prod"
auth:
  console_passcode: "dev-secret"
  jwt_secret: "jwt-secret"
  cookie_secure: false
database:
  postgres:
    db_name: "ragtest"
rag:
  embed_dim: 1024
`), 0o600); err != nil {
		t.Fatal(err)
	}

	err := MustInitServerConfig("gofurry-rag", path)
	if err == nil {
		t.Fatal("expected secure cookie validation error")
	}
	if !strings.Contains(err.Error(), "auth.cookie_secure") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigRejectsUnsupportedEmbedDim(t *testing.T) {
	t.Cleanup(ResetForTest)
	dir := t.TempDir()
	path := filepath.Join(dir, "server.yaml")
	if err := os.WriteFile(path, []byte(`
auth:
  console_passcode: "secret"
  jwt_secret: "jwt"
database:
  postgres:
    db_name: "ragtest"
rag:
  embed_dim: 768
`), 0o600); err != nil {
		t.Fatal(err)
	}

	err := MustInitServerConfig("gofurry-rag", path)
	if err == nil {
		t.Fatal("expected embed_dim validation error")
	}
	if !strings.Contains(err.Error(), "rag.embed_dim") || !strings.Contains(err.Error(), "1024") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTrustProxyRequiresTrustedSource(t *testing.T) {
	t.Cleanup(ResetForTest)
	dir := t.TempDir()
	path := filepath.Join(dir, "server.yaml")
	if err := os.WriteFile(path, []byte(`
server:
  trust_proxy: true
auth:
  console_passcode: "secret"
  jwt_secret: "jwt"
database:
  postgres:
    db_name: "ragtest"
rag:
  embed_dim: 1024
`), 0o600); err != nil {
		t.Fatal(err)
	}

	err := MustInitServerConfig("gofurry-rag", path)
	if err == nil {
		t.Fatal("expected trusted proxy validation error")
	}
	if !strings.Contains(err.Error(), "server.trusted_proxies") {
		t.Fatalf("unexpected error: %v", err)
	}
}
