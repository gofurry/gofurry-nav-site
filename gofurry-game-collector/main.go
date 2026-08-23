package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	gameService "github.com/gofurry/gofurry-game-collector/collector/game/service"
	"github.com/gofurry/gofurry-game-collector/common"
	"github.com/gofurry/gofurry-game-collector/common/log"
	cs "github.com/gofurry/gofurry-game-collector/common/service"
	"github.com/gofurry/gofurry-game-collector/internal/infra/postgres"
	"github.com/gofurry/gofurry-game-collector/roof/env"
	"github.com/gofurry/gofurry-game-collector/schedule"
	"github.com/jackc/pgx/v5/pgxpool"
	kservice "github.com/kardianos/service"
)

var (
	errChan = make(chan error)
)

func main() {
	svcConfig := &kservice.Config{
		Name:        common.COMMON_PROJECT_NAME,
		DisplayName: "gf-game-collector",
		Description: "gf-game-collector",
	}
	prg := &goFurry{}
	s, err := kservice.New(prg, svcConfig)
	if err != nil {
		log.Error(err)
	}

	if len(os.Args) > 1 {
		if os.Args[1] == "install" {
			err = s.Install()
			if err != nil {
				log.Error("服务安装失败: ", err)
			} else {
				log.Info("服务安装成功.")
			}
			return
		}

		if os.Args[1] == "uninstall" {
			err = s.Uninstall()
			if err != nil {
				log.Error("服务卸载失败: ", err)
			} else {
				log.Info("服务卸载成功.")
			}
			return
		}

		if os.Args[1] == "version" {
			log.Info("gf-game-collector V1.0.0")
			return
		}

		if os.Args[1] == "collect" || os.Args[1] == "full" {
			prg.runCollectorOnce(func() {
				gameService.GetGameService().Collect()
			})
			return
		}

		if os.Args[1] == "players" {
			prg.runCollectorOnce(func() {
				gameService.GetGameService().CollectCurrentPlayers()
			})
			return
		}

		if os.Args[1] == "all" {
			prg.runCollectorOnce(func() {
				gameService.GetGameService().CollectCurrentPlayers()
				gameService.GetGameService().Collect()
			})
			return
		}
	}

	// 内存限制和 GC 策略
	debug.SetGCPercent(1000)
	debug.SetMemoryLimit(int64(env.GetServerConfig().Server.MemoryLimit << 30))

	prg.InitOnStart()

	// 启动系统
	err = s.Run()
	if err != nil {
		log.Error(err)
	}
}

func (gf *goFurry) runCollectorOnce(run func()) {
	debug.SetGCPercent(1000)
	debug.SetMemoryLimit(int64(env.GetServerConfig().Server.MemoryLimit << 30))
	gf.InitOnStart()
	defer gf.closePostgres()
	defer func() { _ = log.Sync() }()
	gameService.InitLimiter()
	run()
}

func (gf *goFurry) InitOnStart() {
	initLogger()
	// 初始化 redis
	cs.InitRedisOnStart()
	// 初始化时间调度
	cs.InitTimeWheelOnStart()
	gf.initPostgres()
}

type goFurry struct {
	pool *pgxpool.Pool
}

func (gf *goFurry) Start(s kservice.Service) error {
	go gf.run()
	return nil
}

func (gf *goFurry) run() {
	// 启动 collector
	go func() {
		// 初始化 collector
		fmt.Println("gf-game-collector已启动...")
		schedule.InitSchedule()
	}()
}

func (gf *goFurry) Stop(s kservice.Service) error {
	gf.closePostgres()
	return log.Sync()
}

func (gf *goFurry) initPostgres() {
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
		log.Fatal("open PostgreSQL pool: ", err)
	}
	gf.pool = pool
	gameService.InitPersistence(pool)
}

func (gf *goFurry) closePostgres() {
	if gf.pool != nil {
		gf.pool.Close()
		gf.pool = nil
		gameService.InitPersistence(nil)
	}
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

func initLogger() {
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
		log.Error("日志初始化失败: ", err)
		os.Exit(1)
	}
}
