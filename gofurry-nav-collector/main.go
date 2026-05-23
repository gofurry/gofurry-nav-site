package main

import (
	"os"
	"runtime/debug"

	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/log"
	cs "github.com/gofurry/gofurry-nav-collector/common/service"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
	"github.com/gofurry/gofurry-nav-collector/schedule"
	"github.com/kardianos/service"
)

//@title gofurry-Collector
//@version v1.0.0
//@description Collector for gofurry Nav Page

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

	InitOnStart()

	// 启动系统
	err = s.Run()
	if err != nil {
		log.Error(err)
	}
}

func InitOnStart() {
	// 初始化 redis
	cs.InitRedisOnStart()
	// 初始化时间调度
	cs.InitTimeWheelOnStart()
}

type goFurry struct{}

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
		}, "Collector service started")
		schedule.InitSchedule()
	}()
}

func (gf *goFurry) Stop(s service.Service) error {
	schedule.StopSchedule()
	return nil
}
