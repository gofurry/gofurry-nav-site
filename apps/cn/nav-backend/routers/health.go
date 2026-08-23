package routers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

const healthProbeTimeout = 2 * time.Second

func registerHealthChecks(app *fiber.App, pool *pgxpool.Pool) {
	app.Get(healthcheck.LivenessEndpoint, healthcheck.New(healthcheck.Config{
		ResponseFormat: healthcheck.FormatJSON,
	}))
	app.Get(healthcheck.ReadinessEndpoint, healthcheck.New(healthcheck.Config{
		Probe: func(_ fiber.Ctx) bool {
			return dependenciesReady(pool)
		},
		ResponseFormat: healthcheck.FormatJSON,
	}))
}

func dependenciesReady(pool *pgxpool.Pool) bool {
	ctx, cancel := context.WithTimeout(context.Background(), healthProbeTimeout)
	defer cancel()

	redisClient := cs.GetRedisService()
	if redisClient == nil || redisClient.Ping(ctx).Err() != nil {
		return false
	}
	return pool != nil && pool.Ping(ctx) == nil
}
