package schedule

import (
	"fmt"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/schedule/task"
	"github.com/gofurry/gofurry-nav-backend/common/log"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
)

func InitScheduleOnStart(store task.NavCacheStore, nav task.NavCacheReader, home task.HomeCacheReader, views task.SiteViewStore) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("[InitScheduleOnStart] receive InitScheduleOnStart recover: %v", err))
		}
	}()
	log.Debug("[Schedule] init start module initialization begin...")

	refreshCaches := func() { Schedule(store, nav, home) }
	flushViews := func() { task.UpdateSiteViewCountCache(views) }
	go refreshCaches()
	go flushViews()

	cs.AddCronJob(10*time.Minute, refreshCaches)
	cs.AddCronJob(24*time.Hour, flushViews)

	log.Debug("[Schedule] init end module initialization finished...")
}

func Schedule(store task.NavCacheStore, nav task.NavCacheReader, home task.HomeCacheReader) {
	task.UpdateSiteListCache(store)
	task.UpdateGroupListCache(store)
	task.UpdateFeaturedSiteListCache(store)
	task.UpdateDerivedNavCaches(nav, home)
}
