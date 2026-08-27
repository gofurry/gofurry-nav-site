package scheduler

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gofurry/gofurry-nav-collector/collector/execution"
	"github.com/gofurry/gofurry-nav-collector/common/util"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
)

const (
	StatusRunning  = "running"
	StatusComplete = "complete"
	StatusFailed   = "failed"
	StatusSkipped  = "skipped"
)

// Run is the lightweight in-process probe counter used by the mature protocol
// implementations. Durable ownership, lease, status, and history are managed
// exclusively by PostgreSQL in collector/control.
type Run struct {
	CollectorID string
	JobID       string
	Protocol    string
	StartedAt   time.Time
	FinishedAt  time.Time

	targetCount  atomic.Int64
	successCount atomic.Int64
	failureCount atomic.Int64
	skippedCount atomic.Int64
	errorCount   atomic.Int64
}

type StateDocument struct {
	CollectorID  string    `json:"collector_id"`
	JobID        string    `json:"job_id"`
	Protocol     string    `json:"protocol"`
	Status       string    `json:"status"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at,omitempty"`
	DurationMS   int64     `json:"duration_ms"`
	TargetCount  int64     `json:"target_count"`
	SuccessCount int64     `json:"success_count"`
	FailureCount int64     `json:"failure_count"`
	SkippedCount int64     `json:"skipped_count"`
	ErrorCount   int64     `json:"error_count"`
	SkipReason   string    `json:"skip_reason,omitempty"`
}

func NewRun(protocol string, _ time.Duration) *Run {
	collectorID := CollectorID(env.GetServerConfig().Collector.Scheduler)
	jobID := NewJobID(protocol)
	if request, ok := execution.Current(protocol); ok {
		collectorID = request.InstanceID
		jobID = request.RunID
	}
	return &Run{CollectorID: collectorID, JobID: jobID, Protocol: protocol, StartedAt: time.Now()}
}

func CollectorID(cfg env.SchedulerConfig) string {
	id := strings.TrimSpace(cfg.CollectorID)
	if id != "" {
		return id
	}
	hostname, err := os.Hostname()
	if err != nil || strings.TrimSpace(hostname) == "" {
		return "unknown-host"
	}
	return strings.TrimSpace(hostname)
}

func NewJobID(protocol string) string {
	return fmt.Sprintf("%s-%d", protocol, util.GenerateId())
}

func (r *Run) Fields() map[string]interface{} {
	return map[string]interface{}{"collector_id": r.CollectorID, "job_id": r.JobID, "protocol": r.Protocol}
}

func (r *Run) Start() { r.StartedAt = time.Now() }

func (r *Run) AcquireLeaseOrSkip() bool { return true }

func (r *Run) ReleaseLease() {}

func (r *Run) SetTargetCount(count int) { r.targetCount.Store(int64(count)) }

func (r *Run) RecordSuccess() { r.successCount.Add(1) }

func (r *Run) RecordFailure() {
	r.failureCount.Add(1)
	r.errorCount.Add(1)
}

func (r *Run) RecordSkipped() { r.skippedCount.Add(1) }

func (r *Run) RecordSkippedN(count int) {
	if count > 0 {
		r.skippedCount.Add(int64(count))
	}
}

func (r *Run) RecordRunError() { r.errorCount.Add(1) }

func (r *Run) RefreshRunning() {}

func (r *Run) Complete(targetCount int) {
	r.SetTargetCount(targetCount)
	r.FinishedAt = time.Now()
}

func (r *Run) Fail(_ string, targetCount int) {
	r.SetTargetCount(targetCount)
	r.RecordRunError()
	r.FinishedAt = time.Now()
}

func (r *Run) Skip(_ string, targetCount int) {
	r.SetTargetCount(targetCount)
	r.RecordSkipped()
	r.FinishedAt = time.Now()
}

func (r *Run) Snapshot(status string, skipReason string) StateDocument {
	finishedAt := r.FinishedAt
	if finishedAt.IsZero() && status != StatusRunning {
		finishedAt = time.Now()
	}
	durationMS := int64(0)
	if !finishedAt.IsZero() {
		durationMS = finishedAt.Sub(r.StartedAt).Milliseconds()
	}
	return StateDocument{
		CollectorID: r.CollectorID, JobID: r.JobID, Protocol: r.Protocol,
		Status: status, StartedAt: r.StartedAt, FinishedAt: finishedAt,
		DurationMS: durationMS, TargetCount: r.targetCount.Load(),
		SuccessCount: r.successCount.Load(), FailureCount: r.failureCount.Load(),
		SkippedCount: r.skippedCount.Load(), ErrorCount: r.errorCount.Load(), SkipReason: skipReason,
	}
}
