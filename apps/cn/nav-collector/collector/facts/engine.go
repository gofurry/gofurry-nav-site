package facts

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gofurry/gofurry-nav-collector/common/log"
	navsqlc "github.com/gofurry/gofurry-nav-collector/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	TargetPipeline = "nav.target_facts"
	SitePipeline   = "nav.site_facts"
)

var pipelines = []string{TargetPipeline, SitePipeline}

type Options struct {
	ReconcileInterval time.Duration
	FinalizationGrace time.Duration
	RetentionEnabled  bool
	ObservationKeep   int32
	RetentionBatch    int32
	Now               func() time.Time
}

type Engine struct {
	pool   *pgxpool.Pool
	opts   Options
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type CheckpointStatus struct {
	PipelineKey      string  `json:"pipeline_key"`
	Projection       int32   `json:"projection_version"`
	SourceStartDate  *string `json:"source_start_date"`
	ProcessedThrough *string `json:"processed_through"`
	QualityCutoverAt *string `json:"quality_cutover_at"`
	LatestClosedDay  string  `json:"latest_closed_day"`
	LagDays          *int    `json:"lag_days"`
}

type DayResult struct {
	PipelineKey string `json:"pipeline_key"`
	FactDate    string `json:"fact_date,omitempty"`
	Processed   bool   `json:"processed"`
	Reason      string `json:"reason,omitempty"`
}

type BackfillSummary struct {
	DryRun    bool        `json:"dry_run"`
	Processed int         `json:"processed_days"`
	Days      []DayResult `json:"days"`
}

type BackfillOptions struct {
	Pipeline string
	From     *time.Time
	Through  *time.Time
	MaxDays  int
	DryRun   bool
}

func New(pool *pgxpool.Pool, opts Options) *Engine {
	if opts.ReconcileInterval <= 0 {
		opts.ReconcileInterval = 10 * time.Minute
	}
	if opts.FinalizationGrace <= 0 {
		opts.FinalizationGrace = 30 * time.Minute
	}
	if opts.ObservationKeep <= 0 {
		opts.ObservationKeep = 288
	}
	if opts.RetentionBatch <= 0 {
		opts.RetentionBatch = 500
	}
	return &Engine{pool: pool, opts: opts}
}

func (engine *Engine) Start(parent context.Context) error {
	if engine == nil || engine.pool == nil {
		return errors.New("Nav fact engine PostgreSQL pool is nil")
	}
	ctx, cancel := context.WithCancel(parent)
	engine.cancel = cancel
	if err := engine.Reconcile(ctx); err != nil {
		return fmt.Errorf("startup Nav fact reconcile: %w", err)
	}
	engine.wg.Add(1)
	go func() {
		defer engine.wg.Done()
		ticker := time.NewTicker(engine.opts.ReconcileInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := engine.Reconcile(ctx); err != nil && !errors.Is(err, context.Canceled) {
					log.Error("Nav fact reconcile failed: " + err.Error())
				}
			}
		}
	}()
	return nil
}

