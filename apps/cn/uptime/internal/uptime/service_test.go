package uptime

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-uptime/internal/config"
	"go.uber.org/zap"
)

func TestServiceHealthAndUptimeMiddleware(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(target.Close)
	cfg := testConfig(t, target.URL)
	service, err := NewService(cfg, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = service.Shutdown() })

	for _, path := range []string{"/livez", "/readyz", "/uptime", "/uptime/api/status"} {
		response, testErr := service.App().Test(httptest.NewRequest(http.MethodGet, path, http.NoBody))
		if testErr != nil {
			t.Fatalf("GET %s: %v", path, testErr)
		}
		body, _ := io.ReadAll(response.Body)
		_ = response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("GET %s status = %d body=%s", path, response.StatusCode, body)
		}
		if path == "/uptime/api/status" && !strings.Contains(string(body), "test-target") {
			t.Fatalf("status payload does not contain endpoint id: %s", body)
		}
	}
	service.ready.Store(false)
	response, err := service.App().Test(httptest.NewRequest(http.MethodGet, "/readyz", http.NoBody))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("readyz status = %d, want 503", response.StatusCode)
	}
}

func TestServicePersistsHistoryAcrossRestart(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(target.Close)
	cfg := testConfig(t, target.URL)
	first, err := NewService(cfg, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	if err = first.Shutdown(); err != nil {
		t.Fatal(err)
	}
	second, err := NewService(cfg, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = second.Shutdown() })
	response, err := second.App().Test(httptest.NewRequest(http.MethodGet, "/uptime/api/status", http.NoBody))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK || !strings.Contains(string(body), "test-target") {
		t.Fatalf("persisted status = %d %s", response.StatusCode, body)
	}
}

func TestServiceGracefulShutdown(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(target.Close)
	cfg := testConfig(t, target.URL)
	cfg.Server.Port = "0"
	service, err := NewService(cfg, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- service.Serve(ctx) }()
	time.Sleep(100 * time.Millisecond)
	cancel()
	select {
	case err = <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("service did not shut down")
	}
}

func testConfig(t *testing.T, targetURL string) *config.Config {
	t.Helper()
	return &config.Config{
		Server: config.ServerConfig{Mode: "test", IPAddress: "127.0.0.1", Port: "0", Network: "tcp", AppVersion: "test"},
		Log:    config.LogConfig{LogMode: "dev", LogLevel: "error"},
		Storage: config.StorageConfig{
			Path: filepath.Join(t.TempDir(), "uptime.db"),
		},
		Uptime: config.UptimeConfig{
			SampleIntervalSeconds: 1,
			RetentionDays:         2,
			DaysToShow:            1,
			Timezone:              "Asia/Shanghai",
			UI: config.UIConfig{
				Path: "/uptime", Title: "Status", Description: "Services", Footer: "GoFurry",
				GreenThreshold: 0.999, YellowThreshold: 0.99,
			},
			Endpoints: []config.EndpointConfig{{
				ID: "test-target", Name: "Test", URL: targetURL, Method: http.MethodGet,
				ExpectedStatusCodes: []int{http.StatusOK}, IntervalSeconds: 1, TimeoutSeconds: 1,
			}},
		},
	}
}
