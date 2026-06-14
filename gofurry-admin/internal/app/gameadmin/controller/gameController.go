package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/gameadmin/models"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/shared/adminutil"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/shared/audit"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/infra/db"
	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/common"
	pkgmodels "github.com/gofurry/awesome-fiber-template/v3/medium/pkg/models"
	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/util"
	"gorm.io/gorm"
)

type gameAPI struct{}

var GameAPI = &gameAPI{}

func gameDB() *gorm.DB {
	return db.Databases.DB(db.Game)
}

func (api *gameAPI) ListGames(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	base := adminutil.ApplyKeyword(gameDB().Model(&models.Game{}).Order("id DESC"), page.Keyword, "name", "name_en", "info", "info_en", "CAST(id AS TEXT)")
	var rows []models.Game
	total, err := adminutil.Paginate(base, page, &rows)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	list := make([]models.GameDTO, 0, len(rows))
	for _, row := range rows {
		list = append(list, gameDTO(row))
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, list))
}

func (api *gameAPI) CreateGame(c fiber.Ctx) error {
	var req models.GamePayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if err := validateGamePayload(req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	var created models.Game
	err := gameDB().Transaction(func(tx *gorm.DB) error {
		if dupErr := ensureUniqueGameAppID(tx, req.Appid, 0); dupErr != nil {
			return dupErr
		}
		ids, allocErr := adminutil.AllocateSequentialIDs(tx, created.TableName(), 1)
		if allocErr != nil {
			return allocErr
		}
		created = models.Game{
			ID:           ids[0],
			Name:         strings.TrimSpace(req.Name),
			NameEn:       strings.TrimSpace(req.NameEn),
			Info:         strings.TrimSpace(req.Info),
			InfoEn:       strings.TrimSpace(req.InfoEn),
			Resources:    adminutil.ToJSONStringPtr(normalizeKV(req.Resources)),
			Groups:       adminutil.ToJSONStringPtr(normalizeKV(req.Groups)),
			ReleaseDate:  strings.TrimSpace(req.ReleaseDate),
			Developers:   adminutil.MustJSON(normalizeStringArray(req.Developers)),
			Publishers:   adminutil.MustJSON(normalizeStringArray(req.Publishers)),
			Appid:        req.Appid,
			Header:       strings.TrimSpace(req.Header),
			Links:        adminutil.ToJSONStringPtr(normalizeKV(req.Links)),
			Weight:       req.Weight,
			PrimaryTag:   req.PrimaryTag,
			SecondaryTag: req.SecondaryTag,
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		after, snapErr := audit.SnapshotByID(tx, created.TableName(), created.ID)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "create", created.TableName(), created.ID, nil, after)
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(gameDTO(created))
}

func (api *gameAPI) GetGame(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var row models.Game
	if err := gameDB().First(&row, "id = ?", id).Error; err != nil {
		return common.NewResponse(c).Error(common.NewDaoError(err.Error()))
	}
	return common.NewResponse(c).SuccessWithData(gameDTO(row))
}

func (api *gameAPI) UpdateGame(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req models.GamePayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if err := validateGamePayload(req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	txErr := gameDB().Transaction(func(tx *gorm.DB) error {
		if dupErr := ensureUniqueGameAppID(tx, req.Appid, id); dupErr != nil {
			return dupErr
		}
		before, snapErr := audit.SnapshotByID(tx, (&models.Game{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		if err := tx.Model(&models.Game{}).Where("id = ?", id).Updates(map[string]any{
			"name":          strings.TrimSpace(req.Name),
			"name_en":       strings.TrimSpace(req.NameEn),
			"info":          strings.TrimSpace(req.Info),
			"info_en":       strings.TrimSpace(req.InfoEn),
			"resources":     adminutil.MustJSON(normalizeKV(req.Resources)),
			"groups":        adminutil.MustJSON(normalizeKV(req.Groups)),
			"release_date":  strings.TrimSpace(req.ReleaseDate),
			"developers":    adminutil.MustJSON(normalizeStringArray(req.Developers)),
			"publishers":    adminutil.MustJSON(normalizeStringArray(req.Publishers)),
			"appid":         req.Appid,
			"header":        strings.TrimSpace(req.Header),
			"links":         adminutil.MustJSON(normalizeKV(req.Links)),
			"weight":        req.Weight,
			"primary_tag":   req.PrimaryTag,
			"secondary_tag": req.SecondaryTag,
		}).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		after, snapErr := audit.SnapshotByID(tx, (&models.Game{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "update", (&models.Game{}).TableName(), id, before, after)
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return api.GetGame(c)
}

func (api *gameAPI) DeleteGame(c fiber.Ctx) error {
	return api.deleteHard(c, &models.Game{})
}

func (api *gameAPI) ListComments(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	base := adminutil.ApplyKeyword(gameDB().Model(&models.GameComment{}).Order("id DESC"), page.Keyword, "content", "region", "name", "ip", "CAST(id AS TEXT)", "CAST(game_id AS TEXT)")
	var rows []models.GameComment
	total, err := adminutil.Paginate(base, page, &rows)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, rows))
}

func (api *gameAPI) CreateComment(c fiber.Ctx) error {
	var req models.GameCommentPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if strings.TrimSpace(req.Content) == "" || req.GameID <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("content and game_id are required"))
	}
	row := models.GameComment{
		ID:      util.GenerateId(),
		Region:  strings.TrimSpace(req.Region),
		Content: strings.TrimSpace(req.Content),
		Score:   req.Score,
		GameID:  req.GameID,
		IP:      strings.TrimSpace(req.IP),
		Name:    strings.TrimSpace(req.Name),
	}
	txErr := gameDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		after, snapErr := audit.SnapshotByID(tx, row.TableName(), row.ID)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "create", row.TableName(), row.ID, nil, after)
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return common.NewResponse(c).SuccessWithData(row)
}

func (api *gameAPI) GetComment(c fiber.Ctx) error {
	return api.getOne(c, gameDB(), &models.GameComment{})
}

func (api *gameAPI) UpdateComment(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req models.GameCommentPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if strings.TrimSpace(req.Content) == "" || req.GameID <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("content and game_id are required"))
	}
	txErr := gameDB().Transaction(func(tx *gorm.DB) error {
		before, snapErr := audit.SnapshotByID(tx, (&models.GameComment{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		if err := tx.Model(&models.GameComment{}).Where("id = ?", id).Updates(map[string]any{
			"region":  strings.TrimSpace(req.Region),
			"content": strings.TrimSpace(req.Content),
			"score":   req.Score,
			"game_id": req.GameID,
			"ip":      strings.TrimSpace(req.IP),
			"name":    strings.TrimSpace(req.Name),
		}).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		after, snapErr := audit.SnapshotByID(tx, (&models.GameComment{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "update", (&models.GameComment{}).TableName(), id, before, after)
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return api.GetComment(c)
}

func (api *gameAPI) DeleteComment(c fiber.Ctx) error {
	return api.deleteHard(c, &models.GameComment{})
}

func (api *gameAPI) ListPrizes(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	base := adminutil.ApplyKeyword(gameDB().Model(&models.Prize{}).Order("id DESC"), page.Keyword, "title", "desc", "CAST(id AS TEXT)")
	var rows []models.Prize
	total, err := adminutil.Paginate(base, page, &rows)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	list := make([]models.PrizeDTO, 0, len(rows))
	for _, row := range rows {
		list = append(list, prizeDTO(row))
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, list))
}

func (api *gameAPI) CreatePrize(c fiber.Ctx) error {
	var req models.PrizePayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	startTime, endTime, valErr := parsePrizeTimes(req)
	if valErr != nil {
		return common.NewResponse(c).Error(valErr)
	}
	var created models.Prize
	err := gameDB().Transaction(func(tx *gorm.DB) error {
		ids, allocErr := adminutil.AllocateSequentialIDs(tx, created.TableName(), 1)
		if allocErr != nil {
			return allocErr
		}
		created = models.Prize{
			ID:        ids[0],
			Title:     strings.TrimSpace(req.Title),
			Desc:      strings.TrimSpace(req.Desc),
			Prize:     adminutil.MustJSON(normalizePrizeBody(req.Prize)),
			Key:       strings.TrimSpace(req.Key),
			StartTime: pkgmodels.LocalTime(startTime),
			EndTime:   pkgmodels.LocalTime(endTime),
			Status:    req.Status,
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		after, snapErr := audit.SnapshotByID(tx, created.TableName(), created.ID)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "create", created.TableName(), created.ID, nil, after)
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(prizeDTO(created))
}

func (api *gameAPI) GetPrize(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var row models.Prize
	if err := gameDB().First(&row, "id = ?", id).Error; err != nil {
		return common.NewResponse(c).Error(common.NewDaoError(err.Error()))
	}
	return common.NewResponse(c).SuccessWithData(prizeDTO(row))
}

func (api *gameAPI) UpdatePrize(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req models.PrizePayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	startTime, endTime, valErr := parsePrizeTimes(req)
	if valErr != nil {
		return common.NewResponse(c).Error(valErr)
	}
	txErr := gameDB().Transaction(func(tx *gorm.DB) error {
		before, snapErr := audit.SnapshotByID(tx, (&models.Prize{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		if err := tx.Model(&models.Prize{}).Where("id = ?", id).Updates(map[string]any{
			"title":      strings.TrimSpace(req.Title),
			"desc":       strings.TrimSpace(req.Desc),
			"prize":      adminutil.MustJSON(normalizePrizeBody(req.Prize)),
			"key":        strings.TrimSpace(req.Key),
			"start_time": startTime,
			"end_time":   endTime,
			"status":     req.Status,
		}).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		after, snapErr := audit.SnapshotByID(tx, (&models.Prize{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "update", (&models.Prize{}).TableName(), id, before, after)
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return api.GetPrize(c)
}

func (api *gameAPI) DeletePrize(c fiber.Ctx) error {
	return api.deleteHard(c, &models.Prize{})
}

func (api *gameAPI) ListTags(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	base := adminutil.ApplyKeyword(gameDB().Model(&models.Tag{}).Order("id DESC"), page.Keyword, "name", "name_en", "info", "info_en", "CAST(id AS TEXT)")
	var rows []models.Tag
	total, err := adminutil.Paginate(base, page, &rows)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, rows))
}

func (api *gameAPI) CreateTag(c fiber.Ctx) error {
	var req models.TagPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if req.ID <= 0 || strings.TrimSpace(req.Name) == "" {
		return common.NewResponse(c).Error(common.NewValidationError("id and name are required"))
	}
	row := models.Tag{
		ID:     req.ID,
		Name:   strings.TrimSpace(req.Name),
		NameEn: strings.TrimSpace(req.NameEn),
		Info:   strings.TrimSpace(req.Info),
		InfoEn: strings.TrimSpace(req.InfoEn),
		Prefix: req.Prefix,
	}
	txErr := gameDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		after, snapErr := audit.SnapshotByID(tx, row.TableName(), row.ID)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "create", row.TableName(), row.ID, nil, after)
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return common.NewResponse(c).SuccessWithData(row)
}

func (api *gameAPI) GetTag(c fiber.Ctx) error {
	return api.getOne(c, gameDB(), &models.Tag{})
}

func (api *gameAPI) UpdateTag(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req models.TagPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if strings.TrimSpace(req.Name) == "" {
		return common.NewResponse(c).Error(common.NewValidationError("name is required"))
	}
	txErr := gameDB().Transaction(func(tx *gorm.DB) error {
		before, snapErr := audit.SnapshotByID(tx, (&models.Tag{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		if err := tx.Model(&models.Tag{}).Where("id = ?", id).Updates(map[string]any{
			"name":    strings.TrimSpace(req.Name),
			"name_en": strings.TrimSpace(req.NameEn),
			"info":    strings.TrimSpace(req.Info),
			"info_en": strings.TrimSpace(req.InfoEn),
			"prefix":  req.Prefix,
		}).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		after, snapErr := audit.SnapshotByID(tx, (&models.Tag{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "update", (&models.Tag{}).TableName(), id, before, after)
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return api.GetTag(c)
}

func (api *gameAPI) DeleteTag(c fiber.Ctx) error {
	return api.deleteHard(c, &models.Tag{})
}

func (api *gameAPI) ListTagMaps(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	base := gameDB().Table("gfg_tag_map AS m").
		Select("m.id, m.game_id, m.tag_id, m.create_time, m.update_time, g.name AS game_name, t.name AS tag_name").
		Joins("LEFT JOIN gfg_game g ON g.id = m.game_id").
		Joins("LEFT JOIN gfg_tag t ON t.id = m.tag_id").
		Order("m.id DESC")
	base = adminutil.ApplyKeyword(base, page.Keyword, "CAST(m.id AS TEXT)", "CAST(m.game_id AS TEXT)", "CAST(m.tag_id AS TEXT)", "g.name", "t.name")
	var rows []models.TagMapDTO
	total, err := adminutil.Paginate(base, page, &rows)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, rows))
}

func (api *gameAPI) CreateTagMap(c fiber.Ctx) error {
	var req models.TagMapPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if req.GameID <= 0 || req.TagID <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("game_id and tag_id are required"))
	}
	var created models.TagMap
	err := gameDB().Transaction(func(tx *gorm.DB) error {
		ids, allocErr := adminutil.AllocateSequentialIDs(tx, created.TableName(), 1)
		if allocErr != nil {
			return allocErr
		}
		created = models.TagMap{ID: ids[0], GameID: req.GameID, TagID: req.TagID}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		after, snapErr := audit.SnapshotByID(tx, created.TableName(), created.ID)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "create", created.TableName(), created.ID, nil, after)
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(created)
}

func (api *gameAPI) GetTagMap(c fiber.Ctx) error {
	return api.getOne(c, gameDB(), &models.TagMap{})
}

func (api *gameAPI) UpdateTagMap(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req models.TagMapPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if req.GameID <= 0 || req.TagID <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("game_id and tag_id are required"))
	}
	txErr := gameDB().Transaction(func(tx *gorm.DB) error {
		before, snapErr := audit.SnapshotByID(tx, (&models.TagMap{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		if err := tx.Model(&models.TagMap{}).Where("id = ?", id).Updates(map[string]any{
			"game_id": req.GameID,
			"tag_id":  req.TagID,
		}).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		after, snapErr := audit.SnapshotByID(tx, (&models.TagMap{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "update", (&models.TagMap{}).TableName(), id, before, after)
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return api.GetTagMap(c)
}

func (api *gameAPI) DeleteTagMap(c fiber.Ctx) error {
	return api.deleteHard(c, &models.TagMap{})
}

func (api *gameAPI) BulkReplaceTagMaps(c fiber.Ctx) error {
	var req adminutil.BulkReplaceRequest
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if req.OwnerID <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("owner_id is required"))
	}
	req.IDs = uniqueInt64s(req.IDs)
	err := gameDB().Transaction(func(tx *gorm.DB) error {
		before, snapErr := audit.SnapshotRows(tx, (&models.TagMap{}).TableName(), "game_id = ?", "id ASC", req.OwnerID)
		if snapErr != nil {
			return snapErr
		}
		if err := tx.Where("game_id = ?", req.OwnerID).Delete(&models.TagMap{}).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		if len(req.IDs) == 0 {
			return api.auditTx(c, tx, "bulk_replace", (&models.TagMap{}).TableName(), req.OwnerID, before, []map[string]any{})
		}
		ids, allocErr := adminutil.AllocateSequentialIDs(tx, (&models.TagMap{}).TableName(), len(req.IDs))
		if allocErr != nil {
			return allocErr
		}
		rows := make([]models.TagMap, 0, len(req.IDs))
		for idx, tagID := range req.IDs {
			rows = append(rows, models.TagMap{ID: ids[idx], GameID: req.OwnerID, TagID: tagID})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
		after, snapErr := audit.SnapshotRows(tx, (&models.TagMap{}).TableName(), "game_id = ?", "id ASC", req.OwnerID)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "bulk_replace", (&models.TagMap{}).TableName(), req.OwnerID, before, after)
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).Success()
}

func (api *gameAPI) deleteHard(c fiber.Ctx, model any) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	table := tableNameOf(model)
	txErr := gameDB().Transaction(func(tx *gorm.DB) error {
		before, snapErr := audit.SnapshotByID(tx, table, id)
		if snapErr != nil {
			return snapErr
		}
		if err := tx.Delete(model, "id = ?", id).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		return api.auditTx(c, tx, "delete", table, id, before, nil)
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return common.NewResponse(c).Success()
}

func (api *gameAPI) deleteSoft(c fiber.Ctx, model any) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	table := tableNameOf(model)
	txErr := gameDB().Transaction(func(tx *gorm.DB) error {
		before, snapErr := audit.SnapshotByID(tx, table, id)
		if snapErr != nil {
			return snapErr
		}
		if err := tx.Model(model).Where("id = ?", id).Update("deleted", true).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		after, snapErr := audit.SnapshotByID(tx, table, id)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "delete", table, id, before, after)
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return common.NewResponse(c).Success()
}

func (api *gameAPI) getOne(c fiber.Ctx, base *gorm.DB, out any) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if err := base.First(out, "id = ?", id).Error; err != nil {
		return common.NewResponse(c).Error(common.NewDaoError(err.Error()))
	}
	return common.NewResponse(c).SuccessWithData(out)
}

func validateGamePayload(req models.GamePayload) common.Error {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.NameEn) == "" {
		return common.NewValidationError("name and name_en are required")
	}
	return nil
}

func ensureUniqueGameAppID(tx *gorm.DB, appid, excludeID int64) common.Error {
	if appid <= 0 {
		return nil
	}

	var existing models.Game
	query := tx.Select("id", "name", "appid").Where("appid = ?", appid)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	if err := query.Take(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return common.NewDaoError(err.Error())
	}

	return common.NewValidationError(fmt.Sprintf("appid already exists (game id=%d, name=%s)", existing.ID, existing.Name))
}

func gameDTO(row models.Game) models.GameDTO {
	return models.GameDTO{
		ID:           row.ID,
		Name:         row.Name,
		NameEn:       row.NameEn,
		Info:         row.Info,
		InfoEn:       row.InfoEn,
		CreateTime:   row.CreateTime,
		UpdateTime:   row.UpdateTime,
		Resources:    adminutil.ParseKVArray(row.Resources),
		Groups:       adminutil.ParseKVArray(row.Groups),
		ReleaseDate:  row.ReleaseDate,
		Developers:   adminutil.ParseStringArray(row.Developers),
		Publishers:   adminutil.ParseStringArray(row.Publishers),
		Appid:        row.Appid,
		Header:       row.Header,
		Links:        adminutil.ParseKVArray(row.Links),
		Weight:       row.Weight,
		PrimaryTag:   row.PrimaryTag,
		SecondaryTag: row.SecondaryTag,
	}
}

func prizeDTO(row models.Prize) models.PrizeDTO {
	var prize models.PrizeBody
	_ = json.Unmarshal([]byte(strings.TrimSpace(row.Prize)), &prize)
	return models.PrizeDTO{
		ID:         row.ID,
		Title:      row.Title,
		Desc:       row.Desc,
		Prize:      prize,
		Key:        row.Key,
		StartTime:  row.StartTime,
		EndTime:    row.EndTime,
		CreateTime: row.CreateTime,
		Status:     row.Status,
	}
}

func normalizeStringArray(items []string) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func normalizeKV(items []pkgmodels.KvModel) []pkgmodels.KvModel {
	result := make([]pkgmodels.KvModel, 0, len(items))
	for _, item := range items {
		key := strings.TrimSpace(item.Key)
		value := strings.TrimSpace(item.Value)
		if key == "" && value == "" {
			continue
		}
		result = append(result, pkgmodels.KvModel{Key: key, Value: value})
	}
	return result
}

func normalizePrizeBody(body models.PrizeBody) models.PrizeBody {
	body.Keys = normalizeStringArray(body.Keys)
	body.Title = strings.TrimSpace(body.Title)
	body.Platform = strings.TrimSpace(body.Platform)
	return body
}

func parsePrizeTimes(req models.PrizePayload) (time.Time, time.Time, common.Error) {
	if strings.TrimSpace(req.Title) == "" {
		return time.Time{}, time.Time{}, common.NewValidationError("title is required")
	}
	start, err := parseDateTime(req.StartTime)
	if err != nil {
		return time.Time{}, time.Time{}, common.NewValidationError("invalid start_time")
	}
	end, err := parseDateTime(req.EndTime)
	if err != nil {
		return time.Time{}, time.Time{}, common.NewValidationError("invalid end_time")
	}
	return start, end, nil
}

func parseDateTime(value string) (time.Time, error) {
	layouts := []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02 15:04:05", "2006-01-02 15:04"}
	value = strings.TrimSpace(value)
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fiber.ErrBadRequest
}

func uniqueInt64s(input []int64) []int64 {
	result := make([]int64, 0, len(input))
	seen := make(map[int64]struct{}, len(input))
	for _, item := range input {
		if item <= 0 {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func (api *gameAPI) auditTx(c fiber.Ctx, tx *gorm.DB, action, resource string, targetID int64, before, after any) common.Error {
	return audit.LogTx(tx, audit.MetaFromFiber(c), action, resource, targetID, before, after)
}

type tableNamer interface {
	TableName() string
}

func tableNameOf(model any) string {
	if named, ok := model.(tableNamer); ok {
		return named.TableName()
	}
	return ""
}
