package dao

import (
	"context"

	"github.com/gofurry/gofurry-game-collector/collector/game/models"
	"github.com/gofurry/gofurry-game-collector/common"
	"github.com/gofurry/gofurry-game-collector/common/abstract"
	database "github.com/gofurry/gofurry-game-collector/roof/db"
)

var newsDao = abstract.NewDaoWithDB[models.GfgGameNews](context.Background(), database.Orm.DB().Table(models.TableNameGfgGameNews))

// 获取游戏记录
func GetGameNews(gameID int64, lang string, idx int64) (models.GfgGameNews, common.GFError) {
	var res models.GfgGameNews
	db := newsDao.DB().Table(newsDao.GetTableName()).Where("game_id=? AND lang=? AND index=?", gameID, lang, idx)
	db.Take(&res)
	if err := db.Error; err != nil {
		return res, common.NewDaoError(err.Error())
	}
	return res, nil
}
