package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

type gameInsightsRouteStub struct{}

func (gameInsightsRouteStub) GetInsightsOverview(c fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}
func (gameInsightsRouteStub) GetInsightsMetricTrend(c fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}
func (gameInsightsRouteStub) GetGameInsights(c fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}
func (gameInsightsRouteStub) GetGamePlayerInsights(c fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}
func (gameInsightsRouteStub) GetGamePriceInsights(c fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}

func TestGameInsightsRoutesAreRegistered(t *testing.T) {
	app := fiber.New()
	registerGameInsightRoutes(app.Group("/api/v2/game"), gameInsightsRouteStub{})
	for _, path := range []string{
		"/api/v2/game/insights/overview",
		"/api/v2/game/insights/metrics/free/trend",
		"/api/v2/game/games/1/insights",
		"/api/v2/game/games/1/insights/players",
		"/api/v2/game/games/1/insights/prices",
	} {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, path, http.NoBody))
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("%s status=%d", path, resp.StatusCode)
		}
	}
}
