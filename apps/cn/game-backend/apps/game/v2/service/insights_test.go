package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
)

type fakeInsightsStore struct {
	game          *v2models.InsightGameRecord
	state         *v2models.InsightGameStateRecord
	playerSummary v2models.InsightPlayerSummaryRecord
	price         *v2models.InsightPriceRecord
	players       []v2models.InsightPlayerPointRecord
	prices        []v2models.InsightPriceRecord
	summaries     map[string]*v2models.InsightMetricSummaryRecord
	trend         []v2models.InsightMetricTrendRecord
	changes       []v2models.InsightChangeRecord
}

func (f *fakeInsightsStore) CountInsightEntities(context.Context) (int64, error) { return 1, nil }
func (f *fakeInsightsStore) GetInsightGame(context.Context, int64) (*v2models.InsightGameRecord, error) {
	return f.game, nil
}
func (f *fakeInsightsStore) GetInsightMetricSummary(_ context.Context, c v2models.InsightMetricContract) (*v2models.InsightMetricSummaryRecord, error) {
	return f.summaries[c.PublicKey], nil
}
func (f *fakeInsightsStore) ListInsightMetricTrend(context.Context, v2models.InsightMetricContract, int32) ([]v2models.InsightMetricTrendRecord, error) {
	return f.trend, nil
}
func (f *fakeInsightsStore) GetInsightGameState(context.Context, int64) (*v2models.InsightGameStateRecord, error) {
	return f.state, nil
}
func (f *fakeInsightsStore) GetInsightPlayerSummary(context.Context, v2models.InsightGameStateRecord) (v2models.InsightPlayerSummaryRecord, error) {
	return f.playerSummary, nil
}
func (f *fakeInsightsStore) GetInsightPriceSummary(context.Context, int64) (*v2models.InsightPriceRecord, error) {
	return f.price, nil
}
func (f *fakeInsightsStore) ListInsightPlayerHistory(context.Context, int64, int32) ([]v2models.InsightPlayerPointRecord, error) {
	return f.players, nil
}
func (f *fakeInsightsStore) ListInsightPriceHistory(context.Context, int64, int32) ([]v2models.InsightPriceRecord, error) {
	return f.prices, nil
}
func (f *fakeInsightsStore) CountInsightOverviewChanges(context.Context, []string, []string) (int64, error) {
	return int64(len(f.changes)), nil
}
func (f *fakeInsightsStore) ListInsightOverviewChanges(context.Context, []string, []string, int32) ([]v2models.InsightChangeRecord, error) {
	return f.changes, nil
}
func (f *fakeInsightsStore) ListInsightGameChanges(context.Context, v2models.InsightGameRecord, []string, []string, int32) ([]v2models.InsightChangeRecord, error) {
	return f.changes, nil
}

func TestInsightMetricMappingIsExplicit(t *testing.T) {
	want := map[string]struct {
		key     string
		version int32
	}{"free": {"free_game_share", 1}, "windows": {"windows_support", 1}, "linux": {"linux_support", 1}}
	for publicKey, expected := range want {
		contract, ok := resolveInsightMetric(publicKey)
		if !ok || contract.InternalKey != expected.key || contract.Version != expected.version {
			t.Fatalf("mapping %q = %#v, %v", publicKey, contract, ok)
		}
	}
	if _, ok := resolveInsightMetric("free_game_share"); ok {
		t.Fatal("internal key must not be public")
	}
}

