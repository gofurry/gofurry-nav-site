package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/search/models"
	"github.com/gofurry/gofurry-nav-backend/common"
)

func TestGetSearchSuggestionsReturnsV2Envelope(t *testing.T) {
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)
	reader := &fakeSuggestionsReader{response: models.SearchSuggestionsResponse{
		SchemaVersion: models.SearchSuggestionsSchemaVersion,
		GeneratedAt:   now,
		State:         models.SearchSuggestionsStateReady,
		Engine:        "bing",
		Query:         "furry",
		Suggestions:   []string{"furry game"},
		CacheState:    models.SearchSuggestionsCacheMiss,
	}}
	restoreReader := setSuggestionsReaderForTest(reader)
	t.Cleanup(restoreReader)
	restoreLimiter := setSuggestionsLimiterForTest(&fakeSuggestionsLimiter{allowed: true})
	t.Cleanup(restoreLimiter)

	app := fiber.New()
	app.Get("/search/suggestions", SearchApi.GetSearchSuggestions)
	resp, err := app.Test(httptest.NewRequest("GET", "/search/suggestions?engine=bing&q=furry", nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	var body struct {
		Code int                              `json:"code"`
		Data models.SearchSuggestionsResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if body.Code != common.RETURN_SUCCESS {
		t.Fatalf("unexpected code: %d", body.Code)
	}
	if body.Data.State != models.SearchSuggestionsStateReady || len(body.Data.Suggestions) != 1 {
		t.Fatalf("unexpected data: %#v", body.Data)
	}
	if reader.lastEngine != "bing" || reader.lastQuery != "furry" {
		t.Fatalf("reader got engine/query = %q/%q", reader.lastEngine, reader.lastQuery)
	}
}

func TestGetSearchSuggestionsRateLimited(t *testing.T) {
	restoreReader := setSuggestionsReaderForTest(&fakeSuggestionsReader{})
	t.Cleanup(restoreReader)
	restoreLimiter := setSuggestionsLimiterForTest(&fakeSuggestionsLimiter{allowed: false, retryAfter: 1800})
	t.Cleanup(restoreLimiter)

	app := fiber.New()
	app.Get("/search/suggestions", SearchApi.GetSearchSuggestions)
	resp, err := app.Test(httptest.NewRequest("GET", "/search/suggestions?engine=bing&q=furry", nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Retry-After"); got != "1800" {
		t.Fatalf("Retry-After = %q", got)
	}
}

type fakeSuggestionsReader struct {
	response   models.SearchSuggestionsResponse
	lastEngine string
	lastQuery  string
}

func (reader *fakeSuggestionsReader) GetSearchSuggestions(engine string, query string) models.SearchSuggestionsResponse {
	reader.lastEngine = engine
	reader.lastQuery = query
	return reader.response
}

type fakeSuggestionsLimiter struct {
	allowed    bool
	retryAfter int64
}

func (limiter *fakeSuggestionsLimiter) Allow(string) (bool, int64) {
	return limiter.allowed, limiter.retryAfter
}
