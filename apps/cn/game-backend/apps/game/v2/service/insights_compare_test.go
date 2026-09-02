package service

import (
	"context"
	"errors"
	"testing"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
)

func TestGameComparePreservesSemanticZerosAndCommonHorizons(t *testing.T) {
	horizon := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	snapshot := horizon.Add(12 * time.Hour)
	observed := snapshot.Add(time.Minute)
	priced, free := "priced", "free"
	currency := "CNY"
	zero := int64(0)
	discount := int32(100)
	store := &fakeInsightsStore{
		games: map[int64]*v2models.InsightGameRecord{
			2: {ID: 2, Name: "Second"},
			1: {ID: 1, Name: "First"},
		},
		compareHorizon: &horizon,
		compareFacts: []v2models.InsightGameCompareFactRecord{
			{GameID: 2, TrackingPeriodID: 20, LanguageEvidence: "fresh", LanguageCodes: []string{"en"}, FullAudioLanguageCodes: []string{}, UnknownLanguageNames: []string{}, PriceAvailable: true, PriceState: &priced, Currency: &currency, InitialAmount: &zero, FinalAmount: &zero, DiscountPercent: &discount},
			{GameID: 1, TrackingPeriodID: 10, LanguageEvidence: "stale", LanguageCodes: []string{"zh-CN"}, FullAudioLanguageCodes: []string{"zh-CN"}, UnknownLanguageNames: []string{"Klingon"}, PriceAvailable: true, PriceState: &free},
		},
		compareCurrent: []v2models.InsightGameCompareCurrentPlayerRecord{
			{GameID: 2, SnapshotScheduledFor: &snapshot, Available: true, PlayerCount: 0, CollectedAt: &observed},
			{GameID: 1, SnapshotScheduledFor: &snapshot, Available: false},
		},
		comparePlayer30D: []v2models.InsightGameComparePlayer30DRecord{
			{GameID: 2, FactThrough: &horizon, EligibleFrom: &horizon, Peak30D: &zero, Average30D: float64TestPointer(0), ObservedDays: 1, SuccessfulSamples: 1, SampleCoverage: float64TestPointer(1)},
			{GameID: 1, FactThrough: &horizon, EligibleFrom: &horizon},
		},
		observedLow: &v2models.InsightObservedLowRecord{Amount: 0, Currency: "CNY", FirstSeen: horizon, ObservedSince: horizon, InitialAmount: 0, DiscountPercent: 100},
	}

	got, err := NewInsightsService(store).GetGameCompare(context.Background(), "2,1,2", "CN")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "ready" || got.StateAsOf == nil || *got.StateAsOf != "2026-09-01" || got.PlayerFactThrough == nil || *got.PlayerFactThrough != "2026-09-01" {
		t.Fatalf("horizons = %#v", got)
	}
	if len(got.Games) != 2 || got.Games[0].Game.ID != 2 || got.Games[1].Game.ID != 1 {
		t.Fatalf("order/dedup = %#v", got.Games)
	}
	first := got.Games[0]
	if !first.Players.CurrentAvailable || first.Players.Current == nil || *first.Players.Current != 0 || first.Players.Peak30D == nil || *first.Players.Peak30D != 0 {
		t.Fatalf("real player zero was lost: %#v", first.Players)
	}
	if !first.Price.Available || first.Price.State == nil || *first.Price.State != "priced" || first.Price.FinalAmount == nil || *first.Price.FinalAmount != 0 || first.Price.ObservedLow == nil || first.Price.ObservedLow.Amount != 0 {
		t.Fatalf("priced zero semantics were lost: %#v", first.Price)
	}
	second := got.Games[1]
	if second.Players.CurrentAvailable || second.Players.Current != nil {
		t.Fatalf("unavailable player became zero: %#v", second.Players)
	}
	if second.Price.State == nil || *second.Price.State != "free" || second.Price.FinalAmount != nil {
		t.Fatalf("free became monetary zero: %#v", second.Price)
	}
	if second.Languages.Evidence != "stale" || len(second.Languages.UnknownNames) != 1 {
		t.Fatalf("language freshness/unknown names changed: %#v", second.Languages)
	}
}

func TestGameCompareValidationNotFoundAndInsufficientData(t *testing.T) {
	for _, ids := range []string{"", "1", "1,1", "1,bad", "1,2,3,4,5"} {
		if _, err := NewInsightsService(&fakeInsightsStore{}).GetGameCompare(context.Background(), ids, "CN"); !errors.Is(err, ErrInvalidInsightCompare) {
			t.Fatalf("ids=%q err=%v", ids, err)
		}
	}
	if _, err := NewInsightsService(&fakeInsightsStore{}).GetGameCompare(context.Background(), "1,2", "JP"); !errors.Is(err, ErrInvalidInsightRegion) {
		t.Fatalf("region err=%v", err)
	}
	store := &fakeInsightsStore{games: map[int64]*v2models.InsightGameRecord{1: {ID: 1, Name: "First"}}}
	if _, err := NewInsightsService(store).GetGameCompare(context.Background(), "1,2", "CN"); !errors.Is(err, ErrInsightGameNotFound) {
		t.Fatalf("missing game err=%v", err)
	}
	store.games[2] = &v2models.InsightGameRecord{ID: 2, Name: "Second"}
	got, err := NewInsightsService(store).GetGameCompare(context.Background(), "1,2", "HK")
	if err != nil || got.Status != "insufficient_data" || got.StateAsOf != nil || len(got.Games) != 2 {
		t.Fatalf("insufficient snapshot = %#v err=%v", got, err)
	}
}

func float64TestPointer(value float64) *float64 { return &value }
