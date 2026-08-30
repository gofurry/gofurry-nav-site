-- name: ListNavMetricRegistry :many
SELECT metric_key, metric_version, metric_kind, entity_level, time_grain,
       source_facts, eligibility_policy, state_policy, coverage_policy,
       freshness_seconds, allowed_dimensions, status, description,
       created_at, retired_at
FROM gfn_metric_registry
ORDER BY metric_key, metric_version;

-- name: ListNavMetricCheckpoints :many
SELECT registry.metric_key,
       registry.metric_version,
       registry.status,
       checkpoint.source_start_date,
       checkpoint.processed_through,
       target.processed_through AS target_processed_through,
       site.processed_through AS site_processed_through,
       checkpoint.created_at,
       checkpoint.updated_at
FROM gfn_metric_registry registry
JOIN gfn_metric_checkpoints checkpoint
  ON checkpoint.metric_key = registry.metric_key
 AND checkpoint.metric_version = registry.metric_version
LEFT JOIN gfn_fact_rollup_checkpoints target
  ON target.pipeline_key = 'nav.target_facts'
LEFT JOIN gfn_fact_rollup_checkpoints site
  ON site.pipeline_key = 'nav.site_facts'
ORDER BY registry.metric_key, registry.metric_version;

-- name: LockNavMetricCheckpoint :one
SELECT checkpoint.metric_key,
       checkpoint.metric_version,
       checkpoint.source_start_date,
       checkpoint.processed_through,
       registry.status
FROM gfn_metric_checkpoints checkpoint
JOIN gfn_metric_registry registry
  ON registry.metric_key = checkpoint.metric_key
 AND registry.metric_version = checkpoint.metric_version
WHERE checkpoint.metric_key = sqlc.arg(metric_key)
  AND checkpoint.metric_version = sqlc.arg(metric_version)
FOR UPDATE OF checkpoint;

-- name: NavMetricUpstreamProcessedThrough :one
SELECT target.processed_through AS target_processed_through,
       site.processed_through AS site_processed_through
FROM gfn_fact_rollup_checkpoints target
CROSS JOIN gfn_fact_rollup_checkpoints site
WHERE target.pipeline_key = 'nav.target_facts'
  AND site.pipeline_key = 'nav.site_facts';

-- name: CountNavMetricPopulation :one
SELECT count(*)::bigint
FROM gfn_site_daily
WHERE fact_date = sqlc.arg(fact_date)
  AND finalized_at IS NOT NULL
  AND tracked_at_end;

-- name: ProjectNavMetricDay :one
SELECT CASE
    WHEN sqlc.arg(metric_version)::integer = 2 THEN gfn_project_metric_day_v2(
        sqlc.arg(metric_key)::text,
        sqlc.arg(metric_version)::integer,
        sqlc.arg(fact_date)::date
    )
    ELSE gfn_project_metric_day(
        sqlc.arg(metric_key)::text,
        sqlc.arg(metric_version)::integer,
        sqlc.arg(fact_date)::date
    )
END::bigint;

-- name: AdvanceNavMetricCheckpoint :execrows
UPDATE gfn_metric_checkpoints
SET processed_through = sqlc.arg(processed_through),
    updated_at = transaction_timestamp()
WHERE metric_key = sqlc.arg(metric_key)
  AND metric_version = sqlc.arg(metric_version)
  AND (processed_through IS NULL OR processed_through < sqlc.arg(processed_through));

-- name: GetNavMetricGlobalDaily :one
SELECT population_count, eligible_count, not_applicable_count,
       positive_count, negative_count, stale_count, not_probed_count,
       probe_failed_count, unknown_count, computed_at
FROM gfn_metric_daily
WHERE metric_key = sqlc.arg(metric_key)
  AND metric_version = sqlc.arg(metric_version)
  AND fact_date = sqlc.arg(fact_date)
  AND dimension_key = 'global'
  AND dimension_value = 'all';
