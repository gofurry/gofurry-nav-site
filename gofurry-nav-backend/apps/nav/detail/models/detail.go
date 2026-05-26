package models

import (
	"encoding/json"
	"strings"
	"time"

	readmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/readmodel/models"
	summarymodels "github.com/gofurry/gofurry-nav-backend/apps/nav/summary/models"
)

const (
	DetailSchemaVersion         = 1
	TableNameGfnCollectorDomain = "gfn_collector_domain"
)

type CollectorDomain struct {
	ID      int64   `gorm:"column:id;type:bigint;primaryKey" json:"id"`
	SiteID  int64   `gorm:"column:site_id;type:bigint" json:"site_id"`
	Name    string  `gorm:"column:name;type:character varying(255);not null" json:"name"`
	Proxy   string  `gorm:"column:proxy;type:character varying(4);not null" json:"proxy"`
	Prefix  *string `gorm:"column:prefix;type:character varying(255)" json:"prefix"`
	TLS     string  `gorm:"column:tls;type:character varying(4);not null" json:"tls"`
	Deleted bool    `gorm:"column:deleted;type:boolean" json:"deleted"`
}

func (*CollectorDomain) TableName() string {
	return TableNameGfnCollectorDomain
}

func (d CollectorDomain) TargetName() string {
	return strings.TrimSpace(prefixValue(d.Prefix) + d.Name)
}

type SiteInfo struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Info      string  `json:"info"`
	Icon      *string `json:"icon"`
	Country   *string `json:"country"`
	Nsfw      string  `json:"nsfw"`
	Welfare   string  `json:"welfare"`
	ViewCount int64   `json:"view_count"`
}

type SiteTarget struct {
	Target       string  `json:"target"`
	DomainID     int64   `json:"domain_id"`
	Name         string  `json:"name"`
	Prefix       *string `json:"prefix"`
	TLS          string  `json:"tls"`
	Proxy        string  `json:"proxy"`
	Source       string  `json:"source"`
	Registered   bool    `json:"registered"`
	SummaryOnly  bool    `json:"summary_only"`
	SummaryState string  `json:"summary_state"`
	Status       string  `json:"status"`
}

type DerivedState struct {
	Trend   TargetTrendResponse   `json:"trend"`
	Changes TargetChangesResponse `json:"changes"`
}

type SiteDetailResponse struct {
	Site            SiteInfo                            `json:"site"`
	Targets         []SiteTarget                        `json:"targets"`
	SelectedTarget  string                              `json:"selected_target"`
	SiteSummary     summarymodels.SiteSummaryResponse   `json:"site_summary"`
	TargetSummary   summarymodels.TargetSummaryResponse `json:"target_summary"`
	LatestCore      TargetLatestResponse                `json:"latest_core"`
	Derived         DerivedState                        `json:"derived"`
	LightProbeState TargetLatestResponse                `json:"light_probe_state"`
	GeneratedAt     time.Time                           `json:"generated_at"`
	SchemaVersion   int                                 `json:"schema_version"`
}

type TargetLatestResponse struct {
	State          string                                  `json:"state"`
	SiteID         int64                                   `json:"site_id"`
	Target         string                                  `json:"target"`
	Protocols      map[string]readmodels.CollectorEnvelope `json:"protocols"`
	ReasonCodes    []string                                `json:"reason_codes,omitempty"`
	ReasonMessages []string                                `json:"reason_messages,omitempty"`
	GeneratedAt    time.Time                               `json:"generated_at"`
	SchemaVersion  int                                     `json:"schema_version"`
}

type TargetObservationsResponse struct {
	State          string                         `json:"state"`
	SiteID         int64                          `json:"site_id"`
	Target         string                         `json:"target"`
	Protocol       string                         `json:"protocol"`
	Limit          int                            `json:"limit"`
	Items          []readmodels.CollectorEnvelope `json:"items"`
	ReasonCodes    []string                       `json:"reason_codes,omitempty"`
	ReasonMessages []string                       `json:"reason_messages,omitempty"`
	GeneratedAt    time.Time                      `json:"generated_at"`
	SchemaVersion  int                            `json:"schema_version"`
}

type TargetTrendResponse struct {
	State          string          `json:"state"`
	SiteID         int64           `json:"site_id"`
	Target         string          `json:"target"`
	Windows        json.RawMessage `json:"windows"`
	ReasonCodes    []string        `json:"reason_codes,omitempty"`
	ReasonMessages []string        `json:"reason_messages,omitempty"`
	GeneratedAt    time.Time       `json:"generated_at"`
	SchemaVersion  int             `json:"schema_version"`
}

type TargetChangesResponse struct {
	State          string          `json:"state"`
	SiteID         int64           `json:"site_id"`
	Target         string          `json:"target"`
	Events         json.RawMessage `json:"events"`
	ReasonCodes    []string        `json:"reason_codes,omitempty"`
	ReasonMessages []string        `json:"reason_messages,omitempty"`
	GeneratedAt    time.Time       `json:"generated_at"`
	SchemaVersion  int             `json:"schema_version"`
}

func prefixValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
