-- V3-P0.3 Game Analytics Metric Foundation.
-- +goose Up

CREATE TABLE public.gfg_metric_registry (
    metric_key text NOT NULL,
    metric_version integer NOT NULL,
    metric_kind text NOT NULL,
    entity_level text NOT NULL,
    time_grain text NOT NULL,
    source_facts text[] NOT NULL,
    eligibility_policy text NOT NULL,
    state_policy text NOT NULL,
    coverage_policy text NOT NULL,
    freshness_seconds bigint,
    allowed_dimensions text[] NOT NULL,
    status text NOT NULL,
    description text NOT NULL DEFAULT '',
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    retired_at timestamp with time zone,
    PRIMARY KEY (metric_key, metric_version),
    CONSTRAINT gfg_metric_registry_key_check CHECK (btrim(metric_key) <> ''),
    CONSTRAINT gfg_metric_registry_version_check CHECK (metric_version > 0),
    CONSTRAINT gfg_metric_registry_kind_check CHECK (metric_kind = 'state_ratio'),
    CONSTRAINT gfg_metric_registry_entity_check CHECK (entity_level = 'game'),
    CONSTRAINT gfg_metric_registry_grain_check CHECK (time_grain = 'day'),
    CONSTRAINT gfg_metric_registry_source_shape_check CHECK (
        cardinality(source_facts) > 0 AND array_ndims(source_facts) = 1
    ),
    CONSTRAINT gfg_metric_registry_policy_check CHECK (
        btrim(eligibility_policy) <> ''
        AND btrim(state_policy) <> ''
        AND btrim(coverage_policy) <> ''
    ),
    CONSTRAINT gfg_metric_registry_freshness_check CHECK (
        freshness_seconds IS NULL OR freshness_seconds > 0
    ),
    CONSTRAINT gfg_metric_registry_dimension_shape_check CHECK (
        cardinality(allowed_dimensions) = 0 OR array_ndims(allowed_dimensions) = 1
    ),
    CONSTRAINT gfg_metric_registry_dimension_value_check CHECK (
        allowed_dimensions <@ ARRAY['primary_tag_id', 'tag_id']::text[]
    ),
    CONSTRAINT gfg_metric_registry_status_check CHECK (status IN ('active', 'retired')),
    CONSTRAINT gfg_metric_registry_status_shape_check CHECK (
        (status = 'active' AND retired_at IS NULL)
        OR (status = 'retired' AND retired_at IS NOT NULL)
    )
);

CREATE UNIQUE INDEX uq_gfg_metric_registry_active_key
    ON public.gfg_metric_registry (metric_key)
    WHERE status = 'active';

CREATE TABLE public.gfg_metric_entity_daily (
    metric_key text NOT NULL,
    metric_version integer NOT NULL,
    fact_date date NOT NULL,
    game_id bigint NOT NULL,
    state text NOT NULL,
    reason_code text NOT NULL,
    source_observed_at timestamp with time zone,
    dimension_values jsonb NOT NULL DEFAULT '{}'::jsonb,
    source_projection_versions jsonb NOT NULL DEFAULT '{}'::jsonb,
    evaluated_at timestamp with time zone NOT NULL,
    PRIMARY KEY (metric_key, metric_version, fact_date, game_id),
    CONSTRAINT fk_gfg_metric_entity_registry FOREIGN KEY (metric_key, metric_version)
        REFERENCES public.gfg_metric_registry(metric_key, metric_version) ON DELETE RESTRICT,
    CONSTRAINT gfg_metric_entity_state_check CHECK (state IN (
        'positive', 'negative', 'stale', 'not_probed', 'probe_failed',
        'unknown', 'not_applicable'
    )),
    CONSTRAINT gfg_metric_entity_reason_check CHECK (btrim(reason_code) <> ''),
    CONSTRAINT gfg_metric_entity_dimensions_check CHECK (
        jsonb_typeof(dimension_values) = 'object'
    ),
    CONSTRAINT gfg_metric_entity_projection_check CHECK (
        jsonb_typeof(source_projection_versions) = 'object'
    ),
    CONSTRAINT gfg_metric_entity_stale_evidence_check CHECK (
        state <> 'stale' OR source_observed_at IS NOT NULL
    )
);

