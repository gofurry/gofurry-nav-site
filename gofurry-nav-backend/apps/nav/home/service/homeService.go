package service

import (
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/cachekeys"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/home/models"
	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	navservice "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/service"
	"github.com/gofurry/gofurry-nav-backend/common"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
)

const homeGroupPreviewLimit = 8

type navHomeReader interface {
	GetSiteList(lang string) ([]navmodels.SiteVo, common.GFError)
	GetGroupList(lang string) ([]navmodels.GroupVo, common.GFError)
	GetFeaturedSiteList() ([]navmodels.FeaturedSiteVo, common.GFError)
	GetPingList() (map[string]string, common.GFError)
	GetSayingService(lang string) (navmodels.SayingModel, common.GFError)
	GetImageUrl(t string) string
}

type homeService struct {
	navPage navHomeReader
	now     func() time.Time
}

var (
	homeSingleton = &homeService{}
	homeMu        sync.Mutex
)

func GetHomeService() *homeService {
	homeMu.Lock()
	defer homeMu.Unlock()
	if homeSingleton.navPage == nil {
		homeSingleton.navPage = navservice.GetNavPageService()
	}
	if homeSingleton.now == nil {
		homeSingleton.now = time.Now
	}
	return homeSingleton
}

func newHomeService(navPage navHomeReader, now func() time.Time) *homeService {
	return &homeService{navPage: navPage, now: now}
}

func (svc *homeService) GetHome(lang string) models.HomeResponse {
	lang = normalizeLang(lang)
	response := models.HomeResponse{
		SchemaVersion:  models.HomeSchemaVersion,
		GeneratedAt:    svc.clock()(),
		CacheState:     map[string]string{},
		ReasonMessages: map[string]string{},
		Groups:         []models.HomeGroup{},
		Spotlight: models.HomeSpotlight{
			PageSize: 6,
			Featured: []navmodels.SiteVo{},
			Popular:  []navmodels.SiteVo{},
			Latest:   []navmodels.SiteVo{},
			Random:   []navmodels.SiteVo{},
		},
		Ping:        map[string]string{},
		Backgrounds: models.HomeBackgrounds{},
	}

	var sites []navmodels.SiteVo
	if siteList, err := svc.reader().GetSiteList(lang); err != nil {
		response.CacheState["sites"] = models.HomeStateMissing
		response.ReasonMessages["sites"] = err.GetMsg()
	} else {
		response.CacheState["sites"] = models.HomeStateReady
		sites = siteList
	}

	if groups, err := svc.reader().GetGroupList(lang); err != nil {
		response.CacheState["groups"] = models.HomeStateMissing
		response.ReasonMessages["groups"] = err.GetMsg()
	} else {
		response.CacheState["groups"] = models.HomeStateReady
		if len(sites) > 0 {
			response.Groups = buildHomeGroups(sites, groups)
		}
	}

	if featured, err := svc.reader().GetFeaturedSiteList(); err != nil {
		response.CacheState["spotlight"] = models.HomeStateMissing
		response.ReasonMessages["spotlight"] = err.GetMsg()
		if len(sites) > 0 {
			response.Spotlight = buildHomeSpotlight(sites, nil, svc.clock()())
		}
	} else {
		if len(sites) > 0 {
			response.CacheState["spotlight"] = models.HomeStateReady
			response.Spotlight = buildHomeSpotlight(sites, featured, svc.clock()())
		} else {
			response.CacheState["spotlight"] = models.HomeStateMissing
			response.ReasonMessages["spotlight"] = "站点列表不可用"
		}
	}

	if ping, err := svc.reader().GetPingList(); err != nil {
		response.CacheState["ping"] = models.HomeStateMissing
		response.ReasonMessages["ping"] = err.GetMsg()
	} else {
		response.CacheState["ping"] = models.HomeStateReady
		response.Ping = ping
	}

	if saying, err := svc.reader().GetSayingService(lang); err != nil {
		response.CacheState["saying"] = models.HomeStateMissing
		response.ReasonMessages["saying"] = err.GetMsg()
	} else {
		response.CacheState["saying"] = models.HomeStateReady
		response.Saying = &saying
	}

	response.Backgrounds.Desktop = svc.reader().GetImageUrl("standard")
	response.Backgrounds.Mobile = svc.reader().GetImageUrl("mobile")
	if response.Backgrounds.Desktop == "" && response.Backgrounds.Mobile == "" {
		response.CacheState["backgrounds"] = models.HomeStateMissing
		response.ReasonMessages["backgrounds"] = "背景图不可用"
	} else {
		response.CacheState["backgrounds"] = models.HomeStateReady
	}

	if len(response.ReasonMessages) == 0 {
		response.ReasonMessages = nil
	}
	return response
}

