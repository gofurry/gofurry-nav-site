package controller

import (
	"github.com/gofurry/gofurry-nav-backend/apps/nav/sitePage/service"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/util"
	"github.com/gofiber/fiber/v3"
)

type sitePageApi struct{}

var SitePageApi *sitePageApi

func init() {
	SitePageApi = &sitePageApi{}
}

// @Summary 获取单个站点信息
// @Schemes
// @Description 获取单个站点信息 lang = zh 或 en 默认 zh
// @Tags Site
// @Accept json
// @Produce json
// @Param id query string true "站点id"
// @Param lang query string true "语言"
// @Success 200 {object} models.SiteInfoVo
// @Router /api/nav/site/getSiteDetail [Get]
func (api sitePageApi) GetSiteDetail(c fiber.Ctx) error {
	id, lang := c.Query("id"), c.Query("lang")
	data, err := service.GetSitePageService().GetSiteDetailService(id, lang, util.GetClientIP(c))
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取单个站点的 HTTP 记录
// @Schemes
// @Description 获取单个站点的 HTTP 记录
// @Tags Site
// @Accept json
// @Produce json
// @Param domain query string true "域名"
// @Success 200 {object} common.ResultData
// @Router /api/nav/site/getSiteHttpRecord [Get]
func (api sitePageApi) GetSiteHttpRecord(c fiber.Ctx) error {
	domain := c.Query("domain")
	data, err := service.GetSitePageService().GetSiteHttpRecordService(domain)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取单个站点的 DNS 记录
// @Schemes
// @Description 获取单个站点的 DNS 记录
// @Tags Site
// @Accept json
// @Produce json
// @Param domain query string true "域名"
// @Success 200 {object} common.ResultData
// @Router /api/nav/site/getSiteDnsRecord [Get]
func (api sitePageApi) GetSiteDnsRecord(c fiber.Ctx) error {
	domain := c.Query("domain")
	data, err := service.GetSitePageService().GetSiteDnsRecordService(domain)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取单个站点的 Ping 记录
// @Schemes
// @Description 获取单个站点的 Ping 记录
// @Tags Site
// @Accept json
// @Produce json
// @Param domain query string true "域名"
// @Success 200 {object} models.SiteDelayVo
// @Router /api/nav/site/getSitePingRecord [Get]
func (api sitePageApi) GetSitePingRecord(c fiber.Ctx) error {
	domain := c.Query("domain")
	data, err := service.GetSitePageService().GetSitePingRecordService(domain)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}
