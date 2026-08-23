package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/sitedirectory/service"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type siteDirectoryApi struct{}

var SiteDirectoryApi *siteDirectoryApi

func init() {
	SiteDirectoryApi = &siteDirectoryApi{}
}

func (api siteDirectoryApi) GetSiteDirectory(c fiber.Ctx) error {
	data := service.GetCachedSiteDirectory(c.Query("lang", "zh"))
	return common.NewResponse(c).SuccessWithData(data)
}
