-- Collection writes serialize before row locks; this is unrelated to Collector Control Plane.
-- name: LockGameCollectionDomain :exec
SELECT pg_advisory_xact_lock(hashtextextended('gfg.game-collection-domain',0));

-- name: LockCollectionForUpdate :one
SELECT * FROM gfg_game_collection WHERE id=sqlc.arg(id) FOR UPDATE;

-- name: ListCuratedCollectionIDs :many
WITH selected_collections AS (SELECT id FROM gfg_game_collection WHERE (sqlc.arg(status)::text='' OR status=sqlc.arg(status))),
candidate_members AS (
    SELECT i.collection_id, i.game_id FROM gfg_game_collection_item i
    JOIN selected_collections c ON c.id = i.collection_id
    UNION
    SELECT rule.collection_id, gt.game_id FROM gfg_game_collection_tag rule
    JOIN selected_collections c ON c.id = rule.collection_id
    JOIN gfg_tag t ON t.id = rule.tag_id AND t.archived_at IS NULL
    JOIN gfg_tag_category cat ON cat.id = t.category_id AND cat.archived_at IS NULL
    JOIN gfg_game_tag gt ON gt.tag_id = t.id
), effective_members AS (
    SELECT m.collection_id, m.game_id FROM candidate_members m
    WHERE NOT EXISTS (SELECT 1 FROM gfg_game_collection_exclusion x
        WHERE x.collection_id = m.collection_id AND x.game_id = m.game_id)
)
SELECT c.id FROM gfg_game_collection c
WHERE (sqlc.arg(keyword)::text='' OR c.code ILIKE '%'||sqlc.arg(keyword)||'%' OR c.name ILIKE '%'||sqlc.arg(keyword)||'%' OR c.name_en ILIKE '%'||sqlc.arg(keyword)||'%')
AND (sqlc.arg(status)::text='' OR c.status=sqlc.arg(status))
AND (NOT sqlc.arg(home_eligible)::boolean OR (c.status='published' AND EXISTS (
 SELECT 1 FROM effective_members i WHERE i.collection_id=c.id
 AND NOT EXISTS (SELECT 1 FROM gfg_game_tag gt JOIN gfg_tag t ON t.id=gt.tag_id WHERE gt.game_id=i.game_id AND t.code='adult')
)))
ORDER BY c.updated_at DESC,c.id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: CountCuratedCollections :one
WITH selected_collections AS (SELECT id FROM gfg_game_collection WHERE (sqlc.arg(status)::text='' OR status=sqlc.arg(status))),
candidate_members AS (
    SELECT i.collection_id, i.game_id FROM gfg_game_collection_item i
    JOIN selected_collections c ON c.id = i.collection_id
    UNION
    SELECT rule.collection_id, gt.game_id FROM gfg_game_collection_tag rule
    JOIN selected_collections c ON c.id = rule.collection_id
    JOIN gfg_tag t ON t.id = rule.tag_id AND t.archived_at IS NULL
    JOIN gfg_tag_category cat ON cat.id = t.category_id AND cat.archived_at IS NULL
    JOIN gfg_game_tag gt ON gt.tag_id = t.id
), effective_members AS (
    SELECT m.collection_id, m.game_id FROM candidate_members m
    WHERE NOT EXISTS (SELECT 1 FROM gfg_game_collection_exclusion x
        WHERE x.collection_id = m.collection_id AND x.game_id = m.game_id)
)
SELECT count(*) FROM gfg_game_collection c
WHERE (sqlc.arg(keyword)::text='' OR c.code ILIKE '%'||sqlc.arg(keyword)||'%' OR c.name ILIKE '%'||sqlc.arg(keyword)||'%' OR c.name_en ILIKE '%'||sqlc.arg(keyword)||'%')
AND (sqlc.arg(status)::text='' OR c.status=sqlc.arg(status))
AND (NOT sqlc.arg(home_eligible)::boolean OR (c.status='published' AND EXISTS (
 SELECT 1 FROM effective_members i WHERE i.collection_id=c.id
 AND NOT EXISTS (SELECT 1 FROM gfg_game_tag gt JOIN gfg_tag t ON t.id=gt.tag_id WHERE gt.game_id=i.game_id AND t.code='adult')
)));

