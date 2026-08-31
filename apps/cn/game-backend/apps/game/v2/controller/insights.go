package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	v2service "github.com/gofurry/gofurry-game-backend/apps/game/v2/service"
	"github.com/gofurry/gofurry-game-backend/common"
)

type insightsReader interface {
	GetInsightsOverview(context.Context) (v2models.InsightOverview, error)
	GetInsightsMetricTrend(context.Context, string, string) (v2models.InsightMetricTrend, error)
	GetGameInsights(context.Context, int64) (v2models.GameInsights, error)
	GetGamePlayerInsights(context.Context, int64, string) (v2models.InsightPlayerHistory, error)
	GetGamePriceInsights(context.Context, int64, string) (v2models.InsightPriceHistory, error)
}

func (api *GameV2API) GetInsightsOverview(c fiber.Ctx) error {
	data, err := api.insights.GetInsightsOverview(context.Background())
	return respondInsights(c, data, err)
}

func (api *GameV2API) GetInsightsMetricTrend(c fiber.Ctx) error {
	data, err := api.insights.GetInsightsMetricTrend(context.Background(), c.Params("metricKey"), c.Query("range", "30d"))
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
	data, err := api.insights.GetGamePriceInsights(context.Background(), gameID, c.Query("range", "30d"))
	return respondInsights(c, data, err)
}

func respondInsights(c fiber.Ctx, data any, err error) error {
	if err == nil {
		return common.NewResponse(c).SuccessWithData(data)
	}
	switch {
	case errors.Is(err, v2service.ErrInvalidInsightMetric), errors.Is(err, v2service.ErrInvalidInsightRange):
		return common.NewResponse(c).ErrorWithCode(err.Error(), http.StatusBadRequest)
	case errors.Is(err, v2service.ErrInsightGameNotFound):
		return common.NewResponse(c).ErrorWithCode(err.Error(), http.StatusNotFound)
	default:
		return common.NewResponse(c).ErrorWithCode("internal insights error", http.StatusInternalServerError)
	}
}
