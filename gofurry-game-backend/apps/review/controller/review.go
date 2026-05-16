package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-game-backend/apps/review/models"
	"github.com/gofurry/gofurry-game-backend/apps/review/service"
	"github.com/gofurry/gofurry-game-backend/common"
)

type reviewApi struct{}

var ReviewApi *reviewApi

func init() {
	ReviewApi = &reviewApi{}
}

// @Summary 提交评论
// @Schemes
// @Description 提交评论
// @Tags Review
// @Accept json
// @Produce json
// @Param body body models.AnonymousReviewRequest true "请求body"
// @Success 200 {object} common.ResultData
// @Router /api/review/anonymous [Post]
func (api *reviewApi) AddAnonymousReview(c fiber.Ctx) error {
	req := models.AnonymousReviewRequest{}
	if err := c.Bind().Body(&req); err != nil {
		return common.NewResponse(c).Error("解析请求体失败")
	}
	err := service.GetReviewService().AddAnonymousReview(req, c)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).Success()
}

// @Summary 获取最新评论
// @Schemes
// @Description 获取最新评论
// @Tags Review
// @Accept json
// @Produce json
// @Param lang query string true "语言"
// @Success 200 {object} []models.AnonymousReviewResponse
// @Router /api/review/latest [Get]
func (api *reviewApi) GetLatestReviewList(c fiber.Ctx) error {
	lang := c.Query("lang", "zh")
	data, err := service.GetReviewService().GetLatestReviewList(lang)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}
