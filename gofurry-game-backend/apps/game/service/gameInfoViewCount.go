package service

import (
	"github.com/GoFurry/gofurry-game-backend/apps/game/dao"
	"github.com/GoFurry/gofurry-game-backend/apps/game/models"
	"github.com/GoFurry/gofurry-game-backend/common"
	"github.com/GoFurry/gofurry-game-backend/common/util"
)

func (s gameService) GetGameInfoWithViewCount(id string, lang string, clientIP string) (res models.GameBaseInfoVo, err common.GFError) {
	res, err = s.GetGameInfo(id, lang)
	if err != nil {
		return res, err
	}

	intId, parseErr := util.String2Int64(id)
	if parseErr != nil {
		return res, common.NewServiceError("Game ID 杞崲鏈夎")
	}

	viewCount, viewCountErr := dao.GetGameDao().GetGameViewCount(intId)
	if viewCountErr != nil {
		viewCount = 0
	}

	res.ViewCount = s.touchGameViewCount(intId, viewCount, clientIP)
	return res, nil
}
