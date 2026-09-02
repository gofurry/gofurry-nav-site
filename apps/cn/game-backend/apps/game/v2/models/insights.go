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
	Mac     *bool   `json:"mac"`
	Linux   *bool   `json:"linux"`
	Release *string `json:"release"`
	AsOf    *string `json:"as_of"`
}

type InsightPlayerSummary struct {
	Current              *int64     `json:"current"`
	Peak30D              *int64     `json:"peak_30d"`
	Average30D           *float64   `json:"average_30d"`
	AsOf                 *time.Time `json:"as_of"`
	FactThrough          *string    `json:"fact_through"`
	EligibleFrom30D      *string    `json:"eligible_from_30d"`
	ObservedDays30D      int64      `json:"observed_days_30d"`
	SuccessfulSamples30D int64      `json:"successful_samples_30d"`
	SampleCoverage30D    *float64   `json:"sample_coverage_30d"`
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

type InsightObservedLow struct {
	Amount          int64  `json:"amount"`
	Currency        string `json:"currency"`
	FirstSeen       string `json:"first_seen"`
	ObservedSince   string `json:"observed_since"`
	InitialAmount   int64  `json:"initial_amount"`
	DiscountPercent int32  `json:"discount_percent"`
}

type InsightRegionalPrice struct {
	Region          string              `json:"region"`
	Available       bool                `json:"available"`
	State           *string             `json:"state"`
	Currency        *string             `json:"currency"`
	InitialAmount   *int64              `json:"initial_amount"`
	FinalAmount     *int64              `json:"final_amount"`
	DiscountPercent *int32              `json:"discount_percent"`
	ObservedLow     *InsightObservedLow `json:"observed_low"`
}

type InsightRegionalPrices struct {
	AsOf    *string                `json:"as_of"`
	Regions []InsightRegionalPrice `json:"regions"`
}

type GameInsights struct {
	Game           InsightEntityRef      `json:"game"`
	State          InsightGameState      `json:"state"`
	Players        InsightPlayerSummary  `json:"players"`
	Price          *InsightPrice         `json:"price"`
	RegionalPrices InsightRegionalPrices `json:"regional_prices"`
	RecentChanges  []InsightChange       `json:"recent_changes"`
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
	Region           string              `json:"region"`
	RequestedRange   string              `json:"requested_range"`
	AvailableFrom    *string             `json:"available_from"`
	AvailableThrough *string             `json:"available_through"`
	Points           []InsightPricePoint `json:"points"`
}

type InsightPlayerRankingItem struct {
	Rank              int32            `json:"rank"`
	Game              InsightEntityRef `json:"game"`
	Value             float64          `json:"value"`
	ObservedAt        *time.Time       `json:"observed_at"`
	EligibleFrom      *string          `json:"eligible_from"`
	ObservedDays      *int64           `json:"observed_days"`
	SuccessfulSamples *int64           `json:"successful_samples"`
	SampleCoverage    *float64         `json:"sample_coverage"`
}

type InsightPlayerRanking struct {
	Metric                 string                     `json:"metric"`
	Basis                  string                     `json:"basis"`
	SnapshotScheduledFor   *time.Time                 `json:"snapshot_scheduled_for"`
	LatestSlotScheduledFor *time.Time                 `json:"latest_slot_scheduled_for"`
	ObservedFrom           *time.Time                 `json:"observed_from"`
	ObservedThrough        *time.Time                 `json:"observed_through"`
	WindowFrom             *string                    `json:"window_from"`
	WindowThrough          *string                    `json:"window_through"`
	Population             int64                      `json:"population"`
	Ranked                 int64                      `json:"ranked"`
	EntityCoverage         *float64                   `json:"entity_coverage"`
	Items                  []InsightPlayerRankingItem `json:"items"`
}

type InsightPlayerRankingQuery struct {
	Metric string
	Limit  int32
}

type InsightPriceOverview struct {
	Region          string   `json:"region"`
	AsOf            *string  `json:"as_of"`
	Population      int64    `json:"population"`
	Priced          int64    `json:"priced"`
	Free            int64    `json:"free"`
	Unpriced        int64    `json:"unpriced"`
	Unknown         int64    `json:"unknown"`
	Unavailable     int64    `json:"unavailable"`
	Known           int64    `json:"known"`
	Coverage        *float64 `json:"coverage"`
	Discounted      int64    `json:"discounted"`
	DiscountedShare *float64 `json:"discounted_share"`
}

type InsightDiscountItem struct {
	Game            InsightEntityRef    `json:"game"`
	Currency        string              `json:"currency"`
	InitialAmount   int64               `json:"initial_amount"`
	FinalAmount     int64               `json:"final_amount"`
	DiscountPercent int32               `json:"discount_percent"`
	ObservedLow     *InsightObservedLow `json:"observed_low"`
}
type InsightDiscounts struct {
	Region string                `json:"region"`
	AsOf   *string               `json:"as_of"`
	Items  []InsightDiscountItem `json:"items"`
}

type InsightLanguageItem struct {
	Code                   string   `json:"code"`
	SteamName              string   `json:"steam_name"`
	SupportedGames         int64    `json:"supported_games"`
	Share                  *float64 `json:"share"`
	ExplicitFullAudioGames int64    `json:"explicit_full_audio_games"`
	ExplicitFullAudioShare *float64 `json:"explicit_full_audio_share"`
}
type InsightLanguageOverview struct {
	AsOf                  *string               `json:"as_of"`
	FreshnessSeconds      int64                 `json:"freshness_seconds"`
	Population            int64                 `json:"population"`
	Fresh                 int64                 `json:"fresh"`
	Stale                 int64                 `json:"stale"`
	Unobserved            int64                 `json:"unobserved"`
	Coverage              *float64              `json:"coverage"`
	FullyNormalizedGames  int64                 `json:"fully_normalized_games"`
	UnmappedGames         int64                 `json:"unmapped_games"`
	UnmappedEntries       int64                 `json:"unmapped_entries"`
	NormalizationCoverage *float64              `json:"normalization_coverage"`
	Items                 []InsightLanguageItem `json:"items"`
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
	Mac              *bool
	Linux            *bool
	Release          *string
}

type InsightPlayerSummaryRecord struct {
	HasCurrent        bool
	Current           int64
	CurrentAt         *time.Time
	Peak30D           *int64
	Average30D        *float64
	FactThrough       *time.Time
	EligibleFrom      *time.Time
	ObservedDays      int64
	SuccessfulSamples int64
	SampleCoverage    *float64
}

type InsightPriceRecord struct {
	FactDate        time.Time
	State           string
	Currency        *string
	InitialAmount   *int64
	FinalAmount     *int64
	DiscountPercent *int32
}

type InsightRegionalPriceRecord struct {
	Region          string
	Available       bool
	FactDate        time.Time
	State           *string
	Currency        *string
	InitialAmount   *int64
	FinalAmount     *int64
	DiscountPercent *int32
}
type InsightObservedLowRecord struct {
	Amount          int64
	Currency        string
	FirstSeen       time.Time
	ObservedSince   time.Time
	InitialAmount   int64
	DiscountPercent int32
}
type InsightPlayerRankingMetaRecord struct {
	SnapshotScheduledFor   *time.Time
	LatestSlotScheduledFor *time.Time
	ObservedFrom           *time.Time
	ObservedThrough        *time.Time
	WindowFrom             *time.Time
	WindowThrough          *time.Time
	Population             int64
	Ranked                 int64
}
type InsightPlayerRankingRecord struct {
	GameID            int64
	GameName          string
	Value             float64
	ObservedAt        *time.Time
	EligibleFrom      *time.Time
	ObservedDays      *int64
	SuccessfulSamples *int64
	SampleCoverage    *float64
}
type InsightPriceOverviewRecord struct {
	AsOf                                                                 *time.Time
	Population, Priced, Free, Unpriced, Unknown, Unavailable, Discounted int64
}
type InsightDiscountRecord struct {
	AsOf                       time.Time
	GameID, TrackingPeriodID   int64
	GameName, Currency         string
	InitialAmount, FinalAmount int64
	DiscountPercent            int32
}
type InsightLanguageOverviewRecord struct {
	AsOf                                                                                       *time.Time
	Population, Fresh, Stale, Unobserved, FullyNormalizedGames, UnmappedGames, UnmappedEntries int64
}
type InsightLanguageRecord struct {
	Code, SteamName                        string
	SupportedGames, ExplicitFullAudioGames int64
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
