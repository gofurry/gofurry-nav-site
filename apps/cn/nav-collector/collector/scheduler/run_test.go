package scheduler

import (
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
)

type memoryStore struct {
	values        map[string]string
	ttls          map[string]time.Duration
	setNXResult   bool
	setNXErr      common.GFError
	setNXCalls    int
	compareResult bool
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		values:        map[string]string{},
		ttls:          map[string]time.Duration{},
		setNXResult:   true,
		compareResult: true,
	}
}

func (s *memoryStore) Set(key string, value any) common.GFError {
	s.values[key] = value.(string)
	return nil
}

func (s *memoryStore) SetExpire(key string, value any, expiration time.Duration) common.GFError {
	s.values[key] = value.(string)
	s.ttls[key] = expiration
	return nil
}

func (s *memoryStore) SetNX(key string, value any, expiration time.Duration) (bool, common.GFError) {
	s.setNXCalls++
	if s.setNXErr != nil {
		return false, s.setNXErr
	}
	if !s.setNXResult {
		return false, nil
	}
	s.values[key] = value.(string)
	s.ttls[key] = expiration
	return true, nil
}

func (s *memoryStore) CompareAndDelete(key string, expected string) (bool, common.GFError) {
	if !s.compareResult || s.values[key] != expected {
		return false, nil
	}
	delete(s.values, key)
	return true, nil
}

func TestSchedulerConfigDefaults(t *testing.T) {
	cfg := env.SchedulerConfig{}
	if cfg.LeaseEnabled {
		t.Fatal("lease should be disabled by default")
	}
	if !cfg.RunStateEnabled() {
		t.Fatal("run state should be enabled by default")
	}
	if got := cfg.RunStateTTL(); got != 168*time.Hour {
		t.Fatalf("RunStateTTL() = %s, want 168h", got)
	}
}

func TestCollectorIDPrefersConfigAndFallsBackToHostname(t *testing.T) {
	if got := CollectorID(env.SchedulerConfig{CollectorID: " collector-a "}); got != "collector-a" {
		t.Fatalf("CollectorID() = %q, want collector-a", got)
	}
	if got := CollectorID(env.SchedulerConfig{}); strings.TrimSpace(got) == "" {
		t.Fatal("CollectorID() fallback should be non-empty")
	}
}

func TestNewJobIDIncludesProtocol(t *testing.T) {
	got := NewJobID("ping")
	if !strings.HasPrefix(got, "ping-") {
		t.Fatalf("NewJobID() = %q, want ping-*", got)
	}
}

func TestRunStateKeys(t *testing.T) {
	if got := RunStateLatestKey("http"); got != "collector:v2:run:http:latest" {
		t.Fatalf("RunStateLatestKey() = %q", got)
	}
	if got := RunStateKey("http", "http-1"); got != "collector:v2:run:http:http-1" {
		t.Fatalf("RunStateKey() = %q", got)
	}
	if got := LeaseKey("dns"); got != "collector:v2:lease:dns" {
		t.Fatalf("LeaseKey() = %q", got)
	}
}

func TestLeaseTTLAutoDerivesFromInterval(t *testing.T) {
	cfg := env.SchedulerConfig{}
	if got := LeaseTTL(cfg, 10*time.Second); got != time.Minute {
		t.Fatalf("LeaseTTL(short) = %s, want 1m", got)
	}
	if got := LeaseTTL(cfg, time.Hour); got != 2*time.Hour {
		t.Fatalf("LeaseTTL(hour) = %s, want 2h", got)
	}
	if got := LeaseTTL(env.SchedulerConfig{LeaseTTLSeconds: 90}, time.Hour); got != 90*time.Second {
		t.Fatalf("LeaseTTL(configured) = %s, want 90s", got)
	}
}

func TestRunWritesStateAndCounts(t *testing.T) {
	store := newMemoryStore()
	run := NewRunWithStore("ping", time.Minute, store)
	run.Start()
	run.SetTargetCount(2)
	run.RecordSuccess()
	run.RecordFailure()
	run.Complete(2)

	if store.values[RunStateLatestKey("ping")] == "" {
		t.Fatal("latest run state should be written")
	}
	if store.values[RunStateKey("ping", run.JobID)] == "" {
		t.Fatal("job run state should be written")
	}
	if store.ttls[RunStateKey("ping", run.JobID)] != 168*time.Hour {
		t.Fatalf("run state ttl = %s, want 168h", store.ttls[RunStateKey("ping", run.JobID)])
	}
	doc := run.Snapshot(StatusComplete, "")
	if doc.SuccessCount != 1 || doc.FailureCount != 1 || doc.ErrorCount != 1 || doc.TargetCount != 2 {
		t.Fatalf("snapshot counts wrong: %+v", doc)
	}
}

