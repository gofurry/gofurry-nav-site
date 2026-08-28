-- V3-P0.2 Game historical fact and checkpoint tables.
-- +goose Up

CREATE TABLE public.gfg_game_player_hourly (
    tracking_period_id bigint NOT NULL
        REFERENCES public.gfg_game_tracking_periods(id) ON DELETE RESTRICT,
    game_id bigint NOT NULL,
    appid bigint NOT NULL,
    bucket_start timestamp with time zone NOT NULL,
    min_players bigint,
    max_players bigint,
    avg_players double precision,
    median_players double precision,
    expected_samples integer,
    attempted_samples integer NOT NULL,
    successful_samples integer NOT NULL,
    partial_samples integer NOT NULL,
    failed_samples integer NOT NULL,
    skipped_samples integer,
    missed_samples integer,
    canceled_samples integer,
    unattempted_samples integer,
    failure_kind_counts jsonb NOT NULL DEFAULT '{}'::jsonb,
    first_observed_at timestamp with time zone,
    last_observed_at timestamp with time zone,
    avg_observation_lag_ms bigint,
    max_observation_lag_ms bigint,
    quality_basis text NOT NULL,
    projection_version integer NOT NULL,
    finalized_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (tracking_period_id, bucket_start),
    CONSTRAINT gfg_game_player_hourly_bucket_check
        CHECK (bucket_start = date_trunc('hour', bucket_start)),
    CONSTRAINT gfg_game_player_hourly_nonnegative_check CHECK (
        expected_samples IS NULL OR expected_samples >= 0
    ),
    CONSTRAINT gfg_game_player_hourly_counts_check CHECK (
        attempted_samples >= 0
        AND successful_samples >= 0
        AND partial_samples >= 0
        AND failed_samples >= 0
        AND attempted_samples = successful_samples + partial_samples + failed_samples
    ),
    CONSTRAINT gfg_game_player_hourly_quality_check CHECK (
        (quality_basis = 'legacy_observed_only'
            AND expected_samples IS NULL
            AND skipped_samples IS NULL
            AND missed_samples IS NULL
            AND canceled_samples IS NULL
            AND unattempted_samples IS NULL)
        OR
        (quality_basis = 'acquisition_ledger'
            AND expected_samples IS NOT NULL
            AND skipped_samples >= 0
            AND missed_samples >= 0
            AND canceled_samples >= 0
            AND unattempted_samples >= 0
            AND expected_samples = attempted_samples + skipped_samples
                + missed_samples + canceled_samples + unattempted_samples)
    ),
    CONSTRAINT gfg_game_player_hourly_numeric_check CHECK (
        (successful_samples = 0
            AND min_players IS NULL AND max_players IS NULL
            AND avg_players IS NULL AND median_players IS NULL)
        OR
        (successful_samples > 0
            AND min_players IS NOT NULL AND max_players IS NOT NULL
            AND avg_players IS NOT NULL AND median_players IS NOT NULL)
    ),
    CONSTRAINT gfg_game_player_hourly_failure_json_check
        CHECK (jsonb_typeof(failure_kind_counts) = 'object'),
    CONSTRAINT gfg_game_player_hourly_projection_check
        CHECK (projection_version > 0)
);

CREATE INDEX idx_gfg_game_player_hourly_game_time
    ON public.gfg_game_player_hourly (game_id, bucket_start DESC);

