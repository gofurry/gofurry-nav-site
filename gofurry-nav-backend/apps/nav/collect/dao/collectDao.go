package dao

import (
	"strings"
	"sync"

	collectmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/collect/models"
	readmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/readmodel/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/abstract"
)

var (
	newCollectDao = new(collectDao)
	collectDaoMu  sync.Mutex
)

type collectDao struct{ abstract.Dao }

func GetCollectDao() *collectDao {
	collectDaoMu.Lock()
	defer collectDaoMu.Unlock()
	if newCollectDao.Gm == nil {
		newCollectDao.Init()
	}
	return newCollectDao
}

func (dao collectDao) ListObservationSummary() ([]collectmodels.ObservationStatusSummary, common.GFError) {
	rows := []collectmodels.ObservationStatusSummary{}
	db := dao.Gm.Raw(`
SELECT protocol, status, COUNT(*) AS count
FROM gfn_collector_observation
WHERE observed_at >= NOW() - INTERVAL '7 days'
GROUP BY protocol, status
ORDER BY protocol ASC, status ASC
`).Scan(&rows)
	if db.Error != nil {
		return nil, common.NewDaoError(db.Error.Error())
	}
	return rows, nil
}

func (dao collectDao) ListObservations(query collectmodels.ObservationQuery) ([]collectmodels.ObservationItem, common.GFError) {
	limit := normalizeLimit(query.Limit, 50, 200)
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}
	rows := []collectmodels.ObservationItem{}
	db := dao.Gm.Table(readmodels.TableNameGfnCollectorObservation).
		Select(`
id, site_id, target, protocol, status, observed_at, duration_ms, error_code, error_message,
COALESCE(payload->>'collector_id', '') AS collector_id,
COALESCE(payload->>'job_id', '') AS job_id
`)
	if query.SiteID > 0 {
		db = db.Where("site_id = ?", query.SiteID)
	}
	if strings.TrimSpace(query.Target) != "" {
		db = db.Where("target = ?", strings.TrimSpace(query.Target))
	}
	if strings.TrimSpace(query.Protocol) != "" {
		db = db.Where("protocol = ?", strings.TrimSpace(query.Protocol))
	}
	if strings.TrimSpace(query.Status) != "" {
		db = db.Where("status = ?", strings.TrimSpace(query.Status))
	}
	if err := db.Order("observed_at DESC, id DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	return rows, nil
}

func normalizeLimit(value int, fallback int, max int) int {
	if value <= 0 {
		return fallback
	}
	if value > max {
		return max
	}
	return value
}
