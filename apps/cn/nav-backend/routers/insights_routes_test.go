package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

type navInsightsRouteStub struct{}

func (navInsightsRouteStub) GetOverview(c fiber.Ctx) error { return c.SendStatus(http.StatusNoContent) }
func (navInsightsRouteStub) GetMetricTrend(c fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}
func (navInsightsRouteStub) GetMetricBreakdown(c fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}
func (navInsightsRouteStub) GetMetricSliceTrend(c fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}
func (navInsightsRouteStub) GetChanges(c fiber.Ctx) error { return c.SendStatus(http.StatusNoContent) }
func (navInsightsRouteStub) GetCertificateOverview(c fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}
func (navInsightsRouteStub) GetSiteCompare(c fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}
func (navInsightsRouteStub) GetSiteInsights(c fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}

func TestNavInsightsRoutesAreRegistered(t *testing.T) {
	app := fiber.New()
	registerRoutes(app, NavDependencies{Insights: navInsightsRouteStub{}})
	for _, path := range []string{
		"/api/v2/nav/insights/overview",
		"/api/v2/nav/insights/metrics/ipv6/trend",
		"/api/v2/nav/insights/metrics/ipv6/breakdown?dimension=country",
		"/api/v2/nav/insights/metrics/ipv6/breakdown/country/CN/trend",
		"/api/v2/nav/insights/changes",
		"/api/v2/nav/insights/certificates/overview",
		"/api/v2/nav/insights/compare?ids=1,2",
		"/api/v2/nav/sites/1/insights",
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
