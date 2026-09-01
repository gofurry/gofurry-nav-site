package dao

import (
	"context"
	"errors"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type InsightsDAO struct{ queries *gamesqlc.Queries }

func NewInsightsDAO(queries *gamesqlc.Queries) *InsightsDAO { return &InsightsDAO{queries: queries} }

func (d *InsightsDAO) CountInsightEntities(ctx context.Context) (int64, error) {
	return d.queries.CountGameInsightGames(ctx)
}

func (d *InsightsDAO) GetInsightGame(ctx context.Context, gameID int64) (*v2models.InsightGameRecord, error) {
	row, err := d.queries.GetGameInsightGame(ctx, gameID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	name := row.Name
	if name == "" {
		name = row.NameEn
	}
	return &v2models.InsightGameRecord{ID: row.ID, Name: name}, nil
}

func (d *InsightsDAO) GetInsightMetricSummary(ctx context.Context, contract v2models.InsightMetricContract) (*v2models.InsightMetricSummaryRecord, error) {
	row, err := d.queries.GetGameInsightMetricSummary(ctx, gamesqlc.GetGameInsightMetricSummaryParams{
		MetricKey: contract.InternalKey, MetricVersion: contract.Version,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v2models.InsightMetricSummaryRecord{
		FactDate: row.FactDate.Time, EligibleCount: row.EligibleCount,
		PositiveCount: row.PositiveCount, NegativeCount: row.NegativeCount,
		PreviousPositiveCount: row.PreviousPositiveCount, PreviousNegativeCount: row.PreviousNegativeCount,
		AvailableFrom: datePointer(row.AvailableFrom),
	}, nil
}

func (d *InsightsDAO) ListInsightMetricTrend(ctx context.Context, contract v2models.InsightMetricContract, rangeDays int32) ([]v2models.InsightMetricTrendRecord, error) {
	rows, err := d.queries.ListGameInsightMetricTrend(ctx, gamesqlc.ListGameInsightMetricTrendParams{
		MetricKey: contract.InternalKey, MetricVersion: contract.Version, RangeDays: rangeDays,
	})
	if err != nil {
		return nil, err
	}
	result := make([]v2models.InsightMetricTrendRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, v2models.InsightMetricTrendRecord{
			FactDate: row.FactDate.Time, EligibleCount: row.EligibleCount,
			PositiveCount: row.PositiveCount, NegativeCount: row.NegativeCount,
		})
	}
	return result, nil
}

func (d *InsightsDAO) ListInsightMetricBreakdown(ctx context.Context, contract v2models.InsightMetricContract, dimension v2models.InsightDimensionContract, factDate time.Time) ([]v2models.InsightDimensionRecord, error) {
	rows, err := d.queries.ListGameInsightMetricBreakdown(ctx, gamesqlc.ListGameInsightMetricBreakdownParams{
		MetricKey: contract.InternalKey, MetricVersion: contract.Version,
		FactDate: pgtype.Date{Time: factDate, Valid: true}, DimensionKey: dimension.InternalKey,
	})
	if err != nil {
		return nil, err
	}
	result := make([]v2models.InsightDimensionRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, v2models.InsightDimensionRecord{
			Value: row.DimensionValue, Label: nonemptyStringPointer(row.Label), LabelEn: nonemptyStringPointer(row.LabelEn),
			Population: row.PopulationCount, Eligible: row.EligibleCount,
			PositiveCount: row.PositiveCount, NegativeCount: row.NegativeCount,
		})
	}
	return result, nil
}

func (d *InsightsDAO) GetInsightMetricSliceAvailability(ctx context.Context, contract v2models.InsightMetricContract, dimension v2models.InsightDimensionContract, value string) (v2models.InsightDimensionAvailabilityRecord, error) {
	row, err := d.queries.GetGameInsightMetricSliceAvailability(ctx, gamesqlc.GetGameInsightMetricSliceAvailabilityParams{
		MetricKey: contract.InternalKey, MetricVersion: contract.Version,
		DimensionKey: dimension.InternalKey, DimensionValue: value,
	})
	if err != nil {
		return v2models.InsightDimensionAvailabilityRecord{}, err
	}
	return v2models.InsightDimensionAvailabilityRecord{
		Label: nonemptyStringPointer(row.Label), LabelEn: nonemptyStringPointer(row.LabelEn),
		AvailableFrom: datePointer(row.AvailableFrom), AvailableThrough: datePointer(row.AvailableThrough),
	}, nil
}

func (d *InsightsDAO) ListInsightMetricSliceTrend(ctx context.Context, contract v2models.InsightMetricContract, dimension v2models.InsightDimensionContract, value string, through time.Time, rangeDays int32) ([]v2models.InsightDimensionTrendRecord, error) {
	rows, err := d.queries.ListGameInsightMetricSliceTrend(ctx, gamesqlc.ListGameInsightMetricSliceTrendParams{
		MetricKey: contract.InternalKey, MetricVersion: contract.Version,
		DimensionKey: dimension.InternalKey, DimensionValue: value,
		ThroughDate: pgtype.Date{Time: through, Valid: true}, RangeDays: rangeDays,
	})
	if err != nil {
		return nil, err
	}
	result := make([]v2models.InsightDimensionTrendRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, v2models.InsightDimensionTrendRecord{
			FactDate: row.FactDate.Time, Population: row.PopulationCount, Eligible: row.EligibleCount,
			PositiveCount: row.PositiveCount, NegativeCount: row.NegativeCount,
		})
	}
	return result, nil
}

func (d *InsightsDAO) GetInsightGameState(ctx context.Context, gameID int64) (*v2models.InsightGameStateRecord, error) {
	row, err := d.queries.GetGameInsightState(ctx, gameID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v2models.InsightGameStateRecord{
		GameID: row.GameID, FactDate: row.FactDate.Time, TrackingPeriodID: row.TrackingPeriodID, AppID: row.Appid,
		Free: row.IsFree, Windows: row.Windows, Linux: row.Linux, Release: row.ReleaseAvailability,
	}, nil
}

func (d *InsightsDAO) GetInsightPlayerSummary(ctx context.Context, state v2models.InsightGameStateRecord) (v2models.InsightPlayerSummaryRecord, error) {
	row, err := d.queries.GetGameInsightPlayerSummary(ctx, gamesqlc.GetGameInsightPlayerSummaryParams{
		GameID: state.GameID, Appid: state.AppID, TrackingPeriodID: state.TrackingPeriodID,
		ThroughDate: pgtype.Date{Time: state.FactDate, Valid: true},
	})
	if err != nil {
		return v2models.InsightPlayerSummaryRecord{}, err
	}
	return v2models.InsightPlayerSummaryRecord{
		HasCurrent: row.HasCurrent, Current: row.CurrentPlayers, CurrentAt: timestampPointer(row.CurrentAsOf),
		HasPeak30D: row.HasPeak30d, Peak30D: row.Peak30d,
	}, nil
}

func (d *InsightsDAO) GetInsightPriceSummary(ctx context.Context, trackingPeriodID int64) (*v2models.InsightPriceRecord, error) {
	row, err := d.queries.GetGameInsightPriceSummary(ctx, trackingPeriodID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v2models.InsightPriceRecord{
		FactDate: row.FactDate.Time, State: row.PriceState, Currency: row.Currency,
		InitialAmount: row.InitialAmount, FinalAmount: row.FinalAmount, DiscountPercent: row.DiscountPercent,
	}, nil
}

func (d *InsightsDAO) ListInsightPlayerHistory(ctx context.Context, trackingPeriodID int64, rangeDays int32) ([]v2models.InsightPlayerPointRecord, error) {
	rows, err := d.queries.ListGameInsightPlayerHistory(ctx, gamesqlc.ListGameInsightPlayerHistoryParams{
		TrackingPeriodID: trackingPeriodID, RangeDays: rangeDays,
	})
	if err != nil {
		return nil, err
	}
	result := make([]v2models.InsightPlayerPointRecord, 0, len(rows))
	for _, row := range rows {
		if row.MaxPlayers == nil {
			continue
		}
		result = append(result, v2models.InsightPlayerPointRecord{
			FactDate: row.FactDate.Time, Min: row.MinPlayers, Max: *row.MaxPlayers, Avg: row.AvgPlayers,
		})
	}
	return result, nil
}

func (d *InsightsDAO) ListInsightPriceHistory(ctx context.Context, trackingPeriodID int64, rangeDays int32) ([]v2models.InsightPriceRecord, error) {
	rows, err := d.queries.ListGameInsightPriceHistory(ctx, gamesqlc.ListGameInsightPriceHistoryParams{
		TrackingPeriodID: trackingPeriodID, RangeDays: rangeDays,
	})
	if err != nil {
		return nil, err
	}
	result := make([]v2models.InsightPriceRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, v2models.InsightPriceRecord{
			FactDate: row.FactDate.Time, State: row.PriceState, Currency: row.Currency,
			InitialAmount: row.InitialAmount, FinalAmount: row.FinalAmount, DiscountPercent: row.DiscountPercent,
		})
	}
	return result, nil
}

func (d *InsightsDAO) CountInsightOverviewChanges(ctx context.Context, detectorKeys, contractIDs []string) (int64, error) {
	return d.queries.CountGameInsightOverviewChanges(ctx, gamesqlc.CountGameInsightOverviewChangesParams{
		DetectorKeys: detectorKeys, ContractIds: contractIDs,
	})
}

func (d *InsightsDAO) ListInsightOverviewChanges(ctx context.Context, detectorKeys, contractIDs []string, limit int32) ([]v2models.InsightChangeRecord, error) {
	rows, err := d.queries.ListGameInsightOverviewChanges(ctx, gamesqlc.ListGameInsightOverviewChangesParams{
		LimitCount: limit, DetectorKeys: detectorKeys, ContractIds: contractIDs,
	})
	if err != nil {
		return nil, err
	}
	result := make([]v2models.InsightChangeRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, v2models.InsightChangeRecord{
			EntityID: row.GameID, EntityName: row.GameName, DetectorKey: row.DetectorKey,
			DetectorVersion: row.DetectorVersion, EventCode: row.EventCode,
			ProjectionDate: row.ProjectionDate.Time, TimeBasis: row.TimeBasis, EventAt: timestampPointer(row.EventAt),
		})
	}
	return result, nil
}

func (d *InsightsDAO) ListInsightExplorerChanges(ctx context.Context, conditions v2models.InsightChangeExplorerConditions) ([]v2models.InsightChangeRecord, error) {
	positionDate := conditions.RangeThrough
	positionSortAt := conditions.RangeThrough
	positionRank := int32(0)
	positionTie := ""
	hasPosition := conditions.Position != nil
	if conditions.Position != nil {
		positionDate = conditions.Position.ProjectionDate
		positionSortAt = conditions.Position.EventSortAt
		positionRank = conditions.Position.PrecisionRank
		positionTie = conditions.Position.OpaqueTie
	}
	rows, err := d.queries.ListGameInsightExplorerChanges(ctx, gamesqlc.ListGameInsightExplorerChangesParams{
		DetectorKeys: conditions.DetectorKeys, ContractIds: conditions.ContractIDs,
		RangeThrough: pgtype.Date{Time: conditions.RangeThrough, Valid: true}, RangeDays: conditions.RangeDays,
		HasPosition: hasPosition, PositionDate: pgtype.Date{Time: positionDate, Valid: true},
		PositionRank: positionRank, PositionSortAt: pgtype.Timestamptz{Time: positionSortAt, Valid: true},
		PositionTie: positionTie, LimitCount: conditions.Limit,
	})
	if err != nil {
		return nil, err
	}
	result := make([]v2models.InsightChangeRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, v2models.InsightChangeRecord{
			EntityID: row.GameID, EntityName: row.GameName, DetectorKey: row.DetectorKey,
			DetectorVersion: row.DetectorVersion, EventCode: row.EventCode,
			ProjectionDate: row.ProjectionDate.Time, TimeBasis: row.TimeBasis, EventAt: timestampPointer(row.EventAt),
			PrecisionRank: row.PrecisionRank, EventSortAt: row.EventSortAt.Time, OpaqueTie: row.OpaqueTie,
		})
	}
	return result, nil
}

func (d *InsightsDAO) ListInsightGameChanges(ctx context.Context, game v2models.InsightGameRecord, detectorKeys, contractIDs []string, limit int32) ([]v2models.InsightChangeRecord, error) {
	rows, err := d.queries.ListGameInsightGameChanges(ctx, gamesqlc.ListGameInsightGameChangesParams{
		GameID: game.ID, DetectorKeys: detectorKeys, ContractIds: contractIDs, LimitCount: limit,
	})
	if err != nil {
		return nil, err
	}
	result := make([]v2models.InsightChangeRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, v2models.InsightChangeRecord{
			EntityID: game.ID, EntityName: game.Name, DetectorKey: row.DetectorKey,
			DetectorVersion: row.DetectorVersion, EventCode: row.EventCode,
			ProjectionDate: row.ProjectionDate.Time, TimeBasis: row.TimeBasis, EventAt: timestampPointer(row.EventAt),
		})
	}
	return result, nil
}

func datePointer(value pgtype.Date) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func nonemptyStringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
