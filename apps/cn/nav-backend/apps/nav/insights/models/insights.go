package models

import "time"

type Metric struct {
	Key           string   `json:"key"`
	AsOf          string   `json:"as_of"`
	Value         *float64 `json:"value"`
	Delta30D      *float64 `json:"delta_30d"`
	Coverage      *float64 `json:"coverage"`
	Known         int64    `json:"known"`
	Eligible      int64    `json:"eligible"`
	AvailableFrom *string  `json:"available_from"`
}

type TrendPoint struct {
	Date     string   `json:"date"`
	Value    *float64 `json:"value"`
	Coverage *float64 `json:"coverage"`
}

type MetricTrend struct {
	Key              string       `json:"key"`
	RequestedRange   string       `json:"requested_range"`
	AvailableFrom    *string      `json:"available_from"`
	AvailableThrough *string      `json:"available_through"`
	Points           []TrendPoint `json:"points"`
}

type EntityRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Change struct {
	Type       string     `json:"type"`
	Date       string     `json:"date"`
	OccurredAt *time.Time `json:"occurred_at"`
	Entity     EntityRef  `json:"entity"`
	Detail     any        `json:"detail"`
}

type Overview struct {
	GeneratedAt   time.Time `json:"generated_at"`
	EntityCount   int64     `json:"entity_count"`
	Changes7D     int64     `json:"changes_7d"`
	Metrics       []Metric  `json:"metrics"`
	RecentChanges []Change  `json:"recent_changes"`
}

type Ecosystem struct {
	Value    *float64 `json:"value"`
	Coverage *float64 `json:"coverage"`
}

type Capability struct {
	Key       string    `json:"key"`
	AsOf      string    `json:"as_of"`
	State     string    `json:"state"`
	Ecosystem Ecosystem `json:"ecosystem"`
}

type SiteInsights struct {
	Site          EntityRef    `json:"site"`
	Capabilities  []Capability `json:"capabilities"`
	RecentChanges []Change     `json:"recent_changes"`
}

type MetricContract struct {
	PublicKey   string
	InternalKey string
	Version     int32
}

type MetricSummaryRecord struct {
	FactDate              time.Time
	EligibleCount         int64
	PositiveCount         int64
	NegativeCount         int64
	PreviousPositiveCount *int64
	PreviousNegativeCount *int64
	AvailableFrom         *time.Time
}

type MetricTrendRecord struct {
	FactDate      time.Time
	EligibleCount int64
	PositiveCount int64
	NegativeCount int64
}

type SiteRecord struct {
	ID   int64
	Name string
}

type SiteMetricRecord struct {
	FactDate      time.Time
	State         string
	EligibleCount int64
	PositiveCount int64
	NegativeCount int64
}

type ChangeRecord struct {
	EntityID        int64
	EntityName      string
	DetectorKey     string
	DetectorVersion int32
	EventCode       string
	ProjectionDate  time.Time
	TimeBasis       string
	EventAt         *time.Time
}
