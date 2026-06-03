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
	ID          int64     `gorm:"column:id;type:bigint;primaryKey" json:"id"`
	Title       string    `gorm:"column:title;type:character varying(120);not null" json:"title"`
	TitleEn     string    `gorm:"column:title_en;type:character varying(120);not null" json:"title_en"`
	Body        string    `gorm:"column:body;type:text;not null" json:"body"`
	BodyEn      string    `gorm:"column:body_en;type:text;not null" json:"body_en"`
	PublishedAt time.Time `gorm:"column:published_at;type:timestamp(0) without time zone;not null" json:"published_at"`
	CreateTime  time.Time `gorm:"column:create_time;type:timestamp(0) without time zone;not null;autoCreateTime" json:"create_time"`
	UpdateTime  time.Time `gorm:"column:update_time;type:timestamp(0) without time zone;not null;autoUpdateTime" json:"update_time"`
	Deleted     bool      `gorm:"column:deleted;type:boolean;not null" json:"deleted"`
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
