package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-game-backend/apps/search/models"
	"github.com/gofurry/gofurry-game-backend/apps/search/service"
	"github.com/gofurry/gofurry-game-backend/common"
)

type searchApi struct{}

var SearchApi *searchApi

func init() {
	SearchApi = &searchApi{}
}

// @Summary 简易搜索
// @Schemes
// @Description 简易搜索
// @Tags Search
// @Accept json
// @Produce json
// @Param body body models.SearchRequest true "请求body"
// @Success 200 {object} models.SearchGameVo
// @Router /api/v1/game/search/game/simple [POST]
func (api *searchApi) SimpleSearch(c fiber.Ctx) error {
	req := models.SearchRequest{}
	if err := c.Bind().Body(&req); err != nil {
		return common.NewResponse(c).Error("解析请求体失败")
	}
	data, err := service.GetSearchService().SimpleSearchQuery(req)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 分页高级搜索
// @Schemes
// @Description 分页高级搜索
// @Tags Search
// @Accept json
// @Produce json
// @Param body body models.SearchPageQueryRequest true "请求body"
// @Success 200 {object} models.PageResponse
// @Router /api/v1/game/search/game/page [POST]
func (api *searchApi) PageSearch(c fiber.Ctx) error {
	req := models.SearchPageQueryRequest{}
	if err := c.Bind().Body(&req); err != nil {
		return common.NewResponse(c).Error("解析请求体失败")
	}

	req.InitPageIfAbsent()

	data, err := service.GetSearchService().SearchPageQuery(req)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)

}
