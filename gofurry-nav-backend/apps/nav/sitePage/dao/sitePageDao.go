package dao

import (
	"sync"

	siteModel "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/sitePage/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/abstract"
)

var newSitePageDao = new(sitePageDao)
var sitePageDaoMu sync.Mutex

type sitePageDao struct{ abstract.Dao }

func GetSitePageDao() *sitePageDao {
	sitePageDaoMu.Lock()
	defer sitePageDaoMu.Unlock()
	if newSitePageDao.Gm == nil {
		newSitePageDao.Init()
	}
	return newSitePageDao
}

func (dao sitePageDao) GetSiteById(id int64) (record siteModel.GfnSite, err common.GFError) {
	db := dao.Gm.Table(siteModel.TableNameGfnSite)
	db.Where("id = ?", id).Take(&record)
	if dbErr := db.Error; dbErr != nil {
		return record, common.NewDaoError(dbErr.Error())
	}
	return
}

func (dao sitePageDao) GetDelayList(domain string) (record []models.GfnCollectorLogPing, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfnCollectorLogPing)
	db.Where("name = ?", domain)
	db.Order("create_time DESC").Limit(100).Find(&record)
	if dbErr := db.Error; dbErr != nil {
		return record, common.NewDaoError(dbErr.Error())
	}
	return
}
