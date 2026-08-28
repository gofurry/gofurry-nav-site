-- name: ListNavFactCheckpoints :many
SELECT pipeline_key, projection_version, source_start_date, processed_through,
       quality_cutover_at, created_at, updated_at
FROM gfn_fact_rollup_checkpoints
ORDER BY pipeline_key;

-- name: LockNavFactCheckpoint :one
SELECT pipeline_key, projection_version, source_start_date, processed_through,
       quality_cutover_at, created_at, updated_at
FROM gfn_fact_rollup_checkpoints
WHERE pipeline_key = sqlc.arg(pipeline_key)
FOR UPDATE;

-- name: NavTargetFactDaySettled :one
SELECT NOT EXISTS (
    SELECT 1 FROM gfn_collection_jobs job
    WHERE job.trigger = 'scheduled'
      AND job.job_key LIKE 'nav.%'
      AND job.scheduled_for >= sqlc.arg(day_start)
      AND job.scheduled_for < sqlc.arg(day_end)
      AND job.status IN ('queued', 'running')
)
AND NOT EXISTS (
    SELECT 1 FROM gfn_collection_schedules schedule
    WHERE schedule.job_key LIKE 'nav.%'
      AND schedule.enabled
      AND (schedule.next_scheduled_for IS NULL OR schedule.next_scheduled_for < sqlc.arg(day_end))
) AS settled;

-- name: NavSiteFactDependencyReady :one
SELECT processed_through IS NOT NULL AND processed_through >= sqlc.arg(fact_date)::date AS ready
FROM gfn_fact_rollup_checkpoints
WHERE pipeline_key = 'nav.target_facts';

