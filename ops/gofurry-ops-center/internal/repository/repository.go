package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

type AlertInput struct {
	Key     string
	Region  string
	NodeID  string
	Level   string
	Type    string
	Title   string
	Message string
}

func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Close() {
	if r != nil && r.pool != nil {
		r.pool.Close()
	}
}

func (r *Repository) Migrate(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, schemaSQL)
	return err
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *Repository) Ingest(ctx context.Context, payload model.AgentPayload, alertThreshold int) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	reportedAt := payload.Timestamp
	if reportedAt.IsZero() {
		reportedAt = time.Now().UTC()
	}
	_, err = tx.Exec(ctx, `
INSERT INTO nodes (node_id, region, role, display_name, status, agent_version, last_seen_at, updated_at)
VALUES ($1, $2, $3, $4, 'ok', $5, $6, now())
ON CONFLICT (node_id) DO UPDATE
SET region = EXCLUDED.region,
    role = EXCLUDED.role,
    display_name = EXCLUDED.display_name,
    status = 'ok',
    agent_version = EXCLUDED.agent_version,
    last_seen_at = EXCLUDED.last_seen_at,
    updated_at = now()
`, payload.NodeID, payload.Region, payload.Role, payload.NodeName, payload.AgentVersion, reportedAt)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO node_heartbeats (node_id, region, agent_version, reported_at)
VALUES ($1, $2, $3, $4)
`, payload.NodeID, payload.Region, payload.AgentVersion, reportedAt)
	if err != nil {
		return err
	}
	if payload.System != nil {
		_, err = tx.Exec(ctx, `
