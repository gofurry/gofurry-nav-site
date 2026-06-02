package models

import "time"

const (
	SummaryStateReady   = "ready"
	SummaryStateMissing = "missing"
	SummaryStateStale   = "stale"

	StatusHealthy  = "healthy"
	StatusWarning  = "warning"
	StatusDegraded = "degraded"
	StatusUnknown  = "unknown"
	StatusDown     = "down"
)

type ProtocolSummary struct {
	Protocol          string    `json:"protocol"`
	Status            string    `json:"status"`
	ObservedAt        time.Time `json:"observed_at"`
	DurationMS        int64     `json:"duration_ms"`
	Stale             bool      `json:"stale"`
	StaleAfterSeconds int64     `json:"stale_after_seconds"`
	ErrorCode         string    `json:"error_code,omitempty"`
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

type TargetSummaryResponse struct {
	State             string                     `json:"state"`
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

type SiteSummaryResponse struct {
	State           string                   `json:"state"`
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
