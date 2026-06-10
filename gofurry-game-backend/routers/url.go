package routers

import (
	"github.com/gofiber/fiber/v3"
	game "github.com/gofurry/gofurry-game-backend/apps/game/controller"
	gamev2 "github.com/gofurry/gofurry-game-backend/apps/game/v2/controller"
	prize "github.com/gofurry/gofurry-game-backend/apps/prize/controller"
	recommend "github.com/gofurry/gofurry-game-backend/apps/recommend/controller"
	review "github.com/gofurry/gofurry-game-backend/apps/review/controller"
	search "github.com/gofurry/gofurry-game-backend/apps/search/controller"
)

/*
 * @Desc: 接口层
 * @author: 福狼
 * @version: v1.0.0
 */

func gameApi(g fiber.Router) {
	g.Get("/remark", game.GameApi.GetGameRemark) // 获取单条游戏的评论

	g.Get("/tag/list", game.GameApi.GetTagList) // 获取标签列表

	g.Get("/creator", game.GameApi.GetGameCreator) // 获取相关开发者列表

	recommendApi(g.Group("/recommend"))
	searchApi(g.Group("/search"))
	reviewApi(g.Group("/review"))
	prizeApi(g.Group("/prize"))
}

func gameV2Api(g fiber.Router) {
	g.Get("/list", gamev2.GameV2Api.GetGameList)
	g.Get("/info", gamev2.GameV2Api.GetGameInfo)
	g.Get("/news", gamev2.GameV2Api.GetGameNews)
	g.Get("/news/latest", gamev2.GameV2Api.GetLatestGameNews)
	g.Get("/panel/main", gamev2.GameV2Api.GetPanelMain)
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
	g.Get("/CBF", recommend.RecommendApi.RecommendByCBF)     // 用 CBF 返回游戏记录
	g.Get("/random", recommend.RecommendApi.GetRandomGameID) // 返回一个随机的游戏记录 ID
}

func searchApi(g fiber.Router) {
	g.Post("/simple", search.SearchApi.SimpleSearch) // 简易搜索
	g.Post("/page", search.SearchApi.PageSearch)     // 复杂查询
}

func reviewApi(g fiber.Router) {
	g.Post("/anonymous", review.ReviewApi.AddAnonymousReview) // 匿名评论
	g.Get("/latest", review.ReviewApi.GetLatestReviewList)    // 获取最新的评论列表
}

func prizeApi(g fiber.Router) {
	g.Post("/participation", prize.PrizeApi.PrizeParticipation)            // 参与抽奖
	g.Get("/participation/activation", prize.PrizeApi.ActiveParticipation) // 确认参与

	g.Get("/info", prize.PrizeApi.LotteryInfo) // 抽奖页展示数据
}
