package auditadmin

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

type API struct{ service *Service }

func NewAPI(service *Service) *API { return &API{service: service} }

func (api *API) Logs(c fiber.Ctx) error {
	filter, err := auditFilters(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	page, serviceErr := api.service.List(c.Context(), filter)
	if serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}
	return common.NewResponse(c).SuccessWithData(page)
}

func auditFilters(c fiber.Ctx) (Filters, common.Error) {
	page := positiveInt(c.Query("page", c.Query("page_num", "1")), 1)
	pageSize := positiveInt(c.Query("page_size", "20"), 20)
	from, err := auditTime(c.Query("from", ""), false)
	if err != nil {
		return Filters{}, err
	}
	until, err := auditTime(c.Query("to", c.Query("until", "")), true)
	if err != nil {
		return Filters{}, err
	}
	return Filters{
		Operator: strings.TrimSpace(c.Query("operator", c.Query("keyword", ""))),
		Role:     strings.TrimSpace(c.Query("role", "")), Action: strings.TrimSpace(c.Query("action", "")),
		Resource: strings.TrimSpace(c.Query("resource", "")), From: from, Until: until,
		Page: int32(page), PageSize: int32(pageSize),
	}, nil
}

func positiveInt(value string, fallback int64) int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 32)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}

func auditTime(value string, endOfDate bool) (*time.Time, common.Error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return &parsed, nil
	}
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return nil, common.NewValidationError("audit time must be RFC3339 or YYYY-MM-DD")
	}
	if endOfDate {
		parsed = parsed.AddDate(0, 0, 1)
	}
	return &parsed, nil
}
