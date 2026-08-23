package cmd

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/bootstrap"
	"github.com/gofurry/gofurry-admin/internal/transport/http/router"
)

func runService(ctx context.Context) error {
	cfg := env.GetServerConfig()
	debug.SetGCPercent(cfg.Server.GCPercent)
	debug.SetMemoryLimit(int64(cfg.Server.MemoryLimit << 30))

	application := newApp()
	if err := application.start(); err != nil {
		return err
	}
	return application.wait(ctx)
}

type app struct {
	fiberApp     *fiber.App
	runtime      *bootstrap.Runtime
	shutdownOnce sync.Once
	stopping     atomic.Bool
	listenErr    chan error
}

func newApp() *app {
	return &app{}
}

func (a *app) start() error {
	runtime, err := bootstrap.Start()
	if err != nil {
		return err
	}
	a.runtime = runtime

	a.fiberApp = router.New(runtime).Init()
	a.listenErr = make(chan error, 1)
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
			a.listenErr <- nil
			return
		}
		a.listenErr <- err
		return
	}
	a.listenErr <- nil
}

func (a *app) wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return a.shutdown()
	case err := <-a.listenErr:
		if err == nil {
			err = errors.New("fiber app stopped unexpectedly")
		}
		return errors.Join(err, a.shutdown())
	}
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

		if err := a.runtime.Shutdown(); err != nil {
			shutdownErr = errors.Join(shutdownErr, err)
		}
	})

	return shutdownErr
}
