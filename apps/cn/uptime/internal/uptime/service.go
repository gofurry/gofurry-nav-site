package uptime

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	contribuptime "github.com/gofiber/contrib/v3/uptime"
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-uptime/internal/config"
	"go.uber.org/zap"
)

type Service struct {
	cfg      *config.Config
	logger   *zap.Logger
	app      *fiber.App
	store    *BoltStore
	ready    atomic.Bool
	stopOnce sync.Once
}

func NewService(cfg *config.Config, logger *zap.Logger) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	storagePath, err := filepath.Abs(strings.TrimSpace(cfg.Storage.Path))
	if err != nil {
		return nil, fmt.Errorf("resolve storage path: %w", err)
	}
	if err = os.MkdirAll(filepath.Dir(storagePath), 0o750); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}
	store, err := OpenBoltStore(storagePath)
	if err != nil {
		return nil, fmt.Errorf("open uptime Bbolt storage: %w", err)
	}
	if err = store.Ping(context.Background()); err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("verify uptime Bbolt storage: %w", err)
	}

	app := fiber.New(fiber.Config{
		AppName:      "GoFurry Uptime",
		ServerHeader: "GoFurry-Uptime",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	})
	service := &Service{cfg: cfg, logger: logger, app: app, store: store}
	app.Get("/livez", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("ok")
	})
	app.Get("/readyz", func(c fiber.Ctx) error {
		if !service.ready.Load() {
			return c.Status(fiber.StatusServiceUnavailable).SendString("not ready")
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := service.store.Ping(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).SendString("not ready")
		}
		return c.Status(fiber.StatusOK).SendString("ready")
	})

	middleware, err := buildMiddleware(app, store, cfg.Uptime)
	if err != nil {
		_ = store.Close()
		return nil, err
	}
	app.Use(middleware)
	service.ready.Store(true)
	return service, nil
}

func buildMiddleware(app *fiber.App, store *BoltStore, cfg config.UptimeConfig) (handler fiber.Handler, err error) {
	location, err := time.LoadLocation(strings.TrimSpace(cfg.Timezone))
	if err != nil {
		return nil, err
	}
	endpoints := make([]contribuptime.EndpointConfig, 0, len(cfg.Endpoints))
	for _, endpoint := range cfg.Endpoints {
		endpoints = append(endpoints, contribuptime.EndpointConfig{
			ID:                  endpoint.ID,
			Name:                endpoint.Name,
			Description:         endpoint.Description,
			URL:                 endpoint.URL,
			Method:              endpoint.Method,
			Headers:             cloneMap(endpoint.Headers),
			ExpectedStatusCodes: append([]int(nil), endpoint.ExpectedStatusCodes...),
			Interval:            time.Duration(endpoint.IntervalSeconds) * time.Second,
			Timeout:             time.Duration(endpoint.TimeoutSeconds) * time.Second,
		})
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("initialize Fiber uptime middleware: %v", recovered)
		}
	}()
	return contribuptime.New(contribuptime.Config{
		App:            app,
		Endpoints:      endpoints,
		SampleInterval: time.Duration(cfg.SampleIntervalSeconds) * time.Second,
		RetentionDays:  cfg.RetentionDays,
		DaysToShow:     cfg.DaysToShow,
		Timezone:       location,
		Store:          store.FiberStorage(),
		UI: contribuptime.UIConfig{
			Path:            strings.TrimSpace(cfg.UI.Path),
			Title:           strings.TrimSpace(cfg.UI.Title),
			Description:     strings.TrimSpace(cfg.UI.Description),
			Footer:          strings.TrimSpace(cfg.UI.Footer),
			GreenThreshold:  cfg.UI.GreenThreshold,
			YellowThreshold: cfg.UI.YellowThreshold,
		},
	}), nil
}

func (s *Service) Serve(ctx context.Context) error {
	listenErr := make(chan error, 1)
	go func() {
		address := net.JoinHostPort(strings.TrimSpace(s.cfg.Server.IPAddress), strings.TrimSpace(s.cfg.Server.Port))
		listenErr <- s.app.Listen(address, fiber.ListenConfig{ListenerNetwork: s.cfg.Server.Network})
	}()

	s.logger.Info("uptime service started",
		zap.String("address", net.JoinHostPort(s.cfg.Server.IPAddress, s.cfg.Server.Port)),
		zap.String("version", s.cfg.Server.AppVersion),
	)
	select {
	case <-ctx.Done():
		return s.Shutdown()
	case err := <-listenErr:
		if err == nil {
			err = errors.New("uptime HTTP server stopped unexpectedly")
		}
		return errors.Join(err, s.Shutdown())
	}
}

func (s *Service) Shutdown() error {
	var result error
	s.stopOnce.Do(func() {
		s.ready.Store(false)
		result = errors.Join(result, s.app.ShutdownWithTimeout(10*time.Second))
		result = errors.Join(result, s.store.Close())
		if err := s.logger.Sync(); err != nil && !errors.Is(err, os.ErrInvalid) {
			result = errors.Join(result, err)
		}
	})
	return result
}

func (s *Service) App() *fiber.App { return s.app }

func cloneMap(source map[string]string) map[string]string {
	if len(source) == 0 {
		return nil
	}
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
