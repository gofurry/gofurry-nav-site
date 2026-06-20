package service

import (
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/cachekeys"
	homeModels "github.com/gofurry/gofurry-nav-backend/apps/nav/home/models"
	homeService "github.com/gofurry/gofurry-nav-backend/apps/nav/home/service"
	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/sitegroup/models"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
)

const (
	defaultGroupPageSize = 24
	maxGroupPageSize     = 60
)

func GetCachedSiteGroupPage(lang, groupID string, page, pageSize int) models.SiteGroupPageResponse {
	page = normalizePage(page)
	pageSize = normalizePageSize(pageSize)
	response := models.SiteGroupPageResponse{
		SchemaVersion:  models.SiteGroupSchemaVersion,
		GeneratedAt:    time.Now(),
		State:          models.SiteGroupStateMissing,
		ReasonMessages: []string{"站点分组缓存未命中"},
		Page:           page,
		PageSize:       pageSize,
		Items:          []navmodels.SiteVo{},
	}

	raw, err := cs.GetString(cachekeys.SiteGroup(lang, groupID))
	if err != nil || raw == "" {
		if err != nil {
			response.ReasonMessages = []string{err.GetMsg()}
		}
		return response
	}

	var cached models.CachedSiteGroup
	if unmarshalErr := sonic.Unmarshal([]byte(raw), &cached); unmarshalErr != nil {
		response.ReasonMessages = []string{"站点分组缓存反序列化失败"}
		return response
	}

	response.State = models.SiteGroupStateReady
	response.GeneratedAt = cached.GeneratedAt
	response.Group = &cached.Group
	response.Total = len(cached.Sites)
	response.Items = paginateSites(cached.Sites, page, pageSize)
	response.HasMore = page*pageSize < response.Total
	response.ReasonMessages = nil
	return response
}

func BuildCachedSiteGroupPayload(groups []homeModels.HomeGroup, groupID string, sites []navmodels.SiteVo) models.CachedSiteGroup {
	groupID = strings.TrimSpace(groupID)
	for _, group := range groups {
		if strings.TrimSpace(group.ID) != groupID {
			continue
		}
		info := models.SiteGroupInfo{
			ID:         group.ID,
			Name:       group.Name,
			Info:       group.Info,
			Priority:   group.Priority,
			SiteCount:  len(sites),
			DetailPath: group.DetailPath,
		}
		return models.CachedSiteGroup{
			GeneratedAt: time.Now(),
			Group:       info,
			Sites:       append([]navmodels.SiteVo(nil), sites...),
		}
	}
	return models.CachedSiteGroup{
		GeneratedAt: time.Now(),
		Sites:       append([]navmodels.SiteVo(nil), sites...),
	}
}

func BuildAllSiteGroupCaches(sites []navmodels.SiteVo, groups []navmodels.GroupVo) map[string]models.CachedSiteGroup {
	siteMap := make(map[string]navmodels.SiteVo, len(sites))
	for _, site := range sites {
		siteMap[site.ID] = site
	}

	previews := homeService.BuildHomeGroupsForCache(sites, groups)
	result := make(map[string]models.CachedSiteGroup, len(groups))
	for _, group := range groups {
		items := make([]navmodels.SiteVo, 0, len(group.Sites))
		for _, siteID := range group.Sites {
			if site, ok := siteMap[siteID]; ok {
				if group.SiteWeights != nil {
					site.GroupWeight = group.SiteWeights[siteID]
				}
				items = append(items, site)
			}
		}
		homeService.SortGroupSitesForCache(items)
		result[group.ID] = BuildCachedSiteGroupPayload(previews, group.ID, items)
	}
	return result
}

func normalizePage(value int) int {
	if value > 0 {
		return value
	}
	return 1
}

func normalizePageSize(value int) int {
	if value <= 0 {
		return defaultGroupPageSize
	}
	if value > maxGroupPageSize {
		return maxGroupPageSize
	}
	return value
}

func paginateSites(items []navmodels.SiteVo, page, pageSize int) []navmodels.SiteVo {
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []navmodels.SiteVo{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return append([]navmodels.SiteVo(nil), items[start:end]...)
}

func ParsePositiveInt(raw string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
