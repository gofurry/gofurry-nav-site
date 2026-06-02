package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSuggestionQueryIsTrimmedAndLimited(t *testing.T) {
	query := normalizeSuggestionQuery("  " + strings.Repeat("兽", searchSuggestMaxQueryLen+10) + "  ")
	if len([]rune(query)) != searchSuggestMaxQueryLen {
		t.Fatalf("query rune length = %d", len([]rune(query)))
	}
}

func TestBuildSuggestionURLEncodesQuery(t *testing.T) {
	reqURL, err := buildSuggestionURL("https://example.com/search", map[string]string{"q": "兽 人 空 格"})
	if err != nil {
		t.Fatalf("buildSuggestionURL() error = %v", err)
	}
	if !strings.Contains(reqURL, "q=%E5%85%BD+%E4%BA%BA+%E7%A9%BA+%E6%A0%BC") {
		t.Fatalf("query was not encoded: %s", reqURL)
	}
}

func TestFetchSuggestionBodyLimitsResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", searchSuggestMaxBodyBytes*2)))
	}))
	defer server.Close()

	body, err := fetchSuggestionBody(server.URL, nil)
	if err != nil {
		t.Fatalf("fetchSuggestionBody() error = %v", err)
	}
	if len(body) != searchSuggestMaxBodyBytes {
		t.Fatalf("body length = %d", len(body))
	}
}

func TestSearchSuggestionDegradesOnUpstreamFailure(t *testing.T) {
	previous := baiduSuggestEndpoint
	baiduSuggestEndpoint = "://bad-url"
	t.Cleanup(func() { baiduSuggestEndpoint = previous })

	items, err := GetNavPageService().GetBaiduSuggestion("test")
	if err != nil {
		t.Fatalf("GetBaiduSuggestion() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("items = %v", items)
	}
}

func TestBiliSuggestionParsesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("term"); got != "兽 人" {
			t.Fatalf("term = %q", got)
		}
		_, _ = w.Write([]byte(`{"result":{"tag":[{"value":"兽人游戏"}]}}`))
	}))
	defer server.Close()

	previous := biliSuggestEndpoint
	biliSuggestEndpoint = server.URL
	t.Cleanup(func() { biliSuggestEndpoint = previous })

	items, err := GetNavPageService().GetBiliBiliSuggestion("兽 人")
	if err != nil {
		t.Fatalf("GetBiliBiliSuggestion() error = %v", err)
	}
	if len(items) != 1 || items[0] != "兽人游戏" {
		t.Fatalf("items = %v", items)
	}
}
