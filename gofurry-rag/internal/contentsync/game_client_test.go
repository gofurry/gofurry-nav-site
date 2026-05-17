package contentsync

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGameFetchJSONCapsErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(strings.Repeat("x", maxErrorBodyBytes+1024)))
	}))
	defer server.Close()

	client := NewHTTPGameClient(server.URL, time.Second)
	var target []GameSummary
	err := client.fetchJSON(context.Background(), "/game/sync/list", nil, &target)
	if err == nil {
		t.Fatal("expected upstream error")
	}
	if strings.Count(err.Error(), "x") > maxErrorBodyBytes {
		t.Fatalf("error body was not capped: %d", strings.Count(err.Error(), "x"))
	}
}
