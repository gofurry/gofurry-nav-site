package dao

import (
	"context"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	cm "github.com/gofurry/gofurry-nav-backend/common/models"
	navsqlc "github.com/gofurry/gofurry-nav-backend/internal/db/nav/sqlc"
)

type NavPageDAO struct {
	queries *navsqlc.Queries
}

func New(queries *navsqlc.Queries) *NavPageDAO {
	return &NavPageDAO{queries: queries}
}

func (dao *NavPageDAO) GetSiteList() ([]models.GfnSite, common.GFError) {
	rows, err := dao.queries.ListPublicSites(context.Background())
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	result := make([]models.GfnSite, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.GfnSite{
			ID: row.ID, Name: row.Name, NameEn: row.NameEn, Domain: row.Domain,
			Info: row.Info, InfoEn: row.InfoEn,
			CreateTime: cm.LocalTime(row.CreateTime.Time), UpdateTime: cm.LocalTime(row.UpdateTime.Time),
			Country: row.Country, Nsfw: row.Nsfw, Welfare: row.Welfare,
			ViewCount: row.ViewCount, Icon: row.Icon, Deleted: row.Deleted,
		})
	}
	return result, nil
}

func (dao *NavPageDAO) GetSiteIndexList() ([]models.GfnSiteIndex, common.GFError) {
	rows, err := dao.queries.ListPublicSiteIndex(context.Background())
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	result := make([]models.GfnSiteIndex, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.GfnSiteIndex{
			ID: row.ID, Domain: row.Domain, UpdateTime: cm.LocalTime(row.UpdateTime.Time),
		})
	}
	return result, nil
}

func (dao *NavPageDAO) GetGroupList() ([]models.GfnSiteGroup, common.GFError) {
	rows, err := dao.queries.ListSiteGroups(context.Background())
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	result := make([]models.GfnSiteGroup, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.GfnSiteGroup{
			ID: row.ID, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn,
			Priority: row.Priority, CreateTime: row.CreateTime.Time, UpdateTime: row.UpdateTime.Time,
		})
	}
	return result, nil
}

func (dao *NavPageDAO) GetGroupMapList() ([]models.GfnSiteGroupMap, common.GFError) {
	rows, err := dao.queries.ListSiteGroupMappings(context.Background())
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	result := make([]models.GfnSiteGroupMap, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.GfnSiteGroupMap{
			ID: row.ID, SiteID: row.SiteID, GroupID: row.GroupID, Weight: row.Weight,
			CreateTime: row.CreateTime.Time, UpdateTime: row.UpdateTime.Time,
		})
	}
	return result, nil
}

func (dao *NavPageDAO) GetFeaturedSiteList() ([]models.GfnFeaturedSite, common.GFError) {
	rows, err := dao.queries.ListFeaturedSites(context.Background())
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	result := make([]models.GfnFeaturedSite, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.GfnFeaturedSite{
			ID: row.ID, SiteID: row.SiteID, Weight: row.Weight,
			CreateTime: row.CreateTime.Time, UpdateTime: row.UpdateTime.Time,
		})
	}
	return result, nil
}

func (dao *NavPageDAO) GetSayingByRandom(lang string) (*models.GfnSaying, common.GFError) {
	row, err := dao.queries.GetRandomSaying(context.Background(), lang)
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	return &models.GfnSaying{
		ID: row.ID, Author: row.Author, Language: row.Language, Saying: row.Saying,
		CreateTime: row.CreateTime.Time, UpdateTime: row.UpdateTime.Time,
	}, nil
}
