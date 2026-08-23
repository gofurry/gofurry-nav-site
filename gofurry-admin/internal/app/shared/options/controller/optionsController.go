package controller

import (
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
		list = append(list, adminutil.OptionItem{ID: row.ID, Label: row.Name, Extra: row.NameEn})
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
