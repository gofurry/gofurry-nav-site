-- GoFurry game backend v2 similar recommendation storage.
-- Additive migration. It does not remove the legacy v1 CBF route.

BEGIN;

CREATE TABLE IF NOT EXISTS gfg_game_v2_recommendations (
    source_game_id BIGINT NOT NULL,
    target_game_id BIGINT NOT NULL,
    score DOUBLE PRECISION NOT NULL DEFAULT 0,
    display_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    rank INTEGER NOT NULL DEFAULT 0,
    reason_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    algorithm_version TEXT NOT NULL DEFAULT '',
    computed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (source_game_id, target_game_id)
);

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_recommendations_lookup
    ON gfg_game_v2_recommendations (source_game_id, algorithm_version, rank ASC, score DESC);

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_recommendations_computed_at
    ON gfg_game_v2_recommendations (computed_at);

COMMENT ON TABLE gfg_game_v2_recommendations IS
    'Precomputed game v2 similar recommendations. One algorithm_version represents one scoring contract.';

COMMENT ON COLUMN gfg_game_v2_recommendations.score IS
    'Raw hybrid content similarity score in range 0..1.';

COMMENT ON COLUMN gfg_game_v2_recommendations.display_score IS
    'Presentation score in range 0..1 after non-linear stretching.';

COMMENT ON COLUMN gfg_game_v2_recommendations.reason_json IS
    'Short explainable recommendation reasons for UI display and later tuning.';

COMMIT;
