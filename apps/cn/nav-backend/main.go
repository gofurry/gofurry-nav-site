package main

import (
	"context"
	"errors"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/schedule"
	"github.com/gofurry/gofurry-nav-backend/common"
	gfLog "github.com/gofurry/gofurry-nav-backend/common/log"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
	"github.com/gofurry/gofurry-nav-backend/internal/infra/postgres"
	"github.com/gofurry/gofurry-nav-backend/middleware"
	"github.com/gofurry/gofurry-nav-backend/roof/env"
	"github.com/gofurry/gofurry-nav-backend/routers"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	executeCLI()
}

type goFurry struct {
	mu           sync.Mutex
	app          *fiber.App
	pool         *pgxpool.Pool
	dependencies applicationDependencies
	stopOnce     sync.Once
}

func (gf *goFurry) InitOnStart() error {
	cfg := env.GetServerConfig()
	// 初始化自定义日志
	logCfg := &gfLog.Config{
		ShowLine:   true,
		TimeFormat: common.TIME_FORMAT_DATE,
	}
	if cfg.Server.Mode == "debug" {
		logCfg.Level = "debug"
		logCfg.Mode = "dev"
		logCfg.EncodeJson = false
	} else {
		logCfg.Level = cfg.Log.LogLevel
		logCfg.Mode = cfg.Log.LogMode
		logCfg.FilePath = cfg.Log.LogPath
		logCfg.MaxSize = cfg.Log.LogMaxSize
		logCfg.MaxBackups = cfg.Log.LogMaxBackups
		logCfg.MaxAge = cfg.Log.LogMaxAge
		logCfg.Compress = true
		logCfg.EncodeJson = true
		logCfg.TimeFormat = common.TIME_FORMAT_LOG
	}

	// 初始化自定义日志
	err := gfLog.InitLogger(logCfg)
	if err != nil {
		return err
	}

	// 初始化 Coraza 中间件
	if cfg.Waf.WafSwitch {
		middleware.InitGlobalWAF(cfg.Waf)
	}
	// 初始化 redis
	cs.InitRedisOnStart()
	if err := gf.initPostgres(); err != nil {
		return err
	}
	gf.dependencies = newApplicationDependencies(gf.pool)

	if err := cs.InitSchedulerOnStart(); err != nil {
		return err
	}
	if err := schedule.InitScheduleOnStart(
		gf.dependencies.navStore,
		gf.dependencies.navReader,
		gf.dependencies.home,
		gf.dependencies.views,
	); err != nil {
		return err
	}
	return nil
}

func (gf *goFurry) Serve(ctx context.Context) error {
	cfg := env.GetServerConfig()
	debug.SetGCPercent(cfg.Server.GCPercent)
	debug.SetMemoryLimit(int64(cfg.Server.MemoryLimit << 30))
	if err := gf.InitOnStart(); err != nil {
		return err
	}

	app := routers.Router.Init(gf.pool, gf.dependencies.routes)
	gf.mu.Lock()
	gf.app = app
	gf.mu.Unlock()

	listenErr := make(chan error, 1)
	go func() {
		addr := cfg.Server.IPAddress + ":" + cfg.Server.Port
		listenErr <- app.Listen(addr, fiber.ListenConfig{
			ListenerNetwork:   cfg.Server.Network,
			EnablePrefork:     cfg.Server.EnablePrefork,
			EnablePrintRoutes: cfg.Server.Mode == "debug",
		})
	}()

	select {
	case <-ctx.Done():
		return gf.Shutdown()
	case err := <-listenErr:
		if err == nil {
			err = errors.New("web server stopped unexpectedly")
		}
		return errors.Join(err, gf.Shutdown())
	}
}

func (gf *goFurry) Shutdown() error {
	var shutdownErr error
	gf.stopOnce.Do(func() {
		gf.mu.Lock()
		app := gf.app
		gf.app = nil
		gf.mu.Unlock()
		if app != nil {
			shutdownErr = errors.Join(shutdownErr, app.ShutdownWithTimeout(10*time.Second))
		}
		cs.Stop()
		shutdownErr = errors.Join(shutdownErr, cs.CloseRedis())
		if gf.pool != nil {
			gf.pool.Close()
			gf.pool = nil
		}
		shutdownErr = errors.Join(shutdownErr, gfLog.Sync())
	})
	return shutdownErr
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
	}, "gofurry-nav-backend")
	if err != nil {
		return err
	}
	gf.pool = pool
	return nil
}

func seconds(value int) time.Duration { return time.Duration(value) * time.Second }

func durationOrDefault(totalSeconds int, fallback time.Duration) time.Duration {
	if totalSeconds <= 0 {
		return fallback
	}
	return time.Duration(totalSeconds) * time.Second
}
