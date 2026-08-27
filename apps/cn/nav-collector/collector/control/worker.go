package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gofurry/gofurry-nav-collector/collector/execution"
	"github.com/gofurry/gofurry-nav-collector/common/log"
	cs "github.com/gofurry/gofurry-nav-collector/common/service"
	navsqlc "github.com/gofurry/gofurry-nav-collector/internal/db/nav/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	leaseSeconds     = int64(60)
	leaseRenewEvery  = 18 * time.Second
	cancelPollEvery  = 2 * time.Second
	progressCacheTTL = 24 * time.Hour
)

type Worker struct {
	pool       *pgxpool.Pool
	queries    *navsqlc.Queries
	executor   *ProtocolExecutor
	instanceID string
	now        func() time.Time
	wg         sync.WaitGroup
}

type navProgress struct {
	Expected  int       `json:"expected"`
	Attempted int       `json:"attempted"`
	Success   int       `json:"success"`
	Partial   int       `json:"partial"`
	Failed    int       `json:"failed"`
	Skipped   int       `json:"skipped"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewWorker(pool *pgxpool.Pool, executor *ProtocolExecutor, instanceID string) *Worker {
	return &Worker{pool: pool, queries: navsqlc.New(pool), executor: executor, instanceID: instanceID, now: time.Now}
}

func (w *Worker) ClaimAndRun(ctx context.Context) (bool, error) {
	if _, err := w.queries.SkipOverlappingNavPointInTimeJobs(ctx); err != nil {
		return false, err
	}
	tx, err := w.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	queries := navsqlc.New(tx)
	instanceID := w.instanceID
	job, err := queries.ClaimNextNavCollectionJob(ctx, navsqlc.ClaimNextNavCollectionJobParams{
		InstanceID: &instanceID, LeaseSeconds: leaseSeconds,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, tx.Commit(ctx)
	}
	var constraintErr *pgconn.PgError
	if errors.As(err, &constraintErr) && constraintErr.ConstraintName == "uq_gfn_collection_jobs_running_lane" {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if len(job.Tasks) != 1 {
		if _, updateErr := queries.FinalizeNavCollectionJob(ctx, navsqlc.FinalizeNavCollectionJobParams{
			Status: "failed", ID: job.ID, InstanceID: &instanceID,
		}); updateErr != nil {
			return false, updateErr
		}
		return true, tx.Commit(ctx)
	}
	protocol := job.Tasks[0]
	if _, ok := capability("nav." + protocol); !ok {
		if _, updateErr := queries.FinalizeNavCollectionJob(ctx, navsqlc.FinalizeNavCollectionJobParams{
			Status: "failed", ID: job.ID, InstanceID: &instanceID,
		}); updateErr != nil {
			return false, updateErr
		}
		return true, tx.Commit(ctx)
	}
	targets, err := queries.ListNavCollectionTargets(ctx, navsqlc.ListNavCollectionTargetsParams{
		ScopeType: job.ScopeType, ScopeID: job.ScopeID, Target: job.Target,
	})
	if err != nil {
		return false, err
	}
	if len(targets) == 0 {
		if _, err := queries.FinalizeNavCollectionJob(ctx, navsqlc.FinalizeNavCollectionJobParams{
			Status: "skipped", ID: job.ID, InstanceID: &instanceID,
		}); err != nil {
			return false, err
		}
		return true, tx.Commit(ctx)
	}
	attempt, err := queries.NextNavCollectionAttempt(ctx, job.ID)
	if err != nil {
		return false, err
	}
	runID := "nav-" + protocol + "-" + uuid.NewString()
	clock, err := queries.NavCollectionClock(ctx)
	if err != nil || !clock.Valid {
		return false, fmt.Errorf("read Nav control-plane clock: %w", err)
	}
	delay := int64(0)
	if job.ScheduledFor.Valid {
		delay = maxInt64(0, clock.Time.Sub(job.ScheduledFor.Time).Milliseconds())
	}
	if _, err := queries.InsertNavCollectionRun(ctx, navsqlc.InsertNavCollectionRunParams{
		ID: runID, JobID: job.ID, AttemptNo: attempt, CollectorInstanceID: w.instanceID,
		ScheduledFor: job.ScheduledFor, ExpectedCount: int32(len(targets)), ScheduleDelayMs: delay,
	}); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		if err := w.execute(ctx, job, protocol, runID, len(targets)); err != nil && !errors.Is(err, context.Canceled) {
			log.ErrorFields(map[string]interface{}{"job_id": job.ID, "run_id": runID, "protocol": protocol}, "Nav durable job failed: "+err.Error())
		}
	}()
	return true, nil
}

func (w *Worker) execute(parent context.Context, job navsqlc.GfnCollectionJob, protocol, runID string, expected int) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	leaseDone := make(chan struct{})
	var leaseErr error
	go func() {
		leaseErr = w.maintainLease(ctx, cancel, job.ID)
		close(leaseDone)
	}()
	progress := navProgress{Expected: expected, UpdatedAt: w.now()}
	var progressMu sync.Mutex
	var persistenceErr error
	request := execution.Request{
		JobID: job.ID, RunID: runID, InstanceID: w.instanceID,
		ScopeType: job.ScopeType, ScopeID: job.ScopeID, Target: job.Target,
		OnResult: func(result execution.Result) {
			persistCtx, persistCancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := w.saveTaskResult(persistCtx, runID, result)
			persistCancel()
			if err != nil {
				progressMu.Lock()
				if persistenceErr == nil {
					persistenceErr = err
				}
				progressMu.Unlock()
				cancel()
			}
			progressMu.Lock()
			applyNavProgress(&progress, result.Status)
			progress.UpdatedAt = w.now()
			snapshot := progress
			progressMu.Unlock()
			w.writeProgress(runID, snapshot)
		},
	}
	runErr := w.executor.Run(ctx, protocol, request)
	wasCanceled := ctx.Err() != nil
	cancel()
	<-leaseDone
	progressMu.Lock()
	finalProgress := progress
	storageErr := persistenceErr
	progressMu.Unlock()
	status := finalNavRunStatus(finalProgress, expected, wasCanceled && parent.Err() == nil)
	errorKind, errorMessage := "", ""
	if runErr != nil {
		status, errorKind, errorMessage = "failed", "execution", runErr.Error()
	}
	if storageErr != nil {
		status, errorKind, errorMessage = "failed", "storage", storageErr.Error()
	}
	if leaseErr != nil && !errors.Is(leaseErr, context.Canceled) {
		status, errorKind, errorMessage = "failed", "lease", leaseErr.Error()
	}
	if parent.Err() != nil {
		status, errorKind, errorMessage = "canceled", "canceled", parent.Err().Error()
	}

	finalizeCtx := context.Background()
	tx, err := w.pool.BeginTx(finalizeCtx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(finalizeCtx)
	queries := navsqlc.New(tx)
	if _, err := queries.FinalizeNavCollectionRun(finalizeCtx, navsqlc.FinalizeNavCollectionRunParams{
		Status: status, AttemptedCount: int32(finalProgress.Attempted), SuccessCount: int32(finalProgress.Success),
		PartialCount: int32(finalProgress.Partial), FailureCount: int32(finalProgress.Failed),
		SkippedCount: int32(finalProgress.Skipped), ErrorKind: errorKind, ErrorMessage: errorMessage, ID: runID,
	}); err != nil {
		return err
	}
	instanceID := w.instanceID
	if _, err := queries.FinalizeNavCollectionJob(finalizeCtx, navsqlc.FinalizeNavCollectionJobParams{
		Status: status, ID: job.ID, InstanceID: &instanceID,
	}); err != nil {
		return err
	}
	if err := tx.Commit(finalizeCtx); err != nil {
		return err
	}
	finalProgress.UpdatedAt = w.now()
	w.writeProgress(runID, finalProgress)
	log.InfoFields(map[string]interface{}{
		"job_id": job.ID, "run_id": runID, "protocol": protocol, "status": status,
		"expected": expected, "attempted": finalProgress.Attempted,
	}, "Nav durable job finished")
	return errors.Join(runErr, storageErr, leaseErr)
}

func (w *Worker) maintainLease(ctx context.Context, cancel context.CancelFunc, jobID int64) error {
	ticker := time.NewTicker(cancelPollEvery)
	defer ticker.Stop()
	lastRenewed := time.Now()
	instanceID := w.instanceID
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			requested, err := w.queries.NavCollectionCancelRequested(ctx, navsqlc.NavCollectionCancelRequestedParams{JobID: jobID, InstanceID: &instanceID})
			if err != nil {
				cancel()
				return err
			}
			if requested.Valid {
				cancel()
				return context.Canceled
			}
			if time.Since(lastRenewed) >= leaseRenewEvery {
				rows, err := w.queries.RenewNavCollectionLease(ctx, navsqlc.RenewNavCollectionLeaseParams{
					LeaseSeconds: leaseSeconds, JobID: jobID, InstanceID: &instanceID,
				})
				if err != nil || rows != 1 {
					cancel()
					if err != nil {
						return err
					}
					return fmt.Errorf("lease ownership lost")
				}
				lastRenewed = time.Now()
			}
		}
	}
}

func (w *Worker) RecoverExpired(ctx context.Context) error {
	tx, err := w.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	queries := navsqlc.New(tx)
	rows, err := queries.ListExpiredNavCollectionJobs(ctx)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := queries.FailLostNavCollectionRun(ctx, row.RunID); err != nil {
			return err
		}
		if _, err := queries.FailLostNavCollectionJob(ctx, row.JobID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (w *Worker) Wait(ctx context.Context) error {
	done := make(chan struct{})
	go func() { w.wg.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *Worker) saveTaskResult(ctx context.Context, runID string, result execution.Result) error {
	return w.queries.InsertNavCollectionTaskResult(ctx, navsqlc.InsertNavCollectionTaskResultParams{
		RunID: runID, Protocol: result.Protocol, SiteID: result.SiteID, Target: result.Target,
		Status: result.Status, ObservationID: result.ObservationID, DurationMs: result.DurationMS,
		ErrorKind: result.ErrorKind, ErrorMessage: result.ErrorMessage,
		StartedAt: timestamp(result.StartedAt), EndedAt: timestamp(result.EndedAt),
	})
}

func (w *Worker) writeProgress(runID string, progress navProgress) {
	raw, err := json.Marshal(progress)
	if err == nil {
		_ = cs.SetExpire("collection:nav:run:"+runID+":progress", string(raw), progressCacheTTL)
	}
}

func applyNavProgress(progress *navProgress, status string) {
	switch status {
	case "success":
		progress.Success++
		progress.Attempted++
	case "partial":
		progress.Partial++
		progress.Attempted++
	case "skipped":
		progress.Skipped++
	default:
		progress.Failed++
		progress.Attempted++
	}
}

func finalNavRunStatus(progress navProgress, expected int, canceled bool) string {
	if canceled {
		return "canceled"
	}
	if expected > 0 && progress.Success == expected && progress.Partial == 0 && progress.Failed == 0 && progress.Skipped == 0 {
		return "success"
	}
	if progress.Success+progress.Partial > 0 {
		return "partial"
	}
	return "failed"
}

func maxInt64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
