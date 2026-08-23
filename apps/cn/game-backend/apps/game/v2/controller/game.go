package controller

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v3"
	v2dao "github.com/gofurry/gofurry-game-backend/apps/game/v2/dao"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	v2service "github.com/gofurry/gofurry-game-backend/apps/game/v2/service"
	reviewmodels "github.com/gofurry/gofurry-game-backend/apps/review/models"
	reviewservice "github.com/gofurry/gofurry-game-backend/apps/review/service"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/util"
)

type GameV2API struct {
	readModelDAO  *v2dao.ReadModelDAO
	viewService   *v2service.GameViewService
	reviewService *reviewservice.ReviewService
}

func New(readModelDAO *v2dao.ReadModelDAO, viewService *v2service.GameViewService, reviewService *reviewservice.ReviewService) *GameV2API {
	return &GameV2API{readModelDAO: readModelDAO, viewService: viewService, reviewService: reviewService}
}

func (api *GameV2API) GetGameList(c fiber.Ctx) error {
	data, err := api.newReadModelService().GetGameList(context.Background(), v2models.GameV2ListQuery{
		Lang:   c.Query("lang", "zh"),
		Region: c.Query("region", "CN"),
		Limit:  parseInt(c.Query("limit", "20")),
		Offset: parseInt(c.Query("offset", "0")),
		Sort:   c.Query("sort", "weight"),
	})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *GameV2API) GetGameInfo(c fiber.Ctx) error {
	id := parseInt64(c.Query("id", "0"))
	appid := parseInt64(c.Query("appid", "0"))
	if id <= 0 && appid <= 0 {
		return common.NewResponse(c).Error("id 或 appid 不能为空")
	}
	data, err := api.newReadModelService().GetGameDetail(context.Background(), v2models.GameV2DetailRequest{
		GameID:    id,
		AppID:     appid,
		Lang:      c.Query("lang", "zh"),
		Region:    c.Query("region", "CN"),
		NewsLimit: parseInt(c.Query("news_limit", "5")),
	})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *GameV2API) TouchGameView(c fiber.Ctx) error {
	gameID := parseInt64(c.Params("id", "0"))
	if gameID <= 0 {
		return common.NewResponse(c).Error("id 不能为空")
	}

	clientIP := util.GetClientIP(c)
	if clientIP == "" {
		clientIP = c.IP()
	}

	viewCount, err := api.newGameViewService().TouchGameViewCount(gameID, clientIP)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(v2models.GameV2ViewTouchResponse{
		GameID:    gameID,
		ViewCount: viewCount,
	})
}

func (api *GameV2API) SearchSimple(c fiber.Ctx) error {
	req := v2models.GameV2SearchRequest{}
	if err := c.Bind().Body(&req); err != nil {
		return common.NewResponse(c).Error("解析请求体失败")
	}
	data, err := api.newReadModelService().SimpleSearch(context.Background(), req)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *GameV2API) SearchPage(c fiber.Ctx) error {
	req := v2models.GameV2SearchPageQueryRequest{}
	if err := c.Bind().Body(&req); err != nil {
		return common.NewResponse(c).Error("解析请求体失败")
	}
	data, err := api.newReadModelService().SearchPage(context.Background(), req)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *GameV2API) GetTags(c fiber.Ctx) error {
	data, err := api.newReadModelService().ListTags(context.Background(), c.Query("lang", "zh"))
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *GameV2API) GetGameReviews(c fiber.Ctx) error {
	data, err := api.newReadModelService().GetGameReviews(
		context.Background(),
		c.Query("id", "0"),
		parseInt(c.Query("page", "1")),
		parseInt(c.Query("limit", "5")),
	)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *GameV2API) AddAnonymousReview(c fiber.Ctx) error {
	req := reviewmodels.AnonymousReviewRequest{}
	if err := c.Bind().Body(&req); err != nil {
		return common.NewResponse(c).Error("解析请求体失败")
	}
	if err := api.reviewService.AddAnonymousReview(req, c); err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).Success()
}

func (api *GameV2API) GetLatestReviews(c fiber.Ctx) error {
	data, err := api.newReadModelService().ListLatestReviews(context.Background(), c.Query("lang", "zh"), parseInt(c.Query("limit", "15")))
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *GameV2API) GetRandomGame(c fiber.Ctx) error {
	data, err := api.newReadModelService().GetRandomGameID(context.Background())
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *GameV2API) GetSimilarRecommendations(c fiber.Ctx) error {
	data, err := api.newReadModelService().GetSimilarRecommendations(context.Background(), v2models.GameV2SimilarRecommendationQuery{
		GameID: parseInt64(c.Query("id", "0")),
		Lang:   c.Query("lang", "zh"),
		Region: c.Query("region", "CN"),
		Limit:  parseInt(c.Query("limit", "8")),
	})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *GameV2API) GetGameNews(c fiber.Ctx) error {
	id := parseInt64(c.Query("id", "0"))
	appid := parseInt64(c.Query("appid", "0"))
	if id <= 0 && appid <= 0 {
		return common.NewResponse(c).Error("id 或 appid 不能为空")
	}
	data, err := api.newReadModelService().GetGameNews(context.Background(), v2models.GameV2NewsQuery{
		GameID: id,
		AppID:  appid,
		Lang:   c.Query("lang", "zh"),
		Limit:  parseInt(c.Query("limit", "20")),
		Offset: parseInt(c.Query("offset", "0")),
	})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *GameV2API) GetLatestGameNews(c fiber.Ctx) error {
	data, err := api.newReadModelService().GetLatestGameNews(context.Background(), v2models.GameV2NewsQuery{
		Lang:   c.Query("lang", "zh"),
		Limit:  parseInt(c.Query("limit", "20")),
		Offset: parseInt(c.Query("offset", "0")),
	})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *GameV2API) GetPanelMain(c fiber.Ctx) error {
	data, err := api.newReadModelService().GetPanelMain(context.Background(), v2models.GameV2PanelQuery{
		Lang:           c.Query("lang", "zh"),
		Region:         c.Query("region", "CN"),
		Limit:          parseInt(c.Query("limit", "8")),
		TopOnlineLimit: parseInt(c.Query("top_online_limit", "60")),
		PriceLimit:     parseInt(c.Query("price_limit", "120")),
		NewsLimit:      parseInt(c.Query("news_limit", "8")),
	})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *GameV2API) GetHome(c fiber.Ctx) error {
	data, err := api.newReadModelService().GetHome(context.Background(), c.Query("lang", "zh"), c.Query("region", "CN"))
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *GameV2API) newReadModelService() *v2service.ReadModelService {
	return v2service.NewReadModelServiceWithReader(api.readModelDAO)
}

func (api *GameV2API) newGameViewService() *v2service.GameViewService {
	return api.viewService
}

func parseInt(value string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func parseInt64(value string) int64 {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}
