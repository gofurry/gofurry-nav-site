package dao

import (
	"github.com/gofurry/gofurry-nav-backend/apps/system/site/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/abstract"
)

var newSiteDao = new(siteDao)

func init() {
	newSiteDao.Init()
}

type siteDao struct{ abstract.Dao }

func GetSiteDao() *siteDao { return newSiteDao }

func (dao *siteDao) GetChangeLogList() (res []models.ChangeLogVo, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfnLogUpdate)
	db.Select("title, url, create_time, update_time")
	db.Where("deleted IS NOT TRUE")

	db.Order("create_time DESC").Limit(100)

	if dbErr := db.Find(&res).Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}
	return
}
