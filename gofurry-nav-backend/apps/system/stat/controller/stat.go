package controller

import (
	"github.com/gofurry/gofurry-nav-backend/apps/system/stat/service"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofiber/fiber/v3"
)

type statApi struct{}

var StatApi *statApi

func init() {
	StatApi = &statApi{}
}

// @Summary 查询浏览数
// @Schemes
// @Description 查询浏览数
// @Tags Stat
// @Accept json
// @Produce json
// @Success 200 {object} models.ViewsCountVo
// @Router /api/stat/chart/views/count [Get]
func (api *statApi) GetViewsCount(c fiber.Ctx) error {
	data, err := service.GetStatService().ViewsCount()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 查询内容最多的分组
// @Schemes
// @Description 查询内容最多的分组
// @Tags Stat
// @Accept json
// @Produce json
// @Param lang query string true "语言"
// @Success 200 {object} []models.GroupCountVo
// @Router /api/stat/chart/group/count [Get]
func (api *statApi) GetGroupCount(c fiber.Ctx) error {
	lang := c.Query("lang", "zh")
	data, err := service.GetStatService().GroupCount(lang)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 查询访问最多的国家
// @Schemes
// @Description 查询访问最多的国家
// @Tags Stat
// @Accept json
// @Produce json
// @Success 200 {object} models.RegionCountVo
// @Router /api/stat/chart/views/region/country [Get]
func (api *statApi) GetCountryCount(c fiber.Ctx) error {
	data, err := service.GetStatService().CountryCount()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 查询访问最多的省
// @Schemes
// @Description 查询访问最多的省
// @Tags Stat
// @Accept json
// @Produce json
// @Success 200 {object} models.RegionCountVo
// @Router /api/stat/chart/views/region/province [Get]
func (api *statApi) GetProvinceCount(c fiber.Ctx) error {
	data, err := service.GetStatService().ProvinceCount()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 查询访问最多的城市
// @Schemes
// @Description 查询访问最多的城市
// @Tags Stat
// @Accept json
// @Produce json
// @Success 200 {object} models.RegionCountVo
// @Router /api/stat/chart/views/region/city [Get]
func (api *statApi) GetCityCount(c fiber.Ctx) error {
	data, err := service.GetStatService().CityCount()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取近日收录的站点
// @Schemes
// @Description 获取近日收录的站点
// @Tags Stat
// @Accept json
// @Produce json
// @Param lang query string true "语言"
// @Success 200 {object} []models.SiteListVo
// @Router /api/stat/nav/site/list [Get]
func (api *statApi) GetSiteList(c fiber.Ctx) error {
	lang := c.Query("lang", "zh")
	data, err := service.GetStatService().SiteList(lang)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取导航站点的基本数据
// @Schemes
// @Description 获取导航站点的基本数据
// @Tags Stat
// @Accept json
// @Produce json
// @Success 200 {object} models.SiteCommonInfoVo
// @Router /api/stat/nav/site/common [Get]
func (api *statApi) GetSiteCommonInfo(c fiber.Ctx) error {
	data, err := service.GetStatService().SiteCommonInfo()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取最近的最高延迟的 ping 记录
// @Schemes
// @Description 获取最近的最高延迟的 ping 记录
// @Tags Stat
// @Accept json
// @Produce json
// @Success 200 {object} models.SiteCommonInfoVo
// @Router /api/stat/nav/site/ping/list [Get]
func (api *statApi) GetSitePingList(c fiber.Ctx) error {
	data, err := service.GetStatService().SitePingList()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取接口指标
// @Schemes
// @Description 获取接口指标
// @Tags Stat
// @Accept json
// @Produce json
// @Success 200 {object} models.PromMetricsVo
// @Router /api/stat/prom/metrics [Get]
func (api *statApi) GetPromMetrics(c fiber.Ctx) error {
	data, err := service.GetMetricsService().GetPromMetrics()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取接口时序指标
// @Schemes
// @Description 获取接口时序指标
// @Tags Stat
// @Accept json
// @Produce json
// @Success 200 {object} models.PromMetricsVo
// @Router /api/stat/prom/metrics/history [Get]
func (api *statApi) GetPromMetricsHistory(c fiber.Ctx) error {
	data, err := service.GetMetricsService().GetPromMetricsHistory()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 提供背景随机图片
// @Schemes
// @Description 提供背景随机图片的CDN地址
// @Tags Stat
// @Accept json
// @Produce json
// @Success 200 {object} string
// @Router /api/nav/page/header/image/url [Get]
func (api *statApi) GetImageUrl(c fiber.Ctx) error {
	return common.NewResponse(c).SuccessWithData(service.GetMetricsService().GetImageUrl())
}
