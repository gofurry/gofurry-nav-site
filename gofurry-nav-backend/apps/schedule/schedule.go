package schedule

import (
	"fmt"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/schedule/task"
	"github.com/gofurry/gofurry-nav-backend/common/log"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
)

func InitScheduleOnStart() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("[InitScheduleOnStart] receive InitScheduleOnStart recover: %v", err))
		}
	}()
	log.Debug("[Schedule] init start module initialization begin...")

	go Schedule()
	go OneHourTask()
	go task.UpdateSiteViewCountCache()

	cs.AddCronJob(10*time.Minute, Schedule)
	cs.AddCronJob(1*time.Hour, OneHourTask)
	cs.AddCronJob(24*time.Hour, task.UpdateSiteViewCountCache)

	log.Debug("[Schedule] init end module initialization finished...")
}

func OneHourTask() {
	task.UpdateChangeLog()
}

func Schedule() {
	task.UpdateSiteListCache()
	task.UpdateGroupListCache()
}
