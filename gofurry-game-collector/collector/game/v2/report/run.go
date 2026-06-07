package report

import (
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
)

// CollectRun summarizes one collector v2 execution.
type CollectRun struct {
	ID string `json:"id"`

	Task      domain.TaskType `json:"task"`
	Status    domain.Status   `json:"status"`
	StartedAt time.Time       `json:"started_at"`
	EndedAt   time.Time       `json:"ended_at"`

	TotalCount   int `json:"total_count"`
	SuccessCount int `json:"success_count"`
	FailedCount  int `json:"failed_count"`
	SkippedCount int `json:"skipped_count"`

	Error *ErrorInfo `json:"error,omitempty"`
}

// TaskSummary stores aggregate results for one task family inside a run.
type TaskSummary struct {
	Task domain.TaskType `json:"task"`

	TotalCount   int `json:"total_count"`
	SuccessCount int `json:"success_count"`
	FailedCount  int `json:"failed_count"`
	SkippedCount int `json:"skipped_count"`
	PartialCount int `json:"partial_count"`

	DurationMillis int64 `json:"duration_millis"`
}

// RunSummary summarizes one unified v2 runner execution.
type RunSummary struct {
	ID string `json:"id"`

	Status    domain.Status `json:"status"`
	StartedAt time.Time     `json:"started_at"`
	EndedAt   time.Time     `json:"ended_at"`

	TotalCount   int `json:"total_count"`
	SuccessCount int `json:"success_count"`
	FailedCount  int `json:"failed_count"`
	SkippedCount int `json:"skipped_count"`
	PartialCount int `json:"partial_count"`

	TaskSummaries []TaskSummary `json:"task_summaries"`
	Results       []TaskResult  `json:"results,omitempty"`
	Error         *ErrorInfo    `json:"error,omitempty"`
}

// TaskResult summarizes one app-level task result.
type TaskResult struct {
	RunID string `json:"run_id"`

	Task   domain.TaskType `json:"task"`
	Status domain.Status   `json:"status"`

	GameID int64  `json:"game_id"`
	AppID  uint32 `json:"appid"`

	UpstreamStatusCode int    `json:"upstream_status_code"`
	TrafficBucket      string `json:"traffic_bucket"`
	RetryCount         int    `json:"retry_count"`
	DurationMillis     int64  `json:"duration_millis"`

	Error *ErrorInfo `json:"error,omitempty"`

	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}
