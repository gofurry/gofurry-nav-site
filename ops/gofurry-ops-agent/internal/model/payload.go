package model

import "time"

type Payload struct {
	NodeID       string            `json:"node_id"`
	NodeName     string            `json:"node_name,omitempty"`
	Region       string            `json:"region"`
	Role         string            `json:"role,omitempty"`
	Timestamp    time.Time         `json:"timestamp"`
	AgentVersion string            `json:"agent_version"`
	System       *SystemSample     `json:"system,omitempty"`
	Disks        []DiskSample      `json:"disks,omitempty"`
	Networks     []NetworkSample   `json:"networks,omitempty"`
	Docker       []DockerSample    `json:"docker,omitempty"`
	HTTPChecks   []HTTPCheckResult `json:"http_checks,omitempty"`
	Postgres     []ServiceCheck    `json:"postgres,omitempty"`
	Redis        []ServiceCheck    `json:"redis,omitempty"`
	Certs        []CertCheckResult `json:"certs,omitempty"`
	Errors       []string          `json:"errors,omitempty"`
}

type SystemSample struct {
	CPUUsage      float64 `json:"cpu_usage"`
	MemoryUsage   float64 `json:"memory_usage"`
	MemoryUsed    uint64  `json:"memory_used"`
	MemoryTotal   uint64  `json:"memory_total"`
	Load1         float64 `json:"load1"`
	Load5         float64 `json:"load5"`
	Load15        float64 `json:"load15"`
	UptimeSeconds uint64  `json:"uptime_seconds"`
	ProcessCount  uint64  `json:"process_count,omitempty"`
	BootTime      uint64  `json:"boot_time,omitempty"`
}

type DiskSample struct {
	Mount      string  `json:"mount"`
	Usage      float64 `json:"usage"`
	InodeUsage float64 `json:"inode_usage"`
	Used       uint64  `json:"used"`
	Total      uint64  `json:"total"`
}

type NetworkSample struct {
	Name        string `json:"name"`
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

type DockerSample struct {
	Name         string  `json:"name"`
	ID           string  `json:"id,omitempty"`
	Running      bool    `json:"running"`
	Status       string  `json:"status,omitempty"`
	HealthStatus string  `json:"health_status,omitempty"`
	RestartCount int     `json:"restart_count"`
	CPUPercent   float64 `json:"cpu_percent,omitempty"`
	MemoryUsage  uint64  `json:"memory_usage,omitempty"`
	MemoryLimit  uint64  `json:"memory_limit,omitempty"`
	ErrorMessage string  `json:"error_message,omitempty"`
}

type HTTPCheckResult struct {
	Name         string `json:"name"`
	URL          string `json:"url"`
	Status       string `json:"status"`
	StatusCode   int    `json:"status_code,omitempty"`
	LatencyMS    int64  `json:"latency_ms"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type ServiceCheck struct {
	Name         string `json:"name"`
	Status       string `json:"status"`
	LatencyMS    int64  `json:"latency_ms"`
	ErrorMessage string `json:"error_message,omitempty"`
	DatabaseSize int64  `json:"database_size,omitempty"`
	Connections  int64  `json:"connections,omitempty"`
	MemoryUsed   int64  `json:"memory_used,omitempty"`
	KeyCount     int64  `json:"key_count,omitempty"`
}

type CertCheckResult struct {
	Name          string    `json:"name"`
	Host          string    `json:"host"`
	Status        string    `json:"status"`
	ExpiresAt     time.Time `json:"expires_at,omitempty"`
	DaysRemaining int       `json:"days_remaining"`
	MatchedName   bool      `json:"matched_name"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}
