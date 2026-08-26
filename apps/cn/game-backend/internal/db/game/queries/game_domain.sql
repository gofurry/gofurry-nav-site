-- name: GetReleaseStateByGame :one
SELECT game_id, availability, precision, exact_date, release_year,
       release_month, release_quarter, window_start, window_end, raw_text,
       source, source_region, source_locale, normalizer_version,
       observed_at, updated_at
FROM gfg_game_release_state
WHERE game_id = sqlc.arg(game_id);

-- name: GetFirstAvailableByGame :one
SELECT game_id, precision, exact_date, release_year, release_month,
       release_quarter, window_start, window_end, source, inferred,
       source_raw, source_observed_at, normalizer_version,
       established_at, updated_at
FROM gfg_game_first_available
WHERE game_id = sqlc.arg(game_id);

-- name: ListLanguagesByGame :many
SELECT id, game_id, language_code, steam_name, steam_api_code, steam_web_code,
       tier, interface_supported, subtitles_supported, full_audio_supported,
       sort_order, source, source_region, source_locale, normalizer_version,
       observed_at, updated_at
FROM gfg_game_languages
WHERE game_id = sqlc.arg(game_id)
ORDER BY sort_order, id;

-- name: BatchReleaseStatesByGames :many
SELECT game_id, availability, precision, exact_date, release_year,
       release_month, release_quarter, window_start, window_end, raw_text,
       source, source_region, source_locale, normalizer_version,
       observed_at, updated_at
FROM gfg_game_release_state
WHERE game_id = ANY(sqlc.arg(game_ids)::bigint[])
ORDER BY game_id;

-- name: BatchFirstAvailableByGames :many
SELECT game_id, precision, exact_date, release_year, release_month,
       release_quarter, window_start, window_end, source, inferred,
       source_raw, source_observed_at, normalizer_version,
       established_at, updated_at
FROM gfg_game_first_available
WHERE game_id = ANY(sqlc.arg(game_ids)::bigint[])
ORDER BY game_id;
