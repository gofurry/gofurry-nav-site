package dao

import (
	"context"
	"strconv"
	"time"

	"github.com/gofurry/gofurry-nav-collector/collector/ping/models"
	"github.com/gofurry/gofurry-nav-collector/common"
	cm "github.com/gofurry/gofurry-nav-collector/common/models"
	"github.com/gofurry/gofurry-nav-collector/common/retention"
	navsqlc "github.com/gofurry/gofurry-nav-collector/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PingDAO struct {
	pool    *pgxpool.Pool
	queries *navsqlc.Queries
}

func New(pool *pgxpool.Pool) *PingDAO {
	return &PingDAO{pool: pool, queries: navsqlc.New(pool)}
}

func (dao *PingDAO) GetList() ([]models.GfnCollectorDomain, common.GFError) {
	rows, err := dao.queries.ListCollectorDomains(context.Background())
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	result := make([]models.GfnCollectorDomain, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.GfnCollectorDomain{ID: row.ID, SiteID: row.SiteID, Name: row.Name, Proxy: row.Proxy, Prefix: row.Prefix, TLS: row.Tls, Deleted: row.Deleted})
	}
	return result, nil
}

func (dao *PingDAO) Add(record *models.GfnCollectorLogPing) common.GFError {
	err := dao.queries.InsertPingLog(context.Background(), navsqlc.InsertPingLogParams{
		ID: record.ID, Name: record.Name, Delay: record.Delay, Loss: record.Loss,
		Status: record.Status, CreateTime: timestamp(record.CreateTime),
	})
	if err != nil {
		return common.NewDaoError(err.Error())
	}
	return nil
}

func (dao *PingDAO) DeleteByNum(count string) (int64, common.GFError) {
	keepCount, err := strconv.Atoi(count)
	if err != nil {
		return 0, common.NewDaoError("count 格式错误: " + err.Error())
	}
	deleted, err := retention.DeletePingByNameLimit(dao.pool, keepCount, retention.DefaultBatchSize, 2*time.Minute, time.Second)
	if err != nil {
		return deleted, common.NewDaoError("Ping日志分批删除失败: " + err.Error())
	}
	return deleted, nil
}

func timestamp(value cm.LocalTime) pgtype.Timestamp {
	t := time.Time(value)
	return pgtype.Timestamp{Time: t, Valid: !t.IsZero()}
}
