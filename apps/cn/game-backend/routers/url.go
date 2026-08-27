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
