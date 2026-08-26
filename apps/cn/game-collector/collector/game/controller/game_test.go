package controller

import (
	"testing"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/report"
)

func TestRunStartupCollectionPrioritizesPending(t *testing.T) {
	pendingCalls := 0
	playerCalls := 0
	runStartupCollection(func() bool {
		pendingCalls++
		return true
	}, true, func() { playerCalls++ })
	if pendingCalls != 1 || playerCalls != 0 {
		t.Fatalf("pendingCalls=%d playerCalls=%d", pendingCalls, playerCalls)
	}
}

func TestRunStartupCollectionHonorsPlayersSetting(t *testing.T) {
	for _, test := range []struct {
		name        string
		enabled     bool
		playerCalls int
	}{
		{name: "enabled", enabled: true, playerCalls: 1},
		{name: "disabled", enabled: false, playerCalls: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			runStartupCollection(func() bool { return false }, test.enabled, func() { calls++ })
			if calls != test.playerCalls {
				t.Fatalf("player calls=%d, want %d", calls, test.playerCalls)
			}
		})
	}
}

func TestParsePendingGameCollectMember(t *testing.T) {
	game, ok := parsePendingGameCollectMember("303:3024860")
	if !ok || game.ID != 303 || game.Appid != 3024860 {
		t.Fatalf("game=%+v ok=%v", game, ok)
	}
	for _, invalid := range []string{"", "303", "0:1", "1:0", "x:1", "1:x"} {
		if _, ok := parsePendingGameCollectMember(invalid); ok {
			t.Fatalf("accepted invalid member %q", invalid)
		}
	}
}

func TestPendingGameDetailsCollected(t *testing.T) {
	for _, status := range []domain.Status{domain.StatusSuccess, domain.StatusPartial} {
		if !pendingGameDetailsCollected(report.RunSummary{Results: []report.TaskResult{{Task: domain.TaskDetails, Status: status}}}) {
			t.Fatalf("details status %s should complete pending work", status)
		}
	}
	if pendingGameDetailsCollected(report.RunSummary{Results: []report.TaskResult{{Task: domain.TaskDetails, Status: domain.StatusFailed}}}) {
		t.Fatal("failed details must remain pending")
	}
	if pendingGameDetailsCollected(report.RunSummary{Results: []report.TaskResult{{Task: domain.TaskPlayers, Status: domain.StatusSuccess}}}) {
		t.Fatal("players-only result must remain pending")
	}
}