CREATE TABLE public.gfg_game_player_daily (
    tracking_period_id bigint NOT NULL
        REFERENCES public.gfg_game_tracking_periods(id) ON DELETE RESTRICT,
    game_id bigint NOT NULL,
    appid bigint NOT NULL,
    fact_date date NOT NULL,
    min_players bigint,
    max_players bigint,
    avg_players double precision,
    median_players double precision,
    expected_samples integer,
    attempted_samples integer NOT NULL,
    successful_samples integer NOT NULL,
    partial_samples integer NOT NULL,
    failed_samples integer NOT NULL,
    skipped_samples integer,
    missed_samples integer,
    canceled_samples integer,
    unattempted_samples integer,
    failure_kind_counts jsonb NOT NULL DEFAULT '{}'::jsonb,
    first_observed_at timestamp with time zone,
    last_observed_at timestamp with time zone,
    avg_observation_lag_ms bigint,
    max_observation_lag_ms bigint,
    quality_basis text NOT NULL,
    projection_version integer NOT NULL,
    finalized_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (tracking_period_id, fact_date),
    CONSTRAINT gfg_game_player_daily_nonnegative_check
        CHECK (expected_samples IS NULL OR expected_samples >= 0),
    CONSTRAINT gfg_game_player_daily_counts_check CHECK (
        attempted_samples >= 0
        AND successful_samples >= 0
        AND partial_samples >= 0
        AND failed_samples >= 0
        AND attempted_samples = successful_samples + partial_samples + failed_samples
    ),
    CONSTRAINT gfg_game_player_daily_quality_check CHECK (
        (quality_basis = 'legacy_observed_only'
            AND expected_samples IS NULL
            AND skipped_samples IS NULL
            AND missed_samples IS NULL
            AND canceled_samples IS NULL
            AND unattempted_samples IS NULL)
        OR
        (quality_basis = 'acquisition_ledger'
            AND expected_samples IS NOT NULL
            AND skipped_samples >= 0
            AND missed_samples >= 0
            AND canceled_samples >= 0
            AND unattempted_samples >= 0
            AND expected_samples = attempted_samples + skipped_samples
                + missed_samples + canceled_samples + unattempted_samples)
    ),
    CONSTRAINT gfg_game_player_daily_numeric_check CHECK (
        (successful_samples = 0
            AND min_players IS NULL AND max_players IS NULL
            AND avg_players IS NULL AND median_players IS NULL)
        OR
        (successful_samples > 0
            AND min_players IS NOT NULL AND max_players IS NOT NULL
            AND avg_players IS NOT NULL AND median_players IS NOT NULL)
    ),
    CONSTRAINT gfg_game_player_daily_failure_json_check
        CHECK (jsonb_typeof(failure_kind_counts) = 'object'),
    CONSTRAINT gfg_game_player_daily_projection_check
        CHECK (projection_version > 0)
);

CREATE INDEX idx_gfg_game_player_daily_game_date
    ON public.gfg_game_player_daily (game_id, fact_date DESC);

CREATE TABLE public.gfg_game_price_daily (
    tracking_period_id bigint NOT NULL
        REFERENCES public.gfg_game_tracking_periods(id) ON DELETE RESTRICT,
    game_id bigint NOT NULL,
    appid bigint NOT NULL,
    region text NOT NULL,
    fact_date date NOT NULL,
    price_state text NOT NULL,
    currency text,
    initial_amount bigint,
    final_amount bigint,
    discount_percent integer,
    observed_at timestamp with time zone,
    materialization_source text NOT NULL,
    projection_version integer NOT NULL,
    finalized_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (tracking_period_id, region, fact_date),
    CONSTRAINT gfg_game_price_daily_region_check CHECK (btrim(region) <> ''),
    CONSTRAINT gfg_game_price_daily_state_check
        CHECK (price_state IN ('free', 'priced', 'unpriced', 'unknown')),
    CONSTRAINT gfg_game_price_daily_source_check
        CHECK (materialization_source IN ('bootstrap', 'observed', 'carried_forward')),
    CONSTRAINT gfg_game_price_daily_shape_check CHECK (
        (price_state = 'free'
            AND currency IS NULL AND initial_amount IS NULL
            AND final_amount IS NULL AND discount_percent IS NULL)
        OR
        (price_state = 'priced'
            AND btrim(currency) <> ''
            AND initial_amount >= 0 AND final_amount >= 0
            AND discount_percent BETWEEN 0 AND 100)
        OR
        (price_state IN ('unpriced', 'unknown')
            AND currency IS NULL AND initial_amount IS NULL
            AND final_amount IS NULL AND discount_percent IS NULL)
    ),
    CONSTRAINT gfg_game_price_daily_projection_check
        CHECK (projection_version > 0)
);

CREATE INDEX idx_gfg_game_price_daily_game_date
    ON public.gfg_game_price_daily (game_id, fact_date DESC, region);

