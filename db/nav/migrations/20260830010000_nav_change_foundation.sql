-- V3-P0.4 Nav Change Intelligence Foundation.
-- +goose Up

CREATE TABLE public.gfn_change_registry (
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
    CONSTRAINT gfn_change_registry_key_check CHECK (btrim(detector_key) <> ''),
    CONSTRAINT gfn_change_registry_version_check CHECK (detector_version > 0),
    CONSTRAINT gfn_change_registry_source_kind_check CHECK (source_kind IN ('metric', 'fact', 'domain_history', 'effective_period')),
    CONSTRAINT gfn_change_registry_source_shape_check CHECK (array_ndims(source_contracts) = 1 AND cardinality(source_contracts) > 0),
    CONSTRAINT gfn_change_registry_policy_check CHECK (btrim(detection_policy) <> '' AND btrim(watermark_policy) <> ''),
    CONSTRAINT gfn_change_registry_codes_shape_check CHECK (array_ndims(event_codes) = 1 AND cardinality(event_codes) > 0),
    CONSTRAINT gfn_change_registry_grain_check CHECK (processing_grain = 'day'),
    CONSTRAINT gfn_change_registry_status_check CHECK (status IN ('active', 'retired')),
    CONSTRAINT gfn_change_registry_status_shape_check CHECK (
        (status = 'active' AND retired_at IS NULL) OR (status = 'retired' AND retired_at IS NOT NULL)
    )
);

CREATE UNIQUE INDEX uq_gfn_change_registry_active_key
    ON public.gfn_change_registry(detector_key) WHERE status = 'active';

CREATE TABLE public.gfn_change_events (
    event_key text PRIMARY KEY,
    detector_key text NOT NULL,
    detector_version integer NOT NULL,
    site_id bigint NOT NULL,
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
    CONSTRAINT fk_gfn_change_event_registry FOREIGN KEY (detector_key, detector_version)
        REFERENCES public.gfn_change_registry(detector_key, detector_version) ON DELETE RESTRICT,
    CONSTRAINT uq_gfn_change_event_source UNIQUE (detector_key, detector_version, source_event_key),
    CONSTRAINT gfn_change_event_text_shape_check CHECK (
        btrim(event_key) <> '' AND btrim(event_code) <> '' AND btrim(scope_kind) <> ''
        AND btrim(scope_key) <> '' AND btrim(source_event_key) <> ''
        AND btrim(source_before_key) <> '' AND btrim(source_after_key) <> ''
    ),
    CONSTRAINT gfn_change_event_time_basis_check CHECK (time_basis IN ('effective', 'observed', 'day')),
    CONSTRAINT gfn_change_event_time_shape_check CHECK (
        (time_basis IN ('effective', 'observed') AND event_at IS NOT NULL)
        OR (time_basis = 'day' AND event_at IS NULL)
    ),
    CONSTRAINT gfn_change_event_json_shape_check CHECK (
        jsonb_typeof(old_value) = 'object' AND jsonb_typeof(new_value) = 'object'
        AND jsonb_typeof(source_versions) = 'object'
    )
);

CREATE INDEX idx_gfn_change_events_detector_date
    ON public.gfn_change_events(detector_key, detector_version, projection_date DESC);
CREATE INDEX idx_gfn_change_events_site_date
    ON public.gfn_change_events(site_id, projection_date DESC);
CREATE INDEX idx_gfn_change_events_code_date
    ON public.gfn_change_events(event_code, projection_date DESC);
CREATE INDEX idx_gfn_change_events_scope_date
    ON public.gfn_change_events(scope_kind, scope_key, projection_date DESC);
CREATE INDEX idx_gfn_change_events_event_at
    ON public.gfn_change_events(event_at DESC) WHERE event_at IS NOT NULL;

