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

type InsightDimensionSlice struct {
	Value       string   `json:"value"`
	Label       *string  `json:"label"`
	LabelEn     *string  `json:"label_en"`
	Population  int64    `json:"population"`
	Eligible    int64    `json:"eligible"`
	Known       int64    `json:"known"`
	MetricValue *float64 `json:"metric_value"`
	Coverage    *float64 `json:"coverage"`
}

type InsightDimensionBreakdown struct {
	Key       string                  `json:"key"`
	Dimension string                  `json:"dimension"`
	SliceMode string                  `json:"slice_mode"`
	AsOf      *string                 `json:"as_of"`
	Items     []InsightDimensionSlice `json:"items"`
}

type InsightDimensionSliceRef struct {
	Value   string  `json:"value"`
	Label   *string `json:"label"`
	LabelEn *string `json:"label_en"`
}

type InsightDimensionTrendPoint struct {
	Date        string   `json:"date"`
	Population  int64    `json:"population"`
	Eligible    int64    `json:"eligible"`
	Known       int64    `json:"known"`
	MetricValue *float64 `json:"metric_value"`
	Coverage    *float64 `json:"coverage"`
}

type InsightDimensionTrend struct {
	Key              string                       `json:"key"`
	Dimension        string                       `json:"dimension"`
	Slice            InsightDimensionSliceRef     `json:"slice"`
	SliceMode        string                       `json:"slice_mode"`
	RequestedRange   string                       `json:"requested_range"`
	AsOf             *string                      `json:"as_of"`
	AvailableFrom    *string                      `json:"available_from"`
	AvailableThrough *string                      `json:"available_through"`
	Points           []InsightDimensionTrendPoint `json:"points"`
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

type InsightExplorerChange struct {
	Domain     string           `json:"domain"`
	Category   string           `json:"category"`
	Type       string           `json:"type"`
	Date       string           `json:"date"`
	OccurredAt *time.Time       `json:"occurred_at"`
	Entity     InsightEntityRef `json:"entity"`
	Detail     any              `json:"detail"`
}

type InsightChangeExplorerPage struct {
	Items      []InsightExplorerChange `json:"items"`
	NextCursor *string                 `json:"next_cursor"`
}

type InsightChangeExplorerQuery struct {
	Range    string
	Category string
	Type     string
	Cursor   string
	Limit    int32
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

type InsightDimensionContract struct {
	PublicKey   string
	InternalKey string
	SliceMode   string
}

type InsightDimensionRecord struct {
	Value         string
	Label         *string
	LabelEn       *string
	Population    int64
	Eligible      int64
	PositiveCount int64
	NegativeCount int64
}

type InsightDimensionAvailabilityRecord struct {
	Label            *string
	LabelEn          *string
	AvailableFrom    *time.Time
	AvailableThrough *time.Time
}

type InsightDimensionTrendRecord struct {
	FactDate      time.Time
	Population    int64
	Eligible      int64
	PositiveCount int64
	NegativeCount int64
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
	PrecisionRank   int32
	EventSortAt     time.Time
	OpaqueTie       string
}

type InsightChangeExplorerPosition struct {
	ProjectionDate time.Time
	PrecisionRank  int32
	EventSortAt    time.Time
	OpaqueTie      string
}

type InsightChangeExplorerConditions struct {
	DetectorKeys []string
	ContractIDs  []string
	RangeThrough time.Time
	RangeDays    int32
	Position     *InsightChangeExplorerPosition
	Limit        int32
}
