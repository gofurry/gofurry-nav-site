package runner

import (
	"context"
	"errors"
	"testing"

	"github.com/gofurry/gofurry-game-collector/collector/game/models"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/report"
)

type fakeCollector struct {
	status domain.Status
	err    error
}

func (c fakeCollector) CollectGame(_ context.Context, game models.GameID) (report.TaskResult, error) {
	return report.TaskResult{
		Task:   domain.TaskDetails,
		Status: c.status,
		GameID: game.ID,
		AppID:  uint32(game.Appid),
	}, c.err
}

func TestRunAggregatesResultsAndContinuesAfterFailure(t *testing.T) {
	t.Parallel()

	r := New(Options{RunID: "test-run", MaxWorkers: 2}, []TaskBinding{
		{Task: domain.TaskDetails, Collector: fakeCollector{status: domain.StatusSuccess}},
		{Task: domain.TaskNews, Collector: fakeCollector{status: domain.StatusFailed, err: errors.New("upstream failed")}},
	})

	summary, err := r.Run(context.Background(), []models.GameID{
		{ID: 1, Appid: 440},
		{ID: 2, Appid: 570},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if summary.Status != domain.StatusPartial {
		t.Fatalf("unexpected status: %s", summary.Status)
	}
	if summary.TotalCount != 4 || summary.SuccessCount != 2 || summary.FailedCount != 2 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	if len(summary.Results) != 4 {
		t.Fatalf("unexpected result count: %d", len(summary.Results))
	}
}

func TestRunRejectsEmptyTasks(t *testing.T) {
	t.Parallel()

	r := New(Options{}, nil)
	summary, err := r.Run(context.Background(), []models.GameID{{ID: 1, Appid: 440}})
	if err == nil {
		t.Fatal("expected error")
	}
	if summary.Status != domain.StatusFailed {
		t.Fatalf("unexpected status: %s", summary.Status)
	}
}
