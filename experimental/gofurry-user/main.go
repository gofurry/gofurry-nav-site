package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/gofurry/gofurry-user/common"
	cs "github.com/gofurry/gofurry-user/common/service"
	"github.com/gofurry/gofurry-user/common/util"
	"github.com/gofurry/gofurry-user/roof/env"
	"github.com/gofurry/gofurry-user/routers"
	"github.com/gofiber/fiber/v2/log"
	"github.com/kardianos/service"
)

//@title gofurry-User
//@version v1.0.0
//@description gofurry-User

var (
	errChan = make(chan error)
)

func main() {
	dir, _ := os.Getwd()

	svcConfig := &service.Config{
		Name:        common.COMMON_PROJECT_NAME,
		DisplayName: "gf-user",
		Description: "gf-user",
		Option: service.KeyValue{
			"SystemdScript": `[Unit]
Description=gf-user (自定义配置)
After=network.target
Requires=network.target

[Service]
Type=simple
WorkingDirectory=` + dir + `/
ExecStart=` + dir + `/gf-user
Restart=always
RestartSec=30
LogOutput=true
LogDirectory=/var/log/gf-user
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target`,
		},
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
			log.Info("gf-user V1.0.0")
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
	if err := cs.InitEtcdOnStart(); err != nil {
		log.Error(err)
		os.Exit(0)
	}
	// 初始化 redis
	cs.InitRedisOnStart()
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
	// 启动 web
	go func() {
		app := routers.Router.Init()

		addr := env.GetServerConfig().Server.IPAddress + ":" + env.GetServerConfig().Server.Port
		// nginx 完成 https 就不使用 TLS
		//pem := env.GetServerConfig().Key.TlsPem
		//key := env.GetServerConfig().Key.TlsKey
		//if err := app.ListenTLS(addr, pem, key); err != nil {
		//	fmt.Println(err)
		//	errChan <- err
		//}
		if err := app.Listen(addr); err != nil {
			fmt.Println(err)
			errChan <- err
		}
	}()
	if err := <-errChan; err != nil {
		log.Error(err)
	}
}

func (gf *goFurry) Stop(s service.Service) error {
	// 关闭etcd客户端
	if err := cs.CloseEtcdClient(); err != nil {
		log.Error("etcd客户端关闭失败: %v", err)
	} else {
		log.Info("etcd客户端关闭成功")
	}
	// 关闭 grpc 全局连接池
	util.CloseGrpcConns()
	return nil
}
