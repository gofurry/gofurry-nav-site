package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/models"
)

var (
	ErrInvalidMetricKey        = errors.New("invalid public metric key")
	ErrInvalidRange            = errors.New("invalid insights range")
	ErrInvalidDimension        = errors.New("invalid public dimension")
	ErrInvalidSlice            = errors.New("invalid public dimension slice")
	ErrInvalidChanges          = errors.New("invalid change explorer query")
	ErrInvalidCursor           = errors.New("invalid change explorer cursor")
	ErrInvalidCertificateLimit = errors.New("invalid certificate overview limit")
	ErrNotFound                = errors.New("site not found")
)

var metricContracts = []models.MetricContract{
	{PublicKey: "ipv6", InternalKey: "ipv6_adoption", Version: 2},
	{PublicKey: "tls13", InternalKey: "tls13_adoption", Version: 1},
	{PublicKey: "http2", InternalKey: "http2_adoption", Version: 1},
	{PublicKey: "hsts", InternalKey: "hsts_adoption", Version: 1},
	{PublicKey: "csp", InternalKey: "csp_adoption", Version: 1},
	{PublicKey: "security_txt", InternalKey: "security_txt_adoption", Version: 2},
	{PublicKey: "certificate_verified", InternalKey: "tls_certificate_verification", Version: 1},
}

type changeContract struct {
	detector string
	version  int32
	code     string
	public   string
	overview bool
	category string
}

var changeContracts = []changeContract{
	{"ipv6_transition", 2, "ipv6_enabled", "site.ipv6.enabled", true, "capability"},
	{"ipv6_transition", 2, "ipv6_disabled", "site.ipv6.disabled", true, "capability"},
	{"tls13_transition", 1, "tls13_enabled", "site.tls13.enabled", true, "capability"},
	{"tls13_transition", 1, "tls13_disabled", "site.tls13.disabled", true, "capability"},
	{"http2_transition", 1, "http2_enabled", "site.http2.enabled", true, "capability"},
	{"http2_transition", 1, "http2_disabled", "site.http2.disabled", true, "capability"},
	{"hsts_transition", 1, "hsts_added", "site.hsts.added", true, "capability"},
	{"hsts_transition", 1, "hsts_removed", "site.hsts.removed", true, "capability"},
	{"csp_transition", 1, "csp_added", "site.csp.added", true, "capability"},
	{"csp_transition", 1, "csp_removed", "site.csp.removed", true, "capability"},
	{"security_txt_transition", 2, "security_txt_added", "site.security_txt.added", true, "capability"},
	{"security_txt_transition", 2, "security_txt_removed", "site.security_txt.removed", true, "capability"},
	{"primary_target_transition", 1, "primary_target_changed", "site.primary_target.changed", false, "target"},
	{"tls_certificate_transition", 1, "tls_certificate_changed", "site.tls_certificate.changed", false, "certificate"},
	{"tls_certificate_verification_transition", 1, "tls_certificate_verification_failed", "site.tls_certificate.verification_failed", true, "certificate"},
	{"tls_certificate_verification_transition", 1, "tls_certificate_verification_restored", "site.tls_certificate.verification_restored", true, "certificate"},
}

type Store interface {
	CountEntities(context.Context) (int64, error)
	GetSite(context.Context, int64) (*models.SiteRecord, error)
	GetMetricSummary(context.Context, models.MetricContract) (*models.MetricSummaryRecord, error)
	ListMetricTrend(context.Context, models.MetricContract, int32) ([]models.MetricTrendRecord, error)
	ListMetricBreakdown(context.Context, models.MetricContract, models.DimensionContract, time.Time) ([]models.DimensionRecord, error)
	GetMetricSliceAvailability(context.Context, models.MetricContract, models.DimensionContract, string) (models.DimensionAvailabilityRecord, error)
	ListMetricSliceTrend(context.Context, models.MetricContract, models.DimensionContract, string, time.Time, int32) ([]models.DimensionTrendRecord, error)
	GetSiteMetric(context.Context, int64, models.MetricContract) (*models.SiteMetricRecord, error)
	GetCertificateOverviewSummary(context.Context) (*models.CertificateOverviewRecord, error)
	ListCertificateExpiryAttention(context.Context, int32) ([]models.CertificateItemRecord, error)
	ListCertificateVerificationIssues(context.Context, int32) ([]models.CertificateItemRecord, error)
	CountOverviewChanges(context.Context, []string, []string) (int64, error)
	ListOverviewChanges(context.Context, []string, []string, int32) ([]models.ChangeRecord, error)
	ListExplorerChanges(context.Context, models.ChangeExplorerConditions) ([]models.ChangeRecord, error)
	ListSiteChanges(context.Context, models.SiteRecord, []string, []string, int32) ([]models.ChangeRecord, error)
}

