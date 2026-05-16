package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/gofurry/gofurry-oauth-login/common"
	"github.com/gofurry/gofurry-oauth-login/common/log"
	cs "github.com/gofurry/gofurry-oauth-login/common/service"
	"github.com/gofurry/gofurry-oauth-login/env"
	"github.com/gofurry/gofurry-oauth-login/server/service/oauth/github/api"
	"github.com/kardianos/service"
)

var (
	errChan = make(chan error)
)

func main() {
	svcConfig := &service.Config{
		Name:        common.COMMON_PROJECT_NAME,
		DisplayName: "github-oauth-service",
		Description: "github-oauth-service",
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

		if os.Args[1] == "github-oauth-service version" {
			fmt.Println("V1.0.0")
			return
		}

		if os.Args[1] == "info" {
			fmt.Println("OAuth2 Github 三方登录, 基于gPRC和ETCD.")
			return
		}
	}

	// 内存限制和 GC 策略
	debug.SetGCPercent(1000)
	debug.SetMemoryLimit(int64(env.GetServerConfig().Server.MemoryLimit << 30))

	// 初始化系统服务
	InitOnStart()
	// 启动系统
	err = s.Run()
	if err != nil {
		log.Error(err)
	}
}

type goFurry struct{}

func InitOnStart() {
	// 初始化 etcd
	err := cs.InitEtcdOnStart()
	if err != nil {
		log.Error(err)
	}
}

func (gf *goFurry) Start(s service.Service) error {
	go gf.run()
	return nil
}

func (gf *goFurry) run() {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()
	// 启动 gRPC
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error("服务器内部错误: %v, 堆栈: %s", err, string(debug.Stack())) // 记录堆栈
				// 尝试重启服务
				go gf.run()
			}
		}()

		// 启动服务
		api.Api.Init()
	}()
	if err := <-errChan; err != nil {
		log.Error(err)
	}
}

func (gf *goFurry) Stop(s service.Service) error {
	serviceName := env.GetServerConfig().Etcd.EtcdKey
	addr := "127.0.0.1:50056"

	// 注销etcd服务
	if err := cs.UnregisterFromEtcd(serviceName, addr); err != nil {
		log.Error("etcd服务注销失败: %v", err)
	} else {
		log.Info("etcd服务注销成功")
	}

	// 关闭etcd客户端
	if err := cs.CloseEtcdClient(); err != nil {
		log.Error("etcd客户端关闭失败: %v", err)
	} else {
		log.Info("etcd客户端关闭成功")
	}
	return nil
}
