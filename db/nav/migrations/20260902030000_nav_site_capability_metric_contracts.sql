-- V3-P2.3 Site HTTP/security capability and certificate-verification Metric contracts.
-- +goose Up

INSERT INTO public.gfn_metric_registry (
    metric_key, metric_version, metric_kind, entity_level, time_grain,
    source_facts, eligibility_policy, state_policy, coverage_policy,
    freshness_seconds, allowed_dimensions, status, description
) VALUES
    ('http2_adoption', 1, 'state_ratio', 'site', 'day',
     ARRAY['gfn_site_daily', 'gfn_site_target_daily', 'gfn_site_target_protocol_daily']::text[],
     'active_site_primary_target_v1', 'http2_adoption_state_v1',
     'known_over_eligible_v1', 172800,
     ARRAY['group_id', 'nsfw', 'site_country', 'welfare']::text[], 'active',
     'Share of eligible Site Primary Targets where a reliable GoFurry HTTP probe negotiated HTTP/2.'),
    ('hsts_adoption', 1, 'state_ratio', 'site', 'day',
     ARRAY['gfn_site_daily', 'gfn_site_target_daily', 'gfn_site_target_protocol_daily']::text[],
     'active_site_primary_target_v1', 'hsts_adoption_state_v1',
     'known_over_eligible_v1', 172800,
     ARRAY['group_id', 'nsfw', 'site_country', 'welfare']::text[], 'active',
     'Share of applicable Site Primary Targets where GoFurry observed a Strict-Transport-Security header.'),
    ('csp_adoption', 1, 'state_ratio', 'site', 'day',
     ARRAY['gfn_site_daily', 'gfn_site_target_daily', 'gfn_site_target_protocol_daily']::text[],
     'active_site_primary_target_v1', 'csp_adoption_state_v1',
     'known_over_eligible_v1', 172800,
     ARRAY['group_id', 'nsfw', 'site_country', 'welfare']::text[], 'active',
     'Share of Site Primary Targets where GoFurry observed an enforcement Content-Security-Policy header.'),
    ('tls_certificate_verification', 1, 'state_ratio', 'site', 'day',
     ARRAY['gfn_site_daily', 'gfn_site_target_daily', 'gfn_site_target_protocol_daily']::text[],
     'active_site_primary_target_v1', 'tls_certificate_verification_state_v1',
     'known_over_eligible_v1', 172800,
     ARRAY['group_id', 'nsfw', 'site_country', 'welfare']::text[], 'active',
     'Share of applicable Site Primary Targets whose certificate GoFurry successfully verified.' );

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
  ON site.pipeline_key = 'nav.site_facts'
WHERE registry.metric_key IN (
    'http2_adoption', 'hsts_adoption', 'csp_adoption',
    'tls_certificate_verification'
)
  AND registry.metric_version = 1;

-- +goose StatementBegin
CREATE FUNCTION public.gfn_project_site_capability_metric_day(
    target_metric_key text,
    target_metric_version integer,
    target_date date
) RETURNS bigint
LANGUAGE plpgsql
AS $function$
DECLARE
    day_end timestamp with time zone := (target_date + 1)::timestamp AT TIME ZONE 'UTC';
    freshness bigint;
    entity_rows bigint;
