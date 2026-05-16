package dao

import (
	"errors"

	gm "github.com/gofurry/gofurry-game-backend/apps/game/models"
	"github.com/gofurry/gofurry-game-backend/apps/recommend/models"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/abstract"
	"gorm.io/gorm"
)

var newRecommendDao = new(recommendDao)

func init() {
	newRecommendDao.Init()
}

type recommendDao struct{ abstract.Dao }

func GetRecommendDao() *recommendDao { return newRecommendDao }

func (dao recommendDao) GetTagMappingList() (res []models.GfgTagMap, gfError common.GFError) {
	db := dao.Gm.Table(models.TableNameGfgTagMap).Find(&res)
	if err := db.Error; err != nil {
		return res, common.NewDaoError(err.Error())
	}
	return res, nil
}

func (dao recommendDao) GetTagList() (res []models.GfgTag, gfError common.GFError) {
	db := dao.Gm.Table(models.TableNameGfgTag).Find(&res)
	if err := db.Error; err != nil {
		return res, common.NewDaoError(err.Error())
	}
	return res, nil
}

func (dao recommendDao) GetRecommend(gameIDs []int64, lang string) (res []models.GameTemp, gfError common.GFError) {
	db := dao.Gm.Table(gm.TableNameGfgGame).Where("id IN ?", gameIDs)

	if err := db.Find(&res).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return res, nil // 未找到对应游戏，返回空
		}
		return res, common.NewDaoError("query game info failed: " + err.Error())
	}

	return res, nil
}
