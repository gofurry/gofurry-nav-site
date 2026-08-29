package metrics

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

type Options struct {
	ReconcileInterval time.Duration
}

type Engine struct {
	pool    *pgxpool.Pool
	options Options
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

type Status struct {
	MetricKey                string  `json:"metric_key"`
	MetricVersion            int32   `json:"metric_version"`
	Status                   string  `json:"status"`
	SourceStartDate          *string `json:"source_start_date"`
	ProcessedThrough         *string `json:"processed_through"`
	UpstreamProcessedThrough *string `json:"upstream_processed_through"`
	LatestAvailableDay       *string `json:"latest_available_day"`
	LagDays                  *int    `json:"lag_days"`
}

type DayResult struct {
	MetricKey          string `json:"metric_key"`
	MetricVersion      int32  `json:"metric_version"`
	FactDate           string `json:"fact_date,omitempty"`
	Processed          bool   `json:"processed"`
	PopulationEstimate int64  `json:"population_estimate,omitempty"`
	Reason             string `json:"reason,omitempty"`
}

type BackfillOptions struct {
	Metric  string
	Version int32
	From    *time.Time
	Through *time.Time
	MaxDays int
	DryRun  bool
}

type BackfillSummary struct {
	DryRun    bool        `json:"dry_run"`
	Processed int         `json:"processed_days"`
	Days      []DayResult `json:"days"`
}

type registeredMetric struct {
	Contract Contract
	Status   string
}

func New(pool *pgxpool.Pool, options Options) *Engine {
	if options.ReconcileInterval <= 0 {
		options.ReconcileInterval = 10 * time.Minute
	}
	return &Engine{pool: pool, options: options}
}

func (engine *Engine) Start(parent context.Context) error {
	if err := engine.ValidateCatalog(parent); err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(parent)
	engine.cancel = cancel
	if err := engine.Reconcile(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("startup Game metric reconcile failed without stopping Collector: " + err.Error())
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
					log.Error("Game metric reconcile failed without stopping Collector: " + err.Error())
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

// Reconcile processes at most one UTC day for every active metric. A failed
// derived metric is reported after all other active metrics had a chance to run.
func (engine *Engine) Reconcile(ctx context.Context) error {
	metrics, err := engine.resolveMetrics(ctx, "", 0, true)
	if err != nil {
		return err
	}
	var reconcileErr error
	for _, metric := range metrics {
		result, runErr := engine.RunNext(ctx, metric.Contract.Key, metric.Contract.Version, false)
		if runErr != nil {
			reconcileErr = errors.Join(reconcileErr, fmt.Errorf("%s: %w", catalogID(metric.Contract.Key, metric.Contract.Version), runErr))
			continue
		}
		if result.Processed {
			log.Info("Game metric finalized, metric=", result.MetricKey, " version=", result.MetricVersion, " fact_date=", result.FactDate)
		}
	}
	return reconcileErr
}

func (engine *Engine) Status(ctx context.Context) ([]Status, error) {
	if err := engine.ValidateCatalog(ctx); err != nil {
		return nil, err
	}
	rows, err := gamesqlc.New(engine.pool).ListGameMetricCheckpoints(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Status, 0, len(rows))
	for _, row := range rows {
		result = append(result, Status{
			MetricKey: row.MetricKey, MetricVersion: row.MetricVersion, Status: row.Status,
			SourceStartDate: dateString(row.SourceStartDate), ProcessedThrough: dateString(row.ProcessedThrough),
			UpstreamProcessedThrough: dateString(row.UpstreamProcessedThrough), LatestAvailableDay: dateString(row.UpstreamProcessedThrough),
			LagDays: metricLagDays(row.SourceStartDate, row.ProcessedThrough, row.UpstreamProcessedThrough),
		})
	}
	return result, nil
}

func (engine *Engine) RunNext(ctx context.Context, metricKey string, metricVersion int32, dryRun bool) (DayResult, error) {
	if _, ok := contractFor(metricKey, metricVersion); !ok {
		return DayResult{}, fmt.Errorf("Game metric evaluator is not compiled for %s", catalogID(metricKey, metricVersion))
	}
	tx, err := engine.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return DayResult{}, err
	}
	defer tx.Rollback(ctx)
	queries := gamesqlc.New(tx)
	checkpoint, err := queries.LockGameMetricCheckpoint(ctx, gamesqlc.LockGameMetricCheckpointParams{MetricKey: metricKey, MetricVersion: metricVersion})
	if err != nil {
		return DayResult{}, err
	}
	if checkpoint.Status != "active" {
		return DayResult{MetricKey: metricKey, MetricVersion: metricVersion, Reason: "metric version is not active"}, tx.Commit(ctx)
	}
	next, ok := nextDate(checkpoint.SourceStartDate, checkpoint.ProcessedThrough)
	if !ok {
		return DayResult{MetricKey: metricKey, MetricVersion: metricVersion, Reason: "source_start_date is not available"}, tx.Commit(ctx)
	}
	result := DayResult{MetricKey: metricKey, MetricVersion: metricVersion, FactDate: next.Format(time.DateOnly)}
	upstream, err := queries.GameMetricUpstreamProcessedThrough(ctx)
	if err != nil {
		return DayResult{}, err
	}
	if !upstream.Valid || next.After(utcDate(upstream.Time)) {
		result.Reason = "upstream game.state_facts watermark is not ready"
		return result, tx.Commit(ctx)
	}
	result.PopulationEstimate, err = queries.CountGameMetricPopulation(ctx, pgDate(next))
	if err != nil {
		return DayResult{}, err
	}
	if dryRun {
		result.Reason = "dry-run"
		return result, tx.Commit(ctx)
	}
	if _, err := queries.ProjectGameMetricDay(ctx, gamesqlc.ProjectGameMetricDayParams{
		MetricKey: metricKey, MetricVersion: metricVersion, FactDate: pgDate(next),
	}); err != nil {
		return DayResult{}, err
	}
	rows, err := queries.AdvanceGameMetricCheckpoint(ctx, gamesqlc.AdvanceGameMetricCheckpointParams{
		MetricKey: metricKey, MetricVersion: metricVersion, ProcessedThrough: pgDate(next),
	})
	if err != nil {
		return DayResult{}, err
	}
	if rows != 1 {
		return DayResult{}, fmt.Errorf("advance Game metric checkpoint %s on %s: updated %d rows", catalogID(metricKey, metricVersion), result.FactDate, rows)
	}
	if err := tx.Commit(ctx); err != nil {
		return DayResult{}, err
	}
	result.Processed = true
	return result, nil
}

func (engine *Engine) Backfill(ctx context.Context, options BackfillOptions) (BackfillSummary, error) {
	if err := engine.ValidateCatalog(ctx); err != nil {
		return BackfillSummary{}, err
	}
	if options.MaxDays < 0 {
		return BackfillSummary{}, errors.New("--max-days must not be negative")
	}
	if options.From != nil && options.Through != nil && options.Through.Before(*options.From) {
		return BackfillSummary{}, errors.New("--to must not be before --from")
	}
	metrics, err := engine.resolveMetrics(ctx, options.Metric, options.Version, true)
	if err != nil {
		return BackfillSummary{}, err
	}
	summary := BackfillSummary{DryRun: options.DryRun, Days: make([]DayResult, 0)}
	remaining := options.MaxDays
	for _, metric := range metrics {
		if options.DryRun {
			planned, planErr := engine.plan(ctx, metric.Contract, options.From, options.Through, remaining)
			if planErr != nil {
				return summary, planErr
			}
			summary.Days = append(summary.Days, planned...)
			if remaining > 0 {
				remaining -= len(planned)
				if remaining <= 0 {
					break
				}
			}
			continue
		}
		for {
			if remaining == 0 && options.MaxDays > 0 {
				return summary, nil
			}
			next, ok, nextErr := engine.nextMetricDate(ctx, metric.Contract)
			if nextErr != nil {
				return summary, nextErr
			}
			if !ok {
				break
			}
			if options.From != nil && next.Before(utcDate(*options.From)) {
				return summary, fmt.Errorf("cannot start %s at %s: ordered checkpoint requires preceding day %s", catalogID(metric.Contract.Key, metric.Contract.Version), options.From.Format(time.DateOnly), next.Format(time.DateOnly))
			}
			if options.Through != nil && next.After(utcDate(*options.Through)) {
				break
			}
			day, runErr := engine.RunNext(ctx, metric.Contract.Key, metric.Contract.Version, false)
			if runErr != nil {
				return summary, runErr
			}
			summary.Days = append(summary.Days, day)
			if !day.Processed {
				break
			}
			summary.Processed++
			if remaining > 0 {
				remaining--
			}
		}
	}
	return summary, nil
}

func (engine *Engine) Rebuild(ctx context.Context, metricKey string, metricVersion int32, start, through time.Time, maxDays int, dryRun bool) (BackfillSummary, error) {
	if err := engine.ValidateCatalog(ctx); err != nil {
		return BackfillSummary{}, err
	}
	if _, err := engine.resolveMetrics(ctx, metricKey, metricVersion, false); err != nil {
		return BackfillSummary{}, err
	}
	if through.Before(start) {
		return BackfillSummary{}, errors.New("--through must not be before --from")
	}
	if maxDays < 0 {
		return BackfillSummary{}, errors.New("--max-days must not be negative")
	}
	summary := BackfillSummary{DryRun: dryRun, Days: make([]DayResult, 0)}
	for day := utcDate(start); !day.After(utcDate(through)); day = day.AddDate(0, 0, 1) {
		if maxDays > 0 && len(summary.Days) >= maxDays {
			break
		}
		result, err := engine.rebuildDay(ctx, metricKey, metricVersion, day, dryRun)
		if err != nil {
			return summary, err
		}
		summary.Days = append(summary.Days, result)
		if result.Processed {
			summary.Processed++
		}
	}
	return summary, nil
}

func (engine *Engine) rebuildDay(ctx context.Context, metricKey string, metricVersion int32, day time.Time, dryRun bool) (DayResult, error) {
	tx, err := engine.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return DayResult{}, err
	}
	defer tx.Rollback(ctx)
	queries := gamesqlc.New(tx)
	checkpoint, err := queries.LockGameMetricCheckpoint(ctx, gamesqlc.LockGameMetricCheckpointParams{MetricKey: metricKey, MetricVersion: metricVersion})
	if err != nil {
		return DayResult{}, err
	}
	result := DayResult{MetricKey: metricKey, MetricVersion: metricVersion, FactDate: day.Format(time.DateOnly)}
	if !checkpoint.SourceStartDate.Valid || day.Before(utcDate(checkpoint.SourceStartDate.Time)) {
		return DayResult{}, fmt.Errorf("cannot rebuild %s: day is before source_start_date", result.FactDate)
	}
	if !checkpoint.ProcessedThrough.Valid || day.After(utcDate(checkpoint.ProcessedThrough.Time)) {
		return DayResult{}, fmt.Errorf("cannot rebuild %s: day is beyond processed_through", result.FactDate)
	}
	upstream, err := queries.GameMetricUpstreamProcessedThrough(ctx)
	if err != nil {
		return DayResult{}, err
	}
	if !upstream.Valid || day.After(utcDate(upstream.Time)) {
		return DayResult{}, fmt.Errorf("cannot rebuild %s: upstream game.state_facts watermark is not ready", result.FactDate)
	}
	result.PopulationEstimate, err = queries.CountGameMetricPopulation(ctx, pgDate(day))
	if err != nil {
		return DayResult{}, err
	}
	if dryRun {
		result.Reason = "dry-run rebuild"
		return result, tx.Commit(ctx)
	}
	if _, err := queries.ProjectGameMetricDay(ctx, gamesqlc.ProjectGameMetricDayParams{
		MetricKey: metricKey, MetricVersion: metricVersion, FactDate: pgDate(day),
	}); err != nil {
		return DayResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return DayResult{}, err
	}
	result.Processed = true
	return result, nil
}

func (engine *Engine) plan(ctx context.Context, contract Contract, from, through *time.Time, maxDays int) ([]DayResult, error) {
	tx, err := engine.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	queries := gamesqlc.New(tx)
	checkpoint, err := queries.LockGameMetricCheckpoint(ctx, gamesqlc.LockGameMetricCheckpointParams{MetricKey: contract.Key, MetricVersion: contract.Version})
	if err != nil {
		return nil, err
	}
	next, ok := nextDate(checkpoint.SourceStartDate, checkpoint.ProcessedThrough)
	if !ok {
		return nil, nil
	}
	if from != nil && next.Before(utcDate(*from)) {
		return nil, fmt.Errorf("cannot start %s at %s: ordered checkpoint requires preceding day %s", catalogID(contract.Key, contract.Version), from.Format(time.DateOnly), next.Format(time.DateOnly))
	}
	upstream, err := queries.GameMetricUpstreamProcessedThrough(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]DayResult, 0)
	for {
		if through != nil && next.After(utcDate(*through)) {
			break
		}
		if maxDays > 0 && len(result) >= maxDays {
			break
		}
		day := DayResult{MetricKey: contract.Key, MetricVersion: contract.Version, FactDate: next.Format(time.DateOnly)}
		if !upstream.Valid || next.After(utcDate(upstream.Time)) {
			day.Reason = "upstream game.state_facts watermark is not ready"
			result = append(result, day)
			break
		}
		day.PopulationEstimate, err = queries.CountGameMetricPopulation(ctx, pgDate(next))
		if err != nil {
			return nil, err
		}
		day.Reason = "dry-run"
		result = append(result, day)
		next = next.AddDate(0, 0, 1)
	}
	return result, nil
}

