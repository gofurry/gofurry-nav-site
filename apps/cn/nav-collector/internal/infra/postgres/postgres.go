// Package postgres opens the Nav Collector's explicitly owned pgx pool.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

const defaultMaxConns int32 = 6

type Config struct {
	ConnectionString                                                                                        string
	MaxConns, MinConns                                                                                      int32
	MaxConnLifetime, MaxConnLifetimeJitter, MaxConnIdleTime, HealthCheckPeriod, ConnectTimeout, PingTimeout time.Duration
}

func Open(ctx context.Context, cfg Config, applicationName string) (*pgxpool.Pool, error) {
	poolConfig, pingTimeout, err := ParseConfig(cfg, applicationName)
	if err != nil {
		return nil, err
	}
	connectCtx, cancelConnect := context.WithTimeout(ctx, poolConfig.ConnConfig.ConnectTimeout)
	defer cancelConnect()
	pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL pool: %w", err)
	}
	pingCtx, cancelPing := context.WithTimeout(ctx, pingTimeout)
	defer cancelPing()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}
	return pool, nil
}
func ParseConfig(cfg Config, applicationName string) (*pgxpool.Config, time.Duration, error) {
	if cfg.ConnectionString == "" {
		return nil, 0, errors.New("PostgreSQL connection string is required")
	}
	if applicationName == "" {
		return nil, 0, errors.New("PostgreSQL application name is required")
	}
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectionString)
	if err != nil {
		return nil, 0, fmt.Errorf("parse PostgreSQL pool config: %w", err)
	}
	applyDefaults(poolConfig, &cfg, applicationName)
	return poolConfig, cfg.PingTimeout, nil
}
func applyDefaults(poolConfig *pgxpool.Config, cfg *Config, applicationName string) {
	if cfg.MaxConns <= 0 {
		cfg.MaxConns = defaultMaxConns
	}
	if cfg.MaxConnLifetime <= 0 {
		cfg.MaxConnLifetime = 30 * time.Minute
	}
	if cfg.MaxConnLifetimeJitter <= 0 {
		cfg.MaxConnLifetimeJitter = 5 * time.Minute
	}
	if cfg.MaxConnIdleTime <= 0 {
		cfg.MaxConnIdleTime = 5 * time.Minute
	}
	if cfg.HealthCheckPeriod <= 0 {
		cfg.HealthCheckPeriod = time.Minute
	}
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = 5 * time.Second
	}
	if cfg.PingTimeout <= 0 {
		cfg.PingTimeout = 3 * time.Second
	}
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnLifetimeJitter = cfg.MaxConnLifetimeJitter
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod
	poolConfig.ConnConfig.ConnectTimeout = cfg.ConnectTimeout
	poolConfig.ConnConfig.RuntimeParams["application_name"] = applicationName
}
