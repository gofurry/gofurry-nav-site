-- +goose Up

UPDATE public.gfn_metric_registry
SET status = 'retired', retired_at = transaction_timestamp()
WHERE (metric_key, metric_version) IN (('ipv6_adoption', 1), ('security_txt_adoption', 1));

INSERT INTO public.gfn_metric_registry (
    metric_key, metric_version, metric_kind, entity_level, time_grain,
    source_facts, eligibility_policy, state_policy, coverage_policy,
    freshness_seconds, allowed_dimensions, status, description
) VALUES
    ('ipv6_adoption', 2, 'state_ratio', 'site', 'day',
     ARRAY['gfn_site_daily', 'gfn_site_target_daily', 'gfn_site_target_protocol_daily']::text[],
     'active_site_primary_target_v1', 'ipv6_adoption_state_v2',
     'known_over_eligible_v1', 259200,
     ARRAY['group_id', 'nsfw', 'site_country', 'welfare']::text[], 'active',
     'Share of tracked Sites whose Primary Target has confirmed IPv6 DNS evidence; inconclusive AAAA queries remain unknown.'),
    ('security_txt_adoption', 2, 'state_ratio', 'site', 'day',
     ARRAY['gfn_site_daily', 'gfn_site_target_daily', 'gfn_site_target_protocol_daily']::text[],
     'active_site_primary_target_v1', 'security_txt_adoption_state_v2',
     'known_over_eligible_v1', 1814400,
     ARRAY['group_id', 'nsfw', 'site_country', 'welfare']::text[], 'active',
     'Share of tracked Sites whose Primary Target publishes a recognized valid security.txt document.');

INSERT INTO public.gfn_metric_checkpoints(metric_key, metric_version, source_start_date)
SELECT registry.metric_key, registry.metric_version,
       CASE
           WHEN target.source_start_date IS NULL OR site.source_start_date IS NULL THEN NULL
           ELSE GREATEST(
               target.source_start_date,
               site.source_start_date,
               ((transaction_timestamp() AT TIME ZONE 'UTC')::date + 1)
           )
       END
FROM public.gfn_metric_registry registry
LEFT JOIN public.gfn_fact_rollup_checkpoints target ON target.pipeline_key = 'nav.target_facts'
LEFT JOIN public.gfn_fact_rollup_checkpoints site ON site.pipeline_key = 'nav.site_facts'
WHERE registry.metric_version = 2
  AND registry.metric_key IN ('ipv6_adoption', 'security_txt_adoption');

UPDATE public.gfn_change_registry
SET status = 'retired', retired_at = transaction_timestamp()
WHERE (detector_key, detector_version) IN (('ipv6_transition', 1), ('security_txt_transition', 1));

INSERT INTO public.gfn_change_registry (
    detector_key, detector_version, source_kind, source_contracts,
    detection_policy, watermark_policy, event_codes, processing_grain,
    status, description
) VALUES
    ('ipv6_transition', 2, 'metric', ARRAY['ipv6_adoption/2', 'gfn_site_daily'],
     'metric_semantic_transition_v2', 'metric_checkpoint_v1',
     ARRAY['ipv6_enabled', 'ipv6_disabled'], 'day', 'active',
     'Semantic IPv6 transitions based on confirmed AAAA evidence within one Primary Target identity.'),
    ('security_txt_transition', 2, 'metric', ARRAY['security_txt_adoption/2', 'gfn_site_daily'],
     'metric_semantic_transition_v2', 'metric_checkpoint_v1',
     ARRAY['security_txt_added', 'security_txt_removed'], 'day', 'active',
     'Semantic valid security.txt transitions within one Primary Target identity.');

INSERT INTO public.gfn_change_checkpoints(detector_key, detector_version, source_start_date)
SELECT registry.detector_key, registry.detector_version, checkpoint.source_start_date
FROM public.gfn_change_registry registry
JOIN public.gfn_metric_checkpoints checkpoint
  ON checkpoint.metric_key = CASE registry.detector_key
      WHEN 'ipv6_transition' THEN 'ipv6_adoption'
      WHEN 'security_txt_transition' THEN 'security_txt_adoption'
  END
 AND checkpoint.metric_version = registry.detector_version
WHERE registry.detector_version = 2
  AND registry.detector_key IN ('ipv6_transition', 'security_txt_transition');

-- +goose StatementBegin
CREATE FUNCTION public.gfn_project_metric_day_v2(
    target_metric_key text,
    target_metric_version integer,
    target_date date
) RETURNS bigint
LANGUAGE plpgsql
AS $function$
DECLARE
    day_end timestamp with time zone := (target_date + 1)::timestamp AT TIME ZONE 'UTC';
    freshness bigint;
    source_protocol text;
    entity_rows bigint;
