package routers

import (
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	fiberredis "github.com/gofiber/storage/redis/v3"
	"github.com/gofurry/gofurry-nav-backend/roof/env"
	"github.com/redis/go-redis/v9"
)

func TestBuildUptimeConfig(t *testing.T) {
	app := fiber.New()
	redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	t.Cleanup(func() { _ = redisClient.Close() })
	store := fiberredis.NewFromConnection(redisClient)

	cfg, err := buildUptimeConfig(app, store, env.UptimeConfig{
		ServiceID:             " nav-backend ",
		ServiceName:           " Navigation ",
		ServiceDescription:    " Navigation API ",
		SampleIntervalSeconds: 5,
		RetentionDays:         90,
		DaysToShow:            30,
		Timezone:              "Asia/Shanghai",
		StorageKeyPrefix:      " gofurry:test:uptime ",
		UI: env.UptimeUIConfig{
			Path:            " /uptime ",
			Title:           " GoFurry Status ",
			Description:     " Services ",
			Footer:          " GoFurry ",
			GreenThreshold:  0.999,
			YellowThreshold: 0.99,
		},
		Endpoints: []env.UptimeEndpointConfig{
			{
				ID:                  " game-backend ",
				Name:                " Game ",
				Description:         " Game API ",
				URL:                 " http://127.0.0.1:9998/readyz ",
				Method:              " GET ",
				Headers:             map[string]string{"X-Probe": "uptime"},
				ExpectedStatusCodes: []int{200},
				IntervalSeconds:     30,
				TimeoutSeconds:      5,
			},
		},
	}, 7)
	if err != nil {
		t.Fatalf("build uptime config: %v", err)
	}

	if cfg.App != app || cfg.Store != store {
		t.Fatal("app and storage must be passed through")
	}
	if cfg.ServiceID != "nav-backend" || cfg.ServiceName != "Navigation" {
		t.Fatalf("unexpected service identity: %q, %q", cfg.ServiceID, cfg.ServiceName)
	}
	if cfg.NodeID != 7 || cfg.SampleInterval != 5*time.Second {
		t.Fatalf("unexpected node or interval: %d, %s", cfg.NodeID, cfg.SampleInterval)
	}
	if cfg.Timezone.String() != "Asia/Shanghai" {
		t.Fatalf("unexpected timezone: %s", cfg.Timezone)
	}
	if cfg.StorageKeyPrefix != "gofurry:test:uptime" || cfg.UI.Path != "/uptime" {
		t.Fatalf("unexpected storage prefix or UI path: %q, %q", cfg.StorageKeyPrefix, cfg.UI.Path)
	}
	if len(cfg.Endpoints) != 1 {
		t.Fatalf("expected one endpoint, got %d", len(cfg.Endpoints))
	}
	endpoint := cfg.Endpoints[0]
	if endpoint.ID != "game-backend" || endpoint.Interval != 30*time.Second || endpoint.Timeout != 5*time.Second {
		t.Fatalf("unexpected endpoint config: %+v", endpoint)
	}
}

func TestBuildUptimeConfigRejectsInvalidTimezone(t *testing.T) {
	_, err := buildUptimeConfig(fiber.New(), nil, env.UptimeConfig{Timezone: "Mars/Olympus"}, 1)
	if err == nil {
		t.Fatal("expected invalid timezone error")
	}
}
