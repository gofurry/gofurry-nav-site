-- name: LockSiteGroupCuration :exec
SELECT pg_advisory_xact_lock(hashtext('admin.site_group_curation')::bigint);

-- name: ListGroupCuration :many
SELECT m.id, m.site_id, m.weight, s.name, s.deleted, s.update_time
FROM gfn_site_group_map m JOIN gfn_site s ON s.id=m.site_id
WHERE m.group_id=sqlc.arg(group_id)
ORDER BY s.deleted ASC, m.weight DESC, s.update_time DESC, s.id DESC;

-- name: UpdateGroupCurationWeight :execrows
UPDATE gfn_site_group_map SET weight=sqlc.arg(weight), update_time=NOW()::timestamp(0)
WHERE group_id=sqlc.arg(group_id) AND site_id=sqlc.arg(site_id);
