package observation

import (
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
