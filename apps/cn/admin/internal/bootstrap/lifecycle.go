package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	env "github.com/gofurry/gofurry-admin/config"
	authcontroller "github.com/gofurry/gofurry-admin/internal/app/auth/controller"
	authservice "github.com/gofurry/gofurry-admin/internal/app/auth/service"
	collectioncontroller "github.com/gofurry/gofurry-admin/internal/app/collectionadmin/controller"
	collectionservice "github.com/gofurry/gofurry-admin/internal/app/collectionadmin/service"
	gameadmin "github.com/gofurry/gofurry-admin/internal/app/gameadmin/controller"
	metricadmin "github.com/gofurry/gofurry-admin/internal/app/metricadmin"
	navadmin "github.com/gofurry/gofurry-admin/internal/app/navadmin/controller"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	options "github.com/gofurry/gofurry-admin/internal/app/shared/options/controller"
	cache "github.com/gofurry/gofurry-admin/internal/infra/cache"
	"github.com/gofurry/gofurry-admin/internal/infra/db"
	log "github.com/gofurry/gofurry-admin/internal/infra/logging"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

type Runtime struct {
	Pools         *db.Pools
	Audit         *audit.Logger
	AuthService   *authservice.AuthService
	AuthAPI       *authcontroller.AuthAPI
	NavAPI        *navadmin.NavAPI
	GameAPI       *gameadmin.GameAPI
	OptionsAPI    *options.OptionsAPI
	CollectionAPI *collectioncontroller.API
	MetricAPI     *metricadmin.API

	started      atomic.Bool
	shutdownOnce sync.Once
}

func Start() (*Runtime, error) {
	cfg := env.GetServerConfig()
	if err := initLogger(cfg); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pools, err := db.Open(ctx, cfg)
	if err != nil {
		_ = log.Sync()
		return nil, fmt.Errorf("database init failed: %w", err)
	}
	cleanupOnError := func(cause error) (*Runtime, error) {
		pools.Close()
		return nil, errors.Join(cause, log.Sync())
	}
	if err := db.Validate(pools); err != nil {
		return cleanupOnError(err)
	}
	if cfg.Redis.Enabled {
		if err := cache.InitRedisOnStart(); err != nil {
			return cleanupOnError(fmt.Errorf("redis init failed: %w", err))
		}
	}

	auditLogger := audit.New(pools.Admin)
	auth := authservice.New(pools.Admin, auditLogger)
	runtime := &Runtime{
		Pools: pools, Audit: auditLogger, AuthService: auth,
		AuthAPI: authcontroller.New(auth, auditLogger), NavAPI: navadmin.New(pools.Nav, auditLogger),
		GameAPI: gameadmin.New(pools.Game, auditLogger), OptionsAPI: options.New(pools.Nav, pools.Game),
		CollectionAPI: collectioncontroller.New(collectionservice.New(pools.Game, pools.Nav, auditLogger)),
		MetricAPI:     metricadmin.NewAPI(metricadmin.New(pools.Game, pools.Nav)),
	}
	runtime.started.Store(true)
	log.InfoKV("application bootstrap completed")
	return runtime, nil
}

func (runtime *Runtime) Shutdown() error {
	if runtime == nil {
		return nil
	}
	var shutdownErr error
	runtime.shutdownOnce.Do(func() {
		runtime.started.Store(false)
		if env.GetServerConfig().Redis.Enabled {
			if err := cache.Close(); err != nil {
				shutdownErr = errors.Join(shutdownErr, fmt.Errorf("redis shutdown failed: %w", err))
			}
		}
		runtime.Pools.Close()
		if err := log.Sync(); err != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("logger sync failed: %w", err))
		}
	})
	return shutdownErr
}

func initLogger(cfg *env.ServerConfigHolder) error {
	logCfg := &log.Config{ShowLine: true, TimeFormat: common.TIME_FORMAT_DATE}
	if cfg.Server.Mode == "debug" {
		logCfg.Level, logCfg.Mode, logCfg.EncodeJson = "debug", "dev", false
	} else {
		logCfg.Level, logCfg.Mode, logCfg.FilePath = cfg.Log.LogLevel, cfg.Log.LogMode, cfg.Log.LogPath
		logCfg.MaxSize, logCfg.MaxBackups, logCfg.MaxAge = cfg.Log.LogMaxSize, cfg.Log.LogMaxBackups, cfg.Log.LogMaxAge
		logCfg.Compress, logCfg.EncodeJson, logCfg.TimeFormat = true, true, common.TIME_FORMAT_LOG
	}
	if err := log.InitLogger(logCfg); err != nil {
		return fmt.Errorf("logger init failed: %w", err)
	}
	return nil
}