CREATE INDEX idx_gfg_metric_entity_key_date_state
    ON public.gfg_metric_entity_daily (metric_key, metric_version, fact_date, state);
CREATE INDEX idx_gfg_metric_entity_game_date
    ON public.gfg_metric_entity_daily (game_id, fact_date DESC);

CREATE TABLE public.gfg_metric_daily (
    metric_key text NOT NULL,
    metric_version integer NOT NULL,
    fact_date date NOT NULL,
    dimension_key text NOT NULL,
    dimension_value text NOT NULL,
    population_count bigint NOT NULL,
    eligible_count bigint NOT NULL,
    not_applicable_count bigint NOT NULL,
    positive_count bigint NOT NULL,
    negative_count bigint NOT NULL,
    stale_count bigint NOT NULL,
    not_probed_count bigint NOT NULL,
    probe_failed_count bigint NOT NULL,
    unknown_count bigint NOT NULL,
    computed_at timestamp with time zone NOT NULL,
    PRIMARY KEY (metric_key, metric_version, fact_date, dimension_key, dimension_value),
    CONSTRAINT fk_gfg_metric_daily_registry FOREIGN KEY (metric_key, metric_version)
        REFERENCES public.gfg_metric_registry(metric_key, metric_version) ON DELETE RESTRICT,
    CONSTRAINT gfg_metric_daily_dimension_check CHECK (
        btrim(dimension_key) <> '' AND btrim(dimension_value) <> ''
    ),
    CONSTRAINT gfg_metric_daily_nonnegative_check CHECK (
        population_count >= 0 AND eligible_count >= 0 AND not_applicable_count >= 0
        AND positive_count >= 0 AND negative_count >= 0 AND stale_count >= 0
        AND not_probed_count >= 0 AND probe_failed_count >= 0 AND unknown_count >= 0
    ),
    CONSTRAINT gfg_metric_daily_population_check CHECK (
        population_count = eligible_count + not_applicable_count
    ),
    CONSTRAINT gfg_metric_daily_eligible_check CHECK (
        eligible_count = positive_count + negative_count + stale_count
            + not_probed_count + probe_failed_count + unknown_count
    )
);

CREATE INDEX idx_gfg_metric_daily_key_date
    ON public.gfg_metric_daily (metric_key, metric_version, fact_date DESC);
CREATE INDEX idx_gfg_metric_daily_dimension_date
    ON public.gfg_metric_daily (dimension_key, dimension_value, fact_date DESC);

CREATE TABLE public.gfg_metric_checkpoints (
    metric_key text NOT NULL,
    metric_version integer NOT NULL,
    source_start_date date,
    processed_through date,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (metric_key, metric_version),
    CONSTRAINT fk_gfg_metric_checkpoint_registry FOREIGN KEY (metric_key, metric_version)
        REFERENCES public.gfg_metric_registry(metric_key, metric_version) ON DELETE RESTRICT,
    CONSTRAINT gfg_metric_checkpoint_range_check CHECK (
        source_start_date IS NULL OR processed_through IS NULL
        OR processed_through >= source_start_date
    )
);

