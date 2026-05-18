package collector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/config"
)

func TestCollectHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("gofurry ok"))
	}))
	defer server.Close()

	result := collectHTTP(context.Background(), config.HTTPCheckConfig{
		Name:         "home",
		URL:          server.URL,
		Method:       "GET",
		Timeout:      config.Duration{Duration: time.Second},
		ExpectStatus: http.StatusOK,
		ExpectBody:   "gofurry",
	})
	if result.Status != "ok" {
		t.Fatalf("expected ok, got %#v", result)
	}
	if result.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", result.StatusCode)
	}
}
