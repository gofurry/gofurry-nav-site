package controller

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/home/models"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/home/service"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type homeReader interface {
	GetHomePing() models.HomePingResponse
	GetHomeSaying(lang string) models.HomeSayingResponse
	GetHomeHero(context.Context, models.HeroSelection) models.HomeHeroResponse
	GetHeroes(context.Context, models.HeroCatalogQuery) (models.HeroCatalog, common.GFError)
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
	hero := api.service().GetHomeHero(c.Context(), heroSelection(c))
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
	data := api.service().GetHomeHero(c.Context(), heroSelection(c))
	return common.NewResponse(c).SuccessWithData(data)
}

// Malformed/stale fixed IDs intentionally fall back independently to random.
func positiveID(raw string) int64 {
	if raw == "" || strings.Trim(raw, "0123456789") != "" {
		return 0
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0
	}
	return id
}

func heroSelection(c fiber.Ctx) models.HeroSelection {
	return models.HeroSelection{DesktopID: positiveID(c.Query("hero_desktop_id")), MobileID: positiveID(c.Query("hero_mobile_id")), Local: c.Query("hero_mode") == "local"}
}

func (api homeApi) GetHeroes(c fiber.Ctx) error {
	variant := c.Query("variant")
	page := positiveID(c.Query("page_num", "1"))
	size := positiveID(c.Query("page_size", "12"))
	selectedRaw := c.Query("selected_id")
	selected := positiveID(selectedRaw)
	if (variant != "desktop" && variant != "mobile") || page == 0 || page > 1_000_000 || size == 0 || size > 24 || (selectedRaw != "" && selected == 0) {
		return common.NewResponse(c).ErrorWithCode("Invalid Hero catalog parameters", fiber.StatusBadRequest)
	}
	data, err := api.service().GetHeroes(c.Context(), models.HeroCatalogQuery{Variant: variant, PageNum: int32(page), PageSize: int32(size), SelectedID: selected})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api homeApi) GetPatterns(c fiber.Ctx) error {
	data, err := api.service().GetPatterns(c.Context())
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(data)
}
