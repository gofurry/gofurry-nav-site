package reporter

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/config"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/security"
)

func TestReporterSignsRequest(t *testing.T) {
	var seen bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		nodeID := r.Header.Get("X-GoFurry-Node-ID")
		ts := r.Header.Get("X-GoFurry-Timestamp")
		sig := r.Header.Get("X-GoFurry-Signature")
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		if !security.Verify("secret", ts, nodeID, body, sig) {
			t.Fatal("invalid signature")
		}
		seen = true
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	r := New(config.CenterConfig{
		Endpoint: server.URL,
		Token:    "secret",
		Timeout:  config.Duration{Duration: time.Second},
		Retries:  1,
	}, "cn-business-a")
	if err := r.SendRaw(context.Background(), []byte(`{"node_id":"cn-business-a"}`)); err != nil {
		t.Fatal(err)
	}
	if !seen {
		t.Fatal("server did not receive request")
	}
}
