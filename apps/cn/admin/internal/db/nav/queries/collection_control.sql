-- name: AdminNavCollectionCounts :one
SELECT
    count(*) FILTER (WHERE status = 'running')::bigint AS running_count,
    count(*) FILTER (WHERE status = 'queued')::bigint AS queued_count,
    count(*) FILTER (WHERE status = 'failed' AND created_at >= now() - interval '24 hours')::bigint AS failed_24h,
    count(*) FILTER (WHERE status = 'missed' AND created_at >= now() - interval '24 hours')::bigint AS missed_24h
FROM gfn_collection_jobs;

-- name: AdminNavCollectionClock :one
SELECT now()::timestamptz AS control_now;

-- name: AdminListNavCollectorInstances :many
SELECT instance.*,
       GREATEST(0, extract(epoch FROM (now() - last_heartbeat_at)))::bigint AS heartbeat_age_seconds
FROM gfn_collector_instances instance
ORDER BY last_heartbeat_at DESC
LIMIT sqlc.arg(row_limit);

-- name: AdminListNavCollectionSchedules :many
SELECT schedule.*,
       COALESCE(latest.status, '') AS last_status,
       latest.success_count AS last_success_count,
       latest.expected_count AS last_expected_count
FROM gfn_collection_schedules schedule
LEFT JOIN LATERAL (
    SELECT job.status, run.success_count, run.expected_count
    FROM gfn_collection_jobs job
    LEFT JOIN gfn_collection_runs run ON run.job_id = job.id
    WHERE job.schedule_id = schedule.id
    ORDER BY job.created_at DESC, run.attempt_no DESC NULLS LAST
    LIMIT 1
) latest ON true
ORDER BY schedule.job_key;

-- name: AdminGetNavCollectionSchedule :one
SELECT * FROM gfn_collection_schedules WHERE id = sqlc.arg(id);

-- name: AdminUpdateNavCollectionSchedule :one
UPDATE gfn_collection_schedules
SET enabled = sqlc.arg(enabled), schedule_kind = sqlc.arg(schedule_kind),
    cron_expression = sqlc.narg(cron_expression), interval_seconds = sqlc.narg(interval_seconds),
    anchor_at = sqlc.narg(anchor_at), timezone = sqlc.arg(timezone),
    misfire_policy = sqlc.arg(misfire_policy),
    misfire_grace_seconds = sqlc.arg(misfire_grace_seconds),
    version = version + 1, effective_from = now(), last_materialized_for = now(),
    next_scheduled_for = NULL, updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: AdminInsertNavManualJob :one
INSERT INTO gfn_collection_jobs (
    job_key, trigger, scope_type, scope_id, target, tasks, priority,
    concurrency_key, status, requested_by, dedupe_key, created_at, updated_at
) VALUES (
    sqlc.arg(job_key), 'manual', sqlc.arg(scope_type), sqlc.narg(scope_id),
    sqlc.narg(target), ARRAY[sqlc.arg(protocol)::text], 200,
    sqlc.arg(protocol), 'queued', sqlc.arg(requested_by), sqlc.narg(dedupe_key),
    now(), now()
)
ON CONFLICT (dedupe_key)
    WHERE dedupe_key IS NOT NULL AND status IN ('queued', 'running')
DO UPDATE SET dedupe_key = EXCLUDED.dedupe_key
RETURNING *;

-- name: AdminListNavCollectionJobs :many
SELECT job.*, COALESCE(latest.id, '') AS run_id
FROM gfn_collection_jobs job
LEFT JOIN LATERAL (
    SELECT id FROM gfn_collection_runs run WHERE run.job_id = job.id ORDER BY attempt_no DESC LIMIT 1
) latest ON true
WHERE (sqlc.arg(status)::text = '' OR job.status = sqlc.arg(status))
  AND (sqlc.arg(job_key)::text = '' OR job.job_key = sqlc.arg(job_key))
  AND (sqlc.arg(trigger)::text = '' OR job.trigger = sqlc.arg(trigger))
ORDER BY CASE WHEN job.status = 'running' THEN 0 WHEN job.status = 'queued' THEN 1 ELSE 2 END,
         job.priority DESC, job.created_at DESC
LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: AdminGetNavCollectionJob :one
SELECT * FROM gfn_collection_jobs WHERE id = sqlc.arg(id);

