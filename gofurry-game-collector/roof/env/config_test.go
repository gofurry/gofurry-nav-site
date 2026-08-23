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
	if cfg.Server.Mode != "debug" || cfg.Server.AppName != "GF-Game-Collector" || cfg.DataBase.DBName != "gfg" {
		t.Fatalf("example config fields changed semantics: %+v", cfg)
	}
	if cfg.Collector.V2.Steam.MaxWorkers <= 0 || cfg.Collector.V2.Retention.PlayerCountsDays <= 0 {
		t.Fatalf("collector v2 config fields were not decoded: %+v", cfg.Collector.V2)
	}
}
