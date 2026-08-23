package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	cs "github.com/gofurry/gofurry-game-collector/common/service"
	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultPlayerCacheTTL = 3 * time.Hour

// PlayerRepository writes v2 player counts into PostgreSQL and Redis.
type PlayerRepository struct {
	queries  *gamesqlc.Queries
	cacheTTL time.Duration
}

// NewPlayerRepository creates a repository with an explicit PostgreSQL pool.
func NewPlayerRepository(pool *pgxpool.Pool) *PlayerRepository {
	var queries *gamesqlc.Queries
	if pool != nil {
		queries = gamesqlc.New(pool)
	}
	return &PlayerRepository{
		queries:  queries,
		cacheTTL: defaultPlayerCacheTTL,
	}
}

// SavePlayerCount inserts one player-count collection result.
func (r *PlayerRepository) SavePlayerCount(ctx context.Context, item domain.PlayerCount) error {
	if r == nil || r.queries == nil {
		return fmt.Errorf("player repository database is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	err := r.queries.InsertPlayerCount(ctx, gamesqlc.InsertPlayerCountParams{
		RunID:              item.RunID,
		GameID:             item.GameID,
		Appid:              int64(item.AppID),
		Count:              item.Count,
		Status:             string(item.Status),
		UpstreamStatusCode: int32(item.UpstreamStatusCode),
		ErrorKind:          item.ErrorKind,
		ErrorMessage:       item.ErrorMessage,
		CollectedAt:        timestamptz(item.CollectedAt),
	})
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
