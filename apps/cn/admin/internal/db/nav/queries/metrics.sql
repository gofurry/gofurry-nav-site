-- name: AdminListNavMetricRegistry :many
SELECT metric_key, metric_version, metric_kind, entity_level, time_grain,
       source_facts, eligibility_policy, state_policy, coverage_policy,
       freshness_seconds, allowed_dimensions, status, description,
       created_at, retired_at
FROM gfn_metric_registry
ORDER BY metric_key, metric_version DESC;

-- name: AdminListNavMetricCheckpoints :many
SELECT registry.metric_key, registry.metric_version, registry.status,
       checkpoint.source_start_date, checkpoint.processed_through,
       target.processed_through AS target_processed_through,
       site.processed_through AS site_processed_through,
       checkpoint.created_at, checkpoint.updated_at
FROM gfn_metric_registry registry
JOIN gfn_metric_checkpoints checkpoint
  ON checkpoint.metric_key = registry.metric_key
 AND checkpoint.metric_version = registry.metric_version
LEFT JOIN gfn_fact_rollup_checkpoints target
  ON target.pipeline_key = 'nav.target_facts'
LEFT JOIN gfn_fact_rollup_checkpoints site
  ON site.pipeline_key = 'nav.site_facts'
ORDER BY registry.metric_key, registry.metric_version DESC;

-- name: AdminNavMetricOverview :many
SELECT registry.metric_key, registry.metric_version, registry.description,
       checkpoint.source_start_date, checkpoint.processed_through,
       target.processed_through AS target_processed_through,
       site.processed_through AS site_processed_through,
       latest.fact_date AS latest_fact_date,
       COALESCE(latest.population_count, 0)::bigint AS population_count,
       COALESCE(latest.eligible_count, 0)::bigint AS eligible_count,
       COALESCE(latest.not_applicable_count, 0)::bigint AS not_applicable_count,
       COALESCE(latest.positive_count, 0)::bigint AS positive_count,
       COALESCE(latest.negative_count, 0)::bigint AS negative_count,
       COALESCE(latest.stale_count, 0)::bigint AS stale_count,
       COALESCE(latest.not_probed_count, 0)::bigint AS not_probed_count,
       COALESCE(latest.probe_failed_count, 0)::bigint AS probe_failed_count,
       COALESCE(latest.unknown_count, 0)::bigint AS unknown_count
FROM gfn_metric_registry registry
JOIN gfn_metric_checkpoints checkpoint
  ON checkpoint.metric_key = registry.metric_key
 AND checkpoint.metric_version = registry.metric_version
LEFT JOIN gfn_fact_rollup_checkpoints target
  ON target.pipeline_key = 'nav.target_facts'
LEFT JOIN gfn_fact_rollup_checkpoints site
  ON site.pipeline_key = 'nav.site_facts'
LEFT JOIN LATERAL (
    SELECT daily.*
    FROM gfn_metric_daily daily
    WHERE daily.metric_key = registry.metric_key
      AND daily.metric_version = registry.metric_version
      AND daily.dimension_key = 'global'
      AND daily.dimension_value = 'all'
    ORDER BY daily.fact_date DESC
    LIMIT 1
) latest ON true
WHERE registry.status = 'active'
ORDER BY registry.metric_key;

-- name: AdminListNavMetricDaily :many
SELECT metric_key, metric_version, fact_date, dimension_key, dimension_value,
       population_count, eligible_count, not_applicable_count,
       positive_count, negative_count, stale_count, not_probed_count,
       probe_failed_count, unknown_count, computed_at
FROM gfn_metric_daily
WHERE (sqlc.arg(metric_key)::text = '' OR metric_key = sqlc.arg(metric_key))
  AND (sqlc.arg(metric_version)::integer = 0 OR metric_version = sqlc.arg(metric_version))
  AND (sqlc.narg(from_date)::date IS NULL OR fact_date >= sqlc.narg(from_date))
  AND (sqlc.narg(through_date)::date IS NULL OR fact_date <= sqlc.narg(through_date))
  AND dimension_key = sqlc.arg(dimension_key)
  AND dimension_value = sqlc.arg(dimension_value)
ORDER BY fact_date DESC, metric_key, metric_version DESC
LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: AdminCountNavMetricDaily :one
SELECT count(*)::bigint
FROM gfn_metric_daily
WHERE (sqlc.arg(metric_key)::text = '' OR metric_key = sqlc.arg(metric_key))
  AND (sqlc.arg(metric_version)::integer = 0 OR metric_version = sqlc.arg(metric_version))
  AND (sqlc.narg(from_date)::date IS NULL OR fact_date >= sqlc.narg(from_date))
  AND (sqlc.narg(through_date)::date IS NULL OR fact_date <= sqlc.narg(through_date))
  AND dimension_key = sqlc.arg(dimension_key)
  AND dimension_value = sqlc.arg(dimension_value);

-- name: AdminCountNavMetricEntities :one
SELECT count(*)::bigint
FROM gfn_metric_entity_daily entity
WHERE entity.metric_key = sqlc.arg(metric_key)
  AND entity.metric_version = sqlc.arg(metric_version)
  AND entity.fact_date = sqlc.arg(fact_date)
  AND (sqlc.arg(state)::text = '' OR entity.state = sqlc.arg(state))
  AND (sqlc.arg(reason_code)::text = '' OR entity.reason_code = sqlc.arg(reason_code));

-- name: AdminListNavMetricEntities :many
SELECT entity.site_id AS entity_id,
       COALESCE(NULLIF(fact.name, ''), NULLIF(fact.name_en, ''), 'Site #' || entity.site_id::text)::text AS historical_name,
       entity.state, entity.reason_code, entity.source_observed_at,
       entity.dimension_values, entity.source_projection_versions,
       entity.evaluated_at
FROM gfn_metric_entity_daily entity
LEFT JOIN gfn_site_daily fact
  ON fact.site_id = entity.site_id
 AND fact.fact_date = entity.fact_date
WHERE entity.metric_key = sqlc.arg(metric_key)
  AND entity.metric_version = sqlc.arg(metric_version)
  AND entity.fact_date = sqlc.arg(fact_date)
  AND (sqlc.arg(state)::text = '' OR entity.state = sqlc.arg(state))
  AND (sqlc.arg(reason_code)::text = '' OR entity.reason_code = sqlc.arg(reason_code))
ORDER BY entity.site_id
LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);
