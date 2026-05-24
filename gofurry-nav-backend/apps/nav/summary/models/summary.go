package models

import "time"

const (
	SummaryStateReady   = "ready"
	SummaryStateMissing = "missing"
	SummaryStateStale   = "stale"

	StatusHealthy  = "healthy"
	StatusWarning  = "warning"
	StatusDegraded = "degraded"
	StatusUnknown  = "unknown"
	StatusDown     = "down"
)

type ProtocolSummary struct {
	Protocol          string    `json:"protocol"`
	Status            string    `json:"status"`
	ObservedAt        time.Time `json:"observed_at"`
	DurationMS        int64     `json:"duration_ms"`
	Stale             bool      `json:"stale"`
	StaleAfterSeconds int64     `json:"stale_after_seconds"`
	ErrorCode         string    `json:"error_code,omitempty"`
}

type TargetSummaryItem struct {
	Target         string    `json:"target"`
	Status         string    `json:"status"`
	ReasonCodes    []string  `json:"reason_codes"`
	ReasonMessages []string  `json:"reason_messages"`
	ObservedAt     time.Time `json:"observed_at"`
}

type TargetSummaryResponse struct {
	State          string                     `json:"state"`
	SiteID         int64                      `json:"site_id"`
	Target         string                     `json:"target"`
	Status         string                     `json:"status"`
	ReasonCodes    []string                   `json:"reason_codes"`
	ReasonMessages []string                   `json:"reason_messages"`
	Protocols      map[string]ProtocolSummary `json:"protocols"`
	ObservedAt     time.Time                  `json:"observed_at"`
	GeneratedAt    time.Time                  `json:"generated_at"`
	SchemaVersion  int                        `json:"schema_version"`
}

type SiteSummaryResponse struct {
	State          string              `json:"state"`
	SiteID         int64               `json:"site_id"`
	Status         string              `json:"status"`
	ReasonCodes    []string            `json:"reason_codes"`
	ReasonMessages []string            `json:"reason_messages"`
	TargetCount    int                 `json:"target_count"`
	StatusCounts   map[string]int      `json:"status_counts"`
	Targets        []TargetSummaryItem `json:"targets"`
	GeneratedAt    time.Time           `json:"generated_at"`
	SchemaVersion  int                 `json:"schema_version"`
}
