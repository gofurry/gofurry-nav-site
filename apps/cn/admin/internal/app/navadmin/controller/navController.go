package controller

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/navadmin/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	"github.com/gofurry/gofurry-admin/internal/infra/cache"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/jackc/pgx/v5/pgxpool"
)

type navAPI struct{ store *navStore }

type NavAPI = navAPI

func New(pool *pgxpool.Pool, auditLogger *audit.Logger) *NavAPI {
	return &navAPI{store: newNavStore(pool, auditLogger)}
}

const (
	navSiteListCacheKey     = "site:list:v2"
	navGroupListCacheKey    = "group:list"
	navGroupSiteMapCacheKey = "group:site:map"
	navFeaturedSiteCacheKey = "featured-sites:list"
	siteGroupMapListOrder   = "m.id DESC"
)

func normalizeSayingLanguage(lang string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(lang))
	if normalized == "" {
		return "zh", nil
	}
	if normalized != "zh" && normalized != "en" {
		return "", common.NewValidationError("language must be zh or en")
	}
	return normalized, nil
}

func invalidateNavCache(keys ...string) {
	if len(keys) > 0 && cache.RedisReady() {
		_ = cache.Del(keys...)
	}
}

// Public nav pages are cache-only; derived hot caches remain until Nav Backend refreshes them.
func invalidateNavSiteListCache()  { invalidateNavCache(navSiteListCacheKey) }
func invalidateNavGroupListCache() { invalidateNavCache(navGroupListCacheKey) }
func invalidateNavGroupMapCache()  { invalidateNavCache(navGroupSiteMapCacheKey) }
func invalidateNavFeaturedSiteCache() {
	invalidateNavCache(navFeaturedSiteCacheKey)
}

func (api *navAPI) ListSayings(c fiber.Ctx) error {
	total, items, err := api.store.listSayings(c.Context(), adminutil.ParsePageQuery(c))
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
	language, err := normalizeSayingLanguage(req.Language)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	req.Language = language
	item, storeErr := api.store.createSaying(c.Context(), audit.MetaFromFiber(c), req)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	return common.NewResponse(c).SuccessWithData(item)
}

func (api *navAPI) GetSaying(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	item, storeErr := api.store.getSaying(c.Context(), id)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
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
	language, validateErr := normalizeSayingLanguage(req.Language)
	if validateErr != nil {
		return common.NewResponse(c).Error(validateErr)
	}
	req.Language = language
	if storeErr := api.store.updateSaying(c.Context(), audit.MetaFromFiber(c), id, req); storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	return api.GetSaying(c)
}

func (api *navAPI) DeleteSaying(c fiber.Ctx) error {
	return api.delete(c, api.store.deleteSaying)
}

func (api *navAPI) ListUpdateNotices(c fiber.Ctx) error {
	total, items, err := api.store.listUpdateNotices(c.Context(), adminutil.ParsePageQuery(c))
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}

