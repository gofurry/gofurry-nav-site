package main

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	v2controller "github.com/gofurry/gofurry-game-backend/apps/game/v2/controller"
	v2dao "github.com/gofurry/gofurry-game-backend/apps/game/v2/dao"
	v2service "github.com/gofurry/gofurry-game-backend/apps/game/v2/service"
	prizecontroller "github.com/gofurry/gofurry-game-backend/apps/prize/controller"
	prizedao "github.com/gofurry/gofurry-game-backend/apps/prize/dao"
	prizeservice "github.com/gofurry/gofurry-game-backend/apps/prize/service"
	reviewdao "github.com/gofurry/gofurry-game-backend/apps/review/dao"
	reviewservice "github.com/gofurry/gofurry-game-backend/apps/review/service"
	"github.com/gofurry/gofurry-game-backend/apps/schedule"
	"github.com/gofurry/gofurry-game-backend/common"
	gfLog "github.com/gofurry/gofurry-game-backend/common/log"
	cs "github.com/gofurry/gofurry-game-backend/common/service"
	"github.com/gofurry/gofurry-game-backend/internal/infra/postgres"
	"github.com/gofurry/gofurry-game-backend/middleware"
	"github.com/gofurry/gofurry-game-backend/roof/env"
	"github.com/gofurry/gofurry-game-backend/routers"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	executeCLI()
}

type goFurry struct {
	mu       sync.Mutex
	app      *fiber.App
	stopOnce sync.Once
	pool     *pgxpool.Pool
	readDAO  *v2dao.ReadModelDAO
	viewSvc  *v2service.GameViewService
	prizeDAO *prizedao.PrizeDAO
	gameAPI  *v2controller.GameV2API
	prizeAPI *prizecontroller.PrizeAPI
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
	// 初始化时间调度
	cs.InitTimeWheelOnStart()

	if err := gf.initPostgres(); err != nil {
		return err
	}
	prizeService := prizeservice.New(gf.prizeDAO)
	reviewService := reviewservice.New(reviewdao.New(gf.pool))
	gf.gameAPI = v2controller.New(gf.readDAO, gf.viewSvc, reviewService)
	gf.prizeAPI = prizecontroller.New(prizeService)

	// 初始化定时任务
	schedule.InitScheduleOnStart(gf.readDAO, gf.viewSvc, gf.prizeDAO)
	return nil
}

func (gf *goFurry) Serve(ctx context.Context) error {
	cfg := env.GetServerConfig()
	debug.SetGCPercent(cfg.Server.GCPercent)
	debug.SetMemoryLimit(int64(cfg.Server.MemoryLimit << 30))
	if err := gf.InitOnStart(); err != nil {
		return fmt.Errorf("initialize service: %w", err)
	}

	app := routers.Router.Init(gf.pool, gf.gameAPI, gf.prizeAPI)
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
	}, "gofurry-game-backend")
	if err != nil {
		return err
	}
	gf.pool = pool
	gf.readDAO = v2dao.NewReadModelDAO(pool)
	gf.viewSvc = v2service.NewGameViewService(pool)
	gf.prizeDAO = prizedao.New(pool)
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