type InsightsService struct {
	store Store
	now   func() time.Time
}

func New(store Store) *InsightsService {
	return &InsightsService{store: store, now: func() time.Time { return time.Now().UTC() }}
}

func (s *InsightsService) GetOverview(ctx context.Context) (models.Overview, error) {
	result := models.Overview{GeneratedAt: s.now().UTC(), Metrics: []models.Metric{}, RecentChanges: []models.Change{}}
	var err error
	if result.EntityCount, err = s.store.CountEntities(ctx); err != nil {
		return result, err
	}
	detectors, contracts := changeQueryContracts(true)
	if result.Changes7D, err = s.store.CountOverviewChanges(ctx, detectors, contracts); err != nil {
		return result, err
	}
	for _, contract := range metricContracts {
		record, queryErr := s.store.GetMetricSummary(ctx, contract)
		if queryErr != nil {
			return result, queryErr
		}
		if record != nil {
			result.Metrics = append(result.Metrics, publicMetric(contract.PublicKey, *record))
		}
	}
	changes, err := s.store.ListOverviewChanges(ctx, detectors, contracts, 8)
	if err != nil {
		return result, err
	}
	result.RecentChanges = publicChanges(changes)
	return result, nil
}

func (s *InsightsService) GetMetricTrend(ctx context.Context, publicKey, requestedRange string) (models.MetricTrend, error) {
	result := models.MetricTrend{Key: publicKey, RequestedRange: requestedRange, Points: []models.TrendPoint{}}
	contract, ok := resolveMetric(publicKey)
	if !ok {
		return result, ErrInvalidMetricKey
	}
	rangeDays, ok := parseRange(requestedRange)
	if !ok {
		return result, ErrInvalidRange
	}
	summary, err := s.store.GetMetricSummary(ctx, contract)
	if err != nil {
		return result, err
	}
	rows, err := s.store.ListMetricTrend(ctx, contract, rangeDays)
	if err != nil {
		return result, err
	}
	if summary != nil {
		result.AvailableFrom = dateStringPointer(summary.AvailableFrom)
		through := formatDate(summary.FactDate)
		result.AvailableThrough = &through
	}
	for _, row := range rows {
		result.Points = append(result.Points, models.TrendPoint{
			Date:     formatDate(row.FactDate),
			Value:    ratio(row.PositiveCount, row.PositiveCount+row.NegativeCount),
			Coverage: ratio(row.PositiveCount+row.NegativeCount, row.EligibleCount),
		})
	}
	return result, nil
}

func (s *InsightsService) GetSiteInsights(ctx context.Context, siteID int64) (models.SiteInsights, error) {
	result := models.SiteInsights{Capabilities: []models.Capability{}, RecentChanges: []models.Change{}}
	site, err := s.store.GetSite(ctx, siteID)
	if err != nil {
		return result, err
	}
	if site == nil {
		return result, ErrNotFound
	}
	result.Site = models.EntityRef{ID: site.ID, Name: site.Name}
	for _, contract := range metricContracts {
		record, queryErr := s.store.GetSiteMetric(ctx, siteID, contract)
		if queryErr != nil {
			return result, queryErr
		}
		if record == nil {
			continue
		}
		state, ok := publicState(record.State)
		if !ok {
			return result, fmt.Errorf("map metric state %q", record.State)
		}
		known := record.PositiveCount + record.NegativeCount
		result.Capabilities = append(result.Capabilities, models.Capability{
			Key: contract.PublicKey, AsOf: formatDate(record.FactDate), State: state,
			Ecosystem: models.Ecosystem{
				Value: ratio(record.PositiveCount, known), Coverage: ratio(known, record.EligibleCount),
			},
		})
	}
	detectors, contracts := changeQueryContracts(false)
	changes, err := s.store.ListSiteChanges(ctx, *site, detectors, contracts, 20)
	if err != nil {
		return result, err
	}
	result.RecentChanges = publicChanges(changes)
	return result, nil
}

