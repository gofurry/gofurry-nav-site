package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
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
	restore := setDetailReaderForTest(&fakeDetailReader{
		latest: detailmodels.TargetLatestResponse{
			State:         "ready",
			SiteID:        1,
			Target:        "example.com",
			Protocols:     map[string]readmodels.CollectorEnvelope{},
			GeneratedAt:   time.Date(2026, 5, 27, 13, 0, 0, 0, time.UTC),
			SchemaVersion: detailmodels.DetailSchemaVersion,
		},
	})
	t.Cleanup(restore)

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

func TestDetailRouteSuccessAndServiceErrorTable(t *testing.T) {
	now := time.Date(2026, 5, 27, 13, 0, 0, 0, time.UTC)
	cases := []struct {
		name         string
		route        string
		path         string
		handler      fiber.Handler
		successFake  *fakeDetailReader
		errorFake    *fakeDetailReader
		successNeed  string
		errorMessage string
	}{
		{
			name:    "detail",
			route:   "/sites/:siteId/detail",
			path:    "/sites/1/detail?lang=zh&payload_mode=full",
			handler: DetailApi.GetSiteDetail,
			successFake: &fakeDetailReader{detail: detailmodels.SiteDetailResponse{
				SelectedTarget: "example.com",
				GeneratedAt:    now,
				SchemaVersion:  detailmodels.DetailSchemaVersion,
			}},
			errorFake:    &fakeDetailReader{detailErr: common.NewServiceError("detail failed")},
			successNeed:  `"selected_target":"example.com"`,
			errorMessage: "detail failed",
		},
		{
			name:    "latest",
			route:   "/sites/:siteId/targets/:target/latest",
			path:    "/sites/1/targets/example.com/latest?payload_mode=full",
			handler: DetailApi.GetTargetLatest,
			successFake: &fakeDetailReader{latest: detailmodels.TargetLatestResponse{
				State:         "ready",
				SiteID:        1,
				Target:        "example.com",
				Protocols:     map[string]readmodels.CollectorEnvelope{},
				GeneratedAt:   now,
				SchemaVersion: detailmodels.DetailSchemaVersion,
			}},
			errorFake:    &fakeDetailReader{latestErr: common.NewServiceError("latest failed")},
			successNeed:  `"target":"example.com"`,
			errorMessage: "latest failed",
		},
		{
			name:    "observations",
			route:   "/sites/:siteId/targets/:target/observations",
			path:    "/sites/1/targets/example.com/observations?protocol=http&limit=10&payload_mode=full",
			handler: DetailApi.ListTargetObservations,
			successFake: &fakeDetailReader{observations: detailmodels.TargetObservationsResponse{
				State:         "ready",
				SiteID:        1,
				Target:        "example.com",
				Protocol:      readmodels.ProtocolHTTP,
				Limit:         10,
				Items:         []readmodels.CollectorEnvelope{},
				GeneratedAt:   now,
				SchemaVersion: detailmodels.DetailSchemaVersion,
			}},
			errorFake:    &fakeDetailReader{observeErr: common.NewServiceError("observations failed")},
			successNeed:  `"protocol":"http"`,
			errorMessage: "observations failed",
		},
		{
			name:    "trend",
			route:   "/sites/:siteId/targets/:target/trend",
			path:    "/sites/1/targets/example.com/trend",
			handler: DetailApi.GetTargetTrend,
			successFake: &fakeDetailReader{trend: detailmodels.TargetTrendResponse{
				State:         "ready",
				SiteID:        1,
				Target:        "example.com",
				GeneratedAt:   now,
				SchemaVersion: detailmodels.DetailSchemaVersion,
			}},
			errorFake:    &fakeDetailReader{trendErr: common.NewServiceError("trend failed")},
			successNeed:  `"target":"example.com"`,
			errorMessage: "trend failed",
		},
		{
			name:    "changes",
			route:   "/sites/:siteId/targets/:target/changes",
			path:    "/sites/1/targets/example.com/changes",
			handler: DetailApi.GetTargetChanges,
			successFake: &fakeDetailReader{changes: detailmodels.TargetChangesResponse{
				State:         "ready",
				SiteID:        1,
				Target:        "example.com",
				GeneratedAt:   now,
				SchemaVersion: detailmodels.DetailSchemaVersion,
			}},
			errorFake:    &fakeDetailReader{changesErr: common.NewServiceError("changes failed")},
			successNeed:  `"target":"example.com"`,
			errorMessage: "changes failed",
		},
		{
			name:    "light-probes",
			route:   "/sites/:siteId/targets/:target/light-probes",
			path:    "/sites/1/targets/example.com/light-probes?payload_mode=full",
			handler: DetailApi.GetTargetLightProbes,
			successFake: &fakeDetailReader{light: detailmodels.TargetLatestResponse{
				State:         "ready",
				SiteID:        1,
				Target:        "example.com",
				Protocols:     map[string]readmodels.CollectorEnvelope{},
				GeneratedAt:   now,
				SchemaVersion: detailmodels.DetailSchemaVersion,
			}},
			errorFake:    &fakeDetailReader{lightErr: common.NewServiceError("light failed")},
			successNeed:  `"target":"example.com"`,
			errorMessage: "light failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			restore := setDetailReaderForTest(tc.successFake)
			t.Cleanup(restore)
			app := fiber.New()
			app.Get(tc.route, tc.handler)
			resp, err := app.Test(httptest.NewRequest("GET", tc.path, nil))
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}
			defer resp.Body.Close()
			body := decodeResultData(t, resp)
			if body.Code != common.RETURN_SUCCESS || !strings.Contains(string(body.RawData), tc.successNeed) {
				t.Fatalf("response = %+v raw=%s", body, body.RawData)
			}
		})

		t.Run(tc.name+" service error", func(t *testing.T) {
			restore := setDetailReaderForTest(tc.errorFake)
			t.Cleanup(restore)
			app := fiber.New()
			app.Get(tc.route, tc.handler)
			resp, err := app.Test(httptest.NewRequest("GET", tc.path, nil))
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}
			defer resp.Body.Close()
			body := decodeResultData(t, resp)
			if body.Code != common.RETURN_FAILED || body.Data != tc.errorMessage {
				t.Fatalf("response = %+v", body)
			}
		})
	}
}