INSERT INTO public.gfg_metric_registry (
    metric_key, metric_version, metric_kind, entity_level, time_grain,
    source_facts, eligibility_policy, state_policy, coverage_policy,
    freshness_seconds, allowed_dimensions, status, description
) VALUES
    ('free_game_share', 1, 'state_ratio', 'game', 'day',
     ARRAY['gfg_game_daily']::text[], 'tracked_game_availability_v1',
     'free_game_share_state_v1', 'known_over_eligible_v1', 259200,
     ARRAY['primary_tag_id', 'tag_id']::text[], 'active',
     'Share of available tracked games that are free.'),
    ('linux_support', 1, 'state_ratio', 'game', 'day',
     ARRAY['gfg_game_daily']::text[], 'tracked_game_v1',
     'linux_support_state_v1', 'known_over_eligible_v1', 259200,
     ARRAY['primary_tag_id', 'tag_id']::text[], 'active',
     'Share of tracked games with Linux support.'),
    ('windows_support', 1, 'state_ratio', 'game', 'day',
     ARRAY['gfg_game_daily']::text[], 'tracked_game_v1',
     'windows_support_state_v1', 'known_over_eligible_v1', 259200,
     ARRAY['primary_tag_id', 'tag_id']::text[], 'active',
     'Share of tracked games with Windows support.');

INSERT INTO public.gfg_metric_checkpoints (
    metric_key, metric_version, source_start_date
)
SELECT registry.metric_key, registry.metric_version, facts.source_start_date
FROM public.gfg_metric_registry registry
LEFT JOIN public.gfg_fact_rollup_checkpoints facts
  ON facts.pipeline_key = 'game.state_facts';

-- +goose StatementBegin
CREATE FUNCTION public.gfg_project_metric_day(
    target_metric_key text,
    target_metric_version integer,
    target_date date
) RETURNS bigint
LANGUAGE plpgsql
AS $function$
DECLARE
    day_end timestamp with time zone := (target_date + 1)::timestamp AT TIME ZONE 'UTC';
    freshness bigint;
    dimensions text[];
    entity_rows bigint;
