package env

import (
	"net/url"
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
	if cfg.Server.Mode != "debug" || cfg.Server.Port != "9998" || cfg.DataBase.DBName != "gfg" {
		t.Fatalf("example config fields changed semantics: %+v", cfg)
	}
	if cfg.Middleware.Cors.AllowOrigins == "" || !cfg.Middleware.Limiter.IsOn {
		t.Fatalf("middleware config fields were not decoded: %+v", cfg.Middleware)
	}
	if cfg.DataBase.MaxConns != 12 || cfg.DataBase.ConnectTimeoutSeconds != 5 || cfg.DataBase.PingTimeoutSeconds != 3 {
		t.Fatalf("database pool config was not decoded: %+v", cfg.DataBase)
	}
}

func TestDatabaseConnectionStringEscapesCredentials(t *testing.T) {
	cfg := DataBaseConfig{DBName: "gfg", DBUsername: "user@name", DBPassword: "p@ss:/word", DBHost: "localhost", DBPort: "5432"}
	parsed, err := url.Parse(cfg.ConnectionString())
	if err != nil {
		t.Fatal(err)
	}
	if parsed.User.Username() != cfg.DBUsername {
		t.Fatalf("username = %q, want %q", parsed.User.Username(), cfg.DBUsername)
	}
	password, ok := parsed.User.Password()
	if !ok || password != cfg.DBPassword {
		t.Fatalf("password was not URL escaped and recovered")
	}
	if parsed.Path != "/gfg" || parsed.Query().Get("sslmode") != "prefer" {
		t.Fatalf("connection string fields changed: %s", parsed.Redacted())
	}
}
