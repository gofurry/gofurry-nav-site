package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/report"
	cs "github.com/gofurry/gofurry-game-collector/common/service"
	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultRunSummaryCacheTTL = 7 * 24 * time.Hour

// RunRepository writes v2 runner observation records.
type RunRepository struct {
	pool     *pgxpool.Pool
	cacheTTL time.Duration
}

// NewRunRepository creates a repository with an explicit PostgreSQL pool.
func NewRunRepository(pool *pgxpool.Pool) *RunRepository {
	return &RunRepository{pool: pool, cacheTTL: defaultRunSummaryCacheTTL}
}

// SaveRunSummary persists one unified runner summary and refreshes lightweight Redis status keys.
func (r *RunRepository) SaveRunSummary(ctx context.Context, summary report.RunSummary) error {
	if r == nil || r.pool == nil {
		return fmt.Errorf("run repository database is nil")
	}
	if summary.ID == "" {
		return fmt.Errorf("run summary id is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	queries := gamesqlc.New(tx)
	if err := upsertRunSummary(ctx, queries, summary); err != nil {
		return err
	}
	if err := queries.DeleteTaskResultsByRun(ctx, summary.ID); err != nil {
		return err
	}
	for _, result := range summary.Results {
		if err := insertTaskResult(ctx, queries, result); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	r.refreshRunCache(summary)
	return nil
}

func upsertRunSummary(ctx context.Context, queries *gamesqlc.Queries, summary report.RunSummary) error {
	taskSummary, err := marshalJSON(summary.TaskSummaries)
	if err != nil {
		return fmt.Errorf("marshal task summary: %w", err)
	}
	errorKind, errorMessage := errorFields(summary.Error)
	return queries.UpsertCollectRun(ctx, gamesqlc.UpsertCollectRunParams{
		ID:             summary.ID,
		TaskType:       runTaskType(summary),
		Status:         string(summary.Status),
		TotalCount:     int32(summary.TotalCount),
		SuccessCount:   int32(summary.SuccessCount),
		FailedCount:    int32(summary.FailedCount),
		SkippedCount:   int32(summary.SkippedCount),
		PartialCount:   int32(summary.PartialCount),
		TaskSummary:    taskSummary,
		DurationMillis: runDurationMillis(summary),
		ErrorKind:      errorKind,
		ErrorMessage:   errorMessage,
		StartedAt:      timestamptz(summary.StartedAt),
		EndedAt:        timestamptz(summary.EndedAt),
	})
}

func insertTaskResult(ctx context.Context, queries *gamesqlc.Queries, result report.TaskResult) error {
	errorKind, errorMessage := errorFields(result.Error)
	return queries.InsertTaskResult(ctx, gamesqlc.InsertTaskResultParams{
		RunID:              result.RunID,
		TaskType:           string(result.Task),
		Status:             string(result.Status),
		GameID:             result.GameID,
		Appid:              int64(result.AppID),
		UpstreamStatusCode: int32(result.UpstreamStatusCode),
		TrafficBucket:      result.TrafficBucket,
		RetryCount:         int32(result.RetryCount),
		DurationMillis:     result.DurationMillis,
		ErrorKind:          errorKind,
		ErrorMessage:       errorMessage,
		StartedAt:          timestamptz(result.StartedAt),
		EndedAt:            timestamptz(result.EndedAt),
	})
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
