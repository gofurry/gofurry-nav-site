package controller

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/navadmin/models"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/shared/adminutil"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/shared/audit"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/infra/db"
	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/common"
	pkgmodels "github.com/gofurry/awesome-fiber-template/v3/medium/pkg/models"
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

func (api *navAPI) ListUpdateNotices(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	base := adminutil.ApplyKeyword(navDB().Model(&models.UpdateNotice{}).Where("deleted IS NOT TRUE").Order("published_at DESC, id DESC"), page.Keyword, "title", "title_en", "body", "body_en", "CAST(id AS TEXT)")
	var items []models.UpdateNotice
	total, err := adminutil.Paginate(base, page, &items)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}

func (api *navAPI) CreateUpdateNotice(c fiber.Ctx) error {
	var req models.UpdateNoticePayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	title, titleEn, body, bodyEn, publishedAt, validateErr := normalizeUpdateNoticePayload(req)
	if validateErr != nil {
		return common.NewResponse(c).Error(validateErr)
	}
	var created models.UpdateNotice
	err := navDB().Transaction(func(tx *gorm.DB) error {
		ids, allocErr := adminutil.AllocateSequentialIDs(tx, created.TableName(), 1)
		if allocErr != nil {
			return allocErr
		}
		created = models.UpdateNotice{
			ID:          ids[0],
			Title:       title,
			TitleEn:     titleEn,
			Body:        body,
			BodyEn:      bodyEn,
			PublishedAt: pkgmodels.LocalTime(publishedAt),
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

func (api *navAPI) GetUpdateNotice(c fiber.Ctx) error {
	return api.getOne(c, navDB().Where("deleted IS NOT TRUE"), &models.UpdateNotice{})
}

func (api *navAPI) UpdateUpdateNotice(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req models.UpdateNoticePayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	title, titleEn, body, bodyEn, publishedAt, validateErr := normalizeUpdateNoticePayload(req)
	if validateErr != nil {
		return common.NewResponse(c).Error(validateErr)
	}
	txErr := navDB().Transaction(func(tx *gorm.DB) error {
		before, snapErr := audit.SnapshotByID(tx, (&models.UpdateNotice{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		if err := tx.Model(&models.UpdateNotice{}).Where("id = ? AND deleted IS NOT TRUE", id).Updates(map[string]any{
			"title":        title,
			"title_en":     titleEn,
			"body":         body,
			"body_en":      bodyEn,
			"published_at": pkgmodels.LocalTime(publishedAt),
		}).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		after, snapErr := audit.SnapshotByID(tx, (&models.UpdateNotice{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		return api.auditTx(c, tx, "update", (&models.UpdateNotice{}).TableName(), id, before, after)
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return api.GetUpdateNotice(c)
}

func (api *navAPI) DeleteUpdateNotice(c fiber.Ctx) error {
	return api.deleteSoft(c, &models.UpdateNotice{})
}

func (api *navAPI) ListCollectorDomains(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	base := navDB().Table((&models.CollectorDomain{}).TableName() + " AS cd").
		Select("cd.id, cd.site_id, COALESCE(s.name, '') AS site_name, cd.name, cd.proxy, cd.prefix, cd.tls, cd.deleted").
		Joins("LEFT JOIN " + (&models.Site{}).TableName() + " AS s ON s.id = cd.site_id").
		Where("cd.deleted IS NOT TRUE").
		Order("cd.id DESC")
	base = adminutil.ApplyKeyword(base, page.Keyword, "cd.name", "cd.proxy", "cd.tls", "s.name", "s.name_en", "CAST(cd.id AS TEXT)", "CAST(cd.site_id AS TEXT)")
	var items []models.CollectorDomainDTO
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
	var created models.CollectorDomain
	err := navDB().Transaction(func(tx *gorm.DB) error {
		if validateErr := validateCollectorDomainPayload(tx, req); validateErr != nil {
			return validateErr
		}
		ids, allocErr := adminutil.AllocateSequentialIDs(tx, created.TableName(), 1)
		if allocErr != nil {
			return allocErr
		}
		created = models.CollectorDomain{
			ID:      ids[0],
			SiteID:  req.SiteID,
			Name:    strings.TrimSpace(req.Name),
			Proxy:   strings.TrimSpace(req.Proxy),
			Prefix:  normalizeStringPtr(req.Prefix),
			TLS:     strings.TrimSpace(req.TLS),
			Deleted: false,
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
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var item models.CollectorDomainDTO
	collectorDomainTable := (&models.CollectorDomain{}).TableName()
	siteTable := (&models.Site{}).TableName()
	dbErr := navDB().Table(collectorDomainTable+" AS cd").
		Select("cd.id, cd.site_id, COALESCE(s.name, '') AS site_name, cd.name, cd.proxy, cd.prefix, cd.tls, cd.deleted").
		Joins("LEFT JOIN "+siteTable+" AS s ON s.id = cd.site_id").
		Where("cd.id = ? AND cd.deleted IS NOT TRUE", id).
		Take(&item).Error
	if dbErr != nil {
		return common.NewResponse(c).Error(common.NewDaoError(dbErr.Error()))
	}
	return common.NewResponse(c).SuccessWithData(item)
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
	txErr := navDB().Transaction(func(tx *gorm.DB) error {
		if validateErr := validateCollectorDomainPayload(tx, req); validateErr != nil {
			return validateErr
		}
		before, snapErr := audit.SnapshotByID(tx, (&models.CollectorDomain{}).TableName(), id)
		if snapErr != nil {
			return snapErr
		}
		if err := tx.Model(&models.CollectorDomain{}).Where("id = ? AND deleted IS NOT TRUE", id).Updates(map[string]any{
			"site_id": req.SiteID,
			"name":    strings.TrimSpace(req.Name),
			"proxy":   strings.TrimSpace(req.Proxy),
			"prefix":  normalizeStringPtr(req.Prefix),
			"tls":     strings.TrimSpace(req.TLS),
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
	return api.deleteSoft(c, &models.CollectorDomain{})
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
	return nil
}

func normalizeUpdateNoticePayload(req models.UpdateNoticePayload) (string, string, string, string, time.Time, common.Error) {
	title := strings.TrimSpace(req.Title)
	titleEn := strings.TrimSpace(req.TitleEn)
	body := strings.TrimSpace(req.Body)
	bodyEn := strings.TrimSpace(req.BodyEn)
	if title == "" || titleEn == "" || body == "" || bodyEn == "" {
		return "", "", "", "", time.Time{}, common.NewValidationError("title, title_en, body and body_en are required")
	}
	publishedAt, err := parseDateTime(req.PublishedAt)
	if err != nil {
		return "", "", "", "", time.Time{}, common.NewValidationError("invalid published_at")
	}
	return title, titleEn, body, bodyEn, publishedAt, nil
}

func parseDateTime(value string) (time.Time, error) {
	layouts := []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02 15:04"}
	value = strings.TrimSpace(value)
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fiber.ErrBadRequest
}

func validateCollectorDomainPayload(tx *gorm.DB, req models.CollectorDomainPayload) common.Error {
	if req.SiteID <= 0 {
		return common.NewValidationError("site_id is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return common.NewValidationError("name is required")
	}
	if strings.TrimSpace(req.Proxy) == "" {
		return common.NewValidationError("proxy is required")
	}
	if strings.TrimSpace(req.TLS) == "" {
		return common.NewValidationError("tls is required")
	}

	var count int64
	if err := tx.Model(&models.Site{}).Where("id = ? AND deleted IS NOT TRUE", req.SiteID).Count(&count).Error; err != nil {
		return common.NewDaoError(err.Error())
	}
	if count == 0 {
		return common.NewValidationError("site_id must reference an existing site")
	}
	return nil
}

func normalizeStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func siteDTO(item models.Site) models.SiteDTO {
	return models.SiteDTO{
		ID:         item.ID,
		Name:       item.Name,
		NameEn:     item.NameEn,
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
