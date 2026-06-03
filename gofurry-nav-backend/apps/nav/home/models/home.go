package models

import (
	"time"

	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
)

const (
	HomeSchemaVersion = 2

	HomeStateReady   = "ready"
	HomeStateMissing = "missing"
)

type HomeBackgrounds struct {
	Desktop string `json:"desktop"`
	Mobile  string `json:"mobile"`
}

type HomeResponse struct {
	SchemaVersion  int                    `json:"schema_version"`
	GeneratedAt    time.Time              `json:"generated_at"`
	CacheState     map[string]string      `json:"cache_state"`
	ReasonMessages map[string]string      `json:"reason_messages,omitempty"`
	Sites          []navmodels.SiteVo     `json:"sites"`
	Groups         []navmodels.GroupVo    `json:"groups"`
	Ping           map[string]string      `json:"ping"`
	Saying         *navmodels.SayingModel `json:"saying"`
	Backgrounds    HomeBackgrounds        `json:"backgrounds"`
}

type HomePingResponse struct {
	SchemaVersion  int               `json:"schema_version"`
	GeneratedAt    time.Time         `json:"generated_at"`
	State          string            `json:"state"`
	ReasonMessages []string          `json:"reason_messages,omitempty"`
	Ping           map[string]string `json:"ping"`
}
