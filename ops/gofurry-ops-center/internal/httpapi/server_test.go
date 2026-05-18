package httpapi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/config"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/model"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/repository"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/security"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/service"
)

type fakeStore struct {
	ingested      bool
	metricsSince  time.Time
	metricsBucket time.Duration
	nodeMetricsID string
}

func (f *fakeStore) Ingest(context.Context, model.AgentPayload, int) error {
	f.ingested = true
	return nil
}
func (f *fakeStore) UpsertAlert(context.Context, repository.AlertInput) error { return nil }
func (f *fakeStore) ResolveAlert(context.Context, string) error               { return nil }
func (f *fakeStore) ListNodes(context.Context) ([]model.Node, error)          { return nil, nil }
func (f *fakeStore) GetNode(context.Context, string) (model.Node, error)      { return model.Node{}, nil }
func (f *fakeStore) MarkNodeStatus(context.Context, string, string) error     { return nil }
func (f *fakeStore) ListServiceStatuses(context.Context) ([]model.ServiceStatus, error) {
	return nil, nil
}
func (f *fakeStore) ListAlerts(context.Context, bool) ([]model.AlertState, error)  { return nil, nil }
func (f *fakeStore) UpsertPeerSummary(context.Context, model.PeerSummary) error    { return nil }
func (f *fakeStore) LatestPeerSummary(context.Context) (*model.PeerSummary, error) { return nil, nil }
func (f *fakeStore) CreateSyncRun(context.Context, model.SyncEventRequest) (model.SyncRun, error) {
	return model.SyncRun{}, nil
}
func (f *fakeStore) LatestSyncRun(context.Context) (*model.SyncRun, error)      { return nil, nil }
func (f *fakeStore) ListSyncRuns(context.Context, int) ([]model.SyncRun, error) { return nil, nil }
func (f *fakeStore) CreateDeployEvent(context.Context, model.DeployEventRequest) (model.DeployEvent, error) {
	return model.DeployEvent{}, nil
}
func (f *fakeStore) ListDeployEvents(context.Context, int) ([]model.DeployEvent, error) {
	return nil, nil
}
func (f *fakeStore) OverviewMetrics(_ context.Context, since time.Time, bucket time.Duration) (model.OverviewMetrics, error) {
	f.metricsSince = since
	f.metricsBucket = bucket
	return model.OverviewMetrics{CPUTrend: []model.MetricPoint{}}, nil
}
func (f *fakeStore) NodeMetrics(_ context.Context, nodeID string, since time.Time, bucket time.Duration) (model.NodeMetrics, error) {
	f.nodeMetricsID = nodeID
	f.metricsSince = since
	f.metricsBucket = bucket
	return model.NodeMetrics{}, nil
}

func TestAgentIngestRequiresValidSignature(t *testing.T) {
	store := &fakeStore{}
	cfg := config.Config{
		CenterID: "ops",
		Region:   "cn",
		Server:   config.ServerConfig{Host: "127.0.0.1", Port: 8080},
		Security: config.SecurityConfig{
			DashboardPasscode: "pass",
			SessionSecret:     "session",
			AgentTokens:       []config.AgentToken{{NodeID: "cn-business-a", Token: "agent-token"}},
			SignatureWindow:   config.Duration{Duration: 5 * time.Minute},
		},
	}
	app := New(cfg, service.New(cfg, store))
	ts := time.Now().UTC().Format(time.RFC3339)
	body := []byte(fmt.Sprintf(`{"node_id":"cn-business-a","region":"cn","timestamp":"%s","agent_version":"test"}`, ts))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/ingest", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer agent-token")
	req.Header.Set("X-GoFurry-Node-ID", "cn-business-a")
	req.Header.Set("X-GoFurry-Timestamp", ts)
	req.Header.Set("X-GoFurry-Signature", security.Sign("agent-token", ts, "cn-business-a", body))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusAccepted {
		t.Fatalf("expected accepted, got %d", resp.StatusCode)
	}
	if !store.ingested {
		t.Fatal("expected payload ingested")
	}
}

