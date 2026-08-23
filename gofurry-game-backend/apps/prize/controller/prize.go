package controller

import (
	"net/url"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-game-backend/apps/prize/models"
	"github.com/gofurry/gofurry-game-backend/apps/prize/service"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/util"
	"github.com/gofurry/gofurry-game-backend/roof/env"
)

type PrizeAPI struct {
	service *service.PrizeService
}

func New(prizeService *service.PrizeService) *PrizeAPI {
	return &PrizeAPI{service: prizeService}
}

const defaultPrizeActivationFrontendURL = "https://go-furry.com/games/prize/activation"

func prizeActivationFrontendURL() string {
	value := env.GetServerConfig().Prize.ActivationFrontendURL
	if value == "" {
		return defaultPrizeActivationFrontendURL
	}
	return value
}

func prizeActivationRedirectURL(status string, msg string) string {
	frontendURL := prizeActivationFrontendURL()
	u, err := url.Parse(frontendURL)
	if err != nil {
		return defaultPrizeActivationFrontendURL
	}
	q := u.Query()
	q.Set("status", status)
	q.Set("msg", msg)
	u.RawQuery = q.Encode()
	return u.String()
}

func prizeActivationMessage(p models.GfgPrize, m models.GfgPrizeMember, err common.GFError) string {
	name := m.Name
	email := util.MaskEmail(m.Email)
	title := p.Title
	if name == "" {
		name = "GoFurry 用户"
	}
	if email == "" {
		email = "未知邮箱"
	}
	if title == "" {
		title = "GoFurry 抽奖"
	}

	msg := "尊敬的 [" + name + "-" + email + "], 您参加的 [" + title + "] 抽奖活动报名"
	if err != nil {
		return msg + "失败: " + err.GetMsg()
	}
	return msg + "成功"
}

// @Schemes
// @Description 参与抽奖
func (api *PrizeAPI) PrizeParticipation(c fiber.Ctx) error {
	req := models.PrizeParticipationRequest{}
	if err := c.Bind().Body(&req); err != nil {
		return common.NewResponse(c).Error("解析请求体失败")
	}
	err := api.service.PrizeParticipation(req, c)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).Success()
}

// @Schemes
// @Description 激活抽奖申请
func (api *PrizeAPI) ActiveParticipation(c fiber.Ctx) error {
	id := c.Query("id")
	key := c.Query("key")
	p, m, err := api.service.ActiveParticipation(id, key)

	if err != nil {
		return c.Redirect().Status(fiber.StatusFound).To(prizeActivationRedirectURL("fail", prizeActivationMessage(p, m, err)))
	}
	return c.Redirect().Status(fiber.StatusFound).To(prizeActivationRedirectURL("success", prizeActivationMessage(p, m, nil)))
}

// @Schemes
// @Description 抽奖详情
func (api *PrizeAPI) LotteryInfo(c fiber.Ctx) error {
	data, err := api.service.LotteryInfo()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}