-- name: ListNavProtocolDailyInputs :many
WITH protocols(protocol) AS (
    SELECT unnest(ARRAY[
        'ping', 'http', 'dns', 'rdap', 'robots', 'security_txt',
        'llms_txt', 'page_assets', 'port_check', 'waf_canary'
    ]::text[])
), eligible AS (
    SELECT period.*
    FROM gfn_target_tracking_periods period
    WHERE period.tracked_from < sqlc.arg(day_end)
      AND (period.tracked_until IS NULL OR period.tracked_until > sqlc.arg(day_start))
), jobs AS (
    SELECT job.*, job.tasks[1] AS protocol
    FROM gfn_collection_jobs job
    WHERE job.trigger = 'scheduled'
      AND job.job_key LIKE 'nav.%'
      AND job.scheduled_for >= sqlc.arg(day_start)
      AND job.scheduled_for < sqlc.arg(day_end)
), latest_runs AS (
    SELECT DISTINCT ON (run.job_id)
           run.job_id, run.id AS run_id, run.status AS run_status, run.attempt_no
    FROM gfn_collection_runs run
    JOIN jobs ON jobs.id = run.job_id
    ORDER BY run.job_id, run.attempt_no DESC, run.started_at DESC
), latest_results AS (
    SELECT DISTINCT ON (run.job_id, result.site_id, result.target)
           run.job_id, result.site_id, result.target, result.status,
           NULLIF(result.error_kind, '') AS error_kind,
           result.duration_ms, result.started_at, result.ended_at,
           run.attempt_no
    FROM gfn_collection_runs run
    JOIN gfn_collection_task_results result ON result.run_id = run.id
    JOIN jobs ON jobs.id = run.job_id AND jobs.protocol = result.protocol
    ORDER BY run.job_id, result.site_id, result.target, run.attempt_no DESC,
             result.started_at DESC, result.id DESC
), slots AS (
    SELECT eligible.id AS target_tracking_period_id,
           jobs.protocol,
           jobs.scheduled_for,
           CASE
               WHEN result.status = 'success' THEN 'success'
               WHEN result.status = 'partial' THEN 'partial'
               WHEN result.status = 'failed' THEN 'failure'
               WHEN result.status = 'skipped' THEN 'skipped'
               WHEN jobs.status = 'missed' THEN 'missed'
               WHEN jobs.status = 'skipped' THEN 'skipped'
               WHEN jobs.status = 'canceled' OR run.run_status = 'canceled' THEN 'canceled'
               ELSE 'unattempted'
           END AS outcome,
           result.error_kind,
           result.duration_ms
    FROM eligible
    JOIN jobs
      ON jobs.protocol IN (SELECT protocol FROM protocols)
     AND eligible.tracked_from <= jobs.scheduled_for
     AND (eligible.tracked_until IS NULL OR jobs.scheduled_for < eligible.tracked_until)
     AND (jobs.scope_type = 'all'
          OR (jobs.scope_type = 'site' AND jobs.scope_id = eligible.site_id)
          OR (jobs.scope_type = 'target' AND jobs.scope_id = eligible.site_id
              AND lower(jobs.target) = lower(eligible.target)))
    LEFT JOIN latest_runs run ON run.job_id = jobs.id
    LEFT JOIN latest_results result
      ON result.job_id = jobs.id
     AND result.site_id = eligible.site_id
     AND lower(result.target) = lower(eligible.target)
), quality AS (
    SELECT target_tracking_period_id, protocol,
           count(*)::integer AS expected_count,
           count(*) FILTER (WHERE outcome IN ('success', 'partial', 'failure'))::integer AS attempted_count,
           count(*) FILTER (WHERE outcome = 'success')::integer AS success_count,
           count(*) FILTER (WHERE outcome = 'partial')::integer AS partial_count,
           count(*) FILTER (WHERE outcome = 'failure')::integer AS failure_count,
           count(*) FILTER (WHERE outcome = 'skipped')::integer AS skipped_count,
           count(*) FILTER (WHERE outcome = 'missed')::integer AS missed_count,
           count(*) FILTER (WHERE outcome = 'canceled')::integer AS canceled_count,
           count(*) FILTER (WHERE outcome = 'unattempted')::integer AS unattempted_count,
           avg(duration_ms::double precision) FILTER (WHERE outcome IN ('success', 'partial', 'failure')) AS avg_duration_ms,
           percentile_cont(0.95) WITHIN GROUP (ORDER BY duration_ms)
               FILTER (WHERE outcome IN ('success', 'partial', 'failure')) AS p95_duration_ms
    FROM slots
    GROUP BY target_tracking_period_id, protocol
), failures AS (
    SELECT target_tracking_period_id, protocol,
           jsonb_object_agg(error_kind, failure_count ORDER BY error_kind) AS failure_kind_counts
    FROM (
        SELECT target_tracking_period_id, protocol,
               COALESCE(error_kind, 'unknown') AS error_kind,
               count(*)::bigint AS failure_count
        FROM slots
        WHERE outcome = 'failure'
        GROUP BY target_tracking_period_id, protocol, COALESCE(error_kind, 'unknown')
    ) grouped
    GROUP BY target_tracking_period_id, protocol
)
SELECT eligible.id AS target_tracking_period_id,
       eligible.site_id,
       eligible.collector_domain_id,
       eligible.target,
       protocols.protocol::text AS protocol,
       COALESCE(quality.expected_count, 0)::integer AS expected_count,
       COALESCE(quality.attempted_count, 0)::integer AS attempted_count,
       COALESCE(quality.success_count, 0)::integer AS success_count,
       COALESCE(quality.partial_count, 0)::integer AS partial_count,
       COALESCE(quality.failure_count, 0)::integer AS failure_count,
       COALESCE(quality.skipped_count, 0)::integer AS skipped_count,
       COALESCE(quality.missed_count, 0)::integer AS missed_count,
       COALESCE(quality.canceled_count, 0)::integer AS canceled_count,
       COALESCE(quality.unattempted_count, 0)::integer AS unattempted_count,
       COALESCE(failures.failure_kind_counts, '{}'::jsonb) AS failure_kind_counts,
       COALESCE(latest_scheduled.outcome, '')::text AS latest_scheduled_status,
       latest_scheduled.scheduled_for AS latest_scheduled_at,
       COALESCE(latest_observation.status, '')::text AS latest_observation_status,
       latest_observation.observed_at AS latest_observation_at,
       latest_success.observed_at AS known_state_observed_at,
       latest_success.payload AS known_payload,
       quality.avg_duration_ms,
       quality.p95_duration_ms