CREATE TABLE public.gfn_change_checkpoints (
    detector_key text NOT NULL,
    detector_version integer NOT NULL,
    source_start_date date,
    processed_through date,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (detector_key, detector_version),
    CONSTRAINT fk_gfn_change_checkpoint_registry FOREIGN KEY (detector_key, detector_version)
        REFERENCES public.gfn_change_registry(detector_key, detector_version) ON DELETE RESTRICT,
    CONSTRAINT gfn_change_checkpoint_range_check CHECK (
        source_start_date IS NULL OR processed_through IS NULL OR processed_through >= source_start_date
    )
);

INSERT INTO public.gfn_change_registry (
    detector_key, detector_version, source_kind, source_contracts,
    detection_policy, watermark_policy, event_codes, processing_grain, status, description
) VALUES
    ('ipv6_transition', 1, 'metric', ARRAY['ipv6_adoption/1', 'gfn_site_daily'],
     'metric_semantic_transition_v1', 'metric_checkpoint_v1',
     ARRAY['ipv6_enabled', 'ipv6_disabled'], 'day', 'active',
     'Semantic IPv6 transitions within one Primary Target identity.'),
    ('tls13_transition', 1, 'metric', ARRAY['tls13_adoption/1', 'gfn_site_daily'],
     'metric_semantic_transition_v1', 'metric_checkpoint_v1',
     ARRAY['tls13_enabled', 'tls13_disabled'], 'day', 'active',
     'Semantic TLS 1.3 transitions within one Primary Target identity.'),
    ('security_txt_transition', 1, 'metric', ARRAY['security_txt_adoption/1', 'gfn_site_daily'],
     'metric_semantic_transition_v1', 'metric_checkpoint_v1',
     ARRAY['security_txt_added', 'security_txt_removed'], 'day', 'active',
     'Semantic security.txt transitions within one Primary Target identity.'),
    ('primary_target_transition', 1, 'effective_period', ARRAY['gfn_site_primary_target_periods', 'gfn_target_tracking_periods'],
     'primary_continuous_replacement_v1', 'closed_day_v1',
     ARRAY['primary_target_changed'], 'day', 'active',
     'Continuous effective-dated Primary Target replacements.'),
    ('tls_certificate_transition', 1, 'fact', ARRAY['gfn_site_target_daily'],
     'tls_certificate_semantic_memory_v1', 'fact_checkpoint_v1',
     ARRAY['tls_certificate_changed'], 'day', 'active',
     'TLS certificate fingerprint transitions within one target tracking identity.');

INSERT INTO public.gfn_change_checkpoints(detector_key, detector_version, source_start_date)
SELECT registry.detector_key,
       registry.detector_version,
       CASE registry.detector_key
           WHEN 'ipv6_transition' THEN (SELECT source_start_date FROM public.gfn_metric_checkpoints WHERE metric_key = 'ipv6_adoption' AND metric_version = 1)
           WHEN 'tls13_transition' THEN (SELECT source_start_date FROM public.gfn_metric_checkpoints WHERE metric_key = 'tls13_adoption' AND metric_version = 1)
           WHEN 'security_txt_transition' THEN (SELECT source_start_date FROM public.gfn_metric_checkpoints WHERE metric_key = 'security_txt_adoption' AND metric_version = 1)
           WHEN 'tls_certificate_transition' THEN (SELECT source_start_date FROM public.gfn_fact_rollup_checkpoints WHERE pipeline_key = 'nav.target_facts')
           WHEN 'primary_target_transition' THEN (
               SELECT COALESCE(
                   min((effective_from AT TIME ZONE 'UTC')::date),
                   (transaction_timestamp() AT TIME ZONE 'UTC')::date
               ) FROM public.gfn_site_primary_target_periods
           )
       END
FROM public.gfn_change_registry registry;

