-- name: LockGameForUpdate :one
SELECT id, name, name_en, info, info_en, create_time, update_time,
       resources, groups, developers, publishers, appid, header, links,
       weight, primary_tag, secondary_tag, view_count
FROM gfg_game
WHERE id = sqlc.arg(id)
FOR UPDATE;

-- name: ResetSteamDerivedGameState :exec
WITH deleted_recommendations AS (
    DELETE FROM gfg_game_recommendations AS recommendations
    WHERE recommendations.source_game_id = sqlc.arg(target_id)
       OR recommendations.target_game_id = sqlc.arg(target_id)
), deleted_release_history AS (
    DELETE FROM gfg_game_release_history AS release_history
    WHERE release_history.game_id = sqlc.arg(target_id)
), deleted_first_available AS (
    DELETE FROM gfg_game_first_available AS first_available
    WHERE first_available.game_id = sqlc.arg(target_id)
), deleted_release_state AS (
    DELETE FROM gfg_game_release_state AS release_state
    WHERE release_state.game_id = sqlc.arg(target_id)
), deleted_languages AS (
    DELETE FROM gfg_game_languages AS languages
    WHERE languages.game_id = sqlc.arg(target_id)
), deleted_snapshots AS (
    DELETE FROM gfg_game_detail_snapshots AS snapshots
    WHERE snapshots.game_id = sqlc.arg(target_id)
), deleted_player_counts AS (
    DELETE FROM gfg_game_player_counts AS player_counts
    WHERE player_counts.game_id = sqlc.arg(target_id)
), deleted_news AS (
    DELETE FROM gfg_game_news AS news
    WHERE news.game_id = sqlc.arg(target_id)
), deleted_requirements AS (
    DELETE FROM gfg_game_requirements AS requirements
    WHERE requirements.game_id = sqlc.arg(target_id)
), deleted_assets AS (
    DELETE FROM gfg_game_assets AS assets
    WHERE assets.game_id = sqlc.arg(target_id)
), deleted_media AS (
    DELETE FROM gfg_game_media AS media
    WHERE media.game_id = sqlc.arg(target_id)
), deleted_prices AS (
    DELETE FROM gfg_game_prices AS prices
    WHERE prices.game_id = sqlc.arg(target_id)
), deleted_localized AS (
    DELETE FROM gfg_game_localized_details AS localized
    WHERE localized.game_id = sqlc.arg(target_id)
)
DELETE FROM gfg_game_details AS details
WHERE details.game_id = sqlc.arg(target_id);
