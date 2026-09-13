package cloudops

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	"github.com/gofurry/gofurry-admin/internal/infra/assets"
	infra "github.com/gofurry/gofurry-admin/internal/infra/cloudops"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

type API struct {
	service *infra.Service
	audit   *audit.Logger
}

func New(service *infra.Service, logger *audit.Logger) *API {
	return &API{service: service, audit: logger}
}
func (api *API) Overview(c fiber.Ctx) error {
	return common.NewResponse(c).SuccessWithData(api.service.Overview(c.Context()))
}
func (api *API) Object(c fiber.Ctx) error {
	result, err := api.service.Inspect(c.Context(), strings.TrimSpace(c.Query("key")))
	if err != nil {
		return common.NewResponse(c).Error(common.NewValidationError(err.Error()))
	}
	return common.NewResponse(c).SuccessWithData(result)
}

type operationResult struct {
	Result   any      `json:"result"`
	Warnings []string `json:"warnings,omitempty"`
}

// Durable intent precedes the remote operation. Completion is a second event
// under the same action/request ID because cloud effects cannot be rolled back.
func (api *API) operate(c fiber.Ctx, action, target string, input any, run func(context.Context) (any, error)) error {
	meta := audit.MetaFromFiber(c)
	if err := api.audit.Log(c.Context(), meta, action, "cloud", target, nil, map[string]any{"status": "requested", "input": input}); err != nil {
		return common.NewResponse(c).Error(err)
	}
	result, err := run(c.Context())
	status := "completed"
	detail := map[string]any{"status": status, "result": result}
	if err != nil {
		detail = map[string]any{"status": "failed", "error": err.Error()}
	}
	// A canceled browser request must not prevent writing the result audit.
	auditCtx, cancel := context.WithTimeout(context.WithoutCancel(c.Context()), 3*time.Second)
	defer cancel()
	auditErr := api.audit.Log(auditCtx, meta, action, "cloud", target, nil, detail)
	if err != nil {
		return common.NewResponse(c).Error(common.NewError(common.RETURN_FAILED, 502, err.Error()))
	}
	response := operationResult{Result: result}
	if auditErr != nil {
		response.Warnings = []string{"Cloud operation completed but its result audit could not be saved; the intent audit is retained"}
	}
	return common.NewResponse(c).SuccessWithData(response)
}
func (api *API) RepairMirror(c fiber.Ctx) error {
	var req struct {
		Key string `json:"key"`
	}
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if !assets.ValidKey(req.Key) || req.Key == assets.ProbeKey {
		return common.NewResponse(c).Error(common.NewValidationError("invalid managed object key"))
	}
	return api.operate(c, "cloud.mirror.repair", req.Key, req, func(ctx context.Context) (any, error) {
		if err := api.service.Storage.RepairMirror(ctx, req.Key); err != nil {
			return nil, err
		}
		return api.service.Inspect(ctx, req.Key)
	})
}
func (api *API) purge(c fiber.Ctx, provider string) error {
	var req infra.PurgeRequest
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	req, err := api.service.Validate(provider, req)
	if err != nil {
		return common.NewResponse(c).Error(common.NewValidationError(err.Error()))
	}
	return api.operate(c, "cloud."+provider+".purge."+req.Type, provider, req, func(ctx context.Context) (any, error) { return api.service.Purge(ctx, provider, req) })
}
func (api *API) EdgeOnePurge(c fiber.Ctx) error    { return api.purge(c, "edgeone") }
func (api *API) CloudflarePurge(c fiber.Ctx) error { return api.purge(c, "cloudflare") }
func (api *API) EdgeOnePurgeAll(c fiber.Ctx) error {
	return api.operate(c, "cloud.edgeone.purge.all", "configured EdgeOne zone", map[string]string{"type": "all"}, func(ctx context.Context) (any, error) { return api.service.PurgeAll(ctx) })
}

var jobIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,128}$`)

func (api *API) EdgeOnePurgeTasks(c fiber.Ctx) error {
	jobID := strings.TrimSpace(c.Query("job_id"))
	if jobID != "" && !jobIDPattern.MatchString(jobID) {
		return common.NewResponse(c).Error(common.NewValidationError("invalid job_id"))
	}
	page := adminutil.ParsePageQuery(c)
	result, err := api.service.Tasks(c.Context(), jobID, int64(page.PageNum-1)*20)
	if err != nil {
		return common.NewResponse(c).Error(common.NewError(common.RETURN_FAILED, 502, err.Error()))
	}
	return common.NewResponse(c).SuccessWithData(result)
}