func TestAgentIngestRejectsOversizedBody(t *testing.T) {
	cfg := config.Config{
		CenterID: "ops",
		Region:   "cn",
		Security: config.SecurityConfig{
			DashboardPasscode: "pass",
			SessionSecret:     "session",
			AgentTokens:       []config.AgentToken{{NodeID: "node", Token: "token"}},
		},
	}
	app := New(cfg, service.New(cfg, &fakeStore{}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/ingest", bytes.NewReader(bytes.Repeat([]byte("x"), centerBodyLimit+1)))
	_, err := app.Test(req)
	if err == nil {
		t.Fatal("expected body limit rejection")
	}
	if !strings.Contains(err.Error(), "body size exceeds") {
		t.Fatalf("expected body size error, got %v", err)
	}
}

func TestLoginRateLimitCountsFailuresOnly(t *testing.T) {
	cfg := config.Config{
		CenterID: "ops",
		Region:   "cn",
		Security: config.SecurityConfig{
			DashboardPasscode: "correct-passcode",
			SessionSecret:     "session",
		},
	}
	app := New(cfg, service.New(cfg, &fakeStore{}))
	for i := 0; i < 5; i++ {
		resp, err := app.Test(httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", strings.NewReader(`{"passcode":"wrong"}`)))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Fatalf("expected unauthorized attempt %d, got %d", i+1, resp.StatusCode)
		}
	}
	resp, err := app.Test(httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", strings.NewReader(`{"passcode":"wrong"}`)))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("expected rate limit, got %d", resp.StatusCode)
	}

	successApp := New(cfg, service.New(cfg, &fakeStore{}))
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", strings.NewReader(`{"passcode":"correct-passcode"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := successApp.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("expected successful login attempt %d, got %d", i+1, resp.StatusCode)
		}
	}
}

func TestEmbeddedDashboardServed(t *testing.T) {
	cfg := config.Config{
		CenterID: "ops",
		Region:   "cn",
		Security: config.SecurityConfig{
			DashboardPasscode: "pass",
			SessionSecret:     "session",
			AgentTokens:       []config.AgentToken{{NodeID: "node", Token: "token"}},
		},
	}
	app := New(cfg, service.New(cfg, &fakeStore{}))
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/admin", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected ok, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "GoFurry Ops Center") {
		t.Fatalf("dashboard shell was not served: %q", string(body))
	}
}

func TestDashboardMetricsRequiresAdmin(t *testing.T) {
	cfg := config.Config{
		CenterID: "ops",
		Region:   "cn",
		Security: config.SecurityConfig{
			DashboardPasscode: "pass",
			SessionSecret:     "session",
		},
	}
	app := New(cfg, service.New(cfg, &fakeStore{}))
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/metrics/overview", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", resp.StatusCode)
	}
}

func TestDashboardMetricsRangeFallback(t *testing.T) {
	store := &fakeStore{}
	cfg := config.Config{
		CenterID: "ops",
		Region:   "cn",
		Security: config.SecurityConfig{
			DashboardPasscode: "pass",
			SessionSecret:     "session",
			CookieName:        "ops_session",
		},
	}
	app := New(cfg, service.New(cfg, store))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/metrics/overview?range=bad", nil)
	req.AddCookie(sessionCookie(t, "ops_session", "session"))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected ok, got %d", resp.StatusCode)
	}
	if store.metricsBucket != time.Minute {
		t.Fatalf("expected fallback bucket 1m, got %s", store.metricsBucket)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"range":"1h"`) {
		t.Fatalf("expected fallback range in body, got %s", string(body))
	}
}

func TestNodeMetricsRoute(t *testing.T) {
	store := &fakeStore{}
	cfg := config.Config{
		CenterID: "ops",
		Region:   "cn",
		Security: config.SecurityConfig{
			DashboardPasscode: "pass",
			SessionSecret:     "session",
			CookieName:        "ops_session",
		},
	}
	app := New(cfg, service.New(cfg, store))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/nodes/node-a/metrics?range=6h", nil)
	req.AddCookie(sessionCookie(t, "ops_session", "session"))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected ok, got %d", resp.StatusCode)
	}
	if store.nodeMetricsID != "node-a" {
		t.Fatalf("expected node metrics for node-a, got %q", store.nodeMetricsID)
	}
	if store.metricsBucket != 5*time.Minute {
		t.Fatalf("expected 6h bucket 5m, got %s", store.metricsBucket)
	}
}

func sessionCookie(t *testing.T, name, secret string) *http.Cookie {
	t.Helper()
	token, err := security.NewSession(secret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Cookie{Name: name, Value: token}
}
