package service

import (
	"fmt"
	"time"

	"github.com/GoFurry/gofurry-nav-backend/common/log"
	cs "github.com/GoFurry/gofurry-nav-backend/common/service"
	"github.com/GoFurry/gofurry-nav-backend/common/util"
)

const (
	siteViewCountPrefix = "site:view:count:"
	siteViewDailyPrefix = "site:view:daily:"
	siteViewDailyTTL    = 48 * time.Hour
)

func (svc *sitePageService) touchSiteViewCount(siteID int64, dbCount int64, clientIP string) int64 {
	countKey := siteViewCountPrefix + util.Int642String(siteID)
	current := dbCount

	if countStr, err := cs.GetString(countKey); err == nil {
		if countStr != "" {
			if parsed, parseErr := util.String2Int64(countStr); parseErr == nil {
				current = parsed
			}
		} else if cs.SetNX(countKey, util.Int642String(current), 0) {
			// current already seeded from database
		} else if latest, latestErr := cs.GetString(countKey); latestErr == nil && latest != "" {
			if parsed, parseErr := util.String2Int64(latest); parseErr == nil {
				current = parsed
			}
		}
	} else {
		log.Warn(fmt.Sprintf("[site-view-count] get redis count failed for site %d: %v", siteID, err))
	}

	if clientIP == "" {
		return current
	}

	dailyKey := fmt.Sprintf("%s%d:%s:%s", siteViewDailyPrefix, siteID, time.Now().Format("2006-01-02"), util.CreateMD5(clientIP))
	if cs.SetNX(dailyKey, "1", siteViewDailyTTL) {
		cs.Incr(countKey)
		if countStr, err := cs.GetString(countKey); err == nil && countStr != "" {
			if parsed, parseErr := util.String2Int64(countStr); parseErr == nil {
				current = parsed
			} else {
				current++
			}
		} else {
			current++
		}
	}

	return current
}
