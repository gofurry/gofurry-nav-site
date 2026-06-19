package models

import (
	"time"

	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
)

const (
	SiteGroupSchemaVersion = 1

	SiteGroupStateReady   = "ready"
	SiteGroupStateMissing = "missing"
)

type SiteGroupInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Info       string `json:"info"`
	Priority   int64  `json:"priority"`
	SiteCount  int    `json:"site_count"`
	DetailPath string `json:"detail_path"`
}

type CachedSiteGroup struct {
	GeneratedAt time.Time          `json:"generated_at"`
	Group       SiteGroupInfo      `json:"group"`
	Sites       []navmodels.SiteVo `json:"sites"`
}

type SiteGroupPageResponse struct {
	SchemaVersion  int                `json:"schema_version"`
	GeneratedAt    time.Time          `json:"generated_at"`
	State          string             `json:"state"`
	ReasonMessages []string           `json:"reason_messages,omitempty"`
	Group          *SiteGroupInfo     `json:"group,omitempty"`
	Page           int                `json:"page"`
	PageSize       int                `json:"page_size"`
	Total          int                `json:"total"`
	HasMore        bool               `json:"has_more"`
	Items          []navmodels.SiteVo `json:"items"`
}
