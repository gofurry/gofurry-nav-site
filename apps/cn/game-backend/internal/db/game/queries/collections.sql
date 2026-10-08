-- name: CountPublishedCollections :one
WITH selected_collections AS (SELECT id FROM gfg_game_collection WHERE status = 'published'),
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
), visible_members AS (
    SELECT i.collection_id, g.name, g.name_en,
        (fa.game_id IS NOT NULL OR r.availability = 'available') AS released,
        (fa.game_id IS NULL AND r.availability = 'upcoming') AS upcoming
    FROM effective_members i
    JOIN gfg_game g ON g.id = i.game_id
    LEFT JOIN gfg_game_first_available fa ON fa.game_id = g.id
    LEFT JOIN gfg_game_release_state r ON r.game_id = g.id
    WHERE sqlc.arg(include_adult)::boolean OR NOT EXISTS (
        SELECT 1 FROM gfg_game_tag gt JOIN gfg_tag t ON t.id = gt.tag_id
        WHERE gt.game_id = g.id AND t.code = 'adult'
    )
), membership AS (
    SELECT collection_id, count(*) AS visible_count,
        bool_or(released) AS has_released, bool_or(upcoming) AS has_upcoming,
        bool_or(strpos(lower(name), lower(sqlc.arg(keyword)::text)) > 0
            OR strpos(lower(name_en), lower(sqlc.arg(keyword)::text)) > 0) AS member_match
    FROM visible_members GROUP BY collection_id
), filtered AS (
    SELECT c.*, COALESCE(m.visible_count, 0)::bigint AS visible_count
    FROM gfg_game_collection c LEFT JOIN membership m ON m.collection_id = c.id
    WHERE c.status = 'published'
      AND (sqlc.arg(keyword)::text = '' OR m.member_match
        OR strpos(lower(c.code), lower(sqlc.arg(keyword)::text)) > 0
        OR strpos(lower(c.name), lower(sqlc.arg(keyword)::text)) > 0
        OR strpos(lower(c.name_en), lower(sqlc.arg(keyword)::text)) > 0
        OR strpos(lower(c.info), lower(sqlc.arg(keyword)::text)) > 0
        OR strpos(lower(c.info_en), lower(sqlc.arg(keyword)::text)) > 0)
      AND (sqlc.arg(phase)::text = 'all'
        OR (sqlc.arg(phase)::text = 'released' AND m.has_released)
        OR (sqlc.arg(phase)::text = 'upcoming' AND m.has_upcoming)
        OR (sqlc.arg(phase)::text = 'mixed' AND m.has_released AND m.has_upcoming))
)
SELECT count(*) FROM filtered;

