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
	game              *v2models.InsightGameRecord
	games             map[int64]*v2models.InsightGameRecord
	state             *v2models.InsightGameStateRecord
	playerSummary     v2models.InsightPlayerSummaryRecord
	price             *v2models.InsightPriceRecord
	players           []v2models.InsightPlayerPointRecord
	prices            []v2models.InsightPriceRecord
	summaries         map[string]*v2models.InsightMetricSummaryRecord
	trend             []v2models.InsightMetricTrendRecord
	changes           []v2models.InsightChangeRecord
	breakdown         []v2models.InsightDimensionRecord
	sliceAvailability v2models.InsightDimensionAvailabilityRecord
	sliceTrend        []v2models.InsightDimensionTrendRecord
	lastBreakdownDate time.Time
	lastSliceThrough  time.Time
	lastSliceValue    string
	lastExplorer      v2models.InsightChangeExplorerConditions
	regional          []v2models.InsightRegionalPriceRecord
	observedLow       *v2models.InsightObservedLowRecord
	rankingMeta       v2models.InsightPlayerRankingMetaRecord
	rankingRows       []v2models.InsightPlayerRankingRecord
	priceOverview     v2models.InsightPriceOverviewRecord
	discounts         []v2models.InsightDiscountRecord
	languageOverview  v2models.InsightLanguageOverviewRecord
	languages         []v2models.InsightLanguageRecord
	compareHorizon    *time.Time
	compareFacts      []v2models.InsightGameCompareFactRecord
	compareCurrent    []v2models.InsightGameCompareCurrentPlayerRecord
	comparePlayer30D  []v2models.InsightGameComparePlayer30DRecord
}

