-- name: LockCollectorDomainForUpdate :one
SELECT id, name, proxy, prefix, tls, site_id, deleted
FROM gfn_collector_domain
WHERE id = sqlc.arg(id)
FOR UPDATE;

-- name: LockSiteForUpdate :one
SELECT id, name, name_en, info, info_en, create_time, update_time, country,
       nsfw, welfare, icon, deleted, view_count, deleted_at
FROM gfn_site
WHERE id = sqlc.arg(id)
FOR UPDATE;

-- name: CountConflictingActiveTarget :one
SELECT count(*)::bigint
FROM gfn_target_tracking_periods
WHERE site_id = sqlc.arg(site_id)
  AND lower(target) = lower(sqlc.arg(target)::text)
  AND tracked_until IS NULL
  AND (sqlc.narg(exclude_domain_id)::bigint IS NULL
       OR collector_domain_id IS NULL
       OR collector_domain_id <> sqlc.narg(exclude_domain_id)::bigint);

-- name: GetActiveTargetPeriodByDomain :one
SELECT *
FROM gfn_target_tracking_periods
WHERE collector_domain_id = sqlc.arg(collector_domain_id)
  AND tracked_until IS NULL
FOR UPDATE;

-- name: OpenTargetTrackingPeriod :one
INSERT INTO gfn_target_tracking_periods (
    collector_domain_id, site_id, target, tracked_from,
    tracking_basis, opened_reason
)
VALUES (
    sqlc.arg(collector_domain_id), sqlc.arg(site_id), sqlc.arg(target),
    transaction_timestamp(), 'explicit', sqlc.arg(opened_reason)
)
RETURNING *;

-- name: CloseTargetTrackingPeriod :one
UPDATE gfn_target_tracking_periods
SET tracked_until = GREATEST(transaction_timestamp(), tracked_from + interval '1 microsecond'),
    closed_reason = sqlc.arg(closed_reason),
    updated_at = transaction_timestamp()
WHERE id = sqlc.arg(id) AND tracked_until IS NULL
RETURNING *;

-- name: CloseTargetTrackingPeriodsBySite :execrows
UPDATE gfn_target_tracking_periods
SET tracked_until = GREATEST(transaction_timestamp(), tracked_from + interval '1 microsecond'),
    closed_reason = sqlc.arg(closed_reason),
    updated_at = transaction_timestamp()
WHERE site_id = sqlc.arg(site_id) AND tracked_until IS NULL;

-- name: GetActivePrimaryTarget :one
SELECT primary_period.id,
       primary_period.site_id,
       primary_period.target_tracking_period_id,
       primary_period.basis,
       target_period.collector_domain_id,
       target_period.target
FROM gfn_site_primary_target_periods primary_period
JOIN gfn_target_tracking_periods target_period
  ON target_period.id = primary_period.target_tracking_period_id
WHERE primary_period.site_id = sqlc.arg(site_id)
  AND primary_period.effective_until IS NULL
FOR UPDATE OF primary_period;

-- name: ClosePrimaryTargetBySite :execrows
UPDATE gfn_site_primary_target_periods
SET effective_until = GREATEST(transaction_timestamp(), effective_from + interval '1 microsecond'),
    updated_at = transaction_timestamp()
WHERE site_id = sqlc.arg(site_id) AND effective_until IS NULL;

-- name: OpenPrimaryTarget :one
INSERT INTO gfn_site_primary_target_periods (
    site_id, target_tracking_period_id, effective_from, basis
)
VALUES (
    sqlc.arg(site_id), sqlc.arg(target_tracking_period_id),
    transaction_timestamp(), sqlc.arg(basis)
)
RETURNING *;

-- name: AssignPrimaryTargetIfMissing :execrows
WITH candidates AS (
    SELECT target_period.id,
           target_period.site_id,
           count(*) OVER () AS target_count,
           row_number() OVER (
               ORDER BY CASE WHEN COALESCE(domain.prefix, '') = '' THEN 0 ELSE 1 END,
                        domain.id
           ) AS target_rank
    FROM gfn_target_tracking_periods target_period
    JOIN gfn_collector_domain domain
      ON domain.id = target_period.collector_domain_id
    WHERE target_period.site_id = sqlc.arg(site_id)
      AND target_period.tracked_until IS NULL
      AND NOT domain.deleted
)
INSERT INTO gfn_site_primary_target_periods (
    site_id, target_tracking_period_id, effective_from, basis
)
SELECT site_id,
       id,
       transaction_timestamp(),
       CASE WHEN target_count = 1
            THEN 'single_target_inferred'
            ELSE 'deterministic_fallback'
       END
FROM candidates
WHERE target_rank = 1
  AND NOT EXISTS (
      SELECT 1 FROM gfn_site_primary_target_periods current_primary
      WHERE current_primary.site_id = sqlc.arg(site_id)
        AND current_primary.effective_until IS NULL
  )
ON CONFLICT (site_id) WHERE effective_until IS NULL DO NOTHING;

-- name: RefreshCurrentSiteDaily :exec
SELECT gfn_refresh_current_site_daily(sqlc.arg(site_id)::bigint);
