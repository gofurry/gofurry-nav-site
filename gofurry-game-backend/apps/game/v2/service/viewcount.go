package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofurry/gofurry-game-backend/apps/game/dao"
	gameModels "github.com/gofurry/gofurry-game-backend/apps/game/models"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/log"
	cs "github.com/gofurry/gofurry-game-backend/common/service"
	"github.com/gofurry/gofurry-game-backend/common/util"
	"gorm.io/gorm"
)

const (
	gameViewCountPrefix = "game:view:count:"
	gameViewDailyPrefix = "game:view:daily:"
	gameViewDailyTTL    = 48 * time.Hour
)

type GameViewService struct{}

var gameViewSvc = &GameViewService{}

func GetGameViewService() *GameViewService {
	return gameViewSvc
}

func (svc *GameViewService) TouchGameViewCount(gameID int64, clientIP string) (int64, common.GFError) {
	if gameID <= 0 {
		return 0, common.NewServiceError("id 不能为空")
	}

	dbCount, err := loadGameViewCountFromDB(gameID)
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

func loadGameViewCountFromDB(gameID int64) (int64, common.GFError) {
	var row struct {
		ViewCount int64 `gorm:"column:view_count"`
	}

	err := dao.GetGameDao().Gm.Table(gameModels.TableNameGfgGame).
		Select("view_count").
		Where("id = ?", gameID).
		Take(&row).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return 0, common.NewServiceError("game not found")
	case err != nil:
		log.Error("[game-view-count] load database view count failed:", err)
		return 0, common.NewServiceError("查询游戏浏览量失败")
	default:
		return row.ViewCount, nil
	}
}
