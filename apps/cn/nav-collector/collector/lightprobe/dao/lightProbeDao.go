package dao

import (
	"context"

	"github.com/gofurry/gofurry-nav-collector/collector/execution"
	"github.com/gofurry/gofurry-nav-collector/collector/lightprobe/models"
	"github.com/gofurry/gofurry-nav-collector/common"
	navsqlc "github.com/gofurry/gofurry-nav-collector/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LightProbeDAO struct {
	queries *navsqlc.Queries
}

func New(pool *pgxpool.Pool) *LightProbeDAO {
	return &LightProbeDAO{queries: navsqlc.New(pool)}
}

func (dao *LightProbeDAO) GetList(protocol string) ([]models.GfnCollectorDomain, common.GFError) {
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
		if !execution.Allows(protocol, row.SiteID, target) {
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
