package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	gameService "github.com/gofurry/gofurry-game-collector/collector/game/service"
	"github.com/gofurry/gofurry-game-collector/common"
	"github.com/gofurry/gofurry-game-collector/common/log"
	cs "github.com/gofurry/gofurry-game-collector/common/service"
	"github.com/gofurry/gofurry-game-collector/roof/env"
	"github.com/gofurry/gofurry-game-collector/schedule"
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
			runCollectorOnce(func() {
				gameService.GetGameService().Collect()
			})
			return
		}

		if os.Args[1] == "players" {
			runCollectorOnce(func() {
				gameService.GetGameService().CollectCurrentPlayers()
			})
			return
		}

		if os.Args[1] == "all" {
			runCollectorOnce(func() {
				gameService.GetGameService().CollectCurrentPlayers()
				gameService.GetGameService().Collect()
			})
			return
		}
	}

	// 内存限制和 GC 策略
	debug.SetGCPercent(1000)
	debug.SetMemoryLimit(int64(env.GetServerConfig().Server.MemoryLimit << 30))

	InitOnStart()

	// 启动系统
	err = s.Run()
	if err != nil {
		log.Error(err)
	}
}

func runCollectorOnce(run func()) {
	debug.SetGCPercent(1000)
	debug.SetMemoryLimit(int64(env.GetServerConfig().Server.MemoryLimit << 30))
	InitOnStart()
	defer func() { _ = log.Sync() }()
	gameService.InitLimiter()
	run()
}

func InitOnStart() {
	initLogger()
	// 初始化 redis
	cs.InitRedisOnStart()
	// 初始化时间调度
	cs.InitTimeWheelOnStart()
}

type goFurry struct{}

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
	return log.Sync()
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
