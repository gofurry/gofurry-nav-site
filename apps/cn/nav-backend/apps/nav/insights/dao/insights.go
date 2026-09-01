package dao

import (
	"context"
	"errors"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/models"
	navsqlc "github.com/gofurry/gofurry-nav-backend/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type InsightsDAO struct{ queries *navsqlc.Queries }

func New(queries *navsqlc.Queries) *InsightsDAO { return &InsightsDAO{queries: queries} }

func (d *InsightsDAO) CountEntities(ctx context.Context) (int64, error) {
	return d.queries.CountNavInsightSites(ctx)
}

func (d *InsightsDAO) GetSite(ctx context.Context, siteID int64) (*models.SiteRecord, error) {
	row, err := d.queries.GetNavInsightSite(ctx, siteID)
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
	return &models.SiteRecord{ID: row.ID, Name: name}, nil
}

func (d *InsightsDAO) GetMetricSummary(ctx context.Context, contract models.MetricContract) (*models.MetricSummaryRecord, error) {
	row, err := d.queries.GetNavInsightMetricSummary(ctx, navsqlc.GetNavInsightMetricSummaryParams{
		MetricKey: contract.InternalKey, MetricVersion: contract.Version,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &models.MetricSummaryRecord{
		FactDate:              row.FactDate.Time,
		EligibleCount:         row.EligibleCount,
		PositiveCount:         row.PositiveCount,
		NegativeCount:         row.NegativeCount,
		PreviousPositiveCount: row.PreviousPositiveCount,
		PreviousNegativeCount: row.PreviousNegativeCount,
		AvailableFrom:         datePointer(row.AvailableFrom),
	}, nil
}

func (d *InsightsDAO) ListMetricTrend(ctx context.Context, contract models.MetricContract, rangeDays int32) ([]models.MetricTrendRecord, error) {
	rows, err := d.queries.ListNavInsightMetricTrend(ctx, navsqlc.ListNavInsightMetricTrendParams{
		MetricKey: contract.InternalKey, MetricVersion: contract.Version, RangeDays: rangeDays,
	})
	if err != nil {
		return nil, err
	}
	result := make([]models.MetricTrendRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.MetricTrendRecord{
			FactDate: row.FactDate.Time, EligibleCount: row.EligibleCount,
			PositiveCount: row.PositiveCount, NegativeCount: row.NegativeCount,
		})
	}
	return result, nil
}

func (d *InsightsDAO) ListMetricBreakdown(ctx context.Context, contract models.MetricContract, dimension models.DimensionContract, factDate time.Time) ([]models.DimensionRecord, error) {
	rows, err := d.queries.ListNavInsightMetricBreakdown(ctx, navsqlc.ListNavInsightMetricBreakdownParams{
		MetricKey: contract.InternalKey, MetricVersion: contract.Version,
		FactDate: pgtype.Date{Time: factDate, Valid: true}, DimensionKey: dimension.InternalKey,
	})
	if err != nil {
		return nil, err
	}
	result := make([]models.DimensionRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.DimensionRecord{
			Value: row.DimensionValue, Label: nonemptyStringPointer(row.Label), LabelEn: nonemptyStringPointer(row.LabelEn),
			Population: row.PopulationCount, Eligible: row.EligibleCount,
			PositiveCount: row.PositiveCount, NegativeCount: row.NegativeCount,
		})
	}
	return result, nil
}

func (d *InsightsDAO) GetMetricSliceAvailability(ctx context.Context, contract models.MetricContract, dimension models.DimensionContract, value string) (models.DimensionAvailabilityRecord, error) {
	row, err := d.queries.GetNavInsightMetricSliceAvailability(ctx, navsqlc.GetNavInsightMetricSliceAvailabilityParams{
		MetricKey: contract.InternalKey, MetricVersion: contract.Version,
		DimensionKey: dimension.InternalKey, DimensionValue: value,
	})
	if err != nil {
		return models.DimensionAvailabilityRecord{}, err
	}
	return models.DimensionAvailabilityRecord{
		Label: nonemptyStringPointer(row.Label), LabelEn: nonemptyStringPointer(row.LabelEn),
		AvailableFrom: datePointer(row.AvailableFrom), AvailableThrough: datePointer(row.AvailableThrough),
	}, nil
}

func (d *InsightsDAO) ListMetricSliceTrend(ctx context.Context, contract models.MetricContract, dimension models.DimensionContract, value string, through time.Time, rangeDays int32) ([]models.DimensionTrendRecord, error) {
	rows, err := d.queries.ListNavInsightMetricSliceTrend(ctx, navsqlc.ListNavInsightMetricSliceTrendParams{
		MetricKey: contract.InternalKey, MetricVersion: contract.Version,
		DimensionKey: dimension.InternalKey, DimensionValue: value,
		ThroughDate: pgtype.Date{Time: through, Valid: true}, RangeDays: rangeDays,
	})
	if err != nil {
		return nil, err
	}
	result := make([]models.DimensionTrendRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.DimensionTrendRecord{
			FactDate: row.FactDate.Time, Population: row.PopulationCount, Eligible: row.EligibleCount,
			PositiveCount: row.PositiveCount, NegativeCount: row.NegativeCount,
		})
	}
	return result, nil
}

