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
	defaultCollectTaskResultsRetentionDays = 90
)

// RetentionConfig keeps the existing typed configuration shape. Player-count
// and Run pruning are intentionally frozen until the P0.2 retention design.
type RetentionConfig struct {
	PlayerCountsDays       int
	CollectRunsDays        int
	CollectTaskResultsDays int
}

// RetentionRepository prunes only temporary operational task results.
type RetentionRepository struct {
	pool *pgxpool.Pool
}

// NewRetentionRepository creates a repository with an explicit PostgreSQL pool.
func NewRetentionRepository(pool *pgxpool.Pool) *RetentionRepository {
	return &RetentionRepository{pool: pool}
}

// Prune deliberately preserves gfg_game_player_counts and durable Job/Run
// history. Only high-volume operational task results retain the short window.
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
	if _, err := queries.DeleteGameCollectionTaskResultsOlderThan(ctx, timestamptz(retentionCutoff(cfg.CollectTaskResultsDays))); err != nil {
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
	} else if cfg.CollectTaskResultsDays < 30 {
		cfg.CollectTaskResultsDays = 30
	}
	return cfg
}
