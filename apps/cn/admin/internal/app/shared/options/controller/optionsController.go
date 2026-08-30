package controller

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	gamesqlc "github.com/gofurry/gofurry-admin/internal/db/game/sqlc"
	navsqlc "github.com/gofurry/gofurry-admin/internal/db/nav/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OptionsAPI struct {
	nav  *navsqlc.Queries
	game *gamesqlc.Queries
}

func New(navPool, gamePool *pgxpool.Pool) *OptionsAPI {
	return &OptionsAPI{nav: navsqlc.New(navPool), game: gamesqlc.New(gamePool)}
}

func (api *OptionsAPI) SiteOptions(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	total, err := api.nav.CountSiteOptions(c.Context(), page.Keyword)
	if err != nil {
		return common.NewResponse(c).Error(common.NewDaoError(err.Error()))
	}
	rows, err := api.nav.ListSiteOptions(c.Context(), navsqlc.ListSiteOptionsParams{Keyword: page.Keyword, RowOffset: int32((page.PageNum - 1) * page.PageSize), RowLimit: int32(page.PageSize)})
	if err != nil {
		return common.NewResponse(c).Error(common.NewDaoError(err.Error()))
	}
	list := make([]adminutil.OptionItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, adminutil.OptionItem{ID: row.ID, Label: row.Name, Extra: row.NameEn})
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, list))
}

func (api *OptionsAPI) SiteTargetOptions(c fiber.Ctx) error {
	siteID, err := strconv.ParseInt(strings.TrimSpace(c.Query("site_id", "")), 10, 64)
	if err != nil || siteID <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("site_id must be a positive integer"))
	}
	page := adminutil.ParsePageQuery(c)
	total, queryErr := api.nav.CountSiteTargetOptions(c.Context(), navsqlc.CountSiteTargetOptionsParams{SiteID: &siteID, Keyword: page.Keyword})
	if queryErr != nil {
		return common.NewResponse(c).Error(common.NewDaoError(queryErr.Error()))
	}
	rows, queryErr := api.nav.ListSiteTargetOptions(c.Context(), navsqlc.ListSiteTargetOptionsParams{
		SiteID: &siteID, Keyword: page.Keyword, RowOffset: int32((page.PageNum - 1) * page.PageSize), RowLimit: int32(page.PageSize),
	})
	if queryErr != nil {
		return common.NewResponse(c).Error(common.NewDaoError(queryErr.Error()))
	}
	list := make([]adminutil.OptionItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, adminutil.OptionItem{ID: row.ID, Label: row.Target, Extra: "proxy=" + row.Proxy + " · tls=" + row.Tls})
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, list))
}

func (api *OptionsAPI) SiteGroupOptions(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	total, err := api.nav.CountSiteGroupOptions(c.Context(), page.Keyword)
	if err != nil {
		return common.NewResponse(c).Error(common.NewDaoError(err.Error()))
	}
	rows, err := api.nav.ListSiteGroupOptions(c.Context(), navsqlc.ListSiteGroupOptionsParams{Keyword: page.Keyword, RowOffset: int32((page.PageNum - 1) * page.PageSize), RowLimit: int32(page.PageSize)})
	if err != nil {
		return common.NewResponse(c).Error(common.NewDaoError(err.Error()))
	}
	list := make([]adminutil.OptionItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, adminutil.OptionItem{ID: row.ID, Label: row.Name, Extra: row.NameEn})
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, list))
}

func (api *OptionsAPI) GameOptions(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	total, err := api.game.CountGameOptions(c.Context(), page.Keyword)
	if err != nil {
		return common.NewResponse(c).Error(common.NewDaoError(err.Error()))
	}
	rows, err := api.game.ListGameOptions(c.Context(), gamesqlc.ListGameOptionsParams{Keyword: page.Keyword, RowOffset: int32((page.PageNum - 1) * page.PageSize), RowLimit: int32(page.PageSize)})
	if err != nil {
		return common.NewResponse(c).Error(common.NewDaoError(err.Error()))
	}
	list := make([]adminutil.OptionItem, 0, len(rows))
	for _, row := range rows {
		extra := fmt.Sprintf("AppID %d · ID %d", row.Appid, row.ID)
		if strings.TrimSpace(row.NameEn) != "" {
			extra = row.NameEn + " · " + extra
		}
		list = append(list, adminutil.OptionItem{ID: row.ID, Label: row.Name, Extra: extra})
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, list))
}

func (api *OptionsAPI) TagOptions(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	rows, err := api.game.ListTagOptions(c.Context(), page.Keyword)
	if err != nil {
		return common.NewResponse(c).Error(common.NewDaoError(err.Error()))
	}
	list := make([]adminutil.OptionItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, adminutil.OptionItem{ID: row.ID, Label: row.Name, Extra: row.NameEn})
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(int64(len(list)), list))
}
