-- name: CountNavInsightSites :one
SELECT count(*)::bigint
FROM public.gfn_site
WHERE deleted IS NOT TRUE;

-- name: GetNavInsightSite :one
SELECT id, name, name_en
FROM public.gfn_site
WHERE id = sqlc.arg(site_id)
  AND deleted IS NOT TRUE;

-- name: GetNavInsightMetricSummary :one
WITH latest AS (
    SELECT fact_date, population_count, eligible_count, not_applicable_count,
           positive_count, negative_count, stale_count, not_probed_count,
           probe_failed_count, unknown_count
    FROM public.gfn_metric_daily
    WHERE metric_key = sqlc.arg(metric_key)
      AND metric_version = sqlc.arg(metric_version)
      AND dimension_key = 'global'
      AND dimension_value = 'all'
    ORDER BY fact_date DESC
    LIMIT 1
), available AS (
    SELECT min(fact_date)::date AS available_from
    FROM public.gfn_metric_daily
    WHERE metric_key = sqlc.arg(metric_key)
      AND metric_version = sqlc.arg(metric_version)
      AND dimension_key = 'global'
      AND dimension_value = 'all'
)
SELECT latest.fact_date,
       latest.population_count,
       latest.eligible_count,
       latest.not_applicable_count,
       latest.positive_count,
       latest.negative_count,
       latest.stale_count,
       latest.not_probed_count,
       latest.probe_failed_count,
       latest.unknown_count,
       previous.positive_count AS previous_positive_count,
       previous.negative_count AS previous_negative_count,
       available.available_from
FROM latest
CROSS JOIN available
LEFT JOIN public.gfn_metric_daily previous
  ON previous.metric_key = sqlc.arg(metric_key)
 AND previous.metric_version = sqlc.arg(metric_version)
 AND previous.dimension_key = 'global'
 AND previous.dimension_value = 'all'
 AND previous.fact_date = latest.fact_date - 30;

-- name: ListNavInsightMetricTrend :many
WITH latest AS (
    SELECT max(fact_date)::date AS fact_date
    FROM public.gfn_metric_daily
    WHERE metric_key = sqlc.arg(metric_key)
      AND metric_version = sqlc.arg(metric_version)
      AND dimension_key = 'global'
      AND dimension_value = 'all'
)
SELECT daily.fact_date,
       daily.eligible_count,
       daily.positive_count,
       daily.negative_count
FROM public.gfn_metric_daily daily
CROSS JOIN latest
WHERE daily.metric_key = sqlc.arg(metric_key)
  AND daily.metric_version = sqlc.arg(metric_version)
  AND daily.dimension_key = 'global'
  AND daily.dimension_value = 'all'
  AND (
      sqlc.arg(range_days)::integer = 0
      OR daily.fact_date >= latest.fact_date - (sqlc.arg(range_days)::integer - 1)
  )
ORDER BY daily.fact_date;

-- name: ListNavInsightMetricBreakdown :many
SELECT daily.dimension_value,
       COALESCE(CASE WHEN daily.dimension_key = 'group_id' THEN site_group.name END, '')::text AS label,
       COALESCE(CASE WHEN daily.dimension_key = 'group_id' THEN site_group.name_en END, '')::text AS label_en,
       daily.population_count,
       daily.eligible_count,
       daily.positive_count,
       daily.negative_count
FROM public.gfn_metric_daily daily
LEFT JOIN public.gfn_site_group site_group
  ON daily.dimension_key = 'group_id'
 AND site_group.id::text = daily.dimension_value
WHERE daily.metric_key = sqlc.arg(metric_key)
  AND daily.metric_version = sqlc.arg(metric_version)
  AND daily.fact_date = sqlc.arg(fact_date)::date
  AND daily.dimension_key = sqlc.arg(dimension_key)
ORDER BY daily.eligible_count DESC, daily.dimension_value ASC;

