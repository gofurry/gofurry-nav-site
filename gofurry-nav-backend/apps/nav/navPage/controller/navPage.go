package controller

import (
	"github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/service"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofiber/fiber/v3"
)

type navPageApi struct{}

var NavPageApi *navPageApi

func init() {
	NavPageApi = &navPageApi{}
}

// @Summary 获取所有导航站点信息
// @Schemes
// @Description 获取所有导航站点信息, lang= zh 或 en 默认 zh
// @Tags Nav
// @Accept json
// @Produce json
// @Param lang query string true "语言"
// @Success 200 {object} []models.SiteVo
// @Router /api/nav/page/site/list [Get]
func (api *navPageApi) GetSiteList(c fiber.Ctx) error {
	lang := c.Query("lang", "zh")
	data, err := service.GetNavPageService().GetSiteList(lang)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取所有导航站点分组信息
// @Schemes
// @Description 获取所有导航站点分组信息, lang= zh 或 en 默认 zh
// @Tags Nav
// @Accept json
// @Produce json
// @Param lang query string true "语言"
// @Success 200 {object} []models.GroupVo
// @Router /api/nav/page/group/list [Get]
func (api *navPageApi) GetGroupList(c fiber.Ctx) error {
	lang := c.Query("lang", "zh")
	data, err := service.GetNavPageService().GetGroupList(lang)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取所有导航站点延迟信息
// @Schemes
// @Description 获取所有导航站点延迟信息
// @Tags Nav
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/nav/page/ping/list [Get]
func (api *navPageApi) GetPingList(c fiber.Ctx) error {
	data, err := service.GetNavPageService().GetPingList()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取百度搜索建议
// @Schemes
// @Description 获取百度搜索建议
// @Tags Nav-search
// @Accept json
// @Produce json
// @Param q query string true "查询"
// @Success 200 {object} []string
// @Router /api/nav/page/search/baidu [Get]
func (api *navPageApi) GetBaiduSearchSuggestion(c fiber.Ctx) error {
	q := c.Query("q")
	data, err := service.GetNavPageService().GetBaiduSuggestion(q)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取必应搜索建议
// @Schemes
// @Description 获取必应搜索建议
// @Tags Nav-search
// @Accept json
// @Produce json
// @Param q query string true "查询"
// @Success 200 {object} []string
// @Router /api/nav/page/search/bing [Get]
func (api *navPageApi) GetBingSearchSuggestion(c fiber.Ctx) error {
	q := c.Query("q")
	data, err := service.GetNavPageService().GetBingSuggestion(q)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取谷歌搜索建议
// @Schemes
// @Description 获取谷歌搜索建议
// @Tags Nav-search
// @Accept json
// @Produce json
// @Param q query string true "查询"
// @Success 200 {object} []string
// @Router /api/nav/page/search/google [Get]
func (api *navPageApi) GetGoogleSearchSuggestion(c fiber.Ctx) error {
	q := c.Query("q")
	data, err := service.GetNavPageService().GetGoogleSuggestion(q)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取b站搜索建议
// @Schemes
// @Description 获取b站搜索建议
// @Tags Nav-search
// @Accept json
// @Produce json
// @Param q query string true "查询"
// @Success 200 {object} []string
// @Router /api/nav/page/search/bilibili [Get]
func (api *navPageApi) GetBiliBiliSearchSuggestion(c fiber.Ctx) error {
	q := c.Query("q")
	data, err := service.GetNavPageService().GetBiliBiliSuggestion(q)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取随机金句
// @Schemes
// @Description 获取随机金句
// @Tags Nav-search
// @Accept json
// @Produce json
// @Success 200 {object} string
// @Router /api/nav/page/header/getSaying [Get]
func (api *navPageApi) GetSaying(c fiber.Ctx) error {
	saying, err := service.GetNavPageService().GetSayingService()
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(saying)
}

// @Summary 提供背景随机图片
// @Schemes
// @Description 提供背景随机图片的CDN地址, type= resized 或 normal 默认 normal
// @Tags Nav-search
// @Accept json
// @Produce json
// @Param type query string true "图片类型"
// @Success 200 {object} string
// @Router /api/nav/page/header/image/url [Get]
func (api *navPageApi) GetImageUrl(c fiber.Ctx) error {
	return common.NewResponse(c).SuccessWithData(service.GetNavPageService().GetImageUrl(c.Query("type", "normal")))
}
