package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	cs "github.com/gofurry/gofurry-game-collector/common/service"
	database "github.com/gofurry/gofurry-game-collector/roof/db"
	"gorm.io/gorm"
)

const defaultPlayerCacheTTL = 3 * time.Hour

// PlayerRepository writes v2 player counts into PostgreSQL and Redis.
type PlayerRepository struct {
	db       *gorm.DB
	cacheTTL time.Duration
}

// NewPlayerRepository creates a repository backed by the global PostgreSQL handle.
func NewPlayerRepository() *PlayerRepository {
	return NewPlayerRepositoryWithDB(database.Orm.DB())
}

// NewPlayerRepositoryWithDB creates a repository with an explicit PostgreSQL handle.
func NewPlayerRepositoryWithDB(db *gorm.DB) *PlayerRepository {
	return &PlayerRepository{
		db:       db,
		cacheTTL: defaultPlayerCacheTTL,
	}
}

// SavePlayerCount inserts one player-count collection result.
func (r *PlayerRepository) SavePlayerCount(ctx context.Context, item domain.PlayerCount) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("player repository database is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	err := r.db.WithContext(ctx).Exec(`
INSERT INTO gfg_game_v2_player_counts (
    run_id,
    game_id,
    appid,
    count,
    status,
    upstream_status_code,
    error_kind,
    error_message,
    collected_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
)
`,
		item.RunID,
		item.GameID,
		item.AppID,
		item.Count,
		string(item.Status),
		item.UpstreamStatusCode,
		item.ErrorKind,
		item.ErrorMessage,
		item.CollectedAt,
	).Error
	if err != nil {
		return fmt.Errorf("insert v2 player count appid=%d status=%s: %w", item.AppID, item.Status, err)
	}

	if item.Status == domain.StatusSuccess {
		r.refreshCache(item)
	}
	return nil
}

func (r *PlayerRepository) refreshCache(item domain.PlayerCount) {
	if cs.GetRedisService() == nil {
		return
	}
	payload, err := json.Marshal(item)
	if err != nil {
		return
	}
	_ = cs.SetExpire(playerCacheKey(item.GameID), string(payload), r.cacheTTL)
}

func playerCacheKey(gameID int64) string {
	return fmt.Sprintf("game:v2:players:%d:current", gameID)
}
