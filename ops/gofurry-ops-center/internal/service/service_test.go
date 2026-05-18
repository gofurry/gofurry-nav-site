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
