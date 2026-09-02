package dao

import (
	"context"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

func (d *InsightsDAO) GetInsightGameCompareFactHorizon(ctx context.Context, gameIDs []int64) (*time.Time, error) {
	value, err := d.queries.GetGameInsightCompareFactHorizon(ctx, gameIDs)
	if err != nil {
		return nil, err
	}
	return datePointer(value), nil
}

func (d *InsightsDAO) ListInsightGameCompareFacts(ctx context.Context, gameIDs []int64, factDate time.Time, region string) ([]v2models.InsightGameCompareFactRecord, error) {
	rows, err := d.queries.ListGameInsightCompareFacts(ctx, gamesqlc.ListGameInsightCompareFactsParams{
		FactDate: pgtype.Date{Time: factDate, Valid: true}, Region: region, GameIds: gameIDs,
	})
	if err != nil {
		return nil, err
	}
	result := make([]v2models.InsightGameCompareFactRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, v2models.InsightGameCompareFactRecord{
			GameID: row.GameID, TrackingPeriodID: row.TrackingPeriodID,
			Free: row.IsFree, Windows: row.Windows, Mac: row.Mac, Linux: row.Linux, Release: row.ReleaseAvailability,
			LanguageEvidence: row.LanguageEvidence, LanguageCodes: row.LanguageCodes,
			FullAudioLanguageCodes: row.FullAudioLanguageCodes, UnknownLanguageNames: row.UnknownLanguageNames,
			PriceAvailable: row.PriceAvailable, PriceState: row.PriceState, Currency: row.Currency,
			InitialAmount: row.InitialAmount, FinalAmount: row.FinalAmount, DiscountPercent: row.DiscountPercent,
		})
	}
	return result, nil
}

func (d *InsightsDAO) ListInsightGameCompareCurrentPlayers(ctx context.Context, gameIDs []int64) ([]v2models.InsightGameCompareCurrentPlayerRecord, error) {
	rows, err := d.queries.ListGameInsightCompareCurrentPlayers(ctx, gameIDs)
	if err != nil {
		return nil, err
	}
	result := make([]v2models.InsightGameCompareCurrentPlayerRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, v2models.InsightGameCompareCurrentPlayerRecord{
			GameID: row.GameID, SnapshotScheduledFor: timestampPointer(row.SnapshotScheduledFor),
			Available: row.Available, PlayerCount: row.PlayerCount, CollectedAt: timestampPointer(row.CollectedAt),
		})
	}
	return result, nil
}

func (d *InsightsDAO) ListInsightGameComparePlayer30D(ctx context.Context, gameIDs []int64) ([]v2models.InsightGameComparePlayer30DRecord, error) {
	rows, err := d.queries.ListGameInsightComparePlayer30d(ctx, gameIDs)
	if err != nil {
		return nil, err
	}
	result := make([]v2models.InsightGameComparePlayer30DRecord, 0, len(rows))
	for _, row := range rows {
		record := v2models.InsightGameComparePlayer30DRecord{
			GameID: row.GameID, FactThrough: datePointer(row.FactThrough), EligibleFrom: datePointer(row.EligibleFrom),
			ObservedDays: row.ObservedDays, SuccessfulSamples: row.SuccessfulSamples,
		}
		if row.SuccessfulSamples > 0 {
			record.Peak30D = int64Pointer(row.Peak30d)
			record.Average30D = float64Pointer(row.Average30d)
		}
		if row.HasSampleCoverage {
			record.SampleCoverage = float64Pointer(row.SampleCoverage)
		}
		result = append(result, record)
	}
	return result, nil
}
