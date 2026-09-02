package controller

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	collectionmodels "github.com/gofurry/gofurry-admin/internal/app/collectionadmin/models"
	collectionservice "github.com/gofurry/gofurry-admin/internal/app/collectionadmin/service"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

type API struct{ service *collectionservice.Service }

func New(service *collectionservice.Service) *API { return &API{service: service} }

func (api *API) Overview(c fiber.Ctx) error {
	data, err := api.service.Overview(c.Context())
	return respond(c, data, err)
}

func (api *API) Instances(c fiber.Ctx) error {
	limit, offset := pageBounds(c, 50)
	data, err := api.service.Instances(c.Context(), collectionservice.InstanceFilters{
		Domain: strings.TrimSpace(c.Query("domain", "")), View: strings.TrimSpace(c.Query("view", "current")),
		Limit: limit, Offset: offset,
	})
	return respond(c, data, err)
}

func (api *API) Schedules(c fiber.Ctx) error {
	data, err := api.service.Schedules(c.Context())
	return respond(c, data, err)
}

func (api *API) UpdateSchedule(c fiber.Ctx) error {
	var request collectionmodels.ScheduleUpdate
	if err := adminutil.DecodeBody(c, &request); err != nil {
		return common.NewResponse(c).Error(err)
	}
	data, err := api.service.UpdateSchedule(c.Context(), audit.MetaFromFiber(c), c.Params("domain"), int64Param(c, "id"), request)
	return respond(c, data, err)
}

func (api *API) RunSchedule(c fiber.Ctx) error {
	data, err := api.service.RunScheduleNow(c.Context(), audit.MetaFromFiber(c), c.Params("domain"), int64Param(c, "id"))
	return respond(c, data, err)
}

func (api *API) Jobs(c fiber.Ctx) error {
	data, err := api.service.Jobs(c.Context(), filters(c))
	return respond(c, data, err)
}

func (api *API) CreateJobs(c fiber.Ctx) error {
	var request collectionmodels.ManualJobRequest
	if err := adminutil.DecodeBody(c, &request); err != nil {
		return common.NewResponse(c).Error(err)
	}
	data, err := api.service.CreateManualJobs(c.Context(), audit.MetaFromFiber(c), request)
	return respond(c, data, err)
}

func (api *API) Job(c fiber.Ctx) error {
	data, err := api.service.Job(c.Context(), c.Params("domain"), int64Param(c, "id"))
	return respond(c, data, err)
}

func (api *API) CancelJob(c fiber.Ctx) error {
	data, err := api.service.CancelJob(c.Context(), audit.MetaFromFiber(c), c.Params("domain"), int64Param(c, "id"))
	return respond(c, data, err)
}

func (api *API) RetryJob(c fiber.Ctx) error {
	data, err := api.service.RetryJob(c.Context(), audit.MetaFromFiber(c), c.Params("domain"), int64Param(c, "id"))
	return respond(c, data, err)
}

func (api *API) Runs(c fiber.Ctx) error {
	data, err := api.service.Runs(c.Context(), filters(c))
	return respond(c, data, err)
}

func (api *API) Run(c fiber.Ctx) error {
	data, err := api.service.Run(c.Context(), c.Params("domain"), strings.TrimSpace(c.Params("id")))
	return respond(c, data, err)
}

func (api *API) Results(c fiber.Ctx) error {
	limit, offset := pageBounds(c, 50)
	data, err := api.service.Results(c.Context(), c.Params("domain"), strings.TrimSpace(c.Params("id")), collectionservice.ResultFilters{
		GameID: optionalInt64(c.Query("game_id", "")), AppID: optionalInt64(c.Query("appid", "")),
		SiteID: optionalInt64(c.Query("site_id", "")), Target: optionalString(c.Query("target", "")),
		Protocol: strings.TrimSpace(c.Query("protocol", "")), Limit: limit, Offset: offset,
	})
	return respond(c, data, err)
}

func (api *API) Charts(c fiber.Ctx) error {
	window := 24 * time.Hour
	switch c.Query("window", "24h") {
	case "7d":
		window = 7 * 24 * time.Hour
	case "30d":
		window = 30 * 24 * time.Hour
	}
	data, err := api.service.Charts(c.Context(), strings.TrimSpace(c.Query("domain", "")), strings.TrimSpace(c.Query("job_key", "")), window)
	return respond(c, data, err)
}

func respond(c fiber.Ctx, data any, err common.Error) error {
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func filters(c fiber.Ctx) collectionservice.Filters {
	limit, offset := pageBounds(c, 100)
	return collectionservice.Filters{
		Domain: strings.TrimSpace(c.Query("domain", "")), Status: strings.TrimSpace(c.Query("status", "")),
		JobKey: strings.TrimSpace(c.Query("job_key", "")), Trigger: strings.TrimSpace(c.Query("trigger", "")),
		Since: optionalTime(c.Query("since", "")), Until: optionalTime(c.Query("until", "")),
		Limit: limit, Offset: offset,
	}
}

func pageBounds(c fiber.Ctx, fallback int32) (int32, int32) {
	if c.Query("page", "") == "" && c.Query("page_size", "") == "" {
		return int32Query(c, "limit", fallback), int32Query(c, "offset", 0)
	}
	page := int32Query(c, "page", 1)
	pageSize := int32Query(c, "page_size", fallback)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = fallback
	}
	return pageSize, (page - 1) * pageSize
}

func int64Param(c fiber.Ctx, name string) int64 {
	value, _ := strconv.ParseInt(strings.TrimSpace(c.Params(name)), 10, 64)
	return value
}

func int32Query(c fiber.Ctx, name string, fallback int32) int32 {
	value, err := strconv.ParseInt(strings.TrimSpace(c.Query(name, "")), 10, 32)
	if err != nil {
		return fallback
	}
	return int32(value)
}

func optionalInt64(value string) *int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || parsed <= 0 {
		return nil
	}
	return &parsed
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func optionalTime(value string) *time.Time {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return nil
	}
	return &parsed
}
