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

func TestListSitesUsesV2SyncEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/nav/sync/sites" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("lang"); got != "en" {
			t.Fatalf("lang = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1,"data":[{"id":"1","name":"Site","domain":"example.com","info":"Info","country":"","nsfw":"0","welfare":"0"}]}`))
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

func TestListGroupsUsesDedicatedSyncEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/nav/sync/site-groups" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1,"data":[{"id":"g1","name":"Group","sites":[{"id":"1","name":"Site"},{"site_id":"2"},{"id":3},"4"]}]}`))
	}))
	defer server.Close()

	client := NewHTTPNavClient(server.URL+"/api/v1", time.Second)
	groups, err := client.ListGroups(context.Background(), "zh-CN")
	if err != nil {
		t.Fatalf("ListGroups() error = %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("groups = %+v", groups)
	}
	want := []string{"1", "2", "3", "4"}
	if strings.Join(groups[0].Sites, ",") != strings.Join(want, ",") {
		t.Fatalf("group sites = %#v, want %#v", groups[0].Sites, want)
	}
}

func TestGetSiteDetailUsesV2Endpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/nav/sites/101/detail" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("lang"); got != "zh" {
			t.Fatalf("lang = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1,"data":{"site":{"name":"兽人控游戏索引","info":"简介","country":null,"nsfw":"no","welfare":"no"}}}`))
	}))
	defer server.Close()

	client := NewHTTPNavClient(server.URL+"/api/v1", time.Second)
	detail, err := client.GetSiteDetail(context.Background(), "101", "zh-CN")
	if err != nil {
		t.Fatalf("GetSiteDetail() error = %v", err)
	}
	if detail.Name != "兽人控游戏索引" || detail.Info != "简介" {
		t.Fatalf("detail = %+v", detail)
	}
}

func TestGetSiteHTTPIgnoresRemovedLegacyEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":0,"data":"链接不存在"}`))
	}))
	defer server.Close()

	client := NewHTTPNavClient(server.URL+"/api/v1", time.Second)
	if _, err := client.GetSiteHTTP(context.Background(), "example.com"); err != nil {
		t.Fatalf("GetSiteHTTP() error = %v", err)
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
