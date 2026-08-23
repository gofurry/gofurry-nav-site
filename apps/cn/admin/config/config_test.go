package env

import (
	"strings"
	"testing"
)

func TestDatabaseConfigIsPostgreSQLOnly(t *testing.T) {
	for _, driver := range []string{"mysql", "sqlite"} {
		cfg := DataBaseConfig{Enabled: true, DBType: driver}
		cfg.normalize("gfa")
		err := cfg.validate("database")
		if err == nil || !strings.Contains(err.Error(), "PostgreSQL-only") {
			t.Fatalf("driver %q should be rejected as PostgreSQL-only, got %v", driver, err)
		}
	}
}

func TestDatabaseConfigPreservesLegacyPostgresFields(t *testing.T) {
	cfg := DataBaseConfig{
		Enabled: true,
		DBType:  "postgres",
		DBName:  "gfa_custom",
		DBHost:  "db.internal",
		DBPort:  "5544",
		DBUser:  "admin_user",
		DBPass:  "secret",
	}
	cfg.normalize("gfa")

	if err := cfg.validate("database"); err != nil {
		t.Fatalf("validate() error = %v", err)
	}
	if cfg.Postgres.DBName != cfg.DBName || cfg.Postgres.DBHost != cfg.DBHost || cfg.Postgres.DBPort != cfg.DBPort || cfg.Postgres.DBUser != cfg.DBUser || cfg.Postgres.DBPass != cfg.DBPass {
		t.Fatalf("legacy PostgreSQL fields were not preserved: %+v", cfg.Postgres)
	}
}

func TestExplicitViperConfigDoesNotEnableAutomaticEnv(t *testing.T) {
	t.Setenv("APP_SERVER_PORT", "1")
	var cfg serverConfig
	if err := InitConfig("gofurry-admin", "server.yaml", "server.example.yaml", &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != "10099" {
		t.Fatalf("server port = %q, want file value", cfg.Server.Port)
	}
}
