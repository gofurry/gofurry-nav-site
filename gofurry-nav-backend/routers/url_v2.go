package routers

import (
	"github.com/gofiber/fiber/v3"
	detail "github.com/gofurry/gofurry-nav-backend/apps/nav/detail/controller"
	summary "github.com/gofurry/gofurry-nav-backend/apps/nav/summary/controller"
)

func navV2Api(g fiber.Router) {
	g.Get("/sites/:siteId/detail", detail.DetailApi.GetSiteDetail)
	g.Get("/sites/:siteId/summary", summary.SummaryApi.GetSiteSummary)
	g.Get("/sites/:siteId/targets/:target/summary", summary.SummaryApi.GetTargetSummary)
	g.Get("/sites/:siteId/targets/:target/latest", detail.DetailApi.GetTargetLatest)
	g.Get("/sites/:siteId/targets/:target/observations", detail.DetailApi.ListTargetObservations)
	g.Get("/sites/:siteId/targets/:target/trend", detail.DetailApi.GetTargetTrend)
	g.Get("/sites/:siteId/targets/:target/changes", detail.DetailApi.GetTargetChanges)
	g.Get("/sites/:siteId/targets/:target/light-probes", detail.DetailApi.GetTargetLightProbes)
}
