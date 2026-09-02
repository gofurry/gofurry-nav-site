package controller

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	v2service "github.com/gofurry/gofurry-game-backend/apps/game/v2/service"
)

type fakeInsightsReader struct{}

func (fakeInsightsReader) GetInsightsOverview(context.Context) (v2models.InsightOverview, error) {
	return v2models.InsightOverview{Metrics: []v2models.InsightMetric{}, RecentChanges: []v2models.InsightChange{}}, nil
}
func (fakeInsightsReader) GetInsightsMetricTrend(_ context.Context, key, requestedRange string) (v2models.InsightMetricTrend, error) {
	if key == "bad" {
		return v2models.InsightMetricTrend{}, v2service.ErrInvalidInsightMetric
	}
	if requestedRange == "bad" {
		return v2models.InsightMetricTrend{}, v2service.ErrInvalidInsightRange
	}
	return v2models.InsightMetricTrend{Key: key, RequestedRange: requestedRange, Points: []v2models.InsightTrendPoint{}}, nil
}
func (fakeInsightsReader) GetInsightsMetricBreakdown(_ context.Context, key, dimension string) (v2models.InsightDimensionBreakdown, error) {
	if dimension == "bad" {
		return v2models.InsightDimensionBreakdown{}, v2service.ErrInvalidInsightDimension
	}
	return v2models.InsightDimensionBreakdown{Key: key, Dimension: dimension, SliceMode: "partition", Items: []v2models.InsightDimensionSlice{}}, nil
}
func (fakeInsightsReader) GetInsightsMetricSliceTrend(_ context.Context, key, dimension, value, requestedRange string) (v2models.InsightDimensionTrend, error) {
	if value == "bad" {
		return v2models.InsightDimensionTrend{}, v2service.ErrInvalidInsightSlice
	}
	return v2models.InsightDimensionTrend{Key: key, Dimension: dimension, RequestedRange: requestedRange, Points: []v2models.InsightDimensionTrendPoint{}}, nil
}
func (fakeInsightsReader) GetInsightsChanges(_ context.Context, query v2models.InsightChangeExplorerQuery) (v2models.InsightChangeExplorerPage, error) {
	if query.Cursor == "bad" {
		return v2models.InsightChangeExplorerPage{}, v2service.ErrInvalidInsightCursor
	}
	return v2models.InsightChangeExplorerPage{Items: []v2models.InsightExplorerChange{}}, nil
}
func (fakeInsightsReader) GetPlayerRanking(_ context.Context, query v2models.InsightPlayerRankingQuery) (v2models.InsightPlayerRanking, error) {
	if query.Metric == "bad" {
		return v2models.InsightPlayerRanking{}, v2service.ErrInvalidPlayerRanking
	}
	if query.Limit > 100 {
		return v2models.InsightPlayerRanking{}, v2service.ErrInvalidInsightLimit
	}
	return v2models.InsightPlayerRanking{Items: []v2models.InsightPlayerRankingItem{}}, nil
}
func (fakeInsightsReader) GetPriceOverview(_ context.Context, region string) (v2models.InsightPriceOverview, error) {
	if region == "JP" {
		return v2models.InsightPriceOverview{}, v2service.ErrInvalidInsightRegion
	}
	return v2models.InsightPriceOverview{}, nil
}
func (fakeInsightsReader) GetDiscounts(_ context.Context, region string, limit int32) (v2models.InsightDiscounts, error) {
	if region == "JP" {
		return v2models.InsightDiscounts{}, v2service.ErrInvalidInsightRegion
	}
	if limit > 100 {
		return v2models.InsightDiscounts{}, v2service.ErrInvalidInsightLimit
	}
	return v2models.InsightDiscounts{Items: []v2models.InsightDiscountItem{}}, nil
}
func (fakeInsightsReader) GetLanguageOverview(context.Context) (v2models.InsightLanguageOverview, error) {
	return v2models.InsightLanguageOverview{Items: []v2models.InsightLanguageItem{}}, nil
}
func (fakeInsightsReader) GetGameCompare(_ context.Context, ids, region string) (v2models.GameCompare, error) {
	if ids == "bad" {
		return v2models.GameCompare{}, v2service.ErrInvalidInsightCompare
	}
	if ids == "1,404" {
		return v2models.GameCompare{}, v2service.ErrInsightGameNotFound
	}
	if region == "JP" {
		return v2models.GameCompare{}, v2service.ErrInvalidInsightRegion
	}
	return v2models.GameCompare{Status: "ready", Region: region, Games: []v2models.GameCompareItem{}}, nil
}
func (fakeInsightsReader) GetGameInsights(_ context.Context, id int64) (v2models.GameInsights, error) {
	if id == 404 {
		return v2models.GameInsights{}, v2service.ErrInsightGameNotFound
	}
	return v2models.GameInsights{Game: v2models.InsightEntityRef{ID: id}, RecentChanges: []v2models.InsightChange{}}, nil
}
func (fakeInsightsReader) GetGamePlayerInsights(_ context.Context, id int64, requestedRange string) (v2models.InsightPlayerHistory, error) {
	if id == 404 {
		return v2models.InsightPlayerHistory{}, v2service.ErrInsightGameNotFound
	}
	if requestedRange == "bad" {
		return v2models.InsightPlayerHistory{}, v2service.ErrInvalidInsightRange
	}
	return v2models.InsightPlayerHistory{RequestedRange: requestedRange, Points: []v2models.InsightPlayerPoint{}}, nil
}
func (fakeInsightsReader) GetGamePriceInsights(_ context.Context, id int64, region, requestedRange string) (v2models.InsightPriceHistory, error) {
	if id == 404 {
		return v2models.InsightPriceHistory{}, v2service.ErrInsightGameNotFound
	}
	if requestedRange == "bad" {
		return v2models.InsightPriceHistory{}, v2service.ErrInvalidInsightRange
	}
	if region == "JP" {
		return v2models.InsightPriceHistory{}, v2service.ErrInvalidInsightRegion
	}
	return v2models.InsightPriceHistory{RequestedRange: requestedRange, Points: []v2models.InsightPricePoint{}}, nil
}

