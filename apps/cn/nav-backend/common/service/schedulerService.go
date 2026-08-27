package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/gofurry/gofurry-nav-backend/common/log"
)

var (
	schedulerMu sync.Mutex
	scheduler   gocron.Scheduler
)

func InitSchedulerOnStart() error {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()
	if scheduler != nil {
		return nil
	}
	created, err := gocron.NewScheduler()
	if err != nil {
		return fmt.Errorf("create backend scheduler: %w", err)
	}
	created.Start()
	scheduler = created
	log.Info("Backend scheduler started")
	return nil
}

func AddIntervalJob(name string, interval time.Duration, job func()) error {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()
	if scheduler == nil {
		return fmt.Errorf("backend scheduler is not initialized")
	}
	_, err := scheduler.NewJob(
		gocron.DurationJob(interval),
		gocron.NewTask(job),
		gocron.WithName(name),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		return fmt.Errorf("register backend job %s: %w", name, err)
	}
	return nil
}

func Stop() {
	schedulerMu.Lock()
	current := scheduler
	scheduler = nil
	schedulerMu.Unlock()
	if current != nil {
		if err := current.Shutdown(); err != nil {
			log.Error("stop backend scheduler: ", err)
		}
	}
}
