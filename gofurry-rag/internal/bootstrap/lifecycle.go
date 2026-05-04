package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"

	env "github.com/GoFurry/gofurry-rag/config"
	"github.com/GoFurry/gofurry-rag/internal/db"
	"github.com/GoFurry/gofurry-rag/internal/embedder"
	"github.com/GoFurry/gofurry-rag/internal/ingest"
	ragservice "github.com/GoFurry/gofurry-rag/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	lifecycleMu sync.Mutex
	started     atomic.Bool
	runtime     *Runtime
)

type Runtime struct {
	pool         *pgxpool.Pool
	repo         *db.Repository
	embedClient  embedder.Client
	ragService   *ragservice.Service
	worker       *ingest.Worker
	workerCtx    context.Context
	workerCancel context.CancelFunc
}

func Start() error {
	lifecycleMu.Lock()
	defer lifecycleMu.Unlock()

	if started.Load() {
		return nil
	}

	cfg := env.GetServerConfig()
	initLogger(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	repo := db.NewRepository(pool)
	if cfg.Database.AutoMigrate {
		if err := repo.Migrate(ctx); err != nil {
			pool.Close()
			return fmt.Errorf("migrate database: %w", err)
		}
	}

	embedClient := embedder.NewOllamaClient(cfg.OllamaBaseURL, cfg.EmbedModel, cfg.EmbedDim)
	ragService := ragservice.New(repo, embedClient, *cfg)
	worker := ingest.NewWorker(repo, embedClient, *cfg)
	workerCtx, workerCancel := context.WithCancel(context.Background())
	worker.Start(workerCtx)

	runtime = &Runtime{
		pool:         pool,
		repo:         repo,
		embedClient:  embedClient,
		ragService:   ragService,
		worker:       worker,
		workerCtx:    workerCtx,
		workerCancel: workerCancel,
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

	var shutdownErr error
	if runtime != nil {
		if runtime.workerCancel != nil {
			runtime.workerCancel()
		}
		if runtime.pool != nil {
			runtime.pool.Close()
		}
	}
	started.Store(false)
	runtime = nil
	return shutdownErr
}

func RAGService() *ragservice.Service {
	if runtime == nil {
		return nil
	}
	return runtime.ragService
}

func Worker() *ingest.Worker {
	if runtime == nil {
		return nil
	}
	return runtime.worker
}

func initLogger(cfg *env.Config) {
	level := slog.LevelInfo
	if cfg.Server.Mode == "debug" || cfg.Log.LogLevel == "debug" {
		level = slog.LevelDebug
	}
	var handler slog.Handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	if cfg.Server.Mode != "debug" && cfg.Log.LogMode == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}
	slog.SetDefault(slog.New(handler))
}

func closeRuntimeOnError(cause error, pool *pgxpool.Pool) error {
	if pool != nil {
		pool.Close()
	}
	return errors.Join(cause)
}
