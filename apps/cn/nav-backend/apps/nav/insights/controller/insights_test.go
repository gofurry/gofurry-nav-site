package controller

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/models"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/service"
)

type fakeReader struct{}

func (fakeReader) GetOverview(context.Context) (models.Overview, error) {
	return models.Overview{Metrics: []models.Metric{}, RecentChanges: []models.Change{}}, nil
}
func (fakeReader) GetMetricTrend(_ context.Context, key, requestedRange string) (models.MetricTrend, error) {
	if key == "bad" {
		return models.MetricTrend{}, service.ErrInvalidMetricKey
	}
	if requestedRange == "bad" {
		return models.MetricTrend{}, service.ErrInvalidRange
	}
	return models.MetricTrend{Key: key, RequestedRange: requestedRange, Points: []models.TrendPoint{}}, nil
}
func (fakeReader) GetMetricBreakdown(_ context.Context, key, dimension string) (models.DimensionBreakdown, error) {
	if dimension == "bad" {
		return models.DimensionBreakdown{}, service.ErrInvalidDimension
	}
	return models.DimensionBreakdown{Key: key, Dimension: dimension, SliceMode: "partition", Items: []models.DimensionSlice{}}, nil
}
func (fakeReader) GetMetricSliceTrend(_ context.Context, key, dimension, value, requestedRange string) (models.DimensionTrend, error) {
	if value == "bad" {
		return models.DimensionTrend{}, service.ErrInvalidSlice
	}
	return models.DimensionTrend{Key: key, Dimension: dimension, RequestedRange: requestedRange, Points: []models.DimensionTrendPoint{}}, nil
}
func (fakeReader) GetChanges(_ context.Context, query models.ChangeExplorerQuery) (models.ChangeExplorerPage, error) {
	if query.Cursor == "bad" {
		return models.ChangeExplorerPage{}, service.ErrInvalidCursor
	}
	return models.ChangeExplorerPage{Items: []models.ExplorerChange{}}, nil
}
func (fakeReader) GetCertificateOverview(_ context.Context, limit int32) (models.CertificateOverview, error) {
	if limit > 100 {
		return models.CertificateOverview{}, service.ErrInvalidCertificateLimit
	}
	return models.CertificateOverview{ExpiryAttention: []models.CertificateItem{}, VerificationIssues: []models.CertificateItem{}}, nil
}
func (fakeReader) GetSiteCompare(_ context.Context, ids string) (models.SiteCompare, error) {
	if ids == "bad" {
		return models.SiteCompare{}, service.ErrInvalidCompare
	}
	if ids == "1,404" {
		return models.SiteCompare{}, service.ErrNotFound
	}
	return models.SiteCompare{Status: "ready", Sites: []models.SiteCompareItem{}}, nil
}
func (fakeReader) GetSiteInsights(_ context.Context, id int64) (models.SiteInsights, error) {
	if id == 404 {
		return models.SiteInsights{}, service.ErrNotFound
	}
	return models.SiteInsights{Site: models.EntityRef{ID: id}, Capabilities: []models.Capability{}, RecentChanges: []models.Change{}}, nil
}

func TestNavInsightsHTTPContract(t *testing.T) {
	api := New(fakeReader{})
	app := fiber.New()
	app.Get("/api/v2/nav/insights/overview", api.GetOverview)
	app.Get("/api/v2/nav/insights/metrics/:metricKey/trend", api.GetMetricTrend)
	app.Get("/api/v2/nav/insights/metrics/:metricKey/breakdown", api.GetMetricBreakdown)
	app.Get("/api/v2/nav/insights/metrics/:metricKey/breakdown/:dimension/:value/trend", api.GetMetricSliceTrend)
	app.Get("/api/v2/nav/insights/changes", api.GetChanges)
	app.Get("/api/v2/nav/insights/certificates/overview", api.GetCertificateOverview)
	app.Get("/api/v2/nav/insights/compare", api.GetSiteCompare)
	app.Get("/api/v2/nav/sites/:siteId/insights", api.GetSiteInsights)

	cases := []struct {
		path string
		want int
	}{
		{"/api/v2/nav/insights/overview", http.StatusOK},
		{"/api/v2/nav/insights/metrics/ipv6/trend?range=all", http.StatusOK},
		{"/api/v2/nav/insights/metrics/ipv6/breakdown?dimension=country", http.StatusOK},
		{"/api/v2/nav/insights/metrics/ipv6/breakdown/country/CN/trend?range=90d", http.StatusOK},
		{"/api/v2/nav/insights/changes?range=30d&limit=20", http.StatusOK},
		{"/api/v2/nav/insights/changes?cursor=bad", http.StatusBadRequest},
		{"/api/v2/nav/insights/certificates/overview", http.StatusOK},
		{"/api/v2/nav/insights/certificates/overview?limit=101", http.StatusBadRequest},
		{"/api/v2/nav/insights/certificates/overview?limit=bad", http.StatusBadRequest},
		{"/api/v2/nav/insights/compare?ids=1,2", http.StatusOK},
		{"/api/v2/nav/insights/compare?ids=bad", http.StatusBadRequest},
		{"/api/v2/nav/insights/compare?ids=1,404", http.StatusNotFound},
		{"/api/v2/nav/sites/1/insights", http.StatusOK},
		{"/api/v2/nav/insights/metrics/bad/trend", http.StatusBadRequest},
		{"/api/v2/nav/insights/metrics/ipv6/trend?range=bad", http.StatusBadRequest},
		{"/api/v2/nav/insights/metrics/ipv6/breakdown?dimension=bad", http.StatusBadRequest},
		{"/api/v2/nav/sites/404/insights", http.StatusNotFound},
	}
	for _, tc := range cases {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, tc.path, http.NoBody))
		if err != nil {
			t.Fatalf("%s: %v", tc.path, err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != tc.want {
			t.Fatalf("%s status=%d body=%s", tc.path, resp.StatusCode, body)
		}
		for _, forbidden := range []string{"detector_key", "metric_version", "internal_key"} {
			if strings.Contains(string(body), forbidden) {
				t.Fatalf("%s leaked %q: %s", tc.path, forbidden, body)
			}
		}
	}
}
