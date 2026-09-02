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

type DimensionSlice struct {
	Value       string   `json:"value"`
	Label       *string  `json:"label"`
	LabelEn     *string  `json:"label_en"`
	Population  int64    `json:"population"`
	Eligible    int64    `json:"eligible"`
	Known       int64    `json:"known"`
	MetricValue *float64 `json:"metric_value"`
	Coverage    *float64 `json:"coverage"`
}

type DimensionBreakdown struct {
	Key       string           `json:"key"`
	Dimension string           `json:"dimension"`
	SliceMode string           `json:"slice_mode"`
	AsOf      *string          `json:"as_of"`
	Items     []DimensionSlice `json:"items"`
}

type DimensionSliceRef struct {
	Value   string  `json:"value"`
	Label   *string `json:"label"`
	LabelEn *string `json:"label_en"`
}

type DimensionTrendPoint struct {
	Date        string   `json:"date"`
	Population  int64    `json:"population"`
	Eligible    int64    `json:"eligible"`
	Known       int64    `json:"known"`
	MetricValue *float64 `json:"metric_value"`
	Coverage    *float64 `json:"coverage"`
}

type DimensionTrend struct {
	Key              string                `json:"key"`
	Dimension        string                `json:"dimension"`
	Slice            DimensionSliceRef     `json:"slice"`
	SliceMode        string                `json:"slice_mode"`
	RequestedRange   string                `json:"requested_range"`
	AsOf             *string               `json:"as_of"`
	AvailableFrom    *string               `json:"available_from"`
	AvailableThrough *string               `json:"available_through"`
	Points           []DimensionTrendPoint `json:"points"`
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

type ExplorerChange struct {
	Domain     string     `json:"domain"`
	Category   string     `json:"category"`
	Type       string     `json:"type"`
	Date       string     `json:"date"`
	OccurredAt *time.Time `json:"occurred_at"`
	Entity     EntityRef  `json:"entity"`
	Detail     any        `json:"detail"`
}

type ChangeExplorerPage struct {
	Items      []ExplorerChange `json:"items"`
	NextCursor *string          `json:"next_cursor"`
}

type ChangeExplorerQuery struct {
	Range    string
	Category string
	Type     string
	Cursor   string
	Limit    int32
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

type SiteCompareCapability struct {
	Key   string `json:"key"`
	State string `json:"state"`
}

type SiteCompareCertificate struct {
	Target            string     `json:"target"`
	NotAfter          *time.Time `json:"not_after"`
	DaysToExpiry      *int32     `json:"days_to_expiry"`
	ExpiryStatus      *string    `json:"expiry_status"`
	Verified          *bool      `json:"verified"`
	VerificationIssue *string    `json:"verification_issue"`
	Issuer            *string    `json:"issuer"`
	ObservedAt        *time.Time `json:"observed_at"`
}

type SiteCompareItem struct {
	Site         EntityRef               `json:"site"`
	Capabilities []SiteCompareCapability `json:"capabilities"`
	Certificate  *SiteCompareCertificate `json:"certificate"`
}

type SiteCompare struct {
	Status string            `json:"status"`
	AsOf   *string           `json:"as_of"`
	Sites  []SiteCompareItem `json:"sites"`
}

type CertificateVerificationSummary struct {
	Known    int64    `json:"known"`
	Verified int64    `json:"verified"`
	Failed   int64    `json:"failed"`
	Coverage *float64 `json:"coverage"`
}

type CertificateQualitySummary struct {
	NotApplicable int64 `json:"not_applicable"`
	Stale         int64 `json:"stale"`
	NotProbed     int64 `json:"not_probed"`
	ProbeFailed   int64 `json:"probe_failed"`
	Unknown       int64 `json:"unknown"`
}

type CertificateExpirySummary struct {
	Known           int64    `json:"known"`
	Coverage        *float64 `json:"coverage"`
	Expired         int64    `json:"expired"`
	ExpiresWithin7D int64    `json:"expires_within_7d"`
	ExpiresIn8To30D int64    `json:"expires_in_8_30d"`
	Later           int64    `json:"later"`
}

type CertificateItem struct {
	Site              EntityRef  `json:"site"`
	Target            string     `json:"target"`
	NotAfter          *time.Time `json:"not_after"`
	DaysToExpiry      *int32     `json:"days_to_expiry"`
	ExpiryStatus      *string    `json:"expiry_status"`
	Verified          *bool      `json:"verified"`
	VerificationIssue *string    `json:"verification_issue"`
	Issuer            *string    `json:"issuer"`
	ObservedAt        *time.Time `json:"observed_at"`
}

type CertificateOverview struct {
	AsOf               *string                        `json:"as_of"`
	ReferenceAt        *time.Time                     `json:"reference_at"`
	FreshnessSeconds   int64                          `json:"freshness_seconds"`
	Population         int64                          `json:"population"`
	Eligible           int64                          `json:"eligible"`
	Verification       CertificateVerificationSummary `json:"verification"`
	Quality            CertificateQualitySummary      `json:"quality"`
	Expiry             CertificateExpirySummary       `json:"expiry"`
	ExpiryAttention    []CertificateItem              `json:"expiry_attention"`
	VerificationIssues []CertificateItem              `json:"verification_issues"`
}

type MetricContract struct {
	PublicKey   string
	InternalKey string
	Version     int32
}

type DimensionContract struct {
	PublicKey   string
	InternalKey string
	SliceMode   string
}

type DimensionRecord struct {
	Value         string
	Label         *string
	LabelEn       *string
	Population    int64
	Eligible      int64
	PositiveCount int64
	NegativeCount int64
}

type DimensionAvailabilityRecord struct {
	Label            *string
	LabelEn          *string
	AvailableFrom    *time.Time
	AvailableThrough *time.Time
}

type DimensionTrendRecord struct {
	FactDate      time.Time
	Population    int64
	Eligible      int64
	PositiveCount int64
	NegativeCount int64
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

type SiteCompareCapabilityRecord struct {
	SiteID        int64
	MetricKey     string
	MetricVersion int32
	State         string
}

type CertificateOverviewRecord struct {
	FactDate         time.Time
	ReferenceAt      time.Time
	FreshnessSeconds int64
	Population       int64
	Eligible         int64
	Verified         int64
	Failed           int64
	NotApplicable    int64
	Stale            int64
	NotProbed        int64
	ProbeFailed      int64
	Unknown          int64
	Expired          int64
	ExpiresWithin7D  int64
	ExpiresIn8To30D  int64
	Later            int64
}

type CertificateItemRecord struct {
	SiteID            int64
	SiteName          string
	Target            string
	NotAfter          *time.Time
	Verified          *bool
	VerificationIssue *string
	Issuer            *string
	ObservedAt        *time.Time
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
	PrecisionRank   int32
	EventSortAt     time.Time
	OpaqueTie       string
}

type ChangeExplorerPosition struct {
	ProjectionDate time.Time
	PrecisionRank  int32
	EventSortAt    time.Time
	OpaqueTie      string
}

type ChangeExplorerConditions struct {
	DetectorKeys []string
	ContractIDs  []string
	RangeThrough time.Time
	RangeDays    int32
	Position     *ChangeExplorerPosition
	Limit        int32
}