-- name: GetNavInsightMetricSliceAvailability :one
WITH available AS (
    SELECT min(fact_date)::date AS available_from,
           max(fact_date)::date AS available_through
    FROM public.gfn_metric_daily
    WHERE metric_key = sqlc.arg(metric_key)
      AND metric_version = sqlc.arg(metric_version)
      AND dimension_key = sqlc.arg(dimension_key)
      AND dimension_value = sqlc.arg(dimension_value)
)
SELECT available.available_from,
       available.available_through,
       COALESCE(CASE WHEN sqlc.arg(dimension_key)::text = 'group_id' THEN site_group.name END, '')::text AS label,
       COALESCE(CASE WHEN sqlc.arg(dimension_key)::text = 'group_id' THEN site_group.name_en END, '')::text AS label_en
FROM available
LEFT JOIN public.gfn_site_group site_group
  ON sqlc.arg(dimension_key)::text = 'group_id'
 AND site_group.id::text = sqlc.arg(dimension_value)::text;

-- name: ListNavInsightMetricSliceTrend :many
SELECT daily.fact_date,
       daily.population_count,
       daily.eligible_count,
       daily.positive_count,
       daily.negative_count
FROM public.gfn_metric_daily daily
WHERE daily.metric_key = sqlc.arg(metric_key)
  AND daily.metric_version = sqlc.arg(metric_version)
  AND daily.dimension_key = sqlc.arg(dimension_key)
  AND daily.dimension_value = sqlc.arg(dimension_value)
  AND daily.fact_date <= sqlc.arg(through_date)::date
  AND (
      sqlc.arg(range_days)::integer = 0
      OR daily.fact_date >= sqlc.arg(through_date)::date - (sqlc.arg(range_days)::integer - 1)
  )
ORDER BY daily.fact_date;

-- name: GetNavInsightSiteMetric :one
SELECT entity.fact_date,
       entity.state,
       global.eligible_count,
       global.positive_count,
       global.negative_count
FROM public.gfn_metric_entity_daily entity
JOIN public.gfn_metric_daily global
  ON global.metric_key = entity.metric_key
 AND global.metric_version = entity.metric_version
 AND global.fact_date = entity.fact_date
 AND global.dimension_key = 'global'
 AND global.dimension_value = 'all'
WHERE entity.metric_key = sqlc.arg(metric_key)
  AND entity.metric_version = sqlc.arg(metric_version)
  AND entity.site_id = sqlc.arg(site_id)
ORDER BY entity.fact_date DESC
LIMIT 1;

-- name: CountNavInsightOverviewChanges :one
SELECT count(*)::bigint
FROM public.gfn_change_events event
WHERE event.projection_date >= (now() AT TIME ZONE 'UTC')::date - 6
  AND event.detector_key = ANY(sqlc.arg(detector_keys)::text[])
  AND event.detector_key || '/' || event.detector_version::text || '/' || event.event_code
      = ANY(sqlc.arg(contract_ids)::text[]);

-- name: ListNavInsightOverviewChanges :many
WITH newest AS (
    SELECT DISTINCT ON (event.site_id)
           event.site_id,
           event.detector_key,
           event.detector_version,
           event.event_code,
           event.projection_date,
           event.time_basis,
           event.event_at,
           event.event_key
    FROM public.gfn_change_events event
    WHERE event.detector_key = ANY(sqlc.arg(detector_keys)::text[])
      AND event.detector_key || '/' || event.detector_version::text || '/' || event.event_code
          = ANY(sqlc.arg(contract_ids)::text[])
    ORDER BY event.site_id, event.projection_date DESC,
             event.event_at DESC NULLS LAST, event.event_key DESC
)
SELECT newest.site_id,
       COALESCE(NULLIF(history.name, ''), NULLIF(site.name, ''), '')::text AS site_name,
       newest.detector_key,
       newest.detector_version,
       newest.event_code,
       newest.projection_date,
       newest.time_basis,
       newest.event_at
