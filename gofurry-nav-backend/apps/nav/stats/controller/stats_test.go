package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/stats/models"
	"github.com/gofurry/gofurry-nav-backend/common"
)

func TestTouchPageViewRequiresPage(t *testing.T) {
	app := fiber.New()
	app.Post("/stats/page-view", StatsApi.TouchPageView)

	resp, err := app.Test(httptest.NewRequest(http.MethodPost, "/stats/page-view", nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	var body struct {
		Code int `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if body.Code != common.RETURN_FAILED {
		t.Fatalf("body = %#v", body)
	}
}

func TestTouchPageViewReturnsTrackerEnvelope(t *testing.T) {
	now := time.Date(2026, 6, 4, 17, 0, 0, 0, time.UTC)
	restore := setPageViewTrackerForTest(&fakePageViewTracker{response: models.PageViewResponse{
		SchemaVersion: models.PageViewSchemaVersion,
		GeneratedAt:   now,
		State:         models.PageViewStateReady,
		Page:          "nav_home",
		ViewCount:     12,
	}})
	t.Cleanup(restore)

	app := fiber.New()
	app.Post("/stats/page-view", StatsApi.TouchPageView)
	resp, err := app.Test(httptest.NewRequest(http.MethodPost, "/stats/page-view?page=nav_home", nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	var body struct {
		Code int                     `json:"code"`
		Data models.PageViewResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if body.Code != common.RETURN_SUCCESS || body.Data.Page != "nav_home" || body.Data.ViewCount != 12 {
		t.Fatalf("body = %#v", body)
	}
}

type fakePageViewTracker struct {
	response models.PageViewResponse
}

func (tracker *fakePageViewTracker) TouchPageView(page string, clientIP string) models.PageViewResponse {
	return tracker.response
}
