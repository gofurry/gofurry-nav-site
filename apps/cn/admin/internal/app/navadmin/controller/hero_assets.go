package controller

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	navsqlc "github.com/gofurry/gofurry-admin/internal/db/nav/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

func (api *navAPI) ListHeroAssets(c fiber.Ctx) error {
	variant := c.Query("variant", "desktop")
	if variant != "desktop" && variant != "mobile" {
		return common.NewResponse(c).Error(common.NewValidationError("variant must be desktop or mobile"))
	}
	total, err := api.store.q.CountHeroAssets(c.Context(), variant)
	if err != nil {
		return common.NewResponse(c).Error(navDAOError(err))
	}
	limit, offset := pageArgs(adminutil.ParsePageQuery(c))
	rows, err := api.store.q.ListHeroAssets(c.Context(), navsqlc.ListHeroAssetsParams{Variant: variant, RowLimit: limit, RowOffset: offset})
	if err != nil {
		return common.NewResponse(c).Error(navDAOError(err))
	}
	items := make([]heroDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, api.hero(row))
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}

func (api *navAPI) GetHeroAsset(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	row, e := api.store.q.GetHeroAsset(c.Context(), id)
	if e != nil {
		return common.NewResponse(c).Error(navDAOError(e))
	}
	return common.NewResponse(c).SuccessWithData(api.hero(row))
}

func (api *navAPI) CreateHeroAsset(c fiber.Ctx) error {
	variant, name := c.FormValue("variant"), strings.TrimSpace(c.FormValue("name"))
	if variant != "desktop" && variant != "mobile" {
		return common.NewResponse(c).Error(common.NewValidationError("variant must be desktop or mobile"))
	}
	if !validateAssetName(name) {
		return common.NewResponse(c).Error(common.NewValidationError("name is required (at most 120 characters)"))
	}
	enabled, e := strconv.ParseBool(c.FormValue("enabled", "true"))
	if e != nil {
		return common.NewResponse(c).Error(common.NewValidationError("enabled must be a boolean"))
	}
	object, err := uploadObject(c, "hero-"+variant, 0)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	publication, err := api.publish(c, object)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var result heroDTO
	err = api.store.mutate(c.Context(), audit.MetaFromFiber(c), "hero.create", "gfn_home_hero_asset", func(q *navsqlc.Queries) (int64, any, any, error) {
		id, e := q.NextHeroAssetID(c.Context())
		if e != nil {
			return 0, nil, nil, e
		}
		row, e := q.CreateHeroAsset(c.Context(), navsqlc.CreateHeroAssetParams{ID: id, Variant: variant, Name: name, ObjectKey: object.Key, Enabled: enabled})
		result = api.hero(row)
		return id, nil, row, e
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(publicationDTO{publication, result})
}

func (api *navAPI) UpdateHeroAsset(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req heroPayload
	if err := decodeAssetJSON(c.Body(), &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	req.Name = strings.TrimSpace(req.Name)
	if !validateAssetName(req.Name) {
		return common.NewResponse(c).Error(common.NewValidationError("name is required (at most 120 characters)"))
	}
	// Select the audit verb inside the same locked transaction as the update.
	err = api.store.mutateAsset(c.Context(), audit.MetaFromFiber(c), "hero.update", "gfn_home_hero_asset", func(q *navsqlc.Queries) (int64, any, any, string, error) {
		before, e := q.LockHeroAsset(c.Context(), id)
		if e != nil {
			return id, nil, nil, "", e
		}
		_, e = q.UpdateHeroAsset(c.Context(), navsqlc.UpdateHeroAssetParams{ID: id, Name: req.Name, Enabled: req.Enabled})
		return id, before, req, assetUpdateAction("hero", before.Enabled, req.Enabled), e
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return api.GetHeroAsset(c)
}

func (api *navAPI) ReplaceHeroAssetFile(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	row, e := api.store.q.GetHeroAsset(c.Context(), id)
	if e != nil {
		return common.NewResponse(c).Error(navDAOError(e))
	}
	object, err := uploadObject(c, "hero-"+row.Variant, 0)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	publication, err := api.publish(c, object)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	err = api.store.mutate(c.Context(), audit.MetaFromFiber(c), "hero.file.replace", "gfn_home_hero_asset", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, e := q.LockHeroAsset(c.Context(), id)
		if e != nil {
			return id, nil, nil, e
		}
		_, e = q.ReplaceHeroAssetFile(c.Context(), navsqlc.ReplaceHeroAssetFileParams{ID: id, ObjectKey: object.Key})
		return id, before, map[string]any{"object_key": object.Key}, e
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(publication)
}

func (api *navAPI) DeleteHeroAsset(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	err = api.store.mutate(c.Context(), audit.MetaFromFiber(c), "hero.delete", "gfn_home_hero_asset", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, e := q.LockHeroAsset(c.Context(), id)
		if e != nil {
			return id, nil, nil, e
		}
		_, e = q.DeleteHeroAsset(c.Context(), id)
		return id, before, map[string]any{"deleted": true}, e
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(map[string]bool{"deleted": true})
}
