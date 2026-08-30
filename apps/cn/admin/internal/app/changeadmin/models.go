package changeadmin

import (
	"encoding/json"
	"time"
)

type Filters struct {
	Domain          string
	DetectorKey     string
	DetectorVersion int32
	Status          string
	From            *time.Time
	Through         *time.Time
	EventCode       string
	ScopeKind       string
	ScopeKey        string
	EntityID        int64
	Page            int32
	PageSize        int32
}

type Overview struct {
	Domain                   string  `json:"domain"`
	DetectorKey              string  `json:"detector_key"`
	DetectorVersion          int32   `json:"detector_version"`
	Status                   string  `json:"status"`
	Description              string  `json:"description"`
	WatermarkPolicy          string  `json:"watermark_policy"`
	SourceStartDate          *string `json:"source_start_date"`
	ProcessedThrough         *string `json:"processed_through"`
	UpstreamProcessedThrough *string `json:"upstream_processed_through"`
	LagDays                  *int    `json:"lag_days"`
	LatestProjectionDate     *string `json:"latest_projection_date"`
	LatestEventCount         int64   `json:"latest_event_count"`
	TotalEventCount          int64   `json:"total_event_count"`
}

type Registry struct {
	Domain          string   `json:"domain"`
	DetectorKey     string   `json:"detector_key"`
	DetectorVersion int32    `json:"detector_version"`
	SourceKind      string   `json:"source_kind"`
	SourceContracts []string `json:"source_contracts"`
	DetectionPolicy string   `json:"detection_policy"`
	WatermarkPolicy string   `json:"watermark_policy"`
	EventCodes      []string `json:"event_codes"`
	ProcessingGrain string   `json:"processing_grain"`
	Status          string   `json:"status"`
	Description     string   `json:"description"`
	CreatedAt       string   `json:"created_at"`
	RetiredAt       *string  `json:"retired_at"`
}

type Checkpoint struct {
	Domain                   string  `json:"domain"`
	DetectorKey              string  `json:"detector_key"`
	DetectorVersion          int32   `json:"detector_version"`
	Status                   string  `json:"status"`
	WatermarkPolicy          string  `json:"watermark_policy"`
	SourceStartDate          *string `json:"source_start_date"`
	ProcessedThrough         *string `json:"processed_through"`
	UpstreamProcessedThrough *string `json:"upstream_processed_through"`
	LagDays                  *int    `json:"lag_days"`
	CreatedAt                string  `json:"created_at"`
	UpdatedAt                string  `json:"updated_at"`
}

type Event struct {
	Domain          string          `json:"domain"`
	EventKey        string          `json:"event_key"`
	DetectorKey     string          `json:"detector_key"`
	DetectorVersion int32           `json:"detector_version"`
	EntityID        int64           `json:"entity_id"`
	HistoricalName  string          `json:"historical_name"`
	ProjectionDate  string          `json:"projection_date"`
	EventAt         *string         `json:"event_at"`
	TimeBasis       string          `json:"time_basis"`
	EventCode       string          `json:"event_code"`
	ScopeKind       string          `json:"scope_kind"`
	ScopeKey        string          `json:"scope_key"`
	OldValue        json.RawMessage `json:"old_value"`
	NewValue        json.RawMessage `json:"new_value"`
	SourceEventKey  string          `json:"source_event_key"`
	SourceBeforeKey string          `json:"source_before_key"`
	SourceAfterKey  string          `json:"source_after_key"`
	SourceBeforeAt  *string         `json:"source_before_at"`
	SourceAfterAt   *string         `json:"source_after_at"`
	SourceVersions  json.RawMessage `json:"source_versions"`
	MaterializedAt  string          `json:"materialized_at"`
}

type EventPage struct {
	List     []Event `json:"list"`
	Total    int64   `json:"total"`
	Page     int32   `json:"page"`
	PageSize int32   `json:"page_size"`
}
