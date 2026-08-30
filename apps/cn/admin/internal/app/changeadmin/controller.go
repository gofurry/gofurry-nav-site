package changeadmin

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
func (api *API) Events(c fiber.Ctx) error {
	filter, err := parseFilters(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	data, appErr := api.service.Events(c.Context(), filter)
	return respond(c, data, appErr)
}

func parseFilters(c fiber.Ctx) (Filters, common.Error) {
	version, err := nonNegativeInt(c.Query("version", "0"), "version", 32)
	if err != nil {
		return Filters{}, err
	}
	entity, err := nonNegativeInt(c.Query("entity_id", "0"), "entity_id", 64)
	if err != nil {
		return Filters{}, err
	}
	page, err := nonNegativeInt(c.Query("page", "1"), "page", 32)
	if err != nil {
		return Filters{}, err
	}
	pageSize, err := nonNegativeInt(c.Query("page_size", "50"), "page_size", 32)
	if err != nil {
		return Filters{}, err
	}
	from, err := optionalDate(c.Query("from", ""), "from")
	if err != nil {
		return Filters{}, err
	}
	through, err := optionalDate(c.Query("to", c.Query("through", "")), "to")
	if err != nil {
		return Filters{}, err
	}
	return Filters{Domain: strings.TrimSpace(c.Query("domain", "")), DetectorKey: strings.TrimSpace(c.Query("detector_key", c.Query("detector", ""))), DetectorVersion: int32(version), Status: strings.TrimSpace(c.Query("status", "")), From: from, Through: through, EventCode: strings.TrimSpace(c.Query("event_code", "")), ScopeKind: strings.TrimSpace(c.Query("scope_kind", "")), ScopeKey: strings.TrimSpace(c.Query("scope_key", "")), EntityID: entity, Page: int32(page), PageSize: int32(pageSize)}, nil
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
func nonNegativeInt(value, name string, bits int) (int64, common.Error) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, bits)
	if err != nil || parsed < 0 {
		return 0, common.NewValidationError(name + " must be a non-negative integer")
	}
	return parsed, nil
}
func respond(c fiber.Ctx, data any, err common.Error) error {
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(data)
}