-- name: GetCuratedCollectionSnapshots :many
WITH selected_collections AS (SELECT id FROM gfg_game_collection WHERE id=ANY(sqlc.arg(ids)::bigint[])),
candidate_members AS (
    SELECT i.collection_id, i.game_id FROM gfg_game_collection_item i
    JOIN selected_collections c ON c.id = i.collection_id
    UNION
    SELECT rule.collection_id, gt.game_id FROM gfg_game_collection_tag rule
    JOIN selected_collections c ON c.id = rule.collection_id
    JOIN gfg_tag t ON t.id = rule.tag_id AND t.archived_at IS NULL
    JOIN gfg_tag_category cat ON cat.id = t.category_id AND cat.archived_at IS NULL
    JOIN gfg_game_tag gt ON gt.tag_id = t.id
), effective_members AS (
    SELECT m.collection_id, m.game_id FROM candidate_members m
    WHERE NOT EXISTS (SELECT 1 FROM gfg_game_collection_exclusion x
        WHERE x.collection_id = m.collection_id AND x.game_id = m.game_id)
)
SELECT sqlc.embed(c), h.slot AS home_slot,
 (SELECT count(*) FROM effective_members i WHERE i.collection_id=c.id)::bigint AS member_count,
 (SELECT count(*) FROM effective_members i WHERE i.collection_id=c.id
  AND NOT EXISTS (SELECT 1 FROM gfg_game_tag gt JOIN gfg_tag t ON t.id=gt.tag_id WHERE gt.game_id=i.game_id AND t.code='adult'))::bigint AS sfw_member_count
FROM gfg_game_collection c LEFT JOIN gfg_game_collection_home_slot h ON h.collection_id=c.id
WHERE c.id=ANY(sqlc.arg(ids)::bigint[]) ORDER BY c.updated_at DESC,c.id DESC;

-- name: InsertCuratedCollection :one
INSERT INTO gfg_game_collection(code,name,name_en,info,info_en)
VALUES(sqlc.arg(code),sqlc.arg(name),sqlc.arg(name_en),sqlc.arg(info),sqlc.arg(info_en)) RETURNING id;

-- name: UpdateCuratedCollectionContent :exec
UPDATE gfg_game_collection SET name=sqlc.arg(name),name_en=sqlc.arg(name_en),info=sqlc.arg(info),info_en=sqlc.arg(info_en),version=version+1,updated_at=now() WHERE id=sqlc.arg(id);

-- name: BumpCuratedCollectionVersion :exec
UPDATE gfg_game_collection SET version=version+1,updated_at=now() WHERE id=sqlc.arg(id);

-- name: TransitionCuratedCollection :exec
UPDATE gfg_game_collection SET status=sqlc.arg(status),
 published_at=CASE WHEN sqlc.arg(status)='published' THEN now() ELSE published_at END,
 archived_at=CASE WHEN sqlc.arg(status)='archived' THEN now() ELSE NULL END,
 version=version+1,updated_at=now() WHERE id=sqlc.arg(id);

-- name: GetCuratedCollectionMembers :many
SELECT g.id AS game_id,g.name,g.name_en,g.appid,
 EXISTS(SELECT 1 FROM gfg_game_tag gt JOIN gfg_tag t ON t.id=gt.tag_id WHERE gt.game_id=g.id AND t.code='adult')::boolean AS adult
FROM gfg_game_collection_item i JOIN gfg_game g ON g.id=i.game_id
WHERE i.collection_id=sqlc.arg(collection_id) ORDER BY g.id;

-- name: ExistingCuratedCollectionGameIDs :many
SELECT id FROM gfg_game WHERE id=ANY(sqlc.arg(ids)::bigint[]) ORDER BY id FOR KEY SHARE;

-- name: DeleteRemovedCollectionMembers :exec
DELETE FROM gfg_game_collection_item WHERE collection_id=sqlc.arg(collection_id) AND NOT(game_id=ANY(sqlc.arg(game_ids)::bigint[]));

-- name: InsertCollectionMembers :exec
INSERT INTO gfg_game_collection_item(collection_id,game_id)
SELECT sqlc.arg(collection_id)::bigint,unnest(sqlc.arg(game_ids)::bigint[])
ON CONFLICT(collection_id,game_id) DO NOTHING;

-- name: RemoveCollectionHomeSlot :exec
DELETE FROM gfg_game_collection_home_slot WHERE collection_id=sqlc.arg(collection_id);

