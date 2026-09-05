package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
)

var (
	ErrInvalidInsightMetric    = errors.New("invalid public metric key")
	ErrInvalidInsightRange     = errors.New("invalid insights range")
	ErrInvalidInsightDimension = errors.New("invalid public dimension")
	ErrInvalidInsightSlice     = errors.New("invalid public dimension slice")
	ErrInvalidInsightChanges   = errors.New("invalid change explorer query")
	ErrInvalidInsightCursor    = errors.New("invalid change explorer cursor")
	ErrInvalidInsightRegion    = errors.New("invalid insights region")
	ErrInvalidPlayerRanking    = errors.New("invalid player ranking query")
	ErrInvalidInsightLimit     = errors.New("invalid insights limit")
	ErrInvalidInsightCompare   = errors.New("invalid game comparison")
	ErrInsightGameNotFound     = errors.New("game not found")
)

var insightMetricContracts = []v2models.InsightMetricContract{
	{PublicKey: "free", InternalKey: "free_game_share", Version: 1},
	{PublicKey: "windows", InternalKey: "windows_support", Version: 1},
	{PublicKey: "mac", InternalKey: "mac_support", Version: 1},
	{PublicKey: "linux", InternalKey: "linux_support", Version: 1},
}

type insightChangeContract struct {
	detector string
	version  int32
	code     string
	public   string
	category string
}

var insightChangeContracts = []insightChangeContract{
	{"free_game_transition", 1, "game_became_free", "game.free.enabled", "pricing_model"},
	{"free_game_transition", 1, "game_became_paid", "game.free.disabled", "pricing_model"},
	{"windows_support_transition", 1, "windows_support_added", "game.windows.added", "platform"},
	{"windows_support_transition", 1, "windows_support_removed", "game.windows.removed", "platform"},
	{"linux_support_transition", 1, "linux_support_added", "game.linux.added", "platform"},
	{"linux_support_transition", 1, "linux_support_removed", "game.linux.removed", "platform"},
	{"mac_support_transition", 1, "mac_support_added", "game.mac.added", "platform"},
	{"mac_support_transition", 1, "mac_support_removed", "game.mac.removed", "platform"},
	{"game_release_transition", 1, "game_became_available", "game.release.available", "release"},
	{"game_release_transition", 1, "game_availability_withdrawn", "game.release.withdrawn", "release"},
	{"game_release_transition", 1, "game_release_plan_changed", "game.release.changed", "release"},
	{"game_price_transition", 1, "game_price_increased", "game.price.increased", "price"},
	{"game_price_transition", 1, "game_price_decreased", "game.price.decreased", "price"},
	{"game_price_transition", 1, "game_price_state_changed", "game.price.state_changed", "price"},
	{"game_price_transition", 1, "game_price_currency_changed", "game.price.currency_changed", "price"},
	{"game_price_transition", 1, "game_discount_started", "game.discount.started", "discount"},
	{"game_price_transition", 1, "game_discount_ended", "game.discount.ended", "discount"},
	{"game_price_transition", 1, "game_discount_changed", "game.discount.changed", "discount"},
}

