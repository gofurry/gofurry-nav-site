CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ping_name_time_id
ON gfn_collector_log_ping (name, create_time DESC, id DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_http_name_time_id
ON gfn_collector_log_http (name, create_time DESC, id DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_dns_name_time_id
ON gfn_collector_log_dns (name, create_time DESC, id DESC);