CREATE TABLE public.gfg_game_daily (
    game_id bigint NOT NULL,
    fact_date date NOT NULL,
    tracking_period_id bigint NOT NULL
        REFERENCES public.gfg_game_tracking_periods(id) ON DELETE RESTRICT,
    appid bigint NOT NULL,
    snapshot_at timestamp with time zone NOT NULL,
    tracked_at_end boolean NOT NULL,
    name text NOT NULL,
    name_en text NOT NULL,
    view_count bigint NOT NULL,
    game_type text,
    is_free boolean,
    windows boolean,
    mac boolean,
    linux boolean,
    release_availability text,
    release_precision text,
    release_exact_date date,
    release_year integer,
    release_month integer,
    release_quarter integer,
    release_window_start date,
    release_window_end date,
    release_observed_at timestamp with time zone,
    first_available_precision text,
    first_available_exact_date date,
    first_available_year integer,
    first_available_month integer,
    first_available_quarter integer,
    first_available_window_start date,
    first_available_window_end date,
    first_available_source text,
    first_available_inferred boolean,
    language_codes text[],
    unknown_language_names text[],
    full_audio_language_codes text[],
    languages_observed_at timestamp with time zone,
    developers text[] NOT NULL,
    publishers text[] NOT NULL,
    primary_tag_id bigint,
    secondary_tag_id bigint,
    tag_ids bigint[] NOT NULL,
    details_observed_at timestamp with time zone,
    materialization_source text NOT NULL,
    projection_version integer NOT NULL,
    finalized_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (game_id, fact_date),
    CONSTRAINT gfg_game_daily_source_check
        CHECK (materialization_source IN ('bootstrap', 'observed', 'carried_forward')),
    CONSTRAINT gfg_game_daily_release_availability_check
        CHECK (release_availability IS NULL OR release_availability IN ('upcoming', 'available', 'unknown')),
    CONSTRAINT gfg_game_daily_release_precision_check
        CHECK (release_precision IS NULL OR release_precision IN ('day', 'month', 'quarter', 'year', 'tba', 'none', 'unknown')),
    CONSTRAINT gfg_game_daily_first_available_precision_check
        CHECK (first_available_precision IS NULL OR first_available_precision IN ('day', 'month', 'quarter', 'year')),
    CONSTRAINT gfg_game_daily_projection_check
        CHECK (projection_version > 0)
);

CREATE INDEX idx_gfg_game_daily_date ON public.gfg_game_daily (fact_date DESC, game_id);
CREATE INDEX idx_gfg_game_daily_tracking_period
    ON public.gfg_game_daily (tracking_period_id, fact_date DESC);

CREATE TABLE public.gfg_fact_rollup_checkpoints (
    pipeline_key text PRIMARY KEY,
    projection_version integer NOT NULL,
    source_start_date date,
    processed_through date,
    quality_cutover_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT gfg_fact_rollup_checkpoints_pipeline_check
        CHECK (pipeline_key IN ('game.player_facts', 'game.state_facts')),
    CONSTRAINT gfg_fact_rollup_checkpoints_projection_check
        CHECK (projection_version > 0),
    CONSTRAINT gfg_fact_rollup_checkpoints_range_check
        CHECK (source_start_date IS NULL OR processed_through IS NULL OR processed_through >= source_start_date)
);

INSERT INTO public.gfg_fact_rollup_checkpoints (
    pipeline_key, projection_version, source_start_date, quality_cutover_at
)
SELECT 'game.player_facts',
       1,
       LEAST(
           COALESCE(
               (SELECT min(collected_at AT TIME ZONE 'UTC')::date FROM public.gfg_game_player_counts),
               (transaction_timestamp() AT TIME ZONE 'UTC')::date
           ),
           (transaction_timestamp() AT TIME ZONE 'UTC')::date
       ),
       date_trunc('day', transaction_timestamp() AT TIME ZONE 'UTC')
           AT TIME ZONE 'UTC' + interval '1 day'
UNION ALL
SELECT 'game.state_facts',
       1,
       (transaction_timestamp() AT TIME ZONE 'UTC')::date,
       date_trunc('day', transaction_timestamp() AT TIME ZONE 'UTC')
           AT TIME ZONE 'UTC' + interval '1 day';

