package models

import "time"

const (
	PageViewSchemaVersion = 1
	PageViewStateReady    = "ready"
	PageViewStateError    = "error"
)

type PageViewResponse struct {
	SchemaVersion int       `json:"schema_version"`
	GeneratedAt   time.Time `json:"generated_at"`
	State         string    `json:"state"`
	Page          string    `json:"page"`
	ViewCount     int64     `json:"view_count"`
}
