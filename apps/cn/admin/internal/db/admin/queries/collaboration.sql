-- name: CountContentIdeas :one
SELECT count(*) FROM gfa_content_idea i
WHERE (sqlc.arg(kind)::text = '' OR i.kind = sqlc.arg(kind))
AND (sqlc.arg(status)::text = 'all' OR i.status = sqlc.arg(status) OR (sqlc.arg(status) = 'active' AND i.status IN ('idea', 'researching')))
AND (sqlc.arg(priority)::text = '' OR i.priority = sqlc.arg(priority))
AND (sqlc.arg(researcher)::text = 'all' OR (sqlc.arg(researcher) = 'me' AND i.researching_by_account_id = sqlc.arg(account_id)::bigint) OR (sqlc.arg(researcher) = 'unassigned' AND i.researching_by_account_id IS NULL))
AND (sqlc.arg(keyword)::text = '' OR concat_ws(' ', i.title, i.source, i.note) ILIKE '%' || sqlc.arg(keyword) || '%');

-- name: ListContentIdeas :many
SELECT i.*, EXISTS (SELECT 1 FROM gfa_content_idea duplicate WHERE duplicate.source_key = i.source_key AND duplicate.id <> i.id) AS idea_duplicate, creator.display_name AS creator_name, COALESCE(researcher.display_name, '')::text AS researcher_name
FROM gfa_content_idea i
JOIN gfa_admin_account creator ON creator.id = i.created_by_account_id
LEFT JOIN gfa_admin_account researcher ON researcher.id = i.researching_by_account_id
WHERE (sqlc.arg(kind)::text = '' OR i.kind = sqlc.arg(kind))
AND (sqlc.arg(status)::text = 'all' OR i.status = sqlc.arg(status) OR (sqlc.arg(status) = 'active' AND i.status IN ('idea', 'researching')))
AND (sqlc.arg(priority)::text = '' OR i.priority = sqlc.arg(priority))
AND (sqlc.arg(researcher)::text = 'all' OR (sqlc.arg(researcher) = 'me' AND i.researching_by_account_id = sqlc.arg(account_id)::bigint) OR (sqlc.arg(researcher) = 'unassigned' AND i.researching_by_account_id IS NULL))
AND (sqlc.arg(keyword)::text = '' OR concat_ws(' ', i.title, i.source, i.note) ILIKE '%' || sqlc.arg(keyword) || '%')
ORDER BY i.updated_at DESC, i.id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: GetContentIdea :one
SELECT i.*, creator.display_name AS creator_name, COALESCE(researcher.display_name, '')::text AS researcher_name
FROM gfa_content_idea i JOIN gfa_admin_account creator ON creator.id = i.created_by_account_id
LEFT JOIN gfa_admin_account researcher ON researcher.id = i.researching_by_account_id WHERE i.id = $1;

-- name: ListContentIdeasBySourceKeys :many
SELECT DISTINCT ON (source_key) id, kind, title, source_key FROM gfa_content_idea
WHERE source_key = ANY(sqlc.arg(source_keys)::text[]) ORDER BY source_key, id;

-- name: InsertContentIdea :one
INSERT INTO gfa_content_idea (kind, title, source, source_key, note, priority, created_by_account_id)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: InsertContentIdeasBatch :copyfrom
INSERT INTO gfa_content_idea (kind, title, source, source_key, note, priority, created_by_account_id)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: UpdateContentIdeaVersioned :one
UPDATE gfa_content_idea SET kind = sqlc.arg(kind), title = sqlc.narg(title), source = sqlc.narg(source),
source_key = sqlc.narg(source_key), note = sqlc.arg(note), priority = sqlc.arg(priority), version = version + 1, updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version) RETURNING *;

-- name: ResearchContentIdeaVersioned :one
UPDATE gfa_content_idea SET status = 'researching', researching_by_account_id = sqlc.arg(account_id), researching_at = CURRENT_TIMESTAMP,
version = version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = sqlc.arg(id) AND version = sqlc.arg(version) AND status = 'idea' RETURNING *;

-- name: ReleaseContentIdeaVersioned :one
UPDATE gfa_content_idea SET status = 'idea', researching_by_account_id = NULL, researching_at = NULL,
version = version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = sqlc.arg(id) AND version = sqlc.arg(version) AND status = 'researching' RETURNING *;

-- name: ShelveContentIdeaVersioned :one
UPDATE gfa_content_idea SET status = 'shelved', researching_by_account_id = NULL, researching_at = NULL,
version = version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = sqlc.arg(id) AND version = sqlc.arg(version) AND status IN ('idea', 'researching') RETURNING *;

-- name: RestoreContentIdeaVersioned :one
UPDATE gfa_content_idea SET status = 'idea', version = version + 1, updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version) AND status = 'shelved' RETURNING *;

-- name: LinkContentIdeaVersioned :one
UPDATE gfa_content_idea SET status = 'landed', linked_kind = sqlc.arg(linked_kind), linked_resource_id = sqlc.arg(linked_resource_id), landed_at = CURRENT_TIMESTAMP,
researching_by_account_id = NULL, researching_at = NULL, version = version + 1, updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version) AND status IN ('idea', 'researching') RETURNING *;

-- name: LandOtherContentIdeaVersioned :one
UPDATE gfa_content_idea SET status = 'landed', landed_at = CURRENT_TIMESTAMP, researching_by_account_id = NULL, researching_at = NULL,
version = version + 1, updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version) AND kind = 'other' AND status IN ('idea', 'researching') RETURNING *;

-- name: ReopenContentIdeaVersioned :one
UPDATE gfa_content_idea SET status = 'idea', linked_kind = NULL, linked_resource_id = NULL, landed_at = NULL,
version = version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = sqlc.arg(id) AND version = sqlc.arg(version) AND status = 'landed' RETURNING *;

-- name: CountContentIdeaSummary :one
SELECT count(*) FILTER (WHERE status = 'idea') AS reserve_count,
count(*) FILTER (WHERE status = 'researching') AS researching_count,
count(*) FILTER (WHERE status = 'landed' AND landed_at >= CURRENT_TIMESTAMP - interval '30 days') AS landed_30d FROM gfa_content_idea;

-- name: DeleteContentIdeaVersioned :one
DELETE FROM gfa_content_idea WHERE id = $1 AND version = $2 RETURNING *;
