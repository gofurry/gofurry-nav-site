-- V3-P0.3 Nav Analytics Metric Foundation.
-- +goose Up

CREATE TABLE public.gfn_metric_registry (
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
    CONSTRAINT gfn_metric_registry_key_check CHECK (btrim(metric_key) <> ''),
    CONSTRAINT gfn_metric_registry_version_check CHECK (metric_version > 0),
    CONSTRAINT gfn_metric_registry_kind_check CHECK (metric_kind = 'state_ratio'),
    CONSTRAINT gfn_metric_registry_entity_check CHECK (entity_level = 'site'),
    CONSTRAINT gfn_metric_registry_grain_check CHECK (time_grain = 'day'),
    CONSTRAINT gfn_metric_registry_source_shape_check CHECK (
        cardinality(source_facts) > 0 AND array_ndims(source_facts) = 1
    ),
    CONSTRAINT gfn_metric_registry_policy_check CHECK (
        btrim(eligibility_policy) <> ''
        AND btrim(state_policy) <> ''
        AND btrim(coverage_policy) <> ''
    ),
    CONSTRAINT gfn_metric_registry_freshness_check CHECK (
        freshness_seconds IS NULL OR freshness_seconds > 0
    ),
    CONSTRAINT gfn_metric_registry_dimension_shape_check CHECK (
        cardinality(allowed_dimensions) = 0 OR array_ndims(allowed_dimensions) = 1
    ),
    CONSTRAINT gfn_metric_registry_dimension_value_check CHECK (
        allowed_dimensions <@ ARRAY['group_id', 'nsfw', 'site_country', 'welfare']::text[]
    ),
    CONSTRAINT gfn_metric_registry_status_check CHECK (status IN ('active', 'retired')),
    CONSTRAINT gfn_metric_registry_status_shape_check CHECK (
        (status = 'active' AND retired_at IS NULL)
        OR (status = 'retired' AND retired_at IS NOT NULL)
    )
);

CREATE UNIQUE INDEX uq_gfn_metric_registry_active_key
    ON public.gfn_metric_registry (metric_key)
    WHERE status = 'active';

CREATE TABLE public.gfn_metric_entity_daily (
    metric_key text NOT NULL,
    metric_version integer NOT NULL,
    fact_date date NOT NULL,
    site_id bigint NOT NULL,
    state text NOT NULL,
    reason_code text NOT NULL,
    source_observed_at timestamp with time zone,
    dimension_values jsonb NOT NULL DEFAULT '{}'::jsonb,
    source_projection_versions jsonb NOT NULL DEFAULT '{}'::jsonb,
    evaluated_at timestamp with time zone NOT NULL,
    PRIMARY KEY (metric_key, metric_version, fact_date, site_id),
    CONSTRAINT fk_gfn_metric_entity_registry FOREIGN KEY (metric_key, metric_version)
        REFERENCES public.gfn_metric_registry(metric_key, metric_version) ON DELETE RESTRICT,
    CONSTRAINT gfn_metric_entity_state_check CHECK (state IN (
        'positive', 'negative', 'stale', 'not_probed', 'probe_failed',
        'unknown', 'not_applicable'
    )),
    CONSTRAINT gfn_metric_entity_reason_check CHECK (btrim(reason_code) <> ''),
    CONSTRAINT gfn_metric_entity_dimensions_check CHECK (
        jsonb_typeof(dimension_values) = 'object'
    ),
    CONSTRAINT gfn_metric_entity_projection_check CHECK (
        jsonb_typeof(source_projection_versions) = 'object'
    ),
    CONSTRAINT gfn_metric_entity_stale_evidence_check CHECK (
        state <> 'stale' OR source_observed_at IS NOT NULL
    )
);

CREATE INDEX idx_gfn_metric_entity_key_date_state
    ON public.gfn_metric_entity_daily (metric_key, metric_version, fact_date, state);
CREATE INDEX idx_gfn_metric_entity_site_date
    ON public.gfn_metric_entity_daily (site_id, fact_date DESC);