func (engine *Engine) Shutdown(ctx context.Context) error {
	if engine == nil {
		return nil
	}
	if engine.cancel != nil {
		engine.cancel()
	}
	done := make(chan struct{})
	go func() { engine.wg.Wait(); close(done) }()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

// Reconcile handles at most one UTC day per pipeline in dependency order.
func (engine *Engine) Reconcile(ctx context.Context) error {
	for _, pipeline := range pipelines {
		result, err := engine.RunNext(ctx, pipeline, false)
		if err != nil {
			return err
		}
		if result.Processed {
			log.InfoFields(map[string]interface{}{"pipeline": pipeline, "fact_date": result.FactDate}, "Nav facts finalized")
		}
	}
	if engine.opts.RetentionEnabled {
		_, err := engine.PruneObservations(ctx)
		return err
	}
	return nil
}

func (engine *Engine) Status(ctx context.Context) ([]CheckpointStatus, error) {
	queries := navsqlc.New(engine.pool)
	rows, err := queries.ListNavFactCheckpoints(ctx)
	if err != nil {
		return nil, err
	}
	now, err := engine.now(ctx, queries)
	if err != nil {
		return nil, err
	}
	result := make([]CheckpointStatus, 0, len(rows))
	latestClosed := engine.latestClosedDay(now)
	for _, row := range rows {
		result = append(result, CheckpointStatus{PipelineKey: row.PipelineKey, Projection: row.ProjectionVersion,
			SourceStartDate: dateText(row.SourceStartDate), ProcessedThrough: dateText(row.ProcessedThrough),
			QualityCutoverAt: timestampText(row.QualityCutoverAt), LatestClosedDay: latestClosed.Format(time.DateOnly),
			LagDays: checkpointLagDays(row.SourceStartDate, row.ProcessedThrough, latestClosed)})
	}
	return result, nil
}

func (engine *Engine) RunNext(ctx context.Context, pipeline string, dryRun bool) (DayResult, error) {
	tx, err := engine.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return DayResult{}, err
	}
	defer tx.Rollback(ctx)
	queries := navsqlc.New(tx)
	cp, err := queries.LockNavFactCheckpoint(ctx, pipeline)
	if err != nil {
		return DayResult{}, err
	}
	day, ok := nextDate(cp.SourceStartDate, cp.ProcessedThrough)
	if !ok {
		return DayResult{PipelineKey: pipeline, Reason: "source_start_date is not available"}, tx.Commit(ctx)
	}
	result := DayResult{PipelineKey: pipeline, FactDate: day.Format(time.DateOnly)}
	ready, reason, err := engine.finalizable(ctx, queries, pipeline, day)
	if err != nil {
		return DayResult{}, err
	}
	if !ready {
		result.Reason = reason
		return result, tx.Commit(ctx)
	}
	if dryRun {
		result.Reason = "dry-run"
		return result, tx.Commit(ctx)
	}
	if err := engine.project(ctx, queries, cp, pipeline, day); err != nil {
		return DayResult{}, err
	}
	changed, err := queries.AdvanceNavFactCheckpoint(ctx, navsqlc.AdvanceNavFactCheckpointParams{PipelineKey: pipeline, ProcessedThrough: pgDate(day)})
	if err != nil {
		return DayResult{}, err
	}
	if changed != 1 {
		return DayResult{}, fmt.Errorf("advance %s checkpoint for %s: updated %d rows", pipeline, result.FactDate, changed)
	}
	if err := tx.Commit(ctx); err != nil {
		return DayResult{}, err
	}
	result.Processed = true
	return result, nil
}

func (engine *Engine) project(ctx context.Context, queries *navsqlc.Queries, cp navsqlc.GfnFactRollupCheckpoint, pipeline string, day time.Time) error {
	switch pipeline {
	case TargetPipeline:
		start, end := utcDate(day), utcDate(day).AddDate(0, 0, 1)
		inputs, err := queries.ListNavProtocolDailyInputs(ctx, navsqlc.ListNavProtocolDailyInputsParams{DayStart: pgTimestamp(start), DayEnd: pgTimestamp(end)})
		if err != nil {
			return err
		}
		ledger := cp.QualityCutoverAt.Valid && !start.Before(cp.QualityCutoverAt.Time.UTC())
		for _, input := range inputs {
			knownState, err := projectKnownState(input.Protocol, input.KnownPayload)
			if err != nil {
				return err
			}
			params := navsqlc.UpsertNavProtocolDailyParams{
				TargetTrackingPeriodID: input.TargetTrackingPeriodID, SiteID: input.SiteID,
				CollectorDomainID: input.CollectorDomainID, Target: input.Target, Protocol: input.Protocol,
				FactDate: pgDate(day), AttemptedCount: input.AttemptedCount,
				SuccessCount: input.SuccessCount, PartialCount: input.PartialCount, FailureCount: input.FailureCount,
				FailureKindCounts: input.FailureKindCounts, LatestScheduledStatus: optionalString(input.LatestScheduledStatus),
				LatestScheduledAt: input.LatestScheduledAt, LatestObservationStatus: optionalString(input.LatestObservationStatus),
				LatestObservationAt: input.LatestObservationAt, KnownStateObservedAt: input.KnownStateObservedAt,
				KnownState: knownState, AvgDurationMs: input.AvgDurationMs, P95DurationMs: input.P95DurationMs,
			}
			if ledger {
				params.QualityBasis = "acquisition_ledger"
				params.ExpectedCount = int32Pointer(input.ExpectedCount)
				params.SkippedCount = int32Pointer(input.SkippedCount)
				params.MissedCount = int32Pointer(input.MissedCount)
				params.CanceledCount = int32Pointer(input.CanceledCount)
				params.UnattemptedCount = int32Pointer(input.UnattemptedCount)
			} else {
				params.QualityBasis = "legacy_observed_only"
			}
			if err := queries.UpsertNavProtocolDaily(ctx, params); err != nil {
				return err
			}
		}
		_, err = queries.ProjectNavTargetDaily(ctx, pgDate(day))
		return err
	case SitePipeline:
		_, err := queries.ProjectNavSiteDaily(ctx, pgDate(day))
		return err
	default:
		return fmt.Errorf("unknown Nav fact pipeline %q", pipeline)
	}
}

