package models

import (
	"time"

	cm "github.com/gofurry/gofurry-game-backend/common/models"
)

const (
	TableNameGfgGameV2Details            = "gfg_game_v2_details"
	TableNameGfgGameV2LocalizedDetails   = "gfg_game_v2_localized_details"
	TableNameGfgGameV2Prices             = "gfg_game_v2_prices"
	TableNameGfgGameV2Media              = "gfg_game_v2_media"
	TableNameGfgGameV2Requirements       = "gfg_game_v2_requirements"
	TableNameGfgGameV2News               = "gfg_game_v2_news"
	TableNameGfgGameV2PlayerCounts       = "gfg_game_v2_player_counts"
	TableNameGfgGameV2CollectRuns        = "gfg_game_v2_collect_runs"
	TableNameGfgGameV2CollectTaskResults = "gfg_game_v2_collect_task_results"
)

type GfgGameV2Details struct {
	GameID             int64     `gorm:"column:game_id;primaryKey" json:"game_id"`
	AppID              int64     `gorm:"column:appid" json:"appid"`
	Source             string    `gorm:"column:source" json:"source"`
	Type               string    `gorm:"column:type" json:"type"`
	Name               string    `gorm:"column:name" json:"name"`
	IsFree             bool      `gorm:"column:is_free" json:"is_free"`
	Website            *string   `gorm:"column:website" json:"website"`
	HeaderURL          *string   `gorm:"column:header_url" json:"header_url"`
	Developers         *string   `gorm:"column:developers" json:"developers"`
	Publishers         *string   `gorm:"column:publishers" json:"publishers"`
	ReleaseComingSoon  bool      `gorm:"column:release_coming_soon" json:"release_coming_soon"`
	ReleaseDateText    *string   `gorm:"column:release_date_text" json:"release_date_text"`
	Platforms          *string   `gorm:"column:platforms" json:"platforms"`
	SupportedLanguages *string   `gorm:"column:supported_languages" json:"supported_languages"`
	SupportInfo        *string   `gorm:"column:support_info" json:"support_info"`
	ContentDescriptors *string   `gorm:"column:content_descriptors" json:"content_descriptors"`
	Ratings            *string   `gorm:"column:ratings" json:"ratings"`
	CollectedAt        time.Time `gorm:"column:collected_at" json:"collected_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (*GfgGameV2Details) TableName() string { return TableNameGfgGameV2Details }

type GfgGameV2LocalizedDetails struct {
	GameID              int64     `gorm:"column:game_id;primaryKey" json:"game_id"`
	AppID               int64     `gorm:"column:appid" json:"appid"`
	Lang                string    `gorm:"column:lang;primaryKey" json:"lang"`
	Name                string    `gorm:"column:name" json:"name"`
	ShortDescription    *string   `gorm:"column:short_description" json:"short_description"`
	DetailedDescription *string   `gorm:"column:detailed_description" json:"detailed_description"`
	AboutTheGame        *string   `gorm:"column:about_the_game" json:"about_the_game"`
	CollectedAt         time.Time `gorm:"column:collected_at" json:"collected_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (*GfgGameV2LocalizedDetails) TableName() string { return TableNameGfgGameV2LocalizedDetails }

type GfgGameV2Price struct {
	GameID           int64     `gorm:"column:game_id;primaryKey" json:"game_id"`
	AppID            int64     `gorm:"column:appid" json:"appid"`
	Region           string    `gorm:"column:region;primaryKey" json:"region"`
	IsFree           bool      `gorm:"column:is_free" json:"is_free"`
	Currency         *string   `gorm:"column:currency" json:"currency"`
	InitialAmount    int64     `gorm:"column:initial_amount" json:"initial_amount"`
	FinalAmount      int64     `gorm:"column:final_amount" json:"final_amount"`
	DiscountPercent  int64     `gorm:"column:discount_percent" json:"discount_percent"`
	InitialFormatted *string   `gorm:"column:initial_formatted" json:"initial_formatted"`
	FinalFormatted   *string   `gorm:"column:final_formatted" json:"final_formatted"`
	CollectedAt      time.Time `gorm:"column:collected_at" json:"collected_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (*GfgGameV2Price) TableName() string { return TableNameGfgGameV2Prices }

type GfgGameV2Media struct {
	ID           int64     `gorm:"column:id;primaryKey" json:"id"`
	GameID       int64     `gorm:"column:game_id" json:"game_id"`
	AppID        int64     `gorm:"column:appid" json:"appid"`
	MediaType    string    `gorm:"column:media_type" json:"media_type"`
	MediaKey     string    `gorm:"column:media_key" json:"media_key"`
	Title        *string   `gorm:"column:title" json:"title"`
	URL          *string   `gorm:"column:url" json:"url"`
	ThumbnailURL *string   `gorm:"column:thumbnail_url" json:"thumbnail_url"`
	Extra        *string   `gorm:"column:extra" json:"extra"`
	SortOrder    int       `gorm:"column:sort_order" json:"sort_order"`
	CollectedAt  time.Time `gorm:"column:collected_at" json:"collected_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (*GfgGameV2Media) TableName() string { return TableNameGfgGameV2Media }

type GfgGameV2Requirements struct {
	GameID      int64     `gorm:"column:game_id;primaryKey" json:"game_id"`
	AppID       int64     `gorm:"column:appid" json:"appid"`
	PC          *string   `gorm:"column:pc" json:"pc"`
	Mac         *string   `gorm:"column:mac" json:"mac"`
	Linux       *string   `gorm:"column:linux" json:"linux"`
	CollectedAt time.Time `gorm:"column:collected_at" json:"collected_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (*GfgGameV2Requirements) TableName() string { return TableNameGfgGameV2Requirements }

type GfgGameV2News struct {
	ID              int64     `gorm:"column:id;primaryKey" json:"id"`
	GameID          int64     `gorm:"column:game_id" json:"game_id"`
	AppID           int64     `gorm:"column:appid" json:"appid"`
	Lang            string    `gorm:"column:lang" json:"lang"`
	EventGID        string    `gorm:"column:event_gid" json:"event_gid"`
	AnnouncementGID *string   `gorm:"column:announcement_gid" json:"announcement_gid"`
	ForumTopicID    *string   `gorm:"column:forum_topic_id" json:"forum_topic_id"`
	Headline        string    `gorm:"column:headline" json:"headline"`
	RawBody         *string   `gorm:"column:raw_body" json:"raw_body"`
	HTML            *string   `gorm:"column:html" json:"html"`
	PlainText       *string   `gorm:"column:plain_text" json:"plain_text"`
	Summary         *string   `gorm:"column:summary" json:"summary"`
	URL             *string   `gorm:"column:url" json:"url"`
	Tags            *string   `gorm:"column:tags" json:"tags"`
	VoteUpCount     int64     `gorm:"column:vote_up_count" json:"vote_up_count"`
	VoteDownCount   int64     `gorm:"column:vote_down_count" json:"vote_down_count"`
	CommentCount    int64     `gorm:"column:comment_count" json:"comment_count"`
	RawEvent        *string   `gorm:"column:raw_event" json:"raw_event"`
	PublishedAt     time.Time `gorm:"column:published_at" json:"published_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`
	CollectedAt     time.Time `gorm:"column:collected_at" json:"collected_at"`
}

func (*GfgGameV2News) TableName() string { return TableNameGfgGameV2News }

type GfgGameV2PlayerCount struct {
	ID                 int64     `gorm:"column:id;primaryKey" json:"id"`
	RunID              string    `gorm:"column:run_id" json:"run_id"`
	GameID             int64     `gorm:"column:game_id" json:"game_id"`
	AppID              int64     `gorm:"column:appid" json:"appid"`
	Count              int64     `gorm:"column:count" json:"count"`
	Status             string    `gorm:"column:status" json:"status"`
	UpstreamStatusCode int       `gorm:"column:upstream_status_code" json:"upstream_status_code"`
	ErrorKind          string    `gorm:"column:error_kind" json:"error_kind"`
	ErrorMessage       string    `gorm:"column:error_message" json:"error_message"`
	CollectedAt        time.Time `gorm:"column:collected_at" json:"collected_at"`
}

func (*GfgGameV2PlayerCount) TableName() string { return TableNameGfgGameV2PlayerCounts }

type GfgGameV2CollectRun struct {
	ID             string     `gorm:"column:id;primaryKey" json:"id"`
	TaskType       string     `gorm:"column:task_type" json:"task_type"`
	Status         string     `gorm:"column:status" json:"status"`
	TotalCount     int64      `gorm:"column:total_count" json:"total_count"`
	SuccessCount   int64      `gorm:"column:success_count" json:"success_count"`
	FailedCount    int64      `gorm:"column:failed_count" json:"failed_count"`
	SkippedCount   int64      `gorm:"column:skipped_count" json:"skipped_count"`
	PartialCount   int64      `gorm:"column:partial_count" json:"partial_count"`
	TaskSummary    string     `gorm:"column:task_summary" json:"task_summary"`
	DurationMillis int64      `gorm:"column:duration_millis" json:"duration_millis"`
	ErrorKind      string     `gorm:"column:error_kind" json:"error_kind"`
	ErrorMessage   string     `gorm:"column:error_message" json:"error_message"`
	StartedAt      time.Time  `gorm:"column:started_at" json:"started_at"`
	EndedAt        *time.Time `gorm:"column:ended_at" json:"ended_at"`
}

func (*GfgGameV2CollectRun) TableName() string { return TableNameGfgGameV2CollectRuns }

type GfgGameV2CollectTaskResult struct {
	ID                 int64      `gorm:"column:id;primaryKey" json:"id"`
	RunID              string     `gorm:"column:run_id" json:"run_id"`
	TaskType           string     `gorm:"column:task_type" json:"task_type"`
	Status             string     `gorm:"column:status" json:"status"`
	GameID             int64      `gorm:"column:game_id" json:"game_id"`
	AppID              int64      `gorm:"column:appid" json:"appid"`
	UpstreamStatusCode int        `gorm:"column:upstream_status_code" json:"upstream_status_code"`
	TrafficBucket      string     `gorm:"column:traffic_bucket" json:"traffic_bucket"`
	RetryCount         int        `gorm:"column:retry_count" json:"retry_count"`
	DurationMillis     int64      `gorm:"column:duration_millis" json:"duration_millis"`
	ErrorKind          string     `gorm:"column:error_kind" json:"error_kind"`
	ErrorMessage       string     `gorm:"column:error_message" json:"error_message"`
	StartedAt          time.Time  `gorm:"column:started_at" json:"started_at"`
	EndedAt            *time.Time `gorm:"column:ended_at" json:"ended_at"`
}

func (*GfgGameV2CollectTaskResult) TableName() string { return TableNameGfgGameV2CollectTaskResults }

type GameV2SiteRecord struct {
	ID         int64        `gorm:"column:id" json:"id"`
	Name       string       `gorm:"column:name" json:"name"`
	NameEn     string       `gorm:"column:name_en" json:"name_en"`
	Info       string       `gorm:"column:info" json:"info"`
	InfoEn     string       `gorm:"column:info_en" json:"info_en"`
	Resources  *string      `gorm:"column:resources" json:"resources"`
	Groups     *string      `gorm:"column:groups" json:"groups"`
	Links      *string      `gorm:"column:links" json:"links"`
	AppID      int64        `gorm:"column:appid" json:"appid"`
	Header     string       `gorm:"column:header" json:"header"`
	ViewCount  int64        `gorm:"column:view_count" json:"view_count"`
	Weight     int64        `gorm:"column:weight" json:"weight"`
	CreateTime cm.LocalTime `gorm:"column:create_time" json:"create_time"`
	UpdateTime cm.LocalTime `gorm:"column:update_time" json:"update_time"`
}

type GameV2Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type GameV2Aggregate struct {
	Site         GameV2SiteRecord
	Details      *GfgGameV2Details
	Localized    *GfgGameV2LocalizedDetails
	Prices       []GfgGameV2Price
	Media        []GfgGameV2Media
	Requirements *GfgGameV2Requirements
	News         []GfgGameV2News
	OnlineCount  *GfgGameV2PlayerCount
	Tags         []GameV2Tag
}

type GameV2DetailQuery struct {
	GameID    int64
	AppID     int64
	Lang      string
	NewsLimit int
}

type GameV2ListQuery struct {
	Lang         string
	Region       string
	Limit        int
	Offset       int
	Sort         string
	UpdatedSince time.Time
}

type GameV2NewsQuery struct {
	GameID       int64
	AppID        int64
	Lang         string
	Limit        int
	Offset       int
	UpdatedSince time.Time
}

type GameV2PanelQuery struct {
	Lang      string
	Region    string
	Limit     int
	NewsLimit int
}

type GameV2CollectRunQuery struct {
	TaskType string
	Status   string
	Limit    int
	Offset   int
}

type GameV2CollectTaskResultQuery struct {
	RunID    string
	TaskType string
	Status   string
	GameID   int64
	AppID    int64
	Limit    int
	Offset   int
}

type GameV2CollectStatus struct {
	LatestRun      *GfgGameV2CollectRun             `json:"latest_run"`
	LatestTaskRuns []GfgGameV2CollectRun            `json:"latest_task_runs"`
	Summary        []GameV2CollectTaskStatusSummary `json:"summary"`
	GeneratedAt    time.Time                        `json:"generated_at"`
}

type GameV2CollectTaskStatusSummary struct {
	TaskType string `gorm:"column:task_type" json:"task_type"`
	Status   string `gorm:"column:status" json:"status"`
	Count    int64  `gorm:"column:count" json:"count"`
}

type GameV2CollectLocalizedStatus struct {
	Lang        string    `json:"lang"`
	Name        string    `json:"name"`
	CollectedAt time.Time `json:"collected_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GameV2CollectRegionFreshness struct {
	Region      string    `json:"region"`
	Available   bool      `json:"available"`
	Currency    string    `json:"currency"`
	FinalAmount int64     `json:"final_amount"`
	CollectedAt time.Time `json:"collected_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GameV2CollectGameStatus struct {
	GameID            int64                          `json:"game_id"`
	AppID             int64                          `json:"appid"`
	Name              string                         `json:"name"`
	DetailsUpdatedAt  *time.Time                     `json:"details_updated_at"`
	Localized         []GameV2CollectLocalizedStatus `json:"localized"`
	Prices            []GameV2CollectRegionFreshness `json:"prices"`
	MediaCount        int64                          `json:"media_count"`
	NewsCount         int64                          `json:"news_count"`
	LatestNewsAt      *time.Time                     `json:"latest_news_at"`
	LatestPlayerCount *GfgGameV2PlayerCount          `json:"latest_player_count"`
	LatestTaskResults []GfgGameV2CollectTaskResult   `json:"latest_task_results"`
}

type GameV2DetailRequest struct {
	GameID    int64
	AppID     int64
	Lang      string
	Region    string
	NewsLimit int
}

type GameV2SyncListQuery struct {
	Lang         string
	Region       string
	Limit        int
	Offset       int
	UpdatedSince time.Time
}

type GameV2SyncNewsQuery struct {
	Lang         string
	Limit        int
	Offset       int
	UpdatedSince time.Time
}

type GameV2SyncGameSummary struct {
	ID          string    `json:"id"`
	AppID       string    `json:"appid"`
	Name        string    `json:"name"`
	Info        string    `json:"info"`
	ReleaseDate string    `json:"release_date"`
	Developers  []string  `json:"developers"`
	Publishers  []string  `json:"publishers"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GameV2SyncGameDetail struct {
	ID                  string                   `json:"id"`
	AppID               string                   `json:"appid"`
	Name                string                   `json:"name"`
	Info                string                   `json:"info"`
	Resources           []cm.KvModel             `json:"resources"`
	Groups              []cm.KvModel             `json:"groups"`
	ReleaseDate         string                   `json:"release_date"`
	Developers          []string                 `json:"developers"`
	Publishers          []string                 `json:"publishers"`
	Links               []cm.KvModel             `json:"links"`
	Platform            string                   `json:"platform"`
	Tags                []GameV2Tag              `json:"tags"`
	SupportedLanguages  string                   `json:"supported_languages"`
	Website             string                   `json:"website"`
	DetailedDescription string                   `json:"detailed_description"`
	AboutTheGame        string                   `json:"about_the_game"`
	PcRequirements      GameV2SyncPCRequirements `json:"pc_requirements"`
	UpdatedAt           time.Time                `json:"updated_at"`
}

type GameV2SyncPCRequirements struct {
	Minimum     string `json:"minimum"`
	Recommended string `json:"recommended"`
}

type GameV2SyncNewsItem struct {
	ID          string    `json:"id"`
	GameID      string    `json:"game_id"`
	AppID       string    `json:"appid"`
	Name        string    `json:"name"`
	PostTime    string    `json:"post_time"`
	Headline    string    `json:"headline"`
	Author      string    `json:"author"`
	Content     string    `json:"content"`
	URL         string    `json:"url"`
	Lang        string    `json:"lang"`
	UpdatedAt   time.Time `json:"updated_at"`
	PublishedAt time.Time `json:"published_at"`
}

type GameV2SyncCreator struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Info       string       `json:"info"`
	URL        string       `json:"url"`
	Avatar     string       `json:"avatar"`
	Links      []cm.KvModel `json:"links"`
	Contact    []cm.KvModel `json:"contact"`
	Type       int64        `json:"type"`
	CreateTime cm.LocalTime `json:"create_time"`
	UpdateTime cm.LocalTime `json:"update_time"`
}

type GameV2SyncCreatorRow struct {
	ID         int64        `gorm:"column:id"`
	Name       string       `gorm:"column:name"`
	Info       string       `gorm:"column:info"`
	URL        string       `gorm:"column:url"`
	Avatar     string       `gorm:"column:avatar"`
	Links      *string      `gorm:"column:links"`
	Contact    *string      `gorm:"column:contact"`
	Type       int64        `gorm:"column:type"`
	CreateTime cm.LocalTime `gorm:"column:create_time"`
	UpdateTime cm.LocalTime `gorm:"column:update_time"`
}

type GameV2ListItem struct {
	ID          string            `json:"id"`
	AppID       string            `json:"appid"`
	Name        string            `json:"name"`
	Summary     string            `json:"summary"`
	HeaderURL   string            `json:"header_url"`
	CapsuleURL  string            `json:"capsule_url"`
	ReleaseDate string            `json:"release_date"`
	Developers  []string          `json:"developers"`
	Publishers  []string          `json:"publishers"`
	Platforms   map[string]bool   `json:"platforms"`
	Price       GameV2PriceView   `json:"price"`
	OnlineCount GameV2OnlineCount `json:"online_count"`
	Tags        []GameV2Tag       `json:"tags"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type GameV2PanelReadModel struct {
	LatestGames     []GameV2ListItem `json:"latest_games"`
	UpdatedGames    []GameV2ListItem `json:"updated_games"`
	TopOnline       []GameV2ListItem `json:"top_online"`
	FreeGames       []GameV2ListItem `json:"free_games"`
	HighestDiscount []GameV2ListItem `json:"highest_discount"`
	LowPrice        []GameV2ListItem `json:"low_price"`
	LatestNews      []GameV2NewsItem `json:"latest_news"`
}

type GameV2DetailReadModel struct {
	ID                  string                      `json:"id"`
	AppID               string                      `json:"appid"`
	RequestedLang       string                      `json:"requested_lang"`
	Lang                string                      `json:"lang"`
	Name                string                      `json:"name"`
	Summary             string                      `json:"summary"`
	Type                string                      `json:"type"`
	IsFree              bool                        `json:"is_free"`
	Website             string                      `json:"website"`
	HeaderURL           string                      `json:"header_url"`
	ShortDescription    string                      `json:"short_description"`
	DetailedDescription string                      `json:"detailed_description"`
	AboutTheGame        string                      `json:"about_the_game"`
	Release             GameV2Release               `json:"release"`
	Developers          []string                    `json:"developers"`
	Publishers          []string                    `json:"publishers"`
	Platforms           map[string]bool             `json:"platforms"`
	SupportedLanguages  string                      `json:"supported_languages"`
	SupportInfo         map[string]string           `json:"support_info"`
	Prices              []GameV2PriceView           `json:"prices"`
	Price               GameV2PriceView             `json:"price"`
	Media               GameV2MediaView             `json:"media"`
	Requirements        GameV2RequirementsView      `json:"requirements"`
	News                []GameV2NewsItem            `json:"news"`
	OnlineCount         GameV2OnlineCount           `json:"online_count"`
	Site                GameV2SiteInfo              `json:"site"`
	Tags                []GameV2Tag                 `json:"tags"`
	CollectedAt         time.Time                   `json:"collected_at"`
	UpdatedAt           time.Time                   `json:"updated_at"`
	Extra               GameV2ReadModelExtraPayload `json:"extra"`
}

type GameV2Release struct {
	ComingSoon bool   `json:"coming_soon"`
	Date       string `json:"date"`
}

type GameV2PriceView struct {
	Region            string    `json:"region"`
	Available         bool      `json:"available"`
	UnavailableReason string    `json:"unavailable_reason,omitempty"`
	IsFree            bool      `json:"is_free"`
	Currency          string    `json:"currency"`
	InitialAmount     int64     `json:"initial_amount"`
	FinalAmount       int64     `json:"final_amount"`
	DiscountPercent   int64     `json:"discount_percent"`
	InitialFormatted  string    `json:"initial_formatted"`
	FinalFormatted    string    `json:"final_formatted"`
	CollectedAt       time.Time `json:"collected_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type GameV2MediaView struct {
	HeaderURL        string             `json:"header_url"`
	CapsuleURL       string             `json:"capsule_url"`
	CapsuleV5URL     string             `json:"capsule_v5_url"`
	BackgroundURL    string             `json:"background_url"`
	BackgroundRawURL string             `json:"background_raw_url"`
	Screenshots      []GameV2Screenshot `json:"screenshots"`
	Movies           []GameV2Movie      `json:"movies"`
}

type GameV2Screenshot struct {
	ID           string `json:"id"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
}

type GameV2Movie struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
	Extra        any    `json:"extra,omitempty"`
}

type GameV2RequirementsView struct {
	PC    map[string]string `json:"pc"`
	Mac   map[string]string `json:"mac"`
	Linux map[string]string `json:"linux"`
}

type GameV2NewsItem struct {
	ID            string    `json:"id"`
	GameID        string    `json:"game_id"`
	AppID         string    `json:"appid"`
	Lang          string    `json:"lang"`
	GameName      string    `json:"game_name"`
	HeaderURL     string    `json:"header_url"`
	EventGID      string    `json:"event_gid"`
	Headline      string    `json:"headline"`
	Summary       string    `json:"summary"`
	PlainText     string    `json:"plain_text"`
	HTML          string    `json:"html"`
	URL           string    `json:"url"`
	Tags          []string  `json:"tags"`
	PublishedAt   time.Time `json:"published_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	CommentCount  int64     `json:"comment_count"`
	VoteUpCount   int64     `json:"vote_up_count"`
	VoteDownCount int64     `json:"vote_down_count"`
}

type GameV2NewsRow struct {
	GfgGameV2News
	GameName   string `gorm:"column:game_name" json:"game_name"`
	GameNameEn string `gorm:"column:game_name_en" json:"game_name_en"`
	HeaderURL  string `gorm:"column:header_url" json:"header_url"`
}

type GameV2OnlineCount struct {
	Count       int64     `json:"count"`
	Status      string    `json:"status"`
	CollectedAt time.Time `json:"collected_at"`
}

type GameV2SiteInfo struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Info       string       `json:"info"`
	Header     string       `json:"header"`
	ViewCount  int64        `json:"view_count"`
	Resources  []cm.KvModel `json:"resources"`
	Groups     []cm.KvModel `json:"groups"`
	Links      []cm.KvModel `json:"links"`
	CreateTime cm.LocalTime `json:"create_time"`
	UpdateTime cm.LocalTime `json:"update_time"`
}

type GameV2ReadModelExtraPayload struct {
	ContentDescriptors any `json:"content_descriptors,omitempty"`
	Ratings            any `json:"ratings,omitempty"`
}
