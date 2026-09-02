package workbench

import "time"

type Summary struct {
	Attention        []AttentionItem   `json:"attention"`
	RecentChanges    []RecentChange    `json:"recent_changes"`
	RecentOperations []RecentOperation `json:"recent_operations"`
	SystemStatus     []SystemStatus    `json:"system_status"`
}

type AttentionItem struct {
	Key     string `json:"key"`
	Tone    string `json:"tone"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Href    string `json:"href"`
}

type RecentChange struct {
	Domain         string  `json:"domain"`
	EventKey       string  `json:"event_key"`
	HistoricalName string  `json:"historical_name"`
	EventCode      string  `json:"event_code"`
	EventAt        *string `json:"event_at"`
	ProjectionDate string  `json:"projection_date"`
}

type RecentOperation struct {
	ID           int64     `json:"id"`
	Action       string    `json:"action"`
	Resource     string    `json:"resource"`
	TargetID     string    `json:"target_id"`
	OperatorName string    `json:"operator_name"`
	OperatorRole string    `json:"operator_role"`
	CreatedAt    time.Time `json:"created_at"`
}

type SystemStatus struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Status  string `json:"status"`
	Summary string `json:"summary"`
	Href    string `json:"href"`
}