func (svc *homeService) GetHomePing() models.HomePingResponse {
	response := models.HomePingResponse{
		SchemaVersion: models.HomeSchemaVersion,
		GeneratedAt:   svc.clock()(),
		State:         models.HomeStateMissing,
		Ping:          map[string]string{},
	}

	ping, err := svc.reader().GetPingList()
	if err != nil {
		response.ReasonMessages = []string{err.GetMsg()}
		return response
	}

	response.State = models.HomeStateReady
	response.Ping = ping
	return response
}

func (svc *homeService) GetHomeSaying(lang string) models.HomeSayingResponse {
	lang = normalizeLang(lang)
	response := models.HomeSayingResponse{
		SchemaVersion: models.HomeSchemaVersion,
		GeneratedAt:   svc.clock()(),
		State:         models.HomeStateMissing,
	}

	saying, err := svc.reader().GetSayingService(lang)
	if err != nil {
		response.ReasonMessages = []string{err.GetMsg()}
		return response
	}
	response.State = models.HomeStateReady
	response.Saying = &saying
	return response
}

func (svc *homeService) GetHomeBackgrounds() models.HomeBackgroundsResponse {
	response := models.HomeBackgroundsResponse{
		SchemaVersion: models.HomeSchemaVersion,
		GeneratedAt:   svc.clock()(),
		State:         models.HomeStateMissing,
		Backgrounds: models.HomeBackgrounds{
			Desktop: svc.reader().GetImageUrl("standard"),
			Mobile:  svc.reader().GetImageUrl("mobile"),
		},
	}
	if response.Backgrounds.Desktop == "" && response.Backgrounds.Mobile == "" {
		response.ReasonMessages = []string{"背景图不可用"}
		return response
	}
	response.State = models.HomeStateReady
	return response
}

func (svc *homeService) reader() navHomeReader {
	if svc != nil && svc.navPage != nil {
		return svc.navPage
	}
	return navservice.GetNavPageService()
}

func (svc *homeService) clock() func() time.Time {
	if svc != nil && svc.now != nil {
		return svc.now
	}
	return time.Now
}

func normalizeLang(lang string) string {
	if strings.EqualFold(strings.TrimSpace(lang), "en") {
		return "en"
	}
	return "zh"
}

func buildHomeGroups(sites []navmodels.SiteVo, groups []navmodels.GroupVo) []models.HomeGroup {
	return BuildHomeGroupsForCache(sites, groups)
}

func BuildHomeGroupsForCache(sites []navmodels.SiteVo, groups []navmodels.GroupVo) []models.HomeGroup {
	siteMap := make(map[string]navmodels.SiteVo, len(sites))
	for _, site := range sites {
		siteMap[site.ID] = site
	}

	result := make([]models.HomeGroup, 0, len(groups))
	for _, group := range groups {
		items := make([]navmodels.SiteVo, 0, len(group.Sites))
		for _, siteID := range group.Sites {
			if site, ok := siteMap[siteID]; ok {
				items = append(items, site)
			}
		}
		SortSitesForCache(items)
		siteCount := len(items)
		if siteCount > homeGroupPreviewLimit {
			items = items[:homeGroupPreviewLimit]
		}

		result = append(result, models.HomeGroup{
			ID:         group.ID,
			Name:       group.Name,
			Info:       group.Info,
			Priority:   group.Priority,
			SiteCount:  siteCount,
			HasMore:    siteCount > len(items),
			DetailPath: "/site-groups/" + group.ID,
			Sites:      items,
		})
	}

	return result
}

