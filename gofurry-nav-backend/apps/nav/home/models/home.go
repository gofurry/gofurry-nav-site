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

type HomeGroup struct {
	ID       string             `json:"id"`
	Name     string             `json:"name"`
	Info     string             `json:"info"`
	Priority int64              `json:"priority"`
	Sites    []navmodels.SiteVo `json:"sites"`
}

type HomeSpotlight struct {
	PageSize int                `json:"page_size"`
	Featured []navmodels.SiteVo `json:"featured"`
	Popular  []navmodels.SiteVo `json:"popular"`
	Latest   []navmodels.SiteVo `json:"latest"`
	Random   []navmodels.SiteVo `json:"random"`
}

type HomeResponse struct {
	SchemaVersion  int                    `json:"schema_version"`
	GeneratedAt    time.Time              `json:"generated_at"`
	CacheState     map[string]string      `json:"cache_state"`
	ReasonMessages map[string]string      `json:"reason_messages,omitempty"`
	Sites          []navmodels.SiteVo     `json:"sites"`
	Groups         []HomeGroup            `json:"groups"`
	Spotlight      HomeSpotlight          `json:"spotlight"`
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

type HomeSayingResponse struct {
	SchemaVersion  int                    `json:"schema_version"`
	GeneratedAt    time.Time              `json:"generated_at"`
	State          string                 `json:"state"`
	ReasonMessages []string               `json:"reason_messages,omitempty"`
	Saying         *navmodels.SayingModel `json:"saying"`
}

type HomeBackgroundsResponse struct {
	SchemaVersion  int             `json:"schema_version"`
	GeneratedAt    time.Time       `json:"generated_at"`
	State          string          `json:"state"`
	ReasonMessages []string        `json:"reason_messages,omitempty"`
	Backgrounds    HomeBackgrounds `json:"backgrounds"`
}
