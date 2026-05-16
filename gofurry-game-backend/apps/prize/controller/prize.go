package controller

import (
	"github.com/gofurry/gofurry-game-backend/apps/prize/models"
	"github.com/gofurry/gofurry-game-backend/apps/prize/service"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/util"
	"github.com/gofiber/fiber/v3"
)

type prizeApi struct{}

var PrizeApi *prizeApi

func init() {
	PrizeApi = &prizeApi{}
}

// @Summary 参与抽奖
// @Schemes
// @Description 参与抽奖
// @Tags Prize
// @Accept json
// @Produce json
// @Param body body models.PrizeParticipationRequest true "请求body"
// @Success 200 {object} common.ResultData
// @Router /api/prize/participation [Post]
func (api *prizeApi) PrizeParticipation(c fiber.Ctx) error {
	req := models.PrizeParticipationRequest{}
	if err := c.Bind().Body(&req); err != nil {
		return common.NewResponse(c).Error("解析请求体失败")
	}
	err := service.GetPrizeService().PrizeParticipation(req, c)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).Success()
}

// @Summary 激活抽奖申请
// @Schemes
// @Description 激活抽奖申请
// @Tags Prize
// @Accept json
// @Produce json
// @Param id query string true "抽奖活动id"
// @Param key query string true "激活令牌"
// @Success 200 {object} common.ResultData
// @Router /api/prize/participation/activation [Get]
func (api *prizeApi) ActiveParticipation(c fiber.Ctx) error {
	id := c.Query("id")
	key := c.Query("key")
	p, m, err := service.GetPrizeService().ActiveParticipation(id, key)

	baseURL := "https://go-furry.com/games/prize/activation"
	msg := "尊敬的 [" + m.Name + "-" + util.MaskEmail(m.Email) + "], 您参加的 [" + p.Title + "] 抽奖活动报名"
	if err != nil {
		return c.Redirect().Status(fiber.StatusFound).To(baseURL + "?status=fail&msg=" + msg + "失败: " + err.GetMsg())
	}
	return c.Redirect().Status(fiber.StatusFound).To(baseURL + "?status=success&msg=" + msg + "成功")
}

// @Summary 抽奖详情
// @Schemes
// @Description 抽奖详情
// @Tags Prize
// @Accept json
// @Produce json
// @Success 200 {object} models.LotteryResp
// @Router /api/prize/info [Get]
func (api *prizeApi) LotteryInfo(c fiber.Ctx) error {
	data, err := service.GetPrizeService().LotteryInfo()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}
