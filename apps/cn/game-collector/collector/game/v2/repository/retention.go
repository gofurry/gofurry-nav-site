package repository

import (
	"context"
	"fmt"
	"time"

	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultPlayerCountsRetentionDays       = 90
	defaultCollectRunsRetentionDays        = 90
	defaultCollectTaskResultsRetentionDays = 7
)

// RetentionConfig controls cleanup for append-only v2 observation/history tables.
type RetentionConfig struct {
	PlayerCountsDays       int
	CollectRunsDays        int
	CollectTaskResultsDays int
}

// RetentionRepository prunes v2 append-only tables.
type RetentionRepository struct {
	pool *pgxpool.Pool
}

// NewRetentionRepository creates a repository with an explicit PostgreSQL pool.
func NewRetentionRepository(pool *pgxpool.Pool) *RetentionRepository {
	return &RetentionRepository{pool: pool}
}

// Prune deletes records older than the configured retention windows.
func (r *RetentionRepository) Prune(ctx context.Context, cfg RetentionConfig) error {
	if r == nil || r.pool == nil {
		return fmt.Errorf("retention repository database is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	cfg = normalizeRetentionConfig(cfg)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	queries := gamesqlc.New(tx)
	if _, err := queries.DeleteTaskResultsOlderThan(ctx, timestamptz(retentionCutoff(cfg.CollectTaskResultsDays))); err != nil {
		return err
	}
	if _, err := queries.DeleteCollectRunsOlderThan(ctx, timestamptz(retentionCutoff(cfg.CollectRunsDays))); err != nil {
		return err
	}
	if _, err := queries.DeletePlayerCountsOlderThan(ctx, timestamptz(retentionCutoff(cfg.PlayerCountsDays))); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func retentionCutoff(days int) time.Time {
	return time.Now().Add(-time.Duration(days) * 24 * time.Hour)
}

func normalizeRetentionConfig(cfg RetentionConfig) RetentionConfig {
	if cfg.PlayerCountsDays <= 0 {
		cfg.PlayerCountsDays = defaultPlayerCountsRetentionDays
	}
	if cfg.CollectRunsDays <= 0 {
		cfg.CollectRunsDays = defaultCollectRunsRetentionDays
	}
	if cfg.CollectTaskResultsDays <= 0 {
		cfg.CollectTaskResultsDays = defaultCollectTaskResultsRetentionDays
	}
	return cfg
}
