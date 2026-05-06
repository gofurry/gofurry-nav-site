package controller

import (
	"github.com/GoFurry/gofurry-game-backend/apps/game/service"
	"github.com/GoFurry/gofurry-game-backend/common"
	"github.com/GoFurry/gofurry-game-backend/common/util"
	"github.com/gofiber/fiber/v3"
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
// @Router /api/game/info/list [Get]
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
// @Router /api/game/info/main [Get]
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
// @Router /api/game/panel/main [Get]
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
// @Router /api/game/update/latest [Get]
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
// @Router /api/game/tag/list [Get]
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
// @Router /api/game/info [Get]
func (api *gameApi) GetGameInfo(c fiber.Ctx) error {
	num := c.Query("id", "0")
	lang := c.Query("lang", "zh")
	data, err := service.GetGameService().GetGameInfoWithViewCount(num, lang, util.GetClientIP(c))
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
// @Router /api/game/remark [Get]
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
// @Router /api/game/creator [Get]
func (api *gameApi) GetGameCreator(c fiber.Ctx) error {
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
// @Router /api/game/update/latest/more [Get]
func (api *gameApi) GetUpdateNewsMore(c fiber.Ctx) error {
	lang := c.Query("lang", "zh")
	data, err := service.GetGameService().GetMoreUpdateNews(lang)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}
