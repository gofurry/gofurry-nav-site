package domain

import "time"

// GameDetails is the collector v2 canonical game details contract.
type GameDetails struct {
	GameID int64  `json:"game_id"`
	AppID  uint32 `json:"appid"`

	Type      string `json:"type"`
	Name      string `json:"name"`
	IsFree    bool   `json:"is_free"`
	Website   string `json:"website"`
	HeaderURL string `json:"header_url"`

	Developers []string        `json:"developers"`
	Publishers []string        `json:"publishers"`
	Release    ReleaseDate     `json:"release"`
	Platforms  PlatformSupport `json:"platforms"`

	SupportedLanguages string             `json:"supported_languages"`
	SupportInfo        SupportInfo        `json:"support_info"`
	ContentDescriptors ContentDescriptors `json:"content_descriptors"`
	Ratings            []Rating           `json:"ratings"`

	CollectedAt time.Time `json:"collected_at"`
}

// GameLocalizedDetails stores language-specific copy and rich text.
type GameLocalizedDetails struct {
	GameID   int64    `json:"game_id"`
	AppID    uint32   `json:"appid"`
	Language Language `json:"language"`

	Name                string `json:"name"`
	ShortDescription    string `json:"short_description"`
	DetailedDescription string `json:"detailed_description"`
	AboutTheGame        string `json:"about_the_game"`

	CollectedAt time.Time `json:"collected_at"`
}

// GamePrice stores one regional price snapshot.
type GamePrice struct {
	GameID int64  `json:"game_id"`
	AppID  uint32 `json:"appid"`
	Region Region `json:"region"`

	IsFree           bool   `json:"is_free"`
	Currency         string `json:"currency"`
	Initial          int64  `json:"initial"`
	Final            int64  `json:"final"`
	DiscountPercent  int64  `json:"discount_percent"`
	InitialFormatted string `json:"initial_formatted"`
	FinalFormatted   string `json:"final_formatted"`

	CollectedAt time.Time `json:"collected_at"`
}

// GameMedia groups media assets for one app.
type GameMedia struct {
	GameID int64  `json:"game_id"`
	AppID  uint32 `json:"appid"`

	HeaderURL        string       `json:"header_url"`
	CapsuleURL       string       `json:"capsule_url"`
	CapsuleV5URL     string       `json:"capsule_v5_url"`
	BackgroundURL    string       `json:"background_url"`
	BackgroundRawURL string       `json:"background_raw_url"`
	Screenshots      []Screenshot `json:"screenshots"`
	Movies           []Movie      `json:"movies"`

	CollectedAt time.Time `json:"collected_at"`
}

// GameMediaAsset is the unified Steam media/static asset contract.
type GameMediaAsset struct {
	GameID int64  `json:"game_id"`
	AppID  uint32 `json:"appid"`

	AssetType    string `json:"asset_type"`
	AssetFamily  string `json:"asset_family"`
	Source       string `json:"source"`
	Language     string `json:"language"`
	MediaKey     string `json:"media_key"`
	Title        string `json:"title"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
	Format       string `json:"format"`

	Exists        *bool  `json:"exists"`
	StatusCode    int    `json:"status_code"`
	ContentType   string `json:"content_type"`
	ContentLength int64  `json:"content_length"`
	Extra         any    `json:"extra"`
	SortOrder     int    `json:"sort_order"`

	CheckedAt   *time.Time `json:"checked_at"`
	CollectedAt time.Time  `json:"collected_at"`
}

// Screenshot stores one Steam screenshot asset pair.
type Screenshot struct {
	ID           int    `json:"id"`
	ThumbnailURL string `json:"thumbnail_url"`
	FullURL      string `json:"full_url"`
}

// Movie stores one Steam movie/trailer asset.
type Movie struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	ThumbnailURL string `json:"thumbnail_url"`
	WebM480URL   string `json:"webm_480_url"`
	WebMMaxURL   string `json:"webm_max_url"`
	MP4480URL    string `json:"mp4_480_url"`
	MP4MaxURL    string `json:"mp4_max_url"`
	DASHAV1URL   string `json:"dash_av1_url"`
	DASHH264URL  string `json:"dash_h264_url"`
	HLSH264URL   string `json:"hls_h264_url"`
	Highlight    bool   `json:"highlight"`
}

// SystemRequirements stores platform-specific requirements.
type SystemRequirements struct {
	GameID int64  `json:"game_id"`
	AppID  uint32 `json:"appid"`

	PC    Requirements `json:"pc"`
	Mac   Requirements `json:"mac"`
	Linux Requirements `json:"linux"`

	CollectedAt time.Time `json:"collected_at"`
}

// Requirements keeps Steam's HTML-ish requirements text intact for backend cleaning/display.
type Requirements struct {
	Minimum     string `json:"minimum"`
	Recommended string `json:"recommended"`
}

// SupportInfo stores Steam support contact data.
type SupportInfo struct {
	URL   string `json:"url"`
	Email string `json:"email"`
}

// ContentDescriptors stores Steam content descriptors.
type ContentDescriptors struct {
	IDs   []int  `json:"ids"`
	Notes string `json:"notes"`
}

// Rating stores one rating board value.
type Rating struct {
	Board       string `json:"board"`
	Rating      string `json:"rating"`
	RequiredAge string `json:"required_age"`
}

// DetailsCollection is one complete v2 details collection result for an app.
type DetailsCollection struct {
	Details      GameDetails            `json:"details"`
	Localized    []GameLocalizedDetails `json:"localized"`
	Prices       []GamePrice            `json:"prices"`
	Media        GameMedia              `json:"media"`
	Assets       []GameMediaAsset       `json:"assets"`
	Requirements SystemRequirements     `json:"requirements"`
	Snapshots    []RawSnapshot          `json:"snapshots"`
}
