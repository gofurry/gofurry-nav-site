package models

import (
	"time"

	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
)

const (
	HomeSchemaVersion = 4

	HomeStateReady   = "ready"
	HomeStateMissing = "missing"
)

type HeroAsset struct {
	ID        int64  `json:"id,string"`
	ObjectKey string `json:"object_key"`
}

// Hero preferences belong to the browser. Only these request-time selectors
// enter the resolver; they are never stored in the derived home cache.
type HeroSelection struct {
	DesktopID int64
	MobileID  int64
	Local     bool
}

type HeroCatalogQuery struct {
	Variant    string
	PageNum    int32
	PageSize   int32
	SelectedID int64
}

type HeroCatalogItem struct {
	ID        int64  `json:"id,string"`
	Name      string `json:"name"`
	ObjectKey string `json:"object_key"`
}

type HeroCatalog struct {
	SchemaVersion int               `json:"schema_version"`
	Variant       string            `json:"variant"`
	PageNum       int32             `json:"page_num"`
	PageSize      int32             `json:"page_size"`
	Total         int64             `json:"total"`
	Items         []HeroCatalogItem `json:"items"`
	Selected      *HeroCatalogItem  `json:"selected"`
}
type HomeHero struct {
	Desktop *HeroAsset `json:"desktop"`
	Mobile  *HeroAsset `json:"mobile"`
}

type HomeGroup struct {
	ID         string             `json:"id"`
	Name       string             `json:"name"`
	Info       string             `json:"info"`
	Priority   int64              `json:"priority"`
	SiteCount  int                `json:"site_count"`
	HasMore    bool               `json:"has_more"`
	DetailPath string             `json:"detail_path"`
	Sites      []navmodels.SiteVo `json:"sites"`
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
	Sites          []navmodels.SiteVo     `json:"sites,omitempty"`
	Groups         []HomeGroup            `json:"groups"`
	Spotlight      HomeSpotlight          `json:"spotlight"`
	Ping           map[string]string      `json:"ping"`
	Saying         *navmodels.SayingModel `json:"saying"`
	Hero           HomeHero               `json:"hero"`
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

type HomeHeroResponse struct {
	SchemaVersion  int       `json:"schema_version"`
	GeneratedAt    time.Time `json:"generated_at"`
	State          string    `json:"state"`
	ReasonMessages []string  `json:"reason_messages,omitempty"`
	Hero           HomeHero  `json:"hero"`
}

type BackgroundPattern struct {
	ID            int64   `json:"id,string"`
	Name          string  `json:"name"`
	NameEn        string  `json:"name_en"`
	ObjectKey     string  `json:"object_key"`
	LightColor    string  `json:"light_color"`
	DarkColor     string  `json:"dark_color"`
	LightOpacity  float64 `json:"light_opacity"`
	DarkOpacity   float64 `json:"dark_opacity"`
	DefaultSizePx int32   `json:"default_size_px"`
}
type PatternCatalog struct {
	SchemaVersion int                 `json:"schema_version"`
	Patterns      []BackgroundPattern `json:"patterns"`
}
