package dao

import (
	"context"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/models"
	navsqlc "github.com/gofurry/gofurry-nav-backend/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

func (d *InsightsDAO) GetSiteCompareHorizon(ctx context.Context, siteIDs []int64, contracts []models.MetricContract) (*time.Time, error) {
	keys, versions := siteCompareContracts(contracts)
	value, err := d.queries.GetNavInsightCompareHorizon(ctx, navsqlc.GetNavInsightCompareHorizonParams{
		SiteIds: siteIDs, MetricKeys: keys, MetricVersions: versions,
	})
	if err != nil {
		return nil, err
	}
	return datePointer(value), nil
}

func (d *InsightsDAO) ListSiteCompareCapabilities(ctx context.Context, siteIDs []int64, contracts []models.MetricContract, factDate time.Time) ([]models.SiteCompareCapabilityRecord, error) {
	keys, versions := siteCompareContracts(contracts)
	rows, err := d.queries.ListNavInsightCompareCapabilities(ctx, navsqlc.ListNavInsightCompareCapabilitiesParams{
		FactDate: pgtype.Date{Time: factDate, Valid: true}, SiteIds: siteIDs,
		MetricKeys: keys, MetricVersions: versions,
	})
	if err != nil {
		return nil, err
	}
	result := make([]models.SiteCompareCapabilityRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.SiteCompareCapabilityRecord{
			SiteID: row.SiteID, MetricKey: row.MetricKey, MetricVersion: row.MetricVersion, State: row.State,
		})
	}
	return result, nil
}

func (d *InsightsDAO) ListSiteCompareCertificates(ctx context.Context, siteIDs []int64, factDate time.Time) ([]models.CertificateItemRecord, error) {
	rows, err := d.queries.ListNavInsightCompareCertificates(ctx, navsqlc.ListNavInsightCompareCertificatesParams{
		SiteIds: siteIDs, FactDate: pgtype.Date{Time: factDate, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	result := make([]models.CertificateItemRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.CertificateItemRecord{
			SiteID: row.SiteID, SiteName: row.SiteName, Target: row.Target,
			NotAfter: timestampPointer(row.TlsCertNotAfter), Verified: row.Verified,
			VerificationIssue: nonemptyStringPointer(row.VerificationIssue), Issuer: row.Issuer,
			ObservedAt: timestampPointer(row.ObservedAt),
		})
	}
	return result, nil
}

func siteCompareContracts(contracts []models.MetricContract) ([]string, []int32) {
	keys := make([]string, 0, len(contracts))
	versions := make([]int32, 0, len(contracts))
	for _, contract := range contracts {
		keys = append(keys, contract.InternalKey)
		versions = append(versions, contract.Version)
	}
	return keys, versions
}
