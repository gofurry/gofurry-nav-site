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
