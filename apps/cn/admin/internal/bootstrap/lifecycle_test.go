package bootstrap

import (
	"context"
	"path/filepath"
	"testing"

	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/infra/db"
	log "github.com/gofurry/gofurry-admin/internal/infra/logging"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type stopFunc func()

func (f stopFunc) Stop() { f() }

type syncCheck struct{ check func() }

func (s syncCheck) Write(p []byte) (int, error) { return len(p), nil }
func (s syncCheck) Sync() error                 { s.check(); return nil }

func TestRuntimeStopsSchedulerBeforePoolsAndLogger(t *testing.T) {
	if err := env.MustInitServerConfig("gofurry-admin", filepath.Join("..", "..", "config", "server.example.yaml")); err != nil {
		t.Fatal(err)
	}
	cfg := env.GetServerConfig()
	redisEnabled := cfg.Redis.Enabled
	cfg.Redis.Enabled = false
	defer func() { cfg.Redis.Enabled = redisEnabled }()
	// Lazy, zero-minimum pool: no connection or running PostgreSQL required.
	pool, err := pgxpool.New(context.Background(), "postgres://localhost/gfa?pool_min_conns=0")
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	runtime := &Runtime{Pools: &db.Pools{Admin: pool}}
	runtime.started.Store(true)
	stops, syncs := 0, 0
	runtime.EdgeOneScheduler = stopFunc(func() {
		stops++
		if runtime.started.Load() {
			t.Error("runtime must be unready before scheduler stops")
		}
		if runtime.Pools.Admin != pool || syncs != 0 {
			t.Error("DB/logger torn down before scheduler drained")
		}
	})
	previousLogger := log.GlobalLogger
	log.GlobalLogger = zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), syncCheck{check: func() {
		syncs++
		if stops != 1 || runtime.Pools.Admin != nil {
			t.Error("logger flushed before scheduler/DB shutdown")
		}
	}}, zap.InfoLevel))
	defer func() { log.GlobalLogger = previousLogger }()
	if err := runtime.Shutdown(); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Shutdown(); err != nil {
		t.Fatal(err)
	}
	if stops != 1 || syncs != 1 {
		t.Fatalf("shutdown not idempotent: stops=%d syncs=%d", stops, syncs)
	}
}