func (engine *Engine) Backfill(ctx context.Context, dryRun bool) (BackfillSummary, error) {
	return engine.BackfillWithOptions(ctx, BackfillOptions{DryRun: dryRun})
}

func (engine *Engine) BackfillWithOptions(ctx context.Context, options BackfillOptions) (BackfillSummary, error) {
	selected, err := navPipelines(options.Pipeline)
	if err != nil {
		return BackfillSummary{}, err
	}
	if options.From != nil && options.Through != nil && options.Through.Before(*options.From) {
		return BackfillSummary{}, errors.New("--to must not be before --from")
	}
	summary := BackfillSummary{DryRun: options.DryRun, Days: make([]DayResult, 0)}
	if options.DryRun {
		remaining := options.MaxDays
		for _, pipeline := range selected {
			days, err := engine.plan(ctx, pipeline, options.From, options.Through, remaining)
			if err != nil {
				return summary, err
			}
			summary.Days = append(summary.Days, days...)
			if remaining > 0 {
				remaining -= len(days)
				if remaining <= 0 {
					break
				}
			}
		}
		return summary, nil
	}
	remaining := options.MaxDays
	for _, pipeline := range selected {
		for {
			next, ok, err := engine.nextPipelineDate(ctx, pipeline)
			if err != nil {
				return summary, err
			}
			if !ok {
				break
			}
			if options.From != nil && next.Before(utcDate(*options.From)) {
				return summary, fmt.Errorf("cannot start %s at %s: ordered checkpoint requires preceding day %s", pipeline, options.From.Format(time.DateOnly), next.Format(time.DateOnly))
			}
			if options.Through != nil && next.After(utcDate(*options.Through)) {
				break
			}
			if remaining == 0 && options.MaxDays > 0 {
				return summary, nil
			}
			day, err := engine.RunNext(ctx, pipeline, false)
			if err != nil {
				return summary, err
			}
			if !day.Processed {
				break
			}
			summary.Processed++
			summary.Days = append(summary.Days, day)
			if remaining > 0 {
				remaining--
			}
		}
	}
	return summary, nil
}

func (engine *Engine) plan(ctx context.Context, pipeline string, from, through *time.Time, maxDays int) ([]DayResult, error) {
	queries := navsqlc.New(engine.pool)
	cp, err := queries.LockNavFactCheckpoint(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	next, ok := nextDate(cp.SourceStartDate, cp.ProcessedThrough)
	if !ok {
		return nil, nil
	}
	if from != nil && next.Before(utcDate(*from)) {
		return nil, fmt.Errorf("cannot start %s at %s: ordered checkpoint requires preceding day %s", pipeline, from.Format(time.DateOnly), next.Format(time.DateOnly))
	}
	result := make([]DayResult, 0)
	for {
		if through != nil && next.After(utcDate(*through)) {
			return result, nil
		}
		if maxDays > 0 && len(result) >= maxDays {
			return result, nil
		}
		ready, reason, err := engine.finalizable(ctx, queries, pipeline, next)
		if err != nil {
			return nil, err
		}
		if !ready {
			if len(result) == 0 {
				result = append(result, DayResult{PipelineKey: pipeline, FactDate: next.Format(time.DateOnly), Reason: reason})
			}
			return result, nil
		}
		result = append(result, DayResult{PipelineKey: pipeline, FactDate: next.Format(time.DateOnly), Reason: "dry-run"})
		next = next.AddDate(0, 0, 1)
	}
}

func (engine *Engine) nextPipelineDate(ctx context.Context, pipeline string) (time.Time, bool, error) {
	cp, err := navsqlc.New(engine.pool).LockNavFactCheckpoint(ctx, pipeline)
	if err != nil {
		return time.Time{}, false, err
	}
	next, ok := nextDate(cp.SourceStartDate, cp.ProcessedThrough)
	return next, ok, nil
}

func (engine *Engine) previewNext(ctx context.Context, pipeline string) (DayResult, error) {
	tx, err := engine.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return DayResult{}, err
	}
	defer tx.Rollback(ctx)
	queries := navsqlc.New(tx)
	cp, err := queries.LockNavFactCheckpoint(ctx, pipeline)
	if err != nil {
		return DayResult{}, err
	}
	day, ok := nextDate(cp.SourceStartDate, cp.ProcessedThrough)
	if !ok {
		return DayResult{PipelineKey: pipeline, Reason: "source_start_date is not available"}, nil
	}
	ready, reason, err := engine.finalizable(ctx, queries, pipeline, day)
	if err != nil {
		return DayResult{}, err
	}
	if ready {
		reason = "dry-run"
	}
	return DayResult{PipelineKey: pipeline, FactDate: day.Format(time.DateOnly), Reason: reason}, nil
}

