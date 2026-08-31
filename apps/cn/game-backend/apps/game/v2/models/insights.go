package models

import "time"

type InsightMetric struct {
	Key           string   `json:"key"`
	AsOf          string   `json:"as_of"`
	Value         *float64 `json:"value"`
	Delta30D      *float64 `json:"delta_30d"`
	Coverage      *float64 `json:"coverage"`
	Known         int64    `json:"known"`
	Eligible      int64    `json:"eligible"`
	AvailableFrom *string  `json:"available_from"`
}

type InsightTrendPoint struct {
	Date     string   `json:"date"`
	Value    *float64 `json:"value"`
	Coverage *float64 `json:"coverage"`
}

type InsightMetricTrend struct {
	Key              string              `json:"key"`
	RequestedRange   string              `json:"requested_range"`
	AvailableFrom    *string             `json:"available_from"`
	AvailableThrough *string             `json:"available_through"`
	Points           []InsightTrendPoint `json:"points"`
}

type InsightEntityRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type InsightChange struct {
	Type       string           `json:"type"`
	Date       string           `json:"date"`
	OccurredAt *time.Time       `json:"occurred_at"`
	Entity     InsightEntityRef `json:"entity"`
	Detail     any              `json:"detail"`
}

type InsightOverview struct {
	GeneratedAt   time.Time       `json:"generated_at"`
	EntityCount   int64           `json:"entity_count"`
	Changes7D     int64           `json:"changes_7d"`
	Metrics       []InsightMetric `json:"metrics"`
	RecentChanges []InsightChange `json:"recent_changes"`
}

type InsightGameState struct {
	Free    *bool   `json:"free"`
	Windows *bool   `json:"windows"`
	Linux   *bool   `json:"linux"`
	Release *string `json:"release"`
	AsOf    *string `json:"as_of"`
}

type InsightPlayerSummary struct {
	Current *int64     `json:"current"`
	Peak30D *int64     `json:"peak_30d"`
	AsOf    *time.Time `json:"as_of"`
}

type InsightPrice struct {
	Region          string  `json:"region"`
	State           string  `json:"state"`
	Currency        *string `json:"currency"`
	InitialAmount   *int64  `json:"initial_amount"`
	FinalAmount     *int64  `json:"final_amount"`
	DiscountPercent *int32  `json:"discount_percent"`
	AsOf            string  `json:"as_of"`
}

type GameInsights struct {
	Game          InsightEntityRef     `json:"game"`
	State         InsightGameState     `json:"state"`
	Players       InsightPlayerSummary `json:"players"`
	Price         *InsightPrice        `json:"price"`
	RecentChanges []InsightChange      `json:"recent_changes"`
}

type InsightPlayerPoint struct {
	Date string   `json:"date"`
	Min  *int64   `json:"min"`
	Max  int64    `json:"max"`
	Avg  *float64 `json:"avg"`
}

type InsightPlayerHistory struct {
	RequestedRange   string               `json:"requested_range"`
	AvailableFrom    *string              `json:"available_from"`
	AvailableThrough *string              `json:"available_through"`
	Points           []InsightPlayerPoint `json:"points"`
}

type InsightPricePoint struct {
	Date            string  `json:"date"`
	State           string  `json:"state"`
	Currency        *string `json:"currency"`
	InitialAmount   *int64  `json:"initial_amount"`
	FinalAmount     *int64  `json:"final_amount"`
	DiscountPercent *int32  `json:"discount_percent"`
}

type InsightPriceHistory struct {
	RequestedRange   string              `json:"requested_range"`
	AvailableFrom    *string             `json:"available_from"`
	AvailableThrough *string             `json:"available_through"`
	Points           []InsightPricePoint `json:"points"`
}

type InsightMetricContract struct {
	PublicKey   string
	InternalKey string
	Version     int32
}

type InsightMetricSummaryRecord struct {
	FactDate              time.Time
	EligibleCount         int64
	PositiveCount         int64
	NegativeCount         int64
	PreviousPositiveCount *int64
	PreviousNegativeCount *int64
	AvailableFrom         *time.Time
}

type InsightMetricTrendRecord struct {
	FactDate      time.Time
	EligibleCount int64
	PositiveCount int64
	NegativeCount int64
}

type InsightGameRecord struct {
	ID   int64
	Name string
}

type InsightGameStateRecord struct {
	GameID           int64
	FactDate         time.Time
	TrackingPeriodID int64
	AppID            int64
	Free             *bool
	Windows          *bool
	Linux            *bool
	Release          *string
}

type InsightPlayerSummaryRecord struct {
	HasCurrent bool
	Current    int64
	CurrentAt  *time.Time
	HasPeak30D bool
	Peak30D    int64
}

type InsightPriceRecord struct {
	FactDate        time.Time
	State           string
	Currency        *string
	InitialAmount   *int64
	FinalAmount     *int64
	DiscountPercent *int32
}

type InsightPlayerPointRecord struct {
	FactDate time.Time
	Min      *int64
	Max      int64
	Avg      *float64
}

type InsightChangeRecord struct {
	EntityID        int64
	EntityName      string
	DetectorKey     string
	DetectorVersion int32
	EventCode       string
	ProjectionDate  time.Time
	TimeBasis       string
	EventAt         *time.Time
}
