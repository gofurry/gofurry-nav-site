package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/sitegroup/service"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type siteGroupApi struct{}

var SiteGroupApi *siteGroupApi

func init() {
	SiteGroupApi = &siteGroupApi{}
}

func (api siteGroupApi) GetSiteGroupPage(c fiber.Ctx) error {
	page := service.ParsePositiveInt(c.Query("page", "1"), 1)
	pageSize := service.ParsePositiveInt(c.Query("page_size", "24"), 24)
	data := service.GetCachedSiteGroupPage(c.Query("lang", "zh"), c.Params("groupId"), page, pageSize)
	return common.NewResponse(c).SuccessWithData(data)
}
