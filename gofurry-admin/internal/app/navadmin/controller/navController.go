package controller

import (
	"encoding/json"
	"strings"

	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/navadmin/models"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/shared/adminutil"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/shared/audit"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/infra/db"
	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/common"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type navAPI struct{}

var NavAPI = &navAPI{}

func navDB() *gorm.DB {
	return db.Databases.DB(db.Nav)
}

func (api *navAPI) ListSayings(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	base := adminutil.ApplyKeyword(navDB().Model(&models.Saying{}).Order("id DESC"), page.Keyword, "author", "saying", "CAST(id AS TEXT)")
	var items []models.Saying
	total, err := adminutil.Paginate(base, page, &items)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}

func (api *navAPI) CreateSaying(c fiber.Ctx) error {
	var req models.SayingPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	req.Saying = strings.TrimSpace(req.Saying)
	if req.Saying == "" {
		return common.NewResponse(c).Error(common.NewValidationError("saying is required"))
	}

	var created models.Saying
	err := navDB().Transaction(func(tx *gorm.DB) error {
		ids, allocErr := adminutil.AllocateSequentialIDs(tx, created.TableName(), 1)
		if allocErr != nil {
			return allocErr
		}
		created = models.Saying{ID: ids[0], Author: req.Author, Saying: req.Saying}
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

func (api *navAPI) GetSaying(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var item models.Saying
	if err := navDB().First(&item, "id = ?", id).Error; err != nil {
		return common.NewResponse(c).Error(common.NewDaoError(err.Error()))
	}
	return common.NewResponse(c).SuccessWithData(item)
}

func (api *navAPI) UpdateSaying(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req models.SayingPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	req.Saying = strings.TrimSpace(req.Saying)
	if req.Saying == "" {
		return common.NewResponse(c).Error(common.NewValidationError("saying is required"))
	}
	txErr := navDB().Transaction(func(tx *gorm.DB) error {
		before, snapErr := audit.SnapshotByID(tx, (&models.Saying{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		if err := tx.Model(&models.Saying{}).Where("id = ?", id).Updates(map[string]any{
			"author": req.Author,
			"saying": req.Saying,
		}).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		after, snapErr := audit.SnapshotByID(tx, (&models.Saying{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "update", (&models.Saying{}).TableName(), id, before, after)
	})
	if txErr != nil {
		return common.NewResponse(c).Error(common.NewDaoError(txErr.Error()))
	}
	return api.GetSaying(c)
}

func (api *navAPI) DeleteSaying(c fiber.Ctx) error {
	return api.deleteHard(c, &models.Saying{})
}

func (api *navAPI) ListLogUpdates(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	base := adminutil.ApplyKeyword(navDB().Model(&models.LogUpdate{}).Where("deleted IS NOT TRUE").Order("id DESC"), page.Keyword, "title", "url", "CAST(id AS TEXT)")
	var items []models.LogUpdate
	total, err := adminutil.Paginate(base, page, &items)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}

func (api *navAPI) CreateLogUpdate(c fiber.Ctx) error {
	var req models.LogUpdatePayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.URL) == "" {
		return common.NewResponse(c).Error(common.NewValidationError("title and url are required"))
	}
	var created models.LogUpdate
	err := navDB().Transaction(func(tx *gorm.DB) error {
		ids, allocErr := adminutil.AllocateSequentialIDs(tx, created.TableName(), 1)
		if allocErr != nil {
			return allocErr
		}
		created = models.LogUpdate{ID: ids[0], Title: strings.TrimSpace(req.Title), URL: strings.TrimSpace(req.URL)}
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

func (api *navAPI) GetLogUpdate(c fiber.Ctx) error {
	return api.getOne(c, navDB().Where("deleted IS NOT TRUE"), &models.LogUpdate{})
}

func (api *navAPI) UpdateLogUpdate(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req models.LogUpdatePayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.URL) == "" {
		return common.NewResponse(c).Error(common.NewValidationError("title and url are required"))
	}
	txErr := navDB().Transaction(func(tx *gorm.DB) error {
		before, snapErr := audit.SnapshotByID(tx, (&models.LogUpdate{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		if err := tx.Model(&models.LogUpdate{}).Where("id = ? AND deleted IS NOT TRUE", id).Updates(map[string]any{
			"title": strings.TrimSpace(req.Title),
			"url":   strings.TrimSpace(req.URL),
		}).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		after, snapErr := audit.SnapshotByID(tx, (&models.LogUpdate{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "update", (&models.LogUpdate{}).TableName(), id, before, after)
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return api.GetLogUpdate(c)
}

func (api *navAPI) DeleteLogUpdate(c fiber.Ctx) error {
	return api.deleteSoft(c, &models.LogUpdate{})
}

func (api *navAPI) ListCollectorDomains(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	base := adminutil.ApplyKeyword(navDB().Model(&models.CollectorDomain{}).Order("id DESC"), page.Keyword, "name", "proxy", "tls", "CAST(id AS TEXT)")
	var items []models.CollectorDomain
	total, err := adminutil.Paginate(base, page, &items)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}

func (api *navAPI) CreateCollectorDomain(c fiber.Ctx) error {
	var req models.CollectorDomainPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if strings.TrimSpace(req.Name) == "" {
		return common.NewResponse(c).Error(common.NewValidationError("name is required"))
	}
	var created models.CollectorDomain
	err := navDB().Transaction(func(tx *gorm.DB) error {
		ids, allocErr := adminutil.AllocateSequentialIDs(tx, created.TableName(), 1)
		if allocErr != nil {
			return allocErr
		}
		created = models.CollectorDomain{
			ID:     ids[0],
			Name:   strings.TrimSpace(req.Name),
			Proxy:  strings.TrimSpace(req.Proxy),
			Prefix: req.Prefix,
			TLS:    strings.TrimSpace(req.TLS),
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
	return common.NewResponse(c).SuccessWithData(created)
}

func (api *navAPI) GetCollectorDomain(c fiber.Ctx) error {
	return api.getOne(c, navDB(), &models.CollectorDomain{})
}

func (api *navAPI) UpdateCollectorDomain(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req models.CollectorDomainPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if strings.TrimSpace(req.Name) == "" {
		return common.NewResponse(c).Error(common.NewValidationError("name is required"))
	}
	txErr := navDB().Transaction(func(tx *gorm.DB) error {
		before, snapErr := audit.SnapshotByID(tx, (&models.CollectorDomain{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		if err := tx.Model(&models.CollectorDomain{}).Where("id = ?", id).Updates(map[string]any{
			"name":   strings.TrimSpace(req.Name),
			"proxy":  strings.TrimSpace(req.Proxy),
			"prefix": req.Prefix,
			"tls":    strings.TrimSpace(req.TLS),
		}).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		after, snapErr := audit.SnapshotByID(tx, (&models.CollectorDomain{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "update", (&models.CollectorDomain{}).TableName(), id, before, after)
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return api.GetCollectorDomain(c)
}

func (api *navAPI) DeleteCollectorDomain(c fiber.Ctx) error {
	return api.deleteHard(c, &models.CollectorDomain{})
}

func (api *navAPI) ListSites(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	base := adminutil.ApplyKeyword(navDB().Model(&models.Site{}).Where("deleted IS NOT TRUE").Order("id DESC"), page.Keyword, "name", "name_en", "info", "info_en", "CAST(id AS TEXT)")
	var items []models.Site
	total, err := adminutil.Paginate(base, page, &items)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	result := make([]models.SiteDTO, 0, len(items))
	for _, item := range items {
		result = append(result, siteDTO(item))
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, result))
}

func (api *navAPI) CreateSite(c fiber.Ctx) error {
	var req models.SitePayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if err := validateSitePayload(req); err != nil {
		return common.NewResponse(c).Error(err)
	}

	var created models.Site
	err := navDB().Transaction(func(tx *gorm.DB) error {
		ids, allocErr := adminutil.AllocateSequentialIDs(tx, created.TableName(), 1)
		if allocErr != nil {
			return allocErr
		}
		created = models.Site{
			ID:      ids[0],
			Name:    strings.TrimSpace(req.Name),
			NameEn:  strings.TrimSpace(req.NameEn),
			Domain:  adminutil.MustJSON(map[string][]string{"domain": normalizedDomains(req.Domains)}),
			Info:    strings.TrimSpace(req.Info),
			InfoEn:  strings.TrimSpace(req.InfoEn),
			Country: req.Country,
			Nsfw:    strings.TrimSpace(req.Nsfw),
			Welfare: strings.TrimSpace(req.Welfare),
			Icon:    req.Icon,
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
	return common.NewResponse(c).SuccessWithData(siteDTO(created))
}

func (api *navAPI) GetSite(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var item models.Site
	if err := navDB().Where("deleted IS NOT TRUE").First(&item, "id = ?", id).Error; err != nil {
		return common.NewResponse(c).Error(common.NewDaoError(err.Error()))
	}
	return common.NewResponse(c).SuccessWithData(siteDTO(item))
}

func (api *navAPI) UpdateSite(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req models.SitePayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if err := validateSitePayload(req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	txErr := navDB().Transaction(func(tx *gorm.DB) error {
		before, snapErr := audit.SnapshotByID(tx, (&models.Site{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		if err := tx.Model(&models.Site{}).Where("id = ? AND deleted IS NOT TRUE", id).Updates(map[string]any{
			"name":    strings.TrimSpace(req.Name),
			"name_en": strings.TrimSpace(req.NameEn),
			"domain":  adminutil.MustJSON(map[string][]string{"domain": normalizedDomains(req.Domains)}),
			"info":    strings.TrimSpace(req.Info),
			"info_en": strings.TrimSpace(req.InfoEn),
			"country": req.Country,
			"nsfw":    strings.TrimSpace(req.Nsfw),
			"welfare": strings.TrimSpace(req.Welfare),
			"icon":    req.Icon,
		}).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		after, snapErr := audit.SnapshotByID(tx, (&models.Site{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "update", (&models.Site{}).TableName(), id, before, after)
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return api.GetSite(c)
}

func (api *navAPI) DeleteSite(c fiber.Ctx) error {
	return api.deleteSoft(c, &models.Site{})
}

func (api *navAPI) ListSiteGroups(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	base := adminutil.ApplyKeyword(navDB().Model(&models.SiteGroup{}).Order("priority DESC, id DESC"), page.Keyword, "name", "name_en", "info", "info_en", "CAST(id AS TEXT)")
	var items []models.SiteGroup
	total, err := adminutil.Paginate(base, page, &items)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}

func (api *navAPI) CreateSiteGroup(c fiber.Ctx) error {
	var req models.SiteGroupPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if strings.TrimSpace(req.Name) == "" {
		return common.NewResponse(c).Error(common.NewValidationError("name is required"))
	}
	var created models.SiteGroup
	err := navDB().Transaction(func(tx *gorm.DB) error {
		ids, allocErr := adminutil.AllocateSequentialIDs(tx, created.TableName(), 1)
		if allocErr != nil {
			return allocErr
		}
		created = models.SiteGroup{
			ID:       ids[0],
			Name:     strings.TrimSpace(req.Name),
			NameEn:   strings.TrimSpace(req.NameEn),
			Info:     strings.TrimSpace(req.Info),
			InfoEn:   strings.TrimSpace(req.InfoEn),
			Priority: req.Priority,
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
	return common.NewResponse(c).SuccessWithData(created)
}

func (api *navAPI) GetSiteGroup(c fiber.Ctx) error {
	return api.getOne(c, navDB(), &models.SiteGroup{})
}

func (api *navAPI) UpdateSiteGroup(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req models.SiteGroupPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if strings.TrimSpace(req.Name) == "" {
		return common.NewResponse(c).Error(common.NewValidationError("name is required"))
	}
	txErr := navDB().Transaction(func(tx *gorm.DB) error {
		before, snapErr := audit.SnapshotByID(tx, (&models.SiteGroup{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		if err := tx.Model(&models.SiteGroup{}).Where("id = ?", id).Updates(map[string]any{
			"name":     strings.TrimSpace(req.Name),
			"name_en":  strings.TrimSpace(req.NameEn),
			"info":     strings.TrimSpace(req.Info),
			"info_en":  strings.TrimSpace(req.InfoEn),
			"priority": req.Priority,
		}).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		after, snapErr := audit.SnapshotByID(tx, (&models.SiteGroup{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "update", (&models.SiteGroup{}).TableName(), id, before, after)
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return api.GetSiteGroup(c)
}

func (api *navAPI) DeleteSiteGroup(c fiber.Ctx) error {
	return api.deleteHard(c, &models.SiteGroup{})
}

func (api *navAPI) ListSiteGroupMaps(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	base := navDB().Table("gfn_site_group_map AS m").
		Select("m.id, m.site_id, m.group_id, m.create_time, m.update_time, s.name AS site_name, g.name AS group_name").
		Joins("LEFT JOIN gfn_site s ON s.id = m.site_id").
		Joins("LEFT JOIN gfn_site_group g ON g.id = m.group_id").
		Order("m.id DESC")
	base = adminutil.ApplyKeyword(base, page.Keyword, "CAST(m.id AS TEXT)", "CAST(m.site_id AS TEXT)", "CAST(m.group_id AS TEXT)", "s.name", "g.name")
	var items []models.SiteGroupMapDTO
	total, err := adminutil.Paginate(base, page, &items)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}

func (api *navAPI) CreateSiteGroupMap(c fiber.Ctx) error {
	var req models.SiteGroupMapPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if req.SiteID <= 0 || req.GroupID <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("site_id and group_id are required"))
	}
	var created models.SiteGroupMap
	err := navDB().Transaction(func(tx *gorm.DB) error {
		ids, allocErr := adminutil.AllocateSequentialIDs(tx, created.TableName(), 1)
		if allocErr != nil {
			return allocErr
		}
		created = models.SiteGroupMap{ID: ids[0], SiteID: req.SiteID, GroupID: req.GroupID}
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

func (api *navAPI) GetSiteGroupMap(c fiber.Ctx) error {
	return api.getOne(c, navDB(), &models.SiteGroupMap{})
}

func (api *navAPI) UpdateSiteGroupMap(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req models.SiteGroupMapPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if req.SiteID <= 0 || req.GroupID <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("site_id and group_id are required"))
	}
	txErr := navDB().Transaction(func(tx *gorm.DB) error {
		before, snapErr := audit.SnapshotByID(tx, (&models.SiteGroupMap{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		if err := tx.Model(&models.SiteGroupMap{}).Where("id = ?", id).Updates(map[string]any{
			"site_id":  req.SiteID,
			"group_id": req.GroupID,
		}).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		after, snapErr := audit.SnapshotByID(tx, (&models.SiteGroupMap{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "update", (&models.SiteGroupMap{}).TableName(), id, before, after)
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return api.GetSiteGroupMap(c)
}

func (api *navAPI) DeleteSiteGroupMap(c fiber.Ctx) error {
	return api.deleteHard(c, &models.SiteGroupMap{})
}

func (api *navAPI) BulkReplaceSiteGroupMaps(c fiber.Ctx) error {
	var req adminutil.BulkReplaceRequest
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if req.OwnerID <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("owner_id is required"))
	}
	req.IDs = uniqueInt64s(req.IDs)

	err := navDB().Transaction(func(tx *gorm.DB) error {
		before, snapErr := audit.SnapshotRows(tx, (&models.SiteGroupMap{}).TableName(), "site_id = ?", "id ASC", req.OwnerID)
		if snapErr != nil {
			return snapErr
		}
		if err := tx.Where("site_id = ?", req.OwnerID).Delete(&models.SiteGroupMap{}).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		if len(req.IDs) == 0 {
			return api.auditTx(c, tx, "bulk_replace", (&models.SiteGroupMap{}).TableName(), req.OwnerID, before, []map[string]any{})
		}
		ids, allocErr := adminutil.AllocateSequentialIDs(tx, (&models.SiteGroupMap{}).TableName(), len(req.IDs))
		if allocErr != nil {
			return allocErr
		}
		rows := make([]models.SiteGroupMap, 0, len(req.IDs))
		for idx, groupID := range req.IDs {
			rows = append(rows, models.SiteGroupMap{ID: ids[idx], SiteID: req.OwnerID, GroupID: groupID})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
		after, snapErr := audit.SnapshotRows(tx, (&models.SiteGroupMap{}).TableName(), "site_id = ?", "id ASC", req.OwnerID)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "bulk_replace", (&models.SiteGroupMap{}).TableName(), req.OwnerID, before, after)
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).Success()
}

func (api *navAPI) deleteHard(c fiber.Ctx, model any) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	table := tableNameOf(model)
	txErr := navDB().Transaction(func(tx *gorm.DB) error {
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

func (api *navAPI) deleteSoft(c fiber.Ctx, model any) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	table := tableNameOf(model)
	txErr := navDB().Transaction(func(tx *gorm.DB) error {
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

func (api *navAPI) getOne(c fiber.Ctx, base *gorm.DB, out any) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if err := base.First(out, "id = ?", id).Error; err != nil {
		return common.NewResponse(c).Error(common.NewDaoError(err.Error()))
	}
	return common.NewResponse(c).SuccessWithData(out)
}

func validateSitePayload(req models.SitePayload) common.Error {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.NameEn) == "" {
		return common.NewValidationError("name and name_en are required")
	}
	if len(normalizedDomains(req.Domains)) == 0 {
		return common.NewValidationError("at least one domain is required")
	}
	return nil
}

func normalizedDomains(input []string) []string {
	result := make([]string, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for _, item := range input {
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

func siteDTO(item models.Site) models.SiteDTO {
	type domainPayload struct {
		Domain []string `json:"domain"`
	}
	var payload domainPayload
	_ = jsonUnmarshalString(item.Domain, &payload)
	return models.SiteDTO{
		ID:         item.ID,
		Name:       item.Name,
		NameEn:     item.NameEn,
		Domains:    normalizedDomains(payload.Domain),
		Info:       item.Info,
		InfoEn:     item.InfoEn,
		CreateTime: item.CreateTime,
		UpdateTime: item.UpdateTime,
		Country:    item.Country,
		Nsfw:       item.Nsfw,
		Welfare:    item.Welfare,
		Icon:       item.Icon,
		Deleted:    item.Deleted,
	}
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

func jsonUnmarshalString(raw string, target any) error {
	return json.Unmarshal([]byte(strings.TrimSpace(raw)), target)
}

func (api *navAPI) auditTx(c fiber.Ctx, tx *gorm.DB, action, resource string, targetID int64, before, after any) common.Error {
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