func TestGameInsightsHTTPContract(t *testing.T) {
	api := New(nil, nil, nil, fakeInsightsReader{})
	app := fiber.New()
	app.Get("/api/v2/game/insights/overview", api.GetInsightsOverview)
	app.Get("/api/v2/game/insights/metrics/:metricKey/trend", api.GetInsightsMetricTrend)
	app.Get("/api/v2/game/insights/metrics/:metricKey/breakdown", api.GetInsightsMetricBreakdown)
	app.Get("/api/v2/game/insights/metrics/:metricKey/breakdown/:dimension/:value/trend", api.GetInsightsMetricSliceTrend)
	app.Get("/api/v2/game/insights/changes", api.GetInsightsChanges)
	app.Get("/api/v2/game/insights/players/ranking", api.GetPlayerRanking)
	app.Get("/api/v2/game/insights/prices/overview", api.GetPriceOverview)
	app.Get("/api/v2/game/insights/prices/discounts", api.GetDiscounts)
	app.Get("/api/v2/game/insights/languages/overview", api.GetLanguageOverview)
	app.Get("/api/v2/game/insights/compare", api.GetGameCompare)
	app.Get("/api/v2/game/games/:gameId/insights", api.GetGameInsights)
	app.Get("/api/v2/game/games/:gameId/insights/players", api.GetGamePlayerInsights)
	app.Get("/api/v2/game/games/:gameId/insights/prices", api.GetGamePriceInsights)

	cases := []struct {
		path string
		want int
	}{
		{"/api/v2/game/insights/overview", http.StatusOK},
		{"/api/v2/game/insights/metrics/free/trend?range=90d", http.StatusOK},
		{"/api/v2/game/insights/metrics/free/breakdown?dimension=primary_tag", http.StatusOK},
		{"/api/v2/game/insights/metrics/free/breakdown/tag/1/trend?range=90d", http.StatusOK},
		{"/api/v2/game/insights/changes?range=30d&limit=20", http.StatusOK},
		{"/api/v2/game/insights/changes?cursor=bad", http.StatusBadRequest},
		{"/api/v2/game/insights/players/ranking", http.StatusOK},
		{"/api/v2/game/insights/players/ranking?metric=bad", http.StatusBadRequest},
		{"/api/v2/game/insights/players/ranking?limit=101", http.StatusBadRequest},
		{"/api/v2/game/insights/prices/overview?region=US", http.StatusOK},
		{"/api/v2/game/insights/prices/overview?region=JP", http.StatusBadRequest},
		{"/api/v2/game/insights/prices/discounts?region=HK", http.StatusOK},
		{"/api/v2/game/insights/prices/discounts?limit=101", http.StatusBadRequest},
		{"/api/v2/game/insights/languages/overview", http.StatusOK},
		{"/api/v2/game/insights/compare?ids=1,2&region=US", http.StatusOK},
		{"/api/v2/game/insights/compare?ids=bad", http.StatusBadRequest},
		{"/api/v2/game/insights/compare?ids=1,2&region=JP", http.StatusBadRequest},
		{"/api/v2/game/insights/compare?ids=1,404", http.StatusNotFound},
		{"/api/v2/game/games/1/insights", http.StatusOK},
		{"/api/v2/game/games/1/insights/players?range=all", http.StatusOK},
		{"/api/v2/game/games/1/insights/prices?range=30d", http.StatusOK},
		{"/api/v2/game/games/1/insights/prices?region=JP", http.StatusBadRequest},
		{"/api/v2/game/insights/metrics/bad/trend", http.StatusBadRequest},
		{"/api/v2/game/insights/metrics/free/trend?range=bad", http.StatusBadRequest},
		{"/api/v2/game/games/404/insights", http.StatusNotFound},
		{"/api/v2/game/games/1/insights/players?range=bad", http.StatusBadRequest},
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
		for _, forbidden := range []string{"detector_key", "metric_version", "internal_key", "old_value", "new_value"} {
			if strings.Contains(string(body), forbidden) {
				t.Fatalf("%s leaked %q: %s", tc.path, forbidden, body)
			}
		}
	}
}