-- Shared write-through projector used by Game Collector and Admin. Keeping the
-- set-based projection under Goose prevents the two applications from drifting.
-- +goose StatementBegin
CREATE FUNCTION public.gfg_refresh_current_game_daily(
    target_game_id bigint,
    source_kind text
) RETURNS void
LANGUAGE plpgsql
AS $function$
BEGIN
    IF source_kind NOT IN ('bootstrap', 'observed', 'carried_forward') THEN
        RAISE EXCEPTION 'invalid Game Daily materialization source: %', source_kind;
    END IF;

    WITH current_game AS (
        SELECT g.*, p.id AS tracking_period_id
        FROM public.gfg_game g
        JOIN public.gfg_game_tracking_periods p
          ON p.game_id = g.id
         AND p.appid = g.appid
         AND p.tracking_basis = 'explicit'
         AND p.tracked_until IS NULL
        WHERE g.id = target_game_id
    ), languages AS (
        SELECT l.game_id,
               array_agg(l.language_code ORDER BY l.sort_order, l.id)
                   FILTER (WHERE l.language_code IS NOT NULL) AS language_codes,
               array_agg(l.steam_name ORDER BY l.sort_order, l.id)
                   FILTER (WHERE l.language_code IS NULL) AS unknown_language_names,
               array_agg(l.language_code ORDER BY l.sort_order, l.id)
                   FILTER (WHERE l.language_code IS NOT NULL AND l.full_audio_supported) AS full_audio_language_codes,
               max(l.observed_at) AS languages_observed_at
        FROM public.gfg_game_languages l
        WHERE l.game_id = target_game_id
        GROUP BY l.game_id
    ), tags AS (
        SELECT m.game_id, array_agg(DISTINCT m.tag_id ORDER BY m.tag_id) AS tag_ids
        FROM public.gfg_tag_map m
        WHERE m.game_id = target_game_id
        GROUP BY m.game_id
    ), snapshot AS (
        SELECT g.id AS game_id,
               (transaction_timestamp() AT TIME ZONE 'UTC')::date AS fact_date,
               g.tracking_period_id,
               g.appid,
               transaction_timestamp() AS snapshot_at,
               g.name::text AS name,
               g.name_en::text AS name_en,
               g.view_count,
               NULLIF(d.type, '') AS game_type,
               d.is_free,
               CASE WHEN d.game_id IS NULL THEN NULL ELSE (d.platforms ->> 'windows')::boolean END AS windows,
               CASE WHEN d.game_id IS NULL THEN NULL ELSE (d.platforms ->> 'mac')::boolean END AS mac,
               CASE WHEN d.game_id IS NULL THEN NULL ELSE (d.platforms ->> 'linux')::boolean END AS linux,
               r.availability AS release_availability,
               r.precision AS release_precision,
               r.exact_date AS release_exact_date,
               r.release_year,
               r.release_month,
               r.release_quarter,
               r.window_start AS release_window_start,
               r.window_end AS release_window_end,
               r.observed_at AS release_observed_at,
               f.precision AS first_available_precision,
               f.exact_date AS first_available_exact_date,
               f.release_year AS first_available_year,
               f.release_month AS first_available_month,
               f.release_quarter AS first_available_quarter,
               f.window_start AS first_available_window_start,
               f.window_end AS first_available_window_end,
               f.source AS first_available_source,
               f.inferred AS first_available_inferred,
               l.language_codes,
               l.unknown_language_names,
               l.full_audio_language_codes,
               l.languages_observed_at,
               CASE WHEN jsonb_typeof(g.developers) = 'array'
                    THEN ARRAY(SELECT jsonb_array_elements_text(g.developers))
                    ELSE ARRAY[]::text[] END AS developers,
               CASE WHEN jsonb_typeof(g.publishers) = 'array'
                    THEN ARRAY(SELECT jsonb_array_elements_text(g.publishers))
                    ELSE ARRAY[]::text[] END AS publishers,
               NULLIF(g.primary_tag, 0) AS primary_tag_id,
               NULLIF(g.secondary_tag, 0) AS secondary_tag_id,
               COALESCE(t.tag_ids, ARRAY[]::bigint[]) AS tag_ids,
               d.collected_at AS details_observed_at
        FROM current_game g
        LEFT JOIN public.gfg_game_details d ON d.game_id = g.id AND d.appid = g.appid
        LEFT JOIN public.gfg_game_release_state r ON r.game_id = g.id
        LEFT JOIN public.gfg_game_first_available f ON f.game_id = g.id
        LEFT JOIN languages l ON l.game_id = g.id
        LEFT JOIN tags t ON t.game_id = g.id
    )
    INSERT INTO public.gfg_game_daily (
        game_id, fact_date, tracking_period_id, appid, snapshot_at, tracked_at_end,
        name, name_en, view_count, game_type, is_free, windows, mac, linux,
        release_availability, release_precision, release_exact_date, release_year,
        release_month, release_quarter, release_window_start, release_window_end,
        release_observed_at, first_available_precision, first_available_exact_date,
        first_available_year, first_available_month, first_available_quarter,
        first_available_window_start, first_available_window_end,
        first_available_source, first_available_inferred, language_codes,
        unknown_language_names, full_audio_language_codes, languages_observed_at,
        developers, publishers, primary_tag_id, secondary_tag_id, tag_ids,
        details_observed_at, materialization_source, projection_version,
        finalized_at, created_at, updated_at
    )
    SELECT game_id, fact_date, tracking_period_id, appid, snapshot_at, true,
           name, name_en, view_count, game_type, is_free, windows, mac, linux,
           release_availability, release_precision, release_exact_date, release_year,
           release_month, release_quarter, release_window_start, release_window_end,
           release_observed_at, first_available_precision, first_available_exact_date,
           first_available_year, first_available_month, first_available_quarter,
           first_available_window_start, first_available_window_end,
           first_available_source, first_available_inferred, language_codes,
           unknown_language_names, full_audio_language_codes, languages_observed_at,
           developers, publishers, primary_tag_id, secondary_tag_id, tag_ids,
           details_observed_at, source_kind, 1,
           NULL, transaction_timestamp(), transaction_timestamp()
    FROM snapshot
    ON CONFLICT (game_id, fact_date) DO UPDATE
    SET tracking_period_id = EXCLUDED.tracking_period_id,
        appid = EXCLUDED.appid,
        snapshot_at = EXCLUDED.snapshot_at,
        tracked_at_end = EXCLUDED.tracked_at_end,
        name = EXCLUDED.name,
        name_en = EXCLUDED.name_en,
        view_count = EXCLUDED.view_count,
        game_type = EXCLUDED.game_type,
        is_free = EXCLUDED.is_free,
        windows = EXCLUDED.windows,
        mac = EXCLUDED.mac,
        linux = EXCLUDED.linux,
        release_availability = EXCLUDED.release_availability,
        release_precision = EXCLUDED.release_precision,
        release_exact_date = EXCLUDED.release_exact_date,
        release_year = EXCLUDED.release_year,
        release_month = EXCLUDED.release_month,
        release_quarter = EXCLUDED.release_quarter,
        release_window_start = EXCLUDED.release_window_start,
        release_window_end = EXCLUDED.release_window_end,
        release_observed_at = EXCLUDED.release_observed_at,
        first_available_precision = EXCLUDED.first_available_precision,
        first_available_exact_date = EXCLUDED.first_available_exact_date,
        first_available_year = EXCLUDED.first_available_year,
        first_available_month = EXCLUDED.first_available_month,
        first_available_quarter = EXCLUDED.first_available_quarter,
        first_available_window_start = EXCLUDED.first_available_window_start,
        first_available_window_end = EXCLUDED.first_available_window_end,
        first_available_source = EXCLUDED.first_available_source,
        first_available_inferred = EXCLUDED.first_available_inferred,
        language_codes = EXCLUDED.language_codes,
        unknown_language_names = EXCLUDED.unknown_language_names,
        full_audio_language_codes = EXCLUDED.full_audio_language_codes,
        languages_observed_at = EXCLUDED.languages_observed_at,
        developers = EXCLUDED.developers,
        publishers = EXCLUDED.publishers,
        primary_tag_id = EXCLUDED.primary_tag_id,
        secondary_tag_id = EXCLUDED.secondary_tag_id,
        tag_ids = EXCLUDED.tag_ids,
        details_observed_at = EXCLUDED.details_observed_at,
        materialization_source = EXCLUDED.materialization_source,
        projection_version = EXCLUDED.projection_version,
        updated_at = transaction_timestamp()
    WHERE public.gfg_game_daily.finalized_at IS NULL;
END;
$function$;
-- +goose StatementEnd

SELECT public.gfg_refresh_current_game_daily(id, 'bootstrap')
FROM public.gfg_game;

COMMENT ON TABLE public.gfg_game_player_hourly IS 'UTC Game player facts; scheduled acquisition quality is separate from numeric values.';
COMMENT ON TABLE public.gfg_game_player_daily IS 'UTC Game player facts aggregated directly from successful daily raw samples.';
COMMENT ON TABLE public.gfg_game_price_daily IS 'Historical regional price state with observation provenance and carry-forward source.';
COMMENT ON TABLE public.gfg_game_daily IS 'Compact historical Game entity and canonical domain snapshot.';
COMMENT ON TABLE public.gfg_fact_rollup_checkpoints IS 'Ordered singleton checkpoints for Game fact pipelines.';

-- +goose Down

DROP FUNCTION IF EXISTS public.gfg_refresh_current_game_daily(bigint, text);
DROP TABLE public.gfg_fact_rollup_checkpoints;
DROP TABLE public.gfg_game_daily;
DROP TABLE public.gfg_game_price_daily;
DROP TABLE public.gfg_game_player_daily;
DROP TABLE public.gfg_game_player_hourly;
