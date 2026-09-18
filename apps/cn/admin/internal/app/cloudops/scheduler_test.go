package cloudops

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	infra "github.com/gofurry/gofurry-admin/internal/infra/cloudops"
)

type scheduledAudit struct {
	meta   audit.Meta
	host   string
	detail map[string]any
}

type fakeScheduledStore struct {
	mu                    sync.Mutex
	held                  bool
	acquireErr, lookupErr error
	failAudit             string
	events                []string
	audits                []scheduledAudit
	intents               map[string]bool
}

func (f *fakeScheduledStore) event(event string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, event)
}
func (f *fakeScheduledStore) Acquire(context.Context) (scheduledPurgeLease, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, "acquire")
	if f.acquireErr != nil {
		return nil, f.acquireErr
	}
	if f.held {
		return nil, nil
	}
	f.held = true
	return f, nil
}
func (f *fakeScheduledStore) Recorded(_ context.Context, slot, host string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, "lookup")
	return f.intents[slot+host], f.lookupErr
}
func (f *fakeScheduledStore) Audit(ctx context.Context, meta audit.Meta, host string, detail any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	value := detail.(map[string]any)
	status := value["status"].(string)
	f.events = append(f.events, status)
	if f.failAudit == status {
		return errors.New("audit unavailable")
	}
	f.audits = append(f.audits, scheduledAudit{meta: meta, host: host, detail: value})
	if status == "requested" {
		if f.intents == nil {
			f.intents = make(map[string]bool)
		}
		f.intents[meta.RequestID+host] = true
	}
	return nil
}
func (f *fakeScheduledStore) Release(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, "release")
	f.held = false
	return ctx.Err()
}

type fakeScheduledPurger struct {
	t             *testing.T
	store         *fakeScheduledStore
	calls         atomic.Int32
	run           func(context.Context) (infra.PurgeResult, error)
	validationErr error
}

func (f *fakeScheduledPurger) Validate(provider string, req infra.PurgeRequest) (infra.PurgeRequest, error) {
	if provider != "edgeone" {
		f.t.Errorf("provider = %q", provider)
	}
	if f.validationErr != nil {
		return req, f.validationErr
	}
	return infra.ValidatePurge(req, "main.example.test", "assets.example.test")
}
func (f *fakeScheduledPurger) Purge(ctx context.Context, provider string, req infra.PurgeRequest) (infra.PurgeResult, error) {
	f.calls.Add(1)
	f.store.event("purge")
	if provider != "edgeone" || req.Type != "host" || !reflect.DeepEqual(req.Targets, []string{"main.example.test"}) {
		f.t.Errorf("unexpected purge %q %+v", provider, req)
	}
	deadline, bounded := ctx.Deadline()
	if !bounded || time.Until(deadline) > 25*time.Second {
		f.t.Error("purge requires a bounded background context")
	}
	if f.run != nil {
		return f.run(ctx)
	}
	return infra.PurgeResult{JobID: "job-122", Status: "submitted"}, nil
}

func scheduledConfig() env.EdgeOneConfig {
	return env.EdgeOneConfig{Enabled: true, ZoneID: "zone-test", SecretID: "test-id", SecretKey: "test-secret", MainHost: "main.example.test", AssetHost: "assets.example.test",
		ScheduledPurge: env.EdgeOneScheduledPurgeConfig{Enabled: true, Time: "05:30", Timezone: "Asia/Shanghai"}}
}
func testScheduler(t *testing.T, store *fakeScheduledStore, service *fakeScheduledPurger) *Scheduler {
	t.Helper()
	s, err := newScheduler(scheduledConfig(), service, store)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Stop)
	return s
}
func testSlot() time.Time {
	value, _ := time.Parse(time.RFC3339, "2026-09-19T05:30:00+08:00")
	return value
}

func TestScheduledPurgeAuditAndSlotDedupe(t *testing.T) {
	store := &fakeScheduledStore{}
	service := &fakeScheduledPurger{t: t, store: store}
	s := testScheduler(t, store, service)
	s.runSlot(testSlot())
	want := []string{"acquire", "lookup", "requested", "purge", "completed", "release"}
	if !reflect.DeepEqual(store.events, want) {
		t.Fatalf("operation ordering: %v", store.events)
	}
	for _, record := range store.audits {
		if record.meta.Operator != "system" || record.meta.OperatorAccountID != nil || record.meta.RequestID != "edgeone-scheduled-purge:20260919T0530+0800" || record.host != "main.example.test" {
			t.Fatalf("audit metadata: %+v", record)
		}
	}
	input := store.audits[0].detail["input"].(map[string]any)
	if input["scheduled_for"] != "2026-09-19T05:30:00+08:00" || input["type"] != "host" || !reflect.DeepEqual(input["targets"], []string{"main.example.test"}) {
		t.Fatalf("intent input: %v", input)
	}
	if result := store.audits[1].detail["result"].(infra.PurgeResult); result.JobID != "job-122" || result.Status != "submitted" {
		t.Fatalf("audit result: %+v", result)
	}
	s.runSlot(testSlot())
	// A late second instance/restart acquires after the first releases the lock.
	// It must consult durable intent, not submit another purge for the same slot.
	second := testScheduler(t, store, service)
	second.runSlot(testSlot())
	if service.calls.Load() != 1 || len(store.audits) != 2 || store.held {
		t.Fatal("same slot was resubmitted or lock leaked")
	}
	second.runSlot(testSlot().AddDate(0, 0, 1))
	if service.calls.Load() != 2 || len(store.audits) != 4 {
		t.Fatal("next daily slot did not submit")
	}
}

func TestScheduledPurgeFailureBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name       string
		configure  func(*fakeScheduledStore, *fakeScheduledPurger)
		wantEvents []string
		calls      int32
	}{
		{"lock held", func(s *fakeScheduledStore, _ *fakeScheduledPurger) { s.held = true }, []string{"acquire"}, 0},
		{"lock query failed", func(s *fakeScheduledStore, _ *fakeScheduledPurger) { s.acquireErr = errors.New("DB unavailable") }, []string{"acquire"}, 0},
		{"lookup failed", func(s *fakeScheduledStore, _ *fakeScheduledPurger) { s.lookupErr = errors.New("DB unavailable") }, []string{"acquire", "lookup", "release"}, 0},
		{"intent failed", func(s *fakeScheduledStore, _ *fakeScheduledPurger) { s.failAudit = "requested" }, []string{"acquire", "lookup", "requested", "release"}, 0},
		{"remote error", func(_ *fakeScheduledStore, p *fakeScheduledPurger) {
			p.run = func(context.Context) (infra.PurgeResult, error) { return infra.PurgeResult{}, context.DeadlineExceeded }
		}, []string{"acquire", "lookup", "requested", "purge", "failed", "release"}, 1},
		{"provider rejected target", func(_ *fakeScheduledStore, p *fakeScheduledPurger) {
			p.run = func(context.Context) (infra.PurgeResult, error) {
				return infra.PurgeResult{Status: "failed", Failures: []infra.Failure{{Reason: "not allowed"}}}, nil
			}
		}, []string{"acquire", "lookup", "requested", "purge", "failed", "release"}, 1},
		{"result audit failed", func(s *fakeScheduledStore, _ *fakeScheduledPurger) { s.failAudit = "completed" }, []string{"acquire", "lookup", "requested", "purge", "completed", "release"}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &fakeScheduledStore{}
			service := &fakeScheduledPurger{t: t, store: store}
			tc.configure(store, service)
			s := testScheduler(t, store, service)
			s.runSlot(testSlot())
			s.runSlot(testSlot())
			if service.calls.Load() != tc.calls || !reflect.DeepEqual(store.events, tc.wantEvents) {
				t.Fatalf("calls=%d events=%v", service.calls.Load(), store.events)
			}
			for _, record := range store.audits {
				if record.detail["status"] == "failed" {
					message, ok := record.detail["error"].(string)
					if !ok || message == "" {
						t.Fatal("failed audit lacks error")
					}
				}
			}
			if tc.calls > 0 {
				second := testScheduler(t, store, service)
				second.runSlot(testSlot())
				if service.calls.Load() != tc.calls {
					t.Fatal("another instance retried an ambiguous or completed submission")
				}
			}
		})
	}
}

func TestScheduledPurgeStartStopAndNoCatchUp(t *testing.T) {
	store := &fakeScheduledStore{}
	service := &fakeScheduledPurger{t: t, store: store}
	s := testScheduler(t, store, service)
	// Start/restart at 06:00 must schedule tomorrow, not replay 05:30.
	schedule := s.cron.Entry(s.entry).Schedule
	if next := schedule.Next(testSlot().Add(30 * time.Minute)); !next.Equal(testSlot().AddDate(0, 0, 1)) {
		t.Fatalf("next run = %v", next)
	}
	s.Start()
	s.Start()
	if s.cron.Entry(s.entry).Next.IsZero() {
		t.Fatal("cron did not start")
	}
	s.Stop()
	s.Stop()
	s.Start()
	s.runSlot(testSlot())
	if service.calls.Load() != 0 {
		t.Fatal("startup, shutdown or restart triggered purge")
	}
	var disabled *Scheduler
	disabled.Start()
	disabled.Stop()
	if scheduler, err := NewScheduler(env.EdgeOneConfig{}, nil, nil, nil); err != nil || scheduler != nil {
		t.Fatal("disabled scheduler requires dependencies")
	}
	service.validationErr = errors.New("host is not allowed")
	if _, err := newScheduler(scheduledConfig(), service, store); err == nil {
		t.Fatal("existing host validation was bypassed")
	}
}

func TestScheduledPurgeOverlapAndShutdownDrain(t *testing.T) {
	store := &fakeScheduledStore{}
	entered, canceled, finish := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var finishOnce sync.Once
	allowFinish := func() { finishOnce.Do(func() { close(finish) }) }
	service := &fakeScheduledPurger{t: t, store: store, run: func(ctx context.Context) (infra.PurgeResult, error) {
		close(entered)
		<-ctx.Done()
		close(canceled)
		<-finish
		return infra.PurgeResult{}, ctx.Err()
	}}
	s := testScheduler(t, store, service)
	t.Cleanup(allowFinish)
	go s.runSlot(testSlot())
	waitSignal(t, entered)
	s.runSlot(testSlot().AddDate(0, 0, 1))
	second := testScheduler(t, store, service)
	second.runSlot(testSlot())
	if service.calls.Load() != 1 {
		t.Fatal("overlapping run was not skipped")
	}
	stopped := make(chan struct{})
	go func() { s.Stop(); close(stopped) }()
	waitSignal(t, canceled)
	select {
	case <-stopped:
		t.Fatal("Stop returned before work completed")
	default:
	}
	allowFinish()
	waitSignal(t, stopped)
	if store.held || len(store.audits) != 2 || store.audits[1].detail["status"] != "failed" {
		t.Fatal("shutdown failed to audit outcome and release lock")
	}
	if store.events[len(store.events)-1] != "release" {
		t.Fatal("session released before outcome audit")
	}
}

func waitSignal(t *testing.T, signal <-chan struct{}) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for scheduler")
	}
}
