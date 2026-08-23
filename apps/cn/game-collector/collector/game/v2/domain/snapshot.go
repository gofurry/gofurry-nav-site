package domain

import (
	"encoding/json"
	"time"
)

// SnapshotKind identifies one raw upstream payload family.
type SnapshotKind string

const (
	SnapshotDetails SnapshotKind = "details"
	SnapshotNews    SnapshotKind = "news"
)

// RawSnapshot preserves bounded raw upstream payloads in PostgreSQL jsonb.
type RawSnapshot struct {
	ID int64 `json:"id"`

	GameID int64        `json:"game_id"`
	AppID  uint32       `json:"appid"`
	Kind   SnapshotKind `json:"kind"`

	Language Language `json:"language"`
	Region   Region   `json:"region"`
	Source   Source   `json:"source"`

	PayloadHash string          `json:"payload_hash"`
	RawPayload  json.RawMessage `json:"raw_payload"`

	CollectedAt time.Time `json:"collected_at"`
}
