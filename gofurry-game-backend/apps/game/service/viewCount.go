package service

import (
	"fmt"
	"time"

	"github.com/GoFurry/gofurry-game-backend/common/log"
	cs "github.com/GoFurry/gofurry-game-backend/common/service"
	"github.com/GoFurry/gofurry-game-backend/common/util"
)

const (
	gameViewCountPrefix = "game:view:count:"
	gameViewDailyPrefix = "game:view:daily:"
	gameViewDailyTTL    = 48 * time.Hour
)

func (s gameService) touchGameViewCount(gameID int64, dbCount int64, clientIP string) int64 {
	countKey := gameViewCountPrefix + util.Int642String(gameID)
	current := dbCount

	if countStr, err := cs.GetString(countKey); err == nil {
		if countStr != "" {
			if parsed, parseErr := util.String2Int64(countStr); parseErr == nil {
				current = parsed
			}
		} else if cs.SetNX(countKey, util.Int642String(current), 0) {
			// current already seeded from database
		} else if latest, latestErr := cs.GetString(countKey); latestErr == nil && latest != "" {
			if parsed, parseErr := util.String2Int64(latest); parseErr == nil {
				current = parsed
			}
		}
	} else {
		log.Warn(fmt.Sprintf("[game-view-count] get redis count failed for game %d: %v", gameID, err))
	}

	if clientIP == "" {
		return current
	}

	dailyKey := fmt.Sprintf("%s%d:%s:%s", gameViewDailyPrefix, gameID, time.Now().Format("2006-01-02"), util.CreateMD5(clientIP))
	if cs.SetNX(dailyKey, "1", gameViewDailyTTL) {
		cs.Incr(countKey)
		if countStr, err := cs.GetString(countKey); err == nil && countStr != "" {
			if parsed, parseErr := util.String2Int64(countStr); parseErr == nil {
				current = parsed
			} else {
				current++
			}
		} else {
			current++
		}
	}

	return current
}
