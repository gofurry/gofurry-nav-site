-- name: LockGameTarget :one
SELECT id, appid
FROM gfg_game
WHERE id = sqlc.arg(game_id)
FOR UPDATE;

-- name: GetReleaseState :one
SELECT game_id, availability, precision, exact_date, release_year,
       release_month, release_quarter, window_start, window_end, raw_text,
       source, source_region, source_locale, normalizer_version,
       observed_at, updated_at
FROM gfg_game_release_state
WHERE game_id = sqlc.arg(game_id);

-- name: InsertReleaseState :exec
INSERT INTO gfg_game_release_state (
    game_id, availability, precision, exact_date, release_year,
    release_month, release_quarter, window_start, window_end, raw_text,
    source, source_region, source_locale, normalizer_version, observed_at
) VALUES (
    sqlc.arg(game_id), sqlc.arg(availability), sqlc.arg(precision),
    sqlc.narg(exact_date), sqlc.narg(release_year), sqlc.narg(release_month),
    sqlc.narg(release_quarter), sqlc.narg(window_start), sqlc.narg(window_end),
    sqlc.arg(raw_text), sqlc.arg(source), sqlc.arg(source_region),
    sqlc.arg(source_locale), sqlc.arg(normalizer_version), sqlc.arg(observed_at)
);

-- name: UpdateReleaseState :exec
UPDATE gfg_game_release_state
SET availability = sqlc.arg(availability),
    precision = sqlc.arg(precision),
    exact_date = sqlc.narg(exact_date),
    release_year = sqlc.narg(release_year),
    release_month = sqlc.narg(release_month),
    release_quarter = sqlc.narg(release_quarter),
    window_start = sqlc.narg(window_start),
    window_end = sqlc.narg(window_end),
    raw_text = sqlc.arg(raw_text),
    source = sqlc.arg(source),
    source_region = sqlc.arg(source_region),
    source_locale = sqlc.arg(source_locale),
    normalizer_version = sqlc.arg(normalizer_version),
    observed_at = sqlc.arg(observed_at),
    updated_at = now()
WHERE game_id = sqlc.arg(game_id);

-- name: InsertReleaseHistory :exec
INSERT INTO gfg_game_release_history (
    game_id, availability, precision, exact_date, release_year,
    release_month, release_quarter, window_start, window_end, raw_text,
    source, source_region, source_locale, normalizer_version, observed_at
) VALUES (
    sqlc.arg(game_id), sqlc.arg(availability), sqlc.arg(precision),
    sqlc.narg(exact_date), sqlc.narg(release_year), sqlc.narg(release_month),
    sqlc.narg(release_quarter), sqlc.narg(window_start), sqlc.narg(window_end),
    sqlc.arg(raw_text), sqlc.arg(source), sqlc.arg(source_region),
    sqlc.arg(source_locale), sqlc.arg(normalizer_version), sqlc.arg(observed_at)
);

-- name: InsertFirstAvailableIfAbsent :execrows
INSERT INTO gfg_game_first_available (
    game_id, precision, exact_date, release_year, release_month,
    release_quarter, window_start, window_end, source, inferred,
    source_raw, source_observed_at, normalizer_version
) VALUES (
    sqlc.arg(game_id), sqlc.arg(precision), sqlc.narg(exact_date),
    sqlc.arg(release_year), sqlc.narg(release_month),
    sqlc.narg(release_quarter), sqlc.arg(window_start), sqlc.arg(window_end),
    sqlc.arg(source), sqlc.arg(inferred), sqlc.arg(source_raw),
    sqlc.narg(source_observed_at), sqlc.arg(normalizer_version)
)
ON CONFLICT (game_id) DO NOTHING;

-- name: GetFirstAvailable :one
SELECT game_id, precision, exact_date, release_year, release_month,
       release_quarter, window_start, window_end, source, inferred,
       source_raw, source_observed_at, normalizer_version,
       established_at, updated_at
FROM gfg_game_first_available
WHERE game_id = sqlc.arg(game_id);

-- name: DeleteGameLanguages :exec
DELETE FROM gfg_game_languages
WHERE game_id = sqlc.arg(game_id);

-- name: InsertGameLanguage :exec
INSERT INTO gfg_game_languages (
    game_id, language_code, steam_name, steam_api_code, steam_web_code,
    tier, interface_supported, subtitles_supported, full_audio_supported,
    sort_order, source, source_region, source_locale, normalizer_version,
    observed_at
) VALUES (
    sqlc.arg(game_id), sqlc.narg(language_code), sqlc.arg(steam_name),
    sqlc.narg(steam_api_code), sqlc.narg(steam_web_code), sqlc.arg(tier),
    sqlc.narg(interface_supported), sqlc.narg(subtitles_supported),
    sqlc.narg(full_audio_supported), sqlc.arg(sort_order), sqlc.arg(source),
    sqlc.arg(source_region), sqlc.arg(source_locale),
    sqlc.arg(normalizer_version), sqlc.arg(observed_at)
);

-- name: ListLegacyFirstAvailableCandidates :many
SELECT g.id AS game_id,
       g.release_date,
       (d.game_id IS NOT NULL)::boolean AS has_current_state,
       COALESCE(d.release_coming_soon, false)::boolean AS release_coming_soon,
       d.collected_at AS source_observed_at
FROM gfg_game g
LEFT JOIN gfg_game_v2_details d ON d.game_id = g.id
WHERE btrim(g.release_date) <> ''
ORDER BY g.id;