func publicMetric(key string, record models.MetricSummaryRecord) models.Metric {
	known := record.PositiveCount + record.NegativeCount
	metric := models.Metric{
		Key: key, AsOf: formatDate(record.FactDate), Value: ratio(record.PositiveCount, known),
		Coverage: ratio(known, record.EligibleCount), Known: known, Eligible: record.EligibleCount,
		AvailableFrom: dateStringPointer(record.AvailableFrom),
	}
	if record.PreviousPositiveCount != nil && record.PreviousNegativeCount != nil {
		previousKnown := *record.PreviousPositiveCount + *record.PreviousNegativeCount
		currentValue := ratio(record.PositiveCount, known)
		previousValue := ratio(*record.PreviousPositiveCount, previousKnown)
		if currentValue != nil && previousValue != nil {
			delta := *currentValue - *previousValue
			metric.Delta30D = &delta
		}
	}
	return metric
}

func resolveMetric(publicKey string) (models.MetricContract, bool) {
	for _, contract := range metricContracts {
		if contract.PublicKey == publicKey {
			return contract, true
		}
	}
	return models.MetricContract{}, false
}

func parseRange(value string) (int32, bool) {
	switch value {
	case "30d":
		return 30, true
	case "90d":
		return 90, true
	case "all":
		return 0, true
	default:
		return 0, false
	}
}

func publicState(internal string) (string, bool) {
	states := map[string]string{
		"positive": "supported", "negative": "unsupported", "stale": "stale",
		"not_probed": "not_probed", "probe_failed": "unavailable", "unknown": "unknown",
		"not_applicable": "not_applicable",
	}
	value, ok := states[internal]
	return value, ok
}

func publicChanges(records []models.ChangeRecord) []models.Change {
	result := make([]models.Change, 0, len(records))
	for _, record := range records {
		publicType, ok := publicChangeType(record)
		if !ok {
			continue
		}
		var occurredAt *time.Time
		if record.TimeBasis != "day" {
			occurredAt = record.EventAt
		}
		result = append(result, models.Change{
			Type: publicType, Date: formatDate(record.ProjectionDate), OccurredAt: occurredAt,
			Entity: models.EntityRef{ID: record.EntityID, Name: record.EntityName}, Detail: nil,
		})
	}
	return result
}

func publicChangeType(record models.ChangeRecord) (string, bool) {
	for _, contract := range changeContracts {
		if contract.detector == record.DetectorKey && contract.version == record.DetectorVersion && contract.code == record.EventCode {
			return contract.public, true
		}
	}
	return "", false
}

func changeQueryContracts(overviewOnly bool) ([]string, []string) {
	detectorSet := map[string]struct{}{}
	detectors := []string{}
	contracts := []string{}
	for _, contract := range changeContracts {
		if overviewOnly && !contract.overview {
			continue
		}
		if _, exists := detectorSet[contract.detector]; !exists {
			detectorSet[contract.detector] = struct{}{}
			detectors = append(detectors, contract.detector)
		}
		contracts = append(contracts, fmt.Sprintf("%s/%d/%s", contract.detector, contract.version, contract.code))
	}
	return detectors, contracts
}

func ratio(numerator, denominator int64) *float64 {
	if denominator == 0 {
		return nil
	}
	value := float64(numerator) / float64(denominator)
	return &value
}

func dateStringPointer(value *time.Time) *string {
	if value == nil {
		return nil
	}
	result := formatDate(*value)
	return &result
}

func formatDate(value time.Time) string { return value.UTC().Format("2006-01-02") }
