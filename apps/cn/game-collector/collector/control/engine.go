package control

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	gamerepository "github.com/gofurry/gofurry-game-collector/collector/game/v2/repository"
	"github.com/gofurry/gofurry-game-collector/common/log"
	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
	"github.com/gofurry/gofurry-game-collector/roof/env"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Engine struct {
	pool         *pgxpool.Pool
	queries      *gamesqlc.Queries
	materializer *Materializer
	worker       *Worker
	retention    *gamerepository.RetentionRepository
	scheduler    gocron.Scheduler
	instanceID   string
	ctx          context.Context
	cancel       context.CancelFunc
	stopOnce     sync.Once
}

func NewEngine(pool *pgxpool.Pool, executor Executor) (*Engine, error) {
	if pool == nil {
		return nil, fmt.Errorf("Game control plane requires PostgreSQL")
	}
	if executor == nil {
		return nil, fmt.Errorf("Game control plane requires an executor")
	}
	instanceID := uuid.NewString()
	return &Engine{
		pool: pool, queries: gamesqlc.New(pool), materializer: NewMaterializer(pool),
		worker: NewWorker(pool, executor, instanceID), retention: gamerepository.NewRetentionRepository(pool), instanceID: instanceID,
	}, nil
}

func (e *Engine) Start(parent context.Context) error {
	if parent == nil {
		parent = context.Background()
	}
	e.ctx, e.cancel = context.WithCancel(parent)
	if err := ensureSchedules(e.ctx, e.queries); err != nil {
		return err
	}
	collectorID, hostname, version, commitSHA := collectorIdentity()
	if err := e.queries.RegisterGameCollectorInstance(e.ctx, gamesqlc.RegisterGameCollectorInstanceParams{
		InstanceID: e.instanceID, CollectorID: collectorID, Hostname: hostname,
		Version: version, CommitSha: commitSHA,
		Capabilities: []string{JobKeyMetadata, JobKeyPlayers},
	}); err != nil {
		return fmt.Errorf("register Game collector instance: %w", err)
	}
	if err := e.worker.RecoverExpired(e.ctx); err != nil {
		return fmt.Errorf("recover Game collector leases: %w", err)
	}
	if err := e.materializer.Reconcile(e.ctx); err != nil {
		return fmt.Errorf("initial Game schedule reconciliation: %w", err)
	}
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return err
	}
	e.scheduler = scheduler
	if err := e.addClockJob("game-schedule-reconcile", 15*time.Second, func() error {
		return e.materializer.Reconcile(e.ctx)
	}); err != nil {
		return err
	}
	if err := e.addClockJob("game-worker-claim", time.Second, func() error {
		_, err := e.worker.ClaimAndRun(e.ctx)
		return err
	}); err != nil {
		return err
	}
	if err := e.addClockJob("game-lease-recovery", 30*time.Second, func() error {
		return e.worker.RecoverExpired(e.ctx)
	}); err != nil {
		return err
	}
	if err := e.addClockJob("game-instance-heartbeat", 20*time.Second, func() error {
		rows, err := e.queries.HeartbeatGameCollectorInstance(e.ctx, e.instanceID)
		if err == nil && rows != 1 {
			return fmt.Errorf("Game collector instance heartbeat row missing")
		}
		return err
	}); err != nil {
		return err
	}
	if err := e.addClockJob("game-task-result-retention", 24*time.Hour, func() error {
		cfg := env.GetServerConfig().Collector.V2.Retention
		return e.retention.Prune(e.ctx, gamerepository.RetentionConfig{
			PlayerCountsDays: cfg.PlayerCountsDays, CollectRunsDays: cfg.CollectRunsDays,
			CollectTaskResultsDays: cfg.CollectTaskResultsDays,
		})
	}); err != nil {
		return err
	}
	e.scheduler.Start()
	log.Info("Game collector control plane started, instance_id=", e.instanceID)
	return nil
}

func (e *Engine) addClockJob(name string, interval time.Duration, action func() error) error {
	_, err := e.scheduler.NewJob(
		gocron.DurationJob(interval),
		gocron.NewTask(func() {
			if e.ctx.Err() != nil {
				return
			}
			if err := action(); err != nil && !errors.Is(err, context.Canceled) {
				log.Error("Game control-plane clock job failed, name=", name, " err=", err)
			}
		}),
		gocron.WithName(name),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	return err
}

func (e *Engine) Shutdown(ctx context.Context) error {
	var shutdownErr error
	e.stopOnce.Do(func() {
		if e.cancel != nil {
			e.cancel()
		}
		if e.scheduler != nil {
			shutdownErr = errors.Join(shutdownErr, e.scheduler.ShutdownWithContext(ctx))
		}
		if e.queries != nil && e.instanceID != "" {
			_, err := e.queries.StopGameCollectorInstance(context.Background(), e.instanceID)
			shutdownErr = errors.Join(shutdownErr, err)
		}
	})
	return shutdownErr
}

func (e *Engine) InstanceID() string {
	if e == nil {
		return ""
	}
	return e.instanceID
}
