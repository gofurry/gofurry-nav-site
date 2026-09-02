-- name: ListGameMetricRegistry :many
SELECT metric_key, metric_version, metric_kind, entity_level, time_grain,
       source_facts, eligibility_policy, state_policy, coverage_policy,
       freshness_seconds, allowed_dimensions, status, description,
       created_at, retired_at
FROM gfg_metric_registry
ORDER BY metric_key, metric_version;

-- name: ListGameMetricCheckpoints :many
SELECT registry.metric_key,
       registry.metric_version,
       registry.status,
       checkpoint.source_start_date,
       checkpoint.processed_through,
       upstream.processed_through AS upstream_processed_through,
       checkpoint.created_at,
       checkpoint.updated_at
FROM gfg_metric_registry registry
JOIN gfg_metric_checkpoints checkpoint
  ON checkpoint.metric_key = registry.metric_key
 AND checkpoint.metric_version = registry.metric_version
LEFT JOIN gfg_fact_rollup_checkpoints upstream
  ON upstream.pipeline_key = 'game.state_facts'
ORDER BY registry.metric_key, registry.metric_version;

-- name: LockGameMetricCheckpoint :one
SELECT checkpoint.metric_key,
       checkpoint.metric_version,
       checkpoint.source_start_date,
       checkpoint.processed_through,
       registry.status
FROM gfg_metric_checkpoints checkpoint
JOIN gfg_metric_registry registry
  ON registry.metric_key = checkpoint.metric_key
 AND registry.metric_version = checkpoint.metric_version
WHERE checkpoint.metric_key = sqlc.arg(metric_key)
  AND checkpoint.metric_version = sqlc.arg(metric_version)
FOR UPDATE OF checkpoint;

-- name: GameMetricUpstreamProcessedThrough :one
SELECT processed_through
FROM gfg_fact_rollup_checkpoints
WHERE pipeline_key = 'game.state_facts';

-- name: CountGameMetricPopulation :one
SELECT count(*)::bigint
FROM gfg_game_daily
WHERE fact_date = sqlc.arg(fact_date)
  AND finalized_at IS NOT NULL
  AND tracked_at_end;

-- name: ProjectGameMetricDay :one
SELECT gfg_project_metric_day(
    sqlc.arg(metric_key)::text,
    sqlc.arg(metric_version)::integer,
    sqlc.arg(fact_date)::date
)::bigint;

-- name: ProjectGameMacMetricDay :one
SELECT gfg_project_mac_metric_day(sqlc.arg(fact_date)::date)::bigint;

-- name: AdvanceGameMetricCheckpoint :execrows
UPDATE gfg_metric_checkpoints
SET processed_through = sqlc.arg(processed_through),
    updated_at = transaction_timestamp()
WHERE metric_key = sqlc.arg(metric_key)
  AND metric_version = sqlc.arg(metric_version)
  AND (processed_through IS NULL OR processed_through < sqlc.arg(processed_through));

-- name: GetGameMetricGlobalDaily :one
SELECT population_count, eligible_count, not_applicable_count,
       positive_count, negative_count, stale_count, not_probed_count,
       probe_failed_count, unknown_count, computed_at
FROM gfg_metric_daily
WHERE metric_key = sqlc.arg(metric_key)
  AND metric_version = sqlc.arg(metric_version)
  AND fact_date = sqlc.arg(fact_date)
  AND dimension_key = 'global'
  AND dimension_value = 'all';