func (f *fakeInsightsStore) CountInsightEntities(context.Context) (int64, error) { return 1, nil }
func (f *fakeInsightsStore) GetInsightGame(_ context.Context, id int64) (*v2models.InsightGameRecord, error) {
	if f.games != nil {
		return f.games[id], nil
	}
	return f.game, nil
}
func (f *fakeInsightsStore) GetInsightMetricSummary(_ context.Context, c v2models.InsightMetricContract) (*v2models.InsightMetricSummaryRecord, error) {
	return f.summaries[c.PublicKey], nil
}
func (f *fakeInsightsStore) ListInsightMetricTrend(context.Context, v2models.InsightMetricContract, int32) ([]v2models.InsightMetricTrendRecord, error) {
	return f.trend, nil
}
func (f *fakeInsightsStore) ListInsightMetricBreakdown(_ context.Context, _ v2models.InsightMetricContract, _ v2models.InsightDimensionContract, date time.Time) ([]v2models.InsightDimensionRecord, error) {
	f.lastBreakdownDate = date
	return f.breakdown, nil
}
func (f *fakeInsightsStore) GetInsightMetricSliceAvailability(context.Context, v2models.InsightMetricContract, v2models.InsightDimensionContract, string) (v2models.InsightDimensionAvailabilityRecord, error) {
	return f.sliceAvailability, nil
}
func (f *fakeInsightsStore) ListInsightMetricSliceTrend(_ context.Context, _ v2models.InsightMetricContract, _ v2models.InsightDimensionContract, value string, through time.Time, _ int32) ([]v2models.InsightDimensionTrendRecord, error) {
	f.lastSliceValue, f.lastSliceThrough = value, through
	return f.sliceTrend, nil
}
func (f *fakeInsightsStore) GetInsightGameState(context.Context, int64) (*v2models.InsightGameStateRecord, error) {
	return f.state, nil
}
func (f *fakeInsightsStore) GetInsightGameCompareFactHorizon(context.Context, []int64) (*time.Time, error) {
	return f.compareHorizon, nil
}
func (f *fakeInsightsStore) ListInsightGameCompareFacts(context.Context, []int64, time.Time, string) ([]v2models.InsightGameCompareFactRecord, error) {
	return f.compareFacts, nil
}
func (f *fakeInsightsStore) ListInsightGameCompareCurrentPlayers(context.Context, []int64) ([]v2models.InsightGameCompareCurrentPlayerRecord, error) {
	return f.compareCurrent, nil
}
func (f *fakeInsightsStore) ListInsightGameComparePlayer30D(context.Context, []int64) ([]v2models.InsightGameComparePlayer30DRecord, error) {
	return f.comparePlayer30D, nil
}
func (f *fakeInsightsStore) GetInsightPlayerSummary(context.Context, v2models.InsightGameStateRecord) (v2models.InsightPlayerSummaryRecord, error) {
	return f.playerSummary, nil
}
func (f *fakeInsightsStore) ListInsightPlayerHistory(context.Context, int64, int32) ([]v2models.InsightPlayerPointRecord, error) {
	return f.players, nil
}
func (f *fakeInsightsStore) ListInsightPriceHistory(context.Context, int64, string, time.Time, int32) ([]v2models.InsightPriceRecord, error) {
	return f.prices, nil
}
func (f *fakeInsightsStore) ListInsightRegionalPrices(context.Context, int64, time.Time) ([]v2models.InsightRegionalPriceRecord, error) {
	if f.regional != nil {
		return f.regional, nil
	}
	rows := []v2models.InsightRegionalPriceRecord{{Region: "CN"}, {Region: "US"}, {Region: "HK"}}
	if f.price != nil {
		state := f.price.State
		rows[0] = v2models.InsightRegionalPriceRecord{Region: "CN", Available: true, FactDate: f.price.FactDate, State: &state, Currency: f.price.Currency, InitialAmount: f.price.InitialAmount, FinalAmount: f.price.FinalAmount, DiscountPercent: f.price.DiscountPercent}
	}
	return rows, nil
}
func (f *fakeInsightsStore) GetInsightObservedLow(context.Context, int64, string, time.Time) (*v2models.InsightObservedLowRecord, error) {
	return f.observedLow, nil
}
func (f *fakeInsightsStore) GetLatestPlayerRankingMeta(context.Context) (v2models.InsightPlayerRankingMetaRecord, error) {
	return f.rankingMeta, nil
}
func (f *fakeInsightsStore) ListLatestPlayerRanking(context.Context, int32) ([]v2models.InsightPlayerRankingRecord, error) {
	return f.rankingRows, nil
}
func (f *fakeInsightsStore) GetPlayer30DRankingMeta(context.Context) (v2models.InsightPlayerRankingMetaRecord, error) {
	return f.rankingMeta, nil
}
func (f *fakeInsightsStore) ListPlayer30DRanking(context.Context, bool, int32) ([]v2models.InsightPlayerRankingRecord, error) {
	return f.rankingRows, nil
}
func (f *fakeInsightsStore) GetInsightPriceOverview(context.Context, string) (v2models.InsightPriceOverviewRecord, error) {
	return f.priceOverview, nil
}
func (f *fakeInsightsStore) ListInsightDiscounts(context.Context, string, int32) ([]v2models.InsightDiscountRecord, error) {
	return f.discounts, nil
}
func (f *fakeInsightsStore) GetInsightLanguageOverview(context.Context) (v2models.InsightLanguageOverviewRecord, error) {
	return f.languageOverview, nil
}
func (f *fakeInsightsStore) ListInsightLanguages(context.Context) ([]v2models.InsightLanguageRecord, error) {
	return f.languages, nil
}
func (f *fakeInsightsStore) CountInsightOverviewChanges(context.Context, []string, []string) (int64, error) {
	return int64(len(f.changes)), nil
}
func (f *fakeInsightsStore) ListInsightOverviewChanges(context.Context, []string, []string, int32) ([]v2models.InsightChangeRecord, error) {
	return f.changes, nil
}
func (f *fakeInsightsStore) ListInsightExplorerChanges(_ context.Context, conditions v2models.InsightChangeExplorerConditions) ([]v2models.InsightChangeRecord, error) {
	f.lastExplorer = conditions
	return f.changes, nil
}
func (f *fakeInsightsStore) ListInsightGameChanges(context.Context, v2models.InsightGameRecord, []string, []string, int32) ([]v2models.InsightChangeRecord, error) {
	return f.changes, nil
}