CREATE TABLE public.gfn_metric_daily (
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
    CONSTRAINT fk_gfn_metric_daily_registry FOREIGN KEY (metric_key, metric_version)
        REFERENCES public.gfn_metric_registry(metric_key, metric_version) ON DELETE RESTRICT,
    CONSTRAINT gfn_metric_daily_dimension_check CHECK (
        btrim(dimension_key) <> '' AND btrim(dimension_value) <> ''
    ),
    CONSTRAINT gfn_metric_daily_nonnegative_check CHECK (
        population_count >= 0 AND eligible_count >= 0 AND not_applicable_count >= 0
        AND positive_count >= 0 AND negative_count >= 0 AND stale_count >= 0
        AND not_probed_count >= 0 AND probe_failed_count >= 0 AND unknown_count >= 0
    ),
    CONSTRAINT gfn_metric_daily_population_check CHECK (
        population_count = eligible_count + not_applicable_count
    ),
    CONSTRAINT gfn_metric_daily_eligible_check CHECK (
        eligible_count = positive_count + negative_count + stale_count
            + not_probed_count + probe_failed_count + unknown_count
    )
);

CREATE INDEX idx_gfn_metric_daily_key_date
    ON public.gfn_metric_daily (metric_key, metric_version, fact_date DESC);
CREATE INDEX idx_gfn_metric_daily_dimension_date
    ON public.gfn_metric_daily (dimension_key, dimension_value, fact_date DESC);

CREATE TABLE public.gfn_metric_checkpoints (
    metric_key text NOT NULL,
    metric_version integer NOT NULL,
    source_start_date date,
    processed_through date,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (metric_key, metric_version),
    CONSTRAINT fk_gfn_metric_checkpoint_registry FOREIGN KEY (metric_key, metric_version)
        REFERENCES public.gfn_metric_registry(metric_key, metric_version) ON DELETE RESTRICT,
    CONSTRAINT gfn_metric_checkpoint_range_check CHECK (
        source_start_date IS NULL OR processed_through IS NULL
        OR processed_through >= source_start_date
    )
);

INSERT INTO public.gfn_metric_registry (
    metric_key, metric_version, metric_kind, entity_level, time_grain,
    source_facts, eligibility_policy, state_policy, coverage_policy,
    freshness_seconds, allowed_dimensions, status, description
) VALUES
    ('ipv6_adoption', 1, 'state_ratio', 'site', 'day',
     ARRAY['gfn_site_daily', 'gfn_site_target_daily', 'gfn_site_target_protocol_daily']::text[],
     'active_site_primary_target_v1', 'ipv6_adoption_state_v1',
     'known_over_eligible_v1', 259200,
     ARRAY['group_id', 'nsfw', 'site_country', 'welfare']::text[], 'active',
     'Share of tracked Sites whose historical Primary Target has IPv6 DNS support.'),
    ('security_txt_adoption', 1, 'state_ratio', 'site', 'day',
     ARRAY['gfn_site_daily', 'gfn_site_target_daily', 'gfn_site_target_protocol_daily']::text[],
     'active_site_primary_target_v1', 'security_txt_adoption_state_v1',
     'known_over_eligible_v1', 1814400,
     ARRAY['group_id', 'nsfw', 'site_country', 'welfare']::text[], 'active',
     'Share of tracked Sites whose historical Primary Target publishes security.txt.'),
    ('tls13_adoption', 1, 'state_ratio', 'site', 'day',
     ARRAY['gfn_site_daily', 'gfn_site_target_daily', 'gfn_site_target_protocol_daily']::text[],
     'active_site_primary_target_v1', 'tls13_adoption_state_v1',
     'known_over_eligible_v1', 172800,
     ARRAY['group_id', 'nsfw', 'site_country', 'welfare']::text[], 'active',
     'Share of applicable tracked Sites whose historical Primary Target negotiates TLS 1.3.');

INSERT INTO public.gfn_metric_checkpoints (
    metric_key, metric_version, source_start_date
)
SELECT registry.metric_key,
       registry.metric_version,
       CASE
           WHEN target.source_start_date IS NULL OR site.source_start_date IS NULL THEN NULL
           ELSE GREATEST(target.source_start_date, site.source_start_date)
       END
FROM public.gfn_metric_registry registry
LEFT JOIN public.gfn_fact_rollup_checkpoints target
  ON target.pipeline_key = 'nav.target_facts'
LEFT JOIN public.gfn_fact_rollup_checkpoints site
  ON site.pipeline_key = 'nav.site_facts';

