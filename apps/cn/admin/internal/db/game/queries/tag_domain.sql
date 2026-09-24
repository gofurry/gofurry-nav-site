-- All category, tag and classification mutations acquire this transaction lock first.
-- It serializes archive/assignment races across the small editorial domain and
-- establishes one lock order before any game or category row locks.
-- name: LockTagDomain :exec
SELECT pg_advisory_xact_lock(hashtextextended('gfg.tag-domain',0));

-- name: InvalidateGameRecommendations :exec
DELETE FROM gfg_game_recommendations;

-- name: ListGameTagRelations :many
SELECT * FROM gfg_game_tag WHERE game_id=sqlc.arg(game_id) ORDER BY tag_id;

-- name: DeleteGameTagRelations :exec
DELETE FROM gfg_game_tag WHERE game_id=sqlc.arg(game_id);

-- name: DeleteRemovedGameTags :exec
DELETE FROM gfg_game_tag WHERE game_id=sqlc.arg(game_id) AND NOT(tag_id=ANY(sqlc.arg(tag_ids)::bigint[]));

-- name: ClearChangedGameTagRoles :exec
UPDATE gfg_game_tag SET role='normal',update_time=NOW()::timestamp(0)
WHERE game_id=sqlc.arg(game_id) AND role<>'normal'
AND NOT ((role='primary' AND tag_id=sqlc.arg(primary_id)) OR (role='secondary' AND tag_id=sqlc.arg(secondary_id)));

-- name: UpsertGameTag :exec
INSERT INTO gfg_game_tag(game_id,tag_id,role,create_time,update_time)
VALUES(sqlc.arg(game_id),sqlc.arg(tag_id),sqlc.arg(role),NOW()::timestamp(0),NOW()::timestamp(0))
ON CONFLICT(game_id,tag_id) DO UPDATE SET role=EXCLUDED.role,update_time=EXCLUDED.update_time
WHERE gfg_game_tag.role<>EXCLUDED.role;

-- name: SetGameClassificationWeight :exec
UPDATE gfg_game SET weight=sqlc.arg(weight),update_time=NOW()::timestamp(0) WHERE id=sqlc.arg(id);

-- name: TagAssignmentCount :one
SELECT count(*) FROM gfg_game_tag WHERE tag_id=sqlc.arg(tag_id);

-- name: ActiveCategoryTagCount :one
SELECT count(*) FROM gfg_tag WHERE category_id=sqlc.arg(category_id) AND archived_at IS NULL;

-- name: CountTags :one
SELECT count(*) FROM gfg_tag WHERE sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR code ILIKE '%'||sqlc.arg(keyword)||'%';

-- name: ListTags :many
SELECT * FROM gfg_tag WHERE sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR code ILIKE '%'||sqlc.arg(keyword)||'%'
ORDER BY id LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: GetTag :one
SELECT * FROM gfg_tag WHERE id=sqlc.arg(id);

-- name: InsertTag :one
INSERT INTO gfg_tag(code,name,name_en,info,info_en,category_id,create_time,update_time)
VALUES(sqlc.arg(code),sqlc.arg(name),sqlc.arg(name_en),sqlc.arg(info),sqlc.arg(info_en),sqlc.arg(category_id),NOW()::timestamp(0),NOW()::timestamp(0)) RETURNING *;

-- name: UpdateTag :one
UPDATE gfg_tag SET name=sqlc.arg(name),name_en=sqlc.arg(name_en),info=sqlc.arg(info),info_en=sqlc.arg(info_en),category_id=sqlc.arg(category_id),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) RETURNING *;

-- name: ArchiveTag :one
UPDATE gfg_tag SET archived_at=CASE WHEN sqlc.arg(archive)::boolean THEN COALESCE(archived_at,NOW()::timestamp(0)) ELSE NULL END,update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) RETURNING *;

-- name: CountTagCategories :one
SELECT count(*) FROM gfg_tag_category WHERE sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR code ILIKE '%'||sqlc.arg(keyword)||'%';

-- name: ListTagCategories :many
SELECT * FROM gfg_tag_category WHERE sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR code ILIKE '%'||sqlc.arg(keyword)||'%'
ORDER BY sort_order,id LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: GetTagCategory :one
SELECT * FROM gfg_tag_category WHERE id=sqlc.arg(id);

-- name: InsertTagCategory :one
INSERT INTO gfg_tag_category(code,name,name_en,info,info_en,sort_order,create_time,update_time)
VALUES(sqlc.arg(code),sqlc.arg(name),sqlc.arg(name_en),sqlc.arg(info),sqlc.arg(info_en),sqlc.arg(sort_order),NOW()::timestamp(0),NOW()::timestamp(0)) RETURNING *;

-- name: UpdateTagCategory :one
UPDATE gfg_tag_category SET name=sqlc.arg(name),name_en=sqlc.arg(name_en),info=sqlc.arg(info),info_en=sqlc.arg(info_en),sort_order=sqlc.arg(sort_order),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) RETURNING *;

-- name: ArchiveTagCategory :one
UPDATE gfg_tag_category SET archived_at=CASE WHEN sqlc.arg(archive)::boolean THEN COALESCE(archived_at,NOW()::timestamp(0)) ELSE NULL END,update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) RETURNING *;

-- name: CountTagCategoryOptions :one
SELECT count(*) FROM gfg_tag_category WHERE archived_at IS NULL AND (sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR code ILIKE '%'||sqlc.arg(keyword)||'%' OR id::text=sqlc.arg(keyword));

-- name: ListTagCategoryOptions :many
SELECT id,name,name_en FROM gfg_tag_category WHERE archived_at IS NULL AND (sqlc.arg(keyword)::text='' OR name ILIKE '%'||sqlc.arg(keyword)||'%' OR name_en ILIKE '%'||sqlc.arg(keyword)||'%' OR code ILIKE '%'||sqlc.arg(keyword)||'%' OR id::text=sqlc.arg(keyword)) ORDER BY sort_order,id LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);
