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
	app.Get("/api/v2/nav/sites/:siteId/insights", api.GetSiteInsights)

	cases := []struct {
		path string
		want int
	}{
		{"/api/v2/nav/insights/overview", http.StatusOK},
		{"/api/v2/nav/insights/metrics/ipv6/trend?range=all", http.StatusOK},
		{"/api/v2/nav/sites/1/insights", http.StatusOK},
		{"/api/v2/nav/insights/metrics/bad/trend", http.StatusBadRequest},
		{"/api/v2/nav/insights/metrics/ipv6/trend?range=bad", http.StatusBadRequest},
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
