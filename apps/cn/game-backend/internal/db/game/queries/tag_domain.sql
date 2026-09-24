-- name: ListPublicTagCategories :many
SELECT id,code,CASE WHEN sqlc.arg(lang)::text='en' THEN COALESCE(NULLIF(name_en,''),name) ELSE COALESCE(NULLIF(name,''),name_en) END::text AS name
FROM gfg_tag_category WHERE archived_at IS NULL ORDER BY sort_order,id;

-- name: LockRecommendationDomain :exec
SELECT pg_advisory_xact_lock(hashtextextended('gfg.tag-domain',0));

-- name: DeleteSourceRecommendations :exec
DELETE FROM gfg_game_recommendations WHERE source_game_id=sqlc.arg(source_game_id);

-- name: StoreRecommendation :exec
INSERT INTO gfg_game_recommendations(source_game_id,target_game_id,score,display_score,rank,reason_json,algorithm_version,computed_at)
VALUES(sqlc.arg(source_game_id),sqlc.arg(target_game_id),sqlc.arg(score),sqlc.arg(display_score),sqlc.arg(rank),sqlc.arg(reason_json),sqlc.arg(algorithm_version),sqlc.arg(computed_at));

-- name: ListPublicTags :many
SELECT t.id::text AS id,t.code,
 CASE WHEN sqlc.arg(lang)::text='en' THEN COALESCE(NULLIF(t.name_en,''),t.name) ELSE COALESCE(NULLIF(t.name,''),t.name_en) END::text AS name,
 t.category_id::text AS category_id,c.code AS category_code,
 CASE WHEN sqlc.arg(lang)::text='en' THEN COALESCE(NULLIF(c.name_en,''),c.name) ELSE COALESCE(NULLIF(c.name,''),c.name_en) END::text AS category_name,
 COALESCE(tc.game_count,0)::integer AS game_count
FROM gfg_tag t JOIN gfg_tag_category c ON c.id=t.category_id
LEFT JOIN (SELECT gt.tag_id,count(*) AS game_count FROM gfg_game_tag gt
 JOIN gfg_game_details d ON d.game_id=gt.game_id GROUP BY gt.tag_id) tc ON tc.tag_id=t.id
WHERE t.archived_at IS NULL AND c.archived_at IS NULL
ORDER BY game_count DESC,t.id;
