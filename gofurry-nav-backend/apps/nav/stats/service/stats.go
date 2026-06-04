package service

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/stats/models"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
	"github.com/gofurry/gofurry-nav-backend/common/util"
)

const (
	pageViewCountPrefix = "nav:page:view:count:"
	pageViewDailyPrefix = "nav:page:view:daily:"
	pageViewDailyTTL    = 48 * time.Hour
)

type pageViewService struct {
	now func() time.Time
}

var (
	pageViewSingleton = &pageViewService{}
	pageViewMu        sync.Mutex
)

func GetPageViewService() *pageViewService {
	pageViewMu.Lock()
	defer pageViewMu.Unlock()
	if pageViewSingleton.now == nil {
		pageViewSingleton.now = time.Now
	}
	return pageViewSingleton
}

func newPageViewService(now func() time.Time) *pageViewService {
	return &pageViewService{now: now}
}

func (svc *pageViewService) TouchPageView(page string, clientIP string) models.PageViewResponse {
	page = normalizePageKey(page)
	now := svc.clock()()
	response := models.PageViewResponse{
		SchemaVersion: models.PageViewSchemaVersion,
		GeneratedAt:   now,
		State:         models.PageViewStateReady,
		Page:          page,
		ViewCount:     currentPageViewCount(page),
	}

	if page == "" || clientIP == "" {
		return response
	}

	dailyKey := fmt.Sprintf("%s%s:%s:%s", pageViewDailyPrefix, page, now.Format("2006-01-02"), util.CreateMD5(clientIP))
	if cs.SetNX(dailyKey, "1", pageViewDailyTTL) {
		_ = cs.Incr(pageViewCountKey(page))
		response.ViewCount = currentPageViewCount(page)
	}
	return response
}

func (svc *pageViewService) clock() func() time.Time {
	if svc != nil && svc.now != nil {
		return svc.now
	}
	return time.Now
}

func currentPageViewCount(page string) int64 {
	countStr, err := cs.GetString(pageViewCountKey(page))
	if err != nil || countStr == "" {
		return 0
	}
	count, parseErr := util.String2Int64(countStr)
	if parseErr != nil {
		return 0
	}
	return count
}

func pageViewCountKey(page string) string {
	return pageViewCountPrefix + page
}

func normalizePageKey(page string) string {
	page = strings.ToLower(strings.TrimSpace(page))
	page = strings.ReplaceAll(page, " ", "_")
	page = strings.ReplaceAll(page, "/", "_")
	page = strings.ReplaceAll(page, "\\", "_")
	if page == "" {
		return ""
	}
	var cleaned strings.Builder
	for _, r := range page {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			cleaned.WriteRune(r)
		}
	}
	result := cleaned.String()
	if len(result) > 32 {
		return result[:32]
	}
	return result
}
