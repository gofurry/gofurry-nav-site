package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	env "github.com/GoFurry/gofurry-rag/config"
	"github.com/GoFurry/gofurry-rag/internal/bootstrap"
	"github.com/GoFurry/gofurry-rag/internal/transport/http/router"
	"github.com/GoFurry/gofurry-rag/pkg/common"
	"github.com/gofiber/fiber/v3"
	"github.com/kardianos/service"
	"github.com/spf13/viper"
)

func runService() error {
	cfg := env.GetServerConfig()
	svc, err := newService()
	if err != nil {
		return err
	}

	debug.SetGCPercent(cfg.Server.GCPercent)
	debug.SetMemoryLimit(int64(cfg.Server.MemoryLimit << 30))

	return svc.Run()
}

func newService() (service.Service, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve executable path failed: %w", err)
	}

	appID, appName := appIdentity()
	args := buildServiceArguments(viper.GetString("config"))
	svcConfig := &service.Config{
		Name:             appID,
		DisplayName:      appName,
		Description:      appName,
		Executable:       exePath,
		Arguments:        args,
		WorkingDirectory: filepath.Dir(exePath),
		Option: service.KeyValue{
			"SystemdScript": buildSystemdScript(appID, appName, exePath, args),
		},
	}

	return service.New(newApp(), svcConfig)
}

func buildSystemdScript(appID, appName, exePath string, args []string) string {
	command := buildSystemdCommand(exePath, args)
	return `[Unit]
Description=` + appName + `
After=network.target
Requires=network.target

[Service]
Type=simple
WorkingDirectory=` + filepath.Dir(exePath) + `
ExecStart=` + command + `
Restart=always
RestartSec=30
LogOutput=true
LogDirectory=/var/log/` + appID + `
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target`
}

func buildSystemdCommand(executable string, args []string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, strconv.Quote(executable))
	for _, arg := range args {
		parts = append(parts, strconv.Quote(arg))
	}
	return strings.Join(parts, " ")
}

func buildServiceArguments(configPath string) []string {
	args := []string{"serve"}
	if configPath = strings.TrimSpace(configPath); configPath != "" {
		if absPath, err := filepath.Abs(configPath); err == nil {
			configPath = absPath
		}
		args = append(args, "--config", configPath)
	}
	return args
}

func appIdentity() (string, string) {
	cfg := env.GetServerConfig()
	appID := cfg.Server.AppID
	if appID == "" {
		appID = common.COMMON_PROJECT_NAME
	}

	appName := cfg.Server.AppName
	if appName == "" {
		appName = appID
	}
	return appID, appName
}

type app struct {
	fiberApp     *fiber.App
	shutdownOnce sync.Once
	stopping     atomic.Bool
}

func newApp() *app {
	return &app{}
}

func (a *app) Start(s service.Service) error {
	if err := bootstrap.Start(); err != nil {
		return err
	}
	a.fiberApp = router.New().Init()
	go a.run()
	return nil
}

func (a *app) run() {
	cfg := env.GetServerConfig()
	addr := cfg.Server.IPAddress + ":" + cfg.Server.Port

	if err := a.fiberApp.Listen(addr, fiber.ListenConfig{
		TLSConfig:         nil,
		EnablePrefork:     cfg.Server.EnablePrefork,
		ListenerNetwork:   cfg.Server.Network,
		EnablePrintRoutes: cfg.Server.Mode == "debug",
	}); err != nil {
		if a.stopping.Load() {
			return
		}
		slog.Error("fiber app exited unexpectedly", "error", err)
		if shutdownErr := a.shutdown(); shutdownErr != nil {
			slog.Error("application shutdown failed", "error", shutdownErr)
		}
		os.Exit(1)
	}
}

func (a *app) Stop(s service.Service) error {
	return a.shutdown()
}

func (a *app) shutdown() error {
	var shutdownErr error
	a.shutdownOnce.Do(func() {
		a.stopping.Store(true)
		if a.fiberApp != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := a.fiberApp.ShutdownWithContext(ctx); err != nil {
				shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown fiber app failed: %w", err))
			}
		}
		if err := bootstrap.Shutdown(); err != nil {
			shutdownErr = errors.Join(shutdownErr, err)
		}
	})
	return shutdownErr
}
