package service

/*
 * @Desc: timewheel服务
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"fmt"
	"time"

	"github.com/gofurry/gofurry-nav-backend/common/log"
	"github.com/rfyiamcool/go-timewheel"
)

var timeWheel *timewheel.TimeWheel

func InitTimeWheelOnStart() {
	StartTimeWheel()
	log.Info("[StartTimeWheel finish]")
}

func StartTimeWheel() {
	var err error
	timeWheel, err = timewheel.NewTimeWheel(100*time.Millisecond, 1200, timewheel.TickSafeMode())
	if err != nil {
		panic(err)
	}
	timeWheel.Start()
}

func Stop() {
	if timeWheel != nil {
		timeWheel.Stop()
	}
}

func RemoveTask(task *timewheel.Task) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	log.Info(fmt.Sprintf("remove Cronn Job: %v", task))
	timeWheel.Remove(task)
}

func AddCronJob(tick time.Duration, job func()) *timewheel.Task {
	task := timeWheel.AddCron(tick, job)
	log.Info(fmt.Sprintf("AddOrUpdate Cron Job: %v", task))
	return task
}
