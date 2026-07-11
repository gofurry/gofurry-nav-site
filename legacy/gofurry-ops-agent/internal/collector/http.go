package collector

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/config"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/model"
)

func collectHTTP(ctx context.Context, cfg config.HTTPCheckConfig) model.HTTPCheckResult {
	start := time.Now()
	result := model.HTTPCheckResult{Name: cfg.Name, URL: cfg.URL}
	checkCtx, cancel := context.WithTimeout(ctx, cfg.Timeout.Duration)
	defer cancel()

	req, err := http.NewRequestWithContext(checkCtx, cfg.Method, cfg.URL, nil)
	if err != nil {
		result.Status = "down"
		result.ErrorMessage = err.Error()
		return result
	}
	client := &http.Client{Timeout: cfg.Timeout.Duration}
	resp, err := client.Do(req)
	result.LatencyMS = time.Since(start).Milliseconds()
	if err != nil {
		result.Status = statusFromError(err)
		result.ErrorMessage = err.Error()
		return result
	}
	defer resp.Body.Close()
	result.StatusCode = resp.StatusCode
	result.Status = "ok"
	if cfg.ExpectStatus > 0 && resp.StatusCode != cfg.ExpectStatus {
		result.Status = "down"
		result.ErrorMessage = "unexpected status"
	}
	if cfg.ExpectBody != "" {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 128*1024))
		if readErr != nil {
			result.Status = "down"
			result.ErrorMessage = readErr.Error()
		} else if !strings.Contains(string(body), cfg.ExpectBody) {
			result.Status = "down"
			result.ErrorMessage = "expected body keyword not found"
		}
	}
	return result
}