BEGIN
    SELECT freshness_seconds, allowed_dimensions
    INTO freshness, dimensions
    FROM public.gfg_metric_registry
    WHERE metric_key = target_metric_key
      AND metric_version = target_metric_version;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'unknown Game metric %/%', target_metric_key, target_metric_version;
    END IF;
    IF target_metric_version <> 1
       OR target_metric_key NOT IN ('free_game_share', 'windows_support', 'linux_support') THEN
        RAISE EXCEPTION 'Game metric evaluator is not compiled for %/%', target_metric_key, target_metric_version;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM public.gfg_game_daily fact
        WHERE fact.fact_date = target_date
          AND fact.finalized_at IS NOT NULL
          AND fact.tracked_at_end
          AND fact.details_observed_at > day_end
    ) THEN
        RAISE EXCEPTION 'Game metric source evidence is after UTC day end for %/% on %',
            target_metric_key, target_metric_version, target_date;
    END IF;

    DELETE FROM public.gfg_metric_entity_daily
    WHERE metric_key = target_metric_key
      AND metric_version = target_metric_version
      AND fact_date = target_date;
    DELETE FROM public.gfg_metric_daily
    WHERE metric_key = target_metric_key
      AND metric_version = target_metric_version
      AND fact_date = target_date;

    INSERT INTO public.gfg_metric_entity_daily (
        metric_key, metric_version, fact_date, game_id, state, reason_code,
        source_observed_at, dimension_values, source_projection_versions,
        evaluated_at
    )
    SELECT target_metric_key,
           target_metric_version,
           target_date,
           fact.game_id,
           CASE target_metric_key
               WHEN 'free_game_share' THEN CASE
                   WHEN fact.release_availability = 'upcoming' THEN 'not_applicable'
                   WHEN fact.release_availability IS NULL
                        OR fact.release_availability = 'unknown' THEN 'unknown'
                   WHEN fact.details_observed_at IS NULL THEN 'unknown'
                   WHEN fact.details_observed_at < day_end - make_interval(secs => freshness::double precision) THEN 'stale'
                   WHEN fact.is_free IS TRUE THEN 'positive'
                   WHEN fact.is_free IS FALSE THEN 'negative'
                   ELSE 'unknown'
               END
               WHEN 'windows_support' THEN CASE
                   WHEN fact.details_observed_at IS NULL THEN 'unknown'
                   WHEN fact.details_observed_at < day_end - make_interval(secs => freshness::double precision) THEN 'stale'
                   WHEN fact.windows IS TRUE THEN 'positive'
                   WHEN fact.windows IS FALSE THEN 'negative'
                   ELSE 'unknown'
               END
               WHEN 'linux_support' THEN CASE
                   WHEN fact.details_observed_at IS NULL THEN 'unknown'
                   WHEN fact.details_observed_at < day_end - make_interval(secs => freshness::double precision) THEN 'stale'
                   WHEN fact.linux IS TRUE THEN 'positive'
                   WHEN fact.linux IS FALSE THEN 'negative'
                   ELSE 'unknown'
               END
           END,
           CASE target_metric_key
               WHEN 'free_game_share' THEN CASE
                   WHEN fact.release_availability = 'upcoming' THEN 'game_not_available'
                   WHEN fact.release_availability IS NULL
                        OR fact.release_availability = 'unknown' THEN 'game_availability_unknown'
                   WHEN fact.details_observed_at IS NOT NULL
                        AND fact.details_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'details_state_stale'
                   WHEN fact.is_free IS TRUE THEN 'game_is_free'
                   WHEN fact.is_free IS FALSE THEN 'game_is_paid'
                   ELSE 'is_free_unknown'
               END
               WHEN 'windows_support' THEN CASE
                   WHEN fact.details_observed_at IS NOT NULL
                        AND fact.details_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'details_state_stale'
                   WHEN fact.windows IS TRUE THEN 'windows_supported'
                   WHEN fact.windows IS FALSE THEN 'windows_not_supported'
                   ELSE 'windows_support_unknown'
               END
               WHEN 'linux_support' THEN CASE
                   WHEN fact.details_observed_at IS NOT NULL
                        AND fact.details_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'details_state_stale'
                   WHEN fact.linux IS TRUE THEN 'linux_supported'
                   WHEN fact.linux IS FALSE THEN 'linux_not_supported'
                   ELSE 'linux_support_unknown'
               END
           END,
           fact.details_observed_at,
           jsonb_build_object(
               'primary_tag_id', to_jsonb(fact.primary_tag_id),
               'tag_ids', to_jsonb(fact.tag_ids)
           ),
           jsonb_build_object('gfg_game_daily', fact.projection_version),
           transaction_timestamp()
    FROM public.gfg_game_daily fact
    WHERE fact.fact_date = target_date
      AND fact.finalized_at IS NOT NULL
      AND fact.tracked_at_end;

    GET DIAGNOSTICS entity_rows = ROW_COUNT;

    WITH registry AS (
        SELECT allowed_dimensions
        FROM public.gfg_metric_registry
        WHERE metric_key = target_metric_key
          AND metric_version = target_metric_version
    ), entities AS (
        SELECT *
        FROM public.gfg_metric_entity_daily
        WHERE metric_key = target_metric_key
          AND metric_version = target_metric_version
          AND fact_date = target_date
    ), memberships AS (
        SELECT entity.game_id, 'global'::text AS dimension_key, 'all'::text AS dimension_value
        FROM entities entity
        UNION ALL
        SELECT entity.game_id,
               'primary_tag_id',
               COALESCE(NULLIF(btrim(entity.dimension_values ->> 'primary_tag_id'), ''), 'unknown')
        FROM entities entity CROSS JOIN registry
        WHERE 'primary_tag_id' = ANY(registry.allowed_dimensions)
        UNION ALL
        SELECT entity.game_id, 'tag_id', member.value
        FROM entities entity
        CROSS JOIN registry
        CROSS JOIN LATERAL jsonb_array_elements_text(
            CASE WHEN jsonb_typeof(entity.dimension_values -> 'tag_ids') = 'array'
                 THEN entity.dimension_values -> 'tag_ids'
                 ELSE '[]'::jsonb END
        ) member(value)
        WHERE 'tag_id' = ANY(registry.allowed_dimensions)
          AND member.value ~ '^[0-9]+$'
    ), unique_memberships AS (
        SELECT DISTINCT game_id, dimension_key, dimension_value
        FROM memberships
    )
    INSERT INTO public.gfg_metric_daily (
        metric_key, metric_version, fact_date, dimension_key, dimension_value,
        population_count, eligible_count, not_applicable_count,
        positive_count, negative_count, stale_count, not_probed_count,
        probe_failed_count, unknown_count, computed_at
    )
    SELECT target_metric_key,
           target_metric_version,
           target_date,
           membership.dimension_key,
           membership.dimension_value,
           count(*)::bigint,
           count(*) FILTER (WHERE entity.state <> 'not_applicable')::bigint,
           count(*) FILTER (WHERE entity.state = 'not_applicable')::bigint,
           count(*) FILTER (WHERE entity.state = 'positive')::bigint,
           count(*) FILTER (WHERE entity.state = 'negative')::bigint,
           count(*) FILTER (WHERE entity.state = 'stale')::bigint,
           count(*) FILTER (WHERE entity.state = 'not_probed')::bigint,
           count(*) FILTER (WHERE entity.state = 'probe_failed')::bigint,
           count(*) FILTER (WHERE entity.state = 'unknown')::bigint,
           transaction_timestamp()
    FROM unique_memberships membership
    JOIN entities entity ON entity.game_id = membership.game_id
    GROUP BY membership.dimension_key, membership.dimension_value;

    INSERT INTO public.gfg_metric_daily (
        metric_key, metric_version, fact_date, dimension_key, dimension_value,
        population_count, eligible_count, not_applicable_count,
        positive_count, negative_count, stale_count, not_probed_count,
        probe_failed_count, unknown_count, computed_at
    ) VALUES (
        target_metric_key, target_metric_version, target_date, 'global', 'all',
        0, 0, 0, 0, 0, 0, 0, 0, 0, transaction_timestamp()
    ) ON CONFLICT (metric_key, metric_version, fact_date, dimension_key, dimension_value)
      DO NOTHING;

    IF EXISTS (
        SELECT 1
        FROM public.gfg_metric_daily daily
        WHERE daily.metric_key = target_metric_key
          AND daily.metric_version = target_metric_version
          AND daily.fact_date = target_date
          AND (
              daily.population_count <> daily.eligible_count + daily.not_applicable_count
              OR daily.eligible_count <> daily.positive_count + daily.negative_count
                  + daily.stale_count + daily.not_probed_count
                  + daily.probe_failed_count + daily.unknown_count
          )
    ) THEN
        RAISE EXCEPTION 'Game metric count invariant failed for %/% on %',
            target_metric_key, target_metric_version, target_date;
    END IF;

    RETURN entity_rows;
END;
$function$;
-- +goose StatementEnd

COMMENT ON TABLE public.gfg_metric_registry IS
    'Goose-owned versioned Game metric contracts; runtime and Admin are read-only.';
COMMENT ON TABLE public.gfg_metric_entity_daily IS
    'Explainable per-Game historical metric state derived only from finalized Game Facts.';
COMMENT ON TABLE public.gfg_metric_daily IS
    'Global and single-dimension Game metric state counts; ratios are query-time only.';
COMMENT ON TABLE public.gfg_metric_checkpoints IS
    'Independent ordered checkpoints for each Game metric version.';

-- +goose Down

DROP FUNCTION public.gfg_project_metric_day(text, integer, date);
DROP TABLE public.gfg_metric_checkpoints;
DROP TABLE public.gfg_metric_daily;
DROP TABLE public.gfg_metric_entity_daily;
DROP TABLE public.gfg_metric_registry;
