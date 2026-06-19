package schedule

import (
	"fmt"
	"time"

	"github.com/gofurry/gofurry-game-backend/apps/schedule/task"
	"github.com/gofurry/gofurry-game-backend/common/log"
	cs "github.com/gofurry/gofurry-game-backend/common/service"
)

func InitScheduleOnStart() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("[InitScheduleOnStart] recover: %v", err))
		}
	}()

	log.Info("[Schedule] init start")

	go ScheduleByOneHour()
	go ScheduleByHalfDay()
	go task.UpdateGameViewCountCache()
	task.RefreshGameHomeCache()

	cs.AddCronJob(1*time.Hour, ScheduleByOneHour)
	cs.AddCronJob(1*time.Hour, task.RefreshGameHomeCache)
	cs.AddCronJob(12*time.Hour, ScheduleByHalfDay)
	cs.AddCronJob(24*time.Hour, task.UpdateGameViewCountCache)

	log.Info("[Schedule] init done")
}

func ScheduleByOneHour() {
	task.UpdatePrizeWinner()
}

func ScheduleByHalfDay() {
}
