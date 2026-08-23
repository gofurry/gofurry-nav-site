package models

import (
	"time"

	readmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/readmodel/models"
	summarymodels "github.com/gofurry/gofurry-nav-backend/apps/nav/summary/models"
)

type CollectStatus struct {
	LatestRuns  []readmodels.RunStateResponse `json:"latest_runs"`
	Summary     []ObservationStatusSummary    `json:"summary"`
	GeneratedAt time.Time                     `json:"generated_at"`
}

type ObservationStatusSummary struct {
	Protocol string `json:"protocol"`
	Status   string `json:"status"`
	Count    int64  `json:"count"`
}

type ObservationQuery struct {
	SiteID   int64
	Target   string
	Protocol string
	Status   string
	Limit    int
	Offset   int
}

type ObservationItem struct {
	ID           int64     `json:"id"`
	SiteID       int64     `json:"site_id"`
	Target       string    `json:"target"`
	Protocol     string    `json:"protocol"`
	Status       string    `json:"status"`
	ObservedAt   time.Time `json:"observed_at"`
	DurationMS   int64     `json:"duration_ms"`
	ErrorCode    *string   `json:"error_code,omitempty"`
	ErrorMessage *string   `json:"error_message,omitempty"`
	CollectorID  string    `json:"collector_id,omitempty"`
	JobID        string    `json:"job_id,omitempty"`
}

type SiteCollectStatus struct {
	SiteID      int64                             `json:"site_id"`
	Summary     summarymodels.SiteSummaryResponse `json:"summary"`
	Targets     []summarymodels.TargetSummaryItem `json:"targets"`
	GeneratedAt time.Time                         `json:"generated_at"`
}

type TargetCollectStatus struct {
	SiteID      int64                               `json:"site_id"`
	Target      string                              `json:"target"`
	Summary     summarymodels.TargetSummaryResponse `json:"summary"`
	LatestCore  readmodels.TargetLatestResponse     `json:"latest_core"`
	LatestLight readmodels.TargetLatestResponse     `json:"latest_light"`
	Trend       readmodels.TargetTrendResponse      `json:"trend"`
	Changes     readmodels.TargetChangesResponse    `json:"changes"`
	GeneratedAt time.Time                           `json:"generated_at"`
}
