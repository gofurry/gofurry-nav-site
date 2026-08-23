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

	gameService "github.com/gofurry/gofurry-game-collector/collector/game/service"
	"github.com/gofurry/gofurry-game-collector/common/log"
	cs "github.com/gofurry/gofurry-game-collector/common/service"
	"github.com/gofurry/gofurry-game-collector/internal/infra/postgres"
	"github.com/gofurry/gofurry-game-collector/roof/env"
	"github.com/gofurry/gofurry-game-collector/schedule"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	executeCLI()
}

func (gf *goFurry) RunCollectorOnce(run func()) (err error) {
	debug.SetGCPercent(1000)
	debug.SetMemoryLimit(int64(env.GetServerConfig().Server.MemoryLimit << 30))
	if err := gf.InitOnStart(); err != nil {
		return err
	}
	defer func() { err = errors.Join(err, gf.Shutdown()) }()
	gameService.InitLimiter()
	run()
	return nil
}

func (gf *goFurry) InitOnStart() error {
	if err := initLogger(); err != nil {
		return err
	}
	// 初始化 redis
	cs.InitRedisOnStart()
	// 初始化时间调度
	cs.InitTimeWheelOnStart()
	return gf.initPostgres()
}

type goFurry struct {
	pool     *pgxpool.Pool
	stopOnce sync.Once
}

func (gf *goFurry) Serve(ctx context.Context) error {
	debug.SetGCPercent(1000)
	debug.SetMemoryLimit(int64(env.GetServerConfig().Server.MemoryLimit << 30))
	if err := gf.InitOnStart(); err != nil {
		return err
	}
	fmt.Println("gf-game-collector已启动...")
	schedule.InitSchedule()
	<-ctx.Done()
	return gf.Shutdown()
}

func (gf *goFurry) initPostgres() error {
	dbConfig := env.GetServerConfig().DataBase
	ctx, cancel := context.WithTimeout(context.Background(), durationOrDefault(dbConfig.ConnectTimeoutSeconds+dbConfig.PingTimeoutSeconds, 8*time.Second))
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
	}, "gofurry-game-collector")
	if err != nil {
		return fmt.Errorf("open PostgreSQL pool: %w", err)
	}
	gf.pool = pool
	gameService.InitPersistence(pool)
	return nil
}

func (gf *goFurry) Shutdown() error {
	var shutdownErr error
	gf.stopOnce.Do(func() {
		cs.Stop()
		shutdownErr = errors.Join(shutdownErr, cs.CloseRedis())
		if gf.pool != nil {
			gf.pool.Close()
			gf.pool = nil
			gameService.InitPersistence(nil)
		}
		shutdownErr = errors.Join(shutdownErr, log.Sync())
	})
	return shutdownErr
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
		logCfg.FilePath = "./logs/gf-game-collector.log"
	} else if filepath.Ext(logCfg.FilePath) == "" || strings.HasSuffix(logCfg.FilePath, "/") || strings.HasSuffix(logCfg.FilePath, "\\") {
		logCfg.FilePath = filepath.Join(logCfg.FilePath, "gf-game-collector.log")
	}
	if logCfg.MaxBackups == 0 {
		logCfg.MaxBackups = cfg.Log.LogRotationCount
	}
	if err := log.InitLogger(logCfg); err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}
	return nil
}
