package service

import (
	"context"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
)

var insightRegions = []string{"CN", "US", "HK"}

func validInsightRegion(region string) bool {
	for _, candidate := range insightRegions {
		if region == candidate {
			return true
		}
	}
	return false
}

func (s *InsightsService) regionalPrices(ctx context.Context, state v2models.InsightGameStateRecord) (v2models.InsightRegionalPrices, error) {
	asOf := insightFormatDate(state.FactDate)
	result := v2models.InsightRegionalPrices{AsOf: &asOf, Regions: []v2models.InsightRegionalPrice{}}
	rows, err := s.store.ListInsightRegionalPrices(ctx, state.TrackingPeriodID, state.FactDate)
	if err != nil {
		return result, err
	}
	for _, row := range rows {
		item := v2models.InsightRegionalPrice{
			Region: row.Region, Available: row.Available, State: row.State, Currency: row.Currency,
			InitialAmount: row.InitialAmount, FinalAmount: row.FinalAmount, DiscountPercent: row.DiscountPercent,
		}
		if row.Available && row.State != nil && *row.State == "priced" && row.Currency != nil {
			low, queryErr := s.store.GetInsightObservedLow(ctx, state.TrackingPeriodID, row.Region, state.FactDate)
			if queryErr != nil {
				return result, queryErr
			}
			item.ObservedLow = publicObservedLow(low)
		}
		result.Regions = append(result.Regions, item)
	}
	return result, nil
}

func compatibilityCNPrice(regional v2models.InsightRegionalPrices) *v2models.InsightPrice {
	for _, item := range regional.Regions {
		if item.Region != "CN" || !item.Available || item.State == nil || regional.AsOf == nil {
			continue
		}
		return &v2models.InsightPrice{
			Region: "CN", State: *item.State, Currency: item.Currency, InitialAmount: item.InitialAmount,
			FinalAmount: item.FinalAmount, DiscountPercent: item.DiscountPercent, AsOf: *regional.AsOf,
		}
	}
	return nil
}

func publicObservedLow(record *v2models.InsightObservedLowRecord) *v2models.InsightObservedLow {
	if record == nil {
		return nil
	}
	return &v2models.InsightObservedLow{
		Amount: record.Amount, Currency: record.Currency, FirstSeen: insightFormatDate(record.FirstSeen),
		ObservedSince: insightFormatDate(record.ObservedSince), InitialAmount: record.InitialAmount,
		DiscountPercent: record.DiscountPercent,
	}
}

