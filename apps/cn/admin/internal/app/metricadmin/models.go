package metricadmin

import (
	"encoding/json"
	"time"
)

type Registry struct {
	Domain            string   `json:"domain"`
	MetricKey         string   `json:"metric_key"`
	MetricVersion     int32    `json:"metric_version"`
	MetricKind        string   `json:"metric_kind"`
	EntityLevel       string   `json:"entity_level"`
	TimeGrain         string   `json:"time_grain"`
	SourceFacts       []string `json:"source_facts"`
	EligibilityPolicy string   `json:"eligibility_policy"`
	StatePolicy       string   `json:"state_policy"`
	CoveragePolicy    string   `json:"coverage_policy"`
	FreshnessSeconds  *int64   `json:"freshness_seconds"`
	AllowedDimensions []string `json:"allowed_dimensions"`
	Status            string   `json:"status"`
	Description       string   `json:"description"`
	CreatedAt         string   `json:"created_at"`
	RetiredAt         *string  `json:"retired_at"`
}

type Checkpoint struct {
	Domain                   string  `json:"domain"`
	MetricKey                string  `json:"metric_key"`
	MetricVersion            int32   `json:"metric_version"`
	Status                   string  `json:"status"`
	SourceStartDate          *string `json:"source_start_date"`
	ProcessedThrough         *string `json:"processed_through"`
	UpstreamProcessedThrough *string `json:"upstream_processed_through"`
	LagDays                  *int    `json:"lag_days"`
	CreatedAt                string  `json:"created_at"`
	UpdatedAt                string  `json:"updated_at"`
}

type Overview struct {
	Domain                   string   `json:"domain"`
	MetricKey                string   `json:"metric_key"`
	MetricVersion            int32    `json:"metric_version"`
	Description              string   `json:"description"`
	ProcessedThrough         *string  `json:"processed_through"`
	UpstreamProcessedThrough *string  `json:"upstream_processed_through"`
	LagDays                  *int     `json:"lag_days"`
	LatestFactDate           *string  `json:"latest_fact_date"`
	PopulationCount          int64    `json:"population_count"`
	EligibleCount            int64    `json:"eligible_count"`
	NotApplicableCount       int64    `json:"not_applicable_count"`
	PositiveCount            int64    `json:"positive_count"`
	NegativeCount            int64    `json:"negative_count"`
	StaleCount               int64    `json:"stale_count"`
	NotProbedCount           int64    `json:"not_probed_count"`
	ProbeFailedCount         int64    `json:"probe_failed_count"`
	UnknownCount             int64    `json:"unknown_count"`
	AdoptionRate             *float64 `json:"adoption_rate"`
	CoverageRate             *float64 `json:"coverage_rate"`
}

type Daily struct {
	Domain             string   `json:"domain"`
	MetricKey          string   `json:"metric_key"`
	MetricVersion      int32    `json:"metric_version"`
	FactDate           string   `json:"fact_date"`
	DimensionKey       string   `json:"dimension_key"`
	DimensionValue     string   `json:"dimension_value"`
	PopulationCount    int64    `json:"population_count"`
	EligibleCount      int64    `json:"eligible_count"`
	NotApplicableCount int64    `json:"not_applicable_count"`
	PositiveCount      int64    `json:"positive_count"`
	NegativeCount      int64    `json:"negative_count"`
	StaleCount         int64    `json:"stale_count"`
	NotProbedCount     int64    `json:"not_probed_count"`
	ProbeFailedCount   int64    `json:"probe_failed_count"`
	UnknownCount       int64    `json:"unknown_count"`
	AdoptionRate       *float64 `json:"adoption_rate"`
	CoverageRate       *float64 `json:"coverage_rate"`
	ComputedAt         string   `json:"computed_at"`
}

type Entity struct {
	Domain                   string          `json:"domain"`
	EntityID                 int64           `json:"entity_id"`
	HistoricalName           string          `json:"historical_name"`
	State                    string          `json:"state"`
	ReasonCode               string          `json:"reason_code"`
	SourceObservedAt         *string         `json:"source_observed_at"`
	DimensionValues          json.RawMessage `json:"dimension_values"`
	SourceProjectionVersions json.RawMessage `json:"source_projection_versions"`
	EvaluatedAt              string          `json:"evaluated_at"`
}

type EntityPage struct {
	Total int64    `json:"total"`
	List  []Entity `json:"list"`
}

type Filters struct {
	Domain         string
	MetricKey      string
	MetricVersion  int32
	Status         string
	From           *time.Time
	Through        *time.Time
	DimensionKey   string
	DimensionValue string
	FactDate       time.Time
	State          string
	ReasonCode     string
	Page           int32
	PageSize       int32
}
