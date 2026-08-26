-- +goose Up

CREATE TABLE public.gfg_game_release_state (
    game_id bigint PRIMARY KEY
        REFERENCES public.gfg_game(id) ON DELETE CASCADE,
    availability text NOT NULL,
    precision text NOT NULL,
    exact_date date,
    release_year integer,
    release_month integer,
    release_quarter integer,
    window_start date,
    window_end date,
    raw_text text NOT NULL DEFAULT '',
    source text NOT NULL,
    source_region text NOT NULL,
    source_locale text NOT NULL,
    normalizer_version text NOT NULL,
    observed_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT gfg_game_release_state_availability_check
        CHECK (availability IN ('upcoming', 'available', 'unknown')),
    CONSTRAINT gfg_game_release_state_precision_check
        CHECK (precision IN ('day', 'month', 'quarter', 'year', 'tba', 'none', 'unknown')),
    CONSTRAINT gfg_game_release_state_shape_check CHECK (
        (precision = 'day'
            AND exact_date IS NOT NULL
            AND release_year IS NOT NULL
            AND release_month BETWEEN 1 AND 12
            AND release_quarter IS NULL
            AND window_start = exact_date
            AND window_end = exact_date
            AND extract(year FROM exact_date)::integer = release_year
            AND extract(month FROM exact_date)::integer = release_month)
        OR
        (precision = 'month'
            AND exact_date IS NULL
            AND release_year IS NOT NULL
            AND release_month BETWEEN 1 AND 12
            AND release_quarter IS NULL
            AND window_start IS NOT NULL
            AND window_end IS NOT NULL)
        OR
        (precision = 'quarter'
            AND exact_date IS NULL
            AND release_year IS NOT NULL
            AND release_month IS NULL
            AND release_quarter BETWEEN 1 AND 4
            AND window_start IS NOT NULL
            AND window_end IS NOT NULL)
        OR
        (precision = 'year'
            AND exact_date IS NULL
            AND release_year IS NOT NULL
            AND release_month IS NULL
            AND release_quarter IS NULL
            AND window_start IS NOT NULL
            AND window_end IS NOT NULL)
        OR
        (precision IN ('tba', 'none', 'unknown')
            AND exact_date IS NULL
            AND release_year IS NULL
            AND release_month IS NULL
            AND release_quarter IS NULL
            AND window_start IS NULL
            AND window_end IS NULL)
    ),
    CONSTRAINT gfg_game_release_state_window_check
        CHECK (window_start IS NULL OR window_end IS NULL OR window_start <= window_end)
);

CREATE INDEX idx_gfg_game_release_state_availability_window
    ON public.gfg_game_release_state (availability, window_start);

CREATE TABLE public.gfg_game_first_available (
    game_id bigint PRIMARY KEY
        REFERENCES public.gfg_game(id) ON DELETE CASCADE,
    precision text NOT NULL,
    exact_date date,
    release_year integer NOT NULL,
    release_month integer,
    release_quarter integer,
    window_start date NOT NULL,
    window_end date NOT NULL,
    source text NOT NULL,
    inferred boolean NOT NULL,
    source_raw text NOT NULL DEFAULT '',
    source_observed_at timestamp with time zone,
    normalizer_version text NOT NULL,
    established_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT gfg_game_first_available_precision_check
        CHECK (precision IN ('day', 'month', 'quarter', 'year')),
    CONSTRAINT gfg_game_first_available_source_check
        CHECK (source IN ('legacy_manual', 'observed_transition', 'steam_backfill')),
    CONSTRAINT gfg_game_first_available_source_inferred_check CHECK (
        (source = 'steam_backfill' AND inferred IS TRUE)
        OR (source IN ('legacy_manual', 'observed_transition') AND inferred IS FALSE)
    ),
    CONSTRAINT gfg_game_first_available_shape_check CHECK (
        (precision = 'day'
            AND exact_date IS NOT NULL
            AND release_month BETWEEN 1 AND 12
            AND release_quarter IS NULL
            AND window_start = exact_date
            AND window_end = exact_date
            AND extract(year FROM exact_date)::integer = release_year
            AND extract(month FROM exact_date)::integer = release_month)
        OR
        (precision = 'month'
            AND exact_date IS NULL
            AND release_month BETWEEN 1 AND 12
            AND release_quarter IS NULL)
        OR
        (precision = 'quarter'
            AND exact_date IS NULL
            AND release_month IS NULL
            AND release_quarter BETWEEN 1 AND 4)
        OR
        (precision = 'year'
            AND exact_date IS NULL
            AND release_month IS NULL
            AND release_quarter IS NULL)
    ),
    CONSTRAINT gfg_game_first_available_window_check
        CHECK (window_start <= window_end)
);

CREATE INDEX idx_gfg_game_first_available_latest
    ON public.gfg_game_first_available (window_end DESC, game_id DESC);
