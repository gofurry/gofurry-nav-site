package domain

import "time"

// Language identifies one localized Steam payload language stored by collector v2.
type Language string

const (
	LanguageZH Language = "zh"
	LanguageEN Language = "en"
)

// Region identifies one Steam Store price region.
type Region string

const (
	RegionCN Region = "CN"
	RegionHK Region = "HK"
	RegionUS Region = "US"
)

// Source identifies the upstream data source.
type Source string

const (
	SourceSteam Source = "steam"
)

// TaskType identifies one collector v2 task family.
type TaskType string

const (
	TaskDetails TaskType = "details"
	TaskNews    TaskType = "news"
	TaskPlayers TaskType = "players"
)

// Status identifies a task or item collection result.
type Status string

const (
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
	StatusSkipped Status = "skipped"
	StatusPartial Status = "partial"
)

// PlatformSupport describes Steam platform support.
type PlatformSupport struct {
	Windows bool `json:"windows"`
	Mac     bool `json:"mac"`
	Linux   bool `json:"linux"`
}

// ReleaseDate stores Steam's release date payload without parsing locale-specific text.
type ReleaseDate struct {
	ComingSoon bool   `json:"coming_soon"`
	DateText   string `json:"date_text"`
}

// TimeRange describes one observed collection interval.
type TimeRange struct {
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
}
