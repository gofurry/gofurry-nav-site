package changes

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

type Options struct{ ReconcileInterval time.Duration }
type Engine struct {
	pool    *pgxpool.Pool
	options Options
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}
type Status struct {
	DetectorKey              string  `json:"detector_key"`
	DetectorVersion          int32   `json:"detector_version"`
	Status                   string  `json:"status"`
	WatermarkPolicy          string  `json:"watermark_policy"`
	SourceStartDate          *string `json:"source_start_date"`
	ProcessedThrough         *string `json:"processed_through"`
	UpstreamProcessedThrough *string `json:"upstream_processed_through"`
	LatestAvailableDay       *string `json:"latest_available_day"`
	LagDays                  *int    `json:"lag_days"`
}
type DayResult struct {
	DetectorKey     string `json:"detector_key"`
	DetectorVersion int32  `json:"detector_version"`
	ProjectionDate  string `json:"projection_date,omitempty"`
	Processed       bool   `json:"processed"`
	EventCount      int64  `json:"event_count,omitempty"`
	Reason          string `json:"reason,omitempty"`
}
type BackfillOptions struct {
	Detector string
	Version  int32
	From     *time.Time
	Through  *time.Time
	MaxDays  int
	DryRun   bool
}
type Summary struct {
	DryRun    bool        `json:"dry_run"`
	Processed int         `json:"processed_days"`
	Days      []DayResult `json:"days"`
}
type registeredDetector struct {
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
		log.Error("startup Nav change reconcile failed without stopping Collector: " + err.Error())
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
					log.Error("Nav change reconcile failed without stopping Collector: " + err.Error())
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
func (engine *Engine) Reconcile(ctx context.Context) error {
	detectors, err := engine.resolve(ctx, "", 0, true)
	if err != nil {
		return err
	}
	return reconcileDetectors(detectors, func(detector Contract) (DayResult, error) {
		return engine.RunNext(ctx, detector.Key, detector.Version, false)
	})
}

func reconcileDetectors(detectors []registeredDetector, run func(Contract) (DayResult, error)) error {
	var resultErr error
	for _, detector := range detectors {
		result, runErr := run(detector.Contract)
		if runErr != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("%s: %w", catalogID(detector.Contract.Key, detector.Contract.Version), runErr))
			continue
		}
		if result.Processed {
			log.Info("Nav changes finalized, detector=", result.DetectorKey, " version=", result.DetectorVersion, " projection_date=", result.ProjectionDate, " events=", result.EventCount)
		}
	}
	return resultErr
}
func (engine *Engine) Status(ctx context.Context) ([]Status, error) {
	if err := engine.ValidateCatalog(ctx); err != nil {
		return nil, err
	}
	rows, err := navsqlc.New(engine.pool).ListNavChangeCheckpoints(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Status, 0, len(rows))
	for _, row := range rows {
		result = append(result, Status{DetectorKey: row.DetectorKey, DetectorVersion: row.DetectorVersion, Status: row.Status, WatermarkPolicy: row.WatermarkPolicy, SourceStartDate: dateString(row.SourceStartDate), ProcessedThrough: dateString(row.ProcessedThrough), UpstreamProcessedThrough: dateString(row.UpstreamProcessedThrough), LatestAvailableDay: dateString(row.UpstreamProcessedThrough), LagDays: lagDays(row.SourceStartDate, row.ProcessedThrough, row.UpstreamProcessedThrough)})
	}
	return result, nil
}
func (engine *Engine) RunNext(ctx context.Context, key string, version int32, dryRun bool) (DayResult, error) {
	if _, ok := contractFor(key, version); !ok {
		return DayResult{}, fmt.Errorf("Nav change detector is not compiled for %s", catalogID(key, version))
	}
	tx, err := engine.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return DayResult{}, err
	}
	defer tx.Rollback(ctx)
	queries := navsqlc.New(tx)
	cp, err := queries.LockNavChangeCheckpoint(ctx, navsqlc.LockNavChangeCheckpointParams{DetectorKey: key, DetectorVersion: version})
	if err != nil {
		return DayResult{}, err
	}
	if cp.Status != "active" {
		return DayResult{DetectorKey: key, DetectorVersion: version, Reason: "detector version is not active"}, tx.Commit(ctx)
	}
	day, ok := nextDate(cp.SourceStartDate, cp.ProcessedThrough)
	if !ok {
		return DayResult{DetectorKey: key, DetectorVersion: version, Reason: "source_start_date is not available"}, tx.Commit(ctx)
	}
	result := DayResult{DetectorKey: key, DetectorVersion: version, ProjectionDate: day.Format(time.DateOnly)}
	upstream, err := queries.NavChangeUpstreamProcessedThrough(ctx, navsqlc.NavChangeUpstreamProcessedThroughParams{DetectorKey: key, DetectorVersion: version})
	if err != nil {
		return DayResult{}, err
	}
	if !upstream.Valid || day.After(utcDate(upstream.Time)) {
		result.Reason = "upstream watermark is not ready"
		return result, tx.Commit(ctx)
	}
	if dryRun {
		result.Reason = "dry-run"
		return result, tx.Commit(ctx)
	}
	result.EventCount, err = queries.ProjectNavChangeDay(ctx, navsqlc.ProjectNavChangeDayParams{DetectorKey: key, DetectorVersion: version, ProjectionDate: pgDate(day)})
	if err != nil {
		return DayResult{}, err
	}
	rows, err := queries.AdvanceNavChangeCheckpoint(ctx, navsqlc.AdvanceNavChangeCheckpointParams{ProcessedThrough: pgDate(day), DetectorKey: key, DetectorVersion: version})
	if err != nil {
		return DayResult{}, err
	}
	if rows != 1 {
		return DayResult{}, fmt.Errorf("advance Nav change checkpoint %s on %s: updated %d rows", catalogID(key, version), result.ProjectionDate, rows)
	}
	if err := tx.Commit(ctx); err != nil {
		return DayResult{}, err
	}
	result.Processed = true
	return result, nil
}
func (engine *Engine) Backfill(ctx context.Context, options BackfillOptions) (Summary, error) {
	if err := engine.ValidateCatalog(ctx); err != nil {
		return Summary{}, err
	}
	if options.MaxDays < 0 {
		return Summary{}, errors.New("--max-days must not be negative")
	}
	if options.From != nil && options.Through != nil && options.Through.Before(*options.From) {
		return Summary{}, errors.New("--through must not be before --from")
	}
	detectors, err := engine.resolve(ctx, options.Detector, options.Version, true)
	if err != nil {
		return Summary{}, err
	}
	summary := Summary{DryRun: options.DryRun, Days: make([]DayResult, 0)}
	remaining := options.MaxDays
	for _, detector := range detectors {
		if options.DryRun {
			planned, planErr := engine.plan(ctx, detector.Contract, options.From, options.Through, remaining)
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
			if options.MaxDays > 0 && remaining == 0 {
				return summary, nil
			}
			day, ok, nextErr := engine.nextDetectorDate(ctx, detector.Contract)
			if nextErr != nil {
				return summary, nextErr
			}
			if !ok {
				break
			}
			if options.From != nil && day.Before(utcDate(*options.From)) {
				return summary, fmt.Errorf("cannot start %s at %s: ordered checkpoint requires preceding day %s", catalogID(detector.Contract.Key, detector.Contract.Version), options.From.Format(time.DateOnly), day.Format(time.DateOnly))
			}
			if options.Through != nil && day.After(utcDate(*options.Through)) {
				break
			}
			result, runErr := engine.RunNext(ctx, detector.Contract.Key, detector.Contract.Version, options.DryRun)
			if runErr != nil {
				return summary, runErr
			}
			summary.Days = append(summary.Days, result)
			if options.DryRun || !result.Processed {
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

func (engine *Engine) plan(ctx context.Context, contract Contract, from, through *time.Time, maxDays int) ([]DayResult, error) {
	tx, err := engine.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	queries := navsqlc.New(tx)
	cp, err := queries.LockNavChangeCheckpoint(ctx, navsqlc.LockNavChangeCheckpointParams{DetectorKey: contract.Key, DetectorVersion: contract.Version})
	if err != nil {
		return nil, err
	}
	day, ok := nextDate(cp.SourceStartDate, cp.ProcessedThrough)
	if !ok {
		return nil, nil
	}
	if from != nil && day.Before(utcDate(*from)) {
		return nil, fmt.Errorf("cannot start %s at %s: ordered checkpoint requires preceding day %s", catalogID(contract.Key, contract.Version), from.Format(time.DateOnly), day.Format(time.DateOnly))
	}
	upstream, err := queries.NavChangeUpstreamProcessedThrough(ctx, navsqlc.NavChangeUpstreamProcessedThroughParams{DetectorKey: contract.Key, DetectorVersion: contract.Version})
	if err != nil {
		return nil, err
	}
	result := make([]DayResult, 0)
	for {
		if through != nil && day.After(utcDate(*through)) {
			break
		}
		if maxDays > 0 && len(result) >= maxDays {
			break
		}
		item := DayResult{DetectorKey: contract.Key, DetectorVersion: contract.Version, ProjectionDate: day.Format(time.DateOnly)}
		if !upstream.Valid || day.After(utcDate(upstream.Time)) {
			item.Reason = "upstream watermark is not ready"
			result = append(result, item)
			break
		}
		item.Reason = "dry-run"
		result = append(result, item)
		day = day.AddDate(0, 0, 1)
	}
	return result, nil
}
func (engine *Engine) Rebuild(ctx context.Context, key string, version int32, start time.Time, through *time.Time, maxDays int, dryRun bool) (Summary, error) {
	if err := engine.ValidateCatalog(ctx); err != nil {
		return Summary{}, err
	}
	if key == "" || version <= 0 {
		return Summary{}, errors.New("--detector and --version are required")
	}
	if maxDays < 0 {
		return Summary{}, errors.New("--max-days must not be negative")
	}
	if _, err := engine.resolve(ctx, key, version, false); err != nil {
		return Summary{}, err
	}
	tx, err := engine.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Summary{}, err
	}
	defer tx.Rollback(ctx)
	queries := navsqlc.New(tx)
	cp, err := queries.LockNavChangeCheckpoint(ctx, navsqlc.LockNavChangeCheckpointParams{DetectorKey: key, DetectorVersion: version})
	if err != nil {
		return Summary{}, err
	}
	end, err := validateRebuildBounds(cp.SourceStartDate, cp.ProcessedThrough, start, through, maxDays)
	if err != nil {
		return Summary{}, err
	}
	summary := Summary{DryRun: dryRun, Days: make([]DayResult, 0)}
	for day := utcDate(start); !day.After(end); day = day.AddDate(0, 0, 1) {
		result := DayResult{DetectorKey: key, DetectorVersion: version, ProjectionDate: day.Format(time.DateOnly)}
		upstream, upstreamErr := queries.NavChangeUpstreamProcessedThrough(ctx, navsqlc.NavChangeUpstreamProcessedThroughParams{DetectorKey: key, DetectorVersion: version})
		if upstreamErr != nil {
			return Summary{}, upstreamErr
		}
		if !upstream.Valid || day.After(utcDate(upstream.Time)) {
			return Summary{}, fmt.Errorf("cannot rebuild %s: upstream watermark is not ready", result.ProjectionDate)
		}
		if dryRun {
			result.Reason = "dry-run rebuild"
		} else {
			result.EventCount, err = queries.ProjectNavChangeDay(ctx, navsqlc.ProjectNavChangeDayParams{DetectorKey: key, DetectorVersion: version, ProjectionDate: pgDate(day)})
			if err != nil {
				return Summary{}, err
			}
			result.Processed = true
			summary.Processed++
		}
		summary.Days = append(summary.Days, result)
	}
	if err := tx.Commit(ctx); err != nil {
		return Summary{}, err
	}
	return summary, nil
}
func (engine *Engine) nextDetectorDate(ctx context.Context, contract Contract) (time.Time, bool, error) {
	tx, err := engine.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return time.Time{}, false, err
	}
	defer tx.Rollback(ctx)
	cp, err := navsqlc.New(tx).LockNavChangeCheckpoint(ctx, navsqlc.LockNavChangeCheckpointParams{DetectorKey: contract.Key, DetectorVersion: contract.Version})
	if err != nil {
		return time.Time{}, false, err
	}
	day, ok := nextDate(cp.SourceStartDate, cp.ProcessedThrough)
	return day, ok, nil
}
func (engine *Engine) resolve(ctx context.Context, key string, version int32, activeOnly bool) ([]registeredDetector, error) {
	if version > 0 && key == "" {
		return nil, errors.New("--version requires --detector")
	}
	rows, err := navsqlc.New(engine.pool).ListNavChangeRegistry(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]registeredDetector, 0)
	for _, row := range rows {
		contract, ok := contractFor(row.DetectorKey, row.DetectorVersion)
		if !ok {
			continue
		}
		if (key != "" && key != row.DetectorKey) || (version > 0 && version != row.DetectorVersion) || (activeOnly && row.Status != "active") {
			continue
		}
		result = append(result, registeredDetector{Contract: contract, Status: row.Status})
	}
	if key != "" && len(result) == 0 {
		return nil, fmt.Errorf("Nav change detector %s is not registered%s", catalogID(key, version), map[bool]string{true: " and active", false: ""}[activeOnly])
	}
	if key != "" && version == 0 && len(result) > 1 {
		return nil, fmt.Errorf("Nav change detector %s resolves to multiple versions; specify --version", key)
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

func validateRebuildBounds(source, processed pgtype.Date, start time.Time, through *time.Time, maxDays int) (time.Time, error) {
	if !source.Valid || start.Before(utcDate(source.Time)) {
		return time.Time{}, errors.New("--from is before source_start_date")
	}
	if !processed.Valid {
		return time.Time{}, errors.New("detector has no processed_through to rebuild")
	}
	end := utcDate(processed.Time)
	if through != nil && !utcDate(*through).Equal(end) {
		return time.Time{}, fmt.Errorf("--through must equal processed_through %s", end.Format(time.DateOnly))
	}
	if maxDays > 0 && int(end.Sub(utcDate(start)).Hours()/24)+1 > maxDays {
		return time.Time{}, errors.New("--max-days cannot truncate required forward-propagation rebuild")
	}
	return end, nil
}
func lagDays(source, processed, upstream pgtype.Date) *int {
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
