package controller

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/gameadmin/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	gamesqlc "github.com/gofurry/gofurry-admin/internal/db/game/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	pkgmodels "github.com/gofurry/gofurry-admin/pkg/models"
	"github.com/jackc/pgx/v5/pgtype"
)

var tagCodePattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func validateTagText(code, name, nameEn, info, infoEn string) common.Error {
	if len(code) < 1 || len(code) > 64 || !tagCodePattern.MatchString(code) {
		return common.NewValidationError("code must be 1..64 lowercase kebab-case characters")
	}
	if strings.TrimSpace(name) == "" || strings.TrimSpace(nameEn) == "" {
		return common.NewValidationError("name and name_en are required")
	}
	for _, s := range []string{name, nameEn, info, infoEn} {
		if utf8.RuneCountInString(s) > 255 {
			return common.NewValidationError("tag text must be at most 255 characters")
		}
	}
	return nil
}
func archivedTime(v pgtype.Timestamp) *pkgmodels.LocalTime {
	if !v.Valid {
		return nil
	}
	t := localTime(v)
	return &t
}
func requireActiveCategory(ctx context.Context, q *gamesqlc.Queries, id int64) error {
	c, err := q.GetTagCategory(ctx, id)
	if err != nil {
		return common.NewValidationError("category does not exist")
	}
	if c.ArchivedAt.Valid {
		return common.NewValidationError("category is archived")
	}
	return nil
}

