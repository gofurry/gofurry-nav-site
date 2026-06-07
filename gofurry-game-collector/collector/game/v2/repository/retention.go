package repository

import (
	"context"
	"fmt"
	"time"

	database "github.com/gofurry/gofurry-game-collector/roof/db"
	"gorm.io/gorm"
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
	db *gorm.DB
}

// NewRetentionRepository creates a repository backed by the global PostgreSQL handle.
func NewRetentionRepository() *RetentionRepository {
	return NewRetentionRepositoryWithDB(database.Orm.DB())
}

// NewRetentionRepositoryWithDB creates a repository with an explicit PostgreSQL handle.
func NewRetentionRepositoryWithDB(db *gorm.DB) *RetentionRepository {
	return &RetentionRepository{db: db}
}

// Prune deletes records older than the configured retention windows.
func (r *RetentionRepository) Prune(ctx context.Context, cfg RetentionConfig) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("retention repository database is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	cfg = normalizeRetentionConfig(cfg)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := pruneOlderThan(ctx, tx, "gfg_game_v2_collect_task_results", "started_at", cfg.CollectTaskResultsDays); err != nil {
			return err
		}
		if err := pruneOlderThan(ctx, tx, "gfg_game_v2_collect_runs", "started_at", cfg.CollectRunsDays); err != nil {
			return err
		}
		if err := pruneOlderThan(ctx, tx, "gfg_game_v2_player_counts", "collected_at", cfg.PlayerCountsDays); err != nil {
			return err
		}
		return nil
	})
}

func pruneOlderThan(ctx context.Context, tx *gorm.DB, table string, column string, days int) error {
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	return tx.WithContext(ctx).Exec(fmt.Sprintf("DELETE FROM %s WHERE %s < ?", table, column), cutoff).Error
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