func (s *InsightsService) GetPlayerRanking(ctx context.Context, query v2models.InsightPlayerRankingQuery) (v2models.InsightPlayerRanking, error) {
	if query.Metric == "" {
		query.Metric = "latest_observed"
	}
	if query.Limit == 0 {
		query.Limit = 20
	}
	result := v2models.InsightPlayerRanking{Metric: query.Metric, Items: []v2models.InsightPlayerRankingItem{}}
	if query.Limit < 1 || query.Limit > 100 {
		return result, ErrInvalidInsightLimit
	}
	if query.Metric != "latest_observed" && query.Metric != "peak_30d" && query.Metric != "average_30d" {
		return result, ErrInvalidPlayerRanking
	}
	var meta v2models.InsightPlayerRankingMetaRecord
	var rows []v2models.InsightPlayerRankingRecord
	var err error
	if query.Metric == "latest_observed" {
		result.Basis = "scheduled_snapshot"
		meta, err = s.store.GetLatestPlayerRankingMeta(ctx)
		if err == nil {
			rows, err = s.store.ListLatestPlayerRanking(ctx, query.Limit)
		}
	} else {
		result.Basis = "finalized_daily_facts"
		meta, err = s.store.GetPlayer30DRankingMeta(ctx)
		if err == nil {
			rows, err = s.store.ListPlayer30DRanking(ctx, query.Metric == "average_30d", query.Limit)
		}
	}
	if err != nil {
		return result, err
	}
	result.SnapshotScheduledFor = meta.SnapshotScheduledFor
	result.LatestSlotScheduledFor = meta.LatestSlotScheduledFor
	result.ObservedFrom, result.ObservedThrough = meta.ObservedFrom, meta.ObservedThrough
	result.WindowFrom = insightDateStringPointer(meta.WindowFrom)
	result.WindowThrough = insightDateStringPointer(meta.WindowThrough)
	result.Population, result.Ranked = meta.Population, meta.Ranked
	result.EntityCoverage = insightRatio(meta.Ranked, meta.Population)
	for index, row := range rows {
		item := v2models.InsightPlayerRankingItem{
			Rank: int32(index + 1), Game: v2models.InsightEntityRef{ID: row.GameID, Name: row.GameName},
			Value: row.Value, ObservedAt: row.ObservedAt, EligibleFrom: insightDateStringPointer(row.EligibleFrom),
			ObservedDays: row.ObservedDays, SuccessfulSamples: row.SuccessfulSamples, SampleCoverage: row.SampleCoverage,
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func (s *InsightsService) GetPriceOverview(ctx context.Context, region string) (v2models.InsightPriceOverview, error) {
	result := v2models.InsightPriceOverview{Region: region}
	if !validInsightRegion(region) {
		return result, ErrInvalidInsightRegion
	}
	record, err := s.store.GetInsightPriceOverview(ctx, region)
	if err != nil {
		return result, err
	}
	result.AsOf = insightDateStringPointer(record.AsOf)
	result.Population, result.Priced, result.Free = record.Population, record.Priced, record.Free
	result.Unpriced, result.Unknown, result.Unavailable = record.Unpriced, record.Unknown, record.Unavailable
	result.Known = record.Priced + record.Free + record.Unpriced
	result.Coverage = insightRatio(result.Known, result.Population)
	result.Discounted = record.Discounted
	result.DiscountedShare = insightRatio(record.Discounted, record.Priced)
	return result, nil
}

func (s *InsightsService) GetDiscounts(ctx context.Context, region string, limit int32) (v2models.InsightDiscounts, error) {
	result := v2models.InsightDiscounts{Region: region, Items: []v2models.InsightDiscountItem{}}
	if !validInsightRegion(region) {
		return result, ErrInvalidInsightRegion
	}
	if limit == 0 {
		limit = 20
	}
	if limit < 1 || limit > 100 {
		return result, ErrInvalidInsightLimit
	}
	rows, err := s.store.ListInsightDiscounts(ctx, region, limit)
	if err != nil {
		return result, err
	}
	for _, row := range rows {
		if result.AsOf == nil {
			asOf := insightFormatDate(row.AsOf)
			result.AsOf = &asOf
		}
		low, queryErr := s.store.GetInsightObservedLow(ctx, row.TrackingPeriodID, region, row.AsOf)
		if queryErr != nil {
			return result, queryErr
		}
		result.Items = append(result.Items, v2models.InsightDiscountItem{
			Game: v2models.InsightEntityRef{ID: row.GameID, Name: row.GameName}, Currency: row.Currency,
			InitialAmount: row.InitialAmount, FinalAmount: row.FinalAmount, DiscountPercent: row.DiscountPercent,
			ObservedLow: publicObservedLow(low),
		})
	}
	if result.AsOf == nil {
		overview, queryErr := s.store.GetInsightPriceOverview(ctx, region)
		if queryErr != nil {
			return result, queryErr
		}
		result.AsOf = insightDateStringPointer(overview.AsOf)
	}
	return result, nil
}

func (s *InsightsService) GetLanguageOverview(ctx context.Context) (v2models.InsightLanguageOverview, error) {
	result := v2models.InsightLanguageOverview{FreshnessSeconds: 259200, Items: []v2models.InsightLanguageItem{}}
	record, err := s.store.GetInsightLanguageOverview(ctx)
	if err != nil {
		return result, err
	}
	result.AsOf = insightDateStringPointer(record.AsOf)
	result.Population, result.Fresh, result.Stale, result.Unobserved = record.Population, record.Fresh, record.Stale, record.Unobserved
	result.Coverage = insightRatio(record.Fresh, record.Population)
	result.FullyNormalizedGames, result.UnmappedGames, result.UnmappedEntries = record.FullyNormalizedGames, record.UnmappedGames, record.UnmappedEntries
	result.NormalizationCoverage = insightRatio(record.FullyNormalizedGames, record.Fresh)
	rows, err := s.store.ListInsightLanguages(ctx)
	if err != nil {
		return result, err
	}
	for _, row := range rows {
		result.Items = append(result.Items, v2models.InsightLanguageItem{
			Code: row.Code, SteamName: row.SteamName, SupportedGames: row.SupportedGames,
			Share: insightRatio(row.SupportedGames, record.Fresh), ExplicitFullAudioGames: row.ExplicitFullAudioGames,
			ExplicitFullAudioShare: insightRatio(row.ExplicitFullAudioGames, record.Fresh),
		})
	}
	return result, nil
}
