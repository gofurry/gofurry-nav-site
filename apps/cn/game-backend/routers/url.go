package routers

import (
	"github.com/gofiber/fiber/v3"
	gamev2 "github.com/gofurry/gofurry-game-backend/apps/game/v2/controller"
	prize "github.com/gofurry/gofurry-game-backend/apps/prize/controller"
)

/*
 * @Desc: 接口层
 * @author: 福狼
 * @version: v1.0.0
 */

func gameV2Api(g fiber.Router, gameAPI *gamev2.GameV2API, prizeAPI *prize.PrizeAPI) {
	registerGameInsightRoutes(g, gameAPI)
	g.Get("/list", gameAPI.GetGameList)
	g.Get("/info", gameAPI.GetGameInfo)
	g.Get("/tags", gameAPI.GetTags)
	g.Get("/news", gameAPI.GetGameNews)
	g.Get("/news/latest", gameAPI.GetLatestGameNews)
	g.Get("/home", gameAPI.GetHome)
	g.Get("/panel/main", gameAPI.GetPanelMain)
	g.Post("/games/:id/view", gameAPI.TouchGameView)
	g.Post("/search/simple", gameAPI.SearchSimple)
	g.Post("/search/page", gameAPI.SearchPage)
	g.Get("/reviews", gameAPI.GetGameReviews)
	g.Post("/reviews/anonymous", gameAPI.AddAnonymousReview)
	g.Get("/reviews/latest", gameAPI.GetLatestReviews)
	g.Get("/recommend/random", gameAPI.GetRandomGame)
	g.Get("/recommend/similar", gameAPI.GetSimilarRecommendations)
	g.Get("/prizes", prizeAPI.LotteryInfo)
	g.Post("/prizes/participation", prizeAPI.PrizeParticipation)
	g.Get("/prizes/participation/activation", prizeAPI.ActiveParticipation)

}

type gameInsightsAPI interface {
	GetInsightsOverview(fiber.Ctx) error
	GetInsightsMetricTrend(fiber.Ctx) error
	GetInsightsMetricBreakdown(fiber.Ctx) error
	GetInsightsMetricSliceTrend(fiber.Ctx) error
	GetInsightsChanges(fiber.Ctx) error
	GetPlayerRanking(fiber.Ctx) error
	GetPriceOverview(fiber.Ctx) error
	GetDiscounts(fiber.Ctx) error
	GetLanguageOverview(fiber.Ctx) error
	GetGameInsights(fiber.Ctx) error
	GetGamePlayerInsights(fiber.Ctx) error
	GetGamePriceInsights(fiber.Ctx) error
}

func registerGameInsightRoutes(g fiber.Router, insights gameInsightsAPI) {
	g.Get("/insights/overview", insights.GetInsightsOverview)
	g.Get("/insights/metrics/:metricKey/trend", insights.GetInsightsMetricTrend)
	g.Get("/insights/metrics/:metricKey/breakdown", insights.GetInsightsMetricBreakdown)
	g.Get("/insights/metrics/:metricKey/breakdown/:dimension/:value/trend", insights.GetInsightsMetricSliceTrend)
	g.Get("/insights/changes", insights.GetInsightsChanges)
	g.Get("/insights/players/ranking", insights.GetPlayerRanking)
	g.Get("/insights/prices/overview", insights.GetPriceOverview)
	g.Get("/insights/prices/discounts", insights.GetDiscounts)
	g.Get("/insights/languages/overview", insights.GetLanguageOverview)
	g.Get("/games/:gameId/insights", insights.GetGameInsights)
	g.Get("/games/:gameId/insights/players", insights.GetGamePlayerInsights)
	g.Get("/games/:gameId/insights/prices", insights.GetGamePriceInsights)
}
