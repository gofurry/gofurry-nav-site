package scheduler

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/log"
	cs "github.com/gofurry/gofurry-nav-collector/common/service"
	"github.com/gofurry/gofurry-nav-collector/common/util"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
)

const (
	StatusRunning  = "running"
	StatusComplete = "complete"
	StatusFailed   = "failed"
	StatusSkipped  = "skipped"

	defaultRunStateTTLHours = 168
	minLeaseTTL             = time.Minute
)

type Store interface {
	Set(key string, value any) common.GFError
	SetExpire(key string, value any, expiration time.Duration) common.GFError
	SetNX(key string, value any, expiration time.Duration) (bool, common.GFError)
	CompareAndDelete(key string, expected string) (bool, common.GFError)
}

type redisStore struct{}

func (redisStore) Set(key string, value any) common.GFError {
	return cs.Set(key, value)
}

func (redisStore) SetExpire(key string, value any, expiration time.Duration) common.GFError {
	return cs.SetExpire(key, value, expiration)
}

func (redisStore) SetNX(key string, value any, expiration time.Duration) (bool, common.GFError) {
	return cs.SetNX(key, value, expiration)
}

func (redisStore) CompareAndDelete(key string, expected string) (bool, common.GFError) {
	return cs.CompareAndDelete(key, expected)
}