FROM eligible
CROSS JOIN protocols
LEFT JOIN quality
  ON quality.target_tracking_period_id = eligible.id
 AND quality.protocol = protocols.protocol
LEFT JOIN failures
  ON failures.target_tracking_period_id = eligible.id
 AND failures.protocol = protocols.protocol
LEFT JOIN LATERAL (
    SELECT slot.outcome, slot.scheduled_for
    FROM slots slot
    WHERE slot.target_tracking_period_id = eligible.id
      AND slot.protocol = protocols.protocol
    ORDER BY slot.scheduled_for DESC
    LIMIT 1
) latest_scheduled ON true
LEFT JOIN LATERAL (
    SELECT observation.status, observation.observed_at
    FROM gfn_collector_observation observation
    WHERE observation.site_id = eligible.site_id
      AND lower(observation.target) = lower(eligible.target)
      AND observation.protocol = protocols.protocol
      AND observation.observed_at >= sqlc.arg(day_start)
      AND observation.observed_at < sqlc.arg(day_end)
      AND observation.observed_at >= eligible.tracked_from
      AND (eligible.tracked_until IS NULL OR observation.observed_at < eligible.tracked_until)
    ORDER BY observation.observed_at DESC, observation.id DESC
    LIMIT 1
) latest_observation ON true
LEFT JOIN LATERAL (
    SELECT observation.observed_at, observation.payload
    FROM gfn_collector_observation observation
    WHERE observation.site_id = eligible.site_id
      AND lower(observation.target) = lower(eligible.target)
      AND observation.protocol = protocols.protocol
      AND observation.status = 'success'
      AND observation.observed_at < sqlc.arg(day_end)
      AND observation.observed_at >= eligible.tracked_from
      AND (eligible.tracked_until IS NULL OR observation.observed_at < eligible.tracked_until)
    ORDER BY observation.observed_at DESC, observation.id DESC
    LIMIT 1
) latest_success ON true
ORDER BY eligible.id, protocols.protocol;

-- name: UpsertNavProtocolDaily :exec
INSERT INTO gfn_site_target_protocol_daily (
    target_tracking_period_id, site_id, collector_domain_id, target, protocol,
    fact_date, expected_count, attempted_count, success_count, partial_count,
    failure_count, skipped_count, missed_count, canceled_count,
    unattempted_count, failure_kind_counts, quality_basis,
    latest_scheduled_status, latest_scheduled_at, latest_observation_status,
    latest_observation_at, known_state_observed_at, known_state,
    avg_duration_ms, p95_duration_ms, projection_version, finalized_at,
    created_at, updated_at
)
VALUES (
    sqlc.arg(target_tracking_period_id), sqlc.arg(site_id),
    sqlc.narg(collector_domain_id), sqlc.arg(target), sqlc.arg(protocol),
    sqlc.arg(fact_date), sqlc.narg(expected_count), sqlc.arg(attempted_count),
    sqlc.arg(success_count), sqlc.arg(partial_count), sqlc.arg(failure_count),
    sqlc.narg(skipped_count), sqlc.narg(missed_count), sqlc.narg(canceled_count),
    sqlc.narg(unattempted_count), sqlc.arg(failure_kind_counts)::jsonb,
    sqlc.arg(quality_basis), sqlc.narg(latest_scheduled_status),
    sqlc.narg(latest_scheduled_at), sqlc.narg(latest_observation_status),
    sqlc.narg(latest_observation_at), sqlc.narg(known_state_observed_at),
    sqlc.narg(known_state)::jsonb, sqlc.narg(avg_duration_ms),
    sqlc.narg(p95_duration_ms), 1, transaction_timestamp(),
    transaction_timestamp(), transaction_timestamp()
)
ON CONFLICT (target_tracking_period_id, protocol, fact_date) DO UPDATE
SET site_id = EXCLUDED.site_id,
    collector_domain_id = EXCLUDED.collector_domain_id,
    target = EXCLUDED.target,
    expected_count = EXCLUDED.expected_count,
    attempted_count = EXCLUDED.attempted_count,
    success_count = EXCLUDED.success_count,
    partial_count = EXCLUDED.partial_count,
    failure_count = EXCLUDED.failure_count,
    skipped_count = EXCLUDED.skipped_count,
    missed_count = EXCLUDED.missed_count,
    canceled_count = EXCLUDED.canceled_count,
    unattempted_count = EXCLUDED.unattempted_count,
    failure_kind_counts = EXCLUDED.failure_kind_counts,
    quality_basis = EXCLUDED.quality_basis,
    latest_scheduled_status = EXCLUDED.latest_scheduled_status,
    latest_scheduled_at = EXCLUDED.latest_scheduled_at,
    latest_observation_status = EXCLUDED.latest_observation_status,
    latest_observation_at = EXCLUDED.latest_observation_at,
    known_state_observed_at = EXCLUDED.known_state_observed_at,
    known_state = EXCLUDED.known_state,
    avg_duration_ms = EXCLUDED.avg_duration_ms,
    p95_duration_ms = EXCLUDED.p95_duration_ms,
    projection_version = EXCLUDED.projection_version,
    finalized_at = transaction_timestamp(),
    updated_at = transaction_timestamp();

