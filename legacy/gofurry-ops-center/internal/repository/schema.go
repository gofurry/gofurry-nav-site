package repository

const schemaSQL = `
CREATE TABLE IF NOT EXISTS nodes (
    node_id TEXT PRIMARY KEY,
    region TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT '',
    display_name TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'unknown',
    agent_version TEXT NOT NULL DEFAULT '',
    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS node_heartbeats (
    id BIGSERIAL PRIMARY KEY,
    node_id TEXT NOT NULL,
    region TEXT NOT NULL,
    agent_version TEXT NOT NULL DEFAULT '',
    reported_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_node_heartbeats_node_reported ON node_heartbeats(node_id, reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_node_heartbeats_reported ON node_heartbeats(reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_node_heartbeats_received ON node_heartbeats(received_at);

CREATE TABLE IF NOT EXISTS system_samples (
    id BIGSERIAL PRIMARY KEY,
    node_id TEXT NOT NULL,
    cpu_usage DOUBLE PRECISION,
    memory_usage DOUBLE PRECISION,
    memory_used BIGINT,
    memory_total BIGINT,
    load1 DOUBLE PRECISION,
    load5 DOUBLE PRECISION,
    load15 DOUBLE PRECISION,
    uptime_seconds BIGINT,
    reported_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_system_samples_node_reported ON system_samples(node_id, reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_system_samples_reported ON system_samples(reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_system_samples_received ON system_samples(received_at);

CREATE TABLE IF NOT EXISTS disk_samples (
    id BIGSERIAL PRIMARY KEY,
    node_id TEXT NOT NULL,
    mount TEXT NOT NULL,
    usage DOUBLE PRECISION,
    inode_usage DOUBLE PRECISION,
    used BIGINT,
    total BIGINT,
    reported_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_disk_samples_node_reported ON disk_samples(node_id, reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_disk_samples_reported ON disk_samples(reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_disk_samples_received ON disk_samples(received_at);

CREATE TABLE IF NOT EXISTS network_samples (
    id BIGSERIAL PRIMARY KEY,
    node_id TEXT NOT NULL,
    name TEXT NOT NULL,
    bytes_sent BIGINT,
    bytes_recv BIGINT,
    packets_sent BIGINT,
    packets_recv BIGINT,
    reported_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_network_samples_node_reported ON network_samples(node_id, reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_network_samples_reported ON network_samples(reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_network_samples_received ON network_samples(received_at);

CREATE TABLE IF NOT EXISTS docker_container_samples (
    id BIGSERIAL PRIMARY KEY,
    node_id TEXT NOT NULL,
    name TEXT NOT NULL,
    running BOOLEAN NOT NULL,
    status TEXT NOT NULL DEFAULT '',
    health_status TEXT NOT NULL DEFAULT '',
    restart_count INT NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    reported_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_docker_container_samples_node_reported ON docker_container_samples(node_id, reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_docker_container_samples_reported ON docker_container_samples(reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_docker_container_samples_received ON docker_container_samples(received_at);

CREATE TABLE IF NOT EXISTS http_check_results (
    id BIGSERIAL PRIMARY KEY,
    node_id TEXT NOT NULL,
    name TEXT NOT NULL,
    url TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    status_code INT,
    latency_ms BIGINT,
    error_message TEXT NOT NULL DEFAULT '',
    reported_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_http_check_results_node_reported ON http_check_results(node_id, reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_http_check_results_reported ON http_check_results(reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_http_check_results_received ON http_check_results(received_at);

CREATE TABLE IF NOT EXISTS service_check_results (
    id BIGSERIAL PRIMARY KEY,
    node_id TEXT NOT NULL,
    service_type TEXT NOT NULL,
    name TEXT NOT NULL,
    status TEXT NOT NULL,
    latency_ms BIGINT,
    error_message TEXT NOT NULL DEFAULT '',
    database_size BIGINT,
    connections BIGINT,
    memory_used BIGINT,
    key_count BIGINT,
    reported_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_service_check_results_node_reported ON service_check_results(node_id, reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_service_check_results_reported ON service_check_results(reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_service_check_results_received ON service_check_results(received_at);

CREATE TABLE IF NOT EXISTS cert_check_results (
    id BIGSERIAL PRIMARY KEY,
    node_id TEXT NOT NULL,
    name TEXT NOT NULL,
    host TEXT NOT NULL,
    status TEXT NOT NULL,
    expires_at TIMESTAMPTZ,
    days_remaining INT,
    matched_name BOOLEAN,
    error_message TEXT NOT NULL DEFAULT '',
    reported_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_cert_check_results_node_reported ON cert_check_results(node_id, reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_cert_check_results_reported ON cert_check_results(reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_cert_check_results_received ON cert_check_results(received_at);

CREATE TABLE IF NOT EXISTS service_status (
    key TEXT PRIMARY KEY,
    node_id TEXT NOT NULL,
    service_type TEXT NOT NULL,
    name TEXT NOT NULL,
    status TEXT NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    failure_count INT NOT NULL DEFAULT 0,
    latency_ms BIGINT,
    last_ok_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_service_status_node ON service_status(node_id, service_type);

CREATE TABLE IF NOT EXISTS alert_states (
    key TEXT PRIMARY KEY,
    region TEXT NOT NULL,
    node_id TEXT NOT NULL DEFAULT '',
    level TEXT NOT NULL,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_alert_states_status ON alert_states(status, level);

CREATE TABLE IF NOT EXISTS alert_events (
    id BIGSERIAL PRIMARY KEY,
    alert_key TEXT NOT NULL,
    region TEXT NOT NULL,
    node_id TEXT NOT NULL DEFAULT '',
    level TEXT NOT NULL,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    event TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS peer_summaries (
    peer_region TEXT PRIMARY KEY,
    peer_center_id TEXT NOT NULL,
    status TEXT NOT NULL,
    last_heartbeat_at TIMESTAMPTZ,
    nodes_total INT NOT NULL DEFAULT 0,
    nodes_down INT NOT NULL DEFAULT 0,
    critical_alerts INT NOT NULL DEFAULT 0,
    warning_alerts INT NOT NULL DEFAULT 0,
    last_sync_status TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sync_runs (
    id BIGSERIAL PRIMARY KEY,
    region TEXT NOT NULL,
    sync_name TEXT NOT NULL,
    version TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    items_total INT NOT NULL DEFAULT 0,
    checksum_ok BOOLEAN,
    error_message TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_sync_runs_region_name_created ON sync_runs(region, sync_name, created_at DESC);

CREATE TABLE IF NOT EXISTS deploy_events (
    id BIGSERIAL PRIMARY KEY,
    region TEXT NOT NULL,
    node_id TEXT NOT NULL DEFAULT '',
    service_name TEXT NOT NULL,
    version TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_deploy_events_created ON deploy_events(created_at DESC);
`
