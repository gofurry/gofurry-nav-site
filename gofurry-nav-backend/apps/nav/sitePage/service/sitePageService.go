package service

import (
	"github.com/gofurry/gofurry-nav-backend/apps/nav/sitePage/dao"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type sitePageService struct{}

var sitePageSingleton = new(sitePageService)

func GetSitePageService() *sitePageService { return sitePageSingleton }

func (svc *sitePageService) TouchSiteViewCount(siteID int64, clientIP string) (int64, common.GFError) {
	if siteID <= 0 {
		return 0, common.NewServiceError("siteID 参数非法")
	}
	record, err := dao.GetSitePageDao().GetSiteById(siteID)
	if err != nil {
		return 0, err
	}
	return svc.touchSiteViewCount(siteID, record.ViewCount, clientIP), nil
}