type InsightsStore interface {
	CountInsightEntities(context.Context) (int64, error)
	GetInsightGame(context.Context, int64) (*v2models.InsightGameRecord, error)
	GetInsightMetricSummary(context.Context, v2models.InsightMetricContract) (*v2models.InsightMetricSummaryRecord, error)
	ListInsightMetricTrend(context.Context, v2models.InsightMetricContract, int32) ([]v2models.InsightMetricTrendRecord, error)
	ListInsightMetricBreakdown(context.Context, v2models.InsightMetricContract, v2models.InsightDimensionContract, time.Time) ([]v2models.InsightDimensionRecord, error)
	GetInsightMetricSliceAvailability(context.Context, v2models.InsightMetricContract, v2models.InsightDimensionContract, string) (v2models.InsightDimensionAvailabilityRecord, error)
	ListInsightMetricSliceTrend(context.Context, v2models.InsightMetricContract, v2models.InsightDimensionContract, string, time.Time, int32) ([]v2models.InsightDimensionTrendRecord, error)
	GetInsightGameState(context.Context, int64) (*v2models.InsightGameStateRecord, error)
	GetInsightGameCompareFactHorizon(context.Context, []int64) (*time.Time, error)
	ListInsightGameCompareFacts(context.Context, []int64, time.Time, string) ([]v2models.InsightGameCompareFactRecord, error)
	ListInsightGameCompareCurrentPlayers(context.Context, []int64) ([]v2models.InsightGameCompareCurrentPlayerRecord, error)
	ListInsightGameComparePlayer30D(context.Context, []int64) ([]v2models.InsightGameComparePlayer30DRecord, error)
	GetInsightPlayerSummary(context.Context, v2models.InsightGameStateRecord) (v2models.InsightPlayerSummaryRecord, error)
	ListInsightPlayerHistory(context.Context, int64, int32) ([]v2models.InsightPlayerPointRecord, error)
	ListInsightPriceHistory(context.Context, int64, string, time.Time, int32) ([]v2models.InsightPriceRecord, error)
	ListInsightRegionalPrices(context.Context, int64, time.Time) ([]v2models.InsightRegionalPriceRecord, error)
	GetInsightObservedLow(context.Context, int64, string, time.Time) (*v2models.InsightObservedLowRecord, error)
	GetLatestPlayerRankingMeta(context.Context) (v2models.InsightPlayerRankingMetaRecord, error)
	ListLatestPlayerRanking(context.Context, int32) ([]v2models.InsightPlayerRankingRecord, error)
	GetPlayer30DRankingMeta(context.Context) (v2models.InsightPlayerRankingMetaRecord, error)
	ListPlayer30DRanking(context.Context, bool, int32) ([]v2models.InsightPlayerRankingRecord, error)
	GetInsightPriceOverview(context.Context, string) (v2models.InsightPriceOverviewRecord, error)
	ListInsightDiscounts(context.Context, string, int32) ([]v2models.InsightDiscountRecord, error)
	GetInsightLanguageOverview(context.Context) (v2models.InsightLanguageOverviewRecord, error)
	ListInsightLanguages(context.Context) ([]v2models.InsightLanguageRecord, error)
	CountInsightOverviewChanges(context.Context, []string, []string) (int64, error)
	ListInsightOverviewChanges(context.Context, []string, []string, int32) ([]v2models.InsightChangeRecord, error)
	ListInsightExplorerChanges(context.Context, v2models.InsightChangeExplorerConditions) ([]v2models.InsightChangeRecord, error)
	ListInsightGameChanges(context.Context, v2models.InsightGameRecord, []string, []string, int32) ([]v2models.InsightChangeRecord, error)
}

type InsightsService struct {
	store InsightsStore
	now   func() time.Time
}

func NewInsightsService(store InsightsStore) *InsightsService {
	return &InsightsService{store: store, now: func() time.Time { return time.Now().UTC() }}
}

func (s *InsightsService) GetInsightsOverview(ctx context.Context) (v2models.InsightOverview, error) {
	result := v2models.InsightOverview{GeneratedAt: s.now().UTC(), Metrics: []v2models.InsightMetric{}, RecentChanges: []v2models.InsightChange{}}
	var err error
	if result.EntityCount, err = s.store.CountInsightEntities(ctx); err != nil {
		return result, err
	}
	detectors, contracts := insightChangeQueryContracts()
	if result.Changes7D, err = s.store.CountInsightOverviewChanges(ctx, detectors, contracts); err != nil {
		return result, err
	}
	for _, contract := range insightMetricContracts {
		record, queryErr := s.store.GetInsightMetricSummary(ctx, contract)
		if queryErr != nil {
			return result, queryErr
		}
		if record != nil {
			result.Metrics = append(result.Metrics, insightPublicMetric(contract.PublicKey, *record))
		}
	}
	changes, err := s.store.ListInsightOverviewChanges(ctx, detectors, contracts, 8)
	if err != nil {
		return result, err
	}
	result.RecentChanges = insightPublicChanges(changes)
	return result, nil
}

func (s *InsightsService) GetInsightsMetricTrend(ctx context.Context, publicKey, requestedRange string) (v2models.InsightMetricTrend, error) {
	result := v2models.InsightMetricTrend{Key: publicKey, RequestedRange: requestedRange, Points: []v2models.InsightTrendPoint{}}
	contract, ok := resolveInsightMetric(publicKey)
	if !ok {
		return result, ErrInvalidInsightMetric
	}
	rangeDays, ok := parseInsightRange(requestedRange)
	if !ok {
		return result, ErrInvalidInsightRange
	}
	summary, err := s.store.GetInsightMetricSummary(ctx, contract)
	if err != nil {
		return result, err
	}
	rows, err := s.store.ListInsightMetricTrend(ctx, contract, rangeDays)
	if err != nil {
		return result, err
	}
	if summary != nil {
		result.AvailableFrom = insightDateStringPointer(summary.AvailableFrom)
		through := insightFormatDate(summary.FactDate)
		result.AvailableThrough = &through
	}
	for _, row := range rows {
		known := row.PositiveCount + row.NegativeCount
		result.Points = append(result.Points, v2models.InsightTrendPoint{
			Date: insightFormatDate(row.FactDate), Value: insightRatio(row.PositiveCount, known),
			Coverage: insightRatio(known, row.EligibleCount),
		})
	}
	return result, nil
}

