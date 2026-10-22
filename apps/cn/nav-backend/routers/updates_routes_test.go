package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/roof/env"
	"io"
	"net/http/httptest"
	"testing"
)

type updatesRouteStub struct{}

func (updatesRouteStub) GetUpdates(c fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) }
func (updatesRouteStub) GetUpdateDetail(c fiber.Ctx) error {
	return c.SendString(c.Params("id") + ":" + c.Query("lang"))
}

func TestReleaseNotesPublicRoutes(t *testing.T) {
	app := fiber.New()
	cfg := env.GetServerConfig()
	previous := cfg.NavV2
	t.Cleanup(func() { cfg.NavV2 = previous })
	disabled := false
	cfg.NavV2 = env.NavV2Config{SummaryEnabled: true, DetailEnabled: &disabled, ReadModelEnabled: &disabled}
	stub := updatesRouteStub{}
	registerRoutes(app, NavDependencies{Updates: stub, Home: stub, Search: stub, SiteIndex: stub, NavPage: stub, Insights: navInsightsRouteStub{}})
	for _, tc := range []struct {
		path   string
		status int
	}{{"/api/v2/nav/appearance/patterns", 404}, {"/api/v2/nav/updates", 204}, {"/api/v2/nav/updates/12?lang=en", 200}} {
		response, err := app.Test(httptest.NewRequest("GET", tc.path, nil))
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(response.Body)
		response.Body.Close()
		if tc.status == 200 && string(body) != "12:en" {
			t.Fatalf("detail route parameters: %s", body)
		}
		if response.StatusCode != tc.status {
			t.Fatalf("%s: %d", tc.path, response.StatusCode)
		}
	}
}

func (updatesRouteStub) GetHome(c fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) }

func (updatesRouteStub) GetHomePing(c fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) }

func (updatesRouteStub) GetHomeSaying(c fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) }

func (updatesRouteStub) GetHomeHero(c fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) }

func (updatesRouteStub) GetHeroes(c fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) }

func (updatesRouteStub) GetSearchSuggestions(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func (updatesRouteStub) GetSiteIndex(c fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) }

func (updatesRouteStub) GetGroupList(c fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) }
