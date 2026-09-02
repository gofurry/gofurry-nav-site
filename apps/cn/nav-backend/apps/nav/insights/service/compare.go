package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/models"
)

func (s *InsightsService) GetSiteCompare(ctx context.Context, rawIDs string) (models.SiteCompare, error) {
	result := models.SiteCompare{Status: "insufficient_data", Sites: []models.SiteCompareItem{}}
	siteIDs, err := parseSiteCompareIDs(rawIDs)
	if err != nil {
		return result, err
	}

	for _, siteID := range siteIDs {
		site, queryErr := s.store.GetSite(ctx, siteID)
		if queryErr != nil {
			return result, queryErr
		}
		if site == nil {
			return result, ErrNotFound
		}
		result.Sites = append(result.Sites, models.SiteCompareItem{
			Site: models.EntityRef{ID: site.ID, Name: site.Name}, Capabilities: []models.SiteCompareCapability{},
		})
	}

	horizon, err := s.store.GetSiteCompareHorizon(ctx, siteIDs, metricContracts)
	if err != nil {
		return result, err
	}
	if horizon == nil {
		return result, nil
	}
	asOf := formatDate(*horizon)
	result.AsOf = &asOf

	capabilities, err := s.store.ListSiteCompareCapabilities(ctx, siteIDs, metricContracts, *horizon)
	if err != nil {
		return result, err
	}
	bySite := make(map[int64][]models.SiteCompareCapability, len(siteIDs))
	for _, record := range capabilities {
		key, ok := publicMetricContract(record.MetricKey, record.MetricVersion)
		if !ok {
			return result, fmt.Errorf("map compare metric %s/%d", record.MetricKey, record.MetricVersion)
		}
		state, ok := publicState(record.State)
		if !ok {
			return result, fmt.Errorf("map compare metric state %q", record.State)
		}
		bySite[record.SiteID] = append(bySite[record.SiteID], models.SiteCompareCapability{Key: key, State: state})
	}
	for index := range result.Sites {
		item := &result.Sites[index]
		item.Capabilities = bySite[item.Site.ID]
		if len(item.Capabilities) != len(metricContracts) {
			return result, fmt.Errorf("incomplete site comparison snapshot")
		}
	}

	certificates, err := s.store.ListSiteCompareCertificates(ctx, siteIDs, *horizon)
	if err != nil {
		return result, err
	}
	publicCertificates := publicCertificateItems(certificates, horizon.AddDate(0, 0, 1))
	certificateBySite := make(map[int64]models.CertificateItem, len(publicCertificates))
	for _, certificate := range publicCertificates {
		certificateBySite[certificate.Site.ID] = certificate
	}
	for index := range result.Sites {
		item := &result.Sites[index]
		certificate, ok := certificateBySite[item.Site.ID]
		if !ok {
			continue
		}
		item.Certificate = &models.SiteCompareCertificate{
			Target: certificate.Target, NotAfter: certificate.NotAfter, DaysToExpiry: certificate.DaysToExpiry,
			ExpiryStatus: certificate.ExpiryStatus, Verified: certificate.Verified,
			VerificationIssue: certificate.VerificationIssue, Issuer: certificate.Issuer, ObservedAt: certificate.ObservedAt,
		}
	}
	result.Status = "ready"
	return result, nil
}

func parseSiteCompareIDs(raw string) ([]int64, error) {
	parts := strings.Split(raw, ",")
	result := make([]int64, 0, len(parts))
	seen := map[int64]struct{}{}
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			return nil, ErrInvalidCompare
		}
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil || id <= 0 {
			return nil, ErrInvalidCompare
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	if len(result) < 2 || len(result) > 4 {
		return nil, ErrInvalidCompare
	}
	return result, nil
}

func publicMetricContract(internalKey string, version int32) (string, bool) {
	for _, contract := range metricContracts {
		if contract.InternalKey == internalKey && contract.Version == version {
			return contract.PublicKey, true
		}
	}
	return "", false
}
