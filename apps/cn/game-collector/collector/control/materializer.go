package control

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofurry/gofurry-game-collector/common/log"
	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
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
	queries := gamesqlc.New(m.pool)
	rows, err := queries.ListGameCollectionSchedules(ctx)
	if err != nil {
		return fmt.Errorf("list Game collection schedules: %w", err)
	}
	m.recordVersions(rows)
	clock, err := queries.GameCollectionClock(ctx)
	if err != nil || !clock.Valid {
		return fmt.Errorf("read Game control-plane clock: %w", err)
	}
	now := clock.Time
	for _, schedule := range rows {
		if !schedule.Enabled {
			continue
		}
		if _, ok := capability(schedule.JobKey); !ok {
			log.Warn("ignoring unknown Game collection schedule job_key=", schedule.JobKey)
			continue
		}
		if err := m.materialize(ctx, schedule.ID, now); err != nil {
			return err
		}
	}
	return nil
}

func (m *Materializer) recordVersions(rows []gamesqlc.GfgCollectionSchedule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		seen[row.JobKey] = struct{}{}
		if previous, ok := m.versions[row.JobKey]; !ok || previous != row.Version {
			m.versions[row.JobKey] = row.Version
			log.Info("Game schedule loaded, job_key=", row.JobKey, " version=", row.Version, " enabled=", row.Enabled)
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
	queries := gamesqlc.New(tx)
	schedule, err := queries.GetGameCollectionScheduleForUpdate(ctx, scheduleID)
	if err != nil {
		return err
	}
	if !schedule.Enabled {
		return tx.Commit(ctx)
	}
	item, ok := capability(schedule.JobKey)
	if !ok {
		return fmt.Errorf("unknown Game capability %q", schedule.JobKey)
	}
	due, next, err := dueSlots(schedule, now)
	if err != nil {
		return err
	}
	grace := time.Duration(schedule.MisfireGraceSeconds) * time.Second
	for index, slot := range due {
		status, trigger, priority := classifyDueSlot(schedule.MisfirePolicy, grace, slot, now, index, len(due), schedule.Priority, PriorityCatchUp)
		if status == "queued" && item.PointInTime {
			active, activeErr := queries.GameCollectionLaneActive(ctx, schedule.ConcurrencyKey)
			if activeErr != nil {
				return activeErr
			}
			if active {
				status = "skipped"
			}
		}
		scheduleIDValue := schedule.ID
		scheduleVersion := schedule.Version
		if _, err := queries.InsertGameScheduledJob(ctx, gamesqlc.InsertGameScheduledJobParams{
			ScheduleID:      &scheduleIDValue,
			ScheduleVersion: &scheduleVersion,
			JobKey:          schedule.JobKey,
			Trigger:         trigger,
			Tasks:           append([]string(nil), item.Tasks...),
			Priority:        priority,
			ConcurrencyKey:  schedule.ConcurrencyKey,
			ScheduledFor:    timestamp(slot),
			Status:          status,
			RequestedBy:     "scheduler",
		}); err != nil {
			return fmt.Errorf("materialize %s slot %s: %w", schedule.JobKey, slot.Format(time.RFC3339), err)
		}
	}
	last := schedule.LastMaterializedFor
	if len(due) > 0 {
		last = timestamp(due[len(due)-1])
	}
	if err := queries.UpdateGameCollectionScheduleCursor(ctx, gamesqlc.UpdateGameCollectionScheduleCursorParams{
		LastMaterializedFor: last,
		NextScheduledFor:    timestamp(next),
		ID:                  schedule.ID,
		Version:             schedule.Version,
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
