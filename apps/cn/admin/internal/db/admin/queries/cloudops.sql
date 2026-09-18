-- Dedicated GFA session lock for the daily EdgeOne main-host purge. Keep the
-- owning connection through intent audit, remote submission and result audit.
-- name: TryLockEdgeOneScheduledPurge :one
SELECT pg_try_advisory_lock(718210552122::bigint)::boolean AS acquired;

-- name: UnlockEdgeOneScheduledPurge :one
SELECT pg_advisory_unlock(718210552122::bigint)::boolean AS released;

-- The intent is committed before submission. Check under the session lock to
-- prevent a late instance from resubmitting after the first releases the lock.
-- Terminal rows also imply an intent, so no parsing of audit text is required.
-- name: HasEdgeOneScheduledPurgeRequest :one
SELECT EXISTS (
    SELECT 1 FROM gfa_admin_audit_log
    WHERE action = 'cloud.edgeone.purge.host.scheduled'
      AND resource = 'cloud'
      AND request_id = sqlc.arg(request_id)::text
      AND target_id = sqlc.arg(target_id)::text
)::boolean AS recorded;
