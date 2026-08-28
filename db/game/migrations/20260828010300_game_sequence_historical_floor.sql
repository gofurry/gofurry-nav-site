-- V3-P0.2 final historical Game-ID sequence floor.
-- This data correction is intentionally irreversible: lowering a sequence can
-- reissue an ID, so there is no safe Goose Down operation.
-- +goose Up

WITH sequence_state AS (
    SELECT last_value, is_called
    FROM public.gfg_game_id_seq
), historical_game_ids(game_id) AS (
    SELECT id FROM public.gfg_game
    UNION ALL SELECT scope_id FROM public.gfg_collection_jobs
        WHERE scope_type = 'game' AND scope_id IS NOT NULL
    UNION ALL SELECT game_id FROM public.gfg_collection_task_results
    UNION ALL SELECT game_id FROM public.gfg_game_assets
    UNION ALL SELECT game_id FROM public.gfg_game_comment
    UNION ALL SELECT game_id FROM public.gfg_game_daily
    UNION ALL SELECT game_id FROM public.gfg_game_detail_snapshots
    UNION ALL SELECT game_id FROM public.gfg_game_details
    UNION ALL SELECT game_id FROM public.gfg_game_first_available
    UNION ALL SELECT game_id FROM public.gfg_game_languages
    UNION ALL SELECT game_id FROM public.gfg_game_localized_details
    UNION ALL SELECT game_id FROM public.gfg_game_media
    UNION ALL SELECT game_id FROM public.gfg_game_news
    UNION ALL SELECT game_id FROM public.gfg_game_player_counts
    UNION ALL SELECT game_id FROM public.gfg_game_player_daily
    UNION ALL SELECT game_id FROM public.gfg_game_player_hourly
    UNION ALL SELECT game_id FROM public.gfg_game_price_daily
    UNION ALL SELECT game_id FROM public.gfg_game_prices
    UNION ALL SELECT source_game_id FROM public.gfg_game_recommendations
    UNION ALL SELECT target_game_id FROM public.gfg_game_recommendations
    UNION ALL SELECT game_id FROM public.gfg_game_release_history
    UNION ALL SELECT game_id FROM public.gfg_game_release_state
    UNION ALL SELECT game_id FROM public.gfg_game_requirements
    UNION ALL SELECT game_id FROM public.gfg_game_tracking_periods
    UNION ALL SELECT game_id FROM public.gfg_tag_map
), historical_bound AS (
    SELECT max(game_id) AS max_game_id FROM historical_game_ids
), target AS (
    SELECT GREATEST(
               COALESCE(historical_bound.max_game_id, 1),
               sequence_state.last_value,
               1
           ) AS sequence_value,
           historical_bound.max_game_id IS NOT NULL OR sequence_state.is_called AS sequence_called
    FROM historical_bound
    CROSS JOIN sequence_state
)
SELECT setval(
    'public.gfg_game_id_seq',
    sequence_value,
    sequence_called
)
FROM target;
