package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-game-backend/apps/recommend/service"
	"github.com/gofurry/gofurry-game-backend/common"
)

type recommendApi struct{}

var RecommendApi *recommendApi

func init() {
	RecommendApi = &recommendApi{}
}

// @Summary 随机返回一个游戏记录ID
// @Schemes
// @Description 随机返回一个游戏记录ID
// @Tags Recommend
// @Accept json
// @Produce json
// @Success 200 {object} string
// @Router /api/recommend/game/random [Get]
func (api *recommendApi) GetRandomGameID(c fiber.Ctx) error {
	data, err := service.GetRecommendService().GetRandomGameID()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary CBF 返回游戏记录列表
// @Schemes
// @Description CBF 返回游戏记录列表
// @Tags Recommend
// @Accept json
// @Produce json
// @Param id query string true "初始id"
// @Param lang query string true "语言"
// @Success 200 {object} []models.GameRecommendVo
// @Router /api/recommend/game/CBF [Get]
func (api *recommendApi) RecommendByCBF(c fiber.Ctx) error {
	id := c.Query("id", "-1")
	lang := c.Query("lang", "zh")
	data, err := service.GetRecommendService().RecommendByCBF(id, lang)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}
