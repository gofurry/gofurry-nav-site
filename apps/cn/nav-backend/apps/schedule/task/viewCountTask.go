package task

import (
	"strings"

	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/log"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
	"github.com/gofurry/gofurry-nav-backend/common/util"
)

const siteViewCountPrefix = "site:view:count:"

type SiteViewStore interface {
	UpdateViewCount(siteID int64, viewCount int64) common.GFError
}

func UpdateSiteViewCountCache(store SiteViewStore) {
	keys, err := cs.FindByPrefix(siteViewCountPrefix)
	if err != nil {
		log.Error("[UpdateSiteViewCountCache] find redis keys err:", err)
		return
	}

	for _, key := range keys {
		idStr := strings.TrimPrefix(key, siteViewCountPrefix)
		siteID, parseErr := util.String2Int64(idStr)
		if parseErr != nil {
			continue
		}

		countStr, getErr := cs.GetString(key)
		if getErr != nil || countStr == "" {
			continue
		}

		viewCount, parseCountErr := util.String2Int64(countStr)
		if parseCountErr != nil {
			continue
		}

		if dbErr := store.UpdateViewCount(siteID, viewCount); dbErr != nil {
			log.Error("[UpdateSiteViewCountCache] update site view count err:", dbErr)
		}
	}
}
