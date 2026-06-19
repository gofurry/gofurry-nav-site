package routers

import (
	"github.com/gofiber/fiber/v3"
	collect "github.com/gofurry/gofurry-nav-backend/apps/nav/collect/controller"
	detail "github.com/gofurry/gofurry-nav-backend/apps/nav/detail/controller"
	home "github.com/gofurry/gofurry-nav-backend/apps/nav/home/controller"
	navpage "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/controller"
	search "github.com/gofurry/gofurry-nav-backend/apps/nav/search/controller"
	sitedirectory "github.com/gofurry/gofurry-nav-backend/apps/nav/sitedirectory/controller"
	sitegroup "github.com/gofurry/gofurry-nav-backend/apps/nav/sitegroup/controller"
	siteindex "github.com/gofurry/gofurry-nav-backend/apps/nav/siteindex/controller"
	stats "github.com/gofurry/gofurry-nav-backend/apps/nav/stats/controller"
	summary "github.com/gofurry/gofurry-nav-backend/apps/nav/summary/controller"
	updates "github.com/gofurry/gofurry-nav-backend/apps/nav/updates/controller"
	"github.com/gofurry/gofurry-nav-backend/roof/env"
)

func navV2Api(g fiber.Router, cfg env.NavV2Config) {
	g.Get("/home", home.HomeApi.GetHome)
	g.Get("/home/ping", home.HomeApi.GetHomePing)
	g.Get("/home/saying", home.HomeApi.GetHomeSaying)
	g.Get("/home/backgrounds", home.HomeApi.GetHomeBackgrounds)
	g.Get("/updates", updates.UpdatesApi.GetUpdates)
	g.Get("/search/suggestions", search.SearchApi.GetSearchSuggestions)
	g.Get("/sites/index", siteindex.SiteIndexApi.GetSiteIndex)
	g.Get("/sites/directory", sitedirectory.SiteDirectoryApi.GetSiteDirectory)
	g.Get("/site-groups", navpage.NavPageApi.GetGroupList)
	g.Get("/site-groups/:groupId/sites", sitegroup.SiteGroupApi.GetSiteGroupPage)
	g.Post("/stats/page-view", stats.StatsApi.TouchPageView)
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
	collectApi(g.Group("/collect", collect.RequireAdminToken()))
}

func collectApi(g fiber.Router) {
	g.Get("/status", collect.CollectApi.GetStatus)
	g.Get("/observations", collect.CollectApi.ListObservations)
	g.Get("/sites/:siteId/status", collect.CollectApi.GetSiteStatus)
	g.Get("/sites/:siteId/targets/:target/status", collect.CollectApi.GetTargetStatus)
}
