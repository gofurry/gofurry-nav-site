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
	EdgeProviderHints []EdgeProviderHint         `json:"edge_provider_hints,omitempty"`
	ObservedAt        time.Time                  `json:"observed_at"`
	GeneratedAt       time.Time                  `json:"generated_at"`
	SchemaVersion     int                        `json:"schema_version"`
}

type SiteSummaryDocument struct {
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

type TargetSummaryItem struct {
	Target            string             `json:"target"`
	Status            string             `json:"status"`
	ReasonCodes       []string           `json:"reason_codes"`
	ReasonMessages    []string           `json:"reason_messages"`
	EdgeProviderHints []EdgeProviderHint `json:"edge_provider_hints,omitempty"`
	ObservedAt        time.Time          `json:"observed_at"`
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