CREATE INDEX idx_gfg_game_first_available_window_start
    ON public.gfg_game_first_available (window_start);

CREATE TABLE public.gfg_game_release_history (
    id bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    game_id bigint NOT NULL
        REFERENCES public.gfg_game(id) ON DELETE CASCADE,
    availability text NOT NULL,
    precision text NOT NULL,
    exact_date date,
    release_year integer,
    release_month integer,
    release_quarter integer,
    window_start date,
    window_end date,
    raw_text text NOT NULL DEFAULT '',
    source text NOT NULL,
    source_region text NOT NULL,
    source_locale text NOT NULL,
    normalizer_version text NOT NULL,
    observed_at timestamp with time zone NOT NULL,
    recorded_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT gfg_game_release_history_availability_check
        CHECK (availability IN ('upcoming', 'available', 'unknown')),
    CONSTRAINT gfg_game_release_history_precision_check
        CHECK (precision IN ('day', 'month', 'quarter', 'year', 'tba', 'none', 'unknown')),
    CONSTRAINT gfg_game_release_history_shape_check CHECK (
        (precision = 'day'
            AND exact_date IS NOT NULL
            AND release_year IS NOT NULL
            AND release_month BETWEEN 1 AND 12
            AND release_quarter IS NULL
            AND window_start = exact_date
            AND window_end = exact_date
            AND extract(year FROM exact_date)::integer = release_year
            AND extract(month FROM exact_date)::integer = release_month)
        OR
        (precision = 'month'
            AND exact_date IS NULL
            AND release_year IS NOT NULL
            AND release_month BETWEEN 1 AND 12
            AND release_quarter IS NULL
            AND window_start IS NOT NULL
            AND window_end IS NOT NULL)
        OR
        (precision = 'quarter'
            AND exact_date IS NULL
            AND release_year IS NOT NULL
            AND release_month IS NULL
            AND release_quarter BETWEEN 1 AND 4
            AND window_start IS NOT NULL
            AND window_end IS NOT NULL)
        OR
        (precision = 'year'
            AND exact_date IS NULL
            AND release_year IS NOT NULL
            AND release_month IS NULL
            AND release_quarter IS NULL
            AND window_start IS NOT NULL
            AND window_end IS NOT NULL)
        OR
        (precision IN ('tba', 'none', 'unknown')
            AND exact_date IS NULL
            AND release_year IS NULL
            AND release_month IS NULL
            AND release_quarter IS NULL
            AND window_start IS NULL
            AND window_end IS NULL)
    ),
    CONSTRAINT gfg_game_release_history_window_check
        CHECK (window_start IS NULL OR window_end IS NULL OR window_start <= window_end)
);

CREATE INDEX idx_gfg_game_release_history_game_observed
    ON public.gfg_game_release_history (game_id, observed_at DESC, id DESC);

CREATE TABLE public.gfg_game_languages (
    id bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    game_id bigint NOT NULL
        REFERENCES public.gfg_game(id) ON DELETE CASCADE,
    language_code text,
    steam_name text NOT NULL,
    steam_api_code text,
    steam_web_code text,
    tier text NOT NULL,
    interface_supported boolean,
    subtitles_supported boolean,
    full_audio_supported boolean,
    sort_order integer NOT NULL,
    source text NOT NULL,
    source_region text NOT NULL,
    source_locale text NOT NULL,
    normalizer_version text NOT NULL,
    observed_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT gfg_game_languages_tier_check
        CHECK (tier IN ('platform', 'game_only', 'unknown'))
);

CREATE UNIQUE INDEX uq_gfg_game_languages_known
    ON public.gfg_game_languages (game_id, language_code)
    WHERE language_code IS NOT NULL;
CREATE UNIQUE INDEX uq_gfg_game_languages_unknown
    ON public.gfg_game_languages (game_id, lower(steam_name))
    WHERE language_code IS NULL;
CREATE INDEX idx_gfg_game_languages_order
    ON public.gfg_game_languages (game_id, sort_order, id);
CREATE INDEX idx_gfg_game_languages_code
    ON public.gfg_game_languages (language_code, game_id)
    WHERE language_code IS NOT NULL;

ALTER TABLE public.gfg_game
    ALTER COLUMN release_date SET DEFAULT '';
COMMENT ON COLUMN public.gfg_game.release_date IS
    'Deprecated legacy manual release string; retained temporarily for migration audit and rollback only.';

COMMENT ON TABLE public.gfg_game_release_state IS
    'Current canonical Steam release state normalized from the US/English Storefront response.';
COMMENT ON TABLE public.gfg_game_first_available IS
    'Write-once canonical date or range when a game first became formally purchasable or playable.';
COMMENT ON TABLE public.gfg_game_release_history IS
    'Semantic snapshots of canonical release-state changes; raw-text-only changes do not append rows.';
COMMENT ON TABLE public.gfg_game_languages IS
    'Canonical game language support normalized from the US/English Storefront response.';
