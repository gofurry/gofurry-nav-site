-- name: RandomHeroAsset :one
SELECT id,object_key FROM gfn_home_hero_asset
WHERE enabled AND NOT deleted AND variant=sqlc.arg(variant)
ORDER BY random() LIMIT 1;

-- name: PublicBackgroundPatterns :many
SELECT id,name,name_en,object_key,light_color,dark_color,light_opacity::double precision AS light_opacity,dark_opacity::double precision AS dark_opacity,default_size_px
FROM gfn_background_pattern WHERE enabled AND NOT deleted ORDER BY sort_order,id;