-- name: AdminCancelNavCollectionJob :one
UPDATE gfn_collection_jobs
SET status = CASE WHEN status = 'queued' THEN 'canceled' ELSE status END,
    cancel_requested_at = CASE WHEN status = 'running' THEN now() ELSE cancel_requested_at END,
    completed_at = CASE WHEN status = 'queued' THEN now() ELSE completed_at END,
    updated_at = now()
WHERE id = sqlc.arg(id) AND status IN ('queued', 'running')
RETURNING *;

-- name: AdminRetryNavCollectionJob :one
UPDATE gfn_collection_jobs
SET status = 'queued', claimed_by = NULL, lease_until = NULL,
    cancel_requested_at = NULL, completed_at = NULL,
    requested_by = sqlc.arg(requested_by), updated_at = now()
WHERE id = sqlc.arg(id) AND status IN ('success', 'partial', 'failed', 'canceled')
RETURNING *;

-- name: AdminListNavCollectionRuns :many
SELECT run.*, job.job_key, job.trigger, job.scope_type, job.scope_id, job.target
FROM gfn_collection_runs run
JOIN gfn_collection_jobs job ON job.id = run.job_id
WHERE (sqlc.arg(status)::text = '' OR run.status = sqlc.arg(status))
  AND (sqlc.arg(job_key)::text = '' OR job.job_key = sqlc.arg(job_key))
  AND (sqlc.arg(trigger)::text = '' OR job.trigger = sqlc.arg(trigger))
  AND (sqlc.narg(since)::timestamptz IS NULL OR run.started_at >= sqlc.narg(since))
  AND (sqlc.narg(until)::timestamptz IS NULL OR run.started_at <= sqlc.narg(until))
ORDER BY run.started_at DESC
LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: AdminGetNavCollectionRun :one
SELECT run.*, job.job_key, job.trigger, job.scope_type, job.scope_id, job.target
FROM gfn_collection_runs run JOIN gfn_collection_jobs job ON job.id = run.job_id
WHERE run.id = sqlc.arg(id);

-- name: AdminListNavCollectionResults :many
SELECT * FROM gfn_collection_task_results
WHERE run_id = sqlc.arg(run_id)
  AND (sqlc.narg(site_id)::bigint IS NULL OR site_id = sqlc.narg(site_id))
  AND (sqlc.narg(target)::text IS NULL OR target = sqlc.narg(target))
  AND (sqlc.arg(protocol)::text = '' OR protocol = sqlc.arg(protocol))
ORDER BY started_at DESC, id DESC
LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: AdminListNavCollectionChartRows :many
SELECT job.id AS job_id, job.job_key, job.status AS job_status, job.scheduled_for,
       job.created_at, COALESCE(run.id, '') AS run_id, COALESCE(run.status, '') AS run_status,
       COALESCE(run.expected_count, 0)::integer AS expected_count,
       COALESCE(run.attempted_count, 0)::integer AS attempted_count,
       COALESCE(run.success_count, 0)::integer AS success_count,
       COALESCE(run.partial_count, 0)::integer AS partial_count,
       COALESCE(run.failure_count, 0)::integer AS failure_count,
       COALESCE(run.skipped_count, 0)::integer AS skipped_count,
       COALESCE(run.schedule_delay_ms, 0)::bigint AS schedule_delay_ms,
       COALESCE(run.duration_ms, 0)::bigint AS duration_ms, run.started_at
FROM gfn_collection_jobs job
LEFT JOIN LATERAL (
    SELECT * FROM gfn_collection_runs candidate
    WHERE candidate.job_id = job.id ORDER BY candidate.attempt_no DESC LIMIT 1
) run ON true
WHERE job.created_at >= sqlc.arg(since)
  AND (sqlc.arg(job_key)::text = '' OR job.job_key = sqlc.arg(job_key))
ORDER BY job.created_at;

-- name: AdminNavCollectionSiteExists :one
SELECT EXISTS (SELECT 1 FROM gfn_site WHERE id = sqlc.arg(id) AND deleted IS NOT TRUE)::boolean;

-- name: AdminNavCollectionTargetExists :one
SELECT EXISTS (
    SELECT 1 FROM gfn_collector_domain
    WHERE site_id = sqlc.arg(site_id) AND deleted IS NOT TRUE
      AND lower(COALESCE(prefix, '') || name) = lower(sqlc.arg(target))
)::boolean;
