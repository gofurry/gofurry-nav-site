package routers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	cs "github.com/gofurry/gofurry-game-backend/common/service"
	"github.com/gofurry/gofurry-game-backend/roof/db"
)

const healthProbeTimeout = 2 * time.Second

func registerHealthChecks(app *fiber.App) {
	app.Get(healthcheck.LivenessEndpoint, healthcheck.New(healthcheck.Config{
		ResponseFormat: healthcheck.FormatJSON,
	}))
	app.Get(healthcheck.ReadinessEndpoint, healthcheck.New(healthcheck.Config{
		Probe: func(_ fiber.Ctx) bool {
			return dependenciesReady()
		},
		ResponseFormat: healthcheck.FormatJSON,
	}))
}

func dependenciesReady() bool {
	ctx, cancel := context.WithTimeout(context.Background(), healthProbeTimeout)
	defer cancel()

	redisClient := cs.GetRedisService()
	if redisClient == nil || redisClient.Ping(ctx).Err() != nil {
		return false
	}
	return db.Orm.Ping(ctx) == nil
}
