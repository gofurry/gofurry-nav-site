-- V3-P0.4 Game Change Intelligence Foundation.
-- +goose Up

CREATE TABLE public.gfg_change_registry (
    detector_key text NOT NULL,
    detector_version integer NOT NULL,
    source_kind text NOT NULL,
    source_contracts text[] NOT NULL,
    detection_policy text NOT NULL,
    watermark_policy text NOT NULL,
    event_codes text[] NOT NULL,
    processing_grain text NOT NULL,
    status text NOT NULL,
    description text NOT NULL DEFAULT '',
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    retired_at timestamp with time zone,
    PRIMARY KEY (detector_key, detector_version),
    CONSTRAINT gfg_change_registry_key_check CHECK (btrim(detector_key) <> ''),
    CONSTRAINT gfg_change_registry_version_check CHECK (detector_version > 0),
    CONSTRAINT gfg_change_registry_source_kind_check CHECK (source_kind IN ('metric', 'fact', 'domain_history', 'effective_period')),
    CONSTRAINT gfg_change_registry_source_shape_check CHECK (array_ndims(source_contracts) = 1 AND cardinality(source_contracts) > 0),
    CONSTRAINT gfg_change_registry_policy_check CHECK (btrim(detection_policy) <> '' AND btrim(watermark_policy) <> ''),
    CONSTRAINT gfg_change_registry_codes_shape_check CHECK (array_ndims(event_codes) = 1 AND cardinality(event_codes) > 0),
    CONSTRAINT gfg_change_registry_grain_check CHECK (processing_grain = 'day'),
    CONSTRAINT gfg_change_registry_status_check CHECK (status IN ('active', 'retired')),
    CONSTRAINT gfg_change_registry_status_shape_check CHECK (
        (status = 'active' AND retired_at IS NULL) OR (status = 'retired' AND retired_at IS NOT NULL)
    )
);

CREATE UNIQUE INDEX uq_gfg_change_registry_active_key
    ON public.gfg_change_registry(detector_key) WHERE status = 'active';

CREATE TABLE public.gfg_change_events (
    event_key text PRIMARY KEY,
    detector_key text NOT NULL,
    detector_version integer NOT NULL,
    game_id bigint NOT NULL,
    projection_date date NOT NULL,
    event_at timestamp with time zone,
    time_basis text NOT NULL,
    event_code text NOT NULL,
    scope_kind text NOT NULL,
    scope_key text NOT NULL,
    old_value jsonb NOT NULL DEFAULT '{}'::jsonb,
    new_value jsonb NOT NULL DEFAULT '{}'::jsonb,
    source_event_key text NOT NULL,
    source_before_key text NOT NULL,
    source_after_key text NOT NULL,
    source_before_at timestamp with time zone,
    source_after_at timestamp with time zone,
    source_versions jsonb NOT NULL DEFAULT '{}'::jsonb,
    materialized_at timestamp with time zone NOT NULL,
    CONSTRAINT fk_gfg_change_event_registry FOREIGN KEY (detector_key, detector_version)
        REFERENCES public.gfg_change_registry(detector_key, detector_version) ON DELETE RESTRICT,
    CONSTRAINT uq_gfg_change_event_source UNIQUE (detector_key, detector_version, source_event_key),
    CONSTRAINT gfg_change_event_text_shape_check CHECK (
        btrim(event_key) <> '' AND btrim(event_code) <> '' AND btrim(scope_kind) <> ''
        AND btrim(scope_key) <> '' AND btrim(source_event_key) <> ''
        AND btrim(source_before_key) <> '' AND btrim(source_after_key) <> ''
    ),
    CONSTRAINT gfg_change_event_time_basis_check CHECK (time_basis IN ('effective', 'observed', 'day')),
    CONSTRAINT gfg_change_event_time_shape_check CHECK (
        (time_basis IN ('effective', 'observed') AND event_at IS NOT NULL)
        OR (time_basis = 'day' AND event_at IS NULL)
    ),
    CONSTRAINT gfg_change_event_json_shape_check CHECK (
        jsonb_typeof(old_value) = 'object' AND jsonb_typeof(new_value) = 'object'
        AND jsonb_typeof(source_versions) = 'object'
    )
);

