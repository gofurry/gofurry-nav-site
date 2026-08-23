package controller

import (
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/stats/models"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/stats/service"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/util"
)

type pageViewTracker interface {
	TouchPageView(page string, clientIP string) models.PageViewResponse
}

type statsApi struct{}

var StatsApi *statsApi

var (
	pageViewTrackerMu      sync.RWMutex
	pageViewTrackerForTest pageViewTracker
)

func init() {
	StatsApi = &statsApi{}
}

func (api statsApi) TouchPageView(c fiber.Ctx) error {
	page := c.Query("page")
	if page == "" {
		return common.NewResponse(c).Error("page 参数不能为空")
	}
	return common.NewResponse(c).SuccessWithData(currentPageViewTracker().TouchPageView(page, util.GetClientIP(c)))
}

func currentPageViewTracker() pageViewTracker {
	pageViewTrackerMu.RLock()
	tracker := pageViewTrackerForTest
	pageViewTrackerMu.RUnlock()
	if tracker != nil {
		return tracker
	}
	return service.GetPageViewService()
}

func setPageViewTrackerForTest(tracker pageViewTracker) func() {
	pageViewTrackerMu.Lock()
	previous := pageViewTrackerForTest
	pageViewTrackerForTest = tracker
	pageViewTrackerMu.Unlock()
	return func() {
		pageViewTrackerMu.Lock()
		pageViewTrackerForTest = previous
		pageViewTrackerMu.Unlock()
	}
}
