package task

import (
	"strings"

	navModels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	navDao "github.com/gofurry/gofurry-nav-backend/apps/nav/sitePage/dao"
	"github.com/gofurry/gofurry-nav-backend/common/log"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
	"github.com/gofurry/gofurry-nav-backend/common/util"
)

const siteViewCountPrefix = "site:view:count:"

func UpdateSiteViewCountCache() {
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

		if dbErr := navDao.GetSitePageDao().Gm.Table(navModels.TableNameGfnSite).Where("id = ?", siteID).Update("view_count", viewCount).Error; dbErr != nil {
			log.Error("[UpdateSiteViewCountCache] update site view count err:", dbErr)
		}
	}
}