CREATE INDEX idx_gfg_change_events_detector_date
    ON public.gfg_change_events(detector_key, detector_version, projection_date DESC);
CREATE INDEX idx_gfg_change_events_game_date
    ON public.gfg_change_events(game_id, projection_date DESC);
CREATE INDEX idx_gfg_change_events_code_date
    ON public.gfg_change_events(event_code, projection_date DESC);
CREATE INDEX idx_gfg_change_events_scope_date
    ON public.gfg_change_events(scope_kind, scope_key, projection_date DESC);
CREATE INDEX idx_gfg_change_events_event_at
    ON public.gfg_change_events(event_at DESC) WHERE event_at IS NOT NULL;

CREATE TABLE public.gfg_change_checkpoints (
    detector_key text NOT NULL,
    detector_version integer NOT NULL,
    source_start_date date,
    processed_through date,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (detector_key, detector_version),
    CONSTRAINT fk_gfg_change_checkpoint_registry FOREIGN KEY (detector_key, detector_version)
        REFERENCES public.gfg_change_registry(detector_key, detector_version) ON DELETE RESTRICT,
    CONSTRAINT gfg_change_checkpoint_range_check CHECK (
        source_start_date IS NULL OR processed_through IS NULL OR processed_through >= source_start_date
    )
);

INSERT INTO public.gfg_change_registry (
    detector_key, detector_version, source_kind, source_contracts,
    detection_policy, watermark_policy, event_codes, processing_grain, status, description
) VALUES
    ('free_game_transition', 1, 'metric', ARRAY['free_game_share/1', 'gfg_game_daily'],
     'metric_semantic_transition_v1', 'metric_checkpoint_v1',
     ARRAY['game_became_free', 'game_became_paid'], 'day', 'active',
     'Semantic free/paid transitions within one Game tracking identity.'),
    ('windows_support_transition', 1, 'metric', ARRAY['windows_support/1', 'gfg_game_daily'],
     'metric_semantic_transition_v1', 'metric_checkpoint_v1',
     ARRAY['windows_support_added', 'windows_support_removed'], 'day', 'active',
     'Semantic Windows support transitions within one Game tracking identity.'),
    ('linux_support_transition', 1, 'metric', ARRAY['linux_support/1', 'gfg_game_daily'],
     'metric_semantic_transition_v1', 'metric_checkpoint_v1',
     ARRAY['linux_support_added', 'linux_support_removed'], 'day', 'active',
     'Semantic Linux support transitions within one Game tracking identity.'),
    ('game_release_transition', 1, 'domain_history', ARRAY['gfg_game_release_history', 'gfg_game_tracking_periods'],
     'release_adjacent_history_v1', 'closed_day_v1',
     ARRAY['game_became_available', 'game_availability_withdrawn', 'game_release_plan_changed'], 'day', 'active',
     'Canonical release transitions reliably attributable to one Game tracking period.'),
    ('game_price_transition', 1, 'fact', ARRAY['gfg_game_price_daily'],
     'price_semantic_memory_v1', 'fact_checkpoint_v1',
     ARRAY['game_price_state_changed', 'game_price_currency_changed', 'game_price_decreased', 'game_price_increased',
           'game_discount_started', 'game_discount_ended', 'game_discount_changed'], 'day', 'active',
     'Regional price and discount transitions within one Game tracking identity.');

