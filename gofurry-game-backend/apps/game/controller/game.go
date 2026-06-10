package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-game-backend/apps/game/service"
	"github.com/gofurry/gofurry-game-backend/common"
)

type gameApi struct{}

var GameApi *gameApi

func init() {
	GameApi = &gameApi{}
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
