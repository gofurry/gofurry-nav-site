package controller

import (
	gamemodels "github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/gameadmin/models"
	navmodels "github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/navadmin/models"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/shared/adminutil"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/infra/db"
	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/common"
	"github.com/gofiber/fiber/v3"
)

type optionsAPI struct{}

var OptionsAPI = &optionsAPI{}

func (api *optionsAPI) SiteOptions(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	base := adminutil.ApplyKeyword(db.Databases.DB(db.Nav).Model(&navmodels.Site{}).Where("deleted IS NOT TRUE").Order("id DESC"), page.Keyword, "name", "name_en", "CAST(id AS TEXT)")
	var rows []navmodels.Site
	total, err := adminutil.Paginate(base, page, &rows)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	list := make([]adminutil.OptionItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, adminutil.OptionItem{ID: row.ID, Label: row.Name, Extra: row.NameEn})
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, list))
}

func (api *optionsAPI) SiteGroupOptions(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	base := adminutil.ApplyKeyword(db.Databases.DB(db.Nav).Model(&navmodels.SiteGroup{}).Order("priority DESC, id DESC"), page.Keyword, "name", "name_en", "CAST(id AS TEXT)")
	var rows []navmodels.SiteGroup
	total, err := adminutil.Paginate(base, page, &rows)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	list := make([]adminutil.OptionItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, adminutil.OptionItem{ID: row.ID, Label: row.Name, Extra: row.NameEn})
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, list))
}

func (api *optionsAPI) GameOptions(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	base := adminutil.ApplyKeyword(db.Databases.DB(db.Game).Model(&gamemodels.Game{}).Order("id DESC"), page.Keyword, "name", "name_en", "CAST(id AS TEXT)")
	var rows []gamemodels.Game
	total, err := adminutil.Paginate(base, page, &rows)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	list := make([]adminutil.OptionItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, adminutil.OptionItem{ID: row.ID, Label: row.Name, Extra: row.NameEn})
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, list))
}

func (api *optionsAPI) TagOptions(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	base := adminutil.ApplyKeyword(db.Databases.DB(db.Game).Model(&gamemodels.Tag{}).Order("id DESC"), page.Keyword, "name", "name_en", "CAST(id AS TEXT)")
	var rows []gamemodels.Tag
	total, err := adminutil.Paginate(base, page, &rows)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	list := make([]adminutil.OptionItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, adminutil.OptionItem{ID: row.ID, Label: row.Name, Extra: row.NameEn})
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, list))
}
