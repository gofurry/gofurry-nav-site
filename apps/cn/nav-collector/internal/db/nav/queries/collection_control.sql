-- name: EnsureNavCollectionSchedule :exec
INSERT INTO gfn_collection_schedules (
    job_key, name, enabled, schedule_kind, cron_expression, interval_seconds,
    anchor_at, timezone, misfire_policy, misfire_grace_seconds, overlap_policy,
    priority, concurrency_key, effective_from, created_at, updated_at
) VALUES (
    sqlc.arg(job_key), sqlc.arg(name), sqlc.arg(enabled), sqlc.arg(schedule_kind),
    sqlc.narg(cron_expression), sqlc.narg(interval_seconds), sqlc.narg(anchor_at),
    sqlc.arg(timezone), sqlc.arg(misfire_policy), sqlc.arg(misfire_grace_seconds),
    'skip', sqlc.arg(priority), sqlc.arg(concurrency_key), now(), now(), now()
)
ON CONFLICT (job_key) DO NOTHING;

-- name: ListNavCollectionSchedules :many
SELECT * FROM gfn_collection_schedules ORDER BY job_key;

-- name: NavCollectionClock :one
SELECT now()::timestamptz AS control_now;

-- name: GetNavCollectionScheduleForUpdate :one
SELECT * FROM gfn_collection_schedules WHERE id = sqlc.arg(id) FOR UPDATE;

-- name: UpdateNavCollectionScheduleCursor :exec
UPDATE gfn_collection_schedules
SET last_materialized_for = sqlc.narg(last_materialized_for),
    next_scheduled_for = sqlc.narg(next_scheduled_for), updated_at = now()
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version);

-- name: InsertNavScheduledJob :execrows
INSERT INTO gfn_collection_jobs (
    schedule_id, schedule_version, job_key, trigger, scope_type, tasks,
    priority, concurrency_key, scheduled_for, status, requested_by,
    created_at, updated_at, completed_at
) VALUES (
    sqlc.arg(schedule_id), sqlc.arg(schedule_version), sqlc.arg(job_key),
    sqlc.arg(trigger), 'all', sqlc.arg(tasks), sqlc.arg(priority),
    sqlc.arg(concurrency_key), sqlc.arg(scheduled_for), sqlc.arg(status),
    sqlc.arg(requested_by), now(), now(),
    CASE WHEN sqlc.arg(status)::text IN ('missed', 'skipped') THEN now() ELSE NULL END
)
ON CONFLICT (schedule_id, scheduled_for)
    WHERE schedule_id IS NOT NULL AND scheduled_for IS NOT NULL
DO NOTHING;

-- name: NavCollectionLaneActive :one
SELECT EXISTS (
    SELECT 1 FROM gfn_collection_jobs
    WHERE concurrency_key = sqlc.arg(concurrency_key) AND status = 'running'
)::boolean;

