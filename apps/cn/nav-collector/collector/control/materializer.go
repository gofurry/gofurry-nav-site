package control

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofurry/gofurry-nav-collector/common/log"
	navsqlc "github.com/gofurry/gofurry-nav-collector/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Materializer struct {
	pool     *pgxpool.Pool
	versions map[string]int64
	mu       sync.Mutex
}

func NewMaterializer(pool *pgxpool.Pool) *Materializer {
	return &Materializer{pool: pool, versions: make(map[string]int64)}
}

func (m *Materializer) Reconcile(ctx context.Context) error {
	queries := navsqlc.New(m.pool)
	rows, err := queries.ListNavCollectionSchedules(ctx)
	if err != nil {
		return fmt.Errorf("list Nav collection schedules: %w", err)
	}
	m.recordVersions(rows)
	clock, err := queries.NavCollectionClock(ctx)
	if err != nil || !clock.Valid {
		return fmt.Errorf("read Nav control-plane clock: %w", err)
	}
	now := clock.Time
	for _, schedule := range rows {
		if !schedule.Enabled {
			continue
		}
		if _, ok := capability(schedule.JobKey); !ok {
			log.Warn("ignoring unknown Nav collection schedule job_key=", schedule.JobKey)
			continue
		}
		if err := m.materialize(ctx, schedule.ID, now); err != nil {
			return err
		}
	}
	return nil
}

func (m *Materializer) recordVersions(rows []navsqlc.GfnCollectionSchedule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		seen[row.JobKey] = struct{}{}
		if version, ok := m.versions[row.JobKey]; !ok || version != row.Version {
			m.versions[row.JobKey] = row.Version
			log.InfoFields(map[string]interface{}{"job_key": row.JobKey, "version": row.Version, "enabled": row.Enabled}, "Nav schedule loaded")
		}
	}
	for key := range m.versions {
		if _, ok := seen[key]; !ok {
			delete(m.versions, key)
		}
	}
}

func (m *Materializer) materialize(ctx context.Context, scheduleID int64, now time.Time) error {
	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	queries := navsqlc.New(tx)
	schedule, err := queries.GetNavCollectionScheduleForUpdate(ctx, scheduleID)
	if err != nil {
		return err
	}
	if !schedule.Enabled {
		return tx.Commit(ctx)
	}
	item, ok := capability(schedule.JobKey)
	if !ok {
		return fmt.Errorf("unknown Nav capability %q", schedule.JobKey)
	}
	due, next, err := dueSlots(schedule, now)
	if err != nil {
		return err
	}
	grace := time.Duration(schedule.MisfireGraceSeconds) * time.Second
	for index, slot := range due {
		status, trigger, priority := classifyDueSlot(schedule.MisfirePolicy, grace, slot, now, index, len(due), schedule.Priority, priorityCatchUp)
		if status == "queued" && item.PointInTime {
			active, activeErr := queries.NavCollectionLaneActive(ctx, schedule.ConcurrencyKey)
			if activeErr != nil {
				return activeErr
			}
			if active {
				status = "skipped"
			}
		}
		id, version := schedule.ID, schedule.Version
		if _, err := queries.InsertNavScheduledJob(ctx, navsqlc.InsertNavScheduledJobParams{
			ScheduleID: &id, ScheduleVersion: &version, JobKey: schedule.JobKey,
			Trigger: trigger, Tasks: []string{item.Protocol}, Priority: priority,
			ConcurrencyKey: schedule.ConcurrencyKey, ScheduledFor: timestamp(slot),
			Status: status, RequestedBy: "scheduler",
		}); err != nil {
			return err
		}
	}
	last := schedule.LastMaterializedFor
	if len(due) > 0 {
		last = timestamp(due[len(due)-1])
	}
	if err := queries.UpdateNavCollectionScheduleCursor(ctx, navsqlc.UpdateNavCollectionScheduleCursorParams{
		LastMaterializedFor: last, NextScheduledFor: timestamp(next), ID: schedule.ID, Version: schedule.Version,
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func classifyDueSlot(policy string, grace time.Duration, slot, now time.Time, index, total int, scheduledPriority, catchUpPriority int32) (status, trigger string, priority int32) {
	status, trigger, priority = "queued", "scheduled", scheduledPriority
	if now.Sub(slot) <= grace {
		return
	}
	if policy == "skip" {
		status = "missed"
		return
	}
	if policy == "catch_up_once" {
		if index < total-1 {
			status = "missed"
			return
		}
		trigger, priority = "startup_catchup", catchUpPriority
	}
	return
}

func timestamp(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: !value.IsZero()}
}
