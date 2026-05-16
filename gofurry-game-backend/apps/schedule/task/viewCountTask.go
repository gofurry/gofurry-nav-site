package task

import (
	"strings"

	"github.com/gofurry/gofurry-game-backend/apps/game/dao"
	gameModels "github.com/gofurry/gofurry-game-backend/apps/game/models"
	"github.com/gofurry/gofurry-game-backend/common/log"
	cs "github.com/gofurry/gofurry-game-backend/common/service"
	"github.com/gofurry/gofurry-game-backend/common/util"
)

const gameViewCountPrefix = "game:view:count:"

func UpdateGameViewCountCache() {
	keys, err := cs.FindByPrefix(gameViewCountPrefix)
	if err != nil {
		log.Error("[UpdateGameViewCountCache] find redis keys err:", err)
		return
	}

	for _, key := range keys {
		idStr := strings.TrimPrefix(key, gameViewCountPrefix)
		gameID, parseErr := util.String2Int64(idStr)
		if parseErr != nil {
			continue
		}

		countStr, getErr := cs.GetString(key)
		if getErr != nil || countStr == "" {
			continue
		}

		viewCount, parseCountErr := util.String2Int64(countStr)
		if parseCountErr != nil {
			continue
		}

		if dbErr := dao.GetGameDao().Gm.Table(gameModels.TableNameGfgGame).Where("id = ?", gameID).Update("view_count", viewCount).Error; dbErr != nil {
			log.Error("[UpdateGameViewCountCache] update game view count err:", dbErr)
		}
	}
}
