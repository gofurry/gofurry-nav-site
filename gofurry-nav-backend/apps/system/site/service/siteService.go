package service

import (
	"github.com/gofurry/gofurry-nav-backend/apps/system/site/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
	"github.com/bytedance/sonic"
)

type siteService struct{}

var siteSingleton = new(siteService)

func GetSiteService() *siteService { return siteSingleton }

// 获取更新公告
func (s siteService) GetChangeLog() (res []models.ChangeLogVo, err common.GFError) {
	jsonStr, gfsError := cs.GetString("site-common:changelog")
	if gfsError != nil {
		common.NewServiceError(gfsError.GetMsg())
	}
	jsonErr := sonic.Unmarshal([]byte(jsonStr), &res)
	if jsonErr != nil {
		return res, common.NewServiceError(jsonErr.Error())
	}
	return
}
