package facts

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gofurry/gofurry-game-collector/common/log"
	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	PlayerPipeline = "game.player_facts"
	StatePipeline  = "game.state_facts"
)

var pipelineOrder = []string{PlayerPipeline, StatePipeline}

type Options struct {
	ReconcileInterval time.Duration
	FinalizationGrace time.Duration
	RetentionEnabled  bool
	PlayerRawAge      time.Duration
	RetentionBatch    int32
	Now               func() time.Time
}

type Engine struct {
	pool    *pgxpool.Pool
	options Options
	cancel  context.CancelFunc
	wg      sync.WaitGroup
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
	Planned   []DayResult `json:"days"`
}

type BackfillOptions struct {
	Pipeline string
	From     *time.Time
	Through  *time.Time
	MaxDays  int
	DryRun   bool
}

func New(pool *pgxpool.Pool, options Options) *Engine {
	if options.ReconcileInterval <= 0 {
		options.ReconcileInterval = 10 * time.Minute
	}
	if options.FinalizationGrace <= 0 {
		options.FinalizationGrace = 30 * time.Minute
	}
	if options.PlayerRawAge <= 0 {
		options.PlayerRawAge = 90 * 24 * time.Hour
	}
	if options.RetentionBatch <= 0 {
		options.RetentionBatch = 500
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Engine{pool: pool, options: options}
}

func (engine *Engine) Start(parent context.Context) error {
	if engine == nil || engine.pool == nil {
		return errors.New("Game fact engine PostgreSQL pool is nil")
	}
	ctx, cancel := context.WithCancel(parent)
	engine.cancel = cancel
	if err := engine.Reconcile(ctx); err != nil {
		return fmt.Errorf("startup Game fact reconcile: %w", err)
	}
	engine.wg.Add(1)
	go func() {
		defer engine.wg.Done()
		ticker := time.NewTicker(engine.options.ReconcileInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := engine.Reconcile(ctx); err != nil && !errors.Is(err, context.Canceled) {
					log.Error("Game fact reconcile failed: " + err.Error())
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

// Reconcile processes at most one UTC day per pipeline. Retention runs only
// after the checkpoint transaction has committed.
func (engine *Engine) Reconcile(ctx context.Context) error {
	for _, pipeline := range pipelineOrder {
		result, err := engine.RunNext(ctx, pipeline, false)
		if err != nil {
			return err
		}
		if result.Processed {
			log.Info("Game facts finalized, pipeline=", pipeline, " fact_date=", result.FactDate)
		}
	}
	if engine.options.RetentionEnabled {
		_, err := engine.PrunePlayerRaw(ctx)
		return err
	}
	return nil
}

func (engine *Engine) Status(ctx context.Context) ([]CheckpointStatus, error) {
	rows, err := gamesqlc.New(engine.pool).ListGameFactCheckpoints(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]CheckpointStatus, 0, len(rows))
	latestClosed := engine.latestClosedDay()
	for _, row := range rows {
		result = append(result, CheckpointStatus{
			PipelineKey:      row.PipelineKey,
			Projection:       row.ProjectionVersion,
			SourceStartDate:  dateString(row.SourceStartDate),
			ProcessedThrough: dateString(row.ProcessedThrough),
			QualityCutoverAt: timestampString(row.QualityCutoverAt),
			LatestClosedDay:  latestClosed.Format(time.DateOnly),
			LagDays:          checkpointLagDays(row.SourceStartDate, row.ProcessedThrough, latestClosed),
		})
	}
	return result, nil
}

func (engine *Engine) RunNext(ctx context.Context, pipeline string, dryRun bool) (DayResult, error) {
	tx, err := engine.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return DayResult{}, err
	}
	defer tx.Rollback(ctx)
	queries := gamesqlc.New(tx)
	checkpoint, err := queries.LockGameFactCheckpoint(ctx, pipeline)
	if err != nil {
		return DayResult{}, err
	}
	next, ok := nextDate(checkpoint.SourceStartDate, checkpoint.ProcessedThrough)
	if !ok {
		return DayResult{PipelineKey: pipeline, Reason: "source_start_date is not available"}, tx.Commit(ctx)
	}
	result := DayResult{PipelineKey: pipeline, FactDate: next.Format(time.DateOnly)}
	finalizable, reason, err := engine.finalizable(ctx, queries, pipeline, next)
	if err != nil {
		return DayResult{}, err
	}
	if !finalizable {
		result.Reason = reason
		return result, tx.Commit(ctx)
	}
	if dryRun {
		result.Reason = "dry-run"
		return result, tx.Commit(ctx)
	}
	if err := projectDay(ctx, queries, pipeline, next); err != nil {
		return DayResult{}, err
	}
	rows, err := queries.AdvanceGameFactCheckpoint(ctx, gamesqlc.AdvanceGameFactCheckpointParams{
		PipelineKey: pipeline, ProcessedThrough: pgDate(next),
	})
	if err != nil {
		return DayResult{}, err
	}
	if rows != 1 {
		return DayResult{}, fmt.Errorf("advance %s checkpoint for %s: updated %d rows", pipeline, result.FactDate, rows)
	}
	if err := tx.Commit(ctx); err != nil {
		return DayResult{}, err
	}
	result.Processed = true
	return result, nil
}

func (engine *Engine) Backfill(ctx context.Context, dryRun bool) (BackfillSummary, error) {
	return engine.BackfillWithOptions(ctx, BackfillOptions{DryRun: dryRun})
}

func (engine *Engine) BackfillWithOptions(ctx context.Context, options BackfillOptions) (BackfillSummary, error) {
	pipelines, err := gamePipelines(options.Pipeline)
	if err != nil {
		return BackfillSummary{}, err
	}
	if options.From != nil && options.Through != nil && options.Through.Before(*options.From) {
		return BackfillSummary{}, errors.New("--to must not be before --from")
	}
	summary := BackfillSummary{DryRun: options.DryRun, Planned: make([]DayResult, 0)}
	if options.DryRun {
		remaining := options.MaxDays
		for _, pipeline := range pipelines {
			planned, err := engine.plan(ctx, pipeline, options.From, options.Through, remaining)
			if err != nil {
				return summary, err
			}
			summary.Planned = append(summary.Planned, planned...)
			if remaining > 0 {
				remaining -= len(planned)
				if remaining <= 0 {
					break
				}
			}
		}
		return summary, nil
	}
	remaining := options.MaxDays
	for _, pipeline := range pipelines {
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
			result, err := engine.RunNext(ctx, pipeline, false)
			if err != nil {
				return summary, err
			}
			if !result.Processed {
				break
			}
			summary.Processed++
			summary.Planned = append(summary.Planned, result)
			if remaining > 0 {
				remaining--
			}
		}
	}
	return summary, nil
}

func (engine *Engine) Rebuild(ctx context.Context, pipeline string, start, through time.Time, dryRun bool) (BackfillSummary, error) {
	selected, err := gamePipelines(pipeline)
	if err != nil || len(selected) != 1 {
		return BackfillSummary{}, fmt.Errorf("rebuild requires one Game fact pipeline, got %q", pipeline)
	}
	pipeline = selected[0]
	if through.Before(start) {
		return BackfillSummary{}, errors.New("--through must not be before --from")
	}
	summary := BackfillSummary{DryRun: dryRun, Planned: make([]DayResult, 0)}
	for day := utcDate(start); !day.After(utcDate(through)); day = day.AddDate(0, 0, 1) {
		result := DayResult{PipelineKey: pipeline, FactDate: day.Format(time.DateOnly)}
		if dryRun {
			if err := engine.validateRebuildDay(ctx, pipeline, day); err != nil {
				return summary, err
			}
			result.Reason = "dry-run rebuild"
			summary.Planned = append(summary.Planned, result)
			continue
		}
		if err := engine.rebuildDay(ctx, pipeline, day); err != nil {
			return summary, err
		}
		result.Processed = true
		summary.Processed++
		summary.Planned = append(summary.Planned, result)
	}
	return summary, nil
}

func (engine *Engine) validateRebuildDay(ctx context.Context, pipeline string, day time.Time) error {
	tx, err := engine.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	queries := gamesqlc.New(tx)
	checkpoint, err := queries.LockGameFactCheckpoint(ctx, pipeline)
	if err != nil {
		return err
	}
	if !checkpoint.ProcessedThrough.Valid || day.After(utcDate(checkpoint.ProcessedThrough.Time)) {
		return fmt.Errorf("cannot rebuild %s: day is beyond processed_through", day.Format(time.DateOnly))
	}
	finalizable, reason, err := engine.finalizable(ctx, queries, pipeline, day)
	if err != nil {
		return err
	}
	if !finalizable {
		return fmt.Errorf("cannot rebuild %s: %s", day.Format(time.DateOnly), reason)
	}
	if pipeline == PlayerPipeline {
		available, err := queries.GamePlayerRebuildSourceAvailable(ctx, pgDate(day))
		if err != nil {
			return err
		}
		if !available {
			return fmt.Errorf("cannot rebuild %s: scheduled Player source has been pruned", day.Format(time.DateOnly))
		}
	}
	return nil
}

func (engine *Engine) rebuildDay(ctx context.Context, pipeline string, day time.Time) error {
	tx, err := engine.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	queries := gamesqlc.New(tx)
	checkpoint, err := queries.LockGameFactCheckpoint(ctx, pipeline)
	if err != nil {
		return err
	}
	if !checkpoint.ProcessedThrough.Valid || day.After(utcDate(checkpoint.ProcessedThrough.Time)) {
		return fmt.Errorf("cannot rebuild %s: day is beyond processed_through", day.Format(time.DateOnly))
	}
	finalizable, reason, err := engine.finalizable(ctx, queries, pipeline, day)
	if err != nil {
		return err
	}
	if !finalizable {
		return fmt.Errorf("cannot rebuild %s: %s", day.Format(time.DateOnly), reason)
	}
	if pipeline == PlayerPipeline {
		available, sourceErr := queries.GamePlayerRebuildSourceAvailable(ctx, pgDate(day))
		if sourceErr != nil {
			return sourceErr
		}
		if !available {
			return fmt.Errorf("cannot rebuild %s: scheduled Player source has been pruned", day.Format(time.DateOnly))
		}
	}
	if err := projectDay(ctx, queries, pipeline, day); err != nil {
		return fmt.Errorf("rebuild %s %s (source may have been pruned): %w", pipeline, day.Format(time.DateOnly), err)
	}
	return tx.Commit(ctx)
}

func (engine *Engine) plan(ctx context.Context, pipeline string, from, through *time.Time, maxDays int) ([]DayResult, error) {
	queries := gamesqlc.New(engine.pool)
	checkpoint, err := queries.LockGameFactCheckpoint(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	next, ok := nextDate(checkpoint.SourceStartDate, checkpoint.ProcessedThrough)
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
		finalizable, reason, err := engine.finalizable(ctx, queries, pipeline, next)
		if err != nil {
			return nil, err
		}
		if !finalizable {
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
	checkpoint, err := gamesqlc.New(engine.pool).LockGameFactCheckpoint(ctx, pipeline)
	if err != nil {
		return time.Time{}, false, err
	}
	next, ok := nextDate(checkpoint.SourceStartDate, checkpoint.ProcessedThrough)
	return next, ok, nil
}

func gamePipelines(pipeline string) ([]string, error) {
	switch pipeline {
	case "":
		return pipelineOrder, nil
	case "player", PlayerPipeline:
		return []string{PlayerPipeline}, nil
	case "state", StatePipeline:
		return []string{StatePipeline}, nil
	default:
		return nil, fmt.Errorf("unknown Game fact pipeline %q", pipeline)
	}
}

func (engine *Engine) latestClosedDay() time.Time {
	return utcDate(engine.options.Now().UTC().Add(-engine.options.FinalizationGrace)).AddDate(0, 0, -1)
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

func (engine *Engine) finalizable(ctx context.Context, queries *gamesqlc.Queries, pipeline string, day time.Time) (bool, string, error) {
	start := utcDate(day)
	end := start.AddDate(0, 0, 1)
	if engine.options.Now().UTC().Before(end.Add(engine.options.FinalizationGrace)) {
		return false, "UTC day is not closed past finalization grace", nil
	}
	jobKey := "game.metadata"
	if pipeline == PlayerPipeline {
		jobKey = "game.players"
	}
	settled, err := queries.GameFactDaySettled(ctx, gamesqlc.GameFactDaySettledParams{
		JobKey: jobKey, DayStart: pgTimestamp(start), DayEnd: pgTimestamp(end),
	})
	if err != nil {
		return false, "", err
	}
	if settled == nil || !*settled {
		return false, "canonical scheduled acquisition is not settled", nil
	}
	return true, "", nil
}

func (engine *Engine) PrunePlayerRaw(ctx context.Context) (int64, error) {
	if !engine.options.RetentionEnabled {
		return 0, nil
	}
	return gamesqlc.New(engine.pool).PruneGamePlayerRawBatch(ctx, gamesqlc.PruneGamePlayerRawBatchParams{
		OlderThan: pgTimestamp(engine.options.Now().UTC().Add(-engine.options.PlayerRawAge)),
		BatchSize: engine.options.RetentionBatch,
	})
}

func projectDay(ctx context.Context, queries *gamesqlc.Queries, pipeline string, day time.Time) error {
	switch pipeline {
	case PlayerPipeline:
		_, err := queries.ProjectGamePlayerFactDay(ctx, pgDate(day))
		return err
	case StatePipeline:
		_, err := queries.ProjectGameStateFactDay(ctx, pgDate(day))
		return err
	default:
		return fmt.Errorf("unknown Game fact pipeline %q", pipeline)
	}
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

func utcDate(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func pgDate(value time.Time) pgtype.Date { return pgtype.Date{Time: utcDate(value), Valid: true} }
func pgTimestamp(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func dateString(value pgtype.Date) *string {
	if !value.Valid {
		return nil
	}
	text := utcDate(value.Time).Format(time.DateOnly)
	return &text
}

func timestampString(value pgtype.Timestamptz) *string {
	if !value.Valid {
		return nil
	}
	text := value.Time.UTC().Format(time.RFC3339)
	return &text
}
