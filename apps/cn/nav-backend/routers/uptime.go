package routers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/contrib/v3/uptime"
	"github.com/gofiber/fiber/v3"
	fiberredis "github.com/gofiber/storage/redis/v3"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
	"github.com/gofurry/gofurry-nav-backend/roof/env"
)

func registerUptime(app *fiber.App) {
	cfg := env.GetServerConfig()
	if !cfg.Uptime.Enabled {
		return
	}

	redisClient := cs.GetRedisService()
	if redisClient == nil {
		panic("uptime: redis service is not initialized")
	}

	// Uptime shares the application's Redis client. Its wrapper must not be
	// closed independently because the application owns that connection pool.
	store := fiberredis.NewFromConnection(redisClient)
	uptimeConfig, err := buildUptimeConfig(app, store, cfg.Uptime, cfg.ClusterId)
	if err != nil {
		panic(fmt.Errorf("uptime: invalid configuration: %w", err))
	}

	app.Use(uptime.New(uptimeConfig))
}

func buildUptimeConfig(app *fiber.App, store *fiberredis.Storage, cfg env.UptimeConfig, nodeID int) (uptime.Config, error) {
	location := time.Local
	if timezone := strings.TrimSpace(cfg.Timezone); timezone != "" {
		var err error
		location, err = time.LoadLocation(timezone)
		if err != nil {
			return uptime.Config{}, fmt.Errorf("load timezone %q: %w", timezone, err)
		}
	}

	endpoints := make([]uptime.EndpointConfig, 0, len(cfg.Endpoints))
	for _, endpoint := range cfg.Endpoints {
		endpoints = append(endpoints, uptime.EndpointConfig{
			ID:                  strings.TrimSpace(endpoint.ID),
			Name:                strings.TrimSpace(endpoint.Name),
			Description:         strings.TrimSpace(endpoint.Description),
			URL:                 strings.TrimSpace(endpoint.URL),
			Method:              strings.TrimSpace(endpoint.Method),
			Headers:             cloneStringMap(endpoint.Headers),
			ExpectedStatusCodes: append([]int(nil), endpoint.ExpectedStatusCodes...),
			Interval:            time.Duration(endpoint.IntervalSeconds) * time.Second,
			Timeout:             time.Duration(endpoint.TimeoutSeconds) * time.Second,
		})
	}

	return uptime.Config{
		App:                app,
		ServiceID:          strings.TrimSpace(cfg.ServiceID),
		ServiceName:        strings.TrimSpace(cfg.ServiceName),
		ServiceDescription: strings.TrimSpace(cfg.ServiceDescription),
		Endpoints:          endpoints,
		SampleInterval:     time.Duration(cfg.SampleIntervalSeconds) * time.Second,
		RetentionDays:      cfg.RetentionDays,
		DaysToShow:         cfg.DaysToShow,
		Timezone:           location,
		NodeID:             int64(nodeID),
		Store:              store,
		StorageKeyPrefix:   strings.TrimSpace(cfg.StorageKeyPrefix),
		UI: uptime.UIConfig{
			Path:            strings.TrimSpace(cfg.UI.Path),
			Title:           strings.TrimSpace(cfg.UI.Title),
			Description:     strings.TrimSpace(cfg.UI.Description),
			Footer:          strings.TrimSpace(cfg.UI.Footer),
			GreenThreshold:  cfg.UI.GreenThreshold,
			YellowThreshold: cfg.UI.YellowThreshold,
		},
	}, nil
}

func cloneStringMap(source map[string]string) map[string]string {
	if len(source) == 0 {
		return nil
	}

	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