-- +goose StatementBegin
CREATE FUNCTION public.gfn_project_metric_day(
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
    source_protocol text;
    entity_rows bigint;
BEGIN
    SELECT freshness_seconds, allowed_dimensions
    INTO freshness, dimensions
    FROM public.gfn_metric_registry
    WHERE metric_key = target_metric_key
      AND metric_version = target_metric_version;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'unknown Nav metric %/%', target_metric_key, target_metric_version;
    END IF;
    IF target_metric_version <> 1
       OR target_metric_key NOT IN ('ipv6_adoption', 'tls13_adoption', 'security_txt_adoption') THEN
        RAISE EXCEPTION 'Nav metric evaluator is not compiled for %/%', target_metric_key, target_metric_version;
    END IF;

    source_protocol := CASE target_metric_key
        WHEN 'ipv6_adoption' THEN 'dns'
        WHEN 'tls13_adoption' THEN 'http'
        WHEN 'security_txt_adoption' THEN 'security_txt'
    END;

    IF EXISTS (
        SELECT 1
        FROM public.gfn_site_daily site_fact
        LEFT JOIN public.gfn_site_target_daily target_fact
          ON target_fact.target_tracking_period_id = site_fact.primary_target_tracking_period_id
         AND target_fact.fact_date = site_fact.fact_date
         AND target_fact.finalized_at IS NOT NULL
        LEFT JOIN public.gfn_site_target_protocol_daily protocol_fact
          ON protocol_fact.target_tracking_period_id = site_fact.primary_target_tracking_period_id
         AND protocol_fact.fact_date = site_fact.fact_date
         AND protocol_fact.protocol = source_protocol
         AND protocol_fact.finalized_at IS NOT NULL
        WHERE site_fact.fact_date = target_date
          AND site_fact.finalized_at IS NOT NULL
          AND site_fact.tracked_at_end
          AND CASE target_metric_key
              WHEN 'ipv6_adoption' THEN target_fact.dns_state_observed_at > day_end
              WHEN 'tls13_adoption' THEN target_fact.tls_state_observed_at > day_end
              WHEN 'security_txt_adoption' THEN protocol_fact.known_state_observed_at > day_end
          END
    ) THEN
        RAISE EXCEPTION 'Nav metric source evidence is after UTC day end for %/% on %',
            target_metric_key, target_metric_version, target_date;
    END IF;

    DELETE FROM public.gfn_metric_entity_daily
    WHERE metric_key = target_metric_key
      AND metric_version = target_metric_version
      AND fact_date = target_date;
    DELETE FROM public.gfn_metric_daily
    WHERE metric_key = target_metric_key
      AND metric_version = target_metric_version
      AND fact_date = target_date;

    INSERT INTO public.gfn_metric_entity_daily (
        metric_key, metric_version, fact_date, site_id, state, reason_code,
        source_observed_at, dimension_values, source_projection_versions,
        evaluated_at
    )
    SELECT target_metric_key,
           target_metric_version,
           target_date,
           site_fact.site_id,
           CASE
               WHEN site_fact.primary_target_tracking_period_id IS NULL THEN 'unknown'
               WHEN target_metric_key = 'ipv6_adoption' THEN CASE
                   WHEN target_fact.dns_state_observed_at IS NOT NULL
                        AND target_fact.dns_state_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'stale'
                   WHEN target_fact.dns_state_observed_at IS NOT NULL AND target_fact.dns_has_aaaa IS TRUE THEN 'positive'
                   WHEN target_fact.dns_state_observed_at IS NOT NULL AND target_fact.dns_has_aaaa IS FALSE THEN 'negative'
                   WHEN target_fact.dns_state_observed_at IS NOT NULL THEN 'unknown'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count > 0 AND protocol_fact.success_count = 0
                       THEN 'probe_failed'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count = 0
                       THEN 'not_probed'
                   ELSE 'unknown'
               END
               WHEN target_metric_key = 'tls13_adoption' THEN CASE
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND target_fact.tls_state_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'stale'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND lower(COALESCE(target_fact.tls_handshake, '')) = 'not_tls'
                       THEN 'not_applicable'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND upper(replace(COALESCE(target_fact.tls_version, ''), ' ', '')) = 'TLS1.3'
                       THEN 'positive'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND NULLIF(btrim(target_fact.tls_version), '') IS NOT NULL
                       THEN 'negative'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL THEN 'unknown'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count > 0 AND protocol_fact.success_count = 0
                       THEN 'probe_failed'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count = 0
                       THEN 'not_probed'
                   ELSE 'unknown'
               END
               WHEN target_metric_key = 'security_txt_adoption' THEN CASE
                   WHEN protocol_fact.known_state_observed_at IS NOT NULL
                        AND protocol_fact.known_state_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'stale'
                   WHEN protocol_fact.known_state_observed_at IS NOT NULL
                        AND jsonb_typeof(protocol_fact.known_state -> 'exists') = 'boolean'
                        AND (protocol_fact.known_state ->> 'exists')::boolean
                       THEN 'positive'
                   WHEN protocol_fact.known_state_observed_at IS NOT NULL
                        AND jsonb_typeof(protocol_fact.known_state -> 'exists') = 'boolean'
                        AND NOT (protocol_fact.known_state ->> 'exists')::boolean
                       THEN 'negative'
                   WHEN protocol_fact.known_state_observed_at IS NOT NULL THEN 'unknown'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count > 0 AND protocol_fact.success_count = 0
                       THEN 'probe_failed'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count = 0
                       THEN 'not_probed'
                   ELSE 'unknown'
               END
           END,
           CASE
               WHEN site_fact.primary_target_tracking_period_id IS NULL THEN 'primary_target_unknown'
               WHEN target_metric_key = 'ipv6_adoption' THEN CASE
                   WHEN target_fact.dns_state_observed_at IS NOT NULL
                        AND target_fact.dns_state_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'dns_state_stale'
                   WHEN target_fact.dns_state_observed_at IS NOT NULL AND target_fact.dns_has_aaaa IS TRUE THEN 'aaaa_present'
                   WHEN target_fact.dns_state_observed_at IS NOT NULL AND target_fact.dns_has_aaaa IS FALSE THEN 'aaaa_absent'
                   WHEN target_fact.dns_state_observed_at IS NOT NULL THEN 'dns_metric_field_unknown'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count > 0 AND protocol_fact.success_count = 0
                       THEN 'dns_probe_failed'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count = 0
                       THEN 'dns_not_probed'
                   ELSE 'historical_probe_state_unknown'
               END
               WHEN target_metric_key = 'tls13_adoption' THEN CASE
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND target_fact.tls_state_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'tls_state_stale'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND lower(COALESCE(target_fact.tls_handshake, '')) = 'not_tls'
                       THEN 'target_not_tls'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND upper(replace(COALESCE(target_fact.tls_version, ''), ' ', '')) = 'TLS1.3'
                       THEN 'tls13_negotiated'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND NULLIF(btrim(target_fact.tls_version), '') IS NOT NULL
                       THEN 'tls_version_other'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL THEN 'tls_version_unknown'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count > 0 AND protocol_fact.success_count = 0
                       THEN 'tls_probe_failed'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count = 0
                       THEN 'tls_not_probed'
                   ELSE 'historical_probe_state_unknown'
               END
               WHEN target_metric_key = 'security_txt_adoption' THEN CASE
                   WHEN protocol_fact.known_state_observed_at IS NOT NULL
                        AND protocol_fact.known_state_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'security_txt_state_stale'
                   WHEN protocol_fact.known_state_observed_at IS NOT NULL
                        AND jsonb_typeof(protocol_fact.known_state -> 'exists') = 'boolean'
                        AND (protocol_fact.known_state ->> 'exists')::boolean
                       THEN 'security_txt_present'
                   WHEN protocol_fact.known_state_observed_at IS NOT NULL
                        AND jsonb_typeof(protocol_fact.known_state -> 'exists') = 'boolean'
                        AND NOT (protocol_fact.known_state ->> 'exists')::boolean
                       THEN 'security_txt_absent'
                   WHEN protocol_fact.known_state_observed_at IS NOT NULL THEN 'security_txt_field_unknown'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count > 0 AND protocol_fact.success_count = 0
                       THEN 'security_txt_probe_failed'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count = 0
                       THEN 'security_txt_not_probed'
                   ELSE 'historical_probe_state_unknown'
               END
           END,
           CASE target_metric_key
               WHEN 'ipv6_adoption' THEN target_fact.dns_state_observed_at
               WHEN 'tls13_adoption' THEN target_fact.tls_state_observed_at
               WHEN 'security_txt_adoption' THEN protocol_fact.known_state_observed_at
           END,
           jsonb_build_object(
               'site_country', to_jsonb(site_fact.site_country),
               'group_ids', to_jsonb(site_fact.group_ids),
               'nsfw', to_jsonb(site_fact.nsfw),
               'welfare', to_jsonb(site_fact.welfare)
           ),
           jsonb_strip_nulls(jsonb_build_object(
               'gfn_site_daily', site_fact.projection_version,
               'gfn_site_target_daily', target_fact.projection_version,
               'gfn_site_target_protocol_daily', protocol_fact.projection_version
           )),
           transaction_timestamp()
    FROM public.gfn_site_daily site_fact
    LEFT JOIN public.gfn_site_target_daily target_fact
      ON target_fact.target_tracking_period_id = site_fact.primary_target_tracking_period_id
     AND target_fact.fact_date = site_fact.fact_date
     AND target_fact.finalized_at IS NOT NULL
    LEFT JOIN public.gfn_site_target_protocol_daily protocol_fact
      ON protocol_fact.target_tracking_period_id = site_fact.primary_target_tracking_period_id
     AND protocol_fact.fact_date = site_fact.fact_date
     AND protocol_fact.protocol = source_protocol
     AND protocol_fact.finalized_at IS NOT NULL
    WHERE site_fact.fact_date = target_date
      AND site_fact.finalized_at IS NOT NULL
      AND site_fact.tracked_at_end;

    GET DIAGNOSTICS entity_rows = ROW_COUNT;

    WITH registry AS (
        SELECT allowed_dimensions
        FROM public.gfn_metric_registry
        WHERE metric_key = target_metric_key
          AND metric_version = target_metric_version
    ), entities AS (
        SELECT *
        FROM public.gfn_metric_entity_daily
        WHERE metric_key = target_metric_key
          AND metric_version = target_metric_version
          AND fact_date = target_date
    ), memberships AS (
        SELECT entity.site_id, 'global'::text AS dimension_key, 'all'::text AS dimension_value
        FROM entities entity
        UNION ALL
        SELECT entity.site_id,
               'site_country',
               COALESCE(NULLIF(btrim(entity.dimension_values ->> 'site_country'), ''), 'unknown')
        FROM entities entity CROSS JOIN registry
        WHERE 'site_country' = ANY(registry.allowed_dimensions)
        UNION ALL
        SELECT entity.site_id,
               'nsfw',
               CASE WHEN jsonb_typeof(entity.dimension_values -> 'nsfw') = 'boolean'
                    THEN entity.dimension_values ->> 'nsfw' ELSE 'unknown' END
        FROM entities entity CROSS JOIN registry
        WHERE 'nsfw' = ANY(registry.allowed_dimensions)
        UNION ALL
        SELECT entity.site_id,
               'welfare',
               CASE WHEN jsonb_typeof(entity.dimension_values -> 'welfare') = 'boolean'
                    THEN entity.dimension_values ->> 'welfare' ELSE 'unknown' END
        FROM entities entity CROSS JOIN registry
        WHERE 'welfare' = ANY(registry.allowed_dimensions)
        UNION ALL
        SELECT entity.site_id, 'group_id', member.value
        FROM entities entity
        CROSS JOIN registry
        CROSS JOIN LATERAL jsonb_array_elements_text(
            CASE WHEN jsonb_typeof(entity.dimension_values -> 'group_ids') = 'array'
                 THEN entity.dimension_values -> 'group_ids'
                 ELSE '[]'::jsonb END
        ) member(value)
        WHERE 'group_id' = ANY(registry.allowed_dimensions)
          AND member.value ~ '^[0-9]+$'
    ), unique_memberships AS (
        SELECT DISTINCT site_id, dimension_key, dimension_value
        FROM memberships
    )
    INSERT INTO public.gfn_metric_daily (
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
    JOIN entities entity ON entity.site_id = membership.site_id
    GROUP BY membership.dimension_key, membership.dimension_value;

    INSERT INTO public.gfn_metric_daily (
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
        FROM public.gfn_metric_daily daily
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
        RAISE EXCEPTION 'Nav metric count invariant failed for %/% on %',
            target_metric_key, target_metric_version, target_date;
    END IF;

    RETURN entity_rows;
END;
$function$;
-- +goose StatementEnd

COMMENT ON TABLE public.gfn_metric_registry IS
    'Goose-owned versioned Nav metric contracts; runtime and Admin are read-only.';
COMMENT ON TABLE public.gfn_metric_entity_daily IS
    'Explainable per-Site historical metric state derived only from finalized Nav Facts.';
COMMENT ON TABLE public.gfn_metric_daily IS
    'Global and single-dimension Nav metric state counts; ratios are query-time only.';
COMMENT ON TABLE public.gfn_metric_checkpoints IS
    'Independent ordered checkpoints for each Nav metric version.';

-- +goose Down

DROP FUNCTION public.gfn_project_metric_day(text, integer, date);
DROP TABLE public.gfn_metric_checkpoints;
DROP TABLE public.gfn_metric_daily;
DROP TABLE public.gfn_metric_entity_daily;
DROP TABLE public.gfn_metric_registry;
