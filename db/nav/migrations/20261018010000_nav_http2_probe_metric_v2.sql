-- #92-A: trustworthy HTTP/2 negotiation evidence; historical v1 remains rebuildable.
-- +goose Up

UPDATE public.gfn_metric_registry
SET status = 'retired', retired_at = transaction_timestamp()
WHERE metric_key = 'http2_adoption' AND metric_version = 1;

-- The v1 Change contract cannot advance after its source metric retires.
-- Keep its registry, checkpoint, events and evaluator; no Change v2 is introduced.
UPDATE public.gfn_change_registry
SET status = 'retired', retired_at = transaction_timestamp()
WHERE detector_key = 'http2_transition' AND detector_version = 1;

INSERT INTO public.gfn_metric_registry (
    metric_key, metric_version, metric_kind, entity_level, time_grain,
    source_facts, eligibility_policy, state_policy, coverage_policy,
    freshness_seconds, allowed_dimensions, status, description
) VALUES (
    'http2_adoption', 2, 'state_ratio', 'site', 'day',
    ARRAY['gfn_site_daily', 'gfn_site_target_daily', 'gfn_site_target_protocol_daily']::text[],
    'active_site_primary_target_v1', 'http2_adoption_state_v2',
    'known_over_eligible_v1', 172800,
    ARRAY['group_id', 'nsfw', 'site_country', 'welfare']::text[], 'active',
    'HTTP/2 协商率：GoFurry 当前出口及代理路径到最终 HTTPS 端点的可信 H2/ALPN 观测；旧探测证据未知，明文 HTTP 不适用。'
);

-- This is a lower bound, not proof of a production binary rollout. A delayed
-- rollout requires a separately authorized forward cutover before reconciliation.
INSERT INTO public.gfn_metric_checkpoints(metric_key, metric_version, source_start_date)
SELECT 'http2_adoption', 2,
       CASE WHEN target.source_start_date IS NULL OR site.source_start_date IS NULL THEN NULL
            ELSE GREATEST(target.source_start_date, site.source_start_date,
                          (transaction_timestamp() AT TIME ZONE 'UTC')::date + 1)
       END
FROM (VALUES (1)) AS seed(id)
LEFT JOIN public.gfn_fact_rollup_checkpoints target ON target.pipeline_key = 'nav.target_facts'
LEFT JOIN public.gfn_fact_rollup_checkpoints site ON site.pipeline_key = 'nav.site_facts';

-- +goose StatementBegin
CREATE FUNCTION public.gfn_project_http2_metric_day_v2(
    target_metric_key text,
    target_metric_version integer,
    target_date date
) RETURNS bigint
LANGUAGE plpgsql
AS $function$
DECLARE
    day_end timestamp with time zone := (target_date + 1)::timestamp AT TIME ZONE 'UTC';
    freshness bigint;
    cutover_start date;
    entity_rows bigint;
