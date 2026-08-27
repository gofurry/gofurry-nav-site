package scheduler

import (
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-collector/collector/execution"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
)

func TestCollectorIDPrefersConfigAndFallsBackToHostname(t *testing.T) {
	if got := CollectorID(env.SchedulerConfig{CollectorID: " collector-a "}); got != "collector-a" {
		t.Fatalf("CollectorID() = %q, want collector-a", got)
	}
	if got := CollectorID(env.SchedulerConfig{}); strings.TrimSpace(got) == "" {
		t.Fatal("CollectorID() fallback should be non-empty")
	}
}

func TestNewJobIDIncludesProtocol(t *testing.T) {
	if got := NewJobID("ping"); !strings.HasPrefix(got, "ping-") {
		t.Fatalf("NewJobID() = %q, want ping-*", got)
	}
}

func TestRunUsesDurableExecutionLineageAndCountsLocally(t *testing.T) {
	err := execution.With("ping", execution.Request{
		JobID: 42, RunID: "nav-ping-run", InstanceID: "instance-a", ScopeType: "all",
	}, func() {
		run := NewRun("ping", time.Minute)
		if run.JobID != "nav-ping-run" || run.CollectorID != "instance-a" {
			t.Fatalf("durable lineage not applied: %+v", run)
		}
		if !run.AcquireLeaseOrSkip() {
			t.Fatal("PostgreSQL-owned execution must not acquire a Redis lease")
		}
		run.Start()
		run.SetTargetCount(2)
		run.RecordSuccess()
		run.RecordFailure()
		run.Complete(2)
		doc := run.Snapshot(StatusComplete, "")
		if doc.SuccessCount != 1 || doc.FailureCount != 1 || doc.ErrorCount != 1 || doc.TargetCount != 2 {
			t.Fatalf("snapshot counts wrong: %+v", doc)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}
