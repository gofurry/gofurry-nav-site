package metricadmin

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

type API struct{ service *Service }

func NewAPI(service *Service) *API { return &API{service: service} }

func (api *API) Overview(c fiber.Ctx) error {
	data, err := api.service.Overview(c.Context(), strings.TrimSpace(c.Query("domain", "")))
	return respond(c, data, err)
}

func (api *API) Registry(c fiber.Ctx) error {
	filter, err := parseFilters(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	data, appErr := api.service.Registry(c.Context(), filter)
	return respond(c, data, appErr)
}

func (api *API) Checkpoints(c fiber.Ctx) error {
	filter, err := parseFilters(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	data, appErr := api.service.Checkpoints(c.Context(), filter)
	return respond(c, data, appErr)
}

func (api *API) Daily(c fiber.Ctx) error {
	filter, err := parseFilters(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	data, appErr := api.service.Daily(c.Context(), filter)
	return respond(c, data, appErr)
}

func (api *API) Entities(c fiber.Ctx) error {
	filter, err := parseFilters(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if filter.FactDate.IsZero() {
		return common.NewResponse(c).Error(common.NewValidationError("fact_date is required"))
	}
	data, appErr := api.service.Entities(c.Context(), filter)
	return respond(c, data, appErr)
}

func parseFilters(c fiber.Ctx) (Filters, common.Error) {
	version, err := int32Query(c.Query("version", "0"), "version")
	if err != nil {
		return Filters{}, err
	}
	page, err := int32Query(c.Query("page", "1"), "page")
	if err != nil {
		return Filters{}, err
	}
	pageSize, err := int32Query(c.Query("page_size", "50"), "page_size")
	if err != nil {
		return Filters{}, err
	}
	from, err := optionalDate(c.Query("from", ""), "from")
	if err != nil {
		return Filters{}, err
	}
	throughText := c.Query("to", c.Query("through", ""))
	through, err := optionalDate(throughText, "to")
	if err != nil {
		return Filters{}, err
	}
	factDate, err := optionalDate(c.Query("fact_date", ""), "fact_date")
	if err != nil {
		return Filters{}, err
	}
	filter := Filters{
		Domain: strings.TrimSpace(c.Query("domain", "")), MetricKey: strings.TrimSpace(c.Query("metric_key", c.Query("metric", ""))),
		MetricVersion: version, Status: strings.TrimSpace(c.Query("status", "")), From: from, Through: through,
		DimensionKey:   strings.TrimSpace(c.Query("dimension_key", "global")),
		DimensionValue: strings.TrimSpace(c.Query("dimension_value", "all")), State: strings.TrimSpace(c.Query("state", "")),
		ReasonCode: strings.TrimSpace(c.Query("reason_code", "")), Page: page, PageSize: pageSize,
	}
	if factDate != nil {
		filter.FactDate = *factDate
	}
	return filter, nil
}

func optionalDate(value, name string) (*time.Time, common.Error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return nil, common.NewValidationError(name + " must be YYYY-MM-DD")
	}
	return &parsed, nil
}

func int32Query(value, name string) (int32, common.Error) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 32)
	if err != nil || parsed < 0 {
		return 0, common.NewValidationError(name + " must be a non-negative integer")
	}
	return int32(parsed), nil
}

func respond(c fiber.Ctx, data any, err common.Error) error {
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(data)
}
