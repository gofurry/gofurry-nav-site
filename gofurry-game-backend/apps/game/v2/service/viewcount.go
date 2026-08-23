package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/log"
	cs "github.com/gofurry/gofurry-game-backend/common/service"
	"github.com/gofurry/gofurry-game-backend/common/util"
	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	gameViewCountPrefix = "game:view:count:"
	gameViewDailyPrefix = "game:view:daily:"
	gameViewDailyTTL    = 48 * time.Hour
)

type GameViewService struct {
	queries *gamesqlc.Queries
}

func NewGameViewService(pool *pgxpool.Pool) *GameViewService {
	if pool == nil {
		return &GameViewService{}
	}
	return &GameViewService{queries: gamesqlc.New(pool)}
}

func (svc *GameViewService) TouchGameViewCount(gameID int64, clientIP string) (int64, common.GFError) {
	if gameID <= 0 {
		return 0, common.NewServiceError("id 不能为空")
	}

	dbCount, err := svc.loadGameViewCountFromDB(gameID)
	if err != nil {
		return 0, err
	}
	if cs.GetRedisService() == nil {
		return dbCount, nil
	}

	current := loadGameCurrentViewCount(gameID, dbCount)
	clientIP = strings.TrimSpace(clientIP)
	if clientIP == "" {
		return current, nil
	}

	countKey := gameViewCountPrefix + util.Int642String(gameID)
	dailyKey := fmt.Sprintf("%s%d:%s:%s", gameViewDailyPrefix, gameID, time.Now().Format("2006-01-02"), util.CreateMD5(clientIP))
	if cs.SetNX(dailyKey, "1", gameViewDailyTTL) {
		cs.Incr(countKey)
		if latest, ok := parseRedisInt64(countKey); ok {
			return latest, nil
		}
		return current + 1, nil
	}

	return current, nil
}

func loadGameCurrentViewCount(gameID int64, fallback int64) int64 {
	if gameID <= 0 || cs.GetRedisService() == nil {
		return fallback
	}
	countKey := gameViewCountPrefix + util.Int642String(gameID)
	return seedGameViewCount(countKey, fallback)
}

func seedGameViewCount(countKey string, fallback int64) int64 {
	if latest, ok := parseRedisInt64(countKey); ok {
		return latest
	}

	if cs.SetNX(countKey, util.Int642String(fallback), 0) {
		return fallback
	}

	if latest, ok := parseRedisInt64(countKey); ok {
		return latest
	}

	return fallback
}

func parseRedisInt64(key string) (int64, bool) {
	value, err := cs.GetString(key)
	if err != nil {
		log.Warn("[game-view-count] read redis failed", "key", key, "error", err)
		return 0, false
	}
	if strings.TrimSpace(value) == "" {
		return 0, false
	}
	parsed, parseErr := util.String2Int64(value)
	if parseErr != nil {
		log.Warn("[game-view-count] parse redis value failed", "key", key, "value", value, "error", parseErr)
		return 0, false
	}
	return parsed, true
}

func (svc *GameViewService) loadGameViewCountFromDB(gameID int64) (int64, common.GFError) {
	if svc == nil || svc.queries == nil {
		return 0, common.NewServiceError("查询游戏浏览量失败")
	}
	viewCount, err := svc.queries.GetGameViewCount(context.Background(), gameID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return 0, common.NewServiceError("game not found")
	case err != nil:
		log.Error("[game-view-count] load database view count failed:", err)
		return 0, common.NewServiceError("查询游戏浏览量失败")
	default:
		return viewCount, nil
	}
}

func (svc *GameViewService) PersistViewCount(gameID int64, viewCount int64) error {
	if svc == nil || svc.queries == nil {
		return errors.New("game view persistence is not initialized")
	}
	_, err := svc.queries.UpdateGameViewCount(context.Background(), gamesqlc.UpdateGameViewCountParams{ID: gameID, ViewCount: viewCount})
	return err
}