FROM newest
LEFT JOIN public.gfn_site_daily history
  ON history.site_id = newest.site_id
 AND history.fact_date = newest.projection_date
LEFT JOIN public.gfn_site site ON site.id = newest.site_id
ORDER BY newest.projection_date DESC,
         newest.event_at DESC NULLS LAST,
         newest.site_id DESC
LIMIT sqlc.arg(limit_count);

-- name: ListNavInsightExplorerChanges :many
SELECT event.site_id,
       COALESCE(NULLIF(history.name, ''), NULLIF(site.name, ''), '')::text AS site_name,
       event.detector_key,
       event.detector_version,
       event.event_code,
       event.projection_date,
       event.time_basis,
       event.event_at,
       CASE WHEN event.time_basis = 'day' THEN 0 ELSE 1 END::integer AS precision_rank,
       CASE
           WHEN event.time_basis = 'day' THEN event.projection_date::timestamp AT TIME ZONE 'UTC'
           ELSE event.event_at
       END::timestamptz AS event_sort_at,
       md5(event.event_key)::text AS opaque_tie
FROM public.gfn_change_events event
LEFT JOIN public.gfn_site_daily history
  ON history.site_id = event.site_id
 AND history.fact_date = event.projection_date
LEFT JOIN public.gfn_site site ON site.id = event.site_id
WHERE event.detector_key = ANY(sqlc.arg(detector_keys)::text[])
  AND event.detector_key || '/' || event.detector_version::text || '/' || event.event_code
      = ANY(sqlc.arg(contract_ids)::text[])
  AND event.projection_date <= sqlc.arg(range_through)::date
  AND (
      sqlc.arg(range_days)::integer = 0
      OR event.projection_date >= sqlc.arg(range_through)::date - (sqlc.arg(range_days)::integer - 1)
  )
  AND (
      NOT sqlc.arg(has_position)::boolean
      OR (
          event.projection_date,
          CASE WHEN event.time_basis = 'day' THEN 0 ELSE 1 END,
          CASE
              WHEN event.time_basis = 'day' THEN event.projection_date::timestamp AT TIME ZONE 'UTC'
              ELSE event.event_at
          END,
          md5(event.event_key)
      ) < (
          sqlc.arg(position_date)::date,
          sqlc.arg(position_rank)::integer,
          sqlc.arg(position_sort_at)::timestamptz,
          sqlc.arg(position_tie)::text
      )
  )
ORDER BY event.projection_date DESC,
         precision_rank DESC,
         event_sort_at DESC,
         opaque_tie DESC
LIMIT sqlc.arg(limit_count);

-- name: ListNavInsightSiteChanges :many
SELECT event.detector_key,
       event.detector_version,
       event.event_code,
       event.projection_date,
       event.time_basis,
       event.event_at
FROM public.gfn_change_events event
WHERE event.site_id = sqlc.arg(site_id)
  AND event.detector_key = ANY(sqlc.arg(detector_keys)::text[])
  AND event.detector_key || '/' || event.detector_version::text || '/' || event.event_code
      = ANY(sqlc.arg(contract_ids)::text[])
ORDER BY event.projection_date DESC,
         event.event_at DESC NULLS LAST,
         event.event_key DESC
LIMIT sqlc.arg(limit_count);