BEGIN
    SELECT freshness_seconds
    INTO freshness
    FROM public.gfn_metric_registry
    WHERE metric_key = target_metric_key
      AND metric_version = target_metric_version;

    IF NOT FOUND OR target_metric_version <> 2
       OR target_metric_key NOT IN ('ipv6_adoption', 'security_txt_adoption') THEN
        RAISE EXCEPTION 'Nav metric v2 evaluator is not compiled for %/%', target_metric_key, target_metric_version;
    END IF;

    source_protocol := CASE target_metric_key
        WHEN 'ipv6_adoption' THEN 'dns'
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
                   WHEN target_fact.dns_state_observed_at IS NOT NULL
                        AND protocol_fact.known_state ->> 'aaaa_evidence' = 'present'
                       THEN 'positive'
                   WHEN target_fact.dns_state_observed_at IS NOT NULL
                        AND protocol_fact.known_state ->> 'aaaa_evidence' = 'confirmed_absent'
                       THEN 'negative'
                   WHEN target_fact.dns_state_observed_at IS NOT NULL THEN 'unknown'
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
                        AND protocol_fact.known_state ->> 'recognition' = 'present_valid'
                       THEN 'positive'
                   WHEN protocol_fact.known_state_observed_at IS NOT NULL
                        AND protocol_fact.known_state ->> 'recognition' = 'absent'
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
                   WHEN target_fact.dns_state_observed_at IS NOT NULL
                        AND protocol_fact.known_state ->> 'aaaa_evidence' = 'present'
                       THEN 'aaaa_present'
                   WHEN target_fact.dns_state_observed_at IS NOT NULL
                        AND protocol_fact.known_state ->> 'aaaa_evidence' = 'confirmed_absent'
                       THEN 'aaaa_confirmed_absent'
                   WHEN target_fact.dns_state_observed_at IS NOT NULL
                        AND protocol_fact.known_state ->> 'aaaa_evidence' = 'unavailable'
                       THEN 'aaaa_evidence_unavailable'
                   WHEN target_fact.dns_state_observed_at IS NOT NULL THEN 'aaaa_evidence_unknown'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count > 0 AND protocol_fact.success_count = 0
                       THEN 'dns_probe_failed'
                   WHEN protocol_fact.quality_basis = 'acquisition_ledger'
                        AND protocol_fact.attempted_count = 0
                       THEN 'dns_not_probed'
                   ELSE 'historical_probe_state_unknown'
               END
               WHEN target_metric_key = 'security_txt_adoption' THEN CASE
                   WHEN protocol_fact.known_state_observed_at IS NOT NULL
                        AND protocol_fact.known_state_observed_at < day_end - make_interval(secs => freshness::double precision)
                       THEN 'security_txt_state_stale'
                   WHEN protocol_fact.known_state_observed_at IS NOT NULL
                        AND protocol_fact.known_state ->> 'recognition' = 'present_valid'
                       THEN 'security_txt_valid'
                   WHEN protocol_fact.known_state_observed_at IS NOT NULL
                        AND protocol_fact.known_state ->> 'recognition' = 'absent'
                       THEN 'security_txt_absent'
                   WHEN protocol_fact.known_state_observed_at IS NOT NULL
                        AND protocol_fact.known_state ->> 'recognition' = 'present_invalid'
                       THEN 'security_txt_invalid'
                   WHEN protocol_fact.known_state_observed_at IS NOT NULL THEN 'security_txt_recognition_unknown'
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

