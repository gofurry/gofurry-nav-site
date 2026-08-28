-- V3-P0.2 Nav historical fact and checkpoint tables.
-- +goose Up

CREATE TABLE public.gfn_site_target_protocol_daily (
    target_tracking_period_id bigint NOT NULL
        REFERENCES public.gfn_target_tracking_periods(id) ON DELETE RESTRICT,
    site_id bigint NOT NULL,
    collector_domain_id bigint,
    target text NOT NULL,
    protocol text NOT NULL,
    fact_date date NOT NULL,
    expected_count integer,
    attempted_count integer NOT NULL,
    success_count integer NOT NULL,
    partial_count integer NOT NULL,
    failure_count integer NOT NULL,
    skipped_count integer,
    missed_count integer,
    canceled_count integer,
    unattempted_count integer,
    failure_kind_counts jsonb NOT NULL DEFAULT '{}'::jsonb,
    quality_basis text NOT NULL,
    latest_scheduled_status text,
    latest_scheduled_at timestamp with time zone,
    latest_observation_status text,
    latest_observation_at timestamp with time zone,
    known_state_observed_at timestamp with time zone,
    known_state jsonb,
    avg_duration_ms double precision,
    p95_duration_ms double precision,
    projection_version integer NOT NULL,
    finalized_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (target_tracking_period_id, protocol, fact_date),
    CONSTRAINT gfn_site_target_protocol_daily_protocol_check CHECK (
        protocol IN ('ping', 'http', 'dns', 'rdap', 'robots', 'security_txt',
                     'llms_txt', 'page_assets', 'port_check', 'waf_canary')
    ),
    CONSTRAINT gfn_site_target_protocol_daily_counts_check CHECK (
        attempted_count >= 0
        AND success_count >= 0
        AND partial_count >= 0
        AND failure_count >= 0
        AND attempted_count = success_count + partial_count + failure_count
    ),
    CONSTRAINT gfn_site_target_protocol_daily_quality_check CHECK (
        (quality_basis = 'legacy_observed_only'
            AND expected_count IS NULL
            AND skipped_count IS NULL
            AND missed_count IS NULL
            AND canceled_count IS NULL
            AND unattempted_count IS NULL)
        OR
        (quality_basis = 'acquisition_ledger'
            AND expected_count >= 0
            AND skipped_count >= 0
            AND missed_count >= 0
            AND canceled_count >= 0
            AND unattempted_count >= 0
            AND expected_count = attempted_count + skipped_count
                + missed_count + canceled_count + unattempted_count)
    ),
    CONSTRAINT gfn_site_target_protocol_daily_failure_json_check
        CHECK (jsonb_typeof(failure_kind_counts) = 'object'),
    CONSTRAINT gfn_site_target_protocol_daily_known_state_check
        CHECK (known_state IS NULL OR jsonb_typeof(known_state) = 'object'),
    CONSTRAINT gfn_site_target_protocol_daily_projection_check
        CHECK (projection_version > 0)
);

CREATE INDEX idx_gfn_site_target_protocol_daily_site_date
    ON public.gfn_site_target_protocol_daily (site_id, fact_date DESC, protocol);
CREATE INDEX idx_gfn_site_target_protocol_daily_target_date
    ON public.gfn_site_target_protocol_daily (target, fact_date DESC, protocol);

CREATE TABLE public.gfn_site_target_daily (
    target_tracking_period_id bigint NOT NULL
        REFERENCES public.gfn_target_tracking_periods(id) ON DELETE RESTRICT,
    site_id bigint NOT NULL,
    collector_domain_id bigint,
    target text NOT NULL,
    fact_date date NOT NULL,
    snapshot_at timestamp with time zone NOT NULL,
    tracked_at_end boolean NOT NULL,
    http_probe_status text,
    http_probe_observed_at timestamp with time zone,
    http_state_observed_at timestamp with time zone,
    http_status_code integer,
    http_response_time_ms bigint,
    http_ttfb_ms bigint,
    http_protocol text,
    http_server text,
    http_remote_ip text,
    http_final_url text,
    http_canonical_url text,
    http_security_headers jsonb,
    tls_state_observed_at timestamp with time zone,
    tls_handshake text,
    tls_version text,
    tls_cipher_suite text,
    tls_cert_verified boolean,
    tls_verify_error_category text,
    tls_cert_not_before timestamp with time zone,
    tls_cert_not_after timestamp with time zone,
    tls_cert_issuer text,
    tls_cert_fingerprint_sha256 text,
    tls_cert_spki_sha256 text,
    tls_cert_dns_names text[],
    dns_probe_status text,
    dns_probe_observed_at timestamp with time zone,
    dns_state_observed_at timestamp with time zone,
    dns_has_a boolean,
    dns_has_aaaa boolean,
    dns_ipv4_count integer,
    dns_ipv6_count integer,
    dns_a_records text[],
    dns_aaaa_records text[],
    dns_cname_terminal text,
    dns_cname_depth integer,
    dns_ns_hosts text[],
    dns_mx_hosts text[],
    dns_min_ttl integer,
    dns_max_ttl integer,
    dns_risk_flags text[],
    ping_probe_status text,
    ping_probe_observed_at timestamp with time zone,
    ping_state_observed_at timestamp with time zone,
    ping_avg_rtt_ms double precision,
    ping_min_rtt_ms double precision,
    ping_max_rtt_ms double precision,
    ping_loss_rate double precision,
    ping_jitter_ms double precision,
    ping_selected_ip text,
    ping_ip_family text,
    ping_icmp_blocked_suspected boolean,
    projection_version integer NOT NULL,
    finalized_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (target_tracking_period_id, fact_date),
    CONSTRAINT gfn_site_target_daily_target_check CHECK (btrim(target) <> ''),
    CONSTRAINT gfn_site_target_daily_http_security_headers_check
        CHECK (http_security_headers IS NULL OR jsonb_typeof(http_security_headers) = 'object'),
    CONSTRAINT gfn_site_target_daily_projection_check CHECK (projection_version > 0)
);

