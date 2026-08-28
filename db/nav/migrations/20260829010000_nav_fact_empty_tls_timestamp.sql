-- V3-P0.2 hotfix: legacy successful HTTP observations can carry empty TLS
-- certificate timestamp strings. Project those values as unknown instead of
-- aborting the entire target-day transaction with SQLSTATE 22007.
-- +goose Up

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.gfn_project_target_fact_day(target_date date)
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
           NULLIF(btrim(http.known_state #>> '{tls,cert_not_before}'), '')::timestamptz,
           NULLIF(btrim(http.known_state #>> '{tls,cert_not_after}'), '')::timestamptz,
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

-- There is intentionally no Down migration: reverting would restore a
-- projector that aborts on historical empty timestamp evidence.