BEGIN
    IF target_metric_key IS DISTINCT FROM 'http2_adoption'
       OR target_metric_version IS DISTINCT FROM 2 OR target_date IS NULL THEN
        RAISE EXCEPTION 'Nav HTTP/2 v2 evaluator is not compiled for %/% on %',
            target_metric_key, target_metric_version, target_date;
    END IF;

    SELECT registry.freshness_seconds, checkpoint.source_start_date
    INTO freshness, cutover_start
    FROM public.gfn_metric_registry registry
    JOIN public.gfn_metric_checkpoints checkpoint USING (metric_key, metric_version)
    WHERE registry.metric_key = target_metric_key AND registry.metric_version = target_metric_version;
    IF NOT FOUND OR cutover_start IS NULL OR target_date < cutover_start THEN
        RAISE EXCEPTION 'Nav HTTP/2 v2 day is before or lacks source_start_date: %', target_date;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM public.gfn_site_daily site_fact
        JOIN public.gfn_site_target_protocol_daily protocol_fact
          ON protocol_fact.target_tracking_period_id = site_fact.primary_target_tracking_period_id
         AND protocol_fact.site_id = site_fact.site_id
         AND protocol_fact.fact_date = site_fact.fact_date
         AND protocol_fact.protocol = 'http'
         AND protocol_fact.finalized_at IS NOT NULL
        WHERE site_fact.fact_date = target_date
          AND site_fact.finalized_at IS NOT NULL AND site_fact.tracked_at_end
          AND protocol_fact.known_state_observed_at > day_end
    ) THEN
        RAISE EXCEPTION 'Nav HTTP/2 source evidence is after UTC day end on %', target_date;
    END IF;

    DELETE FROM public.gfn_metric_entity_daily
    WHERE metric_key = target_metric_key AND metric_version = target_metric_version AND fact_date = target_date;
    DELETE FROM public.gfn_metric_daily
    WHERE metric_key = target_metric_key AND metric_version = target_metric_version AND fact_date = target_date;

    WITH evidence AS (
        SELECT site_fact.*,
               target_fact.projection_version AS target_version,
               protocol_fact.projection_version AS protocol_version,
               protocol_fact.known_state_observed_at AS observed_at,
               protocol_fact.known_state AS payload,
               protocol_fact.quality_basis, protocol_fact.attempted_count, protocol_fact.success_count
        FROM public.gfn_site_daily site_fact
        LEFT JOIN public.gfn_site_target_daily target_fact
          ON target_fact.target_tracking_period_id = site_fact.primary_target_tracking_period_id
         AND target_fact.site_id = site_fact.site_id AND target_fact.fact_date = site_fact.fact_date
         AND target_fact.finalized_at IS NOT NULL
        LEFT JOIN public.gfn_site_target_protocol_daily protocol_fact
          ON protocol_fact.target_tracking_period_id = site_fact.primary_target_tracking_period_id
         AND protocol_fact.site_id = site_fact.site_id AND protocol_fact.fact_date = site_fact.fact_date
         AND protocol_fact.protocol = 'http' AND protocol_fact.finalized_at IS NOT NULL
        WHERE site_fact.fact_date = target_date AND site_fact.finalized_at IS NOT NULL
          AND site_fact.tracked_at_end
    ), reasons AS (
        SELECT evidence.*,
               CASE
                   WHEN primary_target_tracking_period_id IS NULL THEN 'primary_target_unknown'
                   WHEN observed_at IS NOT NULL
                        AND observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'http_state_stale'
                   WHEN observed_at IS NULL THEN CASE
                       WHEN quality_basis = 'acquisition_ledger' AND attempted_count > 0 AND success_count = 0
                           THEN 'http_probe_failed'
                       WHEN quality_basis = 'acquisition_ledger' AND attempted_count = 0
                           THEN 'http_not_probed'
                       ELSE 'historical_probe_state_unknown'
                   END
                   WHEN target_version IS NULL THEN 'primary_target_fact_unknown'
                   WHEN payload ->> 'http2_probe_profile' IS DISTINCT FROM 'go_h2_enabled_v1'
                       THEN 'http2_probe_profile_unknown'
                   WHEN jsonb_typeof(payload -> 'final_url') IS DISTINCT FROM 'string'
                     OR jsonb_typeof(payload -> 'final_scheme') IS DISTINCT FROM 'string'
                     OR jsonb_typeof(payload -> 'http_protocol') IS DISTINCT FROM 'string'
                     OR jsonb_typeof(payload -> 'tls_alpn') IS DISTINCT FROM 'string'
                     OR COALESCE(payload ->> 'http_protocol', '') = ''
                       THEN 'http2_evidence_incomplete'
                   -- Require an absolute final endpoint and agreement with its derived scheme.
                   WHEN (payload ->> 'final_url') !~ '^https?://[^/?#[:space:]]+([/?#][^[:space:]]*)?$'
                     OR split_part(payload ->> 'final_url', ':', 1) IS DISTINCT FROM (payload ->> 'final_scheme')
                       THEN 'http2_endpoint_inconsistent'
                   WHEN payload ->> 'final_scheme' = 'http' THEN CASE
                       WHEN payload ->> 'http_protocol' IN ('HTTP/1.0', 'HTTP/1.1') AND payload ->> 'tls_alpn' = ''
                           THEN 'http_plaintext_not_applicable'
                       ELSE 'http2_protocol_alpn_inconsistent'
                   END
                   WHEN payload ->> 'http_protocol' = 'HTTP/2.0' AND payload ->> 'tls_alpn' = 'h2'
                       THEN 'http2_negotiated'
                   WHEN payload ->> 'http_protocol' IN ('HTTP/1.0', 'HTTP/1.1')
                        AND payload ->> 'tls_alpn' IN ('', 'http/1.1')
                       THEN 'http1_negotiated_h2_enabled'
                   WHEN payload ->> 'http_protocol' NOT IN ('HTTP/2.0', 'HTTP/1.0', 'HTTP/1.1')
                       THEN 'http2_protocol_unknown'
                   ELSE 'http2_protocol_alpn_inconsistent'
               END AS reason
        FROM evidence
    )
    INSERT INTO public.gfn_metric_entity_daily (
        metric_key, metric_version, fact_date, site_id, state, reason_code,
        source_observed_at, dimension_values, source_projection_versions, evaluated_at
    )
    SELECT target_metric_key, target_metric_version, target_date, site_id,
           CASE reason
               WHEN 'http2_negotiated' THEN 'positive'
               WHEN 'http1_negotiated_h2_enabled' THEN 'negative'
               WHEN 'http_plaintext_not_applicable' THEN 'not_applicable'
               WHEN 'http_state_stale' THEN 'stale'
               WHEN 'http_probe_failed' THEN 'probe_failed'
               WHEN 'http_not_probed' THEN 'not_probed'
               ELSE 'unknown'
           END,
           reason, observed_at,
           jsonb_build_object('site_country', to_jsonb(site_country), 'group_ids', to_jsonb(group_ids),
                              'nsfw', to_jsonb(nsfw), 'welfare', to_jsonb(welfare)),
           jsonb_strip_nulls(jsonb_build_object(
               'gfn_site_daily', projection_version, 'gfn_site_target_daily', target_version,
               'gfn_site_target_protocol_daily', protocol_version)),
           transaction_timestamp()
    FROM reasons;

    GET DIAGNOSTICS entity_rows = ROW_COUNT;

    WITH registry AS (
        SELECT allowed_dimensions
        FROM public.gfn_metric_registry
        WHERE metric_key = target_metric_key AND metric_version = target_metric_version
    ), entities AS (
        SELECT * FROM public.gfn_metric_entity_daily
        WHERE metric_key = target_metric_key AND metric_version = target_metric_version
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
        FROM entities entity CROSS JOIN registry
        CROSS JOIN LATERAL jsonb_array_elements_text(
            CASE WHEN jsonb_typeof(entity.dimension_values -> 'group_ids') = 'array'
                 THEN entity.dimension_values -> 'group_ids' ELSE '[]'::jsonb END
        ) member(value)
        WHERE 'group_id' = ANY(registry.allowed_dimensions)
          AND member.value ~ '^[0-9]+$'
    ), unique_memberships AS (
        SELECT DISTINCT site_id, dimension_key, dimension_value FROM memberships
    )
    INSERT INTO public.gfn_metric_daily (
        metric_key, metric_version, fact_date, dimension_key, dimension_value,
        population_count, eligible_count, not_applicable_count,
        positive_count, negative_count, stale_count, not_probed_count,
        probe_failed_count, unknown_count, computed_at
    )
    SELECT target_metric_key, target_metric_version, target_date,
           membership.dimension_key, membership.dimension_value,
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
        SELECT 1 FROM public.gfn_metric_daily daily
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

COMMENT ON FUNCTION public.gfn_project_http2_metric_day_v2(text, integer, date) IS
    '按独立切换起点重建指定 UTC 日 HTTP/2 v2 指标：仅可信新版最终 HTTPS 协议与 ALPN 一致时计正负例，明文不适用，保留原始证据时间、失败及未知；原子替换实体和去重单维计数，校验时间与数量守恒，返回实体数。';

-- +goose Down
-- +goose StatementBegin
DO $block$
BEGIN
    RAISE EXCEPTION 'HTTP/2 v2 cutover is irreversible; preserve v1/v2 history and use an explicitly approved forward migration';
END;
$block$;
-- +goose StatementEnd