-- name: ListPublishedCollections :many
WITH selected_collections AS (SELECT id FROM gfg_game_collection WHERE status = 'published'),
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
), visible_members AS (
    SELECT i.collection_id, g.name, g.name_en,
        (fa.game_id IS NOT NULL OR r.availability = 'available') AS released,
        (fa.game_id IS NULL AND r.availability = 'upcoming') AS upcoming
    FROM effective_members i
    JOIN gfg_game g ON g.id = i.game_id
    LEFT JOIN gfg_game_first_available fa ON fa.game_id = g.id
    LEFT JOIN gfg_game_release_state r ON r.game_id = g.id
    WHERE sqlc.arg(include_adult)::boolean OR NOT EXISTS (
        SELECT 1 FROM gfg_game_tag gt JOIN gfg_tag t ON t.id = gt.tag_id
        WHERE gt.game_id = g.id AND t.code = 'adult'
    )
), membership AS (
    SELECT collection_id, count(*) AS visible_count,
        bool_or(released) AS has_released, bool_or(upcoming) AS has_upcoming,
        bool_or(strpos(lower(name), lower(sqlc.arg(keyword)::text)) > 0
            OR strpos(lower(name_en), lower(sqlc.arg(keyword)::text)) > 0) AS member_match
    FROM visible_members GROUP BY collection_id
), filtered AS (
    SELECT c.*, COALESCE(m.visible_count, 0)::bigint AS visible_count
    FROM gfg_game_collection c LEFT JOIN membership m ON m.collection_id = c.id
    WHERE c.status = 'published'
      AND (sqlc.arg(keyword)::text = '' OR m.member_match
        OR strpos(lower(c.code), lower(sqlc.arg(keyword)::text)) > 0
        OR strpos(lower(c.name), lower(sqlc.arg(keyword)::text)) > 0
        OR strpos(lower(c.name_en), lower(sqlc.arg(keyword)::text)) > 0
        OR strpos(lower(c.info), lower(sqlc.arg(keyword)::text)) > 0
        OR strpos(lower(c.info_en), lower(sqlc.arg(keyword)::text)) > 0)
      AND (sqlc.arg(phase)::text = 'all'
        OR (sqlc.arg(phase)::text = 'released' AND m.has_released)
        OR (sqlc.arg(phase)::text = 'upcoming' AND m.has_upcoming)
        OR (sqlc.arg(phase)::text = 'mixed' AND m.has_released AND m.has_upcoming))
)
SELECT id, code, name, name_en, info, info_en, published_at
FROM filtered
ORDER BY
    CASE WHEN sqlc.arg(sort)::text = 'count_desc' THEN visible_count END DESC,
    CASE WHEN sqlc.arg(sort)::text = 'count_asc' THEN visible_count END ASC,
    CASE WHEN sqlc.arg(sort)::text = 'name_asc' THEN
        CASE WHEN sqlc.arg(lang)::text = 'en' THEN CASE WHEN btrim(name_en) = '' THEN name ELSE name_en END
        ELSE CASE WHEN btrim(name) = '' THEN name_en ELSE name END END END ASC,
    CASE WHEN sqlc.arg(sort)::text = 'name_desc' THEN
        CASE WHEN sqlc.arg(lang)::text = 'en' THEN CASE WHEN btrim(name_en) = '' THEN name ELSE name_en END
        ELSE CASE WHEN btrim(name) = '' THEN name_en ELSE name END END END DESC,
    CASE WHEN sqlc.arg(sort)::text IN ('name_asc', 'name_desc') THEN id END ASC,
    published_at DESC, id DESC
LIMIT sqlc.arg(page_size)::bigint OFFSET sqlc.arg(page_offset)::bigint;

-- name: BatchCollectionTimelineDecorations :many
WITH ratings AS (
    SELECT game_id, AVG(score)::double precision AS average, COUNT(*)::bigint AS review_count
    FROM gfg_game_comment WHERE game_id = ANY(sqlc.arg(game_ids)::bigint[])
    GROUP BY game_id
)
SELECT g.id AS game_id, pt.code AS primary_code, st.code AS secondary_code,
    CASE WHEN sqlc.arg(lang)::text = 'en' THEN COALESCE(NULLIF(pt.name_en, ''), pt.name, '')
         ELSE COALESCE(NULLIF(pt.name, ''), pt.name_en, '') END::text AS primary_name,
    CASE WHEN sqlc.arg(lang)::text = 'en' THEN COALESCE(NULLIF(st.name_en, ''), st.name, '')
         ELSE COALESCE(NULLIF(st.name, ''), st.name_en, '') END::text AS secondary_name,
    COALESCE(r.average, 0)::double precision AS average,
    COALESCE(r.review_count, 0)::bigint AS review_count,
    COALESCE(online.count, 0)::bigint AS online_count, online.collected_at AS online_collected_at,
    CASE WHEN jsonb_typeof(g.groups::jsonb) = 'array' THEN (
           SELECT COUNT(*) FROM jsonb_array_elements(g.groups::jsonb) AS entry
           WHERE btrim(COALESCE(entry->>'key', '')) <> ''
             AND btrim(COALESCE(entry->>'value', '')) <> ''
         )
         ELSE 0 END::integer AS community_count