func TestPlayerMissingAndRealZeroStayDistinct(t *testing.T) {
	date := time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC)
	store := &fakeInsightsStore{
		game:      &v2models.InsightGameRecord{ID: 1, Name: "Game"},
		state:     &v2models.InsightGameStateRecord{GameID: 1, FactDate: date, TrackingPeriodID: 10, AppID: 20},
		summaries: map[string]*v2models.InsightMetricSummaryRecord{},
	}
	got, err := NewInsightsService(store).GetGameInsights(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.Players.Current != nil || got.Players.Peak30D != nil {
		t.Fatalf("missing player sample became a value: %#v", got.Players)
	}
	store.playerSummary = v2models.InsightPlayerSummaryRecord{HasCurrent: true, Current: 0, HasPeak30D: true, Peak30D: 0}
	got, err = NewInsightsService(store).GetGameInsights(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.Players.Current == nil || *got.Players.Current != 0 || got.Players.Peak30D == nil || *got.Players.Peak30D != 0 {
		t.Fatalf("real zero was lost: %#v", got.Players)
	}
}

func TestPriceFreeAndPricedZeroStayDistinctAndHistoryIsNotFabricated(t *testing.T) {
	date := time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC)
	zero := int64(0)
	store := &fakeInsightsStore{
		game:  &v2models.InsightGameRecord{ID: 1, Name: "Game"},
		state: &v2models.InsightGameStateRecord{GameID: 1, FactDate: date, TrackingPeriodID: 10},
		prices: []v2models.InsightPriceRecord{
			{FactDate: date.AddDate(0, 0, -1), State: "free"},
			{FactDate: date, State: "priced", FinalAmount: &zero},
		},
		summaries: map[string]*v2models.InsightMetricSummaryRecord{},
	}
	got, err := NewInsightsService(store).GetGamePriceInsights(context.Background(), 1, "30d")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Points) != 2 || got.Points[0].State != "free" || got.Points[0].FinalAmount != nil || got.Points[1].State != "priced" || got.Points[1].FinalAmount == nil || *got.Points[1].FinalAmount != 0 {
		t.Fatalf("price semantics collapsed: %#v", got.Points)
	}
	if got.AvailableFrom == nil || *got.AvailableFrom != "2026-08-29" || got.AvailableThrough == nil || *got.AvailableThrough != "2026-08-30" {
		t.Fatalf("availability = %#v", got)
	}
}

func TestChangeContractAndDayPrecisionDoNotLeakInternals(t *testing.T) {
	date := time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC)
	at := date.Add(12 * time.Hour)
	changes := insightPublicChanges([]v2models.InsightChangeRecord{
		{EntityID: 1, EntityName: "Game", DetectorKey: "game_price_transition", DetectorVersion: 1, EventCode: "game_price_increased", ProjectionDate: date, TimeBasis: "day", EventAt: &at},
		{EntityID: 1, EntityName: "Game", DetectorKey: "game_price_transition", DetectorVersion: 2, EventCode: "game_price_increased", ProjectionDate: date},
	})
	if len(changes) != 1 || changes[0].Type != "game.price.increased" || changes[0].OccurredAt != nil || changes[0].Detail != nil {
		t.Fatalf("changes = %#v", changes)
	}
	payload, err := json.Marshal(changes)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"detector_key", "detector_version", "event_code", "old_value", "new_value"} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("payload leaked %q: %s", forbidden, payload)
		}
	}
}

func TestInsightValidationAndExactDelta(t *testing.T) {
	store := &fakeInsightsStore{summaries: map[string]*v2models.InsightMetricSummaryRecord{}}
	service := NewInsightsService(store)
	if _, err := service.GetInsightsMetricTrend(context.Background(), "bad", "30d"); !errors.Is(err, ErrInvalidInsightMetric) {
		t.Fatalf("invalid key error = %v", err)
	}
	if _, err := service.GetInsightsMetricTrend(context.Background(), "free", "7d"); !errors.Is(err, ErrInvalidInsightRange) {
		t.Fatalf("invalid range error = %v", err)
	}
	record := v2models.InsightMetricSummaryRecord{PositiveCount: 1, NegativeCount: 1, EligibleCount: 2}
	if metric := insightPublicMetric("free", record); metric.Delta30D != nil {
		t.Fatalf("missing exact comparison produced delta: %#v", metric)
	}
	if _, err := service.GetGameInsights(context.Background(), 404); !errors.Is(err, ErrInsightGameNotFound) {
		t.Fatalf("missing game error = %v", err)
	}
}
