package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/home/service"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type homeApi struct{}

var HomeApi *homeApi

func init() {
	HomeApi = &homeApi{}
}

func (api homeApi) GetHome(c fiber.Ctx) error {
	data := service.GetCachedHome(c.Query("lang", "zh"))
	return common.NewResponse(c).SuccessWithData(data)
}

func (api homeApi) GetHomePing(c fiber.Ctx) error {
	data := service.GetHomeService().GetHomePing()
	return common.NewResponse(c).SuccessWithData(data)
}

func (api homeApi) GetHomeSaying(c fiber.Ctx) error {
	data := service.GetHomeService().GetHomeSaying(c.Query("lang", "zh"))
	return common.NewResponse(c).SuccessWithData(data)
}

func (api homeApi) GetHomeBackgrounds(c fiber.Ctx) error {
	data := service.GetHomeService().GetHomeBackgrounds()
	return common.NewResponse(c).SuccessWithData(data)
}
