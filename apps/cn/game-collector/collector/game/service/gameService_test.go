package service

import (
	"testing"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
)

func TestSingleGameTaskTypesExcludePlayers(t *testing.T) {
	tasks := singleGameTaskTypes()
	if len(tasks) != 2 || tasks[0] != domain.TaskDetails || tasks[1] != domain.TaskNews {
		t.Fatalf("single-game tasks = %v, want [details news]", tasks)
	}
	for _, task := range tasks {
		if task == domain.TaskPlayers {
			t.Fatal("single-game collection must not include players")
		}
	}
}
