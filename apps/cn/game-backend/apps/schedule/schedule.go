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

func InitScheduleOnStart(readDAO *v2dao.ReadModelDAO, viewService *v2service.GameViewService, prizeDAO *prizedao.PrizeDAO) error {
	log.Info("[Schedule] init start")

	go ScheduleByOneHour(prizeDAO)
	go ScheduleByHalfDay()
	go task.UpdateGameViewCountCache(viewService)
	task.RefreshGameHomeCache(readDAO)

	jobs := []struct {
		name     string
		interval time.Duration
		run      func()
	}{
		{"game-prize-winner", time.Hour, func() { ScheduleByOneHour(prizeDAO) }},
		{"game-home-cache", time.Hour, func() { task.RefreshGameHomeCache(readDAO) }},
		{"game-half-day", 12 * time.Hour, ScheduleByHalfDay},
		{"game-view-count", 24 * time.Hour, func() { task.UpdateGameViewCountCache(viewService) }},
	}
	for _, job := range jobs {
		if err := cs.AddIntervalJob(job.name, job.interval, job.run); err != nil {
			return fmt.Errorf("initialize Game backend schedules: %w", err)
		}
	}

	log.Info("[Schedule] init done")
	return nil
}

func ScheduleByOneHour(prizeDAO *prizedao.PrizeDAO) {
	task.UpdatePrizeWinner(prizeDAO)
}

func ScheduleByHalfDay() {
}