INSERT INTO system_samples (node_id, cpu_usage, memory_usage, memory_used, memory_total, load1, load5, load15, uptime_seconds, reported_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`, payload.NodeID, payload.System.CPUUsage, payload.System.MemoryUsage, int64(payload.System.MemoryUsed), int64(payload.System.MemoryTotal), payload.System.Load1, payload.System.Load5, payload.System.Load15, int64(payload.System.UptimeSeconds), reportedAt)
		if err != nil {
			return err
		}
	}
	for _, item := range payload.Disks {
		if _, err = tx.Exec(ctx, `
INSERT INTO disk_samples (node_id, mount, usage, inode_usage, used, total, reported_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`, payload.NodeID, item.Mount, item.Usage, item.InodeUsage, int64(item.Used), int64(item.Total), reportedAt); err != nil {
			return err
		}
	}
	for _, item := range payload.Networks {
		if _, err = tx.Exec(ctx, `
INSERT INTO network_samples (node_id, name, bytes_sent, bytes_recv, packets_sent, packets_recv, reported_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`, payload.NodeID, item.Name, int64(item.BytesSent), int64(item.BytesRecv), int64(item.PacketsSent), int64(item.PacketsRecv), reportedAt); err != nil {
			return err
		}
	}
	for _, item := range payload.Docker {
		if _, err = tx.Exec(ctx, `
INSERT INTO docker_container_samples (node_id, name, running, status, health_status, restart_count, error_message, reported_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, payload.NodeID, item.Name, item.Running, item.Status, item.HealthStatus, item.RestartCount, item.ErrorMessage, reportedAt); err != nil {
			return err
		}
		status := "ok"
		message := item.Status
		if !item.Running || item.ErrorMessage != "" || strings.EqualFold(item.HealthStatus, "unhealthy") {
			status = "down"
			message = firstNonEmpty(item.ErrorMessage, item.HealthStatus, item.Status)
		}
		if _, err = upsertServiceStatus(ctx, tx, payload.NodeID, "docker", item.Name, status, message, 0); err != nil {
			return err
		}
	}
	for _, item := range payload.HTTPChecks {
		if _, err = tx.Exec(ctx, `
INSERT INTO http_check_results (node_id, name, url, status, status_code, latency_ms, error_message, reported_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, payload.NodeID, item.Name, item.URL, item.Status, item.StatusCode, item.LatencyMS, item.ErrorMessage, reportedAt); err != nil {
			return err
		}
		if _, err = upsertServiceStatus(ctx, tx, payload.NodeID, "http", item.Name, item.Status, item.ErrorMessage, item.LatencyMS); err != nil {
			return err
		}
	}
	for _, item := range payload.Postgres {
		if err = insertServiceCheck(ctx, tx, payload.NodeID, "postgres", item, reportedAt); err != nil {
			return err
		}
		if _, err = upsertServiceStatus(ctx, tx, payload.NodeID, "postgres", item.Name, item.Status, item.ErrorMessage, item.LatencyMS); err != nil {
			return err
		}
	}
	for _, item := range payload.Redis {
		if err = insertServiceCheck(ctx, tx, payload.NodeID, "redis", item, reportedAt); err != nil {
			return err
		}
		if _, err = upsertServiceStatus(ctx, tx, payload.NodeID, "redis", item.Name, item.Status, item.ErrorMessage, item.LatencyMS); err != nil {
			return err
		}
	}
	for _, item := range payload.Certs {
		if _, err = tx.Exec(ctx, `
INSERT INTO cert_check_results (node_id, name, host, status, expires_at, days_remaining, matched_name, error_message, reported_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`, payload.NodeID, item.Name, item.Host, item.Status, nullableTime(item.ExpiresAt), item.DaysRemaining, item.MatchedName, item.ErrorMessage, reportedAt); err != nil {
			return err
		}
		if _, err = upsertServiceStatus(ctx, tx, payload.NodeID, "cert", item.Name, item.Status, item.ErrorMessage, 0); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func insertServiceCheck(ctx context.Context, tx pgx.Tx, nodeID, typ string, item model.ServiceCheck, reportedAt time.Time) error {
	_, err := tx.Exec(ctx, `
INSERT INTO service_check_results (node_id, service_type, name, status, latency_ms, error_message, database_size, connections, memory_used, key_count, reported_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`, nodeID, typ, item.Name, item.Status, item.LatencyMS, item.ErrorMessage, item.DatabaseSize, item.Connections, item.MemoryUsed, item.KeyCount, reportedAt)
	return err
}

func upsertServiceStatus(ctx context.Context, tx pgx.Tx, nodeID, typ, name, status, message string, latency int64) (int, error) {
	key := serviceKey(nodeID, typ, name)
	var existing int
	err := tx.QueryRow(ctx, `SELECT failure_count FROM service_status WHERE key = $1`, key).Scan(&existing)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}
	nextFailure := 0
	var lastOK any
	if status == "ok" {
		lastOK = time.Now().UTC()
	} else {
		nextFailure = existing + 1
		lastOK = nil
	}
	_, err = tx.Exec(ctx, `
INSERT INTO service_status (key, node_id, service_type, name, status, message, failure_count, latency_ms, last_ok_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
ON CONFLICT (key) DO UPDATE
SET status = EXCLUDED.status,
    message = EXCLUDED.message,
    failure_count = EXCLUDED.failure_count,
    latency_ms = EXCLUDED.latency_ms,
    last_ok_at = COALESCE(EXCLUDED.last_ok_at, service_status.last_ok_at),
    updated_at = now()
`, key, nodeID, typ, name, status, message, nextFailure, latency, lastOK)
	return nextFailure, err
}

func (r *Repository) UpsertAlert(ctx context.Context, input AlertInput) error {
	tag, err := r.pool.Exec(ctx, `
UPDATE alert_states
SET level = $2, type = $3, title = $4, message = $5, status = 'active', last_seen_at = now(), resolved_at = NULL
WHERE key = $1
`, input.Key, input.Level, input.Type, input.Title, input.Message)
	if err != nil {
		return err
	}
	event := "updated"
	if tag.RowsAffected() == 0 {
		event = "created"
		_, err = r.pool.Exec(ctx, `
INSERT INTO alert_states (key, region, node_id, level, type, title, message, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'active')
`, input.Key, input.Region, input.NodeID, input.Level, input.Type, input.Title, input.Message)
		if err != nil {
			return err
		}
	}
	_, err = r.pool.Exec(ctx, `
INSERT INTO alert_events (alert_key, region, node_id, level, type, title, message, event)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, input.Key, input.Region, input.NodeID, input.Level, input.Type, input.Title, input.Message, event)
	return err
}

func (r *Repository) ResolveAlert(ctx context.Context, key string) error {
	tag, err := r.pool.Exec(ctx, `
UPDATE alert_states
SET status = 'resolved', resolved_at = now(), last_seen_at = now()
WHERE key = $1 AND status = 'active'
`, key)
	if err != nil || tag.RowsAffected() == 0 {
		return err
	}
	_, err = r.pool.Exec(ctx, `
INSERT INTO alert_events (alert_key, region, node_id, level, type, title, message, event)
SELECT key, region, node_id, level, type, title, message, 'resolved'
FROM alert_states WHERE key = $1
`, key)
	return err
}

func (r *Repository) ListNodes(ctx context.Context) ([]model.Node, error) {
	rows, err := r.pool.Query(ctx, `
SELECT node_id, region, role, display_name, status, agent_version, last_seen_at, updated_at
FROM nodes ORDER BY region, node_id
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.Node
	for rows.Next() {
		var item model.Node
		if err := rows.Scan(&item.NodeID, &item.Region, &item.Role, &item.DisplayName, &item.Status, &item.AgentVersion, &item.LastSeenAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetNode(ctx context.Context, nodeID string) (model.Node, error) {
	var item model.Node
	err := r.pool.QueryRow(ctx, `
SELECT node_id, region, role, display_name, status, agent_version, last_seen_at, updated_at
FROM nodes WHERE node_id = $1
`, nodeID).Scan(&item.NodeID, &item.Region, &item.Role, &item.DisplayName, &item.Status, &item.AgentVersion, &item.LastSeenAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) MarkNodeStatus(ctx context.Context, nodeID, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE nodes SET status = $2, updated_at = now() WHERE node_id = $1`, nodeID, status)
	return err
}

func (r *Repository) ListServiceStatuses(ctx context.Context) ([]model.ServiceStatus, error) {
	rows, err := r.pool.Query(ctx, `
SELECT key, node_id, service_type, name, status, message, failure_count, COALESCE(latency_ms, 0), last_ok_at, updated_at
FROM service_status ORDER BY service_type, node_id, name
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.ServiceStatus
	for rows.Next() {
		var item model.ServiceStatus
		if err := rows.Scan(&item.Key, &item.NodeID, &item.ServiceType, &item.Name, &item.Status, &item.Message, &item.FailureCount, &item.LatencyMS, &item.LastOKAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListAlerts(ctx context.Context, activeOnly bool) ([]model.AlertState, error) {
	query := `
SELECT key, region, node_id, level, type, title, message, status, first_seen_at, last_seen_at, resolved_at
FROM alert_states`
	if activeOnly {
		query += ` WHERE status = 'active'`
	}
	query += ` ORDER BY status, level, last_seen_at DESC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.AlertState
	for rows.Next() {
		var item model.AlertState
		if err := rows.Scan(&item.Key, &item.Region, &item.NodeID, &item.Level, &item.Type, &item.Title, &item.Message, &item.Status, &item.FirstSeenAt, &item.LastSeenAt, &item.ResolvedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) UpsertPeerSummary(ctx context.Context, item model.PeerSummary) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO peer_summaries (peer_region, peer_center_id, status, last_heartbeat_at, nodes_total, nodes_down, critical_alerts, warning_alerts, last_sync_status, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
ON CONFLICT (peer_region) DO UPDATE
SET peer_center_id = EXCLUDED.peer_center_id,
    status = EXCLUDED.status,
    last_heartbeat_at = EXCLUDED.last_heartbeat_at,
    nodes_total = EXCLUDED.nodes_total,
    nodes_down = EXCLUDED.nodes_down,
    critical_alerts = EXCLUDED.critical_alerts,
    warning_alerts = EXCLUDED.warning_alerts,
    last_sync_status = EXCLUDED.last_sync_status,
    updated_at = now()
`, item.Region, item.CenterID, item.Status, item.LastHeartbeatAt, item.NodesTotal, item.NodesDown, item.CriticalAlerts, item.WarningAlerts, item.LastSyncStatus)
	return err
}

func (r *Repository) LatestPeerSummary(ctx context.Context) (*model.PeerSummary, error) {
	var item model.PeerSummary
	err := r.pool.QueryRow(ctx, `
SELECT peer_region, peer_center_id, status, last_heartbeat_at, nodes_total, nodes_down, critical_alerts, warning_alerts, last_sync_status, updated_at
FROM peer_summaries ORDER BY updated_at DESC LIMIT 1
`).Scan(&item.Region, &item.CenterID, &item.Status, &item.LastHeartbeatAt, &item.NodesTotal, &item.NodesDown, &item.CriticalAlerts, &item.WarningAlerts, &item.LastSyncStatus, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &item, err
}

func (r *Repository) CreateSyncRun(ctx context.Context, req model.SyncEventRequest) (model.SyncRun, error) {
	var item model.SyncRun
	err := r.pool.QueryRow(ctx, `
INSERT INTO sync_runs (region, sync_name, version, status, items_total, checksum_ok, error_message, started_at, finished_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, region, sync_name, version, status, items_total, checksum_ok, error_message, started_at, finished_at, created_at
`, req.Region, req.SyncName, req.Version, req.Status, req.ItemsTotal, req.ChecksumOK, req.ErrorMessage, req.StartedAt, req.FinishedAt).Scan(&item.ID, &item.Region, &item.SyncName, &item.Version, &item.Status, &item.ItemsTotal, &item.ChecksumOK, &item.ErrorMessage, &item.StartedAt, &item.FinishedAt, &item.CreatedAt)
	return item, err
}

func (r *Repository) LatestSyncRun(ctx context.Context) (*model.SyncRun, error) {
	var item model.SyncRun
	err := r.pool.QueryRow(ctx, `
SELECT id, region, sync_name, version, status, items_total, checksum_ok, error_message, started_at, finished_at, created_at
FROM sync_runs ORDER BY created_at DESC, id DESC LIMIT 1
`).Scan(&item.ID, &item.Region, &item.SyncName, &item.Version, &item.Status, &item.ItemsTotal, &item.ChecksumOK, &item.ErrorMessage, &item.StartedAt, &item.FinishedAt, &item.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &item, err
}

func (r *Repository) ListSyncRuns(ctx context.Context, limit int) ([]model.SyncRun, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
SELECT id, region, sync_name, version, status, items_total, checksum_ok, error_message, started_at, finished_at, created_at
FROM sync_runs ORDER BY created_at DESC, id DESC LIMIT $1
`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.SyncRun
	for rows.Next() {
		var item model.SyncRun
		if err := rows.Scan(&item.ID, &item.Region, &item.SyncName, &item.Version, &item.Status, &item.ItemsTotal, &item.ChecksumOK, &item.ErrorMessage, &item.StartedAt, &item.FinishedAt, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateDeployEvent(ctx context.Context, req model.DeployEventRequest) (model.DeployEvent, error) {
	var item model.DeployEvent
	err := r.pool.QueryRow(ctx, `
INSERT INTO deploy_events (region, node_id, service_name, version, status, message)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, region, node_id, service_name, version, status, message, created_at
`, req.Region, req.NodeID, req.ServiceName, req.Version, req.Status, req.Message).Scan(&item.ID, &item.Region, &item.NodeID, &item.ServiceName, &item.Version, &item.Status, &item.Message, &item.CreatedAt)
	return item, err
}

func (r *Repository) ListDeployEvents(ctx context.Context, limit int) ([]model.DeployEvent, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
SELECT id, region, node_id, service_name, version, status, message, created_at
FROM deploy_events ORDER BY created_at DESC, id DESC LIMIT $1
`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.DeployEvent
	for rows.Next() {
		var item model.DeployEvent
		if err := rows.Scan(&item.ID, &item.Region, &item.NodeID, &item.ServiceName, &item.Version, &item.Status, &item.Message, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CleanupRawSamples(ctx context.Context, olderThan time.Time) error {
	for _, table := range []string{"node_heartbeats", "system_samples", "disk_samples", "network_samples", "docker_container_samples", "http_check_results", "service_check_results", "cert_check_results"} {
		column := "received_at"
		if table == "node_heartbeats" {
			column = "received_at"
		}
		if _, err := r.pool.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s < $1", table, column), olderThan); err != nil {
			return err
		}
	}
	return nil
}

func serviceKey(nodeID, typ, name string) string {
	return nodeID + ":" + typ + ":" + name
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
