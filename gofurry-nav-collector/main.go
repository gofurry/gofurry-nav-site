package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/log"
	cs "github.com/gofurry/gofurry-nav-collector/common/service"
	"github.com/gofurry/gofurry-nav-collector/internal/infra/postgres"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
	"github.com/gofurry/gofurry-nav-collector/schedule"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kardianos/service"
)

var (
	errChan = make(chan error)
)

func main() {
	svcConfig := &service.Config{
		Name:        common.COMMON_PROJECT_NAME,
		DisplayName: "gf-nav-collector",
		Description: "gf-nav-collector",
	}
	prg := &goFurry{}
	s, err := service.New(prg, svcConfig)
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
			log.Info("gf-nav-collector V1.0.0")
			return
		}
	}

	// 内存限制和 GC 策略
	debug.SetGCPercent(1000)
	debug.SetMemoryLimit(int64(env.GetServerConfig().Server.MemoryLimit << 30))

	if err := prg.InitOnStart(); err != nil {
		log.Error("服务初始化失败: ", err)
		return
	}

	// 启动系统
	err = s.Run()
	if err != nil {
		log.Error(err)
	}
}

func (gf *goFurry) InitOnStart() error {
	initLogger()
	// 初始化 redis
	cs.InitRedisOnStart()
	// 初始化时间调度
	cs.InitTimeWheelOnStart()
	return gf.initPostgres()
}

type goFurry struct {
	pool *pgxpool.Pool
}

func (gf *goFurry) Start(s service.Service) error {
	go gf.run()
	return nil
}

func (gf *goFurry) run() {
	// 启动 collector
	go func() {
		// 初始化 collector
		log.InfoFields(map[string]interface{}{
			"service": common.COMMON_PROJECT_NAME,
			"version": env.GetServerConfig().Server.AppVersion,
		}, "采集器服务已启动")
		schedule.InitSchedule(gf.pool)
	}()
}

func (gf *goFurry) Stop(s service.Service) error {
	schedule.StopSchedule()
	if gf.pool != nil {
		gf.pool.Close()
		gf.pool = nil
	}
	return log.Sync()
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
		logCfg.FilePath = "./logs/gf-nav-collector.log"
	} else if filepath.Ext(logCfg.FilePath) == "" || strings.HasSuffix(logCfg.FilePath, "/") || strings.HasSuffix(logCfg.FilePath, "\\") {
		logCfg.FilePath = filepath.Join(logCfg.FilePath, "gf-nav-collector.log")
	}
	if logCfg.MaxBackups == 0 {
		logCfg.MaxBackups = cfg.Log.LogRotationCount
	}
	if err := log.InitLogger(logCfg); err != nil {
		log.Error("日志初始化失败: ", err)
		os.Exit(1)
	}
}
