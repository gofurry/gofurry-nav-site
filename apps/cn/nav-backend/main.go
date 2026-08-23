package main

import (
	"context"
	"errors"
	"os"
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
	"github.com/kardianos/service"
)

func main() {
	dir, _ := os.Getwd()

	svcConfig := &service.Config{
		Name:        common.COMMON_PROJECT_NAME,
		DisplayName: "gf-nav",
		Description: "gf-nav",
		Option: service.KeyValue{
			"SystemdScript": `[Unit]
Description=gf-nav
After=network.target
Requires=network.target

[Service]
Type=simple
WorkingDirectory=` + dir + `/
ExecStart=` + dir + `/gf-nav
Restart=always
RestartSec=30
LogOutput=true
LogDirectory=/var/log/gf-nav
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target`,
		},
	}
	prg := &goFurry{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		gfLog.ErrorKV(err.Error())
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			err = s.Install()
			if err != nil {
				gfLog.ErrorKV("service install failed", "error", err)
			} else {
				gfLog.InfoKV(`┏┓  ┏┓
┃┓┏┓┣ ┓┏┏┓┏┓┓┏
┗┛┗┛┻ ┗┻┛ ┛ ┗┫
             ┛
服务安装成功.
				`)
			}
			return
		case "uninstall":
			err = s.Uninstall()
			if err != nil {
				gfLog.ErrorKV("service uninstall failed", "error", err)
			} else {
				gfLog.InfoKV(`┏┓  ┏┓
┃┓┏┓┣ ┓┏┏┓┏┓┓┏
┗┛┗┛┻ ┗┻┛ ┛ ┗┫
             ┛
服务卸载成功.
				`)
			}
			return
		case "version":
			gfLog.InfoKV(`┏┓  ┏┓
┃┓┏┓┣ ┓┏┏┓┏┓┓┏
┗┛┗┛┻ ┗┻┛ ┛ ┗┫
             ┛
gf-nav V1.0.0
				`)
			return
		case "help":
			gfLog.InfoKV(common.COMMON_PROJECT_HELP)
			return
		}
		return
	}

	// 内存限制和 GC 策略
	debug.SetGCPercent(env.GetServerConfig().Server.GCPercent)
	debug.SetMemoryLimit(int64(env.GetServerConfig().Server.MemoryLimit << 30))

	// 初始化系统服务
	if err := prg.InitOnStart(); err != nil {
		gfLog.ErrorKV("initialize service", "error", err)
		return
	}
	// 启动系统
	err = s.Run()
	if err != nil {
		gfLog.ErrorKV(err.Error())
	}
}

type goFurry struct {
	mu           sync.Mutex
	app          *fiber.App
	pool         *pgxpool.Pool
	dependencies applicationDependencies
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
	gf.dependencies = newApplicationDependencies(gf.pool)

	// 初始化定时任务
	schedule.InitScheduleOnStart(
		gf.dependencies.navStore,
		gf.dependencies.navReader,
		gf.dependencies.home,
		gf.dependencies.views,
	)
	return nil
}

func (gf *goFurry) Start(s service.Service) error {
	app := routers.Router.Init(gf.pool, gf.dependencies.routes)
	gf.mu.Lock()
	gf.app = app
	gf.mu.Unlock()

	go gf.run(app)
	return nil
}

func (gf *goFurry) run(app *fiber.App) {
	addr := env.GetServerConfig().Server.IPAddress + ":" + env.GetServerConfig().Server.Port
	if err := app.Listen(addr, fiber.ListenConfig{
		ListenerNetwork:   env.GetServerConfig().Server.Network,
		EnablePrefork:     env.GetServerConfig().Server.EnablePrefork,
		EnablePrintRoutes: env.GetServerConfig().Server.Mode == "debug",
	}); err != nil {
		gfLog.ErrorKV("web server stopped", "error", err)
	}
}

func (gf *goFurry) Stop(s service.Service) error {
	gf.mu.Lock()
	app := gf.app
	gf.app = nil
	gf.mu.Unlock()

	var shutdownErr error
	if app != nil {
		shutdownErr = app.ShutdownWithTimeout(10 * time.Second)
	}
	if gf.pool != nil {
		gf.pool.Close()
		gf.pool = nil
	}
	redisErr := cs.CloseRedis()
	return errors.Join(shutdownErr, redisErr, gfLog.Sync())
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
