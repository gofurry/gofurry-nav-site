package models

import (
	"encoding/json"
	"time"
)

const (
	SummaryStateReady   = "ready"
	SummaryStateMissing = "missing"

	ProtocolPing        = "ping"
	ProtocolHTTP        = "http"
	ProtocolDNS         = "dns"
	ProtocolRDAP        = "rdap"
	ProtocolRobots      = "robots"
	ProtocolSecurityTXT = "security_txt"
	ProtocolPageAssets  = "page_assets"
	ProtocolPortCheck   = "port_check"
	ProtocolWAFCanary   = "waf_canary"

	TableNameGfnCollectorObservation = "gfn_collector_observation"

	DefaultObservationLimit = 100
	MaxObservationLimit     = 500
)

type GfnCollectorObservation struct {
	ID            int64     `gorm:"column:id;type:bigint;primaryKey" json:"id"`
	SiteID        int64     `gorm:"column:site_id;type:bigint;not null" json:"site_id"`
	Target        string    `gorm:"column:target;type:character varying(255);not null" json:"target"`
	Protocol      string    `gorm:"column:protocol;type:character varying(16);not null" json:"protocol"`
	Status        string    `gorm:"column:status;type:character varying(32);not null" json:"status"`
	ObservedAt    time.Time `gorm:"column:observed_at;type:timestamptz;not null" json:"observed_at"`
	DurationMS    int64     `gorm:"column:duration_ms;type:bigint" json:"duration_ms"`
	ErrorCode     *string   `gorm:"column:error_code;type:character varying(64)" json:"error_code,omitempty"`
	ErrorMessage  *string   `gorm:"column:error_message;type:text" json:"error_message,omitempty"`
	Payload       string    `gorm:"column:payload;type:jsonb;not null" json:"payload"`
	SchemaVersion int       `gorm:"column:schema_version;type:int;not null" json:"schema_version"`
	CreateTime    time.Time `gorm:"column:create_time;type:timestamptz;not null;autoCreateTime" json:"create_time"`
}

func (*GfnCollectorObservation) TableName() string {
	return TableNameGfnCollectorObservation
}

type CollectorEnvelope struct {
	SiteID                  int64           `json:"site_id"`
	Target                  string          `json:"target"`
	Protocol                string          `json:"protocol"`
	Status                  string          `json:"status"`
	ObservedAt              time.Time       `json:"observed_at"`
	DurationMS              int64           `json:"duration_ms"`
	ErrorCode               string          `json:"error_code,omitempty"`
	ErrorMessage            string          `json:"error_message,omitempty"`
	Payload                 json.RawMessage `json:"payload"`
	PayloadBytes            int             `json:"payload_bytes,omitempty"`
	PayloadTruncated        bool            `json:"payload_truncated,omitempty"`
	PayloadPreviewMaxBytes  int             `json:"payload_preview_max_bytes,omitempty"`
	PayloadPreviewAvailable bool            `json:"payload_preview_available,omitempty"`
	SchemaVersion           int             `json:"schema_version"`
	CollectorID             string          `json:"collector_id,omitempty"`
	JobID                   string          `json:"job_id,omitempty"`
}

type TargetLatestResponse struct {
	State     string                       `json:"state"`
	SiteID    int64                        `json:"site_id"`
	Target    string                       `json:"target"`
	Protocols map[string]CollectorEnvelope `json:"protocols"`
}

type ObservationsResponse struct {
	State    string              `json:"state"`
	SiteID   int64               `json:"site_id"`
	Target   string              `json:"target"`
	Protocol string              `json:"protocol"`
	Limit    int                 `json:"limit"`
	Items    []CollectorEnvelope `json:"items"`
}

type TargetTrendResponse struct {
	State         string          `json:"state"`
	SiteID        int64           `json:"site_id"`
	Target        string          `json:"target"`
	Windows       json.RawMessage `json:"windows"`
	GeneratedAt   time.Time       `json:"generated_at"`
	SchemaVersion int             `json:"schema_version"`
}

type TargetChangesResponse struct {
	State         string          `json:"state"`
	SiteID        int64           `json:"site_id"`
	Target        string          `json:"target"`
	Events        json.RawMessage `json:"events"`
	GeneratedAt   time.Time       `json:"generated_at"`
	SchemaVersion int             `json:"schema_version"`
}

type RunStateResponse struct {
	State        string    `json:"state"`
	CollectorID  string    `json:"collector_id,omitempty"`
	JobID        string    `json:"job_id,omitempty"`
	Protocol     string    `json:"protocol"`
	Status       string    `json:"status,omitempty"`
	StartedAt    time.Time `json:"started_at,omitempty"`
	FinishedAt   time.Time `json:"finished_at,omitempty"`
	DurationMS   int64     `json:"duration_ms"`
	TargetCount  int64     `json:"target_count"`
	SuccessCount int64     `json:"success_count"`
	FailureCount int64     `json:"failure_count"`
	SkippedCount int64     `json:"skipped_count"`
	ErrorCount   int64     `json:"error_count"`
	SkipReason   string    `json:"skip_reason,omitempty"`
}

func CoreProtocols() []string {
	return []string{ProtocolPing, ProtocolHTTP, ProtocolDNS}
}

func LightProbeProtocols() []string {
	return []string{
		ProtocolRDAP,
		ProtocolRobots,
		ProtocolSecurityTXT,
		ProtocolPageAssets,
		ProtocolPortCheck,
		ProtocolWAFCanary,
	}
}

func AllProtocols() []string {
	values := CoreProtocols()
	values = append(values, LightProbeProtocols()...)
	return values
}

func IsProtocolAllowed(protocol string) bool {
	for _, value := range AllProtocols() {
		if protocol == value {
			return true
		}
	}
	return false
}

func NormalizeObservationLimit(limit int) int {
	if limit <= 0 {
		return DefaultObservationLimit
	}
	if limit > MaxObservationLimit {
		return MaxObservationLimit
	}
	return limit
}
