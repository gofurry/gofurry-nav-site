package observation

import (
	"context"
	"strconv"
	"time"

	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/abstract"
	"github.com/gofurry/gofurry-nav-collector/common/retention"
)

var newObservationDao = new(observationDao)

func init() {
	newObservationDao.Init()
}

type observationDao struct{ abstract.Dao }

func GetObservationDao() *observationDao { return newObservationDao }

func (dao observationDao) AddObservation(record *GfnCollectorObservation) common.GFError {
	db := dao.Gm.Exec(`
INSERT INTO gfn_collector_observation (
	id, site_id, target, protocol, status, observed_at, duration_ms,
	error_code, error_message, payload, schema_version, create_time
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?::jsonb, ?, ?)`,
		record.ID,
		record.SiteID,
		record.Target,
		record.Protocol,
		record.Status,
		record.ObservedAt,
		record.DurationMS,
		record.ErrorCode,
		record.ErrorMessage,
		record.Payload,
		record.SchemaVersion,
		record.CreateTime,
	)
	if err := db.Error; err != nil {
		return common.NewDaoError(err.Error())
	}
	return nil
}

func (dao observationDao) DeleteByProtocolLimit(protocol string, count string) (int64, common.GFError) {
	keepCount, err := strconv.Atoi(count)
	if err != nil {
		return 0, common.NewDaoError("count 格式错误: " + err.Error())
	}

	db := dao.Gm.Table(TableNameGfnCollectorObservation)
	totalDeleted, deleteErr := retention.DeleteObservationByProtocolLimit(db, TableNameGfnCollectorObservation, protocol, keepCount, retention.DefaultBatchSize, 2*time.Minute, time.Second)
	if deleteErr != nil {
		return totalDeleted, common.NewDaoError("v2 observation 分批删除失败: " + deleteErr.Error())
	}

	return totalDeleted, nil
}

type ObservationTrendRow struct {
	Protocol   string
	Status     string
	ObservedAt time.Time
	DurationMS int64
	ErrorCode  *string
	Payload    string
}

func (dao observationDao) ListTrendRows(ctx context.Context, siteID int64, target string, since time.Time, limit int) ([]ObservationTrendRow, common.GFError) {
	if siteID <= 0 || target == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = defaultTrendMaxRows
	}
	rows := []ObservationTrendRow{}
	db := dao.Gm.WithContext(ctx).Raw(`
SELECT protocol, status, observed_at, duration_ms, error_code, payload::text AS payload
FROM gfn_collector_observation
WHERE site_id = ?
  AND target = ?
  AND observed_at >= ?
  AND protocol IN (?, ?, ?)
ORDER BY observed_at DESC, id DESC
LIMIT ?`,
		siteID,
		target,
		since,
		ProtocolPing,
		ProtocolHTTP,
		ProtocolDNS,
		limit,
	).Scan(&rows)
	if db.Error != nil {
		return nil, common.NewDaoError(db.Error.Error())
	}
	return rows, nil
}

func (dao observationDao) ListChangeRows(ctx context.Context, siteID int64, target string, since time.Time, limit int) ([]ObservationTrendRow, common.GFError) {
	if siteID <= 0 || target == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = defaultChangeMaxRows
	}
	rows := []ObservationTrendRow{}
	db := dao.Gm.WithContext(ctx).Raw(`
SELECT protocol, status, observed_at, duration_ms, error_code, payload::text AS payload
FROM gfn_collector_observation
WHERE site_id = ?
  AND target = ?
  AND observed_at >= ?
  AND protocol IN (?, ?, ?, ?)
ORDER BY observed_at DESC, id DESC
LIMIT ?`,
		siteID,
		target,
		since,
		ProtocolHTTP,
		ProtocolDNS,
		ProtocolPortCheck,
		ProtocolRDAP,
		limit,
	).Scan(&rows)
	if db.Error != nil {
		return nil, common.NewDaoError(db.Error.Error())
	}
	return rows, nil
}
