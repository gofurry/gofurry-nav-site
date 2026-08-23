-- name: FoundationPing :one
SELECT 1::bigint AS value;

-- name: ListPublicSites :many
SELECT
    s.id,
    s.name,
    s.name_en,
    json_build_object(
        'domain',
        COALESCE(cd.domains, ARRAY[]::text[])
    )::text AS domain,
    s.info,
    s.info_en,
    s.create_time,
    s.update_time,
    s.country,
    s.nsfw,
    s.welfare,
    s.view_count,
    s.icon,
    s.deleted
FROM gfn_site AS s
LEFT JOIN LATERAL (
    SELECT array_agg(
        trim(COALESCE(prefix, '') || name)
        ORDER BY id ASC
    ) AS domains
    FROM gfn_collector_domain
    WHERE site_id = s.id
      AND site_id > 0
      AND deleted IS NOT TRUE
      AND trim(COALESCE(prefix, '') || name) <> ''
) AS cd ON TRUE
WHERE s.deleted IS NOT TRUE
ORDER BY s.update_time DESC, s.id DESC;

-- name: ListPublicSiteIndex :many
SELECT
    s.id,
    json_build_object(
        'domain',
        COALESCE(cd.domains, ARRAY[]::text[])
    )::text AS domain,
    s.update_time
FROM gfn_site AS s
LEFT JOIN LATERAL (
    SELECT array_agg(
        trim(COALESCE(prefix, '') || name)
        ORDER BY id ASC
    ) AS domains
    FROM gfn_collector_domain
    WHERE site_id = s.id
      AND site_id > 0
      AND deleted IS NOT TRUE
      AND trim(COALESCE(prefix, '') || name) <> ''
) AS cd ON TRUE
WHERE s.deleted IS NOT TRUE
ORDER BY s.id ASC;

-- name: ListSiteGroups :many
SELECT id, name, name_en, info, info_en, priority, create_time, update_time
FROM gfn_site_group
ORDER BY priority ASC;

-- name: ListSiteGroupMappings :many
SELECT id, site_id, group_id, weight, create_time, update_time
FROM gfn_site_group_map
ORDER BY group_id ASC, weight DESC, update_time DESC, id DESC, site_id ASC;

-- name: ListFeaturedSites :many
SELECT f.id, f.site_id, f.weight, f.create_time, f.update_time
FROM gfn_featured_site AS f
INNER JOIN gfn_site AS s
    ON s.id = f.site_id
   AND s.deleted IS NOT TRUE
ORDER BY f.weight DESC, f.id DESC;

-- name: GetRandomSaying :one
SELECT id, author, saying, create_time, update_time, language
FROM gfn_saying
WHERE language = sqlc.arg(language)
   OR (sqlc.arg(language) <> 'zh' AND language = 'zh')
ORDER BY
    CASE WHEN language = sqlc.arg(language) THEN 0 ELSE 1 END,
    random()
LIMIT 1;

-- name: GetSiteByID :one
SELECT id, name, name_en, info, info_en, create_time, update_time,
       country, nsfw, welfare, icon, deleted, view_count
FROM gfn_site
WHERE id = sqlc.arg(site_id)
LIMIT 1;

-- name: GetPublicSiteByID :one
SELECT id, name, name_en, info, info_en, create_time, update_time,
       country, nsfw, welfare, icon, deleted, view_count
FROM gfn_site
WHERE id = sqlc.arg(site_id)
  AND deleted IS NOT TRUE
LIMIT 1;

-- name: ListPublicCollectorDomains :many
SELECT id, name, proxy, prefix, tls, site_id, deleted
FROM gfn_collector_domain
WHERE site_id = sqlc.arg(site_id)
  AND deleted IS NOT TRUE
ORDER BY id ASC;

-- name: ListObservationStatusSummary :many
SELECT protocol, status, COUNT(*)::bigint AS count
FROM gfn_collector_observation
WHERE observed_at >= NOW() - INTERVAL '7 days'
GROUP BY protocol, status
ORDER BY protocol ASC, status ASC;

-- name: ListCollectorObservations :many
SELECT
    id,
    site_id,
    target,
    protocol,
    status,
    observed_at,
    COALESCE(duration_ms, 0)::bigint AS duration_ms,
    error_code,
    error_message,
    COALESCE(payload->>'collector_id', '')::text AS collector_id,
    COALESCE(payload->>'job_id', '')::text AS job_id
FROM gfn_collector_observation
WHERE (sqlc.arg(site_id)::bigint <= 0 OR site_id = sqlc.arg(site_id))
  AND (sqlc.arg(target)::text = '' OR target = sqlc.arg(target))
  AND (sqlc.arg(protocol)::text = '' OR protocol = sqlc.arg(protocol))
  AND (sqlc.arg(status)::text = '' OR status = sqlc.arg(status))
ORDER BY observed_at DESC, id DESC
LIMIT sqlc.arg(row_limit)
OFFSET sqlc.arg(row_offset);

-- name: ListTargetObservations :many
SELECT id, site_id, target, protocol, status, observed_at,
       COALESCE(duration_ms, 0)::bigint AS duration_ms,
       error_code, error_message, payload, schema_version, create_time
FROM gfn_collector_observation
WHERE site_id = sqlc.arg(site_id)
  AND target = sqlc.arg(target)
  AND protocol = sqlc.arg(protocol)
ORDER BY observed_at DESC, id DESC
LIMIT sqlc.arg(row_limit);

-- name: ListPublicUpdateNotices :many
SELECT id, title, title_en, body, body_en, published_at,
       create_time, update_time, deleted
FROM gfn_nav_update_notice
WHERE deleted IS NOT TRUE
ORDER BY published_at DESC, id DESC
LIMIT sqlc.arg(row_limit);

-- name: UpdateSiteViewCount :exec
UPDATE gfn_site
SET view_count = sqlc.arg(view_count)
WHERE id = sqlc.arg(site_id);
