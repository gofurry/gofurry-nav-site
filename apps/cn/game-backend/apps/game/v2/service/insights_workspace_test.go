package service

import (
	"context"
	"encoding/json"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"strings"
	"testing"
	"time"
)

func TestWorkspaceRankingVisualPreservesEvidence(t *testing.T) {
	observed := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	days, samples, coverage := int64(27), int64(180), .75
	asset := "https://example.test/hashed/header.jpg?version=2"
	store := &fakeInsightsStore{
		rankingMeta: v2models.InsightPlayerRankingMetaRecord{Population: 3, Ranked: 2, SnapshotScheduledFor: &observed},
		rankingRows: []v2models.InsightPlayerRankingRecord{
			{GameID: 1, GameName: "With header", VisualAsset: " " + asset + " ", Value: 12.5, ObservedAt: &observed, ObservedDays: &days, SuccessfulSamples: &samples, SampleCoverage: &coverage},
			{GameID: 2, GameName: "No header", Value: 0},
		},
	}
	for _, metric := range []string{"latest_observed", "peak_30d", "average_30d"} {
		got, err := NewInsightsService(store).GetPlayerRanking(context.Background(), v2models.InsightPlayerRankingQuery{Metric: metric})
		if err != nil {
			t.Fatal(err)
		}
		first, second := got.Items[0], got.Items[1]
		if first.Game.Visual == nil || first.Game.Visual.Kind != "game_header" || first.Game.Visual.Asset != asset {
			t.Fatalf("%s visual = %#v", metric, first.Game)
		}
		if first.Rank != 1 || second.Rank != 2 || first.Value != 12.5 || second.Value != 0 || first.ObservedAt != &observed || *first.ObservedDays != days || *first.SuccessfulSamples != samples || *first.SampleCoverage != coverage {
			t.Fatalf("%s ranking evidence changed: %#v", metric, got)
		}
		encoded, err := json.Marshal(second.Game)
		if err != nil || second.Game.Visual != nil || strings.Contains(string(encoded), "visual") {
			t.Fatalf("missing visual should be omitted: %s, %v", encoded, err)
		}
	}
}

func TestWorkspaceDiscountVisualPreservesAmounts(t *testing.T) {
	day := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	store := &fakeInsightsStore{discounts: []v2models.InsightDiscountRecord{
		{GameID: 1, GameName: "Header", VisualAsset: "https://example.test/header.jpg", AsOf: day, Currency: "USD", InitialAmount: 999, FinalAmount: 599, DiscountPercent: 40},
		{GameID: 2, GameName: "Fallback", VisualAsset: " ", AsOf: day, Currency: "USD", InitialAmount: 100, FinalAmount: 0, DiscountPercent: 100},
	}, observedLow: &v2models.InsightObservedLowRecord{Amount: 399, Currency: "USD", FirstSeen: day, ObservedSince: day}}
	got, err := NewInsightsService(store).GetDiscounts(context.Background(), "US", 20)
	if err != nil {
		t.Fatal(err)
	}
	first, second := got.Items[0], got.Items[1]
	if first.Game.Visual == nil || first.Game.Visual.Kind != "game_header" || first.Game.Visual.Asset != store.discounts[0].VisualAsset {
		t.Fatalf("discount visual = %#v", first.Game)
	}
	if first.InitialAmount != 999 || first.FinalAmount != 599 || first.DiscountPercent != 40 || first.ObservedLow.Amount != 399 || first.Currency != "USD" || second.FinalAmount != 0 {
		t.Fatalf("discount evidence changed: %#v", got)
	}
	encoded, err := json.Marshal(second.Game)
	if err != nil || strings.Contains(string(encoded), "visual") {
		t.Fatalf("missing discount visual should be omitted: %s", encoded)
	}
}
