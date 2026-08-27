package schedule

import (
	"fmt"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/schedule/task"
	"github.com/gofurry/gofurry-nav-backend/common/log"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
)

func InitScheduleOnStart(store task.NavCacheStore, nav task.NavCacheReader, home task.HomeCacheReader, views task.SiteViewStore) error {
	log.Debug("[Schedule] init start module initialization begin...")

	refreshCaches := func() { Schedule(store, nav, home) }
	flushViews := func() { task.UpdateSiteViewCountCache(views) }
	go refreshCaches()
	go flushViews()

	if err := cs.AddIntervalJob("nav-derived-caches", 10*time.Minute, refreshCaches); err != nil {
		return fmt.Errorf("initialize Nav cache schedule: %w", err)
	}
	if err := cs.AddIntervalJob("nav-view-count", 24*time.Hour, flushViews); err != nil {
		return fmt.Errorf("initialize Nav view schedule: %w", err)
	}

	log.Debug("[Schedule] init end module initialization finished...")
	return nil
}

func Schedule(store task.NavCacheStore, nav task.NavCacheReader, home task.HomeCacheReader) {
	task.UpdateSiteListCache(store)
	task.UpdateGroupListCache(store)
	task.UpdateFeaturedSiteListCache(store)
	task.UpdateDerivedNavCaches(nav, home)
}
