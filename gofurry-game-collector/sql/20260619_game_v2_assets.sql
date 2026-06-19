CREATE TABLE IF NOT EXISTS gfg_game_v2_assets (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL,
    appid BIGINT NOT NULL,
    asset_type TEXT NOT NULL,
    asset_family TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'steam',
    lang TEXT NOT NULL DEFAULT '',
    media_key TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL DEFAULT '',
    thumbnail_url TEXT NOT NULL DEFAULT '',
    format TEXT NOT NULL DEFAULT '',
    exists BOOLEAN,
    status_code INTEGER NOT NULL DEFAULT 0,
    content_type TEXT NOT NULL DEFAULT '',
    content_length BIGINT NOT NULL DEFAULT 0,
    extra JSONB NOT NULL DEFAULT '{}'::jsonb,
    sort_order INTEGER NOT NULL DEFAULT 0,
    checked_at TIMESTAMPTZ,
    collected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_gfg_game_v2_assets_item
    ON gfg_game_v2_assets (game_id, asset_type, lang, media_key);

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_assets_game_family
    ON gfg_game_v2_assets (game_id, asset_family, sort_order);

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_assets_app_type
    ON gfg_game_v2_assets (appid, asset_type, lang);

CREATE INDEX IF NOT EXISTS idx_gfg_game_v2_assets_exists
    ON gfg_game_v2_assets (game_id, asset_type)
    WHERE exists IS DISTINCT FROM false;