func TestInsightMetricMappingIsExplicit(t *testing.T) {
	want := map[string]struct {
		key     string
		version int32
	}{"free": {"free_game_share", 1}, "windows": {"windows_support", 1}, "mac": {"mac_support", 1}, "linux": {"linux_support", 1}}
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
	zero := int64(0)
	store.playerSummary = v2models.InsightPlayerSummaryRecord{HasCurrent: true, Current: 0, Peak30D: &zero}
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
	got, err := NewInsightsService(store).GetGamePriceInsights(context.Background(), 1, "CN", "30d")
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
		{EntityID: 1, EntityName: "Game", DetectorKey: "mac_support_transition", DetectorVersion: 1, EventCode: "mac_support_added", ProjectionDate: date, TimeBasis: "observed", EventAt: &at},
		{EntityID: 1, EntityName: "Game", DetectorKey: "game_price_transition", DetectorVersion: 2, EventCode: "game_price_increased", ProjectionDate: date},
	})
	if len(changes) != 2 || changes[0].Type != "game.price.increased" || changes[0].OccurredAt != nil || changes[0].Detail != nil || changes[1].Type != "game.mac.added" || changes[1].OccurredAt == nil {
		t.Fatalf("changes = %#v", changes)
	}
	macCategory := ""
	for _, contract := range insightChangeContracts {
		if contract.public == "game.mac.added" {
			macCategory = contract.category
		}
	}
	if macCategory != "platform" {
		t.Fatalf("Mac public category = %q", macCategory)
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

func TestParseInsightRange(t *testing.T) {
	tests := map[string]int32{
		"30d":  30,
		"90d":  90,
		"180d": 180,
		"1y":   365,
		"3y":   1095,
		"5y":   1825,
		"all":  0,
	}
	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			got, ok := parseInsightRange(input)
			if !ok || got != want {
				t.Fatalf("parseInsightRange(%q) = %d, %v; want %d, true", input, got, ok, want)
			}
		})
	}
	if got, ok := parseInsightRange("7d"); ok || got != 0 {
		t.Fatalf("parseInsightRange(7d) = %d, %v; want 0, false", got, ok)
	}
}

func TestInsightDimensionsUseGlobalHorizonNullMathAndDeletedTagFallback(t *testing.T) {
	horizon := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	store := &fakeInsightsStore{
		summaries: map[string]*v2models.InsightMetricSummaryRecord{"free": {FactDate: horizon}},
		breakdown: []v2models.InsightDimensionRecord{
			{Value: "123", Population: 5, Eligible: 5, PositiveCount: 2, NegativeCount: 2},
			{Value: "unknown", Population: 1, Eligible: 0},
		},
	}
	got, err := NewInsightsService(store).GetInsightsMetricBreakdown(context.Background(), "free", "tag")
	if err != nil {
		t.Fatal(err)
	}
	if got.SliceMode != "overlapping" || got.AsOf == nil || *got.AsOf != "2026-09-01" || !store.lastBreakdownDate.Equal(horizon) {
		t.Fatalf("breakdown contract/horizon = %#v date=%v", got, store.lastBreakdownDate)
	}
	if len(got.Items) != 2 || got.Items[0].Label == nil || *got.Items[0].Label != "标签 #123" || got.Items[0].LabelEn == nil || *got.Items[0].LabelEn != "Tag #123" {
		t.Fatalf("deleted tag fallback = %#v", got.Items)
	}
	if got.Items[1].MetricValue != nil || got.Items[1].Coverage != nil || got.Items[1].Value != "unknown" {
		t.Fatalf("unknown/zero denominator collapsed = %#v", got.Items[1])
	}
}

func TestInsightSliceTrendUsesGlobalHorizonAndDoesNotFill(t *testing.T) {
	horizon := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	from, through := horizon.AddDate(0, 0, -10), horizon.AddDate(0, 0, -2)
	store := &fakeInsightsStore{
		summaries:         map[string]*v2models.InsightMetricSummaryRecord{"free": {FactDate: horizon}},
		sliceAvailability: v2models.InsightDimensionAvailabilityRecord{AvailableFrom: &from, AvailableThrough: &through},
		sliceTrend: []v2models.InsightDimensionTrendRecord{
			{FactDate: horizon.AddDate(0, 0, -2), Population: 4, Eligible: 4, PositiveCount: 2, NegativeCount: 2},
			{FactDate: horizon, Population: 4, Eligible: 0},
		},
	}
	got, err := NewInsightsService(store).GetInsightsMetricSliceTrend(context.Background(), "free", "tag", "123", "90d")
	if err != nil {
		t.Fatal(err)
	}
	if !store.lastSliceThrough.Equal(horizon) || store.lastSliceValue != "123" || len(got.Points) != 2 || got.Points[1].MetricValue != nil {
		t.Fatalf("slice anchor/fill/null semantics = %#v through=%v", got, store.lastSliceThrough)
	}
}

