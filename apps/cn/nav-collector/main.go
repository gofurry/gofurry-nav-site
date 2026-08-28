package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gofurry/gofurry-nav-collector/collector/control"
	"github.com/gofurry/gofurry-nav-collector/collector/facts"
	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/log"
	cs "github.com/gofurry/gofurry-nav-collector/common/service"
	internalhealth "github.com/gofurry/gofurry-nav-collector/internal/health"
	"github.com/gofurry/gofurry-nav-collector/internal/infra/postgres"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	executeCLI()
}

func (gf *goFurry) InitOnStart() error {
	if err := initLogger(); err != nil {
		return err
	}
	// 初始化 redis
	cs.InitRedisOnStart()
	return gf.initPostgres()
}

type goFurry struct {
	pool     *pgxpool.Pool
	health   *internalhealth.Server
	control  *control.Engine
	facts    *facts.Engine
	stopOnce sync.Once
}

func (gf *goFurry) Serve(ctx context.Context) error {
	debug.SetGCPercent(1000)
	debug.SetMemoryLimit(int64(env.GetServerConfig().Server.MemoryLimit << 30))
	if err := gf.InitOnStart(); err != nil {
		return err
	}
	healthServer, err := internalhealth.New(env.GetServerConfig().Health, gf.readinessCheck)
	if err != nil {
		return errors.Join(err, gf.Shutdown())
	}
	gf.health = healthServer
	log.InfoFields(map[string]interface{}{
		"service": common.COMMON_PROJECT_NAME,
		"version": env.GetServerConfig().Server.AppVersion,
	}, "采集器服务已启动")
	controlEngine, err := control.NewEngine(gf.pool)
	if err != nil {
		return errors.Join(err, gf.Shutdown())
	}
	gf.control = controlEngine
	if err = gf.control.Start(ctx); err != nil {
		return errors.Join(err, gf.Shutdown())
	}
	factsConfig := env.GetServerConfig().Facts
	gf.facts = facts.New(gf.pool, facts.Options{
		ReconcileInterval: factsConfig.ReconcileInterval(),
		FinalizationGrace: factsConfig.FinalizationGrace(),
		RetentionEnabled:  factsConfig.RetentionEnabled,
		ObservationKeep:   factsConfig.KeepCount(),
		RetentionBatch:    factsConfig.BatchSize(),
	})
	if err = gf.facts.Start(ctx); err != nil {
		return errors.Join(err, gf.Shutdown())
	}
	gf.health.MarkReady()
	if err = gf.health.Start(); err != nil {
		return errors.Join(err, gf.Shutdown())
	}
	select {
	case <-ctx.Done():
		return gf.Shutdown()
	case err = <-gf.health.Errors():
		return errors.Join(err, gf.Shutdown())
	}
}

func (gf *goFurry) Shutdown() error {
	var shutdownErr error
	gf.stopOnce.Do(func() {
		if gf.health != nil {
			gf.health.MarkNotReady()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			shutdownErr = errors.Join(shutdownErr, gf.health.Shutdown(ctx))
			cancel()
		}
		if gf.facts != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			shutdownErr = errors.Join(shutdownErr, gf.facts.Shutdown(ctx))
			cancel()
			gf.facts = nil
		}
		if gf.control != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			shutdownErr = errors.Join(shutdownErr, gf.control.Shutdown(ctx))
			cancel()
			gf.control = nil
		}
		shutdownErr = errors.Join(shutdownErr, cs.CloseRedis())
		if gf.pool != nil {
			gf.pool.Close()
			gf.pool = nil
		}
		shutdownErr = errors.Join(shutdownErr, log.Sync())
	})
	return shutdownErr
}

func (gf *goFurry) readinessCheck(ctx context.Context) error {
	if gf.pool == nil {
		return errors.New("PostgreSQL pool is not initialized")
	}
	if err := gf.pool.Ping(ctx); err != nil {
		return fmt.Errorf("PostgreSQL readiness check: %w", err)
	}
	redisClient := cs.GetRedisService()
	if redisClient == nil {
		return errors.New("Redis client is not initialized")
	}
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis readiness check: %w", err)
	}
	return nil
}

func (gf *goFurry) initPostgres() error {
	dbConfig := env.GetServerConfig().DataBase
	ctx, cancel := context.WithTimeout(context.Background(), durationOrDefault(
		dbConfig.ConnectTimeoutSeconds+dbConfig.PingTimeoutSeconds, 8*time.Second))
	defer cancel()
	pool, err := postgres.Open(ctx, postgres.Config{
		ConnectionString:      dbConfig.ConnectionString(),
		MaxConns:              dbConfig.MaxConns,
		MinConns:              dbConfig.MinConns,
		MaxConnLifetime:       seconds(dbConfig.MaxConnLifetimeSeconds),
		MaxConnLifetimeJitter: seconds(dbConfig.MaxConnLifetimeJitterSeconds),
		MaxConnIdleTime:       seconds(dbConfig.MaxConnIdleTimeSeconds),
		HealthCheckPeriod:     seconds(dbConfig.HealthCheckPeriodSeconds),
		ConnectTimeout:        seconds(dbConfig.ConnectTimeoutSeconds),
		PingTimeout:           seconds(dbConfig.PingTimeoutSeconds),
	}, "gofurry-nav-collector")
	if err != nil {
		return err
	}
	gf.pool = pool
	return nil
}

func seconds(value int) time.Duration {
	return time.Duration(value) * time.Second
}

func durationOrDefault(totalSeconds int, fallback time.Duration) time.Duration {
	if totalSeconds <= 0 {
		return fallback
	}
	return time.Duration(totalSeconds) * time.Second
}

func initLogger() error {
	cfg := env.GetServerConfig()
	logCfg := &log.Config{
		Level:      cfg.Log.LogLevel,
		Mode:       cfg.Log.LogMode,
		FilePath:   cfg.Log.LogPath,
		MaxSize:    cfg.Log.LogMaxSize,
		MaxBackups: cfg.Log.LogMaxBackups,
		MaxAge:     cfg.Log.LogMaxAge,
		Compress:   cfg.Log.LogCompress,
	}
	if cfg.Server.Mode == "debug" {
		logCfg.Level = "debug"
		logCfg.Mode = "dev"
	} else if logCfg.Mode == "" {
		logCfg.Mode = "prod"
	}
	if logCfg.FilePath == "" {
		logCfg.FilePath = "./logs/gf-nav-collector.log"
	} else if filepath.Ext(logCfg.FilePath) == "" || strings.HasSuffix(logCfg.FilePath, "/") || strings.HasSuffix(logCfg.FilePath, "\\") {
		logCfg.FilePath = filepath.Join(logCfg.FilePath, "gf-nav-collector.log")
	}
	if logCfg.MaxBackups == 0 {
		logCfg.MaxBackups = cfg.Log.LogRotationCount
	}
	if err := log.InitLogger(logCfg); err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}
	return nil
}