func (d *InsightsDAO) GetSiteMetric(ctx context.Context, siteID int64, contract models.MetricContract) (*models.SiteMetricRecord, error) {
	row, err := d.queries.GetNavInsightSiteMetric(ctx, navsqlc.GetNavInsightSiteMetricParams{
		MetricKey: contract.InternalKey, MetricVersion: contract.Version, SiteID: siteID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &models.SiteMetricRecord{
		FactDate: row.FactDate.Time, State: row.State, EligibleCount: row.EligibleCount,
		PositiveCount: row.PositiveCount, NegativeCount: row.NegativeCount,
	}, nil
}

func (d *InsightsDAO) CountOverviewChanges(ctx context.Context, detectorKeys, contractIDs []string) (int64, error) {
	return d.queries.CountNavInsightOverviewChanges(ctx, navsqlc.CountNavInsightOverviewChangesParams{
		DetectorKeys: detectorKeys, ContractIds: contractIDs,
	})
}

func (d *InsightsDAO) ListOverviewChanges(ctx context.Context, detectorKeys, contractIDs []string, limit int32) ([]models.ChangeRecord, error) {
	rows, err := d.queries.ListNavInsightOverviewChanges(ctx, navsqlc.ListNavInsightOverviewChangesParams{
		LimitCount: limit, DetectorKeys: detectorKeys, ContractIds: contractIDs,
	})
	if err != nil {
		return nil, err
	}
	result := make([]models.ChangeRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.ChangeRecord{
			EntityID: row.SiteID, EntityName: row.SiteName, DetectorKey: row.DetectorKey,
			DetectorVersion: row.DetectorVersion, EventCode: row.EventCode,
			ProjectionDate: row.ProjectionDate.Time, TimeBasis: row.TimeBasis, EventAt: timestampPointer(row.EventAt),
		})
	}
	return result, nil
}

func (d *InsightsDAO) ListExplorerChanges(ctx context.Context, conditions models.ChangeExplorerConditions) ([]models.ChangeRecord, error) {
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
	rows, err := d.queries.ListNavInsightExplorerChanges(ctx, navsqlc.ListNavInsightExplorerChangesParams{
		DetectorKeys: conditions.DetectorKeys, ContractIds: conditions.ContractIDs,
		RangeThrough: pgtype.Date{Time: conditions.RangeThrough, Valid: true}, RangeDays: conditions.RangeDays,
		HasPosition: hasPosition, PositionDate: pgtype.Date{Time: positionDate, Valid: true},
		PositionRank: positionRank, PositionSortAt: pgtype.Timestamptz{Time: positionSortAt, Valid: true},
		PositionTie: positionTie, LimitCount: conditions.Limit,
	})
	if err != nil {
		return nil, err
	}
	result := make([]models.ChangeRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.ChangeRecord{
			EntityID: row.SiteID, EntityName: row.SiteName, DetectorKey: row.DetectorKey,
			DetectorVersion: row.DetectorVersion, EventCode: row.EventCode,
			ProjectionDate: row.ProjectionDate.Time, TimeBasis: row.TimeBasis, EventAt: timestampPointer(row.EventAt),
			PrecisionRank: row.PrecisionRank, EventSortAt: row.EventSortAt.Time, OpaqueTie: row.OpaqueTie,
		})
	}
	return result, nil
}

func (d *InsightsDAO) ListSiteChanges(ctx context.Context, site models.SiteRecord, detectorKeys, contractIDs []string, limit int32) ([]models.ChangeRecord, error) {
	rows, err := d.queries.ListNavInsightSiteChanges(ctx, navsqlc.ListNavInsightSiteChangesParams{
		SiteID: site.ID, DetectorKeys: detectorKeys, ContractIds: contractIDs, LimitCount: limit,
	})
	if err != nil {
		return nil, err
	}
	result := make([]models.ChangeRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.ChangeRecord{
			EntityID: site.ID, EntityName: site.Name, DetectorKey: row.DetectorKey,
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

func timestampPointer(value pgtype.Timestamptz) *time.Time {
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