INSERT INTO public.gfg_change_checkpoints(detector_key, detector_version, source_start_date)
SELECT registry.detector_key,
       registry.detector_version,
       CASE registry.detector_key
           WHEN 'free_game_transition' THEN (SELECT source_start_date FROM public.gfg_metric_checkpoints WHERE metric_key = 'free_game_share' AND metric_version = 1)
           WHEN 'windows_support_transition' THEN (SELECT source_start_date FROM public.gfg_metric_checkpoints WHERE metric_key = 'windows_support' AND metric_version = 1)
           WHEN 'linux_support_transition' THEN (SELECT source_start_date FROM public.gfg_metric_checkpoints WHERE metric_key = 'linux_support' AND metric_version = 1)
           WHEN 'game_price_transition' THEN (SELECT source_start_date FROM public.gfg_fact_rollup_checkpoints WHERE pipeline_key = 'game.state_facts')
           WHEN 'game_release_transition' THEN (
               SELECT COALESCE(
                   min((history.observed_at AT TIME ZONE 'UTC')::date),
                   (transaction_timestamp() AT TIME ZONE 'UTC')::date
               )
               FROM public.gfg_game_release_history history
               WHERE (SELECT count(*) FROM public.gfg_game_tracking_periods period
                      WHERE period.game_id = history.game_id
                        AND history.observed_at >= period.tracked_from
                        AND (period.tracked_until IS NULL OR history.observed_at < period.tracked_until)) = 1
           )
       END
FROM public.gfg_change_registry registry;

-- +goose StatementBegin
CREATE FUNCTION public.gfg_project_change_day(
    target_detector_key text,
    target_detector_version integer,
    target_date date
) RETURNS bigint
LANGUAGE plpgsql
AS $function$
DECLARE
    inserted_rows bigint := 0;
    source_metric_key text;
