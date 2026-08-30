package dataops

import "time"

type Overview struct {
	GeneratedAt time.Time        `json:"generated_at"`
	Databases   []DatabaseStatus `json:"databases"`
}

type DatabaseStatus struct {
	Key               string          `json:"key"`
	Health            string          `json:"health"`
	PostgreSQLVersion string          `json:"postgresql_version"`
	DatabaseName      string          `json:"database_name"`
	DatabaseSizeBytes int64           `json:"database_size_bytes"`
	TotalConnections  int64           `json:"total_connections"`
	ActiveConnections int64           `json:"active_connections"`
	MaxConnections    int64           `json:"max_connections"`
	ConnectionUsage   *float64        `json:"connection_usage"`
	DatabaseTime      *time.Time      `json:"database_time"`
	Migration         MigrationStatus `json:"migration"`
	Relations         []RelationSize  `json:"relations"`
	Error             string          `json:"error,omitempty"`
}

type MigrationStatus struct {
	CurrentApplied int64  `json:"current_applied"`
	Expected       int64  `json:"expected"`
	PendingCount   int    `json:"pending_count"`
	Status         string `json:"status"`
}

type RelationSize struct {
	Name       string `json:"name"`
	TableBytes int64  `json:"table_bytes"`
	IndexBytes int64  `json:"index_bytes"`
	TotalBytes int64  `json:"total_bytes"`
}
