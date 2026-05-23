package service

/*
 * @Desc: timewheel服务
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"time"

	"github.com/gofurry/gofurry-nav-collector/common/log"
	"github.com/rfyiamcool/go-timewheel"
)

var timeWheel *timewheel.TimeWheel

func InitTimeWheelOnStart() {
	StartTimeWheel()
	log.InfoFields(map[string]interface{}{
		"event": "timewheel_started",
		"slots": 1200,
		"tick":  100 * time.Millisecond,
	}, "时间轮调度器已启动")
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
			log.ErrorFields(map[string]interface{}{
				"event": "cron_remove_recovered",
			}, err)
		}
	}()
	log.InfoFields(map[string]interface{}{
		"event": "cron_remove",
	}, "定时任务已移除")
	timeWheel.Remove(task)
}

func AddCronJob(tick time.Duration, job func()) *timewheel.Task {
	task := timeWheel.AddCron(tick, job)
	log.InfoFields(map[string]interface{}{
		"event":    "cron_add",
		"interval": tick,
	}, "定时任务已注册")
	return task
}
