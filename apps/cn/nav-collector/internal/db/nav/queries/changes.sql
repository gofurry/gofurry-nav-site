-- name: ListNavChangeRegistry :many
SELECT detector_key, detector_version, source_kind, source_contracts,
       detection_policy, watermark_policy, event_codes, processing_grain,
       status, description, created_at, retired_at
FROM gfn_change_registry
ORDER BY detector_key, detector_version;

-- name: ListNavChangeCheckpoints :many
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
FROM gfn_change_registry registry
JOIN gfn_change_checkpoints checkpoint
  ON checkpoint.detector_key = registry.detector_key
 AND checkpoint.detector_version = registry.detector_version
LEFT JOIN gfn_metric_checkpoints metric_checkpoint
  ON metric_checkpoint.metric_key = CASE registry.detector_key
      WHEN 'ipv6_transition' THEN 'ipv6_adoption'
      WHEN 'tls13_transition' THEN 'tls13_adoption'
      WHEN 'security_txt_transition' THEN 'security_txt_adoption'
  END
 AND metric_checkpoint.metric_version = registry.detector_version
LEFT JOIN gfn_fact_rollup_checkpoints fact_checkpoint
  ON fact_checkpoint.pipeline_key = CASE registry.detector_key
      WHEN 'tls_certificate_transition' THEN 'nav.target_facts'
  END
ORDER BY registry.detector_key, registry.detector_version;

-- name: LockNavChangeCheckpoint :one
SELECT checkpoint.detector_key,
       checkpoint.detector_version,
       checkpoint.source_start_date,
       checkpoint.processed_through,
       registry.status,
       registry.watermark_policy
FROM gfn_change_checkpoints checkpoint
JOIN gfn_change_registry registry
  ON registry.detector_key = checkpoint.detector_key
 AND registry.detector_version = checkpoint.detector_version
WHERE checkpoint.detector_key = sqlc.arg(detector_key)
  AND checkpoint.detector_version = sqlc.arg(detector_version)
FOR UPDATE OF checkpoint;

-- name: NavChangeUpstreamProcessedThrough :one
SELECT CASE
    WHEN registry.watermark_policy = 'metric_checkpoint_v1' THEN metric_checkpoint.processed_through
    WHEN registry.watermark_policy = 'fact_checkpoint_v1' THEN fact_checkpoint.processed_through
    WHEN registry.watermark_policy = 'closed_day_v1' THEN ((transaction_timestamp() AT TIME ZONE 'UTC')::date - 1)
END::date AS processed_through
FROM gfn_change_registry registry
LEFT JOIN gfn_metric_checkpoints metric_checkpoint
  ON metric_checkpoint.metric_key = CASE registry.detector_key
      WHEN 'ipv6_transition' THEN 'ipv6_adoption'
      WHEN 'tls13_transition' THEN 'tls13_adoption'
      WHEN 'security_txt_transition' THEN 'security_txt_adoption'
  END
 AND metric_checkpoint.metric_version = registry.detector_version
LEFT JOIN gfn_fact_rollup_checkpoints fact_checkpoint
  ON fact_checkpoint.pipeline_key = CASE registry.detector_key
      WHEN 'tls_certificate_transition' THEN 'nav.target_facts'
  END
WHERE registry.detector_key = sqlc.arg(detector_key)
  AND registry.detector_version = sqlc.arg(detector_version);

-- name: ProjectNavChangeDay :one
SELECT gfn_project_change_day(
    sqlc.arg(detector_key)::text,
    sqlc.arg(detector_version)::integer,
    sqlc.arg(projection_date)::date
)::bigint;

-- name: AdvanceNavChangeCheckpoint :execrows
UPDATE gfn_change_checkpoints
SET processed_through = sqlc.arg(processed_through),
    updated_at = transaction_timestamp()
WHERE detector_key = sqlc.arg(detector_key)
  AND detector_version = sqlc.arg(detector_version)
  AND (processed_through IS NULL OR processed_through < sqlc.arg(processed_through));

-- name: CountNavChangeEventsForDay :one
SELECT count(*)::bigint
FROM gfn_change_events
WHERE detector_key = sqlc.arg(detector_key)
  AND detector_version = sqlc.arg(detector_version)
  AND projection_date = sqlc.arg(projection_date);
