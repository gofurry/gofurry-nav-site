package service

import (
	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type siteStore interface {
	GetSiteById(id int64) (navmodels.GfnSite, common.GFError)
}

type sitePageService struct{ store siteStore }

var sitePageSingleton = new(sitePageService)

func GetSitePageService() *sitePageService { return sitePageSingleton }

func New(store siteStore) *sitePageService { return &sitePageService{store: store} }

func (svc *sitePageService) TouchSiteViewCount(siteID int64, clientIP string) (int64, common.GFError) {
	if siteID <= 0 {
		return 0, common.NewServiceError("siteID 参数非法")
	}
	record, err := svc.store.GetSiteById(siteID)
	if err != nil {
		return 0, err
	}
	return svc.touchSiteViewCount(siteID, record.ViewCount, clientIP), nil
}
