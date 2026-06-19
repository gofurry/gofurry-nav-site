package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/awesome-fiber-template/v3/medium/config"
	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/common"
)

const defaultNavBackendAdminHeader = "X-GoFurry-Admin-Token"

type navBackendEnvelope struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

func (api *navAPI) CollectStatus(c fiber.Ctx) error {
	return api.proxyNavBackend(c, "/api/v2/nav/collect/status", nil)
}

func (api *navAPI) CollectObservations(c fiber.Ctx) error {
	params := url.Values{}
	copyNavQueryParam(c, params, "site_id")
	copyNavQueryParam(c, params, "target")
	copyNavQueryParam(c, params, "protocol")
	copyNavQueryParam(c, params, "status")
	copyNavQueryParam(c, params, "limit")
	copyNavQueryParam(c, params, "offset")
	return api.proxyNavBackend(c, "/api/v2/nav/collect/observations", params)
}

func (api *navAPI) CollectSiteStatus(c fiber.Ctx) error {
	siteID := strings.TrimSpace(c.Params("site_id"))
	if siteID == "" {
		return common.NewResponse(c).Error(common.NewValidationError("site_id is required"))
	}
	return api.proxyNavBackend(c, "/api/v2/nav/collect/sites/"+url.PathEscape(siteID)+"/status", nil)
}

func (api *navAPI) CollectTargetStatus(c fiber.Ctx) error {
	siteID := strings.TrimSpace(c.Params("site_id"))
	target := strings.TrimSpace(c.Params("target"))
	if siteID == "" || target == "" {
		return common.NewResponse(c).Error(common.NewValidationError("site_id and target are required"))
	}
	return api.proxyNavBackend(c, "/api/v2/nav/collect/sites/"+url.PathEscape(siteID)+"/targets/"+url.PathEscape(target)+"/status", nil)
}

func (api *navAPI) proxyNavBackend(c fiber.Ctx, path string, params url.Values) error {
	cfg := env.GetServerConfig().ExternalServices.NavBackend
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return common.NewResponse(c).Error(common.NewServiceError("nav_backend base_url is not configured"))
	}
	token := strings.TrimSpace(cfg.AdminToken)
	if token == "" {
		return common.NewResponse(c).Error(common.NewServiceError("nav_backend admin_token is not configured"))
	}
	header := strings.TrimSpace(cfg.AdminTokenHeader)
	if header == "" {
		header = defaultNavBackendAdminHeader
	}
	timeout := cfg.TimeoutSeconds
	if timeout <= 0 {
		timeout = 10
	}

	target := baseURL + path
	if len(params) > 0 {
		target += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(c.Context(), http.MethodGet, target, nil)
	if err != nil {
		return common.NewResponse(c).Error(common.NewServiceError(err.Error()))
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set(header, token)

	resp, err := (&http.Client{Timeout: time.Duration(timeout) * time.Second}).Do(req)
	if err != nil {
		return common.NewResponse(c).Error(common.NewServiceError(fmt.Sprintf("request nav backend failed: %v", err)))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.NewResponse(c).Error(common.NewServiceError(fmt.Sprintf("read nav backend response failed: %v", err)))
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return common.NewResponse(c).Error(common.NewServiceError(fmt.Sprintf("nav backend returned status %d", resp.StatusCode)))
	}

	var envelope navBackendEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return common.NewResponse(c).Error(common.NewServiceError(fmt.Sprintf("decode nav backend response failed: %v", err)))
	}
	if envelope.Code != common.RETURN_SUCCESS {
		var message any
		if len(envelope.Data) > 0 {
			_ = json.Unmarshal(envelope.Data, &message)
		}
		return common.NewResponse(c).Error(common.NewServiceError(fmt.Sprintf("nav backend request failed: %v", message)))
	}
	var data any
	if len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, &data); err != nil {
			return common.NewResponse(c).Error(common.NewServiceError(fmt.Sprintf("decode nav backend data failed: %v", err)))
		}
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func copyNavQueryParam(c fiber.Ctx, params url.Values, name string) {
	value := strings.TrimSpace(c.Query(name, ""))
	if value != "" {
		params.Set(name, value)
	}
}
