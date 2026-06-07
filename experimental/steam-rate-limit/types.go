package main

import (
	"time"

	steam "github.com/gofurry/steam-go"
)

type taskName string

const (
	taskAppDetails taskName = "appdetails"
	taskEvents     taskName = "events"
	taskPlayers    taskName = "players"
)

type bucketName string

const (
	bucketStore bucketName = "steam_store"
	bucketAPI   bucketName = "steam_api"
)

type config struct {
	RunID             string
	Tasks             []taskName
	AppIDs            []uint32
	Regions           []string
	Languages         []string
	Repeat            int
	Workers           int
	StoreInterval     time.Duration
	APIInterval       time.Duration
	Burst             int
	Timeout           time.Duration
	Retry             int
	RetryBaseDelay    time.Duration
	CooldownOnBlock   time.Duration
	ProgressInterval  time.Duration
	ProxyURLs         []string
	EventsCountBefore int
	EventsCountAfter  int
	OutputDir         string
	StopOnBlock       bool
	FailFast          bool
	PrintEveryResult  bool
}

type requestCase struct {
	Seq      int64      `json:"seq"`
	Task     taskName   `json:"task"`
	Bucket   bucketName `json:"bucket"`
	AppID    uint32     `json:"appid"`
	Region   string     `json:"region,omitempty"`
	Language string     `json:"language,omitempty"`
	Repeat   int        `json:"repeat"`
}

type observedEvent struct {
	TrafficClass  string        `json:"traffic_class"`
	Method        string        `json:"method"`
	Host          string        `json:"host"`
	Path          string        `json:"path"`
	StatusCode    int           `json:"status_code"`
	ErrorKind     string        `json:"error_kind,omitempty"`
	Attempts      int           `json:"attempts"`
	CacheHit      bool          `json:"cache_hit"`
	BlockDetected bool          `json:"block_detected"`
	Duration      time.Duration `json:"duration"`
	ObservedAt    time.Time     `json:"observed_at"`
}

type requestResult struct {
	Seq                int64         `json:"seq"`
	Task               taskName      `json:"task"`
	Bucket             bucketName    `json:"bucket"`
	AppID              uint32        `json:"appid"`
	Region             string        `json:"region,omitempty"`
	Language           string        `json:"language,omitempty"`
	Repeat             int           `json:"repeat"`
	StartedAt          time.Time     `json:"started_at"`
	EndedAt            time.Time     `json:"ended_at"`
	Duration           time.Duration `json:"duration"`
	OK                 bool          `json:"ok"`
	StatusCode         int           `json:"status_code"`
	ErrorKind          string        `json:"error_kind,omitempty"`
	ErrorMessage       string        `json:"error_message,omitempty"`
	Attempts           int           `json:"attempts"`
	CacheHit           bool          `json:"cache_hit"`
	BlockDetected      bool          `json:"block_detected"`
	ResponseBytes      int           `json:"response_bytes"`
	ResponseJSON       bool          `json:"response_json"`
	ResponseSuccessful *bool         `json:"response_successful,omitempty"`
	CooldownWait       time.Duration `json:"cooldown_wait"`
	EventPath          string        `json:"event_path,omitempty"`
	EventHost          string        `json:"event_host,omitempty"`
}

type runReport struct {
	RunID     string          `json:"run_id"`
	Config    reportConfig    `json:"config"`
	StartedAt time.Time       `json:"started_at"`
	EndedAt   time.Time       `json:"ended_at"`
	Duration  time.Duration   `json:"duration"`
	Results   []requestResult `json:"results"`
	Events    []observedEvent `json:"events"`
	Summary   reportSummary   `json:"summary"`
	Generated []string        `json:"generated"`
}

type reportConfig struct {
	Tasks             []taskName    `json:"tasks"`
	AppIDs            []uint32      `json:"appids"`
	Regions           []string      `json:"regions"`
	Languages         []string      `json:"languages"`
	Repeat            int           `json:"repeat"`
	Workers           int           `json:"workers"`
	StoreInterval     time.Duration `json:"store_interval"`
	APIInterval       time.Duration `json:"api_interval"`
	Burst             int           `json:"burst"`
	Timeout           time.Duration `json:"timeout"`
	Retry             int           `json:"retry"`
	RetryBaseDelay    time.Duration `json:"retry_base_delay"`
	CooldownOnBlock   time.Duration `json:"cooldown_on_block"`
	ProgressInterval  time.Duration `json:"progress_interval"`
	ProxyConfigured   bool          `json:"proxy_configured"`
	EventsCountBefore int           `json:"events_count_before"`
	EventsCountAfter  int           `json:"events_count_after"`
}

type reportSummary struct {
	Total              int                        `json:"total"`
	OK                 int                        `json:"ok"`
	Failed             int                        `json:"failed"`
	Blocked            int                        `json:"blocked"`
	HTTP429            int                        `json:"http_429"`
	HTTP403            int                        `json:"http_403"`
	HTTP5xx            int                        `json:"http_5xx"`
	TimeoutOrTransport int                        `json:"timeout_or_transport"`
	ByTask             map[taskName]taskSummary   `json:"by_task"`
	ByBucket           map[bucketName]taskSummary `json:"by_bucket"`
	Recommendation     string                     `json:"recommendation"`
}

type taskSummary struct {
	Total       int           `json:"total"`
	OK          int           `json:"ok"`
	Failed      int           `json:"failed"`
	Blocked     int           `json:"blocked"`
	AvgDuration time.Duration `json:"avg_duration"`
	P95Duration time.Duration `json:"p95_duration"`
}

func bucketForTrafficClass(class steam.TrafficClass) bucketName {
	switch class {
	case steam.TrafficClassOfficialAPI:
		return bucketAPI
	case steam.TrafficClassPublicStorePage:
		return bucketStore
	default:
		return ""
	}
}
