package routers

import (
	"github.com/gofiber/fiber/v3"
	game "github.com/gofurry/gofurry-game-backend/apps/game/controller"
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
	g.Get("/info", game.GameApi.GetGameInfo)                     // 获取单条游戏的基础信息
	g.Get("/info/list", game.GameApi.GetGameList)                // 获取前 num 条游戏记录
	g.Get("/info/main", game.GameApi.GetGameMainList)            // 获取首页展示数据
	g.Get("/panel/main", game.GameApi.GetPanelMainList)          // 获取首页面板数据
	g.Get("/update/latest", game.GameApi.GetUpdateNews)          // 获取首页更新公告
	g.Get("/update/latest/more", game.GameApi.GetUpdateNewsMore) // 获取更多首页更新公告
	g.Get("/sync/list", game.GameApi.GetGameSyncList)            // 获取同步用的全量游戏列表
	g.Get("/sync/info", game.GameApi.GetGameSyncInfo)            // 获取同步用的游戏详情（不触发浏览量）
	g.Get("/sync/news", game.GameApi.GetGameSyncNews)            // 获取同步用的全量游戏新闻
	g.Get("/sync/creators", game.GameApi.GetGameSyncCreators)    // 获取同步用的创作者列表

	g.Get("/remark", game.GameApi.GetGameRemark) // 获取单条游戏的评论

	g.Get("/tag/list", game.GameApi.GetTagList) // 获取标签列表

	g.Get("/creator", game.GameApi.GetGameCreator) // 获取相关开发者列表

	recommendApi(g.Group("/recommend"))
	searchApi(g.Group("/search"))
	reviewApi(g.Group("/review"))
	prizeApi(g.Group("/prize"))
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
