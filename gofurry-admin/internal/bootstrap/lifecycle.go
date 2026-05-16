package bootstrap

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	env "github.com/gofurry/awesome-fiber-template/v3/medium/config"
	authmodels "github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/auth/models"
	auditmodels "github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/shared/audit"
	cache "github.com/gofurry/awesome-fiber-template/v3/medium/internal/infra/cache"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/infra/db"
	log "github.com/gofurry/awesome-fiber-template/v3/medium/internal/infra/logging"
	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/common"
	corazalite "github.com/gofurry/coraza-fiber-lite"
)

var (
	lifecycleMu sync.Mutex
	started     atomic.Bool
)

func databaseModels() []any {
	return []any{
		&authmodels.AdminAccount{},
		&auditmodels.AdminAuditLog{},
	}
}

func Start() error {
	lifecycleMu.Lock()
	defer lifecycleMu.Unlock()

	if started.Load() {
		return nil
	}

	cfg := env.GetServerConfig()

	if err := initLogger(cfg); err != nil {
		return err
	}

	cleanupOnError := func(cause error) error {
		started.Store(false)
		return errors.Join(cause, shutdownComponents(cfg))
	}

	if cfg.Waf.Enabled {
		if err := initWAF(cfg); err != nil {
			return cleanupOnError(err)
		}
	}

	if err := db.InitDatabasesOnStart(databaseModels()...); err != nil {
		return cleanupOnError(fmt.Errorf("database init failed: %w", err))
	}

	if cfg.Redis.Enabled {
		if err := cache.InitRedisOnStart(); err != nil {
			return cleanupOnError(fmt.Errorf("redis init failed: %w", err))
		}
	}

	started.Store(true)
	slog.Info("application bootstrap completed")
	return nil
}

func Shutdown() error {
	lifecycleMu.Lock()
	defer lifecycleMu.Unlock()

	if !started.Load() {
		return nil
	}

	cfg := env.GetServerConfig()
	err := shutdownComponents(cfg)
	started.Store(false)
	return err
}

func initLogger(cfg *env.ServerConfigHolder) error {
	logCfg := &log.Config{
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

	if err := log.InitLogger(logCfg); err != nil {
		return fmt.Errorf("logger init failed: %w", err)
	}
	return nil
}

func initWAF(cfg *env.ServerConfigHolder) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("waf init panic: %v", recovered)
		}
	}()

	corazalite.InitGlobalWAFWithCfg(corazalite.CorazaCfg{
		DirectivesFile:     cfg.Waf.ConfPath,
		RequestBodyAccess:  true,
		ResponseBodyAccess: false,
	})
	corazalite.InitWAFBlockMessage("Request blocked by CorazaWAF")
	return nil
}

func shutdownComponents(cfg *env.ServerConfigHolder) error {
	var shutdownErr error

	if cfg.Redis.Enabled {
		if err := cache.Close(); err != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("redis shutdown failed: %w", err))
		}
	}

	db.Databases.Close()

	if err := log.Sync(); err != nil {
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("logger sync failed: %w", err))
	}

	return shutdownErr
}
