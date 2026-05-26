package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	detailmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/detail/models"
	readmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/readmodel/models"
	"github.com/gofurry/gofurry-nav-backend/common"
)

func TestGetSiteDetailRejectsBadSiteID(t *testing.T) {
	app := fiber.New()
	app.Get("/sites/:siteId/detail", DetailApi.GetSiteDetail)

	req := httptest.NewRequest("GET", "/sites/nope/detail", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	body := decodeResultData(t, resp)
	if body.Code != common.RETURN_FAILED || body.Data != "siteId 参数非法" {
		t.Fatalf("response = %+v", body)
	}
}

func TestListTargetObservationsRejectsBadLimit(t *testing.T) {
	app := fiber.New()
	app.Get("/sites/:siteId/targets/:target/observations", DetailApi.ListTargetObservations)

	req := httptest.NewRequest("GET", "/sites/1/targets/example.com/observations?protocol=http&limit=oops", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	body := decodeResultData(t, resp)
	if body.Code != common.RETURN_FAILED || body.Data != "limit 参数非法" {
		t.Fatalf("response = %+v", body)
	}
}

func TestGetTargetLatestSuccess(t *testing.T) {
	previous := detailSvc
	detailSvc = fakeDetailReader{
		latest: detailmodels.TargetLatestResponse{
			State:         "ready",
			SiteID:        1,
			Target:        "example.com",
			Protocols:     map[string]readmodels.CollectorEnvelope{},
			GeneratedAt:   time.Date(2026, 5, 27, 13, 0, 0, 0, time.UTC),
			SchemaVersion: detailmodels.DetailSchemaVersion,
		},
	}
	t.Cleanup(func() {
		detailSvc = previous
	})

	app := fiber.New()
	app.Get("/sites/:siteId/targets/:target/latest", DetailApi.GetTargetLatest)

	req := httptest.NewRequest("GET", "/sites/1/targets/example.com/latest", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	body := decodeResultData(t, resp)
	if body.Code != common.RETURN_SUCCESS {
		t.Fatalf("response = %+v", body)
	}
	if !strings.Contains(string(body.RawData), "\"target\":\"example.com\"") {
		t.Fatalf("raw data = %s", body.RawData)
	}
}

type fakeDetailReader struct {
	detail       detailmodels.SiteDetailResponse
	detailErr    common.GFError
	latest       detailmodels.TargetLatestResponse
	latestErr    common.GFError
	observations detailmodels.TargetObservationsResponse
	observeErr   common.GFError
	trend        detailmodels.TargetTrendResponse
	trendErr     common.GFError
	changes      detailmodels.TargetChangesResponse
	changesErr   common.GFError
	light        detailmodels.TargetLatestResponse
	lightErr     common.GFError
}

func (f fakeDetailReader) GetSiteDetail(siteID int64, lang string, target string) (detailmodels.SiteDetailResponse, common.GFError) {
	return f.detail, f.detailErr
}

func (f fakeDetailReader) GetTargetLatest(siteID int64, target string) (detailmodels.TargetLatestResponse, common.GFError) {
	return f.latest, f.latestErr
}

func (f fakeDetailReader) ListTargetObservations(siteID int64, target string, protocol string, limit int) (detailmodels.TargetObservationsResponse, common.GFError) {
	return f.observations, f.observeErr
}

func (f fakeDetailReader) GetTargetTrend(siteID int64, target string) (detailmodels.TargetTrendResponse, common.GFError) {
	return f.trend, f.trendErr
}

func (f fakeDetailReader) GetTargetChanges(siteID int64, target string) (detailmodels.TargetChangesResponse, common.GFError) {
	return f.changes, f.changesErr
}

func (f fakeDetailReader) GetTargetLightProbes(siteID int64, target string) (detailmodels.TargetLatestResponse, common.GFError) {
	return f.light, f.lightErr
}

type resultDataBody struct {
	Code    int             `json:"code"`
	Data    interface{}     `json:"data"`
	RawData json.RawMessage `json:"-"`
}

func decodeResultData(t *testing.T, resp *http.Response) resultDataBody {
	t.Helper()
	var decoded struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	body := resultDataBody{Code: decoded.Code, RawData: decoded.Data}
	var stringData string
	if err := json.Unmarshal(decoded.Data, &stringData); err == nil {
		body.Data = stringData
	}
	return body
}
