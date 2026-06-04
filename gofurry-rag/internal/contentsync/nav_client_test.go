package contentsync

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFetchJSONCapsErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(strings.Repeat("x", maxErrorBodyBytes+1024)))
	}))
	defer server.Close()

	client := NewHTTPNavClient(server.URL, time.Second)
	var target []NavSite
	err := client.fetchJSON(context.Background(), "/nav/page/site/list", nil, &target)
	if err == nil {
		t.Fatal("expected upstream error")
	}
	if strings.Count(err.Error(), "x") > maxErrorBodyBytes {
		t.Fatalf("error body was not capped: %d", strings.Count(err.Error(), "x"))
	}
}

func TestListSitesUsesV2HomeEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/nav/home" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("lang"); got != "en" {
			t.Fatalf("lang = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1,"data":{"sites":[{"id":"1","name":"Site","domain":"example.com","info":"Info","country":"","nsfw":"0","welfare":"0"}],"groups":[]}}`))
	}))
	defer server.Close()

	client := NewHTTPNavClient(server.URL+"/api/v1", time.Second)
	sites, err := client.ListSites(context.Background(), "en-US")
	if err != nil {
		t.Fatalf("ListSites() error = %v", err)
	}
	if len(sites) != 1 || sites[0].ID != "1" {
		t.Fatalf("sites = %+v", sites)
	}
}

func TestNormalizeNavLocale(t *testing.T) {
	if got := normalizeNavLocale("en-US"); got != "en" {
		t.Fatalf("normalize en-US = %q", got)
	}
	if got := normalizeNavLocale("zh-CN"); got != "zh" {
		t.Fatalf("normalize zh-CN = %q", got)
	}
}
