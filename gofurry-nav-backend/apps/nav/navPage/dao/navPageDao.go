package dao

import (
	"github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/abstract"
)

var newNavPageDao = new(navPageDao)

func init() {
	newNavPageDao.Init()
}

type navPageDao struct{ abstract.Dao }

func GetNavPageDao() *navPageDao { return newNavPageDao }

func (dao *navPageDao) GetSiteList() (res []models.GfnSite, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfnSite).Where("deleted IS NOT TRUE")
	db.Order("id ASC")
	db.Find(&res)
	if dbErr := db.Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}
	return
}

func (dao *navPageDao) GetGroupList() (res []models.GfnSiteGroup, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfnSiteGroup)
	db.Order("priority ASC")
	db.Find(&res)
	if dbErr := db.Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}
	return
}

func (dao *navPageDao) GetGroupMapList() (res []models.GfnSiteGroupMap, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfnSiteGroupMap)
	db.Find(&res)
	if dbErr := db.Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}
	return
}

// 按 id 序返回指定的金句
func (dao navPageDao) GetSayingByRandom() (*models.GfnSaying, common.GFError) {
	var res models.GfnSaying
	db := dao.Gm.Table("gfn_saying").Order("random()")
	db.Limit(1).First(&res)
	if err := db.Error; err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	return &res, nil
}
