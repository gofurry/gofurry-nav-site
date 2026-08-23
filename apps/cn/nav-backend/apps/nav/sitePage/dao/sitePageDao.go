package dao

import (
	"context"

	sitemodel "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	cm "github.com/gofurry/gofurry-nav-backend/common/models"
	navsqlc "github.com/gofurry/gofurry-nav-backend/internal/db/nav/sqlc"
)

type SitePageDAO struct {
	queries *navsqlc.Queries
}

func New(queries *navsqlc.Queries) *SitePageDAO {
	return &SitePageDAO{queries: queries}
}

func (dao *SitePageDAO) GetSiteById(id int64) (sitemodel.GfnSite, common.GFError) {
	row, err := dao.queries.GetSiteByID(context.Background(), id)
	if err != nil {
		return sitemodel.GfnSite{}, common.NewDaoError(err.Error())
	}
	return sitemodel.GfnSite{
		ID: row.ID, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn,
		CreateTime: cm.LocalTime(row.CreateTime.Time), UpdateTime: cm.LocalTime(row.UpdateTime.Time),
		Country: row.Country, Nsfw: row.Nsfw, Welfare: row.Welfare,
		Icon: row.Icon, Deleted: row.Deleted, ViewCount: row.ViewCount,
	}, nil
}

func (dao *SitePageDAO) UpdateViewCount(siteID int64, viewCount int64) common.GFError {
	err := dao.queries.UpdateSiteViewCount(context.Background(), navsqlc.UpdateSiteViewCountParams{
		SiteID: siteID, ViewCount: viewCount,
	})
	if err != nil {
		return common.NewDaoError(err.Error())
	}
	return nil
}
