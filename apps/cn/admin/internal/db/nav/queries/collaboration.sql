-- Collector targets store prefix and hostname separately. Preserve subdomains;
-- strip only port, trailing dot and one leading www, as the Go normalizer does.
-- name: ListSitesByNormalizedHostsForCollaboration :many
SELECT DISTINCT site.id, site.name, site.name_en,
    regexp_replace(rtrim(regexp_replace(lower(btrim(COALESCE(target.prefix, '') || target.name)), ':[0-9]+$', ''), '.'), '^www\.', '')::text AS host
FROM gfn_site site
JOIN gfn_collector_domain target ON target.site_id = site.id AND target.deleted IS NOT TRUE
WHERE site.deleted IS NOT TRUE
AND regexp_replace(rtrim(regexp_replace(lower(btrim(COALESCE(target.prefix, '') || target.name)), ':[0-9]+$', ''), '.'), '^www\.', '') = ANY(sqlc.arg(hosts)::text[])
ORDER BY site.id;

-- name: GetSiteForCollaborationLink :one
SELECT id, name, name_en FROM gfn_site WHERE id = $1 AND deleted IS NOT TRUE;

-- name: ListBoardSiteReferences :many
SELECT id,name FROM gfn_site WHERE id=ANY(sqlc.arg(ids)::bigint[]) AND deleted IS NOT TRUE;
