package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/report"
	cs "github.com/gofurry/gofurry-game-collector/common/service"
	database "github.com/gofurry/gofurry-game-collector/roof/db"
	"gorm.io/gorm"
)

const defaultRunSummaryCacheTTL = 7 * 24 * time.Hour

// RunRepository writes v2 runner observation records.
type RunRepository struct {
	db       *gorm.DB
	cacheTTL time.Duration
}

// NewRunRepository creates a repository backed by the global PostgreSQL handle.
func NewRunRepository() *RunRepository {
	return NewRunRepositoryWithDB(database.Orm.DB())
}

// NewRunRepositoryWithDB creates a repository with an explicit PostgreSQL handle.
func NewRunRepositoryWithDB(db *gorm.DB) *RunRepository {
	return &RunRepository{db: db, cacheTTL: defaultRunSummaryCacheTTL}
}

// SaveRunSummary persists one unified runner summary and refreshes lightweight Redis status keys.
func (r *RunRepository) SaveRunSummary(ctx context.Context, summary report.RunSummary) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("run repository database is nil")
	}
	if summary.ID == "" {
		return fmt.Errorf("run summary id is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := upsertRunSummary(ctx, tx, summary); err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Exec("DELETE FROM gfg_game_v2_collect_task_results WHERE run_id = ?", summary.ID).Error; err != nil {
			return err
		}
		for _, result := range summary.Results {
			if err := insertTaskResult(ctx, tx, result); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	r.refreshRunCache(summary)
	return nil
}

func upsertRunSummary(ctx context.Context, tx *gorm.DB, summary report.RunSummary) error {
	taskSummary, err := marshalJSON(summary.TaskSummaries)
	if err != nil {
		return fmt.Errorf("marshal task summary: %w", err)
	}
	errorKind, errorMessage := errorFields(summary.Error)
	return tx.WithContext(ctx).Exec(`
INSERT INTO gfg_game_v2_collect_runs (
    id,
    task_type,
    status,
    total_count,
    success_count,
    failed_count,
    skipped_count,
    partial_count,
    task_summary,
    duration_millis,
    error_kind,
    error_message,
    started_at,
    ended_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?::jsonb, ?, ?, ?, ?, ?
)
ON CONFLICT (id)
DO UPDATE SET
    task_type = EXCLUDED.task_type,
    status = EXCLUDED.status,
    total_count = EXCLUDED.total_count,
    success_count = EXCLUDED.success_count,
    failed_count = EXCLUDED.failed_count,
    skipped_count = EXCLUDED.skipped_count,
    partial_count = EXCLUDED.partial_count,
    task_summary = EXCLUDED.task_summary,
    duration_millis = EXCLUDED.duration_millis,
    error_kind = EXCLUDED.error_kind,
    error_message = EXCLUDED.error_message,
    started_at = EXCLUDED.started_at,
    ended_at = EXCLUDED.ended_at
`,
		summary.ID,
		runTaskType(summary),
		string(summary.Status),
		summary.TotalCount,
		summary.SuccessCount,
		summary.FailedCount,
		summary.SkippedCount,
		summary.PartialCount,
		string(taskSummary),
		runDurationMillis(summary),
		errorKind,
		errorMessage,
		summary.StartedAt,
		nullableTime(summary.EndedAt),
	).Error
}

func insertTaskResult(ctx context.Context, tx *gorm.DB, result report.TaskResult) error {
	errorKind, errorMessage := errorFields(result.Error)
	return tx.WithContext(ctx).Exec(`
INSERT INTO gfg_game_v2_collect_task_results (
    run_id,
    task_type,
    status,
    game_id,
    appid,
    upstream_status_code,
    traffic_bucket,
    retry_count,
    duration_millis,
    error_kind,
    error_message,
    started_at,
    ended_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
`,
		result.RunID,
		string(result.Task),
		string(result.Status),
		result.GameID,
		result.AppID,
		result.UpstreamStatusCode,
		result.TrafficBucket,
		result.RetryCount,
		result.DurationMillis,
		errorKind,
		errorMessage,
		result.StartedAt,
		nullableTime(result.EndedAt),
	).Error
}

func (r *RunRepository) refreshRunCache(summary report.RunSummary) {
	if cs.GetRedisService() == nil {
		return
	}
	payload, err := marshalJSON(cacheableRunSummary(summary))
	if err != nil {
		return
	}
	_ = cs.SetExpire("game:v2:collect:last:all", string(payload), r.cacheTTL)
	for _, task := range summary.TaskSummaries {
		if task.Task == "" {
			continue
		}
		_ = cs.SetExpire(collectLastCacheKey(task.Task), string(payload), r.cacheTTL)
	}
}

func collectLastCacheKey(task domain.TaskType) string {
	return fmt.Sprintf("game:v2:collect:last:%s", task)
}

func cacheableRunSummary(summary report.RunSummary) report.RunSummary {
	summary.Results = nil
	return summary
}

func runTaskType(summary report.RunSummary) string {
	parts := make([]string, 0, len(summary.TaskSummaries))
	for _, task := range summary.TaskSummaries {
		if task.Task != "" {
			parts = append(parts, string(task.Task))
		}
	}
	if len(parts) == 0 {
		return "unknown"
	}
	return strings.Join(parts, ",")
}

func runDurationMillis(summary report.RunSummary) int64 {
	if summary.StartedAt.IsZero() || summary.EndedAt.IsZero() {
		return 0
	}
	return summary.EndedAt.Sub(summary.StartedAt).Milliseconds()
}

func errorFields(errorInfo *report.ErrorInfo) (string, string) {
	if errorInfo == nil {
		return "", ""
	}
	return string(errorInfo.Kind), errorInfo.Message
}
