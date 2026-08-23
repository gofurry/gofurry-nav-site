// Package db owns Admin's three explicitly bootstrapped PostgreSQL pools.
package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/infra/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pools struct {
	Admin *pgxpool.Pool
	Nav   *pgxpool.Pool
	Game  *pgxpool.Pool
}

func Open(ctx context.Context, cfg *env.ServerConfigHolder) (*Pools, error) {
	pools := &Pools{}
	var err error
	if cfg.DataBase.Enabled {
		pools.Admin, err = openPool(ctx, cfg.DataBase.Postgres, "gofurry-admin-gfa")
		if err != nil {
			return nil, fmt.Errorf("admin database: %w", err)
		}
	}
	if cfg.BusinessDatabases.Nav.Enabled {
		pools.Nav, err = openPool(ctx, cfg.BusinessDatabases.Nav.Postgres, "gofurry-admin-gfn")
		if err != nil {
			pools.Close()
			return nil, fmt.Errorf("nav database: %w", err)
		}
	}
	if cfg.BusinessDatabases.Game.Enabled {
		pools.Game, err = openPool(ctx, cfg.BusinessDatabases.Game.Postgres, "gofurry-admin-gfg")
		if err != nil {
			pools.Close()
			return nil, fmt.Errorf("game database: %w", err)
		}
	}
	return pools, nil
}

func openPool(ctx context.Context, cfg env.SQLDataBaseConfig, applicationName string) (*pgxpool.Pool, error) {
	return postgres.Open(ctx, postgres.Config{
		ConnectionString:      cfg.ConnectionString(),
		MaxConns:              cfg.MaxConns,
		MinConns:              cfg.MinConns,
		MaxConnLifetime:       seconds(cfg.MaxConnLifetimeSeconds),
		MaxConnLifetimeJitter: seconds(cfg.MaxConnLifetimeJitterSeconds),
		MaxConnIdleTime:       seconds(cfg.MaxConnIdleTimeSeconds),
		HealthCheckPeriod:     seconds(cfg.HealthCheckPeriodSeconds),
		ConnectTimeout:        seconds(cfg.ConnectTimeoutSeconds),
		PingTimeout:           seconds(cfg.PingTimeoutSeconds),
	}, applicationName)
}

func (pools *Pools) Ready(ctx context.Context) bool {
	if pools == nil {
		return false
	}
	for _, pool := range []*pgxpool.Pool{pools.Admin, pools.Nav, pools.Game} {
		if pool == nil || pool.Ping(ctx) != nil {
			return false
		}
	}
	return true
}

func (pools *Pools) Close() {
	if pools == nil {
		return
	}
	for _, pool := range []*pgxpool.Pool{pools.Admin, pools.Nav, pools.Game} {
		if pool != nil {
			pool.Close()
		}
	}
	pools.Admin, pools.Nav, pools.Game = nil, nil, nil
}

func Validate(pools *Pools) error {
	if pools == nil || pools.Admin == nil || pools.Nav == nil || pools.Game == nil {
		return errors.New("admin requires gfa, gfn, and gfg PostgreSQL pools")
	}
	return nil
}

func seconds(value int) time.Duration { return time.Duration(value) * time.Second }