func (s *InsightsService) GetGameInsights(ctx context.Context, gameID int64) (v2models.GameInsights, error) {
	result := v2models.GameInsights{RecentChanges: []v2models.InsightChange{}}
	game, err := s.requireGame(ctx, gameID)
	if err != nil {
		return result, err
	}
	result.Game = v2models.InsightEntityRef{ID: game.ID, Name: game.Name}
	state, err := s.store.GetInsightGameState(ctx, gameID)
	if err != nil {
		return result, err
	}
	if state != nil {
		asOf := insightFormatDate(state.FactDate)
		result.State = v2models.InsightGameState{
			Free: state.Free, Windows: state.Windows, Mac: state.Mac, Linux: state.Linux, Release: state.Release, AsOf: &asOf,
		}
		players, queryErr := s.store.GetInsightPlayerSummary(ctx, *state)
		if queryErr != nil {
			return result, queryErr
		}
		if players.HasCurrent {
			value := players.Current
			result.Players.Current = &value
			result.Players.AsOf = players.CurrentAt
		}
		result.Players.Peak30D = players.Peak30D
		result.Players.Average30D = players.Average30D
		result.Players.FactThrough = insightDateStringPointer(players.FactThrough)
		result.Players.EligibleFrom30D = insightDateStringPointer(players.EligibleFrom)
		result.Players.ObservedDays30D = players.ObservedDays
		result.Players.SuccessfulSamples30D = players.SuccessfulSamples
		result.Players.SampleCoverage30D = players.SampleCoverage
		regional, queryErr := s.regionalPrices(ctx, *state)
		if queryErr != nil {
			return result, queryErr
		}
		result.RegionalPrices = regional
		result.Price = compatibilityCNPrice(regional)
	}
	detectors, contracts := insightChangeQueryContracts()
	changes, err := s.store.ListInsightGameChanges(ctx, *game, detectors, contracts, 20)
	if err != nil {
		return result, err
	}
	result.RecentChanges = insightPublicChanges(changes)
	return result, nil
}

func (s *InsightsService) GetGamePlayerInsights(ctx context.Context, gameID int64, requestedRange string) (v2models.InsightPlayerHistory, error) {
	result := v2models.InsightPlayerHistory{RequestedRange: requestedRange, Points: []v2models.InsightPlayerPoint{}}
	rangeDays, ok := parseInsightRange(requestedRange)
	if !ok {
		return result, ErrInvalidInsightRange
	}
	if _, err := s.requireGame(ctx, gameID); err != nil {
		return result, err
	}
	state, err := s.store.GetInsightGameState(ctx, gameID)
	if err != nil || state == nil {
		return result, err
	}
	rows, err := s.store.ListInsightPlayerHistory(ctx, state.TrackingPeriodID, rangeDays)
	if err != nil {
		return result, err
	}
	for _, row := range rows {
		result.Points = append(result.Points, v2models.InsightPlayerPoint{
			Date: insightFormatDate(row.FactDate), Min: row.Min, Max: row.Max, Avg: row.Avg,
		})
	}
	setPlayerAvailability(&result)
	return result, nil
}

func (s *InsightsService) GetGamePriceInsights(ctx context.Context, gameID int64, region, requestedRange string) (v2models.InsightPriceHistory, error) {
	result := v2models.InsightPriceHistory{Region: region, RequestedRange: requestedRange, Points: []v2models.InsightPricePoint{}}
	if !validInsightRegion(region) {
		return result, ErrInvalidInsightRegion
	}
	rangeDays, ok := parseInsightRange(requestedRange)
	if !ok {
		return result, ErrInvalidInsightRange
	}
	if _, err := s.requireGame(ctx, gameID); err != nil {
		return result, err
	}
	state, err := s.store.GetInsightGameState(ctx, gameID)
	if err != nil || state == nil {
		return result, err
	}
	rows, err := s.store.ListInsightPriceHistory(ctx, state.TrackingPeriodID, region, state.FactDate, rangeDays)
	if err != nil {
		return result, err
	}
	for _, row := range rows {
		result.Points = append(result.Points, v2models.InsightPricePoint{
			Date: insightFormatDate(row.FactDate), State: row.State, Currency: row.Currency,
			InitialAmount: row.InitialAmount, FinalAmount: row.FinalAmount, DiscountPercent: row.DiscountPercent,
		})
	}
	setPriceAvailability(&result)
	return result, nil
}

