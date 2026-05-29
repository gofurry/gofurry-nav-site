package dao

import (
	"errors"
	"sync"

	detailmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/detail/models"
	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/abstract"
	"gorm.io/gorm"
)

var (
	newDetailDao = new(detailDao)
	detailDaoMu  sync.Mutex
)

type detailDao struct{ abstract.Dao }

func GetDetailDao() *detailDao {
	detailDaoMu.Lock()
	defer detailDaoMu.Unlock()
	if newDetailDao.Gm == nil {
		newDetailDao.Init()
	}
	return newDetailDao
}

func (dao detailDao) GetSiteByID(siteID int64) (navmodels.GfnSite, common.GFError) {
	record := navmodels.GfnSite{}
	db := dao.Gm.Table(navmodels.TableNameGfnSite).
		Where("id = ?", siteID).
		Where("deleted IS NOT TRUE").
		Take(&record)
	if db.Error != nil {
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			return record, common.NewDaoError("404")
		}
		return record, common.NewDaoError(db.Error.Error())
	}
	return record, nil
}

func (dao detailDao) ListCollectorDomains(siteID int64) ([]detailmodels.CollectorDomain, common.GFError) {
	records := []detailmodels.CollectorDomain{}
	db := dao.Gm.Table(detailmodels.TableNameGfnCollectorDomain).
		Where("site_id = ?", siteID).
		Where("deleted IS NOT TRUE").
		Order("id ASC").
		Find(&records)
	if db.Error != nil {
		return nil, common.NewDaoError(db.Error.Error())
	}
	return records, nil
}
