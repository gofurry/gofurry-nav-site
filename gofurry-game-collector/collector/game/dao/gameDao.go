package dao

import (
	"context"

	"github.com/gofurry/gofurry-game-collector/collector/game/models"
	"github.com/gofurry/gofurry-game-collector/common"
	"github.com/gofurry/gofurry-game-collector/common/abstract"
	database "github.com/gofurry/gofurry-game-collector/roof/db"
)

var gameDao = abstract.NewDaoWithDB[models.GfgGame](context.Background(), database.Orm.DB().Table(models.TableNameGfgGame))

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
