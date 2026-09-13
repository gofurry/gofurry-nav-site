-- name: NextHeroAssetID :one
WITH lock_row AS MATERIALIZED (SELECT pg_advisory_xact_lock(hashtext('gfn_home_hero_asset')::bigint))
SELECT (COALESCE(MAX(id),0)+1)::bigint FROM gfn_home_hero_asset CROSS JOIN lock_row;

-- name: GetHeroAsset :one
SELECT id,variant,name,object_key,enabled,deleted,deleted_at,create_time,update_time FROM gfn_home_hero_asset WHERE id=sqlc.arg(id) AND NOT deleted;

-- name: LockHeroAsset :one
SELECT id,variant,name,object_key,enabled,deleted,deleted_at,create_time,update_time FROM gfn_home_hero_asset WHERE id=sqlc.arg(id) AND NOT deleted FOR UPDATE;

-- name: ReplaceHeroAssetFile :execrows
UPDATE gfn_home_hero_asset SET object_key=sqlc.arg(object_key),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) AND NOT deleted;

-- name: DeleteHeroAsset :execrows
UPDATE gfn_home_hero_asset SET deleted=true,deleted_at=NOW(),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) AND NOT deleted;

-- name: NextBackgroundPatternID :one
WITH lock_row AS MATERIALIZED (SELECT pg_advisory_xact_lock(hashtext('gfn_background_pattern')::bigint))
SELECT (COALESCE(MAX(id),0)+1)::bigint FROM gfn_background_pattern CROSS JOIN lock_row;

-- name: GetBackgroundPattern :one
SELECT id,name,name_en,object_key,light_color,dark_color,light_opacity::double precision AS light_opacity,dark_opacity::double precision AS dark_opacity,default_size_px,enabled,sort_order,deleted,deleted_at,create_time,update_time FROM gfn_background_pattern WHERE id=sqlc.arg(id) AND NOT deleted;

-- name: LockBackgroundPattern :one
SELECT id,name,name_en,object_key,light_color,dark_color,light_opacity::double precision AS light_opacity,dark_opacity::double precision AS dark_opacity,default_size_px,enabled,sort_order,deleted,deleted_at,create_time,update_time FROM gfn_background_pattern WHERE id=sqlc.arg(id) AND NOT deleted FOR UPDATE;

-- name: ReplaceBackgroundPatternFile :execrows
UPDATE gfn_background_pattern SET object_key=sqlc.arg(object_key),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) AND NOT deleted;

-- name: DeleteBackgroundPattern :execrows
UPDATE gfn_background_pattern SET deleted=true,deleted_at=NOW(),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) AND NOT deleted;

-- name: ListHeroAssets :many
SELECT id,variant,name,object_key,enabled,deleted,deleted_at,create_time,update_time FROM gfn_home_hero_asset
WHERE NOT deleted AND variant=sqlc.arg(variant)
ORDER BY id DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: CountHeroAssets :one
SELECT COUNT(*)::bigint FROM gfn_home_hero_asset WHERE NOT deleted AND variant=sqlc.arg(variant);

-- name: CreateHeroAsset :one
INSERT INTO gfn_home_hero_asset (id,variant,name,object_key,enabled)
VALUES (sqlc.arg(id),sqlc.arg(variant),sqlc.arg(name),sqlc.arg(object_key),sqlc.arg(enabled)) RETURNING id,variant,name,object_key,enabled,deleted,deleted_at,create_time,update_time;

-- name: UpdateHeroAsset :execrows
UPDATE gfn_home_hero_asset SET name=sqlc.arg(name),enabled=sqlc.arg(enabled),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) AND NOT deleted;

-- name: ListBackgroundPatterns :many
SELECT id,name,name_en,object_key,light_color,dark_color,light_opacity::double precision AS light_opacity,dark_opacity::double precision AS dark_opacity,default_size_px,enabled,sort_order,deleted,deleted_at,create_time,update_time FROM gfn_background_pattern WHERE NOT deleted
ORDER BY sort_order,id LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: CountBackgroundPatterns :one
SELECT COUNT(*)::bigint FROM gfn_background_pattern WHERE NOT deleted;

-- name: CreateBackgroundPattern :one
INSERT INTO gfn_background_pattern (id,name,name_en,object_key,light_color,dark_color,light_opacity,dark_opacity,default_size_px,enabled,sort_order)
VALUES (sqlc.arg(id),sqlc.arg(name),sqlc.arg(name_en),sqlc.arg(object_key),sqlc.arg(light_color),sqlc.arg(dark_color),sqlc.arg(light_opacity)::double precision,sqlc.arg(dark_opacity)::double precision,sqlc.arg(default_size_px),sqlc.arg(enabled),sqlc.arg(sort_order)) RETURNING id,name,name_en,object_key,light_color,dark_color,light_opacity::double precision AS light_opacity,dark_opacity::double precision AS dark_opacity,default_size_px,enabled,sort_order,deleted,deleted_at,create_time,update_time;

-- name: UpdateBackgroundPattern :execrows
UPDATE gfn_background_pattern SET name=sqlc.arg(name),name_en=sqlc.arg(name_en),light_color=sqlc.arg(light_color),dark_color=sqlc.arg(dark_color),light_opacity=sqlc.arg(light_opacity)::double precision,dark_opacity=sqlc.arg(dark_opacity)::double precision,default_size_px=sqlc.arg(default_size_px),enabled=sqlc.arg(enabled),sort_order=sqlc.arg(sort_order),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) AND NOT deleted;

-- name: LockSiteIcon :one
SELECT id,icon FROM gfn_site WHERE id=sqlc.arg(id) AND deleted IS NOT TRUE FOR UPDATE;

-- name: ReplaceSiteIcon :execrows
UPDATE gfn_site SET icon=sqlc.narg(icon),update_time=NOW()::timestamp(0)
WHERE id=sqlc.arg(id) AND deleted IS NOT TRUE;
