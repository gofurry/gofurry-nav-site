package controller

import (
	"github.com/gofurry/gofurry-nav-backend/apps/system/site/service"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofiber/fiber/v3"
)

type siteApi struct{}

var SiteApi *siteApi

func init() {
	SiteApi = &siteApi{}
}

// @Summary 查询更新日志
// @Schemes
// @Description 查询更新日志
// @Tags Site
// @Accept json
// @Produce json
// @Success 200 {object} []models.ChangeLogVo
// @Router /api/site/changelog [Get]
func (api *siteApi) GetSiteChangeLog(c fiber.Ctx) error {
	data, err := service.GetSiteService().GetChangeLog()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}
