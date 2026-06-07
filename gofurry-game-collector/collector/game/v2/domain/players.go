package domain

import "time"

// PlayerCount stores one current-player collection result.
type PlayerCount struct {
	GameID int64  `json:"game_id"`
	AppID  uint32 `json:"appid"`

	Count  int64  `json:"count"`
	Status Status `json:"status"`

	UpstreamStatusCode int    `json:"upstream_status_code"`
	ErrorKind          string `json:"error_kind"`
	ErrorMessage       string `json:"error_message"`

	CollectedAt time.Time `json:"collected_at"`
}
