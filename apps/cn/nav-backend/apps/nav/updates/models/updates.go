package models

import "time"

const (
	UpdatesSchemaVersion = 1

	UpdatesStateReady = "ready"
	UpdatesStateEmpty = "empty"
	UpdatesStateError = "error"

	TableNameGfnNavUpdateNotice = "gfn_nav_update_notice"
)

type UpdateNotice struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	TitleEn     string    `json:"title_en"`
	Body        string    `json:"body"`
	BodyEn      string    `json:"body_en"`
	PublishedAt time.Time `json:"published_at"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
	Deleted     bool      `json:"deleted"`
}

func (*UpdateNotice) TableName() string {
	return TableNameGfnNavUpdateNotice
}

type UpdateNoticeItem struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

type UpdatesResponse struct {
	SchemaVersion  int                `json:"schema_version"`
	GeneratedAt    time.Time          `json:"generated_at"`
	State          string             `json:"state"`
	ReasonMessages []string           `json:"reason_messages,omitempty"`
	Items          []UpdateNoticeItem `json:"items"`
}
