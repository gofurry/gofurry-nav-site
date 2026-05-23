-- GoFurry Nav Collector retention cleanup indexes.
--
-- IMPORTANT:
-- PostgreSQL does not allow CREATE INDEX CONCURRENTLY inside a transaction block.
-- Run this file with autocommit enabled. If your SQL client wraps scripts in a
-- transaction, execute the three CREATE INDEX statements one by one with
-- autocommit enabled, or use psql with:
--
--   psql "postgres://USER:PASSWORD@HOST:5432/DB?sslmode=disable" -v ON_ERROR_STOP=1 -f sql/20260523_collector_retention_indexes.sql
--
-- The index names intentionally match the current production/test schema so
-- IF NOT EXISTS is a true no-op when the indexes already exist.

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gfn_collector_log_ping_name_create_time_id
ON gfn_collector_log_ping (name, create_time DESC, id DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gfn_collector_log_http_name_create_time_id
ON gfn_collector_log_http (name, create_time DESC, id DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gfn_collector_log_dns_name_create_time_id
ON gfn_collector_log_dns (name, create_time DESC, id DESC);
