package service

import (
	"context"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/config"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/model"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/repository"
)

type fakeStore struct {
	alerts   map[string]repository.AlertInput
	resolved []string
	nodes    []model.Node
	bucket   time.Duration
}

func (f *fakeStore) Ingest(context.Context, model.AgentPayload, int) error { return nil }
func (f *fakeStore) UpsertAlert(_ context.Context, input repository.AlertInput) error {
	if f.alerts == nil {
		f.alerts = map[string]repository.AlertInput{}
	}
	f.alerts[input.Key] = input
	return nil
}
func (f *fakeStore) ResolveAlert(_ context.Context, key string) error {
	f.resolved = append(f.resolved, key)
	return nil
}
func (f *fakeStore) ListNodes(context.Context) ([]model.Node, error)      { return f.nodes, nil }
func (f *fakeStore) GetNode(context.Context, string) (model.Node, error)  { return model.Node{}, nil }
func (f *fakeStore) MarkNodeStatus(context.Context, string, string) error { return nil }
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
func (f *fakeStore) OverviewMetrics(_ context.Context, _ time.Time, bucket time.Duration) (model.OverviewMetrics, error) {
	f.bucket = bucket
	return model.OverviewMetrics{}, nil
}
func (f *fakeStore) NodeMetrics(_ context.Context, _ string, _ time.Time, bucket time.Duration) (model.NodeMetrics, error) {
	f.bucket = bucket
	return model.NodeMetrics{}, nil
}

func TestIngestCreatesDiskAlert(t *testing.T) {
	store := &fakeStore{}
	svc := New(config.Config{
		Region: "cn",
		Alert: config.AlertConfig{
			Enabled:       true,
			DiskUsageWarn: 85,
		},
	}, store)
	err := svc.Ingest(context.Background(), model.AgentPayload{
		NodeID:    "cn-business-a",
		Region:    "cn",
		Timestamp: time.Now(),
		Disks:     []model.DiskSample{{Mount: "/", Usage: 91}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := store.alerts["disk:cn-business-a:_"]; !ok {
		t.Fatalf("expected disk alert, got %#v", store.alerts)
	}
}

func TestEvaluateNodeOffline(t *testing.T) {
	lastSeen := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	store := &fakeStore{nodes: []model.Node{{NodeID: "n1", Region: "cn", LastSeenAt: &lastSeen}}}
	svc := New(config.Config{
		Alert: config.AlertConfig{
			NodeDownAfter: config.Duration{Duration: time.Minute},
		},
	}, store)
	svc.now = func() time.Time { return lastSeen.Add(2 * time.Minute) }
	if err := svc.EvaluateNodeOffline(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, ok := store.alerts["node_down:cn:n1"]; !ok {
		t.Fatalf("expected node down alert, got %#v", store.alerts)
	}
}

func TestResolveMetricsWindow(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		input  string
		label  string
		since  time.Time
		bucket time.Duration
	}{
		{input: "", label: "1h", since: now.Add(-time.Hour), bucket: time.Minute},
		{input: "bad", label: "1h", since: now.Add(-time.Hour), bucket: time.Minute},
		{input: "6h", label: "6h", since: now.Add(-6 * time.Hour), bucket: 5 * time.Minute},
		{input: "24h", label: "24h", since: now.Add(-24 * time.Hour), bucket: 15 * time.Minute},
	}
	for _, tt := range tests {
		got := resolveMetricsWindow(tt.input, now)
		if got.Label != tt.label || !got.Since.Equal(tt.since) || got.Bucket != tt.bucket {
			t.Fatalf("resolveMetricsWindow(%q) = %#v", tt.input, got)
		}
	}
}

func TestOverviewMetricsAddsMetadata(t *testing.T) {
	store := &fakeStore{}
	svc := New(config.Config{CenterID: "ops", Region: "cn"}, store)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }
	result, err := svc.OverviewMetrics(context.Background(), "6h")
	if err != nil {
		t.Fatal(err)
	}
	if result.CenterID != "ops" || result.Region != "cn" || result.Range != "6h" || !result.GeneratedAt.Equal(now) {
		t.Fatalf("unexpected metadata: %#v", result)
	}
	if store.bucket != 5*time.Minute {
		t.Fatalf("expected 6h bucket, got %s", store.bucket)
	}
}
