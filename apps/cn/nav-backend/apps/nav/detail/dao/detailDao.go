package dao

import (
	"context"
	"errors"

	detailmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/detail/models"
	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	cm "github.com/gofurry/gofurry-nav-backend/common/models"
	navsqlc "github.com/gofurry/gofurry-nav-backend/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5"
)

type DetailDAO struct {
	queries *navsqlc.Queries
}

func New(queries *navsqlc.Queries) *DetailDAO {
	return &DetailDAO{queries: queries}
}

func (dao *DetailDAO) GetSiteByID(siteID int64) (navmodels.GfnSite, common.GFError) {
	row, err := dao.queries.GetPublicSiteByID(context.Background(), siteID)
	if errors.Is(err, pgx.ErrNoRows) {
		return navmodels.GfnSite{}, common.NewDaoError("404")
	}
	if err != nil {
		return navmodels.GfnSite{}, common.NewDaoError(err.Error())
	}
	return navmodels.GfnSite{
		ID: row.ID, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn,
		CreateTime: cm.LocalTime(row.CreateTime.Time), UpdateTime: cm.LocalTime(row.UpdateTime.Time),
		Country: row.Country, Nsfw: row.Nsfw, Welfare: row.Welfare,
		Icon: row.Icon, Deleted: row.Deleted, ViewCount: row.ViewCount,
	}, nil
}

func (dao *DetailDAO) ListCollectorDomains(siteID int64) ([]detailmodels.CollectorDomain, common.GFError) {
	rows, err := dao.queries.ListPublicCollectorDomains(context.Background(), &siteID)
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	result := make([]detailmodels.CollectorDomain, 0, len(rows))
	for _, row := range rows {
		result = append(result, detailmodels.CollectorDomain{
			ID: row.ID, SiteID: pointerValue(row.SiteID), Name: row.Name, Proxy: row.Proxy,
			Prefix: row.Prefix, TLS: row.Tls, Deleted: row.Deleted,
		})
	}
	return result, nil
}

func pointerValue(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}