func TestRunRefreshRunningWritesProgress(t *testing.T) {
	store := newMemoryStore()
	run := NewRunWithStore("waf_canary", time.Hour, store)
	run.Start()
	run.SetTargetCount(3)
	run.RecordSuccess()
	run.RecordSkippedN(2)
	run.RefreshRunning()

	doc := run.Snapshot(StatusRunning, "")
	if doc.TargetCount != 3 || doc.SuccessCount != 1 || doc.SkippedCount != 2 {
		t.Fatalf("running snapshot counts wrong: %+v", doc)
	}
	if store.values[RunStateLatestKey("waf_canary")] == "" || store.values[RunStateKey("waf_canary", run.JobID)] == "" {
		t.Fatal("RefreshRunning should write latest and job run state")
	}
}

func TestLeaseDisabledDoesNotCallSetNX(t *testing.T) {
	oldCfg := env.GetServerConfig().Collector.Scheduler
	env.GetServerConfig().Collector.Scheduler = env.SchedulerConfig{LeaseEnabled: false}
	t.Cleanup(func() { env.GetServerConfig().Collector.Scheduler = oldCfg })

	store := newMemoryStore()
	run := NewRunWithStore("http", time.Hour, store)
	if !run.AcquireLeaseOrSkip() {
		t.Fatal("disabled lease should allow run")
	}
	if store.setNXCalls != 0 {
		t.Fatalf("SetNX calls = %d, want 0", store.setNXCalls)
	}
}

func TestLeaseAcquireFailureSkipsRun(t *testing.T) {
	oldCfg := env.GetServerConfig().Collector.Scheduler
	env.GetServerConfig().Collector.Scheduler = env.SchedulerConfig{LeaseEnabled: true}
	t.Cleanup(func() { env.GetServerConfig().Collector.Scheduler = oldCfg })

	store := newMemoryStore()
	store.setNXResult = false
	run := NewRunWithStore("dns", time.Hour, store)
	if run.AcquireLeaseOrSkip() {
		t.Fatal("held lease should skip run")
	}
	doc := run.Snapshot(StatusSkipped, "lease_held_by_other_collector")
	if doc.SkippedCount != 1 {
		t.Fatalf("SkippedCount = %d, want 1", doc.SkippedCount)
	}
}

func TestLeaseAcquireRedisErrorUsesDifferentSkipReason(t *testing.T) {
	oldCfg := env.GetServerConfig().Collector.Scheduler
	env.GetServerConfig().Collector.Scheduler = env.SchedulerConfig{LeaseEnabled: true}
	t.Cleanup(func() { env.GetServerConfig().Collector.Scheduler = oldCfg })

	store := newMemoryStore()
	store.setNXErr = common.NewServiceError("redis down")
	run := NewRunWithStore("dns", time.Hour, store)
	if run.AcquireLeaseOrSkip() {
		t.Fatal("Redis SetNX error should skip run")
	}
	doc := run.Snapshot(StatusSkipped, "lease_acquire_failed")
	if doc.SkippedCount != 1 {
		t.Fatalf("SkippedCount = %d, want 1", doc.SkippedCount)
	}
}

func TestLeaseReleaseOnlyDeletesMatchingValue(t *testing.T) {
	oldCfg := env.GetServerConfig().Collector.Scheduler
	env.GetServerConfig().Collector.Scheduler = env.SchedulerConfig{LeaseEnabled: true}
	t.Cleanup(func() { env.GetServerConfig().Collector.Scheduler = oldCfg })

	store := newMemoryStore()
	run := NewRunWithStore("http", time.Hour, store)
	if !run.AcquireLeaseOrSkip() {
		t.Fatal("lease should be acquired")
	}
	run.ReleaseLease()
	if _, ok := store.values[LeaseKey("http")]; ok {
		t.Fatal("matching lease should be deleted")
	}

	run = NewRunWithStore("http", time.Hour, store)
	if !run.AcquireLeaseOrSkip() {
		t.Fatal("lease should be acquired")
	}
	store.values[LeaseKey("http")] = "other"
	run.ReleaseLease()
	if store.values[LeaseKey("http")] != "other" {
		t.Fatal("non-matching lease should not be deleted")
	}
}