func TestInsightExplorerCategoryCursorAndNoEntityDedupe(t *testing.T) {
	day := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	store := &fakeInsightsStore{changes: []v2models.InsightChangeRecord{
		{EntityID: 1, DetectorKey: "game_price_transition", DetectorVersion: 1, EventCode: "game_price_decreased", ProjectionDate: day, TimeBasis: "day", PrecisionRank: 0, EventSortAt: day, OpaqueTie: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		{EntityID: 1, DetectorKey: "game_price_transition", DetectorVersion: 1, EventCode: "game_price_increased", ProjectionDate: day.AddDate(0, 0, -1), TimeBasis: "day", PrecisionRank: 0, EventSortAt: day.AddDate(0, 0, -1), OpaqueTie: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		{EntityID: 2, DetectorKey: "game_price_transition", DetectorVersion: 1, EventCode: "game_price_state_changed", ProjectionDate: day.AddDate(0, 0, -2), TimeBasis: "day", PrecisionRank: 0, EventSortAt: day.AddDate(0, 0, -2), OpaqueTie: "cccccccccccccccccccccccccccccccc"},
	}}
	service := NewInsightsService(store)
	service.now = func() time.Time { return day }
	first, err := service.GetInsightsChanges(context.Background(), v2models.InsightChangeExplorerQuery{Range: "30d", Category: "price", Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Items) != 2 || first.Items[0].Entity.ID != first.Items[1].Entity.ID || first.Items[0].Category != "price" || first.NextCursor == nil {
		t.Fatalf("explorer category/dedupe = %#v", first)
	}
	payload, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"detector_key", "detector_version", "event_code", "event_key", "source_key", "game_price_transition"} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("explorer leaked %q: %s", forbidden, payload)
		}
	}
	service.now = func() time.Time { return day.AddDate(0, 0, 1) }
	_, err = service.GetInsightsChanges(context.Background(), v2models.InsightChangeExplorerQuery{Range: "30d", Category: "price", Cursor: *first.NextCursor, Limit: 2})
	if err != nil || !store.lastExplorer.RangeThrough.Equal(day) || store.lastExplorer.Position == nil {
		t.Fatalf("cursor anchor = err=%v conditions=%#v", err, store.lastExplorer)
	}
	if _, err := service.GetInsightsChanges(context.Background(), v2models.InsightChangeExplorerQuery{Range: "30d", Category: "discount", Cursor: *first.NextCursor, Limit: 2}); !errors.Is(err, ErrInvalidInsightCursor) {
		t.Fatalf("cursor/filter mismatch = %v", err)
	}
	if contracts, ok := insightExplorerContracts("", "game.discount.started"); !ok || len(contracts) != 1 || contracts[0].category != "discount" {
		t.Fatalf("exact public type filter = %#v, %v", contracts, ok)
	}
}

