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

func gameV2Api(g fiber.Router) {
	g.Get("/list", gamev2.GameV2Api.GetGameList)
	g.Get("/info", gamev2.GameV2Api.GetGameInfo)
	g.Get("/tags", gamev2.GameV2Api.GetTags)
	g.Get("/news", gamev2.GameV2Api.GetGameNews)
	g.Get("/news/latest", gamev2.GameV2Api.GetLatestGameNews)
	g.Get("/panel/main", gamev2.GameV2Api.GetPanelMain)
	g.Post("/search/simple", gamev2.GameV2Api.SearchSimple)
	g.Post("/search/page", gamev2.GameV2Api.SearchPage)
	g.Get("/reviews", gamev2.GameV2Api.GetGameReviews)
	g.Post("/reviews/anonymous", gamev2.GameV2Api.AddAnonymousReview)
	g.Get("/reviews/latest", gamev2.GameV2Api.GetLatestReviews)
	g.Get("/recommend/random", gamev2.GameV2Api.GetRandomGame)
	g.Get("/recommend/similar", gamev2.GameV2Api.GetSimilarRecommendations)
	g.Get("/prizes", prize.PrizeApi.LotteryInfo)
	g.Post("/prizes/participation", prize.PrizeApi.PrizeParticipation)
	g.Get("/prizes/participation/activation", prize.PrizeApi.ActiveParticipation)
	g.Get("/sync/list", gamev2.GameV2Api.GetSyncGameList)
	g.Get("/sync/info", gamev2.GameV2Api.GetSyncGameInfo)
	g.Get("/sync/news", gamev2.GameV2Api.GetSyncGameNews)

	collect := g.Group("/collect", gamev2.RequireAdminToken())
	collect.Get("/status", gamev2.GameV2Api.GetCollectStatus)
	collect.Get("/runs", gamev2.GameV2Api.ListCollectRuns)
	collect.Get("/runs/:run_id", gamev2.GameV2Api.GetCollectRun)
	collect.Get("/task-results", gamev2.GameV2Api.ListCollectTaskResults)
	collect.Get("/games/:id/status", gamev2.GameV2Api.GetGameCollectStatus)
}
