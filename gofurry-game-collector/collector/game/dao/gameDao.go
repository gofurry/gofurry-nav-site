package dao

import (
	"context"

	"github.com/gofurry/gofurry-game-collector/collector/game/models"
	"github.com/gofurry/gofurry-game-collector/common"
	"github.com/gofurry/gofurry-game-collector/common/abstract"
	database "github.com/gofurry/gofurry-game-collector/roof/db"
)

var gameDao = abstract.NewDaoWithDB[models.GfgGame](context.Background(), database.Orm.DB().Table(models.TableNameGfgGame))
var gameRecordDao = abstract.NewDaoWithDB[models.GfgGameRecord](context.Background(), database.Orm.DB().Table(models.TableNameGfgGameRecord))

// 获取游戏列表
func GetGameList() ([]models.GameID, common.GFError) {
	var res []models.GameID
	db := gameDao.DB().Table(gameDao.GetTableName()).Select("id, appid")
	db.Find(&res)
	if err := db.Error; err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	return res, nil
}

// 获取游戏记录
func GetGameRecordByGameIDAndLang(gameID int64, lang string) (models.GfgGameRecord, common.GFError) {
	var res models.GfgGameRecord
	db := gameRecordDao.DB().Table(gameRecordDao.GetTableName()).Where("game_id=? AND lang=?", gameID, lang)
	db.Take(&res)
	if err := db.Error; err != nil {
		return res, common.NewDaoError(err.Error())
	}
	return res, nil
}