FROM gfg_game g
LEFT JOIN gfg_game_tag pr ON pr.game_id = g.id AND pr.role = 'primary'
LEFT JOIN gfg_tag pt ON pt.id = pr.tag_id
LEFT JOIN gfg_game_tag sr ON sr.game_id = g.id AND sr.role = 'secondary'
LEFT JOIN gfg_tag st ON st.id = sr.tag_id
LEFT JOIN ratings r ON r.game_id = g.id
LEFT JOIN LATERAL (
    SELECT pc.count, pc.collected_at FROM gfg_game_player_counts pc
    WHERE pc.game_id = g.id AND pc.status = 'success'
    ORDER BY pc.collected_at DESC, pc.id DESC LIMIT 1
) online ON true
WHERE g.id = ANY(sqlc.arg(game_ids)::bigint[])
ORDER BY g.id;

-- name: GetPublishedCollection :one
SELECT id, code, name, name_en, info, info_en, published_at
FROM gfg_game_collection WHERE code = sqlc.arg(code) AND status = 'published';

-- name: ListPublishedCollectionHomeSlots :many
WITH selected_collections AS (SELECT c.id FROM gfg_game_collection c JOIN gfg_game_collection_home_slot h ON h.collection_id=c.id WHERE c.status='published'),
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
SELECT s.slot, c.id, c.code, c.name, c.name_en, c.info, c.info_en, c.published_at,
    (SELECT count(*) FROM effective_members i
     WHERE i.collection_id = c.id
       AND (sqlc.arg(include_adult)::boolean OR NOT EXISTS (
           SELECT 1 FROM gfg_game_tag gt JOIN gfg_tag t ON t.id = gt.tag_id
           WHERE gt.game_id = i.game_id AND t.code = 'adult'
       ))) AS visible_game_count
FROM gfg_game_collection_home_slot s
JOIN gfg_game_collection c ON c.id = s.collection_id
WHERE c.status = 'published'
ORDER BY s.slot;

-- name: BatchCollectionMemberships :many
WITH selected_collections AS (SELECT id FROM gfg_game_collection WHERE id = ANY(sqlc.arg(collection_ids)::bigint[])),
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
SELECT collection_id, game_id FROM effective_members
ORDER BY collection_id, game_id;

-- name: BatchCollectionProjectionGames :many
SELECT g.id, g.name, g.name_en, g.info, g.info_en, g.header,
    COALESCE(d.name, '')::text AS detail_name, d.header_url AS detail_header,
    zh.game_id AS zh_id, zh.name AS zh_name, zh.short_description AS zh_summary,
    en.game_id AS en_id, en.name AS en_name, en.short_description AS en_summary,
    EXISTS (SELECT 1 FROM gfg_game_tag gt JOIN gfg_tag t ON t.id = gt.tag_id
        WHERE gt.game_id = g.id AND t.code = 'adult') AS adult
FROM gfg_game g
LEFT JOIN gfg_game_details d ON d.game_id = g.id
LEFT JOIN gfg_game_localized_details zh ON zh.game_id = g.id AND zh.lang = 'zh'
LEFT JOIN gfg_game_localized_details en ON en.game_id = g.id AND en.lang = 'en'
WHERE g.id = ANY(sqlc.arg(game_ids)::bigint[])
ORDER BY g.id;

-- name: BatchCollectionHeaderMedia :many
SELECT game_id, url FROM gfg_game_media
WHERE game_id = ANY(sqlc.arg(game_ids)::bigint[]) AND media_type = 'header'
ORDER BY game_id, media_type, sort_order, id;

-- name: BatchCollectionHeaderAssets :many
SELECT game_id, asset_type, lang, url, exists FROM gfg_game_assets
WHERE game_id = ANY(sqlc.arg(game_ids)::bigint[]) AND asset_type IN ('header', 'header_2x')
ORDER BY game_id, asset_family, sort_order, id;

-- name: CountCollectionBrowse :one
SELECT count(*) FROM gfg_game_collection WHERE status = 'published';

-- name: ListCollectionBrowse :many
SELECT id, code, name, name_en, info, info_en, published_at
FROM gfg_game_collection WHERE status = 'published'
ORDER BY published_at DESC, id DESC
LIMIT sqlc.arg(page_size)::bigint OFFSET sqlc.arg(page_offset)::bigint;