func (s *InsightsService) requireGame(ctx context.Context, gameID int64) (*v2models.InsightGameRecord, error) {
	game, err := s.store.GetInsightGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, ErrInsightGameNotFound
	}
	return game, nil
}

func insightPublicMetric(key string, record v2models.InsightMetricSummaryRecord) v2models.InsightMetric {
	known := record.PositiveCount + record.NegativeCount
	metric := v2models.InsightMetric{
		Key: key, AsOf: insightFormatDate(record.FactDate), Value: insightRatio(record.PositiveCount, known),
		Coverage: insightRatio(known, record.EligibleCount), Known: known, Eligible: record.EligibleCount,
		AvailableFrom: insightDateStringPointer(record.AvailableFrom),
	}
	if record.PreviousPositiveCount != nil && record.PreviousNegativeCount != nil {
		previousKnown := *record.PreviousPositiveCount + *record.PreviousNegativeCount
		currentValue := insightRatio(record.PositiveCount, known)
		previousValue := insightRatio(*record.PreviousPositiveCount, previousKnown)
		if currentValue != nil && previousValue != nil {
			delta := *currentValue - *previousValue
			metric.Delta30D = &delta
		}
	}
	return metric
}

func resolveInsightMetric(publicKey string) (v2models.InsightMetricContract, bool) {
	for _, contract := range insightMetricContracts {
		if contract.PublicKey == publicKey {
			return contract, true
		}
	}
	return v2models.InsightMetricContract{}, false
}

func parseInsightRange(value string) (int32, bool) {
	switch value {
	case "30d":
		return 30, true
	case "90d":
		return 90, true
	case "180d":
		return 180, true
	case "1y":
		return 365, true
	case "3y":
		return 1095, true
	case "5y":
		return 1825, true
	case "all":
		return 0, true
	default:
		return 0, false
	}
}

func insightPublicPrice(record v2models.InsightPriceRecord) *v2models.InsightPrice {
	return &v2models.InsightPrice{
		Region: "CN", State: record.State, Currency: record.Currency, InitialAmount: record.InitialAmount,
		FinalAmount: record.FinalAmount, DiscountPercent: record.DiscountPercent, AsOf: insightFormatDate(record.FactDate),
	}
}

func insightPublicChanges(records []v2models.InsightChangeRecord) []v2models.InsightChange {
	result := make([]v2models.InsightChange, 0, len(records))
	for _, record := range records {
		publicType, ok := insightPublicChangeType(record)
		if !ok {
			continue
		}
		var occurredAt *time.Time
		if record.TimeBasis != "day" {
			occurredAt = record.EventAt
		}
		result = append(result, v2models.InsightChange{
			Type: publicType, Date: insightFormatDate(record.ProjectionDate), OccurredAt: occurredAt,
			Entity: v2models.InsightEntityRef{ID: record.EntityID, Name: record.EntityName}, Detail: nil,
		})
	}
	return result
}

func insightPublicChangeType(record v2models.InsightChangeRecord) (string, bool) {
	for _, contract := range insightChangeContracts {
		if contract.detector == record.DetectorKey && contract.version == record.DetectorVersion && contract.code == record.EventCode {
			return contract.public, true
		}
	}
	return "", false
}

func insightChangeQueryContracts() ([]string, []string) {
	detectorSet := map[string]struct{}{}
	detectors := []string{}
	contracts := []string{}
	for _, contract := range insightChangeContracts {
		if _, exists := detectorSet[contract.detector]; !exists {
			detectorSet[contract.detector] = struct{}{}
			detectors = append(detectors, contract.detector)
		}
		contracts = append(contracts, fmt.Sprintf("%s/%d/%s", contract.detector, contract.version, contract.code))
	}
	return detectors, contracts
}

func setPlayerAvailability(result *v2models.InsightPlayerHistory) {
	if len(result.Points) == 0 {
		return
	}
	from, through := result.Points[0].Date, result.Points[len(result.Points)-1].Date
	result.AvailableFrom, result.AvailableThrough = &from, &through
}

func setPriceAvailability(result *v2models.InsightPriceHistory) {
	if len(result.Points) == 0 {
		return
	}
	from, through := result.Points[0].Date, result.Points[len(result.Points)-1].Date
	result.AvailableFrom, result.AvailableThrough = &from, &through
}

func insightRatio(numerator, denominator int64) *float64 {
	if denominator == 0 {
		return nil
	}
	value := float64(numerator) / float64(denominator)
	return &value
}

func insightDateStringPointer(value *time.Time) *string {
	if value == nil {
		return nil
	}
	result := insightFormatDate(*value)
	return &result
}

func insightFormatDate(value time.Time) string { return value.UTC().Format("2006-01-02") }
