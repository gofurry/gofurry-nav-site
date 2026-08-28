-- V3-P0.2 set-based Nav target and Site fact projectors.
-- +goose Up

-- +goose StatementBegin
CREATE FUNCTION public.gfn_jsonb_text_array(value jsonb)
RETURNS text[]
LANGUAGE sql
IMMUTABLE
PARALLEL SAFE
AS $function$
    SELECT CASE WHEN jsonb_typeof(value) = 'array'
                THEN ARRAY(SELECT jsonb_array_elements_text(value))
                ELSE NULL::text[]
           END;
$function$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.gfn_project_target_fact_day(target_date date)
RETURNS bigint
LANGUAGE plpgsql
AS $function$
DECLARE
    day_start timestamp with time zone := target_date::timestamp AT TIME ZONE 'UTC';
    day_end timestamp with time zone := (target_date + 1)::timestamp AT TIME ZONE 'UTC';
    projected_rows bigint;
BEGIN
    DELETE FROM public.gfn_site_target_daily WHERE fact_date = target_date;

    WITH eligible AS (
        SELECT period.*,
               LEAST(day_end, COALESCE(period.tracked_until, day_end)) AS terminal_at,
               period.tracked_until IS NULL OR period.tracked_until > day_end AS tracked_at_end
        FROM public.gfn_target_tracking_periods period
        WHERE period.tracked_from < day_end
          AND (period.tracked_until IS NULL OR period.tracked_until > day_start)
    ), protocol AS (
        SELECT daily.*
        FROM public.gfn_site_target_protocol_daily daily
        WHERE daily.fact_date = target_date
    )
    INSERT INTO public.gfn_site_target_daily (
        target_tracking_period_id, site_id, collector_domain_id, target,
        fact_date, snapshot_at, tracked_at_end,
        http_probe_status, http_probe_observed_at, http_state_observed_at,
        http_status_code, http_response_time_ms, http_ttfb_ms, http_protocol,
        http_server, http_remote_ip, http_final_url, http_canonical_url,
        http_security_headers, tls_state_observed_at, tls_handshake, tls_version,
        tls_cipher_suite, tls_cert_verified, tls_verify_error_category,
        tls_cert_not_before, tls_cert_not_after, tls_cert_issuer,
        tls_cert_fingerprint_sha256, tls_cert_spki_sha256, tls_cert_dns_names,
        dns_probe_status, dns_probe_observed_at, dns_state_observed_at,
        dns_has_a, dns_has_aaaa, dns_ipv4_count, dns_ipv6_count,
        dns_a_records, dns_aaaa_records, dns_cname_terminal, dns_cname_depth,
        dns_ns_hosts, dns_mx_hosts, dns_min_ttl, dns_max_ttl, dns_risk_flags,
        ping_probe_status, ping_probe_observed_at, ping_state_observed_at,
        ping_avg_rtt_ms, ping_min_rtt_ms, ping_max_rtt_ms, ping_loss_rate,
        ping_jitter_ms, ping_selected_ip, ping_ip_family,
        ping_icmp_blocked_suspected, projection_version, finalized_at,
        created_at, updated_at
    )
    SELECT eligible.id, eligible.site_id, eligible.collector_domain_id,
           eligible.target, target_date, eligible.terminal_at,
           eligible.tracked_at_end,
           http.latest_observation_status, http.latest_observation_at,
           http.known_state_observed_at,
           (http.known_state ->> 'status_code')::integer,
           (http.known_state ->> 'response_time_ms')::bigint,
           (http.known_state ->> 'ttfb_ms')::bigint,
           http.known_state ->> 'http_protocol',
           http.known_state ->> 'server',
           http.known_state ->> 'remote_ip',
           http.known_state ->> 'final_url',
           http.known_state ->> 'canonical_url',
           http.known_state -> 'security_headers',
           http.known_state_observed_at,
           http.known_state #>> '{tls,handshake}',
           http.known_state #>> '{tls,version}',
           http.known_state #>> '{tls,cipher_suite}',
           (http.known_state #>> '{tls,cert_verified}')::boolean,
           http.known_state #>> '{tls,verify_error_category}',
           (http.known_state #>> '{tls,cert_not_before}')::timestamptz,
           (http.known_state #>> '{tls,cert_not_after}')::timestamptz,
           http.known_state #>> '{tls,cert_issuer}',
           http.known_state #>> '{tls,fingerprint_sha256}',
           http.known_state #>> '{tls,spki_sha256}',
           public.gfn_jsonb_text_array(http.known_state #> '{tls,cert_dns_names}'),
           dns.latest_observation_status, dns.latest_observation_at,
           dns.known_state_observed_at,
           (dns.known_state ->> 'has_a')::boolean,
           (dns.known_state ->> 'has_aaaa')::boolean,
           (dns.known_state ->> 'ipv4_count')::integer,
           (dns.known_state ->> 'ipv6_count')::integer,
           public.gfn_jsonb_text_array(dns.known_state -> 'a_records'),
           public.gfn_jsonb_text_array(dns.known_state -> 'aaaa_records'),
           dns.known_state ->> 'cname_terminal',
           (dns.known_state ->> 'cname_depth')::integer,
           public.gfn_jsonb_text_array(dns.known_state -> 'ns_hosts'),
           public.gfn_jsonb_text_array(dns.known_state -> 'mx_hosts'),
           (dns.known_state ->> 'min_ttl')::integer,
           (dns.known_state ->> 'max_ttl')::integer,
           public.gfn_jsonb_text_array(dns.known_state -> 'risk_flags'),
           ping.latest_observation_status, ping.latest_observation_at,
           ping.known_state_observed_at,
           (ping.known_state ->> 'avg_rtt_ms')::double precision,
           (ping.known_state ->> 'min_rtt_ms')::double precision,
           (ping.known_state ->> 'max_rtt_ms')::double precision,
           (ping.known_state ->> 'loss_rate')::double precision,
           (ping.known_state ->> 'jitter_ms')::double precision,
           ping.known_state ->> 'selected_ip',
           ping.known_state ->> 'ip_family',
           (ping.known_state ->> 'icmp_blocked_suspected')::boolean,
           1, transaction_timestamp(), transaction_timestamp(), transaction_timestamp()
    FROM eligible
    LEFT JOIN protocol http
      ON http.target_tracking_period_id = eligible.id AND http.protocol = 'http'
    LEFT JOIN protocol dns
      ON dns.target_tracking_period_id = eligible.id AND dns.protocol = 'dns'
    LEFT JOIN protocol ping
      ON ping.target_tracking_period_id = eligible.id AND ping.protocol = 'ping';

    GET DIAGNOSTICS projected_rows = ROW_COUNT;
    RETURN projected_rows;
END;
$function$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.gfn_project_site_fact_day(target_date date)
RETURNS bigint
LANGUAGE plpgsql
AS $function$
DECLARE
    day_start timestamp with time zone := target_date::timestamp AT TIME ZONE 'UTC';
    day_end timestamp with time zone := (target_date + 1)::timestamp AT TIME ZONE 'UTC';
    projected_rows bigint;
BEGIN
    WITH eligible_sites AS (
        SELECT DISTINCT site_id
        FROM public.gfn_site_daily
        WHERE fact_date <= target_date
          AND snapshot_at < day_end
    ), latest AS (
        SELECT prior.*
        FROM eligible_sites eligible
        JOIN LATERAL (
            SELECT daily.*
            FROM public.gfn_site_daily daily
            WHERE daily.site_id = eligible.site_id
              AND daily.fact_date <= target_date
            ORDER BY daily.fact_date DESC
            LIMIT 1
        ) prior ON true
        WHERE prior.tracked_at_end OR prior.fact_date = target_date
    ), terminal_primary AS (
        SELECT latest.site_id,
               primary_period.target_tracking_period_id,
               target_period.target,
               primary_period.basis
        FROM latest
        LEFT JOIN LATERAL (
            SELECT candidate.*
            FROM public.gfn_site_primary_target_periods candidate
            WHERE candidate.site_id = latest.site_id
              AND candidate.effective_from < day_end
              AND (candidate.effective_until IS NULL OR candidate.effective_until > day_end)
            ORDER BY candidate.effective_from DESC, candidate.id DESC
            LIMIT 1
        ) primary_period ON true
        LEFT JOIN public.gfn_target_tracking_periods target_period
          ON target_period.id = primary_period.target_tracking_period_id
    ), target_counts AS (
        SELECT latest.site_id, count(target_period.id)::integer AS active_target_count
        FROM latest
        LEFT JOIN public.gfn_target_tracking_periods target_period
          ON target_period.site_id = latest.site_id
         AND target_period.tracked_from < day_end
         AND (target_period.tracked_until IS NULL OR target_period.tracked_until > day_end)
        GROUP BY latest.site_id
    )
    INSERT INTO public.gfn_site_daily (
        site_id, fact_date, snapshot_at, tracked_at_end, name, name_en,
        site_country, nsfw, welfare, view_count, group_ids,
        primary_target_tracking_period_id, primary_target, primary_basis,
        active_target_count, projection_version, finalized_at,
        created_at, updated_at
    )
    SELECT latest.site_id, target_date,
           CASE WHEN latest.fact_date = target_date AND NOT latest.tracked_at_end
                THEN latest.snapshot_at ELSE day_end END,
           latest.tracked_at_end, latest.name, latest.name_en,
           latest.site_country, latest.nsfw, latest.welfare, latest.view_count,
           latest.group_ids, primary_state.target_tracking_period_id,
           primary_state.target, primary_state.basis, target_counts.active_target_count,
           1, transaction_timestamp(), latest.created_at, transaction_timestamp()
    FROM latest
    JOIN terminal_primary primary_state ON primary_state.site_id = latest.site_id
    JOIN target_counts ON target_counts.site_id = latest.site_id
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
        finalized_at = transaction_timestamp(),
        updated_at = transaction_timestamp();

    GET DIAGNOSTICS projected_rows = ROW_COUNT;
    RETURN projected_rows;
END;
$function$;
-- +goose StatementEnd

-- +goose Down

DROP FUNCTION public.gfn_project_site_fact_day(date);
DROP FUNCTION public.gfn_project_target_fact_day(date);
DROP FUNCTION public.gfn_jsonb_text_array(jsonb);
