package cloudops

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	infra "github.com/gofurry/gofurry-admin/internal/infra/cloudops"
	log "github.com/gofurry/gofurry-admin/internal/infra/logging"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
)

const scheduledPurgeAction = "cloud.edgeone.purge.host.scheduled"

// This boundary intentionally has no zone-wide purge operation.
type scheduledPurger interface {
	Validate(string, infra.PurgeRequest) (infra.PurgeRequest, error)
	Purge(context.Context, string, infra.PurgeRequest) (infra.PurgeResult, error)
}

type scheduledPurgeStore interface {
	Acquire(context.Context) (scheduledPurgeLease, error)
}

type scheduledPurgeLease interface {
	Recorded(context.Context, string, string) (bool, error)
	Audit(context.Context, audit.Meta, string, any) error
	Release(context.Context) error
}

// Scheduler owns just the daily EdgeOne main-host purge. It never catches up
// missed slots or retries a submission, including ambiguous provider failures.
type Scheduler struct {
	cron     *cron.Cron
	entry    cron.EntryID
	location *time.Location
	request  infra.PurgeRequest
	service  scheduledPurger
	store    scheduledPurgeStore
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	started  bool
	stopped  bool
	running  bool
	lastSlot string
	jobs     sync.WaitGroup
	stopOnce sync.Once
}

func NewScheduler(cfg env.EdgeOneConfig, service *infra.Service, pool *pgxpool.Pool, logger *audit.Logger) (*Scheduler, error) {
	if !cfg.ScheduledPurge.Enabled {
		return nil, nil
	}
	if service == nil || pool == nil || logger == nil {
		return nil, errors.New("EdgeOne scheduled purge requires CloudOps, GFA pool and audit logger")
	}
	return newScheduler(cfg, service, &pgScheduledPurgeStore{pool: pool, logger: logger})
}

func newScheduler(cfg env.EdgeOneConfig, service scheduledPurger, store scheduledPurgeStore) (*Scheduler, error) {
	spec, location, err := cfg.ScheduledPurgeSchedule()
	if err != nil || !cfg.ScheduledPurge.Enabled {
		return nil, err
	}
	request, err := service.Validate("edgeone", infra.PurgeRequest{Type: "host", Targets: []string{cfg.MainHost}})
	if err != nil {
		return nil, fmt.Errorf("EdgeOne scheduled purge main_host: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := &Scheduler{location: location, request: request, service: service, store: store, ctx: ctx, cancel: cancel}
	s.cron = cron.New(cron.WithLocation(location))
	// cron records Prev before servicing Entry requests. Use that scheduled slot,
	// rather than the wall clock after a potentially delayed goroutine dispatch.
	s.entry, err = s.cron.AddFunc(spec, func() { s.runSlot(s.cron.Entry(s.entry).Prev) })
	if err != nil {
		cancel()
		return nil, fmt.Errorf("EdgeOne scheduled purge schedule: %w", err)
	}
	return s, nil
}

func (s *Scheduler) Start() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started || s.stopped {
		return
	}
	s.started = true
	s.cron.Start()
	entry := s.cron.Entry(s.entry)
	log.InfoKV("EdgeOne scheduled purge enabled", "host", s.request.Targets[0], "time", entry.Next.In(s.location).Format("15:04"), "timezone", s.location.String(), "next_run", entry.Next.Format(time.RFC3339))
}

// Stop cancels the remote call, then waits for final audit and session unlock.
// The owner must keep GFA and logging alive until this method returns.
func (s *Scheduler) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		s.mu.Lock()
		s.stopped = true
		s.cancel()
		s.mu.Unlock()
		<-s.cron.Stop().Done()
		s.jobs.Wait()
	})
}

func (s *Scheduler) runSlot(scheduledFor time.Time) {
	scheduledFor = scheduledFor.In(s.location)
	slot := "edgeone-scheduled-purge:" + scheduledFor.Format("20060102T1504-0700")
	s.mu.Lock()
	if s.stopped || s.running || s.lastSlot == slot || scheduledFor.IsZero() {
		s.mu.Unlock()
		return
	}
	s.running, s.lastSlot = true, slot
	s.jobs.Add(1)
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		s.jobs.Done()
	}()

	ctx, cancel := context.WithTimeout(s.ctx, 25*time.Second)
	defer cancel()
	lease, err := s.store.Acquire(ctx)
	if err != nil {
		log.ErrorKV("scheduled EdgeOne purge lock failed", "request_id", slot, "error", err)
		return
	}
	if lease == nil {
		log.InfoKV("scheduled EdgeOne purge skipped: lock held by another instance", "request_id", slot)
		return
	}
	defer func() {
		unlockCtx, unlockCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer unlockCancel()
		if err := lease.Release(unlockCtx); err != nil {
			log.ErrorKV("scheduled EdgeOne purge unlock failed", "request_id", slot, "error", err)
		}
	}()
	host := s.request.Targets[0]
	recorded, err := lease.Recorded(ctx, slot, host)
	if err != nil {
		log.ErrorKV("scheduled EdgeOne purge audit lookup failed", "request_id", slot, "error", err)
		return
	}
	if recorded {
		log.InfoKV("scheduled EdgeOne purge skipped: slot already requested", "request_id", slot, "host", host)
		return
	}
	meta := audit.SystemMeta(slot)
	intent := map[string]any{"status": "requested", "input": map[string]any{
		"type": "host", "targets": s.request.Targets, "scheduled_for": scheduledFor.Format(time.RFC3339),
	}}
	if err := lease.Audit(ctx, meta, host, intent); err != nil {
		log.ErrorKV("scheduled EdgeOne purge intent audit failed", "request_id", slot, "host", host, "error", err)
		return
	}
	result, err := s.service.Purge(ctx, "edgeone", s.request)
	if err == nil && (result.JobID == "" || result.Status != "submitted" || len(result.Failures) != 0) {
		err = errors.New("EdgeOne did not accept the scheduled host purge")
	}
	detail := map[string]any{"status": "completed", "result": result}
	if err != nil {
		detail = map[string]any{"status": "failed", "error": err.Error(), "result": result}
		log.ErrorKV("scheduled EdgeOne purge failed", "request_id", slot, "host", host, "error", err)
	} else {
		log.InfoKV("scheduled EdgeOne purge submitted", "request_id", slot, "host", host, "job_id", result.JobID)
	}
	// Cancellation must not discard the outcome audit. This independent, bounded
	// context also runs on shutdown while the Runtime still owns live DB pools.
	auditCtx, auditCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer auditCancel()
	if err := lease.Audit(auditCtx, meta, host, detail); err != nil {
		log.ErrorKV("scheduled EdgeOne purge result audit failed", "request_id", slot, "host", host, "job_id", result.JobID, "error", err)
	}
}
