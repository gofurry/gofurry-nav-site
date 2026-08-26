package domain

import "time"

// StoreLocale identifies one localized Steam payload locale stored by collector v2.
// It is intentionally separate from a game's canonical supported-language facts.
type StoreLocale string

const (
	StoreLocaleZH StoreLocale = "zh"
	StoreLocaleEN StoreLocale = "en"
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

// RawReleaseDate stores Steam's release date payload without parsing locale-specific text.
type RawReleaseDate struct {
	ComingSoon bool   `json:"coming_soon"`
	DateText   string `json:"date_text"`
}

// ReleaseAvailability is the current normalized Steam availability observation.
type ReleaseAvailability string

const (
	ReleaseUpcoming  ReleaseAvailability = "upcoming"
	ReleaseAvailable ReleaseAvailability = "available"
	ReleaseUnknown   ReleaseAvailability = "unknown"
)

// ReleasePrecision identifies the calendar precision of a normalized release.
type ReleasePrecision string

const (
	ReleasePrecisionDay     ReleasePrecision = "day"
	ReleasePrecisionMonth   ReleasePrecision = "month"
	ReleasePrecisionQuarter ReleasePrecision = "quarter"
	ReleasePrecisionYear    ReleasePrecision = "year"
	ReleasePrecisionTBA     ReleasePrecision = "tba"
	ReleasePrecisionNone    ReleasePrecision = "none"
	ReleasePrecisionUnknown ReleasePrecision = "unknown"
)

// GameReleaseState is one authoritative US/English canonical release observation.
type GameReleaseState struct {
	GameID int64 `json:"game_id"`

	Availability ReleaseAvailability `json:"availability"`
	Precision    ReleasePrecision    `json:"precision"`
	ExactDate    *time.Time          `json:"exact_date"`
	Year         *int                `json:"year"`
	Month        *int                `json:"month"`
	Quarter      *int                `json:"quarter"`
	WindowStart  *time.Time          `json:"window_start"`
	WindowEnd    *time.Time          `json:"window_end"`
	RawText      string              `json:"raw_text"`
	Source       Source              `json:"source"`
	SourceRegion Region              `json:"source_region"`
	SourceLocale StoreLocale         `json:"source_locale"`
	Normalizer   string              `json:"normalizer_version"`
	ObservedAt   time.Time           `json:"observed_at"`
}

// GameLanguage is one normalized supported-language fact.
type GameLanguage struct {
	Code               *string     `json:"code"`
	SteamName          string      `json:"steam_name"`
	SteamAPICode       *string     `json:"steam_api_code"`
	SteamWebCode       *string     `json:"steam_web_code"`
	Tier               string      `json:"tier"`
	InterfaceSupported *bool       `json:"interface_supported"`
	SubtitlesSupported *bool       `json:"subtitles_supported"`
	FullAudioSupported *bool       `json:"full_audio_supported"`
	SortOrder          int         `json:"sort_order"`
	Source             Source      `json:"source"`
	SourceRegion       Region      `json:"source_region"`
	SourceLocale       StoreLocale `json:"source_locale"`
	Normalizer         string      `json:"normalizer_version"`
	ObservedAt         time.Time   `json:"observed_at"`
}

// GameLanguages is an authoritative replacement set. A nil *GameLanguages in
// DetailsCollection means this collection did not contain a reliable observation.
type GameLanguages struct {
	Items []GameLanguage `json:"items"`
}

// TimeRange describes one observed collection interval.
type TimeRange struct {
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
}