func (engine *Engine) Rebuild(ctx context.Context, pipeline string, start, through time.Time, dryRun bool) (BackfillSummary, error) {
	selected, err := navPipelines(pipeline)
	if err != nil || len(selected) != 1 {
		return BackfillSummary{}, fmt.Errorf("rebuild requires one Nav fact pipeline, got %q", pipeline)
	}
	pipeline = selected[0]
	if through.Before(start) {
		return BackfillSummary{}, errors.New("--through must not be before --from")
	}
	summary := BackfillSummary{DryRun: dryRun, Days: make([]DayResult, 0)}
	for day := utcDate(start); !day.After(utcDate(through)); day = day.AddDate(0, 0, 1) {
		result := DayResult{PipelineKey: pipeline, FactDate: day.Format(time.DateOnly)}
		if dryRun {
			if err := engine.validateRebuildDay(ctx, pipeline, day); err != nil {
				return summary, err
			}
			result.Reason = "dry-run rebuild"
			summary.Days = append(summary.Days, result)
			continue
		}
		tx, err := engine.pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return summary, err
		}
		queries := navsqlc.New(tx)
		cp, err := queries.LockNavFactCheckpoint(ctx, pipeline)
		if err == nil && (!cp.ProcessedThrough.Valid || day.After(utcDate(cp.ProcessedThrough.Time))) {
			err = fmt.Errorf("day is beyond processed_through")
		}
		if err == nil {
			var ok bool
			var reason string
			ok, reason, err = engine.finalizable(ctx, queries, pipeline, day)
			if err == nil && !ok {
				err = errors.New(reason)
			}
		}
		if err == nil && pipeline == TargetPipeline {
			var available bool
			available, err = queries.NavTargetRebuildSourceAvailable(ctx, pgDate(day))
			if err == nil && !available {
				err = errors.New("scheduled/observation source has been pruned")
			}
		}
		if err == nil {
			err = engine.project(ctx, queries, cp, pipeline, day)
		}
		if err == nil {
			err = tx.Commit(ctx)
		} else {
			_ = tx.Rollback(ctx)
		}
		if err != nil {
			return summary, fmt.Errorf("rebuild %s %s (source may have been pruned): %w", pipeline, result.FactDate, err)
		}
		result.Processed = true
		summary.Processed++
		summary.Days = append(summary.Days, result)
	}
	return summary, nil
}

func (engine *Engine) validateRebuildDay(ctx context.Context, pipeline string, day time.Time) error {
	tx, err := engine.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	queries := navsqlc.New(tx)
	cp, err := queries.LockNavFactCheckpoint(ctx, pipeline)
	if err != nil {
		return err
	}
	if !cp.ProcessedThrough.Valid || day.After(utcDate(cp.ProcessedThrough.Time)) {
		return fmt.Errorf("rebuild %s %s: day is beyond processed_through", pipeline, day.Format(time.DateOnly))
	}
	ready, reason, err := engine.finalizable(ctx, queries, pipeline, day)
	if err != nil {
		return err
	}
	if !ready {
		return fmt.Errorf("rebuild %s %s: %s", pipeline, day.Format(time.DateOnly), reason)
	}
	if pipeline == TargetPipeline {
		available, err := queries.NavTargetRebuildSourceAvailable(ctx, pgDate(day))
		if err != nil {
			return err
		}
		if !available {
			return fmt.Errorf("rebuild %s %s: scheduled/observation source has been pruned", pipeline, day.Format(time.DateOnly))
		}
	}
	return nil
}

