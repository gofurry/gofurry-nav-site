package service

import (
	"strings"
	"sync"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/home/models"
	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	navservice "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/service"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type navHomeReader interface {
	GetSiteList(lang string) ([]navmodels.SiteVo, common.GFError)
	GetGroupList(lang string) ([]navmodels.GroupVo, common.GFError)
	GetPingList() (map[string]string, common.GFError)
	GetSayingService() (navmodels.SayingModel, common.GFError)
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
		Sites:          []navmodels.SiteVo{},
		Groups:         []models.HomeGroup{},
		Ping:           map[string]string{},
		Backgrounds:    models.HomeBackgrounds{},
	}

	if sites, err := svc.reader().GetSiteList(lang); err != nil {
		response.CacheState["sites"] = models.HomeStateMissing
		response.ReasonMessages["sites"] = err.GetMsg()
	} else {
		response.CacheState["sites"] = models.HomeStateReady
		response.Sites = sites
	}

	if groups, err := svc.reader().GetGroupList(lang); err != nil {
		response.CacheState["groups"] = models.HomeStateMissing
		response.ReasonMessages["groups"] = err.GetMsg()
	} else {
		response.CacheState["groups"] = models.HomeStateReady
		response.Groups = buildHomeGroups(response.Sites, groups)
	}

	if ping, err := svc.reader().GetPingList(); err != nil {
		response.CacheState["ping"] = models.HomeStateMissing
		response.ReasonMessages["ping"] = err.GetMsg()
	} else {
		response.CacheState["ping"] = models.HomeStateReady
		response.Ping = ping
	}

	if saying, err := svc.reader().GetSayingService(); err != nil {
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

func (svc *homeService) GetHomeSaying() models.HomeSayingResponse {
	response := models.HomeSayingResponse{
		SchemaVersion: models.HomeSchemaVersion,
		GeneratedAt:   svc.clock()(),
		State:         models.HomeStateMissing,
	}

	saying, err := svc.reader().GetSayingService()
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

		result = append(result, models.HomeGroup{
			ID:       group.ID,
			Name:     group.Name,
			Info:     group.Info,
			Priority: group.Priority,
			Sites:    items,
		})
	}

	return result
}
