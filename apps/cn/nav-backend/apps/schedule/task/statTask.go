package task

import (
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/cachekeys"
	homemodels "github.com/gofurry/gofurry-nav-backend/apps/nav/home/models"
	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	siteGroupService "github.com/gofurry/gofurry-nav-backend/apps/nav/sitegroup/service"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/log"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
)

type NavCacheStore interface {
	GetSiteList() ([]navmodels.GfnSite, common.GFError)
	GetGroupList() ([]navmodels.GfnSiteGroup, common.GFError)
	GetGroupMapList() ([]navmodels.GfnSiteGroupMap, common.GFError)
	GetFeaturedSiteList() ([]navmodels.GfnFeaturedSite, common.GFError)
}

type NavCacheReader interface {
	GetSiteList(lang string) ([]navmodels.SiteVo, common.GFError)
	GetGroupList(lang string) ([]navmodels.GroupVo, common.GFError)
}

type HomeCacheReader interface {
	GetHome(lang string) homemodels.HomeResponse
}

func UpdateSiteListCache(store NavCacheStore) {
	start := time.Now()
	log.Debug("[StatTask UpdateSiteListCache] start...")
	records, err := store.GetSiteList() // 所有站点记录
	if err != nil {
		log.Error("[StatTask UpdateSiteListCache] GetSiteList err:", err)
		return
	}
	if len(records) == 0 {
		log.Warn("[StatTask UpdateSiteListCache] skip empty site list cache write")
		return
	}

	if b, jsonErr := sonic.Marshal(records); jsonErr == nil {
		cs.Set(cachekeys.SiteListV2, string(b))
	}
	log.Debug("[StatTask UpdateSiteListCache] update site list finished, cost: %v", time.Since(start))
}

func UpdateGroupListCache(store NavCacheStore) {
	start := time.Now()
	log.Debug("[StatTask UpdateGroupListCache] start...")
	groupRecords, err := store.GetGroupList()
	if err != nil {
		log.Error("[StatTask UpdateGroupListCache] GetGroupList err:", err)
		return
	}
	if len(groupRecords) == 0 {
		log.Warn("[StatTask UpdateGroupListCache] skip empty group list cache write")
		return
	}

	mappingRecords, err := store.GetGroupMapList()
	if err != nil {
		log.Error("[StatTask UpdateGroupListCache] GetGroupMapList err:", err)
		return
	}
	if len(mappingRecords) == 0 {
		log.Warn("[StatTask UpdateGroupListCache] skip empty group map cache write")
		return
	}

	if b, err := sonic.Marshal(groupRecords); err == nil {
		cs.Set(cachekeys.GroupList, string(b))
	}
	if b, err := sonic.Marshal(mappingRecords); err == nil {
		cs.Set(cachekeys.GroupSiteMap, string(b))
	}
	log.Debug("[StatTask UpdateGroupListCache] update site group list finished, cost: %v", time.Since(start))
}

func UpdateFeaturedSiteListCache(store NavCacheStore) {
	start := time.Now()
	log.Debug("[StatTask UpdateFeaturedSiteListCache] start...")
	records, err := store.GetFeaturedSiteList()
	if err != nil {
		log.Error("[StatTask UpdateFeaturedSiteListCache] GetFeaturedSiteList err:", err)
		return
	}

	if b, jsonErr := sonic.Marshal(records); jsonErr == nil {
		cs.Set(cachekeys.FeaturedSiteList, string(b))
	}
	log.Debug("[StatTask UpdateFeaturedSiteListCache] update featured site list finished, cost: %v", time.Since(start))
}

func UpdateDerivedNavCaches(nav NavCacheReader, home HomeCacheReader) {
	start := time.Now()
	log.Debug("[StatTask UpdateDerivedNavCaches] start...")

	for _, lang := range []string{"zh", "en"} {
		sites, err := nav.GetSiteList(lang)
		if err != nil {
			log.Error("[StatTask UpdateDerivedNavCaches] GetSiteList err:", err)
			continue
		}
		if len(sites) == 0 {
			log.Warn("[StatTask UpdateDerivedNavCaches] skip derived cache write: empty site list, lang=", lang)
			continue
		}

		if b, jsonErr := sonic.Marshal(sites); jsonErr == nil {
			cs.Set(cachekeys.SiteDirectory(lang), string(b))
		}

		groups, err := nav.GetGroupList(lang)
		if err != nil {
			log.Error("[StatTask UpdateDerivedNavCaches] GetGroupList err:", err)
			continue
		}
		if len(groups) == 0 {
			log.Warn("[StatTask UpdateDerivedNavCaches] skip derived cache write: empty group list, lang=", lang)
			continue
		}

		homePayload := home.GetHome(lang)
		if len(homePayload.Groups) == 0 {
			log.Warn("[StatTask UpdateDerivedNavCaches] skip home cache write: empty home groups, lang=", lang)
			continue
		}
		homePayload.Sites = nil
		if b, jsonErr := sonic.Marshal(homePayload); jsonErr == nil {
			cs.Set(cachekeys.Home(lang), string(b))
		}

		if err := cs.DelByPrefix(cachekeys.SiteGroupPrefix + lang + ":"); err != nil {
			log.Error("[StatTask UpdateDerivedNavCaches] DelByPrefix err:", err)
			continue
		}

		for groupID, payload := range siteGroupService.BuildAllSiteGroupCaches(sites, groups) {
			if b, jsonErr := sonic.Marshal(payload); jsonErr == nil {
				cs.Set(cachekeys.SiteGroup(lang, groupID), string(b))
			}
		}
	}

	log.Debug("[StatTask UpdateDerivedNavCaches] update derived nav caches finished, cost: %v", time.Since(start))
}
