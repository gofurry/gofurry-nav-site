-- V3-P2.2 Mac support Metric contract and projector extension.
-- +goose Up

INSERT INTO public.gfg_metric_registry (
    metric_key, metric_version, metric_kind, entity_level, time_grain,
    source_facts, eligibility_policy, state_policy, coverage_policy,
    freshness_seconds, allowed_dimensions, status, description
) VALUES (
    'mac_support', 1, 'state_ratio', 'game', 'day',
    ARRAY['gfg_game_daily']::text[], 'tracked_game_v1',
    'mac_support_state_v1', 'known_over_eligible_v1', 259200,
    ARRAY['primary_tag_id', 'tag_id']::text[], 'active',
    'Share of tracked games with macOS support.'
);

INSERT INTO public.gfg_metric_checkpoints (
    metric_key, metric_version, source_start_date
)
SELECT 'mac_support', 1, source_start_date
FROM public.gfg_fact_rollup_checkpoints
WHERE pipeline_key = 'game.state_facts';

-- +goose StatementBegin
CREATE FUNCTION public.gfg_project_mac_metric_day(target_date date) RETURNS bigint
LANGUAGE plpgsql
AS $function$
DECLARE
    day_end timestamp with time zone := (target_date + 1)::timestamp AT TIME ZONE 'UTC';
    freshness bigint;
    entity_rows bigint;
BEGIN
    SELECT freshness_seconds INTO freshness
    FROM public.gfg_metric_registry
    WHERE metric_key = 'mac_support' AND metric_version = 1;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'unknown Game metric mac_support/1';
    END IF;

    IF EXISTS (
        SELECT 1 FROM public.gfg_game_daily fact
        WHERE fact.fact_date = target_date AND fact.finalized_at IS NOT NULL
          AND fact.tracked_at_end AND fact.details_observed_at > day_end
    ) THEN
        RAISE EXCEPTION 'Game metric source evidence is after UTC day end for mac_support/1 on %', target_date;
    END IF;

    DELETE FROM public.gfg_metric_entity_daily
    WHERE metric_key = 'mac_support' AND metric_version = 1 AND fact_date = target_date;
    DELETE FROM public.gfg_metric_daily
    WHERE metric_key = 'mac_support' AND metric_version = 1 AND fact_date = target_date;

    INSERT INTO public.gfg_metric_entity_daily (
        metric_key, metric_version, fact_date, game_id, state, reason_code,
        source_observed_at, dimension_values, source_projection_versions, evaluated_at
    )
    SELECT 'mac_support', 1, target_date, fact.game_id,
           CASE
               WHEN fact.details_observed_at IS NULL THEN 'unknown'
               WHEN fact.details_observed_at < day_end - make_interval(secs => freshness::double precision) THEN 'stale'
               WHEN fact.mac IS TRUE THEN 'positive'
               WHEN fact.mac IS FALSE THEN 'negative'
               ELSE 'unknown'
           END,
           CASE
               WHEN fact.details_observed_at IS NOT NULL
                    AND fact.details_observed_at < day_end - make_interval(secs => freshness::double precision)
                   THEN 'details_state_stale'
               WHEN fact.mac IS TRUE THEN 'mac_supported'
               WHEN fact.mac IS FALSE THEN 'mac_not_supported'
               ELSE 'mac_support_unknown'
           END,
           fact.details_observed_at,
           jsonb_build_object('primary_tag_id', to_jsonb(fact.primary_tag_id), 'tag_ids', to_jsonb(fact.tag_ids)),
           jsonb_build_object('gfg_game_daily', fact.projection_version),
           transaction_timestamp()
    FROM public.gfg_game_daily fact
    WHERE fact.fact_date = target_date
      AND fact.finalized_at IS NOT NULL
      AND fact.tracked_at_end;

    GET DIAGNOSTICS entity_rows = ROW_COUNT;

    WITH entities AS (
        SELECT * FROM public.gfg_metric_entity_daily
        WHERE metric_key = 'mac_support' AND metric_version = 1 AND fact_date = target_date
    ), memberships AS (
        SELECT game_id, 'global'::text AS dimension_key, 'all'::text AS dimension_value FROM entities
        UNION ALL
        SELECT game_id, 'primary_tag_id',
               COALESCE(NULLIF(btrim(dimension_values ->> 'primary_tag_id'), ''), 'unknown')
        FROM entities
        UNION ALL
        SELECT entity.game_id, 'tag_id', member.value
        FROM entities entity
        CROSS JOIN LATERAL jsonb_array_elements_text(
            CASE WHEN jsonb_typeof(entity.dimension_values -> 'tag_ids') = 'array'
                 THEN entity.dimension_values -> 'tag_ids' ELSE '[]'::jsonb END
        ) member(value)
        WHERE member.value ~ '^[0-9]+$'
    ), unique_memberships AS (
        SELECT DISTINCT game_id, dimension_key, dimension_value FROM memberships
    )
    INSERT INTO public.gfg_metric_daily (
        metric_key, metric_version, fact_date, dimension_key, dimension_value,
        population_count, eligible_count, not_applicable_count,
        positive_count, negative_count, stale_count, not_probed_count,
        probe_failed_count, unknown_count, computed_at
    )
    SELECT 'mac_support', 1, target_date, membership.dimension_key, membership.dimension_value,
           count(*)::bigint, count(*)::bigint, 0,
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
        population_count, eligible_count, not_applicable_count, positive_count,
        negative_count, stale_count, not_probed_count, probe_failed_count,
        unknown_count, computed_at
    ) VALUES (
        'mac_support', 1, target_date, 'global', 'all',
        0, 0, 0, 0, 0, 0, 0, 0, 0, transaction_timestamp()
    ) ON CONFLICT (metric_key, metric_version, fact_date, dimension_key, dimension_value) DO NOTHING;

    IF EXISTS (
        SELECT 1 FROM public.gfg_metric_daily daily
        WHERE daily.metric_key = 'mac_support' AND daily.metric_version = 1
          AND daily.fact_date = target_date
          AND (daily.population_count <> daily.eligible_count + daily.not_applicable_count
            OR daily.eligible_count <> daily.positive_count + daily.negative_count
               + daily.stale_count + daily.not_probed_count + daily.probe_failed_count + daily.unknown_count)
    ) THEN
        RAISE EXCEPTION 'Game metric count invariant failed for mac_support/1 on %', target_date;
    END IF;
    RETURN entity_rows;
END;
$function$;
-- +goose StatementEnd

-- +goose Down

DELETE FROM public.gfg_metric_daily WHERE metric_key = 'mac_support' AND metric_version = 1;
DELETE FROM public.gfg_metric_entity_daily WHERE metric_key = 'mac_support' AND metric_version = 1;
DELETE FROM public.gfg_metric_checkpoints WHERE metric_key = 'mac_support' AND metric_version = 1;
DELETE FROM public.gfg_metric_registry WHERE metric_key = 'mac_support' AND metric_version = 1;
DROP FUNCTION public.gfg_project_mac_metric_day(date);
