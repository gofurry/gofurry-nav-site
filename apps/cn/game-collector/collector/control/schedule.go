package control

import (
	"fmt"
	"time"

	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
	"github.com/robfig/cron/v3"
)

const maxMaterializedSlots = 10000

func dueSlots(schedule gamesqlc.GfgCollectionSchedule, now time.Time) ([]time.Time, time.Time, error) {
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

func nextSlot(schedule gamesqlc.GfgCollectionSchedule, after time.Time) (time.Time, error) {
	switch schedule.ScheduleKind {
	case "interval":
		if schedule.IntervalSeconds == nil || *schedule.IntervalSeconds <= 0 || !schedule.AnchorAt.Valid {
			return time.Time{}, fmt.Errorf("invalid interval schedule %s", schedule.JobKey)
		}
		step := time.Duration(*schedule.IntervalSeconds) * time.Second
		anchor := schedule.AnchorAt.Time
		if after.Before(anchor) {
			return anchor, nil
		}
		elapsed := after.Sub(anchor)
		return anchor.Add((elapsed/step + 1) * step), nil
	case "cron":
		if schedule.CronExpression == nil {
			return time.Time{}, fmt.Errorf("cron schedule %s has no expression", schedule.JobKey)
		}
		location, err := time.LoadLocation(schedule.Timezone)
		if err != nil {
			return time.Time{}, fmt.Errorf("load timezone %s: %w", schedule.Timezone, err)
		}
		parsed, err := cron.ParseStandard(*schedule.CronExpression)
		if err != nil {
			return time.Time{}, fmt.Errorf("parse schedule %s: %w", schedule.JobKey, err)
		}
		return parsed.Next(after.In(location)), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported schedule kind %q", schedule.ScheduleKind)
	}
}
