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
	GetMetricBreakdown(context.Context, string, string) (models.DimensionBreakdown, error)
	GetMetricSliceTrend(context.Context, string, string, string, string) (models.DimensionTrend, error)
	GetChanges(context.Context, models.ChangeExplorerQuery) (models.ChangeExplorerPage, error)
	GetCertificateOverview(context.Context, int32) (models.CertificateOverview, error)
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

func (api *InsightsAPI) GetMetricBreakdown(c fiber.Ctx) error {
	data, err := api.service.GetMetricBreakdown(context.Background(), c.Params("metricKey"), c.Query("dimension"))
	return respond(c, data, err)
}

func (api *InsightsAPI) GetMetricSliceTrend(c fiber.Ctx) error {
	data, err := api.service.GetMetricSliceTrend(
		context.Background(), c.Params("metricKey"), c.Params("dimension"), c.Params("value"), c.Query("range", "30d"),
	)
	return respond(c, data, err)
}

func (api *InsightsAPI) GetChanges(c fiber.Ctx) error {
	limit, err := strconv.ParseInt(c.Query("limit", "20"), 10, 32)
	if err != nil {
		return common.NewResponse(c).ErrorWithCode(service.ErrInvalidChanges.Error(), http.StatusBadRequest)
	}
	data, err := api.service.GetChanges(context.Background(), models.ChangeExplorerQuery{
		Range: c.Query("range", "30d"), Category: c.Query("category"), Type: c.Query("type"),
		Cursor: c.Query("cursor"), Limit: int32(limit),
	})
	return respond(c, data, err)
}

func (api *InsightsAPI) GetCertificateOverview(c fiber.Ctx) error {
	limit, err := strconv.ParseInt(c.Query("limit", "20"), 10, 32)
	if err != nil {
		return common.NewResponse(c).ErrorWithCode(service.ErrInvalidCertificateLimit.Error(), http.StatusBadRequest)
	}
	data, err := api.service.GetCertificateOverview(context.Background(), int32(limit))
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
	case errors.Is(err, service.ErrInvalidMetricKey), errors.Is(err, service.ErrInvalidRange),
		errors.Is(err, service.ErrInvalidDimension), errors.Is(err, service.ErrInvalidSlice),
		errors.Is(err, service.ErrInvalidChanges), errors.Is(err, service.ErrInvalidCursor),
		errors.Is(err, service.ErrInvalidCertificateLimit):
		return common.NewResponse(c).ErrorWithCode(err.Error(), http.StatusBadRequest)
	case errors.Is(err, service.ErrNotFound):
		return common.NewResponse(c).ErrorWithCode(err.Error(), http.StatusNotFound)
	default:
		return common.NewResponse(c).ErrorWithCode("internal insights error", http.StatusInternalServerError)
	}
}
