package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-game-backend/apps/game/service"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/util"
)

type gameApi struct{}

var GameApi *gameApi

func init() {
	GameApi = &gameApi{}
}

// @Summary 获取所有游戏的记录
// @Schemes
// @Description 获取所有游戏记录
// @Tags Game
// @Accept json
// @Produce json
// @Param num query string true "请求数量"
// @Param lang query string true "语言"
// @Success 200 {object} []models.GameRespVo
// @Router /api/v1/game/info/list [Get]
func (api *gameApi) GetGameList(c fiber.Ctx) error {
	num := c.Query("num", "100")
	lang := c.Query("lang", "zh")
	data, err := service.GetGameService().GetGameList(num, lang)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取首页展示数据
// @Schemes
// @Description 获取首页展示数据
// @Tags Game
// @Accept json
// @Produce json
// @Success 200 {object} models.GameMainInfoVo
// @Router /api/v1/game/info/main [Get]
func (api *gameApi) GetGameMainList(c fiber.Ctx) error {
	data, err := service.GetGameService().GetGameMainList()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取首页面板数据
// @Schemes
// @Description 获取首页面板数据
// @Tags Game
// @Accept json
// @Produce json
// @Success 200 {object} models.GameMainPanelVo
// @Router /api/v1/game/panel/main [Get]
func (api *gameApi) GetPanelMainList(c fiber.Ctx) error {
	data, err := service.GetGameService().GetPanelMainList()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取首页更新公告
// @Schemes
// @Description 获取首页更新公告
// @Tags Game
// @Accept json
// @Produce json
// @Success 200 {object} models.UpdateNewsVo
// @Router /api/v1/game/update/latest [Get]
func (api *gameApi) GetUpdateNews(c fiber.Ctx) error {
	data, err := service.GetGameService().GetUpdateNews()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取标签列表
// @Schemes
// @Description 获取标签列表
// @Tags Game
// @Accept json
// @Produce json
// @Param lang query string true "语言"
// @Success 200 {object} []models.TagModelVo
// @Router /api/v1/game/tag/list [Get]
func (api *gameApi) GetTagList(c fiber.Ctx) error {
	lang := c.Query("lang", "zh")
	data, err := service.GetGameService().GetTagList(lang)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取游戏的基本信息
// @Schemes
// @Description 获取游戏ID对应的基本信息
// @Tags Game
// @Accept json
// @Produce json
// @Param id query string true "游戏id"
// @Param lang query string true "语言"
// @Success 200 {object} []models.GameBaseInfoVo
// @Router /api/v1/game/info [Get]
func (api *gameApi) GetGameInfo(c fiber.Ctx) error {
	num := c.Query("id", "0")
	lang := c.Query("lang", "zh")
	data, err := service.GetGameService().GetGameInfoWithViewCount(num, lang, util.GetClientIP(c))
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取同步用的游戏列表
// @Schemes
// @Description 返回全量游戏轻量列表，供 RAG 同步使用
// @Tags Game
// @Accept json
// @Produce json
// @Param lang query string true "语言"
// @Success 200 {object} []models.GameRespVo
// @Router /api/v1/game/sync/list [Get]
func (api *gameApi) GetGameSyncList(c fiber.Ctx) error {
	lang := c.Query("lang", "zh")
	data, err := service.GetGameService().GetGameSyncList(lang)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取同步用的游戏详情
// @Schemes
// @Description 获取游戏 ID 对应的基础信息，不触发浏览量统计
// @Tags Game
// @Accept json
// @Produce json
// @Param id query string true "游戏id"
// @Param lang query string true "语言"
// @Success 200 {object} []models.GameBaseInfoVo
// @Router /api/v1/game/sync/info [Get]
func (api *gameApi) GetGameSyncInfo(c fiber.Ctx) error {
	num := c.Query("id", "0")
	lang := c.Query("lang", "zh")
	data, err := service.GetGameService().GetGameInfo(num, lang)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取游戏的评论
// @Schemes
// @Description 获取游戏ID对应的评论
// @Tags Game
// @Accept json
// @Produce json
// @Param id query string true "游戏id"
// @Success 200 {object} []models.GameRemarkVo
// @Router /api/v1/game/remark [Get]
func (api *gameApi) GetGameRemark(c fiber.Ctx) error {
	num := c.Query("id", "0")
	data, err := service.GetGameService().GetGameRemark(num)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取游戏创作者列表
// @Schemes
// @Description 获取游戏创作者列表
// @Tags Game
// @Accept json
// @Produce json
// @Param lang query string true "语言"
// @Success 200 {object} []models.CreatorVo
// @Router /api/v1/game/creator [Get]
func (api *gameApi) GetGameCreator(c fiber.Ctx) error {
	lang := c.Query("lang", "zh")
	data, err := service.GetGameService().GetGameCreator(lang)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取同步用的创作者列表
// @Schemes
// @Description 获取创作者列表，供 RAG 同步使用
// @Tags Game
// @Accept json
// @Produce json
// @Param lang query string true "语言"
// @Success 200 {object} []models.CreatorVo
// @Router /api/v1/game/sync/creators [Get]
func (api *gameApi) GetGameSyncCreators(c fiber.Ctx) error {
	lang := c.Query("lang", "zh")
	data, err := service.GetGameService().GetGameCreator(lang)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取更多首页更新公告
// @Schemes
// @Description 获取更多首页更新公告
// @Tags Game
// @Accept json
// @Produce json
// @Param lang query string true "语言"
// @Success 200 {object} []models.UpdateNewsModels
// @Router /api/v1/game/update/latest/more [Get]
func (api *gameApi) GetUpdateNewsMore(c fiber.Ctx) error {
	lang := c.Query("lang", "zh")
	data, err := service.GetGameService().GetMoreUpdateNews(lang)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 获取同步用的游戏新闻
// @Schemes
// @Description 获取全量游戏新闻列表，供 RAG 同步使用
// @Tags Game
// @Accept json
// @Produce json
// @Param lang query string true "语言"
// @Success 200 {object} []models.UpdateNewsModels
// @Router /api/v1/game/sync/news [Get]
func (api *gameApi) GetGameSyncNews(c fiber.Ctx) error {
	lang := c.Query("lang", "zh")
	data, err := service.GetGameService().GetMoreUpdateNews(lang)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}
