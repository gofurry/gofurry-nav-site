package routers

import (
	"github.com/gofiber/fiber/v3"
	detail "github.com/gofurry/gofurry-nav-backend/apps/nav/detail/controller"
	home "github.com/gofurry/gofurry-nav-backend/apps/nav/home/controller"
	summary "github.com/gofurry/gofurry-nav-backend/apps/nav/summary/controller"
	updates "github.com/gofurry/gofurry-nav-backend/apps/nav/updates/controller"
	"github.com/gofurry/gofurry-nav-backend/roof/env"
)

func navV2Api(g fiber.Router, cfg env.NavV2Config) {
	g.Get("/home", home.HomeApi.GetHome)
	g.Get("/home/ping", home.HomeApi.GetHomePing)
	g.Get("/updates", updates.UpdatesApi.GetUpdates)
	if cfg.DetailRoutesEnabled() {
		g.Get("/sites/:siteId/detail", detail.DetailApi.GetSiteDetail)
		g.Post("/sites/:siteId/view", detail.DetailApi.TouchSiteView)
	}
	if cfg.SummaryRoutesEnabled() {
		g.Get("/sites/:siteId/summary", summary.SummaryApi.GetSiteSummary)
		g.Get("/sites/:siteId/targets/:target/summary", summary.SummaryApi.GetTargetSummary)
	}
	if cfg.ReadModelRoutesEnabled() {
		g.Get("/sites/:siteId/targets/:target/latest", detail.DetailApi.GetTargetLatest)
		g.Get("/sites/:siteId/targets/:target/observations", detail.DetailApi.ListTargetObservations)
		g.Get("/sites/:siteId/targets/:target/trend", detail.DetailApi.GetTargetTrend)
		g.Get("/sites/:siteId/targets/:target/changes", detail.DetailApi.GetTargetChanges)
		g.Get("/sites/:siteId/targets/:target/light-probes", detail.DetailApi.GetTargetLightProbes)
	}
}