-- name: ProjectNavTargetDaily :one
SELECT gfn_project_target_fact_day(sqlc.arg(fact_date)::date)::bigint;

-- name: ProjectNavSiteDaily :one
SELECT gfn_project_site_fact_day(sqlc.arg(fact_date)::date)::bigint;

-- name: NavTargetRebuildSourceAvailable :one
SELECT NOT EXISTS (
    SELECT 1
    FROM gfn_site_target_protocol_daily fact
    WHERE fact.fact_date = sqlc.arg(fact_date)
      AND (
          (fact.latest_observation_at IS NOT NULL AND NOT EXISTS (
              SELECT 1 FROM gfn_collector_observation observation
              WHERE observation.site_id = fact.site_id
                AND lower(observation.target) = lower(fact.target)
                AND observation.protocol = fact.protocol
                AND observation.observed_at = fact.latest_observation_at
          ))
          OR
          (fact.known_state_observed_at IS NOT NULL AND NOT EXISTS (
              SELECT 1 FROM gfn_collector_observation observation
              WHERE observation.site_id = fact.site_id
                AND lower(observation.target) = lower(fact.target)
                AND observation.protocol = fact.protocol
                AND observation.status = 'success'
                AND observation.observed_at = fact.known_state_observed_at
          ))
          OR
          (fact.attempted_count > 0 AND (
              SELECT count(*)
              FROM gfn_collection_task_results result
              JOIN gfn_collection_runs run ON run.id = result.run_id
              JOIN gfn_collection_jobs job ON job.id = run.job_id
              WHERE result.site_id = fact.site_id
                AND lower(result.target) = lower(fact.target)
                AND result.protocol = fact.protocol
                AND job.trigger = 'scheduled'
                AND (job.scheduled_for AT TIME ZONE 'UTC')::date = fact.fact_date
          ) < fact.attempted_count)
      )
) AS available;

-- name: AdvanceNavFactCheckpoint :execrows
UPDATE gfn_fact_rollup_checkpoints
SET processed_through = sqlc.arg(processed_through),
    updated_at = transaction_timestamp()
WHERE pipeline_key = sqlc.arg(pipeline_key)
  AND (processed_through IS NULL OR processed_through < sqlc.arg(processed_through));

-- name: PruneNavObservationsBatch :one
SELECT gfn_prune_observations_batch(
    sqlc.arg(keep_count)::integer,
    sqlc.arg(batch_size)::integer
)::bigint;
