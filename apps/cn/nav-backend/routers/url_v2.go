package routers

import (
	"github.com/gofiber/fiber/v3"
	sitedirectory "github.com/gofurry/gofurry-nav-backend/apps/nav/sitedirectory/controller"
	sitegroup "github.com/gofurry/gofurry-nav-backend/apps/nav/sitegroup/controller"
	stats "github.com/gofurry/gofurry-nav-backend/apps/nav/stats/controller"
	summary "github.com/gofurry/gofurry-nav-backend/apps/nav/summary/controller"
	"github.com/gofurry/gofurry-nav-backend/roof/env"
)

type HomeAPI interface {
	GetHome(fiber.Ctx) error
	GetHomePing(fiber.Ctx) error
	GetHomeSaying(fiber.Ctx) error
	GetHomeBackgrounds(fiber.Ctx) error
}

type UpdatesAPI interface{ GetUpdates(fiber.Ctx) error }
type SearchAPI interface{ GetSearchSuggestions(fiber.Ctx) error }
type SiteIndexAPI interface{ GetSiteIndex(fiber.Ctx) error }
type NavPageAPI interface{ GetGroupList(fiber.Ctx) error }
type DetailAPI interface {
	GetSiteDetail(fiber.Ctx) error
	TouchSiteView(fiber.Ctx) error
	GetTargetLatest(fiber.Ctx) error
	ListTargetObservations(fiber.Ctx) error
	GetTargetTrend(fiber.Ctx) error
	GetTargetChanges(fiber.Ctx) error
	GetTargetLightProbes(fiber.Ctx) error
}
type InsightsAPI interface {
	GetOverview(fiber.Ctx) error
	GetMetricTrend(fiber.Ctx) error
	GetSiteInsights(fiber.Ctx) error
}
type NavDependencies struct {
	Home      HomeAPI
	Updates   UpdatesAPI
	Search    SearchAPI
	SiteIndex SiteIndexAPI
	NavPage   NavPageAPI
	Detail    DetailAPI
	Insights  InsightsAPI
}

func navV2Api(g fiber.Router, cfg env.NavV2Config, dependencies NavDependencies) {
	g.Get("/home", dependencies.Home.GetHome)
	g.Get("/home/ping", dependencies.Home.GetHomePing)
	g.Get("/home/saying", dependencies.Home.GetHomeSaying)
	g.Get("/home/backgrounds", dependencies.Home.GetHomeBackgrounds)
	g.Get("/updates", dependencies.Updates.GetUpdates)
	g.Get("/search/suggestions", dependencies.Search.GetSearchSuggestions)
	g.Get("/sites/index", dependencies.SiteIndex.GetSiteIndex)
	g.Get("/sites/directory", sitedirectory.SiteDirectoryApi.GetSiteDirectory)
	g.Get("/site-groups", dependencies.NavPage.GetGroupList)
	g.Get("/site-groups/:groupId/sites", sitegroup.SiteGroupApi.GetSiteGroupPage)
	g.Post("/stats/page-view", stats.StatsApi.TouchPageView)
	if cfg.DetailRoutesEnabled() {
		g.Get("/sites/:siteId/detail", dependencies.Detail.GetSiteDetail)
		g.Post("/sites/:siteId/view", dependencies.Detail.TouchSiteView)
	}
	if cfg.SummaryRoutesEnabled() {
		g.Get("/sites/:siteId/summary", summary.SummaryApi.GetSiteSummary)
		g.Get("/sites/:siteId/targets/:target/summary", summary.SummaryApi.GetTargetSummary)
	}
	if cfg.ReadModelRoutesEnabled() {
		g.Get("/sites/:siteId/targets/:target/latest", dependencies.Detail.GetTargetLatest)
		g.Get("/sites/:siteId/targets/:target/observations", dependencies.Detail.ListTargetObservations)
		g.Get("/sites/:siteId/targets/:target/trend", dependencies.Detail.GetTargetTrend)
		g.Get("/sites/:siteId/targets/:target/changes", dependencies.Detail.GetTargetChanges)
		g.Get("/sites/:siteId/targets/:target/light-probes", dependencies.Detail.GetTargetLightProbes)
	}
}

func registerNavInsightsRoutes(g fiber.Router, insights InsightsAPI) {
	g.Get("/insights/overview", insights.GetOverview)
	g.Get("/insights/metrics/:metricKey/trend", insights.GetMetricTrend)
	g.Get("/sites/:siteId/insights", insights.GetSiteInsights)
}
