-- GoFurry game collector v2 player count contract extension.
-- This migration is additive and keeps alpha.3 installations compatible.

BEGIN;

ALTER TABLE gfg_game_v2_player_counts
    ADD COLUMN IF NOT EXISTS run_id TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_player_counts_run
    ON gfg_game_v2_player_counts (run_id)
    WHERE run_id <> '';

COMMIT;
