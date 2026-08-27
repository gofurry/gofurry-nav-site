package control

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/gofurry/gofurry-nav-collector/common/log"
	navsqlc "github.com/gofurry/gofurry-nav-collector/internal/db/nav/sqlc"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Engine struct {
	queries      *navsqlc.Queries
	materializer *Materializer
	worker       *Worker
	executor     *ProtocolExecutor
	scheduler    gocron.Scheduler
	instanceID   string
	ctx          context.Context
	cancel       context.CancelFunc
	stopOnce     sync.Once
}

const navTaskResultRetentionDays int64 = 90

func NewEngine(pool *pgxpool.Pool) (*Engine, error) {
	if pool == nil {
		return nil, fmt.Errorf("Nav control plane requires PostgreSQL")
	}
	instanceID := uuid.NewString()
	executor := NewProtocolExecutor(pool)
	return &Engine{
		queries: navsqlc.New(pool), materializer: NewMaterializer(pool),
		worker: NewWorker(pool, executor, instanceID), executor: executor,
		instanceID: instanceID,
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
	collectorID, hostname, version, commitSHA := navCollectorIdentity()
	capabilityNames := make([]string, 0, len(catalog()))
	for _, item := range catalog() {
		capabilityNames = append(capabilityNames, item.JobKey)
	}
	if err := e.queries.RegisterNavCollectorInstance(e.ctx, navsqlc.RegisterNavCollectorInstanceParams{
		InstanceID: e.instanceID, CollectorID: collectorID, Hostname: hostname,
		Version: version, CommitSha: commitSHA, Capabilities: capabilityNames,
	}); err != nil {
		return err
	}
	if err := e.worker.RecoverExpired(e.ctx); err != nil {
		return err
	}
	if err := e.materializer.Reconcile(e.ctx); err != nil {
		return err
	}
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return err
	}
	e.scheduler = scheduler
	jobs := []struct {
		name     string
		interval time.Duration
		action   func() error
	}{
		{"nav-schedule-reconcile", 15 * time.Second, func() error { return e.materializer.Reconcile(e.ctx) }},
		{"nav-worker-claim", time.Second, func() error { _, err := e.worker.ClaimAndRun(e.ctx); return err }},
		{"nav-lease-recovery", 30 * time.Second, func() error { return e.worker.RecoverExpired(e.ctx) }},
		{"nav-instance-heartbeat", 20 * time.Second, func() error {
			rows, err := e.queries.HeartbeatNavCollectorInstance(e.ctx, e.instanceID)
			if err == nil && rows != 1 {
				return fmt.Errorf("Nav collector instance heartbeat row missing")
			}
			return err
		}},
		{"nav-task-result-retention", 24 * time.Hour, func() error {
			_, err := e.queries.DeleteNavCollectionTaskResultsOlderThan(e.ctx, navTaskResultRetentionDays)
			return err
		}},
		{"nav-legacy-log-retention", 24 * time.Hour, func() error { e.executor.RetainLegacyLogs(); return nil }},
	}
	for _, job := range jobs {
		if err := e.addClockJob(job.name, job.interval, job.action); err != nil {
			return err
		}
	}
	e.scheduler.Start()
	log.InfoFields(map[string]interface{}{"instance_id": e.instanceID}, "Nav collector control plane started")
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
				log.ErrorFields(map[string]interface{}{"job": name}, "Nav control-plane clock job failed: "+err.Error())
			}
		}),
		gocron.WithName(name), gocron.WithSingletonMode(gocron.LimitModeReschedule),
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
		shutdownErr = errors.Join(shutdownErr, e.worker.Wait(ctx))
		if e.queries != nil {
			_, err := e.queries.StopNavCollectorInstance(context.Background(), e.instanceID)
			shutdownErr = errors.Join(shutdownErr, err)
		}
		if e.executor != nil {
			e.executor.Close()
		}
	})
	return shutdownErr
}

func navCollectorIdentity() (collectorID, hostname, version, commitSHA string) {
	hostname, _ = os.Hostname()
	if strings.TrimSpace(hostname) == "" {
		hostname = "unknown-host"
	}
	collectorID = strings.TrimSpace(env.GetServerConfig().Collector.Scheduler.CollectorID)
	if collectorID == "" {
		collectorID = hostname
	}
	version = env.GetServerConfig().Server.AppVersion
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				commitSHA = setting.Value
				break
			}
		}
	}
	return
}
