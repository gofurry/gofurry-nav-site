package models

import "time"

type Overview struct {
	RunningCount int64 `json:"running_count"`
	QueuedCount  int64 `json:"queued_count"`
	Failed24h    int64 `json:"failed_24h"`
	Missed24h    int64 `json:"missed_24h"`
}

type Instance struct {
	Domain          string     `json:"domain"`
	InstanceID      string     `json:"instance_id"`
	CollectorID     string     `json:"collector_id"`
	Hostname        string     `json:"hostname"`
	Version         string     `json:"version"`
	CommitSHA       string     `json:"commit_sha"`
	Capabilities    []string   `json:"capabilities"`
	StartedAt       time.Time  `json:"started_at"`
	LastHeartbeatAt time.Time  `json:"last_heartbeat_at"`
	StoppedAt       *time.Time `json:"stopped_at,omitempty"`
	Health          string     `json:"health"`
	HeartbeatAgeSec int64      `json:"heartbeat_age_seconds"`
}

type Schedule struct {
	Domain              string     `json:"domain"`
	ID                  int64      `json:"id"`
	JobKey              string     `json:"job_key"`
	Name                string     `json:"name"`
	Enabled             bool       `json:"enabled"`
	ScheduleKind        string     `json:"schedule_kind"`
	CronExpression      *string    `json:"cron_expression,omitempty"`
	IntervalSeconds     *int64     `json:"interval_seconds,omitempty"`
	AnchorAt            *time.Time `json:"anchor_at,omitempty"`
	Timezone            string     `json:"timezone"`
	MisfirePolicy       string     `json:"misfire_policy"`
	MisfireGraceSeconds int32      `json:"misfire_grace_seconds"`
	OverlapPolicy       string     `json:"overlap_policy"`
	Priority            int32      `json:"priority"`
	ConcurrencyKey      string     `json:"concurrency_key"`
	Version             int64      `json:"version"`
	LastMaterializedFor *time.Time `json:"last_materialized_for,omitempty"`
	NextScheduledFor    *time.Time `json:"next_scheduled_for,omitempty"`
	LastStatus          string     `json:"last_status"`
	LastSuccessCount    int32      `json:"last_success_count"`
	LastExpectedCount   int32      `json:"last_expected_count"`
	LastSuccessCoverage float64    `json:"last_success_coverage"`
}

type ScheduleUpdate struct {
	Enabled             bool       `json:"enabled"`
	ScheduleKind        string     `json:"schedule_kind"`
	CronExpression      *string    `json:"cron_expression"`
	IntervalSeconds     *int64     `json:"interval_seconds"`
	AnchorAt            *time.Time `json:"anchor_at"`
	Timezone            string     `json:"timezone"`
	MisfirePolicy       string     `json:"misfire_policy"`
	MisfireGraceSeconds int32      `json:"misfire_grace_seconds"`
}

type ManualJobRequest struct {
	Domain    string   `json:"domain"`
	ScopeType string   `json:"scope_type"`
	ScopeID   *int64   `json:"scope_id"`
	Target    *string  `json:"target"`
	Tasks     []string `json:"tasks"`
}

type Job struct {
	Domain            string     `json:"domain"`
	ID                int64      `json:"id"`
	ScheduleID        *int64     `json:"schedule_id,omitempty"`
	JobKey            string     `json:"job_key"`
	Trigger           string     `json:"trigger"`
	ScopeType         string     `json:"scope_type"`
	ScopeID           *int64     `json:"scope_id,omitempty"`
	Target            *string    `json:"target,omitempty"`
	Tasks             []string   `json:"tasks"`
	Priority          int32      `json:"priority"`
	ScheduledFor      *time.Time `json:"scheduled_for,omitempty"`
	Status            string     `json:"status"`
	RequestedBy       string     `json:"requested_by"`
	ClaimedBy         *string    `json:"claimed_by,omitempty"`
	LeaseUntil        *time.Time `json:"lease_until,omitempty"`
	CancelRequestedAt *time.Time `json:"cancel_requested_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	RunID             string     `json:"run_id,omitempty"`
	Progress          any        `json:"progress,omitempty"`
}

type Run struct {
	Domain              string     `json:"domain"`
	ID                  string     `json:"id"`
	JobID               int64      `json:"job_id"`
	JobKey              string     `json:"job_key"`
	Trigger             string     `json:"trigger"`
	ScopeType           string     `json:"scope_type"`
	ScopeID             *int64     `json:"scope_id,omitempty"`
	Target              *string    `json:"target,omitempty"`
	AttemptNo           int32      `json:"attempt_no"`
	CollectorInstanceID string     `json:"collector_instance_id"`
	Status              string     `json:"status"`
	ScheduledFor        *time.Time `json:"scheduled_for,omitempty"`
	StartedAt           time.Time  `json:"started_at"`
	EndedAt             *time.Time `json:"ended_at,omitempty"`
	ExpectedCount       int32      `json:"expected_count"`
	AttemptedCount      int32      `json:"attempted_count"`
	SuccessCount        int32      `json:"success_count"`
	PartialCount        int32      `json:"partial_count"`
	FailureCount        int32      `json:"failure_count"`
	SkippedCount        int32      `json:"skipped_count"`
	ScheduleDelayMS     int64      `json:"schedule_delay_ms"`
	DurationMS          int64      `json:"duration_ms"`
	ErrorKind           string     `json:"error_kind"`
	ErrorMessage        string     `json:"error_message"`
}

type Result struct {
	Domain        string     `json:"domain"`
	ID            int64      `json:"id"`
	RunID         string     `json:"run_id"`
	Task          string     `json:"task"`
	EntityID      int64      `json:"entity_id"`
	Target        string     `json:"target,omitempty"`
	Status        string     `json:"status"`
	ObservationID *int64     `json:"observation_id,omitempty"`
	DurationMS    int64      `json:"duration_ms"`
	ErrorKind     string     `json:"error_kind"`
	ErrorMessage  string     `json:"error_message"`
	StartedAt     time.Time  `json:"started_at"`
	EndedAt       *time.Time `json:"ended_at,omitempty"`
	AppID         int64      `json:"appid,omitempty"`
}

type ChartPoint struct {
	Domain          string     `json:"domain"`
	JobID           int64      `json:"job_id"`
	JobKey          string     `json:"job_key"`
	JobStatus       string     `json:"job_status"`
	RunID           string     `json:"run_id,omitempty"`
	RunStatus       string     `json:"run_status,omitempty"`
	Expected        int32      `json:"expected"`
	Attempted       int32      `json:"attempted"`
	Success         int32      `json:"success"`
	Partial         int32      `json:"partial"`
	Failed          int32      `json:"failed"`
	Skipped         int32      `json:"skipped"`
	Coverage        float64    `json:"coverage"`
	ScheduleDelayMS int64      `json:"schedule_delay_ms"`
	DurationMS      int64      `json:"duration_ms"`
	CreatedAt       time.Time  `json:"created_at"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
}
