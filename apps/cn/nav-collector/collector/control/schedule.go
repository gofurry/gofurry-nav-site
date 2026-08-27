package control

import (
	"fmt"
	"time"

	navsqlc "github.com/gofurry/gofurry-nav-collector/internal/db/nav/sqlc"
	"github.com/robfig/cron/v3"
)

const maxMaterializedSlots = 10000

func dueSlots(schedule navsqlc.GfnCollectionSchedule, now time.Time) ([]time.Time, time.Time, error) {
	cursor := schedule.EffectiveFrom.Time
	if schedule.LastMaterializedFor.Valid && schedule.LastMaterializedFor.Time.After(cursor) {
		cursor = schedule.LastMaterializedFor.Time
	}
	next, err := nextSlot(schedule, cursor)
	if err != nil {
		return nil, time.Time{}, err
	}
	due := make([]time.Time, 0)
	for !next.After(now) {
		due = append(due, next)
		if len(due) >= maxMaterializedSlots {
			return nil, time.Time{}, fmt.Errorf("schedule %s has more than %d missed slots", schedule.JobKey, maxMaterializedSlots)
		}
		next, err = nextSlot(schedule, next)
		if err != nil {
			return nil, time.Time{}, err
		}
	}
	return due, next, nil
}

func nextSlot(schedule navsqlc.GfnCollectionSchedule, after time.Time) (time.Time, error) {
	if schedule.ScheduleKind == "interval" {
		if schedule.IntervalSeconds == nil || *schedule.IntervalSeconds <= 0 || !schedule.AnchorAt.Valid {
			return time.Time{}, fmt.Errorf("invalid interval schedule %s", schedule.JobKey)
		}
		step := time.Duration(*schedule.IntervalSeconds) * time.Second
		anchor := schedule.AnchorAt.Time
		if after.Before(anchor) {
			return anchor, nil
		}
		return anchor.Add((after.Sub(anchor)/step + 1) * step), nil
	}
	if schedule.ScheduleKind == "cron" && schedule.CronExpression != nil {
		location, err := time.LoadLocation(schedule.Timezone)
		if err != nil {
			return time.Time{}, err
		}
		parsed, err := cron.ParseStandard(*schedule.CronExpression)
		if err != nil {
			return time.Time{}, err
		}
		return parsed.Next(after.In(location)), nil
	}
	return time.Time{}, fmt.Errorf("invalid schedule %s", schedule.JobKey)
}