-- +goose StatementBegin
CREATE FUNCTION public.gfn_project_change_day_v2(
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
    IF target_detector_version <> 2
       OR target_detector_key NOT IN ('ipv6_transition', 'security_txt_transition') THEN
        RAISE EXCEPTION 'Nav change v2 detector is not compiled for %/%', target_detector_key, target_detector_version;
    END IF;

    source_metric_key := CASE target_detector_key
        WHEN 'ipv6_transition' THEN 'ipv6_adoption'
        WHEN 'security_txt_transition' THEN 'security_txt_adoption'
    END;

    DELETE FROM public.gfn_change_events
    WHERE detector_key = target_detector_key
      AND detector_version = target_detector_version
      AND projection_date = target_date;

    INSERT INTO public.gfn_change_events (
        event_key, detector_key, detector_version, site_id, projection_date,
        event_at, time_basis, event_code, scope_kind, scope_key,
        old_value, new_value, source_event_key, source_before_key, source_after_key,
        source_before_at, source_after_at, source_versions, materialized_at
    )
    SELECT target_detector_key || '/2/' || event_identity.source_event_key,
           target_detector_key, 2, current_metric.site_id, target_date,
           current_metric.source_observed_at,
           CASE WHEN current_metric.source_observed_at IS NULL THEN 'day' ELSE 'observed' END,
           CASE target_detector_key
               WHEN 'ipv6_transition' THEN CASE WHEN current_metric.state = 'positive' THEN 'ipv6_enabled' ELSE 'ipv6_disabled' END
               WHEN 'security_txt_transition' THEN CASE WHEN current_metric.state = 'positive' THEN 'security_txt_added' ELSE 'security_txt_removed' END
           END,
           'global', 'all',
           jsonb_build_object('state', previous_metric.state, 'reason_code', previous_metric.reason_code),
           jsonb_build_object('state', current_metric.state, 'reason_code', current_metric.reason_code),
           event_identity.source_event_key,
           source_metric_key || '/2/' || previous_metric.fact_date || '/' || current_metric.site_id,
           source_metric_key || '/2/' || current_metric.fact_date || '/' || current_metric.site_id,
           previous_metric.source_observed_at, current_metric.source_observed_at,
           jsonb_build_object(
               'metric_key', source_metric_key, 'metric_version', 2,
               'before_fact_projection_versions', previous_metric.source_projection_versions,
               'after_fact_projection_versions', current_metric.source_projection_versions,
               'primary_target_tracking_period_id', current_fact.primary_target_tracking_period_id
           ),
           transaction_timestamp()
    FROM public.gfn_metric_entity_daily current_metric
    JOIN public.gfn_site_daily current_fact
      ON current_fact.site_id = current_metric.site_id
     AND current_fact.fact_date = current_metric.fact_date
     AND current_fact.finalized_at IS NOT NULL
    JOIN LATERAL (
        SELECT previous_metric.*
        FROM public.gfn_metric_entity_daily previous_metric
        JOIN public.gfn_site_daily previous_fact
          ON previous_fact.site_id = previous_metric.site_id
         AND previous_fact.fact_date = previous_metric.fact_date
         AND previous_fact.finalized_at IS NOT NULL
        WHERE previous_metric.metric_key = current_metric.metric_key
          AND previous_metric.metric_version = current_metric.metric_version
          AND previous_metric.site_id = current_metric.site_id
          AND previous_metric.fact_date < current_metric.fact_date
          AND previous_metric.state IN ('positive', 'negative')
          AND previous_fact.primary_target_tracking_period_id = current_fact.primary_target_tracking_period_id
        ORDER BY previous_metric.fact_date DESC
        LIMIT 1
    ) previous_metric ON true
    CROSS JOIN LATERAL (
        SELECT source_metric_key || '/2/' || current_fact.primary_target_tracking_period_id || '/'
               || current_metric.site_id || '/' || previous_metric.fact_date || '/' || current_metric.fact_date AS source_event_key
    ) event_identity
    WHERE current_metric.metric_key = source_metric_key
      AND current_metric.metric_version = 2
      AND current_metric.fact_date = target_date
      AND current_fact.primary_target_tracking_period_id IS NOT NULL
      AND current_metric.state IN ('positive', 'negative')
      AND current_metric.state <> previous_metric.state;

    GET DIAGNOSTICS inserted_rows = ROW_COUNT;

    IF EXISTS (
        SELECT 1
        FROM public.gfn_change_events event
        JOIN public.gfn_change_registry registry
          ON registry.detector_key = event.detector_key
         AND registry.detector_version = event.detector_version
        WHERE event.detector_key = target_detector_key
          AND event.detector_version = target_detector_version
          AND event.projection_date = target_date
          AND NOT (event.event_code = ANY(registry.event_codes))
    ) THEN
        RAISE EXCEPTION 'Nav change detector emitted an undeclared event code for %/% on %',
            target_detector_key, target_detector_version, target_date;
    END IF;

    IF EXISTS (
        SELECT 1 FROM public.gfn_change_events
        WHERE detector_key = target_detector_key AND detector_version = target_detector_version
          AND projection_date = target_date AND time_basis = 'observed'
          AND event_at IS DISTINCT FROM source_after_at
    ) THEN
        RAISE EXCEPTION 'Nav observed change event timestamp differs from source_after_at for %/% on %',
            target_detector_key, target_detector_version, target_date;
    END IF;

    RETURN inserted_rows;
END;
$function$;
-- +goose StatementEnd

-- +goose Down

DROP FUNCTION public.gfn_project_change_day_v2(text, integer, date);
DROP FUNCTION public.gfn_project_metric_day_v2(text, integer, date);

DELETE FROM public.gfn_change_events
WHERE detector_version = 2 AND detector_key IN ('ipv6_transition', 'security_txt_transition');
DELETE FROM public.gfn_change_checkpoints
WHERE detector_version = 2 AND detector_key IN ('ipv6_transition', 'security_txt_transition');
DELETE FROM public.gfn_change_registry
WHERE detector_version = 2 AND detector_key IN ('ipv6_transition', 'security_txt_transition');
UPDATE public.gfn_change_registry
SET status = 'active', retired_at = NULL
WHERE detector_version = 1 AND detector_key IN ('ipv6_transition', 'security_txt_transition');

DELETE FROM public.gfn_metric_daily
WHERE metric_version = 2 AND metric_key IN ('ipv6_adoption', 'security_txt_adoption');
DELETE FROM public.gfn_metric_entity_daily
WHERE metric_version = 2 AND metric_key IN ('ipv6_adoption', 'security_txt_adoption');
DELETE FROM public.gfn_metric_checkpoints
WHERE metric_version = 2 AND metric_key IN ('ipv6_adoption', 'security_txt_adoption');
DELETE FROM public.gfn_metric_registry
WHERE metric_version = 2 AND metric_key IN ('ipv6_adoption', 'security_txt_adoption');
UPDATE public.gfn_metric_registry
SET status = 'active', retired_at = NULL
WHERE metric_version = 1 AND metric_key IN ('ipv6_adoption', 'security_txt_adoption');