func TestDetailRoutesRejectBadSiteID(t *testing.T) {
	cases := []struct {
		name    string
		route   string
		path    string
		handler fiber.Handler
	}{
		{"detail", "/sites/:siteId/detail", "/sites/nope/detail", DetailApi.GetSiteDetail},
		{"latest", "/sites/:siteId/targets/:target/latest", "/sites/nope/targets/example.com/latest", DetailApi.GetTargetLatest},
		{"observations", "/sites/:siteId/targets/:target/observations", "/sites/nope/targets/example.com/observations?protocol=http", DetailApi.ListTargetObservations},
		{"trend", "/sites/:siteId/targets/:target/trend", "/sites/nope/targets/example.com/trend", DetailApi.GetTargetTrend},
		{"changes", "/sites/:siteId/targets/:target/changes", "/sites/nope/targets/example.com/changes", DetailApi.GetTargetChanges},
		{"light-probes", "/sites/:siteId/targets/:target/light-probes", "/sites/nope/targets/example.com/light-probes", DetailApi.GetTargetLightProbes},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Get(tc.route, tc.handler)
			resp, err := app.Test(httptest.NewRequest("GET", tc.path, nil))
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}
			defer resp.Body.Close()
			body := decodeResultData(t, resp)
			if body.Code != common.RETURN_FAILED || body.Data != "siteId 参数非法" {
				t.Fatalf("response = %+v", body)
			}
		})
	}
}

func TestTargetParamIsURLDecoded(t *testing.T) {
	fake := &fakeDetailReader{latest: detailmodels.TargetLatestResponse{
		State:         "ready",
		SiteID:        1,
		Target:        "example.com:443",
		Protocols:     map[string]readmodels.CollectorEnvelope{},
		SchemaVersion: detailmodels.DetailSchemaVersion,
	}}
	restore := setDetailReaderForTest(fake)
	t.Cleanup(restore)

	app := fiber.New()
	app.Get("/sites/:siteId/targets/:target/latest", DetailApi.GetTargetLatest)
	resp, err := app.Test(httptest.NewRequest("GET", "/sites/1/targets/example.com%3A443/latest?payload_mode=full", nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	body := decodeResultData(t, resp)
	if body.Code != common.RETURN_SUCCESS {
		t.Fatalf("response = %+v", body)
	}
	if fake.lastTarget != "example.com:443" || fake.lastPayloadMode != "full" {
		t.Fatalf("decoded target/payload mode = %q/%q", fake.lastTarget, fake.lastPayloadMode)
	}
}

func TestCurrentDetailServiceConcurrentWithInjectedReader(t *testing.T) {
	fake := &fakeDetailReader{}
	restore := setDetailReaderForTest(fake)
	t.Cleanup(restore)

	var wg sync.WaitGroup
	mismatches := make(chan struct{}, 64)
	for range 64 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if currentDetailService() != fake {
				mismatches <- struct{}{}
			}
		}()
	}
	wg.Wait()
	close(mismatches)
	if len(mismatches) > 0 {
		t.Fatalf("currentDetailService returned mismatched reader %d times", len(mismatches))
	}
}

type fakeDetailReader struct {
	detail          detailmodels.SiteDetailResponse
	detailErr       common.GFError
	latest          detailmodels.TargetLatestResponse
	latestErr       common.GFError
	observations    detailmodels.TargetObservationsResponse
	observeErr      common.GFError
	trend           detailmodels.TargetTrendResponse
	trendErr        common.GFError
	changes         detailmodels.TargetChangesResponse
	changesErr      common.GFError
	light           detailmodels.TargetLatestResponse
	lightErr        common.GFError
	lastTarget      string
	lastPayloadMode string
}

func (f *fakeDetailReader) GetSiteDetail(siteID int64, lang string, target string, payloadMode string) (detailmodels.SiteDetailResponse, common.GFError) {
	f.lastTarget = target
	f.lastPayloadMode = payloadMode
	return f.detail, f.detailErr
}

func (f *fakeDetailReader) GetTargetLatest(siteID int64, target string, payloadMode string) (detailmodels.TargetLatestResponse, common.GFError) {
	f.lastTarget = target
	f.lastPayloadMode = payloadMode
	return f.latest, f.latestErr
}

func (f *fakeDetailReader) ListTargetObservations(siteID int64, target string, protocol string, limit int, payloadMode string) (detailmodels.TargetObservationsResponse, common.GFError) {
	f.lastTarget = target
	f.lastPayloadMode = payloadMode
	return f.observations, f.observeErr
}

func (f *fakeDetailReader) GetTargetTrend(siteID int64, target string) (detailmodels.TargetTrendResponse, common.GFError) {
	f.lastTarget = target
	return f.trend, f.trendErr
}

func (f *fakeDetailReader) GetTargetChanges(siteID int64, target string) (detailmodels.TargetChangesResponse, common.GFError) {
	f.lastTarget = target
	return f.changes, f.changesErr
}

func (f *fakeDetailReader) GetTargetLightProbes(siteID int64, target string, payloadMode string) (detailmodels.TargetLatestResponse, common.GFError) {
	f.lastTarget = target
	f.lastPayloadMode = payloadMode
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