-- name: GetCollectionHomePlacements :many
SELECT slot,collection_id FROM gfg_game_collection_home_slot ORDER BY slot;

-- name: ClearCollectionHomePlacements :exec
DELETE FROM gfg_game_collection_home_slot;

-- name: InsertCollectionHomePlacements :exec
INSERT INTO gfg_game_collection_home_slot(slot,collection_id)
SELECT unnest(sqlc.arg(slots)::smallint[]),unnest(sqlc.arg(collection_ids)::bigint[]);

-- name: GetCollectionRuleTags :many
SELECT t.id AS tag_id, t.code, t.name, t.name_en,
 (t.archived_at IS NULL AND cat.archived_at IS NULL)::boolean AS active
FROM gfg_game_collection_tag rule JOIN gfg_tag t ON t.id=rule.tag_id
JOIN gfg_tag_category cat ON cat.id=t.category_id
WHERE rule.collection_id=sqlc.arg(collection_id) ORDER BY t.id;

-- name: ValidateCollectionRuleTags :many
SELECT t.id, (t.archived_at IS NULL AND cat.archived_at IS NULL)::boolean AS active,
 EXISTS(SELECT 1 FROM gfg_game_collection_tag rule WHERE rule.collection_id=sqlc.arg(collection_id) AND rule.tag_id=t.id)::boolean AS already_bound
FROM gfg_tag t JOIN gfg_tag_category cat ON cat.id=t.category_id
WHERE t.id=ANY(sqlc.arg(tag_ids)::bigint[]) ORDER BY t.id FOR KEY SHARE OF t;

-- name: GetCollectionCompositionMembers :many
WITH sources AS (
 SELECT gt.game_id, true AS automatic, false AS manual, false AS excluded
 FROM gfg_game_collection_tag rule
 JOIN gfg_tag t ON t.id=rule.tag_id AND t.archived_at IS NULL
 JOIN gfg_tag_category cat ON cat.id=t.category_id AND cat.archived_at IS NULL
 JOIN gfg_game_tag gt ON gt.tag_id=t.id
 WHERE rule.collection_id=sqlc.arg(collection_id)
 UNION ALL
 SELECT game_id, false, true, false FROM gfg_game_collection_item WHERE collection_id=sqlc.arg(collection_id)
 UNION ALL
 SELECT game_id, false, false, true FROM gfg_game_collection_exclusion WHERE collection_id=sqlc.arg(collection_id)
), membership AS (
 SELECT game_id, bool_or(automatic)::boolean AS automatic, bool_or(manual)::boolean AS manual, bool_or(excluded)::boolean AS excluded
 FROM sources GROUP BY game_id
)
SELECT g.id AS game_id, g.name, g.name_en, g.appid, m.automatic, m.manual, m.excluded,
 EXISTS(SELECT 1 FROM gfg_game_tag gt JOIN gfg_tag t ON t.id=gt.tag_id WHERE gt.game_id=g.id AND t.code='adult')::boolean AS adult
FROM membership m JOIN gfg_game g ON g.id=m.game_id ORDER BY g.id;

-- name: GetCollectionExcludedIDs :many
SELECT game_id FROM gfg_game_collection_exclusion WHERE collection_id=sqlc.arg(collection_id) ORDER BY game_id;

-- name: DeleteRemovedCollectionRules :exec
DELETE FROM gfg_game_collection_tag WHERE collection_id=sqlc.arg(collection_id) AND NOT(tag_id=ANY(sqlc.arg(tag_ids)::bigint[]));

-- name: InsertCollectionRules :exec
INSERT INTO gfg_game_collection_tag(collection_id,tag_id)
SELECT sqlc.arg(collection_id)::bigint,unnest(sqlc.arg(tag_ids)::bigint[])
ON CONFLICT(collection_id,tag_id) DO NOTHING;

-- name: DeleteRemovedCollectionExclusions :exec
DELETE FROM gfg_game_collection_exclusion WHERE collection_id=sqlc.arg(collection_id) AND NOT(game_id=ANY(sqlc.arg(game_ids)::bigint[]));

-- name: InsertCollectionExclusions :exec
INSERT INTO gfg_game_collection_exclusion(collection_id,game_id)
SELECT sqlc.arg(collection_id)::bigint,unnest(sqlc.arg(game_ids)::bigint[])
ON CONFLICT(collection_id,game_id) DO NOTHING;