BEGIN
    SELECT freshness_seconds
    INTO freshness
    FROM public.gfn_metric_registry
    WHERE metric_key = target_metric_key
      AND metric_version = target_metric_version;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'unknown Nav Site capability metric %/%', target_metric_key, target_metric_version;
    END IF;
    IF target_metric_version <> 1 OR target_metric_key NOT IN (
        'http2_adoption', 'hsts_adoption', 'csp_adoption',
        'tls_certificate_verification'
    ) THEN
        RAISE EXCEPTION 'Nav Site capability metric evaluator is not compiled for %/%',
            target_metric_key, target_metric_version;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM public.gfn_site_daily site_fact
        LEFT JOIN public.gfn_site_target_daily target_fact
          ON target_fact.target_tracking_period_id = site_fact.primary_target_tracking_period_id
         AND target_fact.fact_date = site_fact.fact_date
         AND target_fact.finalized_at IS NOT NULL
        WHERE site_fact.fact_date = target_date
          AND site_fact.finalized_at IS NOT NULL
          AND site_fact.tracked_at_end
          AND CASE target_metric_key
              WHEN 'http2_adoption' THEN target_fact.http_state_observed_at > day_end
              WHEN 'csp_adoption' THEN target_fact.http_state_observed_at > day_end
              WHEN 'hsts_adoption' THEN target_fact.http_state_observed_at > day_end
                   OR target_fact.tls_state_observed_at > day_end
              ELSE target_fact.tls_state_observed_at > day_end
          END
    ) THEN
        RAISE EXCEPTION 'Nav Site capability source evidence is after UTC day end for %/% on %',
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
               WHEN target_metric_key = 'http2_adoption' THEN CASE
                   WHEN target_fact.http_state_observed_at IS NOT NULL
                        AND target_fact.http_state_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'stale'
                   WHEN target_fact.http_state_observed_at IS NOT NULL
                        AND upper(btrim(COALESCE(target_fact.http_protocol, ''))) IN ('HTTP/2', 'HTTP/2.0')
                       THEN 'positive'
                   WHEN target_fact.http_state_observed_at IS NOT NULL
                        AND NULLIF(btrim(target_fact.http_protocol), '') IS NOT NULL
                       THEN 'negative'
                   WHEN target_fact.http_state_observed_at IS NOT NULL THEN 'unknown'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count > 0 AND protocol_fact.success_count = 0
                       THEN 'probe_failed'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count = 0
                       THEN 'not_probed'
                   ELSE 'unknown'
               END
               WHEN target_metric_key = 'hsts_adoption' THEN CASE
                   WHEN target_fact.tls_state_observed_at < day_end - make_interval(secs => freshness::double precision)
                        OR target_fact.http_state_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'stale'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND lower(COALESCE(target_fact.tls_handshake, '')) = 'not_tls'
                       THEN 'not_applicable'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND target_fact.http_state_observed_at IS NOT NULL
                        AND jsonb_typeof(target_fact.http_security_headers -> 'strict_transport_security') = 'boolean'
                        AND (target_fact.http_security_headers ->> 'strict_transport_security')::boolean
                       THEN 'positive'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND target_fact.http_state_observed_at IS NOT NULL
                        AND jsonb_typeof(target_fact.http_security_headers -> 'strict_transport_security') = 'boolean'
                        AND NOT (target_fact.http_security_headers ->> 'strict_transport_security')::boolean
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
               WHEN target_metric_key = 'csp_adoption' THEN CASE
                   WHEN target_fact.http_state_observed_at IS NOT NULL
                        AND target_fact.http_state_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'stale'
                   WHEN target_fact.http_state_observed_at IS NOT NULL
                        AND jsonb_typeof(target_fact.http_security_headers -> 'content_security_policy') = 'boolean'
                        AND (target_fact.http_security_headers ->> 'content_security_policy')::boolean
                       THEN 'positive'
                   WHEN target_fact.http_state_observed_at IS NOT NULL
                        AND jsonb_typeof(target_fact.http_security_headers -> 'content_security_policy') = 'boolean'
                        AND NOT (target_fact.http_security_headers ->> 'content_security_policy')::boolean
                       THEN 'negative'
                   WHEN target_fact.http_state_observed_at IS NOT NULL THEN 'unknown'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count > 0 AND protocol_fact.success_count = 0
                       THEN 'probe_failed'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count = 0
                       THEN 'not_probed'
                   ELSE 'unknown'
               END
               WHEN target_metric_key = 'tls_certificate_verification' THEN CASE
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND target_fact.tls_state_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'stale'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND lower(COALESCE(target_fact.tls_handshake, '')) = 'not_tls'
                       THEN 'not_applicable'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND lower(COALESCE(target_fact.tls_handshake, '')) = 'collected'
                        AND target_fact.tls_cert_verified IS TRUE
                       THEN 'positive'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND lower(COALESCE(target_fact.tls_handshake, '')) = 'collected'
                        AND target_fact.tls_cert_verified IS FALSE
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
           END,
           CASE
               WHEN site_fact.primary_target_tracking_period_id IS NULL THEN 'primary_target_unknown'
               WHEN target_metric_key = 'http2_adoption' THEN CASE
                   WHEN target_fact.http_state_observed_at IS NOT NULL
                        AND target_fact.http_state_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'http_state_stale'
                   WHEN target_fact.http_state_observed_at IS NOT NULL
                        AND upper(btrim(COALESCE(target_fact.http_protocol, ''))) IN ('HTTP/2', 'HTTP/2.0')
                       THEN 'http2_negotiated'
                   WHEN target_fact.http_state_observed_at IS NOT NULL
                        AND NULLIF(btrim(target_fact.http_protocol), '') IS NOT NULL
                       THEN 'http_protocol_other'
                   WHEN target_fact.http_state_observed_at IS NOT NULL THEN 'http_protocol_unknown'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count > 0 AND protocol_fact.success_count = 0
                       THEN 'http_probe_failed'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count = 0
                       THEN 'http_not_probed'
                   ELSE 'historical_probe_state_unknown'
               END
               WHEN target_metric_key = 'hsts_adoption' THEN CASE
                   WHEN target_fact.tls_state_observed_at < day_end - make_interval(secs => freshness::double precision)
                        OR target_fact.http_state_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'tls_state_stale'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND lower(COALESCE(target_fact.tls_handshake, '')) = 'not_tls'
                       THEN 'target_not_tls'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND target_fact.http_state_observed_at IS NOT NULL
                        AND jsonb_typeof(target_fact.http_security_headers -> 'strict_transport_security') = 'boolean'
                        AND (target_fact.http_security_headers ->> 'strict_transport_security')::boolean
                       THEN 'hsts_present'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND target_fact.http_state_observed_at IS NOT NULL
                        AND jsonb_typeof(target_fact.http_security_headers -> 'strict_transport_security') = 'boolean'
                        AND NOT (target_fact.http_security_headers ->> 'strict_transport_security')::boolean
                       THEN 'hsts_absent'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL THEN 'hsts_field_unknown'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count > 0 AND protocol_fact.success_count = 0
                       THEN 'http_probe_failed'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count = 0
                       THEN 'http_not_probed'
                   ELSE 'historical_probe_state_unknown'
               END
               WHEN target_metric_key = 'csp_adoption' THEN CASE
                   WHEN target_fact.http_state_observed_at IS NOT NULL
                        AND target_fact.http_state_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'http_state_stale'
                   WHEN target_fact.http_state_observed_at IS NOT NULL
                        AND jsonb_typeof(target_fact.http_security_headers -> 'content_security_policy') = 'boolean'
                        AND (target_fact.http_security_headers ->> 'content_security_policy')::boolean
                       THEN 'csp_enforcement_present'
                   WHEN target_fact.http_state_observed_at IS NOT NULL
                        AND jsonb_typeof(target_fact.http_security_headers -> 'content_security_policy') = 'boolean'
                        AND NOT (target_fact.http_security_headers ->> 'content_security_policy')::boolean
                       THEN 'csp_enforcement_absent'
                   WHEN target_fact.http_state_observed_at IS NOT NULL THEN 'csp_field_unknown'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count > 0 AND protocol_fact.success_count = 0
                       THEN 'http_probe_failed'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count = 0
                       THEN 'http_not_probed'
                   ELSE 'historical_probe_state_unknown'
               END
               WHEN target_metric_key = 'tls_certificate_verification' THEN CASE
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND target_fact.tls_state_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'tls_state_stale'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND lower(COALESCE(target_fact.tls_handshake, '')) = 'not_tls'
                       THEN 'target_not_tls'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND lower(COALESCE(target_fact.tls_handshake, '')) = 'collected'
                        AND target_fact.tls_cert_verified IS TRUE
                       THEN 'certificate_verified'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND lower(COALESCE(target_fact.tls_handshake, '')) = 'collected'
                        AND target_fact.tls_cert_verified IS FALSE
                       THEN 'certificate_verification_failed'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL
                        AND lower(COALESCE(target_fact.tls_handshake, '')) = 'failed'
                       THEN 'tls_handshake_failed'
                   WHEN target_fact.tls_state_observed_at IS NOT NULL THEN 'certificate_verification_unknown'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count > 0 AND protocol_fact.success_count = 0
                       THEN 'tls_probe_failed'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count = 0
                       THEN 'tls_not_probed'
                   ELSE 'historical_probe_state_unknown'
               END
           END,
           CASE target_metric_key
               WHEN 'http2_adoption' THEN target_fact.http_state_observed_at
               WHEN 'csp_adoption' THEN target_fact.http_state_observed_at
               WHEN 'hsts_adoption' THEN target_fact.http_state_observed_at
               ELSE target_fact.tls_state_observed_at
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
     AND protocol_fact.protocol = 'http'
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
        SELECT entity.site_id, 'site_country',
               COALESCE(NULLIF(btrim(entity.dimension_values ->> 'site_country'), ''), 'unknown')
        FROM entities entity CROSS JOIN registry
        WHERE 'site_country' = ANY(registry.allowed_dimensions)
        UNION ALL
        SELECT entity.site_id, 'nsfw',
               CASE WHEN jsonb_typeof(entity.dimension_values -> 'nsfw') = 'boolean'
                    THEN entity.dimension_values ->> 'nsfw' ELSE 'unknown' END
        FROM entities entity CROSS JOIN registry
        WHERE 'nsfw' = ANY(registry.allowed_dimensions)
        UNION ALL
        SELECT entity.site_id, 'welfare',
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
                 THEN entity.dimension_values -> 'group_ids' ELSE '[]'::jsonb END
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
        RAISE EXCEPTION 'Nav Site capability metric count invariant failed for %/% on %',
            target_metric_key, target_metric_version, target_date;
    END IF;

    RETURN entity_rows;
END;
$function$;
-- +goose StatementEnd

-- +goose Down

DROP FUNCTION public.gfn_project_site_capability_metric_day(text, integer, date);
DELETE FROM public.gfn_metric_daily
WHERE metric_key IN ('http2_adoption', 'hsts_adoption', 'csp_adoption', 'tls_certificate_verification')
  AND metric_version = 1;
DELETE FROM public.gfn_metric_entity_daily
WHERE metric_key IN ('http2_adoption', 'hsts_adoption', 'csp_adoption', 'tls_certificate_verification')
  AND metric_version = 1;
DELETE FROM public.gfn_metric_checkpoints
WHERE metric_key IN ('http2_adoption', 'hsts_adoption', 'csp_adoption', 'tls_certificate_verification')
  AND metric_version = 1;
DELETE FROM public.gfn_metric_registry
WHERE metric_key IN ('http2_adoption', 'hsts_adoption', 'csp_adoption', 'tls_certificate_verification')
  AND metric_version = 1;
