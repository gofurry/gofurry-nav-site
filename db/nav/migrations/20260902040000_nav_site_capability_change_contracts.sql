-- V3-P2.3 Site HTTP/security capability and certificate-verification Change contracts.
-- +goose Up

INSERT INTO public.gfn_change_registry (
    detector_key, detector_version, source_kind, source_contracts,
    detection_policy, watermark_policy, event_codes, processing_grain,
    status, description
) VALUES
    ('http2_transition', 1, 'metric', ARRAY['http2_adoption/1', 'gfn_site_daily'],
     'metric_semantic_transition_v1', 'metric_checkpoint_v1',
     ARRAY['http2_enabled', 'http2_disabled'], 'day', 'active',
     'Semantic HTTP/2 transitions within one Primary Target identity.'),
    ('hsts_transition', 1, 'metric', ARRAY['hsts_adoption/1', 'gfn_site_daily'],
     'metric_semantic_transition_v1', 'metric_checkpoint_v1',
     ARRAY['hsts_added', 'hsts_removed'], 'day', 'active',
     'Semantic HSTS-presence transitions within one Primary Target identity.'),
    ('csp_transition', 1, 'metric', ARRAY['csp_adoption/1', 'gfn_site_daily'],
     'metric_semantic_transition_v1', 'metric_checkpoint_v1',
     ARRAY['csp_added', 'csp_removed'], 'day', 'active',
     'Semantic enforcement-CSP-presence transitions within one Primary Target identity.'),
    ('tls_certificate_verification_transition', 1, 'metric',
     ARRAY['tls_certificate_verification/1', 'gfn_site_daily'],
     'metric_semantic_transition_v1', 'metric_checkpoint_v1',
     ARRAY['tls_certificate_verification_failed', 'tls_certificate_verification_restored'],
     'day', 'active',
     'Semantic TLS certificate-verification transitions within one Primary Target identity.');

INSERT INTO public.gfn_change_checkpoints (
    detector_key, detector_version, source_start_date
)
SELECT registry.detector_key,
       registry.detector_version,
       metric.source_start_date
FROM public.gfn_change_registry registry
JOIN public.gfn_metric_checkpoints metric
  ON metric.metric_key = CASE registry.detector_key
      WHEN 'http2_transition' THEN 'http2_adoption'
      WHEN 'hsts_transition' THEN 'hsts_adoption'
      WHEN 'csp_transition' THEN 'csp_adoption'
      WHEN 'tls_certificate_verification_transition' THEN 'tls_certificate_verification'
  END
 AND metric.metric_version = registry.detector_version
WHERE registry.detector_key IN (
    'http2_transition', 'hsts_transition', 'csp_transition',
    'tls_certificate_verification_transition'
)
  AND registry.detector_version = 1;

-- +goose StatementBegin
CREATE FUNCTION public.gfn_project_site_capability_change_day(
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
        'http2_transition', 'hsts_transition', 'csp_transition',
        'tls_certificate_verification_transition'
    ) THEN
        RAISE EXCEPTION 'Nav Site capability detector is not compiled for %/%',
            target_detector_key, target_detector_version;
    END IF;

    source_metric_key := CASE target_detector_key
        WHEN 'http2_transition' THEN 'http2_adoption'
        WHEN 'hsts_transition' THEN 'hsts_adoption'
        WHEN 'csp_transition' THEN 'csp_adoption'
        WHEN 'tls_certificate_verification_transition' THEN 'tls_certificate_verification'
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
    SELECT target_detector_key || '/1/' || event_identity.source_event_key,
           target_detector_key, 1, current_metric.site_id, target_date,
           current_metric.source_observed_at,
           CASE WHEN current_metric.source_observed_at IS NULL THEN 'day' ELSE 'observed' END,
           CASE target_detector_key
               WHEN 'http2_transition' THEN CASE
                   WHEN current_metric.state = 'positive' THEN 'http2_enabled' ELSE 'http2_disabled' END
               WHEN 'hsts_transition' THEN CASE
                   WHEN current_metric.state = 'positive' THEN 'hsts_added' ELSE 'hsts_removed' END
               WHEN 'csp_transition' THEN CASE
                   WHEN current_metric.state = 'positive' THEN 'csp_added' ELSE 'csp_removed' END
               WHEN 'tls_certificate_verification_transition' THEN CASE
                   WHEN current_metric.state = 'positive'
                       THEN 'tls_certificate_verification_restored'
                       ELSE 'tls_certificate_verification_failed'
                   END
           END,
           'global', 'all',
           jsonb_build_object('state', previous_metric.state, 'reason_code', previous_metric.reason_code),
           jsonb_build_object('state', current_metric.state, 'reason_code', current_metric.reason_code),
           event_identity.source_event_key,
           source_metric_key || '/1/' || previous_metric.fact_date || '/' || current_metric.site_id,
           source_metric_key || '/1/' || current_metric.fact_date || '/' || current_metric.site_id,
           previous_metric.source_observed_at, current_metric.source_observed_at,
           jsonb_build_object(
               'metric_key', source_metric_key,
               'metric_version', 1,
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
        SELECT source_metric_key || '/1/' || current_fact.primary_target_tracking_period_id || '/'
               || current_metric.site_id || '/' || previous_metric.fact_date || '/'
               || current_metric.fact_date AS source_event_key
    ) event_identity
    WHERE current_metric.metric_key = source_metric_key
      AND current_metric.metric_version = 1
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
        RAISE EXCEPTION 'Nav Site capability detector emitted an undeclared event code for %/% on %',
            target_detector_key, target_detector_version, target_date;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM public.gfn_change_events
        WHERE detector_key = target_detector_key
          AND detector_version = target_detector_version
          AND projection_date = target_date
          AND time_basis = 'observed'
          AND event_at IS DISTINCT FROM source_after_at
    ) THEN
        RAISE EXCEPTION 'Nav Site capability event timestamp differs from source_after_at for %/% on %',
            target_detector_key, target_detector_version, target_date;
    END IF;

    RETURN inserted_rows;
END;
$function$;
-- +goose StatementEnd

-- +goose Down

DROP FUNCTION public.gfn_project_site_capability_change_day(text, integer, date);
DELETE FROM public.gfn_change_events
WHERE detector_key IN (
    'http2_transition', 'hsts_transition', 'csp_transition',
    'tls_certificate_verification_transition'
)
  AND detector_version = 1;
DELETE FROM public.gfn_change_checkpoints
WHERE detector_key IN (
    'http2_transition', 'hsts_transition', 'csp_transition',
    'tls_certificate_verification_transition'
)
  AND detector_version = 1;
DELETE FROM public.gfn_change_registry
WHERE detector_key IN (
    'http2_transition', 'hsts_transition', 'csp_transition',
    'tls_certificate_verification_transition'
)
  AND detector_version = 1;
