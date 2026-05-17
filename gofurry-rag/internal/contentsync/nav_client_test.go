package contentsync

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFetchMarkdownRejectsUnsupportedScheme(t *testing.T) {
	client := NewHTTPNavClient("https://nav.example.test/api", time.Second, nil)
	if _, err := client.FetchMarkdown(context.Background(), "file:///etc/passwd"); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported scheme error, got %v", err)
	}
}

func TestFetchMarkdownRejectsHostOutsideAllowlist(t *testing.T) {
	client := NewHTTPNavClient("https://nav.example.test/api", time.Second, nil)
	if _, err := client.FetchMarkdown(context.Background(), "https://example.com/changelog.md"); err == nil || !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("expected host allowlist error, got %v", err)
	}
}

func TestFetchMarkdownUsesAllowlistAndSizeLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("# changelog"))
	}))
	defer server.Close()

	client := NewHTTPNavClient(server.URL+"/api", time.Second, nil)
	body, err := client.FetchMarkdown(context.Background(), server.URL+"/changelog.md")
	if err != nil {
		t.Fatal(err)
	}
	if body != "# changelog" {
		t.Fatalf("body = %q", body)
	}

	largeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", maxMarkdownBytes+1)))
	}))
	defer largeServer.Close()

	client = NewHTTPNavClient(largeServer.URL+"/api", time.Second, nil)
	if _, err := client.FetchMarkdown(context.Background(), largeServer.URL+"/large.md"); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected size limit error, got %v", err)
	}
}

func TestFetchJSONCapsErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(strings.Repeat("x", maxErrorBodyBytes+1024)))
	}))
	defer server.Close()

	client := NewHTTPNavClient(server.URL, time.Second, nil)
	var target []NavSite
	err := client.fetchJSON(context.Background(), "/nav/page/site/list", nil, &target)
	if err == nil {
		t.Fatal("expected upstream error")
	}
	if strings.Count(err.Error(), "x") > maxErrorBodyBytes {
		t.Fatalf("error body was not capped: %d", strings.Count(err.Error(), "x"))
	}
}
