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
	if cfg.Server.Mode != "debug" || cfg.Server.AppName != "GF-Game-Collector" || cfg.DataBase.DBName != "gfg" {
		t.Fatalf("example config fields changed semantics: %+v", cfg)
	}
	if cfg.Collector.V2.Steam.MaxWorkers <= 0 || cfg.Collector.V2.Retention.PlayerCountsDays <= 0 {
		t.Fatalf("collector v2 config fields were not decoded: %+v", cfg.Collector.V2)
	}
	if cfg.DataBase.MaxConns != 6 || cfg.DataBase.ConnectTimeoutSeconds != 5 || cfg.DataBase.PingTimeoutSeconds != 3 {
		t.Fatalf("PostgreSQL pool config fields were not decoded: %+v", cfg.DataBase)
	}
}

func TestDatabaseConnectionStringEscapesCredentials(t *testing.T) {
	t.Parallel()
	cfg := DataBaseConfig{DBName: "gfg", DBUsername: "user@name", DBPassword: "p:a/ss?word", DBHost: "localhost", DBPort: "5432"}
	parsed, err := url.Parse(cfg.ConnectionString())
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Path != "/gfg" || parsed.User.Username() != cfg.DBUsername {
		t.Fatalf("unexpected connection string identity: %s", parsed.Redacted())
	}
	password, ok := parsed.User.Password()
	if !ok || password != cfg.DBPassword {
		t.Fatal("connection string password did not round trip")
	}
}

func TestExplicitViperConfigDoesNotEnableAutomaticEnv(t *testing.T) {
	previous := configuration
	t.Cleanup(func() { configuration = previous })
	t.Setenv("APP_SERVER_MEMORY_LIMIT", "99")
	if err := LoadServerConfig("../../conf/server.example.yaml"); err != nil {
		t.Fatal(err)
	}
	if configuration.Server.MemoryLimit != 1 {
		t.Fatalf("memory limit = %d, want file value", configuration.Server.MemoryLimit)
	}
}
