package observation

import "time"

const (
	ProtocolPing        = "ping"
	ProtocolHTTP        = "http"
	ProtocolDNS         = "dns"
	ProtocolRDAP        = "rdap"
	ProtocolRobots      = "robots"
	ProtocolSecurityTXT = "security_txt"
	ProtocolPageAssets  = "page_assets"
	ProtocolPortCheck   = "port_check"
	ProtocolWAFCanary   = "waf_canary"

	StatusSuccess = "success"
	StatusFailure = "failure"

	TableNameGfnCollectorObservation = "gfn_collector_observation"
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

type Input struct {
	SiteID       int64
	Target       string
	Protocol     string
	Status       string
	ObservedAt   time.Time
	DurationMS   int64
	ErrorCode    string
	ErrorMessage string
	Payload      any
	CollectorID  string
	JobID        string
}

type LatestDocument struct {
	SiteID        int64     `json:"site_id"`
	Target        string    `json:"target"`
	Protocol      string    `json:"protocol"`
	Status        string    `json:"status"`
	ObservedAt    time.Time `json:"observed_at"`
	DurationMS    int64     `json:"duration_ms"`
	ErrorCode     string    `json:"error_code,omitempty"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	Payload       any       `json:"payload"`
	SchemaVersion int       `json:"schema_version"`
	CollectorID   string    `json:"collector_id,omitempty"`
	JobID         string    `json:"job_id,omitempty"`
}

type ProtocolSummary struct {
	Protocol          string    `json:"protocol"`
	Status            string    `json:"status"`
	ObservedAt        time.Time `json:"observed_at"`
	DurationMS        int64     `json:"duration_ms"`
	Stale             bool      `json:"stale"`
	StaleAfterSeconds int64     `json:"stale_after_seconds"`
	ErrorCode         string    `json:"error_code,omitempty"`
}

type TargetSummaryDocument struct {
	SiteID            int64                      `json:"site_id"`
	Target            string                     `json:"target"`
	Status            string                     `json:"status"`
	ReasonCodes       []string                   `json:"reason_codes"`
	ReasonMessages    []string                   `json:"reason_messages"`
	Protocols         map[string]ProtocolSummary `json:"protocols"`
	CanonicalTarget   *CanonicalTargetHint       `json:"canonical_target_hint,omitempty"`
	TargetRelations   []TargetRelationHint       `json:"target_relation_hints,omitempty"`
	EdgeProviderHints []EdgeProviderHint         `json:"edge_provider_hints,omitempty"`
	ObservedAt        time.Time                  `json:"observed_at"`
	GeneratedAt       time.Time                  `json:"generated_at"`
	SchemaVersion     int                        `json:"schema_version"`
}

type SiteSummaryDocument struct {
	SiteID          int64                    `json:"site_id"`
	Status          string                   `json:"status"`
	ReasonCodes     []string                 `json:"reason_codes"`
	ReasonMessages  []string                 `json:"reason_messages"`
	TargetCount     int                      `json:"target_count"`
	StatusCounts    map[string]int           `json:"status_counts"`
	Targets         []TargetSummaryItem      `json:"targets"`
	TargetRelations []SiteTargetRelationHint `json:"target_relation_hints,omitempty"`
	GeneratedAt     time.Time                `json:"generated_at"`
	SchemaVersion   int                      `json:"schema_version"`
}

type TargetSummaryItem struct {
	Target            string               `json:"target"`
	Status            string               `json:"status"`
	ReasonCodes       []string             `json:"reason_codes"`
	ReasonMessages    []string             `json:"reason_messages"`
	CanonicalTarget   *CanonicalTargetHint `json:"canonical_target_hint,omitempty"`
	TargetRelations   []TargetRelationHint `json:"target_relation_hints,omitempty"`
	EdgeProviderHints []EdgeProviderHint   `json:"edge_provider_hints,omitempty"`
	ObservedAt        time.Time            `json:"observed_at"`
}

type CanonicalTargetHint struct {
	TargetHost    string `json:"target_host"`
	FinalHost     string `json:"final_host,omitempty"`
	CanonicalHost string `json:"canonical_host,omitempty"`
	PreferredHost string `json:"preferred_host,omitempty"`
	Relation      string `json:"relation"`
	Source        string `json:"source"`
	FinalURL      string `json:"final_url,omitempty"`
	CanonicalURL  string `json:"canonical_url,omitempty"`
}

type TargetRelationHint struct {
	Relation    string `json:"relation"`
	Source      string `json:"source"`
	TargetHost  string `json:"target_host"`
	RelatedHost string `json:"related_host,omitempty"`
	Value       string `json:"value,omitempty"`
}

type SiteTargetRelationHint struct {
	Relation string   `json:"relation"`
	Host     string   `json:"host"`
	Targets  []string `json:"targets"`
}

type TargetTrendDocument struct {
	SiteID        int64                  `json:"site_id"`
	Target        string                 `json:"target"`
	Windows       map[string]TrendWindow `json:"windows"`
	GeneratedAt   time.Time              `json:"generated_at"`
	SchemaVersion int                    `json:"schema_version"`
}

type TrendWindow struct {
	Since     time.Time                `json:"since"`
	Until     time.Time                `json:"until"`
	Protocols map[string]ProtocolTrend `json:"protocols"`
}

type ProtocolTrend struct {
	Protocol         string     `json:"protocol"`
	ObservationCount int        `json:"observation_count"`
	SuccessCount     int        `json:"success_count"`
	FailureCount     int        `json:"failure_count"`
	SuccessRate      *float64   `json:"success_rate"`
	AvgDurationMS    *float64   `json:"avg_duration_ms,omitempty"`
	P95DurationMS    *float64   `json:"p95_duration_ms,omitempty"`
	LastObservedAt   *time.Time `json:"last_observed_at,omitempty"`
	LastFailureAt    *time.Time `json:"last_failure_at,omitempty"`
	HTTP             *HTTPTrend `json:"http,omitempty"`
	Ping             *PingTrend `json:"ping,omitempty"`
	DNS              *DNSTrend  `json:"dns,omitempty"`
	TLS              *TLSTrend  `json:"tls,omitempty"`
}

type HTTPTrend struct {
	AvgResponseTimeMS *float64   `json:"avg_response_time_ms,omitempty"`
	P95ResponseTimeMS *float64   `json:"p95_response_time_ms,omitempty"`
	LatestFailureAt   *time.Time `json:"latest_failure_at,omitempty"`
}

type PingTrend struct {
	AvgRTTMS       *float64 `json:"avg_rtt_ms,omitempty"`
	AvgLossRate    *float64 `json:"avg_loss_rate,omitempty"`
	AvgJitterMS    *float64 `json:"avg_jitter_ms,omitempty"`
	LatestLossRate *float64 `json:"latest_loss_rate,omitempty"`
	LatestAvgRTTMS *float64 `json:"latest_avg_rtt_ms,omitempty"`
	LatestJitterMS *float64 `json:"latest_jitter_ms,omitempty"`
}

type DNSTrend struct {
	SuccessRate     *float64       `json:"success_rate,omitempty"`
	LatestTTLMin    *float64       `json:"latest_ttl_min,omitempty"`
	LatestTTLMax    *float64       `json:"latest_ttl_max,omitempty"`
	LatestTTLAvg    *float64       `json:"latest_ttl_avg,omitempty"`
	PreviousTTLMin  *float64       `json:"previous_ttl_min,omitempty"`
	PreviousTTLMax  *float64       `json:"previous_ttl_max,omitempty"`
	PreviousTTLAvg  *float64       `json:"previous_ttl_avg,omitempty"`
	RiskFlagCounts  map[string]int `json:"risk_flag_counts,omitempty"`
	LatestRiskFlags []string       `json:"latest_risk_flags,omitempty"`
}

type TLSTrend struct {
	LatestCertDaysLeft     *int       `json:"latest_cert_days_left,omitempty"`
	PreviousCertDaysLeft   *int       `json:"previous_cert_days_left,omitempty"`
	CertIssuerChanged      bool       `json:"cert_issuer_changed"`
	CertFingerprintChanged bool       `json:"cert_fingerprint_changed"`
	LatestCertIssuer       string     `json:"latest_cert_issuer,omitempty"`
	LatestCertFingerprint  string     `json:"latest_cert_fingerprint_sha256,omitempty"`
	LatestCertNotAfter     string     `json:"latest_cert_not_after,omitempty"`
	LatestCertObservedAt   *time.Time `json:"latest_cert_observed_at,omitempty"`
}

type EdgeProviderHint struct {
	Provider   string                 `json:"provider"`
	HintType   string                 `json:"hint_type"`
	Confidence string                 `json:"confidence"`
	Evidence   []EdgeProviderEvidence `json:"evidence"`
}

type EdgeProviderEvidence struct {
	Source string `json:"source"`
	Field  string `json:"field"`
	Value  string `json:"value"`
}
