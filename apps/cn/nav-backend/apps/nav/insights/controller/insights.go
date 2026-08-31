package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/models"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/service"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type insightsReader interface {
	GetOverview(context.Context) (models.Overview, error)
	GetMetricTrend(context.Context, string, string) (models.MetricTrend, error)
	GetSiteInsights(context.Context, int64) (models.SiteInsights, error)
}

type InsightsAPI struct{ service insightsReader }

func New(reader insightsReader) *InsightsAPI { return &InsightsAPI{service: reader} }

func (api *InsightsAPI) GetOverview(c fiber.Ctx) error {
	data, err := api.service.GetOverview(context.Background())
	return respond(c, data, err)
}

func (api *InsightsAPI) GetMetricTrend(c fiber.Ctx) error {
	data, err := api.service.GetMetricTrend(context.Background(), c.Params("metricKey"), c.Query("range", "30d"))
	return respond(c, data, err)
}

func (api *InsightsAPI) GetSiteInsights(c fiber.Ctx) error {
	siteID, err := strconv.ParseInt(c.Params("siteId"), 10, 64)
	if err != nil || siteID <= 0 {
		return common.NewResponse(c).ErrorWithCode("invalid site id", http.StatusBadRequest)
	}
	data, err := api.service.GetSiteInsights(context.Background(), siteID)
	return respond(c, data, err)
}

func respond(c fiber.Ctx, data any, err error) error {
	if err == nil {
		return common.NewResponse(c).SuccessWithData(data)
	}
	switch {
	case errors.Is(err, service.ErrInvalidMetricKey), errors.Is(err, service.ErrInvalidRange):
		return common.NewResponse(c).ErrorWithCode(err.Error(), http.StatusBadRequest)
	case errors.Is(err, service.ErrNotFound):
		return common.NewResponse(c).ErrorWithCode(err.Error(), http.StatusNotFound)
	default:
		return common.NewResponse(c).ErrorWithCode("internal insights error", http.StatusInternalServerError)
	}
}
