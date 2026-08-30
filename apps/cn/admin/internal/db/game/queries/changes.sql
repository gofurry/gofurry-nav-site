-- name: AdminListGameChangeRegistry :many
SELECT detector_key, detector_version, source_kind, source_contracts,
       detection_policy, watermark_policy, event_codes, processing_grain,
       status, description, created_at, retired_at
FROM gfg_change_registry
ORDER BY detector_key, detector_version DESC;

-- name: AdminGameChangeOverview :many
SELECT registry.detector_key, registry.detector_version, registry.status,
       registry.description, registry.watermark_policy,
       checkpoint.source_start_date, checkpoint.processed_through,
       CASE WHEN registry.watermark_policy = 'metric_checkpoint_v1' THEN metric.processed_through
            WHEN registry.watermark_policy = 'fact_checkpoint_v1' THEN fact.processed_through
            WHEN registry.watermark_policy = 'closed_day_v1' THEN ((transaction_timestamp() AT TIME ZONE 'UTC')::date - 1)
       END::date AS upstream_processed_through,
       latest.projection_date AS latest_projection_date,
       COALESCE(latest.event_count, 0)::bigint AS latest_event_count,
       COALESCE(totals.event_count, 0)::bigint AS total_event_count
FROM gfg_change_registry registry
JOIN gfg_change_checkpoints checkpoint USING (detector_key, detector_version)
LEFT JOIN gfg_metric_checkpoints metric
  ON metric.metric_key = CASE registry.detector_key WHEN 'free_game_transition' THEN 'free_game_share' WHEN 'windows_support_transition' THEN 'windows_support' WHEN 'linux_support_transition' THEN 'linux_support' END
 AND metric.metric_version = registry.detector_version
LEFT JOIN gfg_fact_rollup_checkpoints fact
  ON fact.pipeline_key = CASE registry.detector_key WHEN 'game_price_transition' THEN 'game.state_facts' END
LEFT JOIN LATERAL (SELECT projection_date, count(*)::bigint AS event_count FROM gfg_change_events event WHERE event.detector_key=registry.detector_key AND event.detector_version=registry.detector_version GROUP BY projection_date ORDER BY projection_date DESC LIMIT 1) latest ON true
LEFT JOIN LATERAL (SELECT count(*)::bigint AS event_count FROM gfg_change_events event WHERE event.detector_key=registry.detector_key AND event.detector_version=registry.detector_version) totals ON true
WHERE registry.status='active'
ORDER BY registry.detector_key;

-- name: AdminListGameChangeCheckpoints :many
SELECT registry.detector_key, registry.detector_version, registry.status,
       registry.watermark_policy, checkpoint.source_start_date, checkpoint.processed_through,
       CASE WHEN registry.watermark_policy = 'metric_checkpoint_v1' THEN metric.processed_through
            WHEN registry.watermark_policy = 'fact_checkpoint_v1' THEN fact.processed_through
            WHEN registry.watermark_policy = 'closed_day_v1' THEN ((transaction_timestamp() AT TIME ZONE 'UTC')::date - 1)
       END::date AS upstream_processed_through,
       checkpoint.created_at, checkpoint.updated_at
FROM gfg_change_registry registry
JOIN gfg_change_checkpoints checkpoint USING (detector_key, detector_version)
LEFT JOIN gfg_metric_checkpoints metric
  ON metric.metric_key = CASE registry.detector_key WHEN 'free_game_transition' THEN 'free_game_share' WHEN 'windows_support_transition' THEN 'windows_support' WHEN 'linux_support_transition' THEN 'linux_support' END
 AND metric.metric_version = registry.detector_version
LEFT JOIN gfg_fact_rollup_checkpoints fact
  ON fact.pipeline_key = CASE registry.detector_key WHEN 'game_price_transition' THEN 'game.state_facts' END
ORDER BY registry.detector_key, registry.detector_version DESC;

-- name: AdminCountGameChangeEvents :one
SELECT count(*)::bigint FROM gfg_change_events event
WHERE (sqlc.arg(detector_key)::text='' OR event.detector_key=sqlc.arg(detector_key))
  AND (sqlc.arg(detector_version)::integer=0 OR event.detector_version=sqlc.arg(detector_version))
  AND (sqlc.narg(from_date)::date IS NULL OR event.projection_date>=sqlc.narg(from_date))
  AND (sqlc.narg(through_date)::date IS NULL OR event.projection_date<=sqlc.narg(through_date))
  AND (sqlc.arg(event_code)::text='' OR event.event_code=sqlc.arg(event_code))
  AND (sqlc.arg(scope_kind)::text='' OR event.scope_kind=sqlc.arg(scope_kind))
  AND (sqlc.arg(scope_key)::text='' OR event.scope_key=sqlc.arg(scope_key))
  AND (sqlc.arg(entity_id)::bigint=0 OR event.game_id=sqlc.arg(entity_id));

-- name: AdminListGameChangeEvents :many
SELECT event.event_key, event.detector_key, event.detector_version,
       event.game_id AS entity_id,
       COALESCE(NULLIF(fact.name,''), NULLIF(fact.name_en,''), 'Game #' || event.game_id::text)::text AS historical_name,
       event.projection_date, event.event_at, event.time_basis, event.event_code,
       event.scope_kind, event.scope_key, event.old_value, event.new_value,
       event.source_event_key, event.source_before_key, event.source_after_key,
       event.source_before_at, event.source_after_at, event.source_versions, event.materialized_at
FROM gfg_change_events event
LEFT JOIN gfg_game_daily fact ON fact.game_id=event.game_id AND fact.fact_date=event.projection_date
WHERE (sqlc.arg(detector_key)::text='' OR event.detector_key=sqlc.arg(detector_key))
  AND (sqlc.arg(detector_version)::integer=0 OR event.detector_version=sqlc.arg(detector_version))
  AND (sqlc.narg(from_date)::date IS NULL OR event.projection_date>=sqlc.narg(from_date))
  AND (sqlc.narg(through_date)::date IS NULL OR event.projection_date<=sqlc.narg(through_date))
  AND (sqlc.arg(event_code)::text='' OR event.event_code=sqlc.arg(event_code))
  AND (sqlc.arg(scope_kind)::text='' OR event.scope_kind=sqlc.arg(scope_kind))
  AND (sqlc.arg(scope_key)::text='' OR event.scope_key=sqlc.arg(scope_key))
  AND (sqlc.arg(entity_id)::bigint=0 OR event.game_id=sqlc.arg(entity_id))
ORDER BY event.projection_date DESC, event.event_at DESC NULLS LAST, event.event_key
LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);