func (engine *Engine) nextMetricDate(ctx context.Context, contract Contract) (time.Time, bool, error) {
	tx, err := engine.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return time.Time{}, false, err
	}
	defer tx.Rollback(ctx)
	checkpoint, err := gamesqlc.New(tx).LockGameMetricCheckpoint(ctx, gamesqlc.LockGameMetricCheckpointParams{MetricKey: contract.Key, MetricVersion: contract.Version})
	if err != nil {
		return time.Time{}, false, err
	}
	next, ok := nextDate(checkpoint.SourceStartDate, checkpoint.ProcessedThrough)
	return next, ok, nil
}

func (engine *Engine) resolveMetrics(ctx context.Context, metricKey string, metricVersion int32, activeOnly bool) ([]registeredMetric, error) {
	rows, err := gamesqlc.New(engine.pool).ListGameMetricRegistry(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]registeredMetric, 0)
	for _, row := range rows {
		contract, compiled := contractFor(row.MetricKey, row.MetricVersion)
		if !compiled {
			continue
		}
		if metricKey != "" && row.MetricKey != metricKey {
			continue
		}
		if metricVersion > 0 && row.MetricVersion != metricVersion {
			continue
		}
		if activeOnly && row.Status != "active" {
			continue
		}
		result = append(result, registeredMetric{Contract: contract, Status: row.Status})
	}
	if metricVersion > 0 && metricKey == "" {
		return nil, errors.New("--version requires --metric")
	}
	if metricKey != "" && len(result) == 0 {
		return nil, fmt.Errorf("Game metric %s is not registered%s", catalogID(metricKey, metricVersion), map[bool]string{true: " and active", false: ""}[activeOnly])
	}
	if metricKey != "" && metricVersion == 0 && len(result) > 1 {
		return nil, fmt.Errorf("Game metric %s resolves to multiple versions; specify --version", metricKey)
	}
	return result, nil
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

func metricLagDays(source, processed, upstream pgtype.Date) *int {
	if !upstream.Valid {
		return nil
	}
	next, ok := nextDate(source, processed)
	if !ok {
		return nil
	}
	lag := int(utcDate(upstream.Time).Sub(next).Hours()/24) + 1
	if lag < 0 {
		lag = 0
	}
	return &lag
}

func utcDate(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func pgDate(value time.Time) pgtype.Date { return pgtype.Date{Time: utcDate(value), Valid: true} }

func dateString(value pgtype.Date) *string {
	if !value.Valid {
		return nil
	}
	result := utcDate(value.Time).Format(time.DateOnly)
	return &result
}