func tagModel(row gamesqlc.GfgTag) models.Tag {
	return models.Tag{ID: row.ID, Code: row.Code, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn, CategoryID: row.CategoryID, ArchivedAt: archivedTime(row.ArchivedAt), CreateTime: localTime(row.CreateTime), UpdateTime: localTime(row.UpdateTime)}
}
func (api *GameAPI) ListTags(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	ctx := c.Context()
	total, err := api.store.q.CountTags(ctx, page.Keyword)
	if err != nil {
		return common.NewResponse(c).Error(gameDAOError(err))
	}
	limit, offset := pageArgsGame(page)
	rows, err := api.store.q.ListTags(ctx, gamesqlc.ListTagsParams{Keyword: page.Keyword, RowLimit: limit, RowOffset: offset})
	if err != nil {
		return common.NewResponse(c).Error(gameDAOError(err))
	}
	items := make([]models.Tag, 0, len(rows))
	for _, row := range rows {
		items = append(items, tagModel(row))
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}
func (api *GameAPI) GetTag(c fiber.Ctx) error {
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return common.NewResponse(c).Error(e)
	}
	row, err := api.store.q.GetTag(c.Context(), id)
	if err != nil {
		return common.NewResponse(c).Error(gameDAOError(err))
	}
	return common.NewResponse(c).SuccessWithData(tagModel(row))
}
func (api *GameAPI) CreateTag(c fiber.Ctx) error {
	var req models.TagPayload
	if e := adminutil.DecodeBody(c, &req); e != nil {
		return common.NewResponse(c).Error(e)
	}
	if e := validateTagText(req.Code, req.Name, req.NameEn, req.Info, req.InfoEn); e != nil {
		return common.NewResponse(c).Error(e)
	}
	var result models.Tag
	ctx := c.Context()
	err := api.store.mutate(ctx, audit.MetaFromFiber(c), "create", "gfg_tag", func(q *gamesqlc.Queries) (int64, any, any, error) {
		if err := requireActiveCategory(ctx, q, req.CategoryID); err != nil {
			return 0, nil, nil, err
		}
		row, err := q.InsertTag(ctx, gamesqlc.InsertTagParams{Code: req.Code, Name: strings.TrimSpace(req.Name), NameEn: strings.TrimSpace(req.NameEn), Info: req.Info, InfoEn: req.InfoEn, CategoryID: req.CategoryID})
		result = tagModel(row)
		return row.ID, nil, result, err
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(result)
}
func (api *GameAPI) UpdateTag(c fiber.Ctx) error {
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return common.NewResponse(c).Error(e)
	}
	var req models.TagPayload
	if e := adminutil.DecodeBody(c, &req); e != nil {
		return common.NewResponse(c).Error(e)
	}
	if e := validateTagText(req.Code, req.Name, req.NameEn, req.Info, req.InfoEn); e != nil {
		return common.NewResponse(c).Error(e)
	}
	ctx := c.Context()
	err := api.store.mutate(ctx, audit.MetaFromFiber(c), "update", "gfg_tag", func(q *gamesqlc.Queries) (int64, any, any, error) {
		before, err := q.GetTag(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		if req.Code != before.Code {
			return id, before, nil, common.NewValidationError("code is immutable after creation")
		}
		if err := requireActiveCategory(ctx, q, req.CategoryID); err != nil {
			return 0, nil, nil, err
		}
		after, err := q.UpdateTag(ctx, gamesqlc.UpdateTagParams{ID: id, Name: strings.TrimSpace(req.Name), NameEn: strings.TrimSpace(req.NameEn), Info: req.Info, InfoEn: req.InfoEn, CategoryID: req.CategoryID})
		if err == nil {
			if before.CategoryID != after.CategoryID {
				err = q.InvalidateGameRecommendations(ctx)
			}
		}
		return id, tagModel(before), tagModel(after), err
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return api.GetTag(c)
}
func (api *GameAPI) ArchiveTag(c fiber.Ctx) error { return api.setTagArchive(c, true) }
func (api *GameAPI) RestoreTag(c fiber.Ctx) error { return api.setTagArchive(c, false) }
func (api *GameAPI) setTagArchive(c fiber.Ctx, archive bool) error {
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return common.NewResponse(c).Error(e)
	}
	ctx := c.Context()
	action := "restore"
	if archive {
		action = "archive"
	}
	err := api.store.mutate(ctx, audit.MetaFromFiber(c), action, "gfg_tag", func(q *gamesqlc.Queries) (int64, any, any, error) {
		before, err := q.GetTag(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		if archive {
			n, e := q.TagAssignmentCount(ctx, id)
			if e != nil {
				return id, before, nil, e
			}
			if n > 0 {
				return id, before, nil, common.NewValidationError("detach tag from every game before archiving")
			}
		} else {
			if e := requireActiveCategory(ctx, q, before.CategoryID); e != nil {
				return id, before, nil, e
			}
		}
		after, err := q.ArchiveTag(ctx, gamesqlc.ArchiveTagParams{ID: id, Archive: archive})
		if err == nil && before.ArchivedAt.Valid != archive {
			err = q.InvalidateGameRecommendations(ctx)
		}
		return id, tagModel(before), tagModel(after), err
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return api.GetTag(c)
}

func tagCategoryModel(row gamesqlc.GfgTagCategory) models.TagCategory {
	return models.TagCategory{ID: row.ID, Code: row.Code, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn, SortOrder: row.SortOrder, ArchivedAt: archivedTime(row.ArchivedAt), CreateTime: localTime(row.CreateTime), UpdateTime: localTime(row.UpdateTime)}
}
func (api *GameAPI) ListTagCategories(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	ctx := c.Context()
	total, err := api.store.q.CountTagCategories(ctx, page.Keyword)
	if err != nil {
		return common.NewResponse(c).Error(gameDAOError(err))
	}
	limit, offset := pageArgsGame(page)
	rows, err := api.store.q.ListTagCategories(ctx, gamesqlc.ListTagCategoriesParams{Keyword: page.Keyword, RowLimit: limit, RowOffset: offset})
	if err != nil {
		return common.NewResponse(c).Error(gameDAOError(err))
	}
	items := make([]models.TagCategory, 0, len(rows))
	for _, row := range rows {
		items = append(items, tagCategoryModel(row))
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}
func (api *GameAPI) GetTagCategory(c fiber.Ctx) error {
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return common.NewResponse(c).Error(e)
	}
	row, err := api.store.q.GetTagCategory(c.Context(), id)
	if err != nil {
		return common.NewResponse(c).Error(gameDAOError(err))
	}
	return common.NewResponse(c).SuccessWithData(tagCategoryModel(row))
}
func (api *GameAPI) CreateTagCategory(c fiber.Ctx) error {
	var req models.TagCategoryPayload
	if e := adminutil.DecodeBody(c, &req); e != nil {
		return common.NewResponse(c).Error(e)
	}
	if e := validateTagText(req.Code, req.Name, req.NameEn, req.Info, req.InfoEn); e != nil {
		return common.NewResponse(c).Error(e)
	}
	var result models.TagCategory
	ctx := c.Context()
	err := api.store.mutate(ctx, audit.MetaFromFiber(c), "create", "gfg_tag_category", func(q *gamesqlc.Queries) (int64, any, any, error) {

		row, err := q.InsertTagCategory(ctx, gamesqlc.InsertTagCategoryParams{Code: req.Code, Name: strings.TrimSpace(req.Name), NameEn: strings.TrimSpace(req.NameEn), Info: req.Info, InfoEn: req.InfoEn, SortOrder: req.SortOrder})
		result = tagCategoryModel(row)
		return row.ID, nil, result, err
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(result)
}
func (api *GameAPI) UpdateTagCategory(c fiber.Ctx) error {
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return common.NewResponse(c).Error(e)
	}
	var req models.TagCategoryPayload
	if e := adminutil.DecodeBody(c, &req); e != nil {
		return common.NewResponse(c).Error(e)
	}
	if e := validateTagText(req.Code, req.Name, req.NameEn, req.Info, req.InfoEn); e != nil {
		return common.NewResponse(c).Error(e)
	}
	ctx := c.Context()
	err := api.store.mutate(ctx, audit.MetaFromFiber(c), "update", "gfg_tag_category", func(q *gamesqlc.Queries) (int64, any, any, error) {
		before, err := q.GetTagCategory(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		if req.Code != before.Code {
			return id, before, nil, common.NewValidationError("code is immutable after creation")
		}

		after, err := q.UpdateTagCategory(ctx, gamesqlc.UpdateTagCategoryParams{ID: id, Name: strings.TrimSpace(req.Name), NameEn: strings.TrimSpace(req.NameEn), Info: req.Info, InfoEn: req.InfoEn, SortOrder: req.SortOrder})
		return id, tagCategoryModel(before), tagCategoryModel(after), err
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return api.GetTagCategory(c)
}
func (api *GameAPI) ArchiveTagCategory(c fiber.Ctx) error { return api.setTagCategoryArchive(c, true) }
func (api *GameAPI) RestoreTagCategory(c fiber.Ctx) error { return api.setTagCategoryArchive(c, false) }
func (api *GameAPI) setTagCategoryArchive(c fiber.Ctx, archive bool) error {
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return common.NewResponse(c).Error(e)
	}
	ctx := c.Context()
	action := "restore"
	if archive {
		action = "archive"
	}
	err := api.store.mutate(ctx, audit.MetaFromFiber(c), action, "gfg_tag_category", func(q *gamesqlc.Queries) (int64, any, any, error) {
		before, err := q.GetTagCategory(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		if archive {
			n, e := q.ActiveCategoryTagCount(ctx, id)
			if e != nil {
				return id, before, nil, e
			}
			if n > 0 {
				return id, before, nil, common.NewValidationError("archive active tags before archiving their category")
			}
		}
		after, err := q.ArchiveTagCategory(ctx, gamesqlc.ArchiveTagCategoryParams{ID: id, Archive: archive})
		if err == nil && before.ArchivedAt.Valid != archive {
			err = q.InvalidateGameRecommendations(ctx)
		}
		return id, tagCategoryModel(before), tagCategoryModel(after), err
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return api.GetTagCategory(c)
}

func classificationRoles(req models.GameClassificationPayload) (map[int64]string, common.Error) {
	roles := make(map[int64]string, len(req.TagIDs)+2)
	for _, id := range req.TagIDs {
		if id <= 0 {
			return nil, common.NewValidationError("tag_ids must be positive")
		}
		roles[id] = "normal"
	}
	if req.PrimaryTagID != nil {
		if *req.PrimaryTagID <= 0 {
			return nil, common.NewValidationError("primary_tag_id must be positive or null")
		}
		roles[*req.PrimaryTagID] = "primary"
	}
	if req.SecondaryTagID != nil {
		if *req.SecondaryTagID <= 0 {
			return nil, common.NewValidationError("secondary_tag_id must be positive or null")
		}
		if roles[*req.SecondaryTagID] == "primary" {
			return nil, common.NewValidationError("primary and secondary must differ")
		}
		roles[*req.SecondaryTagID] = "secondary"
	}
	return roles, nil
}
func (api *GameAPI) SaveClassification(c fiber.Ctx) error {
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return common.NewResponse(c).Error(e)
	}
	var req models.GameClassificationPayload
	if e := adminutil.DecodeBody(c, &req); e != nil {
		return common.NewResponse(c).Error(e)
	}
	roles, e := classificationRoles(req)
	if e != nil {
		return common.NewResponse(c).Error(e)
	}
	ctx := c.Context()
	err := api.store.mutate(ctx, audit.MetaFromFiber(c), "classification", "gfg_game_tag", func(q *gamesqlc.Queries) (int64, any, any, error) {
		game, err := q.LockGameForUpdate(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		before, err := q.ListGameTagRelations(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		ids := make([]int64, 0, len(roles))
		var primary, secondary int64
		for tagID, role := range roles {
			tag, e := q.GetTag(ctx, tagID)
			if e != nil || tag.ArchivedAt.Valid {
				return id, before, nil, common.NewValidationError("tag does not exist or is archived")
			}
			if e := requireActiveCategory(ctx, q, tag.CategoryID); e != nil {
				return id, before, nil, e
			}
			ids = append(ids, tagID)
			if role == "primary" {
				primary = tagID
			}
			if role == "secondary" {
				secondary = tagID
			}
		}
		changed := len(before) != len(roles)
		for _, row := range before {
			if roles[row.TagID] != row.Role {
				changed = true
			}
		}
		if changed {
			if err = q.DeleteRemovedGameTags(ctx, gamesqlc.DeleteRemovedGameTagsParams{GameID: id, TagIds: ids}); err != nil {
				return id, before, nil, err
			}
			if err = q.ClearChangedGameTagRoles(ctx, gamesqlc.ClearChangedGameTagRolesParams{GameID: id, PrimaryID: primary, SecondaryID: secondary}); err != nil {
				return id, before, nil, err
			}
			for tagID, role := range roles {
				if err = q.UpsertGameTag(ctx, gamesqlc.UpsertGameTagParams{GameID: id, TagID: tagID, Role: role}); err != nil {
					return id, before, nil, err
				}
			}
			if err = q.InvalidateGameRecommendations(ctx); err != nil {
				return id, before, nil, err
			}
		}
		if changed || game.Weight != req.Weight {
			if err = q.SetGameClassificationWeight(ctx, gamesqlc.SetGameClassificationWeightParams{ID: id, Weight: req.Weight}); err != nil {
				return id, before, nil, err
			}
			if err = q.RefreshCurrentGameDaily(ctx, gamesqlc.RefreshCurrentGameDailyParams{GameID: id, MaterializationSource: "observed"}); err != nil {
				return id, before, nil, err
			}
		}
		after, err := q.ListGameTagRelations(ctx, id)
		return id, struct {
			Weight int64
			Tags   any
		}{game.Weight, before}, struct {
			Weight int64
			Tags   any
		}{req.Weight, after}, err
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return api.GetGameWorkspace(c)
}