BEGIN
    IF target_detector_version <> 1 OR target_detector_key NOT IN (
        'free_game_transition', 'windows_support_transition', 'linux_support_transition',
        'game_release_transition', 'game_price_transition'
    ) THEN
        RAISE EXCEPTION 'Game change detector is not compiled for %/%', target_detector_key, target_detector_version;
    END IF;

    DELETE FROM public.gfg_change_events
    WHERE detector_key = target_detector_key
      AND detector_version = target_detector_version
      AND projection_date = target_date;

    IF target_detector_key IN ('free_game_transition', 'windows_support_transition', 'linux_support_transition') THEN
        source_metric_key := CASE target_detector_key
            WHEN 'free_game_transition' THEN 'free_game_share'
            WHEN 'windows_support_transition' THEN 'windows_support'
            WHEN 'linux_support_transition' THEN 'linux_support'
        END;

        INSERT INTO public.gfg_change_events (
            event_key, detector_key, detector_version, game_id, projection_date,
            event_at, time_basis, event_code, scope_kind, scope_key,
            old_value, new_value, source_event_key, source_before_key, source_after_key,
            source_before_at, source_after_at, source_versions, materialized_at
        )
        SELECT target_detector_key || '/1/' || source_event_key,
               target_detector_key, 1, current_metric.game_id, target_date,
               current_metric.source_observed_at,
               CASE WHEN current_metric.source_observed_at IS NULL THEN 'day' ELSE 'observed' END,
               CASE target_detector_key
                   WHEN 'free_game_transition' THEN CASE WHEN current_metric.state = 'positive' THEN 'game_became_free' ELSE 'game_became_paid' END
                   WHEN 'windows_support_transition' THEN CASE WHEN current_metric.state = 'positive' THEN 'windows_support_added' ELSE 'windows_support_removed' END
                   WHEN 'linux_support_transition' THEN CASE WHEN current_metric.state = 'positive' THEN 'linux_support_added' ELSE 'linux_support_removed' END
               END,
               'global', 'all',
               jsonb_build_object('state', previous_metric.state, 'reason_code', previous_metric.reason_code),
               jsonb_build_object('state', current_metric.state, 'reason_code', current_metric.reason_code),
               source_event_key,
               source_metric_key || '/1/' || previous_metric.fact_date || '/' || current_metric.game_id,
               source_metric_key || '/1/' || current_metric.fact_date || '/' || current_metric.game_id,
               previous_metric.source_observed_at, current_metric.source_observed_at,
               jsonb_build_object(
                   'metric_key', source_metric_key, 'metric_version', 1,
                   'before_fact_projection_versions', previous_metric.source_projection_versions,
                   'after_fact_projection_versions', current_metric.source_projection_versions,
                   'tracking_period_id', current_fact.tracking_period_id
               ),
               transaction_timestamp()
        FROM public.gfg_metric_entity_daily current_metric
        JOIN public.gfg_game_daily current_fact
          ON current_fact.game_id = current_metric.game_id
         AND current_fact.fact_date = current_metric.fact_date
         AND current_fact.finalized_at IS NOT NULL
        JOIN LATERAL (
            SELECT previous_metric.*
            FROM public.gfg_metric_entity_daily previous_metric
            JOIN public.gfg_game_daily previous_fact
              ON previous_fact.game_id = previous_metric.game_id
             AND previous_fact.fact_date = previous_metric.fact_date
             AND previous_fact.finalized_at IS NOT NULL
            WHERE previous_metric.metric_key = current_metric.metric_key
              AND previous_metric.metric_version = current_metric.metric_version
              AND previous_metric.game_id = current_metric.game_id
              AND previous_metric.fact_date < current_metric.fact_date
              AND previous_metric.state IN ('positive', 'negative')
              AND previous_fact.tracking_period_id = current_fact.tracking_period_id
            ORDER BY previous_metric.fact_date DESC
            LIMIT 1
        ) previous_metric ON true
        CROSS JOIN LATERAL (
            SELECT source_metric_key || '/1/' || current_fact.tracking_period_id || '/' || current_metric.game_id || '/'
                   || previous_metric.fact_date || '/' || current_metric.fact_date AS source_event_key
        ) event_identity
        WHERE current_metric.metric_key = source_metric_key
          AND current_metric.metric_version = 1
          AND current_metric.fact_date = target_date
          AND current_metric.state IN ('positive', 'negative')
          AND current_metric.state <> previous_metric.state;

    ELSIF target_detector_key = 'game_release_transition' THEN
        WITH reliable AS (
            SELECT history.*,
                   (SELECT min(period.id)
                    FROM public.gfg_game_tracking_periods period
                    WHERE period.game_id = history.game_id
                      AND history.observed_at >= period.tracked_from
                      AND (period.tracked_until IS NULL OR history.observed_at < period.tracked_until)) AS tracking_period_id,
                   (SELECT count(*)
                    FROM public.gfg_game_tracking_periods period
                    WHERE period.game_id = history.game_id
                      AND history.observed_at >= period.tracked_from
                      AND (period.tracked_until IS NULL OR history.observed_at < period.tracked_until)) AS period_count
            FROM public.gfg_game_release_history history
        ), chained AS (
            SELECT reliable.*,
                   lag(id) OVER (PARTITION BY tracking_period_id ORDER BY observed_at, id) AS before_id,
                   lag(availability) OVER (PARTITION BY tracking_period_id ORDER BY observed_at, id) AS before_availability,
                   lag(precision) OVER (PARTITION BY tracking_period_id ORDER BY observed_at, id) AS before_precision,
                   lag(exact_date) OVER (PARTITION BY tracking_period_id ORDER BY observed_at, id) AS before_exact_date,
                   lag(release_year) OVER (PARTITION BY tracking_period_id ORDER BY observed_at, id) AS before_release_year,
                   lag(release_month) OVER (PARTITION BY tracking_period_id ORDER BY observed_at, id) AS before_release_month,
                   lag(release_quarter) OVER (PARTITION BY tracking_period_id ORDER BY observed_at, id) AS before_release_quarter,
                   lag(window_start) OVER (PARTITION BY tracking_period_id ORDER BY observed_at, id) AS before_window_start,
                   lag(window_end) OVER (PARTITION BY tracking_period_id ORDER BY observed_at, id) AS before_window_end,
                   lag(observed_at) OVER (PARTITION BY tracking_period_id ORDER BY observed_at, id) AS before_observed_at,
                   lag(normalizer_version) OVER (PARTITION BY tracking_period_id ORDER BY observed_at, id) AS before_normalizer_version
            FROM reliable
            WHERE period_count = 1
        ), events AS (
            SELECT chained.*,
                   CASE
                       WHEN before_availability <> 'available' AND availability = 'available' THEN 'game_became_available'
                       WHEN before_availability = 'available' AND availability <> 'available' THEN 'game_availability_withdrawn'
                       WHEN ROW(before_availability, before_precision, before_exact_date, before_release_year, before_release_month,
                                before_release_quarter, before_window_start, before_window_end)
                            IS DISTINCT FROM
                            ROW(availability, precision, exact_date, release_year, release_month,
                                release_quarter, window_start, window_end) THEN 'game_release_plan_changed'
                   END AS event_code
            FROM chained
            WHERE before_id IS NOT NULL
              AND (observed_at AT TIME ZONE 'UTC')::date = target_date
        )
        INSERT INTO public.gfg_change_events (
            event_key, detector_key, detector_version, game_id, projection_date,
            event_at, time_basis, event_code, scope_kind, scope_key,
            old_value, new_value, source_event_key, source_before_key, source_after_key,
            source_before_at, source_after_at, source_versions, materialized_at
        )
        SELECT 'game_release_transition/1/history/' || id,
               target_detector_key, 1, game_id, target_date,
               observed_at, 'observed', event_code, 'global', 'all',
               jsonb_build_object('availability', before_availability, 'precision', before_precision,
                   'exact_date', before_exact_date, 'release_year', before_release_year,
                   'release_month', before_release_month, 'release_quarter', before_release_quarter,
                   'window_start', before_window_start, 'window_end', before_window_end),
               jsonb_build_object('availability', availability, 'precision', precision,
                   'exact_date', exact_date, 'release_year', release_year,
                   'release_month', release_month, 'release_quarter', release_quarter,
                   'window_start', window_start, 'window_end', window_end),
               'history/' || id, 'history/' || before_id, 'history/' || id,
               before_observed_at, observed_at,
               jsonb_build_object('before_normalizer_version', before_normalizer_version,
                                  'after_normalizer_version', normalizer_version,
                                  'tracking_period_id', tracking_period_id),
               transaction_timestamp()
        FROM events
        WHERE event_code IS NOT NULL;

    ELSE
        WITH current_price AS (
            SELECT price.*
            FROM public.gfg_game_price_daily price
            WHERE price.fact_date = target_date
              AND price.finalized_at IS NOT NULL
              AND price.price_state IN ('free', 'priced', 'unpriced')
        ), paired AS (
            SELECT current_price.*,
                   previous.fact_date AS before_date,
                   previous.price_state AS before_price_state,
                   previous.currency AS before_currency,
                   previous.initial_amount AS before_initial_amount,
                   previous.final_amount AS before_final_amount,
                   previous.discount_percent AS before_discount_percent,
                   previous.observed_at AS before_observed_at,
                   previous.projection_version AS before_projection_version
            FROM current_price
            JOIN LATERAL (
                SELECT candidate.*
                FROM public.gfg_game_price_daily candidate
                WHERE candidate.tracking_period_id = current_price.tracking_period_id
                  AND candidate.region = current_price.region
                  AND candidate.fact_date < current_price.fact_date
                  AND candidate.finalized_at IS NOT NULL
                  AND candidate.price_state IN ('free', 'priced', 'unpriced')
                ORDER BY candidate.fact_date DESC
                LIMIT 1
            ) previous ON true
        ), event_rows AS (
            SELECT paired.*,
                   main.event_code,
                   main.event_suffix
            FROM paired
            CROSS JOIN LATERAL (
                SELECT CASE
                    WHEN before_price_state <> price_state THEN 'game_price_state_changed'
                    WHEN before_currency IS DISTINCT FROM currency THEN 'game_price_currency_changed'
                    WHEN before_final_amount > final_amount THEN 'game_price_decreased'
                    WHEN before_final_amount < final_amount THEN 'game_price_increased'
                END AS event_code,
                'main'::text AS event_suffix
            ) main
            WHERE main.event_code IS NOT NULL
            UNION ALL
            SELECT paired.*,
                   CASE
                       WHEN COALESCE(before_discount_percent, 0) = 0 AND COALESCE(discount_percent, 0) > 0 THEN 'game_discount_started'
                       WHEN COALESCE(before_discount_percent, 0) > 0 AND COALESCE(discount_percent, 0) = 0 THEN 'game_discount_ended'
                       WHEN before_discount_percent IS DISTINCT FROM discount_percent THEN 'game_discount_changed'
                   END,
                   'discount'::text
            FROM paired
            WHERE before_price_state = 'priced' AND price_state = 'priced'
              AND before_currency IS NOT DISTINCT FROM currency
              AND CASE
                  WHEN COALESCE(before_discount_percent, 0) = 0 AND COALESCE(discount_percent, 0) > 0 THEN true
                  WHEN COALESCE(before_discount_percent, 0) > 0 AND COALESCE(discount_percent, 0) = 0 THEN true
                  ELSE before_discount_percent IS DISTINCT FROM discount_percent
              END
        )
        INSERT INTO public.gfg_change_events (
            event_key, detector_key, detector_version, game_id, projection_date,
            event_at, time_basis, event_code, scope_kind, scope_key,
            old_value, new_value, source_event_key, source_before_key, source_after_key,
            source_before_at, source_after_at, source_versions, materialized_at
        )
        SELECT 'game_price_transition/1/' || tracking_period_id || '/' || region || '/' || fact_date || '/' || event_suffix,
               target_detector_key, 1, game_id, target_date,
               observed_at, CASE WHEN observed_at IS NULL THEN 'day' ELSE 'observed' END,
               event_code, 'region', region,
               jsonb_build_object('price_state', before_price_state, 'currency', before_currency,
                                  'initial_amount', before_initial_amount, 'final_amount', before_final_amount,
                                  'discount_percent', before_discount_percent),
               jsonb_build_object('price_state', price_state, 'currency', currency,
                                  'initial_amount', initial_amount, 'final_amount', final_amount,
                                  'discount_percent', discount_percent),
               tracking_period_id || '/' || region || '/' || fact_date || '/' || event_suffix,
               'price/' || tracking_period_id || '/' || region || '/' || before_date,
               'price/' || tracking_period_id || '/' || region || '/' || fact_date,
               before_observed_at, observed_at,
               jsonb_build_object('before_projection_version', before_projection_version,
                                  'after_projection_version', projection_version,
                                  'tracking_period_id', tracking_period_id),
               transaction_timestamp()
        FROM event_rows
        WHERE event_code IS NOT NULL;
    END IF;

    GET DIAGNOSTICS inserted_rows = ROW_COUNT;

    IF EXISTS (
        SELECT 1
        FROM public.gfg_change_events event
        JOIN public.gfg_change_registry registry
          ON registry.detector_key = event.detector_key
         AND registry.detector_version = event.detector_version
        WHERE event.detector_key = target_detector_key
          AND event.detector_version = target_detector_version
          AND event.projection_date = target_date
          AND NOT (event.event_code = ANY(registry.event_codes))
    ) THEN
        RAISE EXCEPTION 'Game change detector emitted an undeclared event code for %/% on %',
            target_detector_key, target_detector_version, target_date;
    END IF;

    IF EXISTS (
        SELECT 1 FROM public.gfg_change_events
        WHERE detector_key = target_detector_key AND detector_version = target_detector_version
          AND projection_date = target_date AND time_basis = 'observed'
          AND event_at IS DISTINCT FROM source_after_at
    ) THEN
        RAISE EXCEPTION 'Game observed change event timestamp differs from source_after_at for %/% on %',
            target_detector_key, target_detector_version, target_date;
    END IF;

    RETURN inserted_rows;
END;
$function$;
-- +goose StatementEnd

COMMENT ON TABLE public.gfg_change_registry IS
    'Goose-owned versioned Game change detector contracts; Runtime and Admin are read-only.';
COMMENT ON TABLE public.gfg_change_events IS
    'Deterministic canonical Game change events derived only from history, facts, metrics, and periods.';
COMMENT ON TABLE public.gfg_change_checkpoints IS
    'Independent ordered checkpoints for each Game detector version.';

-- +goose Down

DROP FUNCTION public.gfg_project_change_day(text, integer, date);
DROP TABLE public.gfg_change_checkpoints;
DROP TABLE public.gfg_change_events;
DROP TABLE public.gfg_change_registry;
