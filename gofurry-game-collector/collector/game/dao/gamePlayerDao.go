package dao

import (
	"context"

	"github.com/gofurry/gofurry-game-collector/collector/game/models"
	"github.com/gofurry/gofurry-game-collector/common"
	"github.com/gofurry/gofurry-game-collector/common/abstract"
	database "github.com/gofurry/gofurry-game-collector/roof/db"
)

var playerDao = abstract.NewDaoWithDB[models.GfgGamePlayerCount](context.Background(), database.Orm.DB().Table(models.TableNameGfgGamePlayerCount))

// 获取在线人数记录数量
func GetPlayerCountByID(id int64) (cnt int64, gfError common.GFError) {
	db := playerDao.DB().Table(playerDao.GetTableName()).Where("game_id=?", id).Count(&cnt)
	if err := db.Error; err != nil {
		return 0, common.NewDaoError(err.Error())
	}
	return
}

// 获取最后一条在线人数记录的 ID
func GetLastRecordByID(id int64, skipCount int) (recordId []int64, gfError common.GFError) {
	db := playerDao.DB().Table(playerDao.GetTableName()).Select("id").Where("game_id=?", id)
	db = db.Order("create_time ASC")
	db = db.Offset(skipCount).Find(&recordId)
	if err := db.Error; err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	return recordId, nil
}
