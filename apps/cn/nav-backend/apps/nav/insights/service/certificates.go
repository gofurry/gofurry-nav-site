package service

import (
	"context"
	"math"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/models"
)

const (
	defaultCertificateLimit int32 = 20
	maximumCertificateLimit int32 = 100
)

func (s *InsightsService) GetCertificateOverview(ctx context.Context, limit int32) (models.CertificateOverview, error) {
	result := models.CertificateOverview{
		FreshnessSeconds:   172800,
		ExpiryAttention:    []models.CertificateItem{},
		VerificationIssues: []models.CertificateItem{},
	}
	if limit == 0 {
		limit = defaultCertificateLimit
	}
	if limit < 1 || limit > maximumCertificateLimit {
		return result, ErrInvalidCertificateLimit
	}

	summary, err := s.store.GetCertificateOverviewSummary(ctx)
	if err != nil {
		return result, err
	}
	if summary == nil {
		return result, nil
	}
	asOf := formatDate(summary.FactDate)
	result.AsOf = &asOf
	referenceAt := summary.ReferenceAt.UTC()
	result.ReferenceAt = &referenceAt
	result.FreshnessSeconds = summary.FreshnessSeconds
	result.Population = summary.Population
	result.Eligible = summary.Eligible
	result.Verification = models.CertificateVerificationSummary{
		Known: summary.Verified + summary.Failed, Verified: summary.Verified, Failed: summary.Failed,
		Coverage: ratio(summary.Verified+summary.Failed, summary.Eligible),
	}
	result.Quality = models.CertificateQualitySummary{
		NotApplicable: summary.NotApplicable, Stale: summary.Stale, NotProbed: summary.NotProbed,
		ProbeFailed: summary.ProbeFailed, Unknown: summary.Unknown,
	}
	result.Expiry = models.CertificateExpirySummary{
		Known:    summary.Expired + summary.ExpiresWithin7D + summary.ExpiresIn8To30D + summary.Later,
		Coverage: ratio(summary.Expired+summary.ExpiresWithin7D+summary.ExpiresIn8To30D+summary.Later, summary.Eligible),
		Expired:  summary.Expired, ExpiresWithin7D: summary.ExpiresWithin7D,
		ExpiresIn8To30D: summary.ExpiresIn8To30D, Later: summary.Later,
	}

	attention, err := s.store.ListCertificateExpiryAttention(ctx, limit)
	if err != nil {
		return result, err
	}
	issues, err := s.store.ListCertificateVerificationIssues(ctx, limit)
	if err != nil {
		return result, err
	}
	result.ExpiryAttention = publicCertificateItems(attention, referenceAt)
	result.VerificationIssues = publicCertificateItems(issues, referenceAt)
	return result, nil
}

func publicCertificateItems(records []models.CertificateItemRecord, referenceAt time.Time) []models.CertificateItem {
	items := make([]models.CertificateItem, 0, len(records))
	for _, record := range records {
		daysToExpiry, expiryStatus := certificateExpiry(record.NotAfter, referenceAt)
		items = append(items, models.CertificateItem{
			Site:   models.EntityRef{ID: record.SiteID, Name: record.SiteName},
			Target: record.Target, NotAfter: record.NotAfter, DaysToExpiry: daysToExpiry,
			ExpiryStatus: expiryStatus, Verified: record.Verified,
			VerificationIssue: normalizeVerificationIssue(record.VerificationIssue),
			Issuer:            record.Issuer, ObservedAt: record.ObservedAt,
		})
	}
	return items
}

func certificateExpiry(notAfter *time.Time, referenceAt time.Time) (*int32, *string) {
	if notAfter == nil {
		return nil, nil
	}
	days := int32(math.Floor(notAfter.Sub(referenceAt).Hours() / 24))
	status := "later"
	switch {
	case !notAfter.After(referenceAt):
		status = "expired"
	case !notAfter.After(referenceAt.Add(7 * 24 * time.Hour)):
		status = "expires_within_7d"
	case !notAfter.After(referenceAt.Add(30 * 24 * time.Hour)):
		status = "expires_in_8_30d"
	}
	return &days, &status
}

func normalizeVerificationIssue(value *string) *string {
	if value == nil {
		return nil
	}
	switch *value {
	case "hostname_mismatch", "unknown_authority", "expired", "not_yet_valid", "incompatible_usage", "other":
		return value
	default:
		other := "other"
		return &other
	}
}
