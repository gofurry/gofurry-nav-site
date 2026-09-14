package controller

import (
	"context"
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/home/models"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/home/service"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type homeReader interface {
	GetHomePing() models.HomePingResponse
	GetHomeSaying(lang string) models.HomeSayingResponse
	GetHomeHero(context.Context) models.HomeHeroResponse
	GetPatterns(context.Context) (models.PatternCatalog, common.GFError)
}

type homeApi struct{ reader homeReader }

var HomeApi *homeApi

func init() {
	HomeApi = &homeApi{}
}

func New(reader homeReader) *homeApi { return &homeApi{reader: reader} }

func (api homeApi) service() homeReader {
	if api.reader != nil {
		return api.reader
	}
	return service.GetHomeService()
}

func (api homeApi) GetHome(c fiber.Ctx) error {
	data := service.GetCachedHome(c.Query("lang", "zh"))
	hero := api.service().GetHomeHero(c.Context())
	data.Hero = hero.Hero
	data.CacheState["hero"] = hero.State
	return common.NewResponse(c).SuccessWithData(data)
}

func (api homeApi) GetHomePing(c fiber.Ctx) error {
	data := api.service().GetHomePing()
	return common.NewResponse(c).SuccessWithData(data)
}

func (api homeApi) GetHomeSaying(c fiber.Ctx) error {
	data := api.service().GetHomeSaying(c.Query("lang", "zh"))
	return common.NewResponse(c).SuccessWithData(data)
}

func (api homeApi) GetHomeHero(c fiber.Ctx) error {
	data := api.service().GetHomeHero(c.Context())
	return common.NewResponse(c).SuccessWithData(data)
}

func (api homeApi) GetPatterns(c fiber.Ctx) error {
	data, err := api.service().GetPatterns(c.Context())
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(data)
}
