package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/service"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type navPageApi struct{}

var NavPageApi *navPageApi

func init() {
	NavPageApi = &navPageApi{}
}

// @Schemes
// @Description 获取所有导航站点信息, lang= zh 或 en 默认 zh
func (api *navPageApi) GetSiteList(c fiber.Ctx) error {
	lang := c.Query("lang", "zh")
	data, err := service.GetNavPageService().GetSiteList(lang)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Schemes
// @Description 获取所有导航站点分组信息, lang= zh 或 en 默认 zh
func (api *navPageApi) GetGroupList(c fiber.Ctx) error {
	lang := c.Query("lang", "zh")
	data, err := service.GetNavPageService().GetGroupList(lang)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Schemes
// @Description 获取所有导航站点延迟信息
func (api *navPageApi) GetPingList(c fiber.Ctx) error {
	data, err := service.GetNavPageService().GetPingList()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Schemes
// @Description 获取百度搜索建议
func (api *navPageApi) GetBaiduSearchSuggestion(c fiber.Ctx) error {
	q := c.Query("q")
	data, err := service.GetNavPageService().GetBaiduSuggestion(q)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Schemes
// @Description 获取必应搜索建议
func (api *navPageApi) GetBingSearchSuggestion(c fiber.Ctx) error {
	q := c.Query("q")
	data, err := service.GetNavPageService().GetBingSuggestion(q)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Schemes
// @Description 获取谷歌搜索建议
func (api *navPageApi) GetGoogleSearchSuggestion(c fiber.Ctx) error {
	q := c.Query("q")
	data, err := service.GetNavPageService().GetGoogleSuggestion(q)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Schemes
// @Description 获取b站搜索建议
func (api *navPageApi) GetBiliBiliSearchSuggestion(c fiber.Ctx) error {
	q := c.Query("q")
	data, err := service.GetNavPageService().GetBiliBiliSuggestion(q)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Schemes
// @Description 获取随机金句
func (api *navPageApi) GetSaying(c fiber.Ctx) error {
	saying, err := service.GetNavPageService().GetSayingService(c.Query("lang", "zh"))
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(saying)
}

// @Schemes
// @Description 提供背景随机图片的CDN地址, type= resized 或 normal 默认 normal
func (api *navPageApi) GetImageUrl(c fiber.Ctx) error {
	return common.NewResponse(c).SuccessWithData(service.GetNavPageService().GetImageUrl(c.Query("type", "normal")))
}
