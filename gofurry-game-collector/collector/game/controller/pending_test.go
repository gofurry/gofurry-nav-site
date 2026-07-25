package controller

import (
	"testing"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/report"
)

func TestPendingGameDetailsCollected(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		results []report.TaskResult
		want    bool
	}{
		{
			name: "details success",
			results: []report.TaskResult{
				{Task: domain.TaskDetails, Status: domain.StatusSuccess},
				{Task: domain.TaskNews, Status: domain.StatusFailed},
			},
			want: true,
		},
		{
			name: "details partial",
			results: []report.TaskResult{
				{Task: domain.TaskDetails, Status: domain.StatusPartial},
			},
			want: true,
		},
		{
			name: "details failed",
			results: []report.TaskResult{
				{Task: domain.TaskDetails, Status: domain.StatusFailed},
				{Task: domain.TaskPlayers, Status: domain.StatusSuccess},
			},
			want: false,
		},
		{
			name: "details missing",
			results: []report.TaskResult{
				{Task: domain.TaskNews, Status: domain.StatusSuccess},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := pendingGameDetailsCollected(report.RunSummary{Results: tt.results}); got != tt.want {
				t.Fatalf("pendingGameDetailsCollected() = %v, want %v", got, tt.want)
			}
		})
	}
}
