package controller

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	navsqlc "github.com/gofurry/gofurry-admin/internal/db/nav/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

func (api *navAPI) ListBackgroundPatterns(c fiber.Ctx) error {
	total, err := api.store.q.CountBackgroundPatterns(c.Context())
	if err != nil {
		return common.NewResponse(c).Error(navDAOError(err))
	}
	limit, offset := pageArgs(adminutil.ParsePageQuery(c))
	rows, err := api.store.q.ListBackgroundPatterns(c.Context(), navsqlc.ListBackgroundPatternsParams{RowLimit: limit, RowOffset: offset})
	if err != nil {
		return common.NewResponse(c).Error(navDAOError(err))
	}
	items := make([]patternDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, api.pattern(navsqlc.GetBackgroundPatternRow(row)))
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}

func (api *navAPI) GetBackgroundPattern(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	row, e := api.store.q.GetBackgroundPattern(c.Context(), id)
	if e != nil {
		return common.NewResponse(c).Error(navDAOError(e))
	}
	return common.NewResponse(c).SuccessWithData(api.pattern(row))
}

func patternForm(c fiber.Ctx) (patternPayload, common.Error) {
	// Multipart fields have the same names and validation as JSON metadata updates.
	values := map[string]any{}
	for _, key := range []string{"name", "name_en", "light_color", "dark_color"} {
		values[key] = strings.TrimSpace(c.FormValue(key))
	}
	for _, key := range []string{"light_opacity", "dark_opacity"} {
		value, err := strconv.ParseFloat(c.FormValue(key), 64)
		if err != nil {
			return patternPayload{}, common.NewValidationError("invalid " + key)
		}
		values[key] = value
	}
	for _, key := range []string{"default_size_px", "sort_order"} {
		value, err := strconv.ParseInt(c.FormValue(key, "0"), 10, 64)
		if err != nil {
			return patternPayload{}, common.NewValidationError("invalid " + key)
		}
		values[key] = value
	}
	enabled, err := strconv.ParseBool(c.FormValue("enabled", "true"))
	if err != nil {
		return patternPayload{}, common.NewValidationError("invalid enabled")
	}
	values["enabled"] = enabled
	data, err := json.Marshal(values)
	if err != nil {
		return patternPayload{}, common.NewValidationError("invalid pattern metadata")
	}
	var req patternPayload
	if err := decodeAssetJSON(data, &req); err != nil {
		return req, err
	}
	return req, validatePattern(req)
}

func (api *navAPI) CreateBackgroundPattern(c fiber.Ctx) error {
	req, err := patternForm(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	object, err := uploadObject(c, "pattern", 0)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	publication, err := api.publish(c, object)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var result patternDTO
	err = api.store.mutate(c.Context(), audit.MetaFromFiber(c), "pattern.create", "gfn_background_pattern", func(q *navsqlc.Queries) (int64, any, any, error) {
		id, e := q.NextBackgroundPatternID(c.Context())
		if e != nil {
			return 0, nil, nil, e
		}
		row, e := q.CreateBackgroundPattern(c.Context(), navsqlc.CreateBackgroundPatternParams{ID: id, Name: req.Name, NameEn: req.NameEn, ObjectKey: object.Key, LightColor: req.LightColor, DarkColor: req.DarkColor, LightOpacity: req.LightOpacity, DarkOpacity: req.DarkOpacity, DefaultSizePx: req.DefaultSizePx, Enabled: req.Enabled, SortOrder: req.SortOrder})
		result = api.pattern(navsqlc.GetBackgroundPatternRow(row))
		return id, nil, row, e
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(publicationDTO{publication, result})
}

func (api *navAPI) UpdateBackgroundPattern(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req patternPayload
	if err := decodeAssetJSON(c.Body(), &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	req.Name, req.NameEn = strings.TrimSpace(req.Name), strings.TrimSpace(req.NameEn)
	if err := validatePattern(req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	err = api.store.mutateAsset(c.Context(), audit.MetaFromFiber(c), "pattern.update", "gfn_background_pattern", func(q *navsqlc.Queries) (int64, any, any, string, error) {
		before, e := q.LockBackgroundPattern(c.Context(), id)
		if e != nil {
			return id, nil, nil, "", e
		}
		_, e = q.UpdateBackgroundPattern(c.Context(), navsqlc.UpdateBackgroundPatternParams{ID: id, Name: req.Name, NameEn: req.NameEn, LightColor: req.LightColor, DarkColor: req.DarkColor, LightOpacity: req.LightOpacity, DarkOpacity: req.DarkOpacity, DefaultSizePx: req.DefaultSizePx, Enabled: req.Enabled, SortOrder: req.SortOrder})
		return id, before, req, assetUpdateAction("pattern", before.Enabled, req.Enabled), e
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return api.GetBackgroundPattern(c)
}

func (api *navAPI) ReplaceBackgroundPatternFile(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if _, e := api.store.q.GetBackgroundPattern(c.Context(), id); e != nil {
		return common.NewResponse(c).Error(navDAOError(e))
	}
	object, err := uploadObject(c, "pattern", 0)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	publication, err := api.publish(c, object)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	err = api.store.mutate(c.Context(), audit.MetaFromFiber(c), "pattern.file.replace", "gfn_background_pattern", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, e := q.LockBackgroundPattern(c.Context(), id)
		if e != nil {
			return id, nil, nil, e
		}
		_, e = q.ReplaceBackgroundPatternFile(c.Context(), navsqlc.ReplaceBackgroundPatternFileParams{ID: id, ObjectKey: object.Key})
		return id, before, map[string]any{"object_key": object.Key}, e
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(publication)
}

func (api *navAPI) DeleteBackgroundPattern(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	err = api.store.mutate(c.Context(), audit.MetaFromFiber(c), "pattern.delete", "gfn_background_pattern", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, e := q.LockBackgroundPattern(c.Context(), id)
		if e != nil {
			return id, nil, nil, e
		}
		_, e = q.DeleteBackgroundPattern(c.Context(), id)
		return id, before, map[string]any{"deleted": true}, e
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(map[string]bool{"deleted": true})
}

func assetUpdateAction(kind string, before, after bool) string {
	if before == after {
		return kind + ".update"
	}
	if after {
		return kind + ".enable"
	}
	return kind + ".disable"
}

func (store *navStore) mutateAsset(ctx context.Context, meta audit.Meta, action, resource string, change func(*navsqlc.Queries) (int64, any, any, string, error)) common.Error {
	tx, err := store.pool.Begin(ctx)
	if err != nil {
		return navDAOError(err)
	}
	defer tx.Rollback(ctx)
	id, before, after, verb, err := change(store.q.WithTx(tx))
	if err != nil {
		return navDAOError(err)
	}
	if verb != "" {
		action = verb
	}
	if err := store.audit.Log(ctx, meta, action, resource, id, before, after); err != nil {
		return err
	}
	return navDAOError(tx.Commit(ctx))
}
