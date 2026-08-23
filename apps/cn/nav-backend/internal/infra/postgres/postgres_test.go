package postgres

import (
	"testing"
	"time"
)

func TestParseConfigDefaults(t *testing.T) {
	config, pingTimeout, err := ParseConfig(Config{ConnectionString: "postgres://user:pass@localhost:5432/gfn?sslmode=disable"}, "gofurry-nav-backend")
	if err != nil {
		t.Fatal(err)
	}
	if config.MaxConns != 12 || config.MaxConnLifetime != 30*time.Minute || config.MaxConnLifetimeJitter != 5*time.Minute || config.MaxConnIdleTime != 5*time.Minute || config.HealthCheckPeriod != time.Minute || config.ConnConfig.ConnectTimeout != 5*time.Second || pingTimeout != 3*time.Second {
		t.Fatalf("unexpected pool defaults: %+v ping=%s", config, pingTimeout)
	}
	if got := config.ConnConfig.RuntimeParams["application_name"]; got != "gofurry-nav-backend" {
		t.Fatalf("application_name = %q", got)
	}
}
