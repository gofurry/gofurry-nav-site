-- +goose Up
CREATE TABLE public.gfn_home_hero_asset (
    id bigint PRIMARY KEY,
    variant varchar(16) NOT NULL CHECK (variant IN ('desktop', 'mobile')),
    name varchar(120) NOT NULL CHECK (length(trim(name)) > 0),
    object_key varchar(255) NOT NULL,
    enabled boolean NOT NULL DEFAULT true,
    deleted boolean NOT NULL DEFAULT false,
    deleted_at timestamptz,
    create_time timestamp(0) NOT NULL DEFAULT NOW(),
    update_time timestamp(0) NOT NULL DEFAULT NOW(),
    CONSTRAINT gfn_home_hero_asset_key CHECK (object_key ~ ('^nav/hero/' || variant || '/[a-f0-9]{32}\.avif$'))
);
CREATE INDEX gfn_home_hero_asset_pool_idx ON public.gfn_home_hero_asset (variant,id) WHERE enabled AND NOT deleted;
COMMENT ON TABLE public.gfn_home_hero_asset IS 'Independent desktop/mobile AVIF random pools; no pairing or cross-pool fallback';
COMMENT ON COLUMN public.gfn_home_hero_asset.object_key IS 'Immutable provider-neutral managed object key';

CREATE TABLE public.gfn_background_pattern (
    id bigint PRIMARY KEY,
    name varchar(120) NOT NULL CHECK (length(trim(name)) > 0),
    name_en varchar(120) NOT NULL CHECK (length(trim(name_en)) > 0),
    object_key varchar(255) NOT NULL CHECK (object_key ~ '^nav/patterns/[a-f0-9]{32}\.svg$'),
    light_color varchar(7) NOT NULL CHECK (light_color ~ '^#[a-fA-F0-9]{6}$'),
    dark_color varchar(7) NOT NULL CHECK (dark_color ~ '^#[a-fA-F0-9]{6}$'),
    light_opacity numeric(5,4) NOT NULL CHECK (light_opacity >= 0 AND light_opacity <= 1),
    dark_opacity numeric(5,4) NOT NULL CHECK (dark_opacity >= 0 AND dark_opacity <= 1),
    default_size_px integer NOT NULL CHECK (default_size_px > 0),
    enabled boolean NOT NULL DEFAULT true,
    sort_order bigint NOT NULL DEFAULT 0,
    deleted boolean NOT NULL DEFAULT false,
    deleted_at timestamptz,
    create_time timestamp(0) NOT NULL DEFAULT NOW(),
    update_time timestamp(0) NOT NULL DEFAULT NOW()
);
CREATE INDEX gfn_background_pattern_catalog_idx ON public.gfn_background_pattern (sort_order,id) WHERE enabled AND NOT deleted;
COMMENT ON TABLE public.gfn_background_pattern IS 'Self-contained SVG pattern catalog; defaults are not copied into user overrides';
COMMENT ON COLUMN public.gfn_background_pattern.object_key IS 'Immutable provider-neutral managed object key';
COMMENT ON COLUMN public.gfn_site.icon IS 'Nullable managed object key nav/sites/{id}/icon/{hash}[.{ext}]; populated during maintenance cutover';

-- +goose Down
DROP TABLE public.gfn_background_pattern;
DROP TABLE public.gfn_home_hero_asset;
COMMENT ON COLUMN public.gfn_site.icon IS NULL;
