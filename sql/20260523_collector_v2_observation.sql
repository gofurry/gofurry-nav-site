-- GoFurry Nav Collector v2 observation schema.
--
-- IMPORTANT:
-- 1. This script is intended for manual production execution before deploying
--    the collector build that reads gfn_collector_domain.site_id/deleted.
-- 2. PostgreSQL does not allow CREATE INDEX CONCURRENTLY inside a transaction
--    block. Run this file with autocommit enabled, or execute the CREATE INDEX
--    statements one by one in a SQL client that does not wrap them in BEGIN.
-- 3. The collector does not run this migration automatically.

ALTER TABLE IF EXISTS gfn_collector_domain
ADD COLUMN IF NOT EXISTS site_id bigint;

ALTER TABLE IF EXISTS gfn_collector_domain
ADD COLUMN IF NOT EXISTS deleted boolean NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS gfn_collector_observation (
    id bigint PRIMARY KEY,
    site_id bigint NOT NULL,
    target varchar(255) NOT NULL,
    protocol varchar(16) NOT NULL,
    status varchar(32) NOT NULL,
    observed_at timestamptz NOT NULL,
    duration_ms bigint,
    error_code varchar(64),
    error_message text,
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    schema_version int NOT NULL DEFAULT 1,
    create_time timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gfn_collector_observation_protocol_site_time_id
ON gfn_collector_observation (protocol, site_id, observed_at DESC, id DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gfn_collector_observation_site_protocol_time_id
ON gfn_collector_observation (site_id, protocol, observed_at DESC, id DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gfn_collector_observation_protocol_time_id
ON gfn_collector_observation (protocol, observed_at DESC, id DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gfn_collector_domain_site_id
ON gfn_collector_domain (site_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gfn_collector_domain_deleted
ON gfn_collector_domain (deleted);
