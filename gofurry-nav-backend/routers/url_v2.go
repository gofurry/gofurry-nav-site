package routers

import (
	"github.com/gofiber/fiber/v3"
	summary "github.com/gofurry/gofurry-nav-backend/apps/nav/summary/controller"
)

func navV2Api(g fiber.Router) {
	g.Get("/sites/:siteId/summary", summary.SummaryApi.GetSiteSummary)
	g.Get("/sites/:siteId/targets/:target/summary", summary.SummaryApi.GetTargetSummary)
}
