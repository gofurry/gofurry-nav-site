package control

import (
	"testing"
	"time"

	navsqlc "github.com/gofurry/gofurry-nav-collector/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestDueSlotsAnchoredIntervalAndDurableCursor(t *testing.T) {
	anchor := time.Date(2026, 1, 1, 0, 5, 0, 0, time.UTC)
	step := int64(300)
	schedule := navsqlc.GfnCollectionSchedule{
		JobKey: "nav.ping", ScheduleKind: "interval", IntervalSeconds: &step,
		AnchorAt:      pgtype.Timestamptz{Time: anchor, Valid: true},
		EffectiveFrom: pgtype.Timestamptz{Time: anchor.Add(time.Minute), Valid: true},
	}
	due, next, err := dueSlots(schedule, anchor.Add(16*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(due) != 3 || !due[0].Equal(anchor.Add(5*time.Minute)) || !due[2].Equal(anchor.Add(15*time.Minute)) {
		t.Fatalf("due = %v", due)
	}
	if !next.Equal(anchor.Add(20 * time.Minute)) {
		t.Fatalf("next = %s", next)
	}

	schedule.LastMaterializedFor = pgtype.Timestamptz{Time: anchor.Add(10 * time.Minute), Valid: true}
	due, next, err = dueSlots(schedule, anchor.Add(16*time.Minute))
	if err != nil || len(due) != 1 || !due[0].Equal(anchor.Add(15*time.Minute)) || !next.Equal(anchor.Add(20*time.Minute)) {
		t.Fatalf("restart slots = %v next=%s err=%v", due, next, err)
	}
}

func TestFinalNavRunStatus(t *testing.T) {
	if got := finalNavRunStatus(navProgress{Expected: 1, Attempted: 1, Success: 1}, 1, false); got != "success" {
		t.Fatalf("success status = %s", got)
	}
	if got := finalNavRunStatus(navProgress{Expected: 2, Attempted: 2, Success: 1, Failed: 1}, 2, false); got != "partial" {
		t.Fatalf("partial status = %s", got)
	}
	if got := finalNavRunStatus(navProgress{}, 2, true); got != "canceled" {
		t.Fatalf("cancel status = %s", got)
	}
}

func TestNavMisfireClassification(t *testing.T) {
	now := time.Date(2026, 8, 27, 3, 0, 0, 0, time.UTC)
	status, trigger, priority := classifyDueSlot("skip", 90*time.Second, now.Add(-2*time.Minute), now, 0, 1, priorityScheduled, priorityCatchUp)
	if status != "missed" || trigger != "scheduled" || priority != priorityScheduled {
		t.Fatalf("point-in-time misfire = %s/%s/%d", status, trigger, priority)
	}
	status, _, _ = classifyDueSlot("catch_up_once", 5*time.Minute, now.Add(-time.Hour), now, 0, 4, priorityScheduled, priorityCatchUp)
	if status != "missed" {
		t.Fatalf("older state-refresh slot status = %s", status)
	}
	status, trigger, priority = classifyDueSlot("catch_up_once", 5*time.Minute, now.Add(-20*time.Minute), now, 3, 4, priorityScheduled, priorityCatchUp)
	if status != "queued" || trigger != "startup_catchup" || priority != priorityCatchUp {
		t.Fatalf("latest state-refresh slot = %s/%s/%d", status, trigger, priority)
	}
}
