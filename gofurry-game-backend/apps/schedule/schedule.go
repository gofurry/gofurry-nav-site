package schedule

import (
	"fmt"
	"time"

	"github.com/GoFurry/gofurry-game-backend/apps/schedule/task"
	"github.com/GoFurry/gofurry-game-backend/common/log"
	cs "github.com/GoFurry/gofurry-game-backend/common/service"
)

func InitScheduleOnStart() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("[InitScheduleOnStart] recover: %v", err))
		}
	}()

	log.Info("[Schedule] init start")

	go ScheduleByTenMinutes()
	go ScheduleByOneHour()
	go ScheduleByHalfDay()
	go task.UpdateGameViewCountCache()

	cs.AddCronJob(10*time.Minute, ScheduleByTenMinutes)
	cs.AddCronJob(1*time.Hour, ScheduleByOneHour)
	cs.AddCronJob(12*time.Hour, ScheduleByHalfDay)
	cs.AddCronJob(24*time.Hour, task.UpdateGameViewCountCache)

	log.Info("[Schedule] init done")
}

func ScheduleByTenMinutes() {
	task.UpdateMainInfoCache()
}

func ScheduleByOneHour() {
	task.UpdateGamePanelCache()
	task.UpdateGameNewsCache()
	task.UpdateGameCreatorCache()
	task.UpdateMoreGameNewsCache()
	task.UpdatePrizeWinner()
}

func ScheduleByHalfDay() {
}