-- name: ClaimNextNavCollectionJob :one
WITH candidate AS (
    SELECT queued.id
    FROM gfn_collection_jobs queued
    WHERE queued.status = 'queued'
      AND (queued.scheduled_for IS NULL OR queued.scheduled_for <= now())
      AND NOT EXISTS (
          SELECT 1 FROM gfn_collection_jobs active
          WHERE active.status = 'running'
            AND active.concurrency_key = queued.concurrency_key
      )
    ORDER BY queued.priority DESC, queued.created_at ASC, queued.id ASC
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
UPDATE gfn_collection_jobs job
SET status = 'running', claimed_by = sqlc.arg(instance_id),
    lease_until = now() + sqlc.arg(lease_seconds)::bigint * interval '1 second',
    updated_at = now()
FROM candidate
WHERE job.id = candidate.id
RETURNING job.*;

-- name: SkipOverlappingNavPointInTimeJobs :execrows
UPDATE gfn_collection_jobs queued
SET status = 'skipped', completed_at = now(), updated_at = now()
WHERE queued.status = 'queued'
  AND queued.trigger = 'scheduled'
  AND queued.job_key IN ('nav.ping', 'nav.http', 'nav.dns', 'nav.port_check')
  AND EXISTS (
      SELECT 1
      FROM gfn_collection_jobs active
      WHERE active.status = 'running'
        AND active.concurrency_key = queued.concurrency_key
  );

-- name: ListNavCollectionTargets :many
SELECT DISTINCT cd.site_id::bigint AS site_id,
       (COALESCE(cd.prefix, '') || cd.name)::text AS target
FROM gfn_collector_domain cd
JOIN gfn_site site ON site.id = cd.site_id
WHERE cd.deleted IS NOT TRUE
  AND site.deleted IS NOT TRUE
  AND cd.site_id > 0
  AND (
      sqlc.arg(scope_type)::text = 'all'
      OR cd.site_id = sqlc.narg(scope_id)
  )
  AND (
      sqlc.arg(scope_type)::text <> 'target'
      OR lower(COALESCE(cd.prefix, '') || cd.name) = lower(sqlc.narg(target)::text)
  )
ORDER BY site_id, target;

-- name: NextNavCollectionAttempt :one
SELECT COALESCE(max(attempt_no), 0)::integer + 1
FROM gfn_collection_runs WHERE job_id = sqlc.arg(job_id);

-- name: InsertNavCollectionRun :one
INSERT INTO gfn_collection_runs (
    id, job_id, attempt_no, collector_instance_id, status, scheduled_for,
    started_at, expected_count, attempted_count, success_count, partial_count,
    failure_count, skipped_count, schedule_delay_ms, duration_ms
) VALUES (
    sqlc.arg(id), sqlc.arg(job_id), sqlc.arg(attempt_no),
    sqlc.arg(collector_instance_id), 'running', sqlc.narg(scheduled_for), now(),
    sqlc.arg(expected_count), 0, 0, 0, 0, 0, sqlc.arg(schedule_delay_ms), 0
)
RETURNING *;

-- name: RenewNavCollectionLease :execrows
UPDATE gfn_collection_jobs
SET lease_until = now() + sqlc.arg(lease_seconds)::bigint * interval '1 second', updated_at = now()
WHERE id = sqlc.arg(job_id) AND claimed_by = sqlc.arg(instance_id) AND status = 'running';

-- name: NavCollectionCancelRequested :one
SELECT cancel_requested_at
FROM gfn_collection_jobs
WHERE id = sqlc.arg(job_id) AND claimed_by = sqlc.arg(instance_id) AND status = 'running';

-- name: ListExpiredNavCollectionJobs :many
SELECT j.id AS job_id, r.id AS run_id
FROM gfn_collection_jobs j
JOIN gfn_collection_runs r ON r.job_id = j.id AND r.status = 'running'
WHERE j.status = 'running' AND j.lease_until < now()
ORDER BY j.lease_until
FOR UPDATE OF j SKIP LOCKED;

-- name: FailLostNavCollectionRun :execrows
UPDATE gfn_collection_runs
SET status = 'failed', ended_at = now(),
    duration_ms = GREATEST(0, (extract(epoch FROM (now() - started_at)) * 1000)::bigint),
    error_kind = 'worker_lost', error_message = 'Collector lease expired.'
WHERE id = sqlc.arg(run_id) AND status = 'running';

-- name: FailLostNavCollectionJob :execrows
UPDATE gfn_collection_jobs
SET status = 'failed', completed_at = now(), updated_at = now(), lease_until = NULL
WHERE id = sqlc.arg(job_id) AND status = 'running';

-- name: FinalizeNavCollectionRun :execrows
UPDATE gfn_collection_runs
SET status = sqlc.arg(status), ended_at = now(),
    attempted_count = sqlc.arg(attempted_count), success_count = sqlc.arg(success_count),
    partial_count = sqlc.arg(partial_count), failure_count = sqlc.arg(failure_count),
    skipped_count = sqlc.arg(skipped_count),
    duration_ms = GREATEST(0, (extract(epoch FROM (now() - started_at)) * 1000)::bigint),
    error_kind = sqlc.arg(error_kind), error_message = sqlc.arg(error_message)
WHERE id = sqlc.arg(id) AND status = 'running';

-- name: FinalizeNavCollectionJob :execrows
UPDATE gfn_collection_jobs
SET status = sqlc.arg(status), completed_at = now(), updated_at = now(), lease_until = NULL
WHERE id = sqlc.arg(id) AND claimed_by = sqlc.arg(instance_id) AND status = 'running';

-- name: InsertNavCollectionTaskResult :exec
INSERT INTO gfn_collection_task_results (
    run_id, protocol, site_id, target, status, observation_id, duration_ms,
    error_kind, error_message, started_at, ended_at
) VALUES (
    sqlc.arg(run_id), sqlc.arg(protocol), sqlc.arg(site_id), sqlc.arg(target),
    sqlc.arg(status), sqlc.narg(observation_id), sqlc.arg(duration_ms),
    sqlc.arg(error_kind), sqlc.arg(error_message), sqlc.arg(started_at), sqlc.narg(ended_at)
)
ON CONFLICT (run_id, protocol, site_id, target) DO UPDATE SET
    status = EXCLUDED.status, observation_id = EXCLUDED.observation_id,
    duration_ms = EXCLUDED.duration_ms, error_kind = EXCLUDED.error_kind,
    error_message = EXCLUDED.error_message, started_at = EXCLUDED.started_at,
    ended_at = EXCLUDED.ended_at;

-- name: RegisterNavCollectorInstance :exec
INSERT INTO gfn_collector_instances (
    instance_id, collector_id, hostname, version, commit_sha, capabilities,
    started_at, last_heartbeat_at
) VALUES (
    sqlc.arg(instance_id), sqlc.arg(collector_id), sqlc.arg(hostname),
    sqlc.arg(version), sqlc.arg(commit_sha), sqlc.arg(capabilities), now(), now()
);

-- name: HeartbeatNavCollectorInstance :execrows
UPDATE gfn_collector_instances SET last_heartbeat_at = now()
WHERE instance_id = sqlc.arg(instance_id) AND stopped_at IS NULL;

-- name: StopNavCollectorInstance :execrows
UPDATE gfn_collector_instances SET last_heartbeat_at = now(), stopped_at = now()
WHERE instance_id = sqlc.arg(instance_id) AND stopped_at IS NULL;
