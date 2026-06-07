package details

import (
	"context"
	"testing"

	"github.com/gofurry/gofurry-game-collector/collector/game/models"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
)

type recordingRepository struct{}

func (recordingRepository) SaveDetails(context.Context, domain.DetailsCollection) error { return nil }

func TestCollectGameRejectsMissingAdapter(t *testing.T) {
	t.Parallel()

	collector := NewCollector(nil, recordingRepository{})
	result, err := collector.CollectGame(context.Background(), models.GameID{ID: 1, Appid: 550})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if result.Status != domain.StatusFailed {
		t.Fatalf("unexpected status: %s", result.Status)
	}
}
