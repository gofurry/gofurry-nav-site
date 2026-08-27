package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/models"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/report"
	"github.com/gofurry/gofurry-game-collector/common/log"
	cs "github.com/gofurry/gofurry-game-collector/common/service"
	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
	"github.com/gofurry/gofurry-game-collector/roof/env"
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
	gameHomeCacheZH  = "game:v2:home:zh:CN"
	gameHomeCacheEN  = "game:v2:home:en:CN"
)

type Executor interface {
	ResolveControlTargets(context.Context, *int64) ([]models.GameID, error)
	RunControlJob(context.Context, []models.GameID, []string, string, func(report.TaskResult)) (report.RunSummary, error)
}

type Worker struct {
	pool       *pgxpool.Pool
	queries    *gamesqlc.Queries
	executor   Executor
	instanceID string
	now        func() time.Time
	mu         sync.Mutex
}

type progressDocument struct {
	Expected  int       `json:"expected"`
	Attempted int       `json:"attempted"`
	Success   int       `json:"success"`
	Partial   int       `json:"partial"`
	Failed    int       `json:"failed"`
	Skipped   int       `json:"skipped"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewWorker(pool *pgxpool.Pool, executor Executor, instanceID string) *Worker {
	return &Worker{pool: pool, queries: gamesqlc.New(pool), executor: executor, instanceID: instanceID, now: time.Now}
}

func (w *Worker) ClaimAndRun(ctx context.Context) (bool, error) {
	// One callback per process is sufficient for the single Steam lane and also
	// avoids unnecessary local claim contention.
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, err := w.queries.SkipOverlappingGamePointInTimeJobs(ctx); err != nil {
		return false, err
	}

	tx, err := w.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	queries := gamesqlc.New(tx)
	instanceID := w.instanceID
	job, err := queries.ClaimNextGameCollectionJob(ctx, gamesqlc.ClaimNextGameCollectionJobParams{
		InstanceID: &instanceID, LeaseSeconds: leaseSeconds,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, tx.Commit(ctx)
	}
	var constraintErr *pgconn.PgError
	if errors.As(err, &constraintErr) && constraintErr.ConstraintName == "uq_gfg_collection_jobs_running_lane" {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var gameID *int64
	if job.ScopeType == "game" {
		gameID = job.ScopeID
	}
	targets, err := w.executor.ResolveControlTargets(ctx, gameID)
	if err != nil {
		return false, w.failUnstarted(ctx, tx, queries, job, "target_resolution", err.Error())
	}
	expected := len(targets) * len(job.Tasks)
	if expected == 0 {
		if _, err := queries.FinalizeGameCollectionJob(ctx, gamesqlc.FinalizeGameCollectionJobParams{
			Status: "skipped", ID: job.ID, InstanceID: &instanceID,
		}); err != nil {
			return false, err
		}
		return true, tx.Commit(ctx)
	}
	attempt, err := queries.NextGameCollectionAttempt(ctx, job.ID)
	if err != nil {
		return false, err
	}
	runID := "game-" + uuid.NewString()
	clock, err := queries.GameCollectionClock(ctx)
	if err != nil || !clock.Valid {
		return false, fmt.Errorf("read Game control-plane clock: %w", err)
	}
	delay := int64(0)
	if job.ScheduledFor.Valid {
		delay = maxInt64(0, clock.Time.Sub(job.ScheduledFor.Time).Milliseconds())
	}
	if _, err := queries.InsertGameCollectionRun(ctx, gamesqlc.InsertGameCollectionRunParams{
		ID: runID, JobID: job.ID, AttemptNo: attempt,
		CollectorInstanceID: w.instanceID, ScheduledFor: job.ScheduledFor,
		ExpectedCount: int32(expected), ScheduleDelayMs: delay,
	}); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	log.Info("Game durable job started, job_id=", job.ID, " run_id=", runID, " trigger=", job.Trigger, " scope=", job.ScopeType)
	return true, w.execute(ctx, job, targets, runID, expected)
}

func (w *Worker) failUnstarted(ctx context.Context, tx pgx.Tx, queries *gamesqlc.Queries, job gamesqlc.GfgCollectionJob, kind, message string) error {
	instanceID := w.instanceID
	if _, err := queries.FinalizeGameCollectionJob(ctx, gamesqlc.FinalizeGameCollectionJobParams{
		Status: "failed", ID: job.ID, InstanceID: &instanceID,
	}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return fmt.Errorf("%s: %s", kind, message)
}

func (w *Worker) execute(parent context.Context, job gamesqlc.GfgCollectionJob, targets []models.GameID, runID string, expected int) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	leaseDone := make(chan struct{})
	var leaseErr error
	go func() {
		leaseErr = w.maintainLease(ctx, cancel, job.ID)
		close(leaseDone)
	}()

	progress := progressDocument{Expected: expected, UpdatedAt: w.now()}
	var progressMu sync.Mutex
	var persistenceErr error
	onResult := func(result report.TaskResult) {
		persistCtx, persistCancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := w.saveTaskResult(persistCtx, result)
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
		applyProgress(&progress, result.Status)
		progress.UpdatedAt = w.now()
		snapshot := progress
		progressMu.Unlock()
		w.writeProgress(runID, snapshot)
	}
	summary, runErr := w.executor.RunControlJob(ctx, targets, job.Tasks, runID, onResult)
	wasCanceled := ctx.Err() != nil
	cancel()
	<-leaseDone
	maintenanceErr := leaseErr
	progressMu.Lock()
	finalProgress := progress
	storageErr := persistenceErr
	progressMu.Unlock()

	status := finalRunStatus(finalProgress, expected, wasCanceled && parent.Err() == nil)
	errorKind, errorMessage := summaryError(summary, runErr)
	if storageErr != nil {
		status, errorKind, errorMessage = "failed", "storage", storageErr.Error()
	}
	if maintenanceErr != nil && !errors.Is(maintenanceErr, context.Canceled) {
		status, errorKind, errorMessage = "failed", "lease", maintenanceErr.Error()
	}
	if parent.Err() != nil {
		status, errorKind, errorMessage = "canceled", "canceled", parent.Err().Error()
	}

	tx, err := w.pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	queries := gamesqlc.New(tx)
	if _, err := queries.FinalizeGameCollectionRun(context.Background(), gamesqlc.FinalizeGameCollectionRunParams{
		Status: status, AttemptedCount: int32(finalProgress.Attempted), SuccessCount: int32(finalProgress.Success),
		PartialCount: int32(finalProgress.Partial), FailureCount: int32(finalProgress.Failed),
		SkippedCount: int32(finalProgress.Skipped), ErrorKind: errorKind, ErrorMessage: errorMessage, ID: runID,
	}); err != nil {
		return err
	}
	instanceID := w.instanceID
	if _, err := queries.FinalizeGameCollectionJob(context.Background(), gamesqlc.FinalizeGameCollectionJobParams{
		Status: status, ID: job.ID, InstanceID: &instanceID,
	}); err != nil {
		return err
	}
	if err := tx.Commit(context.Background()); err != nil {
		return err
	}
	// Final Redis/cache state is refreshed only after the durable transaction.
	finalProgress.UpdatedAt = w.now()
	w.writeProgress(runID, finalProgress)
	if containsMetadataTask(job.Tasks) {
		_ = cs.Del(gameHomeCacheZH, gameHomeCacheEN)
	}
	log.Info("Game durable job finished, job_id=", job.ID, " run_id=", runID, " status=", status,
		" expected=", expected, " attempted=", finalProgress.Attempted)
	return errors.Join(runErr, storageErr, maintenanceErr)
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
			requested, err := w.queries.GameCollectionCancelRequested(ctx, gamesqlc.GameCollectionCancelRequestedParams{
				JobID: jobID, InstanceID: &instanceID,
			})
			if err != nil {
				cancel()
				return err
			}
			if requested.Valid {
				cancel()
				return context.Canceled
			}
			if time.Since(lastRenewed) >= leaseRenewEvery {
				rows, err := w.queries.RenewGameCollectionLease(ctx, gamesqlc.RenewGameCollectionLeaseParams{
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
	queries := gamesqlc.New(tx)
	rows, err := queries.ListExpiredGameCollectionJobs(ctx)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := queries.FailLostGameCollectionRun(ctx, row.RunID); err != nil {
			return err
		}
		if _, err := queries.FailLostGameCollectionJob(ctx, row.JobID); err != nil {
			return err
		}
		log.Warn("recovered expired Game collector lease, job_id=", row.JobID, " run_id=", row.RunID)
	}
	return tx.Commit(ctx)
}

func (w *Worker) saveTaskResult(ctx context.Context, result report.TaskResult) error {
	kind, message := "", ""
	if result.Error != nil {
		kind, message = string(result.Error.Kind), result.Error.Message
	}
	status := string(result.Status)
	if status == "" {
		status = "failed"
	}
	return w.queries.InsertGameCollectionTaskResult(ctx, gamesqlc.InsertGameCollectionTaskResultParams{
		RunID: result.RunID, TaskType: string(result.Task), GameID: result.GameID,
		Appid: int64(result.AppID), Status: status, UpstreamStatusCode: int32(result.UpstreamStatusCode),
		TrafficBucket: result.TrafficBucket, RetryCount: int32(result.RetryCount),
		DurationMs: result.DurationMillis, ErrorKind: kind, ErrorMessage: message,
		StartedAt: timestamp(result.StartedAt), EndedAt: timestamp(result.EndedAt),
	})
}

func (w *Worker) writeProgress(runID string, progress progressDocument) {
	raw, err := json.Marshal(progress)
	if err != nil {
		return
	}
	_ = cs.SetExpire("collection:game:run:"+runID+":progress", string(raw), progressCacheTTL)
}

func applyProgress(progress *progressDocument, status domain.Status) {
	switch status {
	case domain.StatusSuccess:
		progress.Success++
		progress.Attempted++
	case domain.StatusPartial:
		progress.Partial++
		progress.Attempted++
	case domain.StatusSkipped:
		progress.Skipped++
	default:
		progress.Failed++
		progress.Attempted++
	}
}

func finalRunStatus(progress progressDocument, expected int, canceled bool) string {
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

func summaryError(summary report.RunSummary, err error) (string, string) {
	if summary.Error != nil {
		return string(summary.Error.Kind), summary.Error.Message
	}
	if err != nil {
		return "execution", err.Error()
	}
	return "", ""
}

func containsMetadataTask(tasks []string) bool {
	for _, task := range tasks {
		if task == "details" || task == "news" {
			return true
		}
	}
	return false
}

func maxInt64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}

func collectorIdentity() (collectorID, hostname, version, commitSHA string) {
	hostname, _ = os.Hostname()
	if strings.TrimSpace(hostname) == "" {
		hostname = "unknown-host"
	}
	collectorID = hostname
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
