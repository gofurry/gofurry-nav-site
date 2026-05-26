package dao

import (
	"strings"
	"sync"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/readmodel/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/abstract"
)

var (
	newObservationDao = new(observationDao)
	observationDaoMu  sync.Mutex
)

type observationDao struct{ abstract.Dao }

func GetObservationDao() *observationDao {
	observationDaoMu.Lock()
	defer observationDaoMu.Unlock()
	if newObservationDao.Gm == nil {
		newObservationDao.Init()
	}
	return newObservationDao
}

func (dao observationDao) ListObservations(siteID int64, target string, protocol string, limit int) ([]models.GfnCollectorObservation, common.GFError) {
	target = strings.TrimSpace(target)
	protocol = strings.TrimSpace(protocol)
	limit = models.NormalizeObservationLimit(limit)
	records := []models.GfnCollectorObservation{}

	db := dao.Gm.Table(models.TableNameGfnCollectorObservation).
		Where("site_id = ?", siteID).
		Where("target = ?", target).
		Where("protocol = ?", protocol).
		Order("observed_at DESC, id DESC").
		Limit(limit).
		Find(&records)
	if db.Error != nil {
		return nil, common.NewDaoError(db.Error.Error())
	}
	return records, nil
}
