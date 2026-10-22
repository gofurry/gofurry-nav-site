-- name: RandomHeroAsset :one
SELECT id,object_key FROM gfn_home_hero_asset
WHERE enabled AND NOT deleted AND variant=sqlc.arg(variant)
ORDER BY random() LIMIT 1;

-- name: PublicHeroAsset :one
SELECT id, name, object_key FROM gfn_home_hero_asset
WHERE id = sqlc.arg(id) AND variant = sqlc.arg(variant) AND enabled AND NOT deleted;

-- name: CountPublicHeroAssets :one
SELECT count(*) FROM gfn_home_hero_asset
WHERE variant = sqlc.arg(variant) AND enabled AND NOT deleted;

-- name: PublicHeroAssets :many
SELECT id, name, object_key FROM gfn_home_hero_asset
WHERE variant = sqlc.arg(variant) AND enabled AND NOT deleted
ORDER BY id
LIMIT sqlc.arg(page_size)::integer OFFSET sqlc.arg(page_offset)::integer;
