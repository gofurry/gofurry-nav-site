package observation

import "time"

const (
	ProtocolPing = "ping"
	ProtocolHTTP = "http"
	ProtocolDNS  = "dns"

	StatusSuccess = "success"
	StatusFailure = "failure"

	TableNameGfnCollectorObservation = "gfn_collector_observation"
)

type GfnCollectorObservation struct {
	ID            int64     `gorm:"column:id;type:bigint;primaryKey" json:"id"`
	SiteID        int64     `gorm:"column:site_id;type:bigint;not null" json:"site_id"`
	Target        string    `gorm:"column:target;type:character varying(255);not null" json:"target"`
	Protocol      string    `gorm:"column:protocol;type:character varying(16);not null" json:"protocol"`
	Status        string    `gorm:"column:status;type:character varying(32);not null" json:"status"`
	ObservedAt    time.Time `gorm:"column:observed_at;type:timestamptz;not null" json:"observed_at"`
	DurationMS    int64     `gorm:"column:duration_ms;type:bigint" json:"duration_ms"`
	ErrorCode     *string   `gorm:"column:error_code;type:character varying(64)" json:"error_code,omitempty"`
	ErrorMessage  *string   `gorm:"column:error_message;type:text" json:"error_message,omitempty"`
	Payload       string    `gorm:"column:payload;type:jsonb;not null" json:"payload"`
	SchemaVersion int       `gorm:"column:schema_version;type:int;not null" json:"schema_version"`
	CreateTime    time.Time `gorm:"column:create_time;type:timestamptz;not null;autoCreateTime" json:"create_time"`
}

func (*GfnCollectorObservation) TableName() string {
	return TableNameGfnCollectorObservation
}

type Input struct {
	SiteID       int64
	Target       string
	Protocol     string
	Status       string
	ObservedAt   time.Time
	DurationMS   int64
	ErrorCode    string
	ErrorMessage string
	Payload      any
}

type LatestDocument struct {
	SiteID        int64     `json:"site_id"`
	Target        string    `json:"target"`
	Protocol      string    `json:"protocol"`
	Status        string    `json:"status"`
	ObservedAt    time.Time `json:"observed_at"`
	DurationMS    int64     `json:"duration_ms"`
	ErrorCode     string    `json:"error_code,omitempty"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	Payload       any       `json:"payload"`
	SchemaVersion int       `json:"schema_version"`
}
