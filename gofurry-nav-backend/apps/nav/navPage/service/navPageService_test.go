package service

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	commonmodels "github.com/gofurry/gofurry-nav-backend/common/models"
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

func TestDuckDuckGoSuggestionParsesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "furry" {
			t.Fatalf("q = %q", got)
		}
		_, _ = w.Write([]byte(`[{"phrase":"furry"},{"phrase":"furry suit"},{"phrase":""}]`))
	}))
	defer server.Close()

	previous := duckSuggestEndpoint
	duckSuggestEndpoint = server.URL
	t.Cleanup(func() { duckSuggestEndpoint = previous })

	items, err := GetNavPageService().GetDuckDuckGoSuggestion("furry")
	if err != nil {
		t.Fatalf("GetDuckDuckGoSuggestion() error = %v", err)
	}
	if len(items) != 2 || items[0] != "furry" || items[1] != "furry suit" {
		t.Fatalf("items = %v", items)
	}
}

func TestConvertGroupRecordsOrdersMappingsDeterministically(t *testing.T) {
	svc := &navPageService{}
	groups := []models.GfnSiteGroup{
		{ID: 10, Name: "论坛社区", NameEn: "Forums", Info: "论坛", InfoEn: "Forums", Priority: 1},
	}
	mappings := []models.GfnSiteGroupMap{
		{ID: 9, GroupID: 10, SiteID: 300},
		{ID: 3, GroupID: 10, SiteID: 100},
		{ID: 5, GroupID: 10, SiteID: 200},
	}

	result := svc.convertGroupRecords(groups, mappings, "zh")
	if len(result) != 1 {
		t.Fatalf("group count = %d", len(result))
	}

	want := []string{"100", "200", "300"}
	if !reflect.DeepEqual(result[0].Sites, want) {
		t.Fatalf("sites order = %v, want %v", result[0].Sites, want)
	}
}

func TestConvertRecordsIncludesUpdateTime(t *testing.T) {
	svc := &navPageService{}
	records := []models.GfnSite{
		{
			ID:         7,
			Name:       "兽时社区",
			NameEn:     "Shoutime",
			Domain:     `{"domain":["www.shoutime.net"]}`,
			Info:       "中文站点",
			InfoEn:     "Forum site",
			Nsfw:       "0",
			Welfare:    "0",
			UpdateTime: models2LocalTime("2026-06-05 12:34:56"),
		},
	}

	result := svc.convertRecords(records, "zh")
	if len(result) != 1 {
		t.Fatalf("record count = %d", len(result))
	}
	if result[0].UpdateTime != "2026-06-05 12:34:56" {
		t.Fatalf("update_time = %q", result[0].UpdateTime)
	}
}

func models2LocalTime(value string) commonmodels.LocalTime {
	return commonmodels.LocalTime(parseTestTime(value))
}

func parseTestTime(value string) time.Time {
	tm, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
	if err != nil {
		panic(err)
	}
	return tm
}
