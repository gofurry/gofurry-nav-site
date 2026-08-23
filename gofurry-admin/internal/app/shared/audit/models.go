package audit

import "time"

type AdminAuditLog struct {
	ID             int64     `json:"id"`
	Action         string    `json:"action"`
	Resource       string    `json:"resource"`
	TargetID       string    `json:"target_id"`
	Operator       string    `json:"operator"`
	SessionVersion int64     `json:"session_version"`
	RequestID      string    `json:"request_id"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	BeforeData     string    `json:"before_data"`
	AfterData      string    `json:"after_data"`
	CreatedAt      time.Time `json:"created_at"`
}

func (AdminAuditLog) TableName() string {
	return "gfa_admin_audit_log"
}

type Meta struct {
	Operator       string
	SessionVersion int64
	RequestID      string
	IPAddress      string
	UserAgent      string
}