func (engine *Engine) finalizable(ctx context.Context, queries *navsqlc.Queries, pipeline string, day time.Time) (bool, string, error) {
	start, end := utcDate(day), utcDate(day).AddDate(0, 0, 1)
	now, err := engine.now(ctx, queries)
	if err != nil {
		return false, "", err
	}
	if now.UTC().Before(end.Add(engine.opts.FinalizationGrace)) {
		return false, "UTC day is not closed past finalization grace", nil
	}
	if pipeline == TargetPipeline {
		settled, err := queries.NavTargetFactDaySettled(ctx, navsqlc.NavTargetFactDaySettledParams{DayStart: pgTimestamp(start), DayEnd: pgTimestamp(end)})
		if err != nil {
			return false, "", err
		}
		if settled == nil || !*settled {
			return false, "canonical scheduled acquisition is not settled", nil
		}
		return true, "", nil
	}
	ready, err := queries.NavSiteFactDependencyReady(ctx, pgDate(day))
	if err != nil {
		return false, "", err
	}
	if ready == nil || !*ready {
		return false, "nav.target_facts dependency is not finalized", nil
	}
	return true, "", nil
}

func (engine *Engine) PruneObservations(ctx context.Context) (int64, error) {
	if !engine.opts.RetentionEnabled {
		return 0, nil
	}
	return navsqlc.New(engine.pool).PruneNavObservationsBatch(ctx, navsqlc.PruneNavObservationsBatchParams{KeepCount: engine.opts.ObservationKeep, BatchSize: engine.opts.RetentionBatch})
}

func navPipelines(pipeline string) ([]string, error) {
	switch pipeline {
	case "":
		return pipelines, nil
	case "target", TargetPipeline:
		return []string{TargetPipeline}, nil
	case "site", SitePipeline:
		return []string{SitePipeline}, nil
	default:
		return nil, fmt.Errorf("unknown Nav fact pipeline %q", pipeline)
	}
}

func (engine *Engine) now(ctx context.Context, queries *navsqlc.Queries) (time.Time, error) {
	if engine.opts.Now != nil {
		return engine.opts.Now().UTC(), nil
	}
	clock, err := queries.NavCollectionClock(ctx)
	if err != nil {
		return time.Time{}, err
	}
	if !clock.Valid {
		return time.Time{}, errors.New("Nav database clock is unavailable")
	}
	return clock.Time.UTC(), nil
}

func (engine *Engine) latestClosedDay(now time.Time) time.Time {
	return utcDate(now.UTC().Add(-engine.opts.FinalizationGrace)).AddDate(0, 0, -1)
}

func checkpointLagDays(source, processed pgtype.Date, latestClosed time.Time) *int {
	var next time.Time
	if processed.Valid {
		next = utcDate(processed.Time).AddDate(0, 0, 1)
	} else if source.Valid {
		next = utcDate(source.Time)
	} else {
		return nil
	}
	lag := int(latestClosed.Sub(next).Hours()/24) + 1
	if lag < 0 {
		lag = 0
	}
	return &lag
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
func int32Pointer(value int32) *int32 { return &value }
func utcDate(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
func pgDate(value time.Time) pgtype.Date { return pgtype.Date{Time: utcDate(value), Valid: true} }
func pgTimestamp(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}
func nextDate(source, processed pgtype.Date) (time.Time, bool) {
	if !source.Valid {
		return time.Time{}, false
	}
	if processed.Valid {
		return utcDate(processed.Time).AddDate(0, 0, 1), true
	}
	return utcDate(source.Time), true
}
func dateText(value pgtype.Date) *string {
	if !value.Valid {
		return nil
	}
	text := utcDate(value.Time).Format(time.DateOnly)
	return &text
}
func timestampText(value pgtype.Timestamptz) *string {
	if !value.Valid {
		return nil
	}
	text := value.Time.UTC().Format(time.RFC3339)
	return &text
}
