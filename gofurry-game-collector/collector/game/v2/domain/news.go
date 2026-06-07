package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// GameNews is the collector v2 canonical Steam event news contract.
type GameNews struct {
	GameID   int64    `json:"game_id"`
	AppID    uint32   `json:"appid"`
	Language Language `json:"language"`

	EventGID        string `json:"event_gid"`
	AnnouncementGID string `json:"announcement_gid"`
	ForumTopicID    string `json:"forum_topic_id"`

	Headline  string `json:"headline"`
	RawBody   string `json:"raw_body"`
	HTML      string `json:"html"`
	PlainText string `json:"plain_text"`
	Summary   string `json:"summary"`
	URL       string `json:"url"`

	Tags          []string        `json:"tags"`
	VoteUpCount   int             `json:"vote_up_count"`
	VoteDownCount int             `json:"vote_down_count"`
	CommentCount  int             `json:"comment_count"`
	RawEvent      json.RawMessage `json:"raw_event"`

	PublishedAt time.Time `json:"published_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CollectedAt time.Time `json:"collected_at"`
}

// SteamNewsURL builds the canonical Steam news URL when the events payload omits one.
func SteamNewsURL(appID uint32, announcementGID string) string {
	if appID == 0 || announcementGID == "" {
		return ""
	}
	return fmt.Sprintf("https://store.steampowered.com/news/app/%d/view/%s", appID, announcementGID)
}
