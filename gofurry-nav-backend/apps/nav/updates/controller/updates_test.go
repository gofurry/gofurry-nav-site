package controller

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/updates/models"
	"github.com/gofurry/gofurry-nav-backend/common"
)

func TestGetUpdatesReturnsV2EnvelopeWithoutLegacyURL(t *testing.T) {
	now := time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC)
	reader := &fakeUpdatesReader{response: models.UpdatesResponse{
		SchemaVersion: models.UpdatesSchemaVersion,
		GeneratedAt:   now,
		State:         models.UpdatesStateReady,
		Items: []models.UpdateNoticeItem{{
			ID:          1,
			Title:       "公告重构",
			Body:        "告别 CDN markdown",
			PublishedAt: now,
			CreateTime:  now,
			UpdateTime:  now,
		}},
	}}
	restore := setUpdatesReaderForTest(reader)
	t.Cleanup(restore)

	app := fiber.New()
	app.Get("/updates", UpdatesApi.GetUpdates)
	resp, err := app.Test(httptest.NewRequest("GET", "/updates", nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	var body struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if body.Code != common.RETURN_SUCCESS {
		t.Fatalf("unexpected code: %d", body.Code)
	}
	raw := string(body.Data)
	if !strings.Contains(raw, `"state":"ready"`) || !strings.Contains(raw, `"body":"告别 CDN markdown"`) {
		t.Fatalf("unexpected data: %s", raw)
	}
	if strings.Contains(raw, `"url"`) {
		t.Fatalf("legacy url leaked into updates response: %s", raw)
	}
	if reader.lastLang != "zh" {
		t.Fatalf("expected default zh lang, got %q", reader.lastLang)
	}
}

type fakeUpdatesReader struct {
	response models.UpdatesResponse
	lastLang string
}

func (reader *fakeUpdatesReader) GetUpdates(lang string) models.UpdatesResponse {
	reader.lastLang = lang
	return reader.response
}
