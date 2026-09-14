package env

import (
	"net/url"
	"os"
	"testing"

	"go.yaml.in/yaml/v4"
)

func TestExampleConfigDecodesWithYAMLV4(t *testing.T) {
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
	if cfg.Redis.RedisUsername != "gofurry_app" {
		t.Fatalf("redis username = %q, want gofurry_app", cfg.Redis.RedisUsername)
	}
	if cfg.Middleware.Cors.AllowOrigins == "" || !cfg.Middleware.Limiter.IsOn {
		t.Fatalf("middleware config fields were not decoded: %+v", cfg.Middleware)
	}
	if cfg.DataBase.MaxConns != 12 || cfg.DataBase.ConnectTimeoutSeconds != 5 || cfg.DataBase.PingTimeoutSeconds != 3 {
		t.Fatalf("database pool config was not decoded: %+v", cfg.DataBase)
	}
}

func TestRedisUsernameIsOptional(t *testing.T) {
	var cfg serverConfig
	if err := yaml.Unmarshal([]byte("redis:\n  redis_addr: 127.0.0.1:6379\n"), &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Redis.RedisUsername != "" {
		t.Fatalf("omitted redis username = %q, want empty", cfg.Redis.RedisUsername)
	}
}

func TestDevelopmentHomeCacheCannotBeEnabledInProduction(t *testing.T) {
	data, err := os.ReadFile("../../conf/server.example.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		mode    string
		seconds int
		valid   bool
	}{
		{"debug", 0, true}, {"debug", 10, true}, {"debug", 31, false}, {"debug", -1, false},
		{"release", 0, true}, {"release", 10, false}, {"production", 10, false},
	} {
		var cfg serverConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			t.Fatal(err)
		}
		cfg.Server.Mode = tc.mode
		cfg.Server.DevelopmentHomeCacheSeconds = tc.seconds
		if err := cfg.validate(); (err == nil) != tc.valid {
			t.Fatalf("mode=%s seconds=%d error=%v", tc.mode, tc.seconds, err)
		}
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

func TestExplicitViperConfigDoesNotEnableAutomaticEnv(t *testing.T) {
	previous := configuration
	t.Cleanup(func() { configuration = previous })
	t.Setenv("APP_SERVER_PORT", "1")
	if err := LoadServerConfig("../../conf/server.example.yaml"); err != nil {
		t.Fatal(err)
	}
	if configuration.Server.Port != "9998" {
		t.Fatalf("server port = %q, want file value", configuration.Server.Port)
	}
}
