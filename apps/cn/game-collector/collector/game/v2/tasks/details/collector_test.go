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

func TestRequestPlanKeepsCNBaseAndUSCanonicalAuthority(t *testing.T) {
	t.Parallel()
	collector := NewCollector(nil, recordingRepository{})
	if len(collector.requests) != 3 || collector.requests[0].region != domain.RegionCN || !collector.requests[0].preferAsBase || collector.requests[0].canonical {
		t.Fatalf("unexpected CN base plan: %+v", collector.requests)
	}
	if collector.requests[1].region != domain.RegionUS || collector.requests[1].lang != domain.StoreLocaleEN || !collector.requests[1].canonical {
		t.Fatalf("unexpected US canonical plan: %+v", collector.requests)
	}
	if collector.requests[2].canonical {
		t.Fatalf("HK fallback must not be canonical: %+v", collector.requests[2])
	}
}
