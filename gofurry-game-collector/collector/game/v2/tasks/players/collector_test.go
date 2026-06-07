package players

import (
	"context"
	"testing"

	"github.com/gofurry/gofurry-game-collector/collector/game/models"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
)

type recordingRepository struct {
	items []domain.PlayerCount
}

func (r *recordingRepository) SavePlayerCount(_ context.Context, item domain.PlayerCount) error {
	r.items = append(r.items, item)
	return nil
}

func TestCollectGameRejectsMissingAdapterAndRecordsFailure(t *testing.T) {
	t.Parallel()

	repo := &recordingRepository{}
	collector := NewCollector(nil, repo)
	result, err := collector.CollectGame(context.Background(), models.GameID{ID: 1, Appid: 550})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if result.Status != domain.StatusFailed {
		t.Fatalf("unexpected status: %s", result.Status)
	}
	if len(repo.items) != 1 || repo.items[0].Status != domain.StatusFailed {
		t.Fatalf("expected one failed player count record, got %#v", repo.items)
	}
}
