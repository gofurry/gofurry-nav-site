-- name: ListGameChangeRegistry :many
SELECT detector_key, detector_version, source_kind, source_contracts,
       detection_policy, watermark_policy, event_codes, processing_grain,
       status, description, created_at, retired_at
FROM gfg_change_registry
ORDER BY detector_key, detector_version;

-- name: ListGameChangeCheckpoints :many
SELECT registry.detector_key,
       registry.detector_version,
       registry.status,
       registry.watermark_policy,
       checkpoint.source_start_date,
       checkpoint.processed_through,
       CASE
           WHEN registry.watermark_policy = 'metric_checkpoint_v1' THEN metric_checkpoint.processed_through
           WHEN registry.watermark_policy = 'fact_checkpoint_v1' THEN fact_checkpoint.processed_through
           WHEN registry.watermark_policy = 'closed_day_v1' THEN ((transaction_timestamp() AT TIME ZONE 'UTC')::date - 1)
       END::date AS upstream_processed_through,
       checkpoint.created_at,
       checkpoint.updated_at
FROM gfg_change_registry registry
JOIN gfg_change_checkpoints checkpoint
  ON checkpoint.detector_key = registry.detector_key
 AND checkpoint.detector_version = registry.detector_version
LEFT JOIN gfg_metric_checkpoints metric_checkpoint
  ON metric_checkpoint.metric_key = CASE registry.detector_key
      WHEN 'free_game_transition' THEN 'free_game_share'
      WHEN 'windows_support_transition' THEN 'windows_support'
      WHEN 'linux_support_transition' THEN 'linux_support'
      WHEN 'mac_support_transition' THEN 'mac_support'
  END
 AND metric_checkpoint.metric_version = registry.detector_version
LEFT JOIN gfg_fact_rollup_checkpoints fact_checkpoint
  ON fact_checkpoint.pipeline_key = CASE registry.detector_key
      WHEN 'game_price_transition' THEN 'game.state_facts'
  END
ORDER BY registry.detector_key, registry.detector_version;

-- name: LockGameChangeCheckpoint :one
SELECT checkpoint.detector_key,
       checkpoint.detector_version,
       checkpoint.source_start_date,
       checkpoint.processed_through,
       registry.status,
       registry.watermark_policy
FROM gfg_change_checkpoints checkpoint
JOIN gfg_change_registry registry
  ON registry.detector_key = checkpoint.detector_key
 AND registry.detector_version = checkpoint.detector_version
WHERE checkpoint.detector_key = sqlc.arg(detector_key)
  AND checkpoint.detector_version = sqlc.arg(detector_version)
FOR UPDATE OF checkpoint;

-- name: GameChangeUpstreamProcessedThrough :one
SELECT CASE
    WHEN registry.watermark_policy = 'metric_checkpoint_v1' THEN metric_checkpoint.processed_through
    WHEN registry.watermark_policy = 'fact_checkpoint_v1' THEN fact_checkpoint.processed_through
    WHEN registry.watermark_policy = 'closed_day_v1' THEN ((transaction_timestamp() AT TIME ZONE 'UTC')::date - 1)
END::date AS processed_through
FROM gfg_change_registry registry
LEFT JOIN gfg_metric_checkpoints metric_checkpoint
  ON metric_checkpoint.metric_key = CASE registry.detector_key
      WHEN 'free_game_transition' THEN 'free_game_share'
      WHEN 'windows_support_transition' THEN 'windows_support'
      WHEN 'linux_support_transition' THEN 'linux_support'
      WHEN 'mac_support_transition' THEN 'mac_support'
  END
 AND metric_checkpoint.metric_version = registry.detector_version
LEFT JOIN gfg_fact_rollup_checkpoints fact_checkpoint
  ON fact_checkpoint.pipeline_key = CASE registry.detector_key
      WHEN 'game_price_transition' THEN 'game.state_facts'
  END
WHERE registry.detector_key = sqlc.arg(detector_key)
  AND registry.detector_version = sqlc.arg(detector_version);

-- name: ProjectGameChangeDay :one
SELECT gfg_project_change_day(
    sqlc.arg(detector_key)::text,
    sqlc.arg(detector_version)::integer,
    sqlc.arg(projection_date)::date
)::bigint;

-- name: ProjectGameMacChangeDay :one
SELECT gfg_project_mac_change_day(sqlc.arg(projection_date)::date)::bigint;

-- name: AdvanceGameChangeCheckpoint :execrows
UPDATE gfg_change_checkpoints
SET processed_through = sqlc.arg(processed_through),
    updated_at = transaction_timestamp()
WHERE detector_key = sqlc.arg(detector_key)
  AND detector_version = sqlc.arg(detector_version)
  AND (processed_through IS NULL OR processed_through < sqlc.arg(processed_through));

-- name: CountGameChangeEventsForDay :one
SELECT count(*)::bigint
FROM gfg_change_events
WHERE detector_key = sqlc.arg(detector_key)
  AND detector_version = sqlc.arg(detector_version)
  AND projection_date = sqlc.arg(projection_date);
