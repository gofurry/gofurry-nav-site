package control

import (
	"testing"
	"time"

	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestDueSlotsAnchoredIntervalDoesNotUseProcessStart(t *testing.T) {
	anchor := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	step := int64(3600)
	schedule := gamesqlc.GfgCollectionSchedule{
		JobKey: "game.players", ScheduleKind: "interval", IntervalSeconds: &step,
		AnchorAt:      pgtype.Timestamptz{Time: anchor, Valid: true},
		EffectiveFrom: pgtype.Timestamptz{Time: anchor.Add(15 * time.Minute), Valid: true},
	}
	due, next, err := dueSlots(schedule, anchor.Add(3*time.Hour+20*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	want := []time.Time{anchor.Add(time.Hour), anchor.Add(2 * time.Hour), anchor.Add(3 * time.Hour)}
	if len(due) != len(want) {
		t.Fatalf("due slots = %v, want %v", due, want)
	}
	for index := range want {
		if !due[index].Equal(want[index]) {
			t.Fatalf("due[%d] = %s, want %s", index, due[index], want[index])
		}
	}
	if !next.Equal(anchor.Add(4 * time.Hour)) {
		t.Fatalf("next = %s", next)
	}

	// A restart after the second slot resumes from the durable cursor rather
	// than shifting the phase to the new process start time.
	schedule.LastMaterializedFor = pgtype.Timestamptz{Time: anchor.Add(2 * time.Hour), Valid: true}
	due, next, err = dueSlots(schedule, anchor.Add(3*time.Hour+20*time.Minute))
	if err != nil || len(due) != 1 || !due[0].Equal(anchor.Add(3*time.Hour)) || !next.Equal(anchor.Add(4*time.Hour)) {
		t.Fatalf("restart slots = %v next=%s err=%v", due, next, err)
	}
}

func TestNextSlotFixedCronTimezone(t *testing.T) {
	expression := "0 3 * * *"
	schedule := gamesqlc.GfgCollectionSchedule{
		JobKey: "game.metadata", ScheduleKind: "cron", CronExpression: &expression,
		Timezone: "Asia/Shanghai",
	}
	next, err := nextSlot(schedule, time.Date(2026, 8, 26, 18, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 8, 27, 3, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	if !next.Equal(want) {
		t.Fatalf("next = %s, want %s", next, want)
	}
}

func TestFinalRunStatus(t *testing.T) {
	if got := finalRunStatus(progressDocument{Expected: 2, Attempted: 2, Success: 2}, 2, false); got != "success" {
		t.Fatalf("success status = %s", got)
	}
	if got := finalRunStatus(progressDocument{Expected: 2, Attempted: 1, Success: 1}, 2, false); got != "partial" {
		t.Fatalf("partial status = %s", got)
	}
	if got := finalRunStatus(progressDocument{}, 2, true); got != "canceled" {
		t.Fatalf("cancel status = %s", got)
	}
}

func TestMisfireClassification(t *testing.T) {
	now := time.Date(2026, 8, 27, 3, 0, 0, 0, time.UTC)
	status, trigger, priority := classifyDueSlot("skip", 90*time.Second, now.Add(-2*time.Minute), now, 0, 1, PriorityScheduled, PriorityCatchUp)
	if status != "missed" || trigger != "scheduled" || priority != PriorityScheduled {
		t.Fatalf("skip misfire = %s/%s/%d", status, trigger, priority)
	}
	status, _, _ = classifyDueSlot("catch_up_once", 5*time.Minute, now.Add(-time.Hour), now, 0, 3, PriorityScheduled, PriorityCatchUp)
	if status != "missed" {
		t.Fatalf("older refresh slot status = %s", status)
	}
	status, trigger, priority = classifyDueSlot("catch_up_once", 5*time.Minute, now.Add(-20*time.Minute), now, 2, 3, PriorityScheduled, PriorityCatchUp)
	if status != "queued" || trigger != "startup_catchup" || priority != PriorityCatchUp {
		t.Fatalf("latest refresh slot = %s/%s/%d", status, trigger, priority)
	}
}
