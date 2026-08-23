package schedule

import (
	"fmt"
	"time"

	v2dao "github.com/gofurry/gofurry-game-backend/apps/game/v2/dao"
	v2service "github.com/gofurry/gofurry-game-backend/apps/game/v2/service"
	prizedao "github.com/gofurry/gofurry-game-backend/apps/prize/dao"
	"github.com/gofurry/gofurry-game-backend/apps/schedule/task"
	"github.com/gofurry/gofurry-game-backend/common/log"
	cs "github.com/gofurry/gofurry-game-backend/common/service"
)

func InitScheduleOnStart(readDAO *v2dao.ReadModelDAO, viewService *v2service.GameViewService, prizeDAO *prizedao.PrizeDAO) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("[InitScheduleOnStart] recover: %v", err))
		}
	}()

	log.Info("[Schedule] init start")

	go ScheduleByOneHour(prizeDAO)
	go ScheduleByHalfDay()
	go task.UpdateGameViewCountCache(viewService)
	task.RefreshGameHomeCache(readDAO)

	cs.AddCronJob(1*time.Hour, func() { ScheduleByOneHour(prizeDAO) })
	cs.AddCronJob(1*time.Hour, func() { task.RefreshGameHomeCache(readDAO) })
	cs.AddCronJob(12*time.Hour, ScheduleByHalfDay)
	cs.AddCronJob(24*time.Hour, func() { task.UpdateGameViewCountCache(viewService) })

	log.Info("[Schedule] init done")
}

func ScheduleByOneHour(prizeDAO *prizedao.PrizeDAO) {
	task.UpdatePrizeWinner(prizeDAO)
}

func ScheduleByHalfDay() {
}
