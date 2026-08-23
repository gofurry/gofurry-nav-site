package env

import (
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
}
