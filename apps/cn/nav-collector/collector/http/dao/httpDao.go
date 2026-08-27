package dao

import (
	"context"
	"strconv"
	"time"

	"github.com/gofurry/gofurry-nav-collector/collector/execution"
	"github.com/gofurry/gofurry-nav-collector/collector/http/models"
	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/retention"
	navsqlc "github.com/gofurry/gofurry-nav-collector/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HTTPDAO struct {
	pool    *pgxpool.Pool
	queries *navsqlc.Queries
}

func New(pool *pgxpool.Pool) *HTTPDAO {
	return &HTTPDAO{pool: pool, queries: navsqlc.New(pool)}
}

func (dao *HTTPDAO) GetList() ([]models.GfnCollectorDomain, common.GFError) {
	rows, err := dao.queries.ListCollectorDomains(context.Background())
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	result := make([]models.GfnCollectorDomain, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		target := row.Name
		if row.Prefix != nil {
			target = *row.Prefix + row.Name
		}
		if !execution.Allows("http", row.SiteID, target) {
			continue
		}
		key := execution.TargetKey(row.SiteID, target)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, models.GfnCollectorDomain{ID: row.ID, SiteID: row.SiteID, Name: row.Name, Proxy: row.Proxy, Prefix: row.Prefix, TLS: row.Tls, Deleted: row.Deleted})
	}
	return result, nil
}

func (dao *HTTPDAO) Add(record *models.GfnCollectorLogHTTP) common.GFError {
	t := time.Time(record.CreateTime)
	err := dao.queries.InsertHTTPLog(context.Background(), navsqlc.InsertHTTPLogParams{
		ID: record.ID, Name: record.Name, Info: []byte(record.Info), Status: record.Status,
		CreateTime: pgtype.Timestamp{Time: t, Valid: !t.IsZero()},
	})
	if err != nil {
		return common.NewDaoError(err.Error())
	}
	return nil
}

func (dao *HTTPDAO) DeleteByNum(count string) (int64, common.GFError) {
	keepCount, err := strconv.Atoi(count)
	if err != nil {
		return 0, common.NewDaoError("count 格式错误: " + err.Error())
	}
	deleted, err := retention.DeleteHTTPByNameLimit(dao.pool, keepCount, retention.DefaultBatchSize, 2*time.Minute, time.Second)
	if err != nil {
		return deleted, common.NewDaoError("HTTP日志分批删除失败: " + err.Error())
	}
	return deleted, nil
}
