package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v3"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	v2service "github.com/gofurry/gofurry-game-backend/apps/game/v2/service"
	"github.com/gofurry/gofurry-game-backend/common"
)

type insightsReader interface {
	GetInsightsOverview(context.Context) (v2models.InsightOverview, error)
	GetInsightsMetricTrend(context.Context, string, string) (v2models.InsightMetricTrend, error)
	GetInsightsMetricBreakdown(context.Context, string, string) (v2models.InsightDimensionBreakdown, error)
	GetInsightsMetricSliceTrend(context.Context, string, string, string, string) (v2models.InsightDimensionTrend, error)
	GetInsightsChanges(context.Context, v2models.InsightChangeExplorerQuery) (v2models.InsightChangeExplorerPage, error)
	GetPlayerRanking(context.Context, v2models.InsightPlayerRankingQuery) (v2models.InsightPlayerRanking, error)
	GetPriceOverview(context.Context, string) (v2models.InsightPriceOverview, error)
	GetDiscounts(context.Context, string, int32) (v2models.InsightDiscounts, error)
	GetLanguageOverview(context.Context) (v2models.InsightLanguageOverview, error)
	GetGameCompare(context.Context, string, string) (v2models.GameCompare, error)
	GetGameInsights(context.Context, int64) (v2models.GameInsights, error)
	GetGamePlayerInsights(context.Context, int64, string) (v2models.InsightPlayerHistory, error)
	GetGamePriceInsights(context.Context, int64, string, string) (v2models.InsightPriceHistory, error)
}

func (api *GameV2API) GetInsightsOverview(c fiber.Ctx) error {
	data, err := api.insights.GetInsightsOverview(context.Background())
	return respondInsights(c, data, err)
}

func (api *GameV2API) GetInsightsMetricTrend(c fiber.Ctx) error {
	data, err := api.insights.GetInsightsMetricTrend(context.Background(), c.Params("metricKey"), c.Query("range", "30d"))
	return respondInsights(c, data, err)
}

func (api *GameV2API) GetInsightsMetricBreakdown(c fiber.Ctx) error {
	data, err := api.insights.GetInsightsMetricBreakdown(context.Background(), c.Params("metricKey"), c.Query("dimension"))
	return respondInsights(c, data, err)
}

func (api *GameV2API) GetInsightsMetricSliceTrend(c fiber.Ctx) error {
	data, err := api.insights.GetInsightsMetricSliceTrend(
		context.Background(), c.Params("metricKey"), c.Params("dimension"), c.Params("value"), c.Query("range", "30d"),
	)
	return respondInsights(c, data, err)
}

func (api *GameV2API) GetInsightsChanges(c fiber.Ctx) error {
	limit, err := strconv.ParseInt(c.Query("limit", "20"), 10, 32)
	if err != nil {
		return common.NewResponse(c).ErrorWithCode(v2service.ErrInvalidInsightChanges.Error(), http.StatusBadRequest)
	}
	data, err := api.insights.GetInsightsChanges(context.Background(), v2models.InsightChangeExplorerQuery{
		Range: c.Query("range", "30d"), Category: c.Query("category"), Type: c.Query("type"),
		Cursor: c.Query("cursor"), Limit: int32(limit),
	})
	return respondInsights(c, data, err)
}

func (api *GameV2API) GetPlayerRanking(c fiber.Ctx) error {
	limit, err := strconv.ParseInt(c.Query("limit", "20"), 10, 32)
	if err != nil {
		return common.NewResponse(c).ErrorWithCode(v2service.ErrInvalidInsightLimit.Error(), http.StatusBadRequest)
	}
	data, err := api.insights.GetPlayerRanking(context.Background(), v2models.InsightPlayerRankingQuery{Metric: c.Query("metric", "latest_observed"), Limit: int32(limit)})
	return respondInsights(c, data, err)
}

func (api *GameV2API) GetPriceOverview(c fiber.Ctx) error {
	data, err := api.insights.GetPriceOverview(context.Background(), c.Query("region", "CN"))
	return respondInsights(c, data, err)
}

func (api *GameV2API) GetDiscounts(c fiber.Ctx) error {
	limit, err := strconv.ParseInt(c.Query("limit", "20"), 10, 32)
	if err != nil {
		return common.NewResponse(c).ErrorWithCode(v2service.ErrInvalidInsightLimit.Error(), http.StatusBadRequest)
	}
	data, err := api.insights.GetDiscounts(context.Background(), c.Query("region", "CN"), int32(limit))
	return respondInsights(c, data, err)
}

func (api *GameV2API) GetLanguageOverview(c fiber.Ctx) error {
	data, err := api.insights.GetLanguageOverview(context.Background())
	return respondInsights(c, data, err)
}

func (api *GameV2API) GetGameCompare(c fiber.Ctx) error {
	data, err := api.insights.GetGameCompare(context.Background(), c.Query("ids"), c.Query("region", "CN"))
	return respondInsights(c, data, err)
}

func (api *GameV2API) GetGameInsights(c fiber.Ctx) error {
	gameID := parseInt64(c.Params("gameId", "0"))
	if gameID <= 0 {
		return common.NewResponse(c).ErrorWithCode("invalid game id", http.StatusBadRequest)
	}
	data, err := api.insights.GetGameInsights(context.Background(), gameID)
	return respondInsights(c, data, err)
}

func (api *GameV2API) GetGamePlayerInsights(c fiber.Ctx) error {
	gameID := parseInt64(c.Params("gameId", "0"))
	if gameID <= 0 {
		return common.NewResponse(c).ErrorWithCode("invalid game id", http.StatusBadRequest)
	}
	data, err := api.insights.GetGamePlayerInsights(context.Background(), gameID, c.Query("range", "30d"))
	return respondInsights(c, data, err)
}

func (api *GameV2API) GetGamePriceInsights(c fiber.Ctx) error {
	gameID := parseInt64(c.Params("gameId", "0"))
	if gameID <= 0 {
		return common.NewResponse(c).ErrorWithCode("invalid game id", http.StatusBadRequest)
	}
	data, err := api.insights.GetGamePriceInsights(context.Background(), gameID, c.Query("region", "CN"), c.Query("range", "30d"))
	return respondInsights(c, data, err)
}

func respondInsights(c fiber.Ctx, data any, err error) error {
	if err == nil {
		return common.NewResponse(c).SuccessWithData(data)
	}
	switch {
	case errors.Is(err, v2service.ErrInvalidInsightMetric), errors.Is(err, v2service.ErrInvalidInsightRange),
		errors.Is(err, v2service.ErrInvalidInsightDimension), errors.Is(err, v2service.ErrInvalidInsightSlice),
		errors.Is(err, v2service.ErrInvalidInsightChanges), errors.Is(err, v2service.ErrInvalidInsightCursor):
		return common.NewResponse(c).ErrorWithCode(err.Error(), http.StatusBadRequest)
	case errors.Is(err, v2service.ErrInvalidInsightRegion), errors.Is(err, v2service.ErrInvalidPlayerRanking),
		errors.Is(err, v2service.ErrInvalidInsightCompare),
		errors.Is(err, v2service.ErrInvalidInsightLimit):
		return common.NewResponse(c).ErrorWithCode(err.Error(), http.StatusBadRequest)
	case errors.Is(err, v2service.ErrInsightGameNotFound):
		return common.NewResponse(c).ErrorWithCode(err.Error(), http.StatusNotFound)
	default:
		return common.NewResponse(c).ErrorWithCode("internal insights error", http.StatusInternalServerError)
	}
}
