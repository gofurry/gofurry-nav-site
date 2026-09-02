-- V3-P2.2 Mac support Change contract and projector extension.
-- +goose Up

INSERT INTO public.gfg_change_registry (
    detector_key, detector_version, source_kind, source_contracts,
    detection_policy, watermark_policy, event_codes, processing_grain,
    status, description
) VALUES (
    'mac_support_transition', 1, 'metric',
    ARRAY['mac_support/1', 'gfg_game_daily']::text[],
    'metric_semantic_transition_v1', 'metric_checkpoint_v1',
    ARRAY['mac_support_added', 'mac_support_removed']::text[],
    'day', 'active', 'Reliable macOS support transitions within one tracking period.'
);

INSERT INTO public.gfg_change_checkpoints (
    detector_key, detector_version, source_start_date
)
SELECT 'mac_support_transition', 1, source_start_date
FROM public.gfg_metric_checkpoints
WHERE metric_key = 'mac_support' AND metric_version = 1;

-- +goose StatementBegin
CREATE FUNCTION public.gfg_project_mac_change_day(target_date date) RETURNS bigint
LANGUAGE plpgsql
AS $function$
DECLARE
    inserted_rows bigint;
BEGIN
    DELETE FROM public.gfg_change_events
    WHERE detector_key = 'mac_support_transition' AND detector_version = 1
      AND projection_date = target_date;

    INSERT INTO public.gfg_change_events (
        event_key, detector_key, detector_version, game_id, projection_date,
        event_at, time_basis, event_code, scope_kind, scope_key,
        old_value, new_value, source_event_key, source_before_key, source_after_key,
        source_before_at, source_after_at, source_versions, materialized_at
    )
    SELECT 'mac_support_transition/1/' || current_fact.tracking_period_id || '/'
               || current_metric.game_id || '/' || previous_metric.fact_date || '/' || current_metric.fact_date,
           'mac_support_transition', 1, current_metric.game_id, target_date,
           current_metric.source_observed_at,
           CASE WHEN current_metric.source_observed_at IS NULL THEN 'day' ELSE 'observed' END,
           CASE WHEN current_metric.state = 'positive' THEN 'mac_support_added' ELSE 'mac_support_removed' END,
           'global', 'all',
           jsonb_build_object('state', previous_metric.state, 'reason_code', previous_metric.reason_code),
           jsonb_build_object('state', current_metric.state, 'reason_code', current_metric.reason_code),
           'mac_support/1/' || current_fact.tracking_period_id || '/' || current_metric.game_id || '/'
               || previous_metric.fact_date || '/' || current_metric.fact_date,
           'mac_support/1/' || previous_metric.fact_date || '/' || current_metric.game_id,
           'mac_support/1/' || current_metric.fact_date || '/' || current_metric.game_id,
           previous_metric.source_observed_at, current_metric.source_observed_at,
           jsonb_build_object(
               'metric_key', 'mac_support', 'metric_version', 1,
               'before_fact_projection_versions', previous_metric.source_projection_versions,
               'after_fact_projection_versions', current_metric.source_projection_versions,
               'tracking_period_id', current_fact.tracking_period_id
           ), transaction_timestamp()
    FROM public.gfg_metric_entity_daily current_metric
    JOIN public.gfg_game_daily current_fact
      ON current_fact.game_id = current_metric.game_id
     AND current_fact.fact_date = current_metric.fact_date
     AND current_fact.finalized_at IS NOT NULL
    JOIN LATERAL (
        SELECT candidate.*
        FROM public.gfg_metric_entity_daily candidate
        JOIN public.gfg_game_daily previous_fact
          ON previous_fact.game_id = candidate.game_id
         AND previous_fact.fact_date = candidate.fact_date
         AND previous_fact.finalized_at IS NOT NULL
        WHERE candidate.metric_key = 'mac_support' AND candidate.metric_version = 1
          AND candidate.game_id = current_metric.game_id
          AND candidate.fact_date = current_metric.fact_date - 1
          AND previous_fact.tracking_period_id = current_fact.tracking_period_id
        ORDER BY candidate.fact_date DESC
        LIMIT 1
    ) previous_metric ON true
    WHERE current_metric.metric_key = 'mac_support' AND current_metric.metric_version = 1
      AND current_metric.fact_date = target_date
      AND current_metric.state IN ('positive', 'negative')
      AND previous_metric.state IN ('positive', 'negative')
      AND current_metric.state <> previous_metric.state;

    GET DIAGNOSTICS inserted_rows = ROW_COUNT;
    RETURN inserted_rows;
END;
$function$;
-- +goose StatementEnd

-- +goose Down

DELETE FROM public.gfg_change_events WHERE detector_key = 'mac_support_transition' AND detector_version = 1;
DELETE FROM public.gfg_change_checkpoints WHERE detector_key = 'mac_support_transition' AND detector_version = 1;
DELETE FROM public.gfg_change_registry WHERE detector_key = 'mac_support_transition' AND detector_version = 1;
DROP FUNCTION public.gfg_project_mac_change_day(date);
