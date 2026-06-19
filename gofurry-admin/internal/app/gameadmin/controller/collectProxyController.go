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

const defaultGameBackendAdminHeader = "X-GoFurry-Admin-Token"

type gameBackendEnvelope struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

func (api *gameAPI) CollectStatus(c fiber.Ctx) error {
	return api.proxyGameBackend(c, "/api/v2/game/collect/status", nil)
}

func (api *gameAPI) CollectRuns(c fiber.Ctx) error {
	params := url.Values{}
	copyQueryParam(c, params, "task_type")
	copyQueryParam(c, params, "status")
	copyQueryParam(c, params, "limit")
	copyQueryParam(c, params, "offset")
	return api.proxyGameBackend(c, "/api/v2/game/collect/runs", params)
}

func (api *gameAPI) CollectRun(c fiber.Ctx) error {
	runID := strings.TrimSpace(c.Params("run_id"))
	if runID == "" {
		return common.NewResponse(c).Error(common.NewValidationError("run_id is required"))
	}
	return api.proxyGameBackend(c, "/api/v2/game/collect/runs/"+url.PathEscape(runID), nil)
}

func (api *gameAPI) CollectTaskResults(c fiber.Ctx) error {
	params := url.Values{}
	copyQueryParam(c, params, "run_id")
	copyQueryParam(c, params, "task_type")
	copyQueryParam(c, params, "status")
	copyQueryParam(c, params, "game_id")
	copyQueryParam(c, params, "appid")
	copyQueryParam(c, params, "limit")
	copyQueryParam(c, params, "offset")
	return api.proxyGameBackend(c, "/api/v2/game/collect/task-results", params)
}

func (api *gameAPI) CollectGameStatus(c fiber.Ctx) error {
	gameID := strings.TrimSpace(c.Params("id"))
	if gameID == "" {
		return common.NewResponse(c).Error(common.NewValidationError("game id is required"))
	}
	params := url.Values{}
	copyQueryParam(c, params, "appid")
	return api.proxyGameBackend(c, "/api/v2/game/collect/games/"+url.PathEscape(gameID)+"/status", params)
}

func (api *gameAPI) proxyGameBackend(c fiber.Ctx, path string, params url.Values) error {
	cfg := env.GetServerConfig().ExternalServices.GameBackend
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return common.NewResponse(c).Error(common.NewServiceError("game_backend base_url is not configured"))
	}
	token := strings.TrimSpace(cfg.AdminToken)
	if token == "" {
		return common.NewResponse(c).Error(common.NewServiceError("game_backend admin_token is not configured"))
	}
	header := strings.TrimSpace(cfg.AdminTokenHeader)
	if header == "" {
		header = defaultGameBackendAdminHeader
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
		return common.NewResponse(c).Error(common.NewServiceError(fmt.Sprintf("request game backend failed: %v", err)))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.NewResponse(c).Error(common.NewServiceError(fmt.Sprintf("read game backend response failed: %v", err)))
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return common.NewResponse(c).Error(common.NewServiceError(fmt.Sprintf("game backend returned status %d", resp.StatusCode)))
	}

	var envelope gameBackendEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return common.NewResponse(c).Error(common.NewServiceError(fmt.Sprintf("decode game backend response failed: %v", err)))
	}
	if envelope.Code != common.RETURN_SUCCESS {
		var message any
		if len(envelope.Data) > 0 {
			_ = json.Unmarshal(envelope.Data, &message)
		}
		return common.NewResponse(c).Error(common.NewServiceError(fmt.Sprintf("game backend request failed: %v", message)))
	}
	var data any
	if len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, &data); err != nil {
			return common.NewResponse(c).Error(common.NewServiceError(fmt.Sprintf("decode game backend data failed: %v", err)))
		}
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func copyQueryParam(c fiber.Ctx, params url.Values, name string) {
	value := strings.TrimSpace(c.Query(name, ""))
	if value != "" {
		params.Set(name, value)
	}
}
