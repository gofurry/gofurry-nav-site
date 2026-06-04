package service

import (
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-backend/common"
)

func TestSearchSuggestionsUsesCacheBeforeProvider(t *testing.T) {
	cache := newMemorySuggestionCache()
	cache.Set(searchSuggestionCacheKey("bing", "兽人"), []string{"cached"})
	provider := &fakeSuggestionProvider{}
	svc := newSearchService(provider, cache, fixedSearchNow)

	response := svc.GetSearchSuggestions("bing", "兽人")
	if response.State != "ready" || response.CacheState != "hit" {
		t.Fatalf("unexpected response state/cache: %#v", response)
	}
	if len(response.Suggestions) != 1 || response.Suggestions[0] != "cached" {
		t.Fatalf("suggestions = %v", response.Suggestions)
	}
	if provider.calls != 0 {
		t.Fatalf("provider calls = %d", provider.calls)
	}
}

func TestSearchSuggestionsFetchesAndCachesSanitizedItems(t *testing.T) {
	cache := newMemorySuggestionCache()
	provider := &fakeSuggestionProvider{items: []string{"  furry  ", "furry", "", "兽人"}}
	svc := newSearchService(provider, cache, fixedSearchNow)

	response := svc.GetSearchSuggestions("google", "  test  ")
	if response.State != "ready" || response.CacheState != "miss" {
		t.Fatalf("unexpected response state/cache: %#v", response)
	}
	if provider.lastEngine != "google" || provider.lastQuery != "test" {
		t.Fatalf("provider got engine/query = %q/%q", provider.lastEngine, provider.lastQuery)
	}
	if got := strings.Join(response.Suggestions, ","); got != "furry,兽人" {
		t.Fatalf("suggestions = %v", response.Suggestions)
	}

	cached, ok := cache.Get(searchSuggestionCacheKey("google", "test"))
	if !ok || strings.Join(cached, ",") != "furry,兽人" {
		t.Fatalf("cache = %v, %v", cached, ok)
	}
}

func TestSearchSuggestionsRejectsUnsupportedEngine(t *testing.T) {
	svc := newSearchService(&fakeSuggestionProvider{}, newMemorySuggestionCache(), fixedSearchNow)
	response := svc.GetSearchSuggestions("duckduckgo", "furry")
	if response.State != "error" {
		t.Fatalf("state = %q", response.State)
	}
	if len(response.ReasonMessages) == 0 {
		t.Fatalf("expected reason_messages")
	}
}

func TestSearchSuggestionsNormalizesQueryLength(t *testing.T) {
	provider := &fakeSuggestionProvider{items: []string{"ok"}}
	svc := newSearchService(provider, newMemorySuggestionCache(), fixedSearchNow)
	query := "  " + strings.Repeat("兽", searchSuggestionMaxQueryLen+8) + "  "

	response := svc.GetSearchSuggestions("baidu", query)
	if len([]rune(response.Query)) != searchSuggestionMaxQueryLen {
		t.Fatalf("query length = %d", len([]rune(response.Query)))
	}
	if provider.lastQuery != response.Query {
		t.Fatalf("provider query = %q, response query = %q", provider.lastQuery, response.Query)
	}
}

func fixedSearchNow() time.Time {
	return time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)
}

type memorySuggestionCache struct {
	items map[string][]string
}

func newMemorySuggestionCache() *memorySuggestionCache {
	return &memorySuggestionCache{items: map[string][]string{}}
}

func (cache *memorySuggestionCache) Get(key string) ([]string, bool) {
	items, ok := cache.items[key]
	return copySuggestions(items), ok
}

func (cache *memorySuggestionCache) Set(key string, suggestions []string) {
	cache.items[key] = copySuggestions(suggestions)
}

type fakeSuggestionProvider struct {
	items      []string
	err        common.GFError
	calls      int
	lastEngine string
	lastQuery  string
}

func (provider *fakeSuggestionProvider) GetBaiduSuggestion(q string) ([]string, common.GFError) {
	return provider.record("baidu", q)
}

func (provider *fakeSuggestionProvider) GetBingSuggestion(q string) ([]string, common.GFError) {
	return provider.record("bing", q)
}

func (provider *fakeSuggestionProvider) GetGoogleSuggestion(q string) ([]string, common.GFError) {
	return provider.record("google", q)
}

func (provider *fakeSuggestionProvider) GetBiliBiliSuggestion(q string) ([]string, common.GFError) {
	return provider.record("bilibili", q)
}

func (provider *fakeSuggestionProvider) record(engine string, q string) ([]string, common.GFError) {
	provider.calls++
	provider.lastEngine = engine
	provider.lastQuery = q
	return provider.items, provider.err
}