func (api *navAPI) CreateUpdateNotice(c fiber.Ctx) error {
	req, publishedAt, err := decodeUpdateNotice(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	item, storeErr := api.store.createUpdateNotice(c.Context(), audit.MetaFromFiber(c), req, publishedAt)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	return common.NewResponse(c).SuccessWithData(item)
}

func (api *navAPI) GetUpdateNotice(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	item, storeErr := api.store.getUpdateNotice(c.Context(), id)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	return common.NewResponse(c).SuccessWithData(item)
}

func (api *navAPI) UpdateUpdateNotice(c fiber.Ctx) error {
	id, idErr := adminutil.ParseIDParam(c)
	if idErr != nil {
		return common.NewResponse(c).Error(idErr)
	}
	req, publishedAt, err := decodeUpdateNotice(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if storeErr := api.store.updateUpdateNotice(c.Context(), audit.MetaFromFiber(c), id, req, publishedAt); storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	return api.GetUpdateNotice(c)
}

func (api *navAPI) DeleteUpdateNotice(c fiber.Ctx) error {
	return api.delete(c, api.store.deleteUpdateNotice)
}

func (api *navAPI) ListCollectorDomains(c fiber.Ctx) error {
	total, items, err := api.store.listCollectorDomains(c.Context(), adminutil.ParsePageQuery(c))
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}

func (api *navAPI) CreateCollectorDomain(c fiber.Ctx) error {
	req, err := decodeCollectorDomain(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	item, storeErr := api.store.createCollectorDomain(c.Context(), audit.MetaFromFiber(c), req)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	invalidateNavSiteListCache()
	return common.NewResponse(c).SuccessWithData(item)
}

func (api *navAPI) GetCollectorDomain(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	item, storeErr := api.store.getCollectorDomain(c.Context(), id)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	return common.NewResponse(c).SuccessWithData(item)
}

func (api *navAPI) UpdateCollectorDomain(c fiber.Ctx) error {
	id, idErr := adminutil.ParseIDParam(c)
	if idErr != nil {
		return common.NewResponse(c).Error(idErr)
	}
	req, err := decodeCollectorDomain(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if storeErr := api.store.updateCollectorDomain(c.Context(), audit.MetaFromFiber(c), id, req); storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	invalidateNavSiteListCache()
	return api.GetCollectorDomain(c)
}

func (api *navAPI) DeleteCollectorDomain(c fiber.Ctx) error {
	if err := api.delete(c, api.store.deleteCollectorDomain); err != nil {
		return err
	}
	invalidateNavSiteListCache()
	return nil
}

func (api *navAPI) SetPrimaryCollectorDomain(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if storeErr := api.store.setPrimaryCollectorDomain(c.Context(), audit.MetaFromFiber(c), id); storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	invalidateNavSiteListCache()
	return common.NewResponse(c).Success()
}

func (api *navAPI) ListSites(c fiber.Ctx) error {
	total, items, err := api.store.listSites(c.Context(), adminutil.ParsePageQuery(c))
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}

func (api *navAPI) ListSiteWorkspaceSummaries(c fiber.Ctx) error {
	total, items, err := api.store.listSiteWorkspaceSummaries(c.Context(), adminutil.ParsePageQuery(c))
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}

func (api *navAPI) CreateSite(c fiber.Ctx) error {
	req, err := decodeSite(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	item, storeErr := api.store.createSite(c.Context(), audit.MetaFromFiber(c), req)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	invalidateNavSiteListCache()
	return common.NewResponse(c).SuccessWithData(item)
}

func (api *navAPI) GetSite(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	item, storeErr := api.store.getSite(c.Context(), id)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	return common.NewResponse(c).SuccessWithData(item)
}

func (api *navAPI) GetSiteWorkspace(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	item, storeErr := api.store.getSiteWorkspace(c.Context(), id)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	return common.NewResponse(c).SuccessWithData(item)
}

func (api *navAPI) UpdateSite(c fiber.Ctx) error {
	id, idErr := adminutil.ParseIDParam(c)
	if idErr != nil {
		return common.NewResponse(c).Error(idErr)
	}
	req, err := decodeSite(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if storeErr := api.store.updateSite(c.Context(), audit.MetaFromFiber(c), id, req); storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	invalidateNavSiteListCache()
	return api.GetSite(c)
}

func (api *navAPI) DeleteSite(c fiber.Ctx) error {
	if err := api.delete(c, api.store.deleteSite); err != nil {
		return err
	}
	invalidateNavSiteListCache()
	return nil
}

func (api *navAPI) ListSiteGroups(c fiber.Ctx) error {
	total, items, err := api.store.listSiteGroups(c.Context(), adminutil.ParsePageQuery(c))
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}

func (api *navAPI) CreateSiteGroup(c fiber.Ctx) error {
	req, err := decodeSiteGroup(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	item, storeErr := api.store.createSiteGroup(c.Context(), audit.MetaFromFiber(c), req)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	invalidateNavGroupListCache()
	return common.NewResponse(c).SuccessWithData(item)
}

func (api *navAPI) GetSiteGroup(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	item, storeErr := api.store.getSiteGroup(c.Context(), id)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	return common.NewResponse(c).SuccessWithData(item)
}

func (api *navAPI) UpdateSiteGroup(c fiber.Ctx) error {
	id, idErr := adminutil.ParseIDParam(c)
	if idErr != nil {
		return common.NewResponse(c).Error(idErr)
	}
	req, err := decodeSiteGroup(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if storeErr := api.store.updateSiteGroup(c.Context(), audit.MetaFromFiber(c), id, req); storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	invalidateNavGroupListCache()
	return api.GetSiteGroup(c)
}

func (api *navAPI) DeleteSiteGroup(c fiber.Ctx) error {
	if err := api.delete(c, api.store.deleteSiteGroup); err != nil {
		return err
	}
	invalidateNavCache(navGroupListCacheKey, navGroupSiteMapCacheKey)
	return nil
}

func (api *navAPI) ListSiteGroupMaps(c fiber.Ctx) error {
	total, items, err := api.store.listSiteGroupMaps(c.Context(), adminutil.ParsePageQuery(c))
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}

func (api *navAPI) CreateSiteGroupMap(c fiber.Ctx) error {
	req, err := decodeSiteGroupMap(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	item, storeErr := api.store.createSiteGroupMap(c.Context(), audit.MetaFromFiber(c), req)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	invalidateNavGroupMapCache()
	return common.NewResponse(c).SuccessWithData(item)
}

func (api *navAPI) GetSiteGroupMap(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	item, storeErr := api.store.getSiteGroupMap(c.Context(), id)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	return common.NewResponse(c).SuccessWithData(item)
}

func (api *navAPI) UpdateSiteGroupMap(c fiber.Ctx) error {
	id, idErr := adminutil.ParseIDParam(c)
	if idErr != nil {
		return common.NewResponse(c).Error(idErr)
	}
	req, err := decodeSiteGroupMap(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if storeErr := api.store.updateSiteGroupMap(c.Context(), audit.MetaFromFiber(c), id, req); storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	invalidateNavGroupMapCache()
	return api.GetSiteGroupMap(c)
}

func (api *navAPI) DeleteSiteGroupMap(c fiber.Ctx) error {
	if err := api.delete(c, api.store.deleteSiteGroupMap); err != nil {
		return err
	}
	invalidateNavGroupMapCache()
	return nil
}

func (api *navAPI) BulkReplaceSiteGroupMaps(c fiber.Ctx) error {
	var req adminutil.BulkReplaceRequest
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if req.OwnerID <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("owner_id is required"))
	}
	if err := api.store.replaceSiteGroupMaps(c.Context(), audit.MetaFromFiber(c), req.OwnerID, uniqueInt64s(req.IDs)); err != nil {
		return common.NewResponse(c).Error(err)
	}
	invalidateNavGroupMapCache()
	return common.NewResponse(c).Success()
}

func (api *navAPI) ListFeaturedSites(c fiber.Ctx) error {
	total, items, err := api.store.listFeaturedSites(c.Context(), adminutil.ParsePageQuery(c))
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}

func (api *navAPI) CreateFeaturedSite(c fiber.Ctx) error {
	req, err := decodeFeaturedSite(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	item, storeErr := api.store.createFeaturedSite(c.Context(), audit.MetaFromFiber(c), req)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	invalidateNavFeaturedSiteCache()
	return common.NewResponse(c).SuccessWithData(item)
}

func (api *navAPI) GetFeaturedSite(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	item, storeErr := api.store.getFeaturedSite(c.Context(), id)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	return common.NewResponse(c).SuccessWithData(item)
}

func (api *navAPI) UpdateFeaturedSite(c fiber.Ctx) error {
	id, idErr := adminutil.ParseIDParam(c)
	if idErr != nil {
		return common.NewResponse(c).Error(idErr)
	}
	req, err := decodeFeaturedSite(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if storeErr := api.store.updateFeaturedSite(c.Context(), audit.MetaFromFiber(c), id, req); storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	invalidateNavFeaturedSiteCache()
	return api.GetFeaturedSite(c)
}

func (api *navAPI) DeleteFeaturedSite(c fiber.Ctx) error {
	if err := api.delete(c, api.store.deleteFeaturedSite); err != nil {
		return err
	}
	invalidateNavFeaturedSiteCache()
	return nil
}

type deleteNavResource func(context.Context, audit.Meta, int64) common.Error

func (api *navAPI) delete(c fiber.Ctx, remove deleteNavResource) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if removeErr := remove(c.Context(), audit.MetaFromFiber(c), id); removeErr != nil {
		return common.NewResponse(c).Error(removeErr)
	}
	return common.NewResponse(c).Success()
}

func decodeUpdateNotice(c fiber.Ctx) (models.UpdateNoticePayload, time.Time, common.Error) {
	var req models.UpdateNoticePayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return req, time.Time{}, err
	}
	title, titleEn, body, bodyEn, publishedAt, err := normalizeUpdateNoticePayload(req)
	req.Title, req.TitleEn, req.Body, req.BodyEn = title, titleEn, body, bodyEn
	return req, publishedAt, err
}

func decodeCollectorDomain(c fiber.Ctx) (models.CollectorDomainPayload, common.Error) {
	var req models.CollectorDomainPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return req, err
	}
	req.Name, req.Proxy, req.TLS = strings.TrimSpace(req.Name), strings.TrimSpace(req.Proxy), strings.TrimSpace(req.TLS)
	req.Prefix = normalizeStringPtr(req.Prefix)
	if req.SiteID <= 0 {
		return req, common.NewValidationError("site_id is required")
	}
	if req.Name == "" {
		return req, common.NewValidationError("name is required")
	}
	if req.Proxy == "" {
		return req, common.NewValidationError("proxy is required")
	}
	if req.TLS == "" {
		return req, common.NewValidationError("tls is required")
	}
	return req, nil
}

func decodeSite(c fiber.Ctx) (models.SitePayload, common.Error) {
	var req models.SitePayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return req, err
	}
	req.Name, req.NameEn = strings.TrimSpace(req.Name), strings.TrimSpace(req.NameEn)
	req.Info, req.InfoEn = strings.TrimSpace(req.Info), strings.TrimSpace(req.InfoEn)
	req.Nsfw, req.Welfare = strings.TrimSpace(req.Nsfw), strings.TrimSpace(req.Welfare)
	return req, validateSitePayload(req)
}

func decodeSiteGroup(c fiber.Ctx) (models.SiteGroupPayload, common.Error) {
	var req models.SiteGroupPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return req, err
	}
	req.Name, req.NameEn = strings.TrimSpace(req.Name), strings.TrimSpace(req.NameEn)
	req.Info, req.InfoEn = strings.TrimSpace(req.Info), strings.TrimSpace(req.InfoEn)
	if req.Name == "" {
		return req, common.NewValidationError("name is required")
	}
	return req, nil
}

func decodeSiteGroupMap(c fiber.Ctx) (models.SiteGroupMapPayload, common.Error) {
	var req models.SiteGroupMapPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return req, err
	}
	if req.SiteID <= 0 || req.GroupID <= 0 {
		return req, common.NewValidationError("site_id and group_id are required")
	}
	return req, nil
}

func decodeFeaturedSite(c fiber.Ctx) (models.FeaturedSitePayload, common.Error) {
	var req models.FeaturedSitePayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return req, err
	}
	if req.SiteID <= 0 {
		return req, common.NewValidationError("site_id is required")
	}
	return req, nil
}

func validateSitePayload(req models.SitePayload) common.Error {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.NameEn) == "" {
		return common.NewValidationError("name and name_en are required")
	}
	return nil
}

func normalizeUpdateNoticePayload(req models.UpdateNoticePayload) (string, string, string, string, time.Time, common.Error) {
	title, titleEn := strings.TrimSpace(req.Title), strings.TrimSpace(req.TitleEn)
	body, bodyEn := strings.TrimSpace(req.Body), strings.TrimSpace(req.BodyEn)
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
	return models.SiteDTO{ID: item.ID, Name: item.Name, NameEn: item.NameEn, Info: item.Info, InfoEn: item.InfoEn,
		CreateTime: item.CreateTime, UpdateTime: item.UpdateTime, Country: item.Country, Nsfw: item.Nsfw,
		Welfare: item.Welfare, Icon: item.Icon, Deleted: item.Deleted}
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