func TestRegionalPricesPreserveAvailabilityStatesPricedZeroAndObservedLow(t *testing.T) {
	day := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	priced, free := "priced", "free"
	cny := "CNY"
	zero := int64(0)
	hundred := int32(100)
	store := &fakeInsightsStore{
		game:  &v2models.InsightGameRecord{ID: 1, Name: "Game"},
		state: &v2models.InsightGameStateRecord{GameID: 1, FactDate: day, TrackingPeriodID: 9, Mac: boolPointer(true)},
		regional: []v2models.InsightRegionalPriceRecord{
			{Region: "CN", Available: true, FactDate: day, State: &priced, Currency: &cny, InitialAmount: &zero, FinalAmount: &zero, DiscountPercent: &hundred},
			{Region: "US", Available: true, FactDate: day, State: &free},
			{Region: "HK", Available: false},
		},
		observedLow: &v2models.InsightObservedLowRecord{Amount: 0, Currency: "CNY", FirstSeen: day, ObservedSince: day, InitialAmount: 0, DiscountPercent: 100},
		summaries:   map[string]*v2models.InsightMetricSummaryRecord{},
	}
	got, err := NewInsightsService(store).GetGameInsights(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.RegionalPrices.Regions) != 3 || got.Price == nil || got.Price.State != "priced" || got.Price.FinalAmount == nil || *got.Price.FinalAmount != 0 {
		t.Fatalf("regional/CN compatibility = %#v", got)
	}
	if got.State.Mac == nil || !*got.State.Mac {
		t.Fatalf("entity Mac state missing: %#v", got.State)
	}
	if got.RegionalPrices.Regions[0].ObservedLow == nil || got.RegionalPrices.Regions[0].ObservedLow.Amount != 0 || got.RegionalPrices.Regions[1].State == nil || *got.RegionalPrices.Regions[1].State != "free" || got.RegionalPrices.Regions[2].State != nil {
		t.Fatalf("priced zero/free/unavailable semantics = %#v", got.RegionalPrices)
	}
	if _, err := NewInsightsService(store).GetGamePriceInsights(context.Background(), 1, "JP", "30d"); !errors.Is(err, ErrInvalidInsightRegion) {
		t.Fatalf("invalid region = %v", err)
	}
}

func TestPlayerRankingRealZeroQualityAndValidation(t *testing.T) {
	day := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	coverage := .5
	store := &fakeInsightsStore{
		rankingMeta: v2models.InsightPlayerRankingMetaRecord{WindowFrom: &day, WindowThrough: &day, Population: 2, Ranked: 1},
		rankingRows: []v2models.InsightPlayerRankingRecord{{GameID: 1, GameName: "Zero", Value: 0, SampleCoverage: &coverage}},
	}
	got, err := NewInsightsService(store).GetPlayerRanking(context.Background(), v2models.InsightPlayerRankingQuery{Metric: "average_30d", Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 || got.Items[0].Value != 0 || got.EntityCoverage == nil || *got.EntityCoverage != .5 || got.Items[0].SampleCoverage == nil {
		t.Fatalf("ranking zero/coverage = %#v", got)
	}
	if _, err := NewInsightsService(store).GetPlayerRanking(context.Background(), v2models.InsightPlayerRankingQuery{Metric: "growth"}); !errors.Is(err, ErrInvalidPlayerRanking) {
		t.Fatalf("invalid metric = %v", err)
	}
}

func TestPriceOverviewAndLanguageMathUseCorrectDenominators(t *testing.T) {
	day := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	store := &fakeInsightsStore{
		priceOverview:    v2models.InsightPriceOverviewRecord{AsOf: &day, Population: 10, Priced: 4, Free: 1, Unpriced: 1, Unknown: 2, Unavailable: 2, Discounted: 2},
		languageOverview: v2models.InsightLanguageOverviewRecord{AsOf: &day, Population: 10, Fresh: 5, Stale: 3, Unobserved: 2, FullyNormalizedGames: 4, UnmappedGames: 1, UnmappedEntries: 2},
		languages:        []v2models.InsightLanguageRecord{{Code: "en", SteamName: "English", SupportedGames: 5, ExplicitFullAudioGames: 2}},
	}
	price, err := NewInsightsService(store).GetPriceOverview(context.Background(), "CN")
	if err != nil || price.Known != 6 || price.Coverage == nil || *price.Coverage != .6 || price.DiscountedShare == nil || *price.DiscountedShare != .5 {
		t.Fatalf("price overview = %#v err=%v", price, err)
	}
	languages, err := NewInsightsService(store).GetLanguageOverview(context.Background())
	if err != nil || languages.Coverage == nil || *languages.Coverage != .5 || languages.Items[0].Share == nil || *languages.Items[0].Share != 1 || languages.Items[0].ExplicitFullAudioShare == nil || *languages.Items[0].ExplicitFullAudioShare != .4 {
		t.Fatalf("language overview = %#v err=%v", languages, err)
	}
}

func boolPointer(value bool) *bool { return &value }
