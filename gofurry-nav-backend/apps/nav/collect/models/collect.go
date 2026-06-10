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
	Protocol string `gorm:"column:protocol" json:"protocol"`
	Status   string `gorm:"column:status" json:"status"`
	Count    int64  `gorm:"column:count" json:"count"`
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
	ID           int64     `gorm:"column:id" json:"id"`
	SiteID       int64     `gorm:"column:site_id" json:"site_id"`
	Target       string    `gorm:"column:target" json:"target"`
	Protocol     string    `gorm:"column:protocol" json:"protocol"`
	Status       string    `gorm:"column:status" json:"status"`
	ObservedAt   time.Time `gorm:"column:observed_at" json:"observed_at"`
	DurationMS   int64     `gorm:"column:duration_ms" json:"duration_ms"`
	ErrorCode    *string   `gorm:"column:error_code" json:"error_code,omitempty"`
	ErrorMessage *string   `gorm:"column:error_message" json:"error_message,omitempty"`
	CollectorID  string    `gorm:"column:collector_id" json:"collector_id,omitempty"`
	JobID        string    `gorm:"column:job_id" json:"job_id,omitempty"`
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