-- name: GetNavCertificateInsightSummary :one
WITH horizon AS (
    SELECT daily.fact_date,
           ((daily.fact_date + 1)::timestamp AT TIME ZONE 'UTC')::timestamptz AS reference_at,
           COALESCE(registry.freshness_seconds, 0)::bigint AS freshness_seconds,
           daily.population_count,
           daily.eligible_count,
           daily.positive_count,
           daily.negative_count,
           daily.not_applicable_count,
           daily.stale_count,
           daily.not_probed_count,
           daily.probe_failed_count,
           daily.unknown_count
    FROM public.gfn_metric_daily daily
    JOIN public.gfn_metric_registry registry
      ON registry.metric_key = daily.metric_key
     AND registry.metric_version = daily.metric_version
    WHERE daily.metric_key = 'tls_certificate_verification'
      AND daily.metric_version = 1
      AND daily.dimension_key = 'global'
      AND daily.dimension_value = 'all'
    ORDER BY daily.fact_date DESC
    LIMIT 1
), certificates AS (
    SELECT target.tls_cert_not_after
    FROM horizon
    JOIN public.gfn_metric_entity_daily entity
      ON entity.metric_key = 'tls_certificate_verification'
     AND entity.metric_version = 1
     AND entity.fact_date = horizon.fact_date
    JOIN public.gfn_site_daily site
      ON site.site_id = entity.site_id
     AND site.fact_date = horizon.fact_date
     AND site.finalized_at IS NOT NULL
    JOIN public.gfn_site_target_daily target
      ON target.target_tracking_period_id = site.primary_target_tracking_period_id
     AND target.fact_date = horizon.fact_date
     AND target.finalized_at IS NOT NULL
    WHERE site.primary_target_tracking_period_id IS NOT NULL
      AND target.tls_state_observed_at IS NOT NULL
      AND target.tls_state_observed_at <= horizon.reference_at
      AND target.tls_state_observed_at >= horizon.reference_at
          - make_interval(secs => horizon.freshness_seconds::double precision)
      AND lower(COALESCE(target.tls_handshake, '')) = 'collected'
      AND target.tls_cert_not_after IS NOT NULL
)
SELECT horizon.fact_date,
       horizon.reference_at,
       horizon.freshness_seconds,
       horizon.population_count,
       horizon.eligible_count,
       horizon.positive_count AS verified_count,
       horizon.negative_count AS failed_count,
       horizon.not_applicable_count,
       horizon.stale_count,
       horizon.not_probed_count,
       horizon.probe_failed_count,
       horizon.unknown_count,
       count(*) FILTER (WHERE certificates.tls_cert_not_after <= horizon.reference_at)::bigint AS expired_count,
       count(*) FILTER (WHERE certificates.tls_cert_not_after > horizon.reference_at
                          AND certificates.tls_cert_not_after <= horizon.reference_at + interval '7 days')::bigint AS expires_within_7d_count,
       count(*) FILTER (WHERE certificates.tls_cert_not_after > horizon.reference_at + interval '7 days'
                          AND certificates.tls_cert_not_after <= horizon.reference_at + interval '30 days')::bigint AS expires_in_8_30d_count,
       count(*) FILTER (WHERE certificates.tls_cert_not_after > horizon.reference_at + interval '30 days')::bigint AS later_count
FROM horizon
LEFT JOIN certificates ON true
GROUP BY horizon.fact_date, horizon.reference_at, horizon.freshness_seconds,
         horizon.population_count, horizon.eligible_count, horizon.positive_count,
         horizon.negative_count, horizon.not_applicable_count, horizon.stale_count,
         horizon.not_probed_count, horizon.probe_failed_count, horizon.unknown_count;

-- name: ListNavCertificateExpiryAttention :many
WITH horizon AS (
    SELECT daily.fact_date,
           ((daily.fact_date + 1)::timestamp AT TIME ZONE 'UTC')::timestamptz AS reference_at,
           COALESCE(registry.freshness_seconds, 0)::bigint AS freshness_seconds
    FROM public.gfn_metric_daily daily
    JOIN public.gfn_metric_registry registry
      ON registry.metric_key = daily.metric_key
     AND registry.metric_version = daily.metric_version
    WHERE daily.metric_key = 'tls_certificate_verification'
      AND daily.metric_version = 1
      AND daily.dimension_key = 'global'
      AND daily.dimension_value = 'all'
    ORDER BY daily.fact_date DESC
    LIMIT 1
)
SELECT site.site_id,
       COALESCE(NULLIF(site.name, ''), NULLIF(site.name_en, ''), '')::text AS site_name,
       target.target,
       target.tls_cert_not_after,
       target.tls_cert_verified AS verified,
       COALESCE(CASE
           WHEN target.tls_cert_verified IS NOT FALSE THEN NULL
           WHEN target.tls_verify_error_category IN (
               'hostname_mismatch', 'unknown_authority', 'expired',
               'not_yet_valid', 'incompatible_usage', 'other'
           ) THEN target.tls_verify_error_category
           ELSE 'other'
       END, '')::text AS verification_issue,
       target.tls_cert_issuer AS issuer,
       target.tls_state_observed_at AS observed_at