type Run struct {
	CollectorID string
	JobID       string
	Protocol    string
	StartedAt   time.Time
	FinishedAt  time.Time

	interval  time.Duration
	store     Store
	leaseKey  string
	leaseJSON string

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

type LeaseDocument struct {
	CollectorID string    `json:"collector_id"`
	JobID       string    `json:"job_id"`
	Protocol    string    `json:"protocol"`
	AcquiredAt  time.Time `json:"acquired_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func NewRun(protocol string, interval time.Duration) *Run {
	return NewRunWithStore(protocol, interval, redisStore{})
}

func NewRunWithStore(protocol string, interval time.Duration, store Store) *Run {
	if store == nil {
		store = redisStore{}
	}
	return &Run{
		CollectorID: CollectorID(env.GetServerConfig().Collector.Scheduler),
		JobID:       NewJobID(protocol),
		Protocol:    protocol,
		StartedAt:   time.Now(),
		interval:    interval,
		store:       store,
	}
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

func RunStateLatestKey(protocol string) string {
	return fmt.Sprintf("collector:v2:run:%s:latest", protocol)
}

func RunStateKey(protocol string, jobID string) string {
	return fmt.Sprintf("collector:v2:run:%s:%s", protocol, jobID)
}

func LeaseKey(protocol string) string {
	return fmt.Sprintf("collector:v2:lease:%s", protocol)
}

func LeaseTTL(cfg env.SchedulerConfig, interval time.Duration) time.Duration {
	if cfg.LeaseTTLSeconds > 0 {
		return time.Duration(cfg.LeaseTTLSeconds) * time.Second
	}
	ttl := interval * 2
	if ttl < minLeaseTTL {
		return minLeaseTTL
	}
	return ttl
}

func (r *Run) Fields() map[string]interface{} {
	return map[string]interface{}{
		"collector_id": r.CollectorID,
		"job_id":       r.JobID,
		"protocol":     r.Protocol,
	}
}

func (r *Run) Start() {
	r.StartedAt = time.Now()
	r.writeState(StatusRunning, "")
}

func (r *Run) AcquireLeaseOrSkip() bool {
	cfg := env.GetServerConfig().Collector.Scheduler
	if !cfg.LeaseEnabled {
		return true
	}

	ttl := LeaseTTL(cfg, r.interval)
	now := time.Now()
	lease := LeaseDocument{
		CollectorID: r.CollectorID,
		JobID:       r.JobID,
		Protocol:    r.Protocol,
		AcquiredAt:  now,
		ExpiresAt:   now.Add(ttl),
	}
	leaseBytes, err := sonic.Marshal(lease)
	if err != nil {
		r.Skip("lease_encode_failed", 0)
		return false
	}

	key := LeaseKey(r.Protocol)
	value := string(leaseBytes)
	created, setErr := r.store.SetNX(key, value, ttl)
	if setErr != nil {
		r.Skip("lease_acquire_failed", 0)
		fields := r.Fields()
		fields["event"] = "lease_acquire_failed"
		fields["redis_key"] = key
		log.WarnFields(fields, "采集 lease 获取失败: "+setErr.GetMsg())
		return false
	}
	if !created {
		r.Skip("lease_held_by_other_collector", 0)
		return false
	}
	r.leaseKey = key
	r.leaseJSON = value
	return true
}

func (r *Run) ReleaseLease() {
	if r.leaseKey == "" || r.leaseJSON == "" {
		return
	}
	deleted, err := r.store.CompareAndDelete(r.leaseKey, r.leaseJSON)
	if err != nil {
		fields := r.Fields()
		fields["event"] = "lease_release_failed"
		fields["redis_key"] = r.leaseKey
		log.WarnFields(fields, "采集 lease 释放失败: "+err.GetMsg())
		return
	}
	if !deleted {
		fields := r.Fields()
		fields["event"] = "lease_release_skipped"
		fields["redis_key"] = r.leaseKey
		log.WarnFields(fields, "采集 lease 未释放：当前持有者已变化")
	}
}

func (r *Run) SetTargetCount(count int) {
	r.targetCount.Store(int64(count))
}

func (r *Run) RecordSuccess() {
	r.successCount.Add(1)
}

func (r *Run) RecordFailure() {
	r.failureCount.Add(1)
	r.errorCount.Add(1)
}

func (r *Run) RecordSkipped() {
	r.skippedCount.Add(1)
}

func (r *Run) RecordSkippedN(count int) {
	if count <= 0 {
		return
	}
	r.skippedCount.Add(int64(count))
}

func (r *Run) RecordRunError() {
	r.errorCount.Add(1)
}

func (r *Run) RefreshRunning() {
	r.writeState(StatusRunning, "")
}

func (r *Run) Complete(targetCount int) {
	r.SetTargetCount(targetCount)
	r.writeState(StatusComplete, "")
}

func (r *Run) Fail(skipReason string, targetCount int) {
	r.SetTargetCount(targetCount)
	r.RecordRunError()
	r.writeState(StatusFailed, skipReason)
}

func (r *Run) Skip(reason string, targetCount int) {
	r.SetTargetCount(targetCount)
	r.RecordSkipped()
	r.writeState(StatusSkipped, reason)
}

func (r *Run) Snapshot(status string, skipReason string) StateDocument {
	finishedAt := time.Now()
	durationMS := finishedAt.Sub(r.StartedAt).Milliseconds()
	if status == StatusRunning {
		finishedAt = time.Time{}
		durationMS = 0
	}
	return StateDocument{
		CollectorID:  r.CollectorID,
		JobID:        r.JobID,
		Protocol:     r.Protocol,
		Status:       status,
		StartedAt:    r.StartedAt,
		FinishedAt:   finishedAt,
		DurationMS:   durationMS,
		TargetCount:  r.targetCount.Load(),
		SuccessCount: r.successCount.Load(),
		FailureCount: r.failureCount.Load(),
		SkippedCount: r.skippedCount.Load(),
		ErrorCount:   r.errorCount.Load(),
		SkipReason:   skipReason,
	}
}

func (r *Run) writeState(status string, skipReason string) {
	cfg := env.GetServerConfig().Collector.Scheduler
	if !cfg.RunStateEnabled() {
		return
	}
	doc := r.Snapshot(status, skipReason)
	raw, err := sonic.Marshal(doc)
	if err != nil {
		fields := r.Fields()
		fields["event"] = "run_state_encode_failed"
		log.WarnFields(fields, "采集运行状态 JSON 编码失败: "+err.Error())
		return
	}
	if err := r.store.Set(RunStateLatestKey(r.Protocol), string(raw)); err != nil {
		fields := r.Fields()
		fields["event"] = "run_state_latest_write_failed"
		fields["redis_key"] = RunStateLatestKey(r.Protocol)
		log.WarnFields(fields, "采集 latest 运行状态写入 Redis 失败: "+err.GetMsg())
	}
	if err := r.store.SetExpire(RunStateKey(r.Protocol, r.JobID), string(raw), cfg.RunStateTTL()); err != nil {
		fields := r.Fields()
		fields["event"] = "run_state_write_failed"
		fields["redis_key"] = RunStateKey(r.Protocol, r.JobID)
		log.WarnFields(fields, "采集运行状态写入 Redis 失败: "+err.GetMsg())
	}
}
