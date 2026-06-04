package controller

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/siteindex/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	cm "github.com/gofurry/gofurry-nav-backend/common/models"
)

func TestGetSiteIndexReturnsEnvelope(t *testing.T) {
	now := time.Date(2026, 6, 4, 16, 0, 0, 0, time.UTC)
	restore := setSiteIndexReaderForTest(&fakeSiteIndexReader{response: models.SiteIndexResponse{
		SchemaVersion: models.SiteIndexSchemaVersion,
		GeneratedAt:   now,
		State:         models.SiteIndexStateReady,
		Items:         []models.SiteIndexItem{{ID: 1, Domains: []string{"example.com"}, UpdatedAt: cm.LocalTime(now)}},
	}})
	t.Cleanup(restore)

	app := fiber.New()
	app.Get("/sites/index", SiteIndexApi.GetSiteIndex)
	resp, err := app.Test(httptest.NewRequest("GET", "/sites/index", nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	var body struct {
		Code int                      `json:"code"`
		Data models.SiteIndexResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if body.Code != common.RETURN_SUCCESS || body.Data.State != models.SiteIndexStateReady {
		t.Fatalf("body = %#v", body)
	}
}

type fakeSiteIndexReader struct {
	response models.SiteIndexResponse
}

func (reader *fakeSiteIndexReader) GetSiteIndex() models.SiteIndexResponse {
	return reader.response
}