func buildHomeSpotlight(sites []navmodels.SiteVo, featured []navmodels.FeaturedSiteVo, now time.Time) models.HomeSpotlight {
	spotlight := models.HomeSpotlight{
		PageSize: 6,
		Featured: []navmodels.SiteVo{},
		Popular:  []navmodels.SiteVo{},
		Latest:   []navmodels.SiteVo{},
		Random:   []navmodels.SiteVo{},
	}

	if len(sites) == 0 {
		return spotlight
	}

	siteMap := make(map[string]navmodels.SiteVo, len(sites))
	for _, site := range sites {
		siteMap[site.ID] = site
	}

	for _, item := range featured {
		if site, ok := siteMap[item.SiteID]; ok {
			spotlight.Featured = append(spotlight.Featured, site)
		}
	}

	spotlight.Popular = append(spotlight.Popular, sites...)
	sort.SliceStable(spotlight.Popular, func(i, j int) bool {
		if spotlight.Popular[i].ViewCount != spotlight.Popular[j].ViewCount {
			return spotlight.Popular[i].ViewCount > spotlight.Popular[j].ViewCount
		}
		return spotlight.Popular[i].ID < spotlight.Popular[j].ID
	})

	spotlight.Latest = append(spotlight.Latest, sites...)
	sort.SliceStable(spotlight.Latest, func(i, j int) bool {
		if spotlight.Latest[i].CreateTime != spotlight.Latest[j].CreateTime {
			return spotlight.Latest[i].CreateTime > spotlight.Latest[j].CreateTime
		}
		return spotlight.Latest[i].ID > spotlight.Latest[j].ID
	})

	spotlight.Random = append(spotlight.Random, sites...)
	rng := rand.New(rand.NewSource(now.UnixNano()))
	rng.Shuffle(len(spotlight.Random), func(i, j int) {
		spotlight.Random[i], spotlight.Random[j] = spotlight.Random[j], spotlight.Random[i]
	})

	return spotlight
}

func sortSitesByWeight(items []navmodels.SiteVo) {
	SortSitesForCache(items)
}

func SortSitesForCache(items []navmodels.SiteVo) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Weight != items[j].Weight {
			return items[i].Weight > items[j].Weight
		}
		if items[i].UpdateTime != items[j].UpdateTime {
			return items[i].UpdateTime > items[j].UpdateTime
		}
		leftID, _ := strconv.ParseInt(items[i].ID, 10, 64)
		rightID, _ := strconv.ParseInt(items[j].ID, 10, 64)
		return leftID > rightID
	})
}

func GetCachedHome(lang string) models.HomeResponse {
	lang = normalizeLang(lang)
	response := models.HomeResponse{
		SchemaVersion:  models.HomeSchemaVersion,
		GeneratedAt:    time.Now(),
		CacheState:     map[string]string{"home": models.HomeStateMissing},
		ReasonMessages: map[string]string{"home": "首页缓存未命中"},
		Groups:         []models.HomeGroup{},
		Spotlight: models.HomeSpotlight{
			PageSize: 6,
			Featured: []navmodels.SiteVo{},
			Popular:  []navmodels.SiteVo{},
			Latest:   []navmodels.SiteVo{},
			Random:   []navmodels.SiteVo{},
		},
		Ping:        map[string]string{},
		Backgrounds: models.HomeBackgrounds{},
	}

	raw, err := cs.GetString(cachekeys.Home(lang))
	if err != nil || raw == "" {
		if err != nil {
			response.ReasonMessages["home"] = err.GetMsg()
		}
		return response
	}

	if unmarshalErr := sonic.Unmarshal([]byte(raw), &response); unmarshalErr != nil {
		response.ReasonMessages["home"] = "首页缓存反序列化失败"
		response.CacheState["home"] = models.HomeStateMissing
		return response
	}

	if response.CacheState == nil {
		response.CacheState = map[string]string{}
	}
	response.CacheState["home"] = models.HomeStateReady
	if len(response.ReasonMessages) == 0 {
		response.ReasonMessages = nil
	}
	return response
}