CREATE INDEX idx_gfn_site_target_daily_site_date
    ON public.gfn_site_target_daily (site_id, fact_date DESC);
CREATE INDEX idx_gfn_site_target_daily_target_date
    ON public.gfn_site_target_daily (target, fact_date DESC);

CREATE TABLE public.gfn_site_daily (
    site_id bigint NOT NULL,
    fact_date date NOT NULL,
    snapshot_at timestamp with time zone NOT NULL,
    tracked_at_end boolean NOT NULL,
    name text NOT NULL,
    name_en text NOT NULL,
    site_country text,
    nsfw boolean,
    welfare boolean,
    view_count bigint NOT NULL,
    group_ids bigint[] NOT NULL,
    primary_target_tracking_period_id bigint
        REFERENCES public.gfn_target_tracking_periods(id) ON DELETE RESTRICT,
    primary_target text,
    primary_basis text,
    active_target_count integer NOT NULL,
    projection_version integer NOT NULL,
    finalized_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (site_id, fact_date),
    CONSTRAINT gfn_site_daily_primary_shape_check CHECK (
        (primary_target_tracking_period_id IS NULL
            AND primary_target IS NULL AND primary_basis IS NULL)
        OR
        (primary_target_tracking_period_id IS NOT NULL
            AND btrim(primary_target) <> ''
            AND primary_basis IN ('explicit', 'single_target_inferred', 'deterministic_fallback', 'legacy_backfill'))
    ),
    CONSTRAINT gfn_site_daily_active_target_count_check CHECK (active_target_count >= 0),
    CONSTRAINT gfn_site_daily_projection_check CHECK (projection_version > 0)
);

CREATE INDEX idx_gfn_site_daily_date ON public.gfn_site_daily (fact_date DESC, site_id);
CREATE INDEX idx_gfn_site_daily_primary
    ON public.gfn_site_daily (primary_target_tracking_period_id, fact_date DESC)
    WHERE primary_target_tracking_period_id IS NOT NULL;

CREATE TABLE public.gfn_fact_rollup_checkpoints (
    pipeline_key text PRIMARY KEY,
    projection_version integer NOT NULL,
    source_start_date date,
    processed_through date,
    quality_cutover_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT gfn_fact_rollup_checkpoints_pipeline_check
        CHECK (pipeline_key IN ('nav.target_facts', 'nav.site_facts')),
    CONSTRAINT gfn_fact_rollup_checkpoints_projection_check
        CHECK (projection_version > 0),
    CONSTRAINT gfn_fact_rollup_checkpoints_range_check
        CHECK (source_start_date IS NULL OR processed_through IS NULL OR processed_through >= source_start_date)
);

INSERT INTO public.gfn_fact_rollup_checkpoints (
    pipeline_key, projection_version, source_start_date, quality_cutover_at
)
SELECT 'nav.target_facts',
       1,
       LEAST(
           COALESCE(
               (SELECT min(observed_at AT TIME ZONE 'UTC')::date FROM public.gfn_collector_observation),
               (transaction_timestamp() AT TIME ZONE 'UTC')::date
           ),
           (transaction_timestamp() AT TIME ZONE 'UTC')::date
       ),
       date_trunc('day', transaction_timestamp() AT TIME ZONE 'UTC')
           AT TIME ZONE 'UTC' + interval '1 day'
UNION ALL
SELECT 'nav.site_facts',
       1,
       (transaction_timestamp() AT TIME ZONE 'UTC')::date,
       date_trunc('day', transaction_timestamp() AT TIME ZONE 'UTC')
           AT TIME ZONE 'UTC' + interval '1 day';

