-- GoFurry game collector v2 stable observation extension.
-- This migration makes collect_runs match the unified v2 runner summary.

BEGIN;

ALTER TABLE gfg_game_v2_collect_runs
    ADD COLUMN IF NOT EXISTS partial_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS task_summary JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS duration_millis BIGINT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_collect_runs_status_time
    ON gfg_game_v2_collect_runs (status, started_at DESC);

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_collect_task_results_started_at
    ON gfg_game_v2_collect_task_results (started_at);

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_player_counts_collected_at
    ON gfg_game_v2_player_counts (collected_at);

COMMIT;
