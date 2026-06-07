package task

import (
	"time"

	"github.com/bytedance/sonic"
	navDao "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/dao"
	"github.com/gofurry/gofurry-nav-backend/common/log"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
)

const siteListCacheKey = "site:list:v2"

func UpdateSiteListCache() {
	start := time.Now()
	log.Debug("[StatTask UpdateSiteListCache] start...")
	records, err := navDao.GetNavPageDao().GetSiteList() // 所有站点记录
	if err != nil {
		log.Error("[StatTask UpdateSiteListCache] GetSiteList err:", err)
	}

	if b, jsonErr := sonic.Marshal(records); jsonErr == nil {
		cs.Set(siteListCacheKey, string(b))
	}
	log.Debug("[StatTask UpdateSiteListCache] update site list finished, cost: %v", time.Since(start))
}

func UpdateGroupListCache() {
	start := time.Now()
	log.Debug("[StatTask UpdateGroupListCache] start...")
	groupRecords, err := navDao.GetNavPageDao().GetGroupList()
	if err != nil {
		log.Error("[StatTask UpdateGroupListCache] GetGroupList err:", err)
		return
	}

	mappingRecords, err := navDao.GetNavPageDao().GetGroupMapList()
	if err != nil {
		log.Error("[StatTask UpdateGroupListCache] GetGroupMapList err:", err)
		return
	}

	if b, err := sonic.Marshal(groupRecords); err == nil {
		cs.Set("group:list", string(b))
	}
	if b, err := sonic.Marshal(mappingRecords); err == nil {
		cs.Set("group:site:map", string(b))
	}
	log.Debug("[StatTask UpdateGroupListCache] update site group list finished, cost: %v", time.Since(start))
}

func UpdateFeaturedSiteListCache() {
	start := time.Now()
	log.Debug("[StatTask UpdateFeaturedSiteListCache] start...")
	records, err := navDao.GetNavPageDao().GetFeaturedSiteList()
	if err != nil {
		log.Error("[StatTask UpdateFeaturedSiteListCache] GetFeaturedSiteList err:", err)
		return
	}

	if b, jsonErr := sonic.Marshal(records); jsonErr == nil {
		cs.Set("featured-sites:list", string(b))
	}
	log.Debug("[StatTask UpdateFeaturedSiteListCache] update featured site list finished, cost: %v", time.Since(start))
}