-- Shared current-day Site dimension projector used by Admin lifecycle writes.
-- A NULL finalized_at denotes only this mutable current-day marker.
-- +goose StatementBegin
CREATE FUNCTION public.gfn_refresh_current_site_daily(target_site_id bigint)
RETURNS void
LANGUAGE plpgsql
AS $function$
BEGIN
    WITH current_primary AS (
        SELECT primary_period.site_id,
               target_period.id AS target_tracking_period_id,
               target_period.target,
               primary_period.basis
        FROM public.gfn_site_primary_target_periods primary_period
        JOIN public.gfn_target_tracking_periods target_period
          ON target_period.id = primary_period.target_tracking_period_id
        WHERE primary_period.site_id = target_site_id
          AND primary_period.effective_until IS NULL
    ), current_targets AS (
        SELECT site_id, count(*)::integer AS active_target_count
        FROM public.gfn_target_tracking_periods
        WHERE site_id = target_site_id AND tracked_until IS NULL
        GROUP BY site_id
    ), site_groups AS (
        SELECT site_id, array_agg(DISTINCT group_id ORDER BY group_id) AS group_ids
        FROM public.gfn_site_group_map
        WHERE site_id = target_site_id
        GROUP BY site_id
    )
    INSERT INTO public.gfn_site_daily (
        site_id, fact_date, snapshot_at, tracked_at_end, name, name_en,
        site_country, nsfw, welfare, view_count, group_ids,
        primary_target_tracking_period_id, primary_target, primary_basis,
        active_target_count, projection_version, finalized_at,
        created_at, updated_at
    )
    SELECT s.id,
           (transaction_timestamp() AT TIME ZONE 'UTC')::date,
           transaction_timestamp(),
           NOT s.deleted,
           s.name::text,
           s.name_en::text,
           NULLIF(btrim(s.country), ''),
           CASE lower(btrim(s.nsfw))
               WHEN '1' THEN true WHEN 'true' THEN true WHEN 'yes' THEN true
               WHEN '0' THEN false WHEN 'false' THEN false WHEN 'no' THEN false
               ELSE NULL
           END,
           CASE lower(btrim(s.welfare))
               WHEN '1' THEN true WHEN 'true' THEN true WHEN 'yes' THEN true
               WHEN '0' THEN false WHEN 'false' THEN false WHEN 'no' THEN false
               ELSE NULL
           END,
           s.view_count,
           COALESCE(groups.group_ids, ARRAY[]::bigint[]),
           primary_target.target_tracking_period_id,
           primary_target.target,
           primary_target.basis,
           COALESCE(targets.active_target_count, 0),
           1,
           NULL,
           transaction_timestamp(),
           transaction_timestamp()
    FROM public.gfn_site s
    LEFT JOIN current_primary primary_target ON primary_target.site_id = s.id
    LEFT JOIN current_targets targets ON targets.site_id = s.id
    LEFT JOIN site_groups groups ON groups.site_id = s.id
    WHERE s.id = target_site_id
    ON CONFLICT (site_id, fact_date) DO UPDATE
    SET snapshot_at = EXCLUDED.snapshot_at,
        tracked_at_end = EXCLUDED.tracked_at_end,
        name = EXCLUDED.name,
        name_en = EXCLUDED.name_en,
        site_country = EXCLUDED.site_country,
        nsfw = EXCLUDED.nsfw,
        welfare = EXCLUDED.welfare,
        view_count = EXCLUDED.view_count,
        group_ids = EXCLUDED.group_ids,
        primary_target_tracking_period_id = EXCLUDED.primary_target_tracking_period_id,
        primary_target = EXCLUDED.primary_target,
        primary_basis = EXCLUDED.primary_basis,
        active_target_count = EXCLUDED.active_target_count,
        projection_version = EXCLUDED.projection_version,
        updated_at = transaction_timestamp()
    WHERE public.gfn_site_daily.finalized_at IS NULL;
END;
$function$;
-- +goose StatementEnd

SELECT public.gfn_refresh_current_site_daily(id)
FROM public.gfn_site;

COMMENT ON TABLE public.gfn_site_target_protocol_daily IS
    'Scheduled protocol quality, latest observation outcome, and whitelisted last-known state.';
COMMENT ON TABLE public.gfn_site_target_daily IS
    'Typed compact HTTP/TLS/DNS/Ping historical target fact.';
COMMENT ON TABLE public.gfn_site_daily IS
    'Historical Site dimensions and effective-dated Primary Target selection. finalized_at is NULL only for a current-day write-through marker.';
COMMENT ON TABLE public.gfn_fact_rollup_checkpoints IS
    'Ordered singleton checkpoints for Nav fact pipelines.';

-- +goose Down

DROP FUNCTION IF EXISTS public.gfn_refresh_current_site_daily(bigint);
DROP TABLE public.gfn_fact_rollup_checkpoints;
DROP TABLE public.gfn_site_daily;
DROP TABLE public.gfn_site_target_daily;
DROP TABLE public.gfn_site_target_protocol_daily;
