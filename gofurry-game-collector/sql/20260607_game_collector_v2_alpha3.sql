-- GoFurry game collector v2 storage contract draft.
-- This migration is intentionally additive. It does not touch v1 tables.

BEGIN;

CREATE TABLE IF NOT EXISTS gfg_game_v2_details (
    game_id BIGINT PRIMARY KEY,
    appid BIGINT NOT NULL UNIQUE,
    source TEXT NOT NULL DEFAULT 'steam',
    type TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    is_free BOOLEAN NOT NULL DEFAULT FALSE,
    website TEXT NOT NULL DEFAULT '',
    header_url TEXT NOT NULL DEFAULT '',
    developers JSONB NOT NULL DEFAULT '[]'::jsonb,
    publishers JSONB NOT NULL DEFAULT '[]'::jsonb,
    release_coming_soon BOOLEAN NOT NULL DEFAULT FALSE,
    release_date_text TEXT NOT NULL DEFAULT '',
    platforms JSONB NOT NULL DEFAULT '{}'::jsonb,
    supported_languages TEXT NOT NULL DEFAULT '',
    support_info JSONB NOT NULL DEFAULT '{}'::jsonb,
    content_descriptors JSONB NOT NULL DEFAULT '{}'::jsonb,
    ratings JSONB NOT NULL DEFAULT '[]'::jsonb,
    collected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS gfg_game_v2_localized_details (
    game_id BIGINT NOT NULL,
    appid BIGINT NOT NULL,
    lang TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    short_description TEXT NOT NULL DEFAULT '',
    detailed_description TEXT NOT NULL DEFAULT '',
    about_the_game TEXT NOT NULL DEFAULT '',
    collected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (game_id, lang)
);

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_localized_details_app_lang
    ON gfg_game_v2_localized_details (appid, lang);

CREATE TABLE IF NOT EXISTS gfg_game_v2_prices (
    game_id BIGINT NOT NULL,
    appid BIGINT NOT NULL,
    region TEXT NOT NULL,
    is_free BOOLEAN NOT NULL DEFAULT FALSE,
    currency TEXT NOT NULL DEFAULT '',
    initial_amount BIGINT NOT NULL DEFAULT 0,
    final_amount BIGINT NOT NULL DEFAULT 0,
    discount_percent BIGINT NOT NULL DEFAULT 0,
    initial_formatted TEXT NOT NULL DEFAULT '',
    final_formatted TEXT NOT NULL DEFAULT '',
    collected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (game_id, region)
);

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_prices_app_region
    ON gfg_game_v2_prices (appid, region);

CREATE TABLE IF NOT EXISTS gfg_game_v2_media (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL,
    appid BIGINT NOT NULL,
    media_type TEXT NOT NULL,
    media_key TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL DEFAULT '',
    thumbnail_url TEXT NOT NULL DEFAULT '',
    extra JSONB NOT NULL DEFAULT '{}'::jsonb,
    sort_order INTEGER NOT NULL DEFAULT 0,
    collected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_gfg_game_v2_media_item
    ON gfg_game_v2_media (game_id, media_type, media_key);

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_media_app_type
    ON gfg_game_v2_media (appid, media_type, sort_order);

CREATE TABLE IF NOT EXISTS gfg_game_v2_requirements (
    game_id BIGINT PRIMARY KEY,
    appid BIGINT NOT NULL UNIQUE,
    pc JSONB NOT NULL DEFAULT '{}'::jsonb,
    mac JSONB NOT NULL DEFAULT '{}'::jsonb,
    linux JSONB NOT NULL DEFAULT '{}'::jsonb,
    collected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS gfg_game_v2_detail_snapshots (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL,
    appid BIGINT NOT NULL,
    lang TEXT NOT NULL,
    region TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT 'steam',
    payload_hash TEXT NOT NULL DEFAULT '',
    raw_payload JSONB NOT NULL,
    collected_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_detail_snapshots_lookup
    ON gfg_game_v2_detail_snapshots (appid, lang, region, collected_at DESC);

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_detail_snapshots_hash
    ON gfg_game_v2_detail_snapshots (appid, lang, region, payload_hash);

CREATE TABLE IF NOT EXISTS gfg_game_v2_news (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL,
    appid BIGINT NOT NULL,
    lang TEXT NOT NULL,
    event_gid TEXT NOT NULL DEFAULT '',
    announcement_gid TEXT NOT NULL DEFAULT '',
    forum_topic_id TEXT NOT NULL DEFAULT '',
    headline TEXT NOT NULL DEFAULT '',
    raw_body TEXT NOT NULL DEFAULT '',
    html TEXT NOT NULL DEFAULT '',
    plain_text TEXT NOT NULL DEFAULT '',
    summary TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL DEFAULT '',
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    vote_up_count INTEGER NOT NULL DEFAULT 0,
    vote_down_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,
    raw_event JSONB NOT NULL DEFAULT '{}'::jsonb,
    published_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    collected_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_gfg_game_v2_news_event_lang
    ON gfg_game_v2_news (appid, lang, event_gid, announcement_gid);

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_news_feed
    ON gfg_game_v2_news (game_id, lang, published_at DESC NULLS LAST, collected_at DESC);

CREATE TABLE IF NOT EXISTS gfg_game_v2_player_counts (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL,
    appid BIGINT NOT NULL,
    count BIGINT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'success',
    upstream_status_code INTEGER NOT NULL DEFAULT 0,
    error_kind TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    collected_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_player_counts_latest
    ON gfg_game_v2_player_counts (game_id, collected_at DESC);

CREATE TABLE IF NOT EXISTS gfg_game_v2_collect_runs (
    id TEXT PRIMARY KEY,
    task_type TEXT NOT NULL,
    status TEXT NOT NULL,
    total_count INTEGER NOT NULL DEFAULT 0,
    success_count INTEGER NOT NULL DEFAULT 0,
    failed_count INTEGER NOT NULL DEFAULT 0,
    skipped_count INTEGER NOT NULL DEFAULT 0,
    error_kind TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_collect_runs_task_time
    ON gfg_game_v2_collect_runs (task_type, started_at DESC);

CREATE TABLE IF NOT EXISTS gfg_game_v2_collect_task_results (
    id BIGSERIAL PRIMARY KEY,
    run_id TEXT NOT NULL REFERENCES gfg_game_v2_collect_runs(id) ON DELETE CASCADE,
    task_type TEXT NOT NULL,
    status TEXT NOT NULL,
    game_id BIGINT NOT NULL,
    appid BIGINT NOT NULL,
    upstream_status_code INTEGER NOT NULL DEFAULT 0,
    traffic_bucket TEXT NOT NULL DEFAULT '',
    retry_count INTEGER NOT NULL DEFAULT 0,
    duration_millis BIGINT NOT NULL DEFAULT 0,
    error_kind TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_collect_task_results_run
    ON gfg_game_v2_collect_task_results (run_id, status);

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_collect_task_results_app_time
    ON gfg_game_v2_collect_task_results (appid, task_type, started_at DESC);

CREATE OR REPLACE FUNCTION gfg_game_v2_prune_detail_snapshots(
    p_appid BIGINT,
    p_lang TEXT,
    p_region TEXT,
    p_keep_count INTEGER DEFAULT 5
)
RETURNS INTEGER
LANGUAGE plpgsql
AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    WITH ranked AS (
        SELECT id,
               row_number() OVER (
                   PARTITION BY appid, lang, region
                   ORDER BY collected_at DESC, id DESC
               ) AS rn
        FROM gfg_game_v2_detail_snapshots
        WHERE appid = p_appid
          AND lang = p_lang
          AND region = p_region
    ),
    deleted AS (
        DELETE FROM gfg_game_v2_detail_snapshots s
        USING ranked r
        WHERE s.id = r.id
          AND r.rn > GREATEST(p_keep_count, 0)
        RETURNING s.id
    )
    SELECT count(*) INTO deleted_count FROM deleted;

    RETURN deleted_count;
END;
$$;

COMMIT;