-- +goose StatementBegin
CREATE FUNCTION public.gfn_project_change_day(
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
        'ipv6_transition', 'tls13_transition', 'security_txt_transition',
        'primary_target_transition', 'tls_certificate_transition'
    ) THEN
        RAISE EXCEPTION 'Nav change detector is not compiled for %/%', target_detector_key, target_detector_version;
    END IF;

    DELETE FROM public.gfn_change_events
    WHERE detector_key = target_detector_key
      AND detector_version = target_detector_version
      AND projection_date = target_date;

    IF target_detector_key IN ('ipv6_transition', 'tls13_transition', 'security_txt_transition') THEN
        source_metric_key := CASE target_detector_key
            WHEN 'ipv6_transition' THEN 'ipv6_adoption'
            WHEN 'tls13_transition' THEN 'tls13_adoption'
            WHEN 'security_txt_transition' THEN 'security_txt_adoption'
        END;

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
                   WHEN 'ipv6_transition' THEN CASE WHEN current_metric.state = 'positive' THEN 'ipv6_enabled' ELSE 'ipv6_disabled' END
                   WHEN 'tls13_transition' THEN CASE WHEN current_metric.state = 'positive' THEN 'tls13_enabled' ELSE 'tls13_disabled' END
                   WHEN 'security_txt_transition' THEN CASE WHEN current_metric.state = 'positive' THEN 'security_txt_added' ELSE 'security_txt_removed' END
               END,
               'global', 'all',
               jsonb_build_object('state', previous_metric.state, 'reason_code', previous_metric.reason_code),
               jsonb_build_object('state', current_metric.state, 'reason_code', current_metric.reason_code),
               event_identity.source_event_key,
               source_metric_key || '/1/' || previous_metric.fact_date || '/' || current_metric.site_id,
               source_metric_key || '/1/' || current_metric.fact_date || '/' || current_metric.site_id,
               previous_metric.source_observed_at, current_metric.source_observed_at,
               jsonb_build_object(
                   'metric_key', source_metric_key, 'metric_version', 1,
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
                   || current_metric.site_id || '/' || previous_metric.fact_date || '/' || current_metric.fact_date AS source_event_key
        ) event_identity
        WHERE current_metric.metric_key = source_metric_key
          AND current_metric.metric_version = 1
          AND current_metric.fact_date = target_date
          AND current_fact.primary_target_tracking_period_id IS NOT NULL
          AND current_metric.state IN ('positive', 'negative')
          AND current_metric.state <> previous_metric.state;

    ELSIF target_detector_key = 'primary_target_transition' THEN
        INSERT INTO public.gfn_change_events (
            event_key, detector_key, detector_version, site_id, projection_date,
            event_at, time_basis, event_code, scope_kind, scope_key,
            old_value, new_value, source_event_key, source_before_key, source_after_key,
            source_before_at, source_after_at, source_versions, materialized_at
        )
        SELECT 'primary_target_transition/1/period/' || current_period.id,
               target_detector_key, 1, current_period.site_id, target_date,
               current_period.effective_from, 'effective', 'primary_target_changed', 'global', 'all',
               jsonb_build_object('target_tracking_period_id', previous_period.target_tracking_period_id,
                                  'target', previous_target.target, 'basis', previous_period.basis),
               jsonb_build_object('target_tracking_period_id', current_period.target_tracking_period_id,
                                  'target', current_target.target, 'basis', current_period.basis),
               'period/' || current_period.id, 'period/' || previous_period.id, 'period/' || current_period.id,
               previous_period.effective_from, current_period.effective_from,
               jsonb_build_object('period_contract', 1), transaction_timestamp()
        FROM public.gfn_site_primary_target_periods current_period
        JOIN public.gfn_site_primary_target_periods previous_period
          ON previous_period.site_id = current_period.site_id
         AND previous_period.effective_until = current_period.effective_from
        JOIN public.gfn_target_tracking_periods previous_target
          ON previous_target.id = previous_period.target_tracking_period_id
        JOIN public.gfn_target_tracking_periods current_target
          ON current_target.id = current_period.target_tracking_period_id
        WHERE (current_period.effective_from AT TIME ZONE 'UTC')::date = target_date
          AND previous_period.target_tracking_period_id <> current_period.target_tracking_period_id;

    ELSE
        WITH current_certificate AS (
            SELECT fact.*
            FROM public.gfn_site_target_daily fact
            WHERE fact.fact_date = target_date
              AND fact.finalized_at IS NOT NULL
              AND NULLIF(btrim(fact.tls_cert_fingerprint_sha256), '') IS NOT NULL
        ), paired AS (
            SELECT current_certificate.*,
                   previous.fact_date AS before_date,
                   previous.tls_state_observed_at AS before_observed_at,
                   previous.tls_cert_fingerprint_sha256 AS before_fingerprint,
                   previous.tls_cert_spki_sha256 AS before_spki,
                   previous.tls_cert_issuer AS before_issuer,
                   previous.tls_cert_not_before AS before_not_before,
                   previous.tls_cert_not_after AS before_not_after,
                   previous.tls_cert_verified AS before_verified,
                   previous.projection_version AS before_projection_version
            FROM current_certificate
            JOIN LATERAL (
                SELECT candidate.*
                FROM public.gfn_site_target_daily candidate
                WHERE candidate.target_tracking_period_id = current_certificate.target_tracking_period_id
                  AND candidate.fact_date < current_certificate.fact_date
                  AND candidate.finalized_at IS NOT NULL
                  AND NULLIF(btrim(candidate.tls_cert_fingerprint_sha256), '') IS NOT NULL
                ORDER BY candidate.fact_date DESC
                LIMIT 1
            ) previous ON true
        )
        INSERT INTO public.gfn_change_events (
            event_key, detector_key, detector_version, site_id, projection_date,
            event_at, time_basis, event_code, scope_kind, scope_key,
            old_value, new_value, source_event_key, source_before_key, source_after_key,
            source_before_at, source_after_at, source_versions, materialized_at
        )
        SELECT 'tls_certificate_transition/1/' || target_tracking_period_id || '/' || fact_date,
               target_detector_key, 1, site_id, target_date,
               tls_state_observed_at,
               CASE WHEN tls_state_observed_at IS NULL THEN 'day' ELSE 'observed' END,
               'tls_certificate_changed', 'target', target_tracking_period_id::text,
               jsonb_build_object('fingerprint_sha256', before_fingerprint, 'spki_sha256', before_spki,
                                  'issuer', before_issuer, 'not_before', before_not_before,
                                  'not_after', before_not_after, 'verified', before_verified),
               jsonb_build_object('fingerprint_sha256', tls_cert_fingerprint_sha256,
                                  'spki_sha256', tls_cert_spki_sha256, 'issuer', tls_cert_issuer,
                                  'not_before', tls_cert_not_before, 'not_after', tls_cert_not_after,
                                  'verified', tls_cert_verified),
               target_tracking_period_id || '/' || fact_date,
               'certificate/' || target_tracking_period_id || '/' || before_date,
               'certificate/' || target_tracking_period_id || '/' || fact_date,
               before_observed_at, tls_state_observed_at,
               jsonb_build_object('before_projection_version', before_projection_version,
                                  'after_projection_version', projection_version,
                                  'target_tracking_period_id', target_tracking_period_id),
               transaction_timestamp()
        FROM paired
        WHERE before_fingerprint <> tls_cert_fingerprint_sha256;
    END IF;

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

COMMENT ON TABLE public.gfn_change_registry IS
    'Goose-owned versioned Nav change detector contracts; Runtime and Admin are read-only.';
COMMENT ON TABLE public.gfn_change_events IS
    'Deterministic canonical Nav change events derived only from facts, metrics, and effective periods.';
COMMENT ON TABLE public.gfn_change_checkpoints IS
    'Independent ordered checkpoints for each Nav detector version.';

-- +goose Down

DROP FUNCTION public.gfn_project_change_day(text, integer, date);
DROP TABLE public.gfn_change_checkpoints;
DROP TABLE public.gfn_change_events;
DROP TABLE public.gfn_change_registry;
