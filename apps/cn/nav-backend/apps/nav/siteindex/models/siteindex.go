package models

import (
	"time"

	cm "github.com/gofurry/gofurry-nav-backend/common/models"
)

const (
	SiteIndexSchemaVersion = 1

	SiteIndexStateReady = "ready"
	SiteIndexStateEmpty = "empty"
	SiteIndexStateError = "error"
)

type SiteIndexItem struct {
	ID        int64        `json:"id"`
	Domains   []string     `json:"domains"`
	UpdatedAt cm.LocalTime `json:"updated_at"`
}

type SiteIndexResponse struct {
	SchemaVersion  int             `json:"schema_version"`
	GeneratedAt    time.Time       `json:"generated_at"`
	State          string          `json:"state"`
	ReasonMessages []string        `json:"reason_messages,omitempty"`
	Items          []SiteIndexItem `json:"items"`
}
