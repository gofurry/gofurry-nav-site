package collector

import (
	"context"
	"time"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/config"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/model"
	"github.com/jackc/pgx/v5"
)

func collectPostgres(ctx context.Context, cfg config.PostgresConfig) model.ServiceCheck {
	result := model.ServiceCheck{Name: cfg.Name}
	checkCtx, cancel := context.WithTimeout(ctx, cfg.Timeout.Duration)
	defer cancel()

	start := time.Now()
	conn, err := pgx.Connect(checkCtx, cfg.DSN)
	if err != nil {
		result.Status = statusFromError(err)
		result.ErrorMessage = err.Error()
		return result
	}
	defer conn.Close(context.Background())

	if err := conn.Ping(checkCtx); err != nil {
		result.LatencyMS = time.Since(start).Milliseconds()
		result.Status = statusFromError(err)
		result.ErrorMessage = err.Error()
		return result
	}
	var one int
	if err := conn.QueryRow(checkCtx, `SELECT 1`).Scan(&one); err != nil {
		result.LatencyMS = time.Since(start).Milliseconds()
		result.Status = statusFromError(err)
		result.ErrorMessage = err.Error()
		return result
	}
	result.LatencyMS = time.Since(start).Milliseconds()
	result.Status = "ok"
	_ = conn.QueryRow(checkCtx, `SELECT pg_database_size(current_database())`).Scan(&result.DatabaseSize)
	_ = conn.QueryRow(checkCtx, `SELECT count(*)::bigint FROM pg_stat_activity WHERE datname = current_database()`).Scan(&result.Connections)
	return result
}