FROM horizon
JOIN public.gfn_metric_entity_daily entity
  ON entity.metric_key = 'tls_certificate_verification'
 AND entity.metric_version = 1
 AND entity.fact_date = horizon.fact_date
JOIN public.gfn_site_daily site
  ON site.site_id = entity.site_id
 AND site.fact_date = horizon.fact_date
 AND site.finalized_at IS NOT NULL
JOIN public.gfn_site_target_daily target
  ON target.target_tracking_period_id = site.primary_target_tracking_period_id
 AND target.fact_date = horizon.fact_date
 AND target.finalized_at IS NOT NULL
WHERE site.primary_target_tracking_period_id IS NOT NULL
  AND target.tls_state_observed_at IS NOT NULL
  AND target.tls_state_observed_at <= horizon.reference_at
  AND target.tls_state_observed_at >= horizon.reference_at
      - make_interval(secs => horizon.freshness_seconds::double precision)
  AND lower(COALESCE(target.tls_handshake, '')) = 'collected'
  AND target.tls_cert_not_after IS NOT NULL
  AND target.tls_cert_not_after <= horizon.reference_at + interval '30 days'
ORDER BY target.tls_cert_not_after ASC, site.site_id ASC
LIMIT sqlc.arg(limit_count);

-- name: ListNavCertificateVerificationIssues :many
WITH horizon AS (
    SELECT fact_date,
           ((fact_date + 1)::timestamp AT TIME ZONE 'UTC')::timestamptz AS reference_at
    FROM public.gfn_metric_daily
    WHERE metric_key = 'tls_certificate_verification'
      AND metric_version = 1
      AND dimension_key = 'global'
      AND dimension_value = 'all'
    ORDER BY fact_date DESC
    LIMIT 1
), issues AS (
    SELECT site.site_id,
           COALESCE(NULLIF(site.name, ''), NULLIF(site.name_en, ''), '')::text AS site_name,
           target.target,
           target.tls_cert_not_after,
           target.tls_cert_verified AS verified,
           CASE WHEN target.tls_verify_error_category IN (
                    'hostname_mismatch', 'unknown_authority', 'expired',
                    'not_yet_valid', 'incompatible_usage', 'other'
                ) THEN target.tls_verify_error_category
                ELSE 'other'
           END::text AS verification_issue,
           target.tls_cert_issuer AS issuer,
           target.tls_state_observed_at AS observed_at
    FROM horizon
    JOIN public.gfn_metric_entity_daily entity
      ON entity.metric_key = 'tls_certificate_verification'
     AND entity.metric_version = 1
     AND entity.fact_date = horizon.fact_date
     AND entity.state = 'negative'
    JOIN public.gfn_site_daily site
      ON site.site_id = entity.site_id
     AND site.fact_date = horizon.fact_date
     AND site.finalized_at IS NOT NULL
    JOIN public.gfn_site_target_daily target
      ON target.target_tracking_period_id = site.primary_target_tracking_period_id
     AND target.fact_date = horizon.fact_date
     AND target.finalized_at IS NOT NULL
)
SELECT site_id, site_name, target, tls_cert_not_after,
       verified, verification_issue, issuer, observed_at
FROM issues
ORDER BY verification_issue ASC, site_id ASC
LIMIT sqlc.arg(limit_count);
