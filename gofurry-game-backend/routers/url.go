package routers

import (
	"github.com/gofiber/fiber/v3"
	game "github.com/gofurry/gofurry-game-backend/apps/game/controller"
	gamev2 "github.com/gofurry/gofurry-game-backend/apps/game/v2/controller"
	prize "github.com/gofurry/gofurry-game-backend/apps/prize/controller"
	recommend "github.com/gofurry/gofurry-game-backend/apps/recommend/controller"
)

/*
 * @Desc: 接口层
 * @author: 福狼
 * @version: v1.0.0
 */

func gameApi(g fiber.Router) {
	g.Get("/creator", game.GameApi.GetGameCreator) // 获取相关开发者列表

	recommendApi(g.Group("/recommend"))
	prizeApi(g.Group("/prize"))
}

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
	g.Get("/sync/list", gamev2.GameV2Api.GetSyncGameList)
	g.Get("/sync/info", gamev2.GameV2Api.GetSyncGameInfo)
	g.Get("/sync/news", gamev2.GameV2Api.GetSyncGameNews)
	g.Get("/sync/creators", gamev2.GameV2Api.GetSyncCreators)

	collect := g.Group("/collect", gamev2.RequireAdminToken())
	collect.Get("/status", gamev2.GameV2Api.GetCollectStatus)
	collect.Get("/runs", gamev2.GameV2Api.ListCollectRuns)
	collect.Get("/runs/:run_id", gamev2.GameV2Api.GetCollectRun)
	collect.Get("/task-results", gamev2.GameV2Api.ListCollectTaskResults)
	collect.Get("/games/:id/status", gamev2.GameV2Api.GetGameCollectStatus)
}

func recommendApi(g fiber.Router) {
	// TODO: 标签表新增一个权重字段

	// 基于内容的推荐（Content-based Filtering）
	// 优点: 存储小 速度快 无冷启动 无需用户行为数据
	// 缺点: 需要传入初始物品, 特征值永远为静态, 每次推荐相同
	// 实现重点: 余弦相似度 特征提取-独热编码
	g.Get("/CBF", recommend.RecommendApi.RecommendByCBF) // 用 CBF 返回游戏记录
}

func prizeApi(g fiber.Router) {
	g.Post("/participation", prize.PrizeApi.PrizeParticipation)            // 参与抽奖
	g.Get("/participation/activation", prize.PrizeApi.ActiveParticipation) // 确认参与

	g.Get("/info", prize.PrizeApi.LotteryInfo) // 抽奖页展示数据
}
