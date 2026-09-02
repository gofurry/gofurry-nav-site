-- name: GetNavInsightCompareHorizon :one
WITH requested AS (
    SELECT site_id
    FROM unnest(sqlc.arg(site_ids)::bigint[]) AS selected(site_id)
), required AS (
    SELECT keys.metric_key, versions.metric_version
    FROM unnest(sqlc.arg(metric_keys)::text[]) WITH ORDINALITY AS keys(metric_key, position)
    JOIN unnest(sqlc.arg(metric_versions)::integer[]) WITH ORDINALITY AS versions(metric_version, position)
      USING (position)
), complete_dates AS (
    SELECT entity.fact_date
    FROM public.gfn_metric_entity_daily entity
    JOIN requested ON requested.site_id = entity.site_id
    JOIN required
      ON required.metric_key = entity.metric_key
     AND required.metric_version = entity.metric_version
    GROUP BY entity.fact_date
    HAVING count(*) = cardinality(sqlc.arg(site_ids)::bigint[])
        * cardinality(sqlc.arg(metric_keys)::text[])
)
SELECT max(fact_date)::date AS fact_date
FROM complete_dates;

-- name: ListNavInsightCompareCapabilities :many
WITH requested AS (
    SELECT site_id, position
    FROM unnest(sqlc.arg(site_ids)::bigint[]) WITH ORDINALITY AS selected(site_id, position)
), required AS (
    SELECT keys.metric_key, versions.metric_version, keys.position
    FROM unnest(sqlc.arg(metric_keys)::text[]) WITH ORDINALITY AS keys(metric_key, position)
    JOIN unnest(sqlc.arg(metric_versions)::integer[]) WITH ORDINALITY AS versions(metric_version, position)
      USING (position)
)
SELECT entity.site_id,
       entity.metric_key,
       entity.metric_version,
       entity.state
FROM requested
CROSS JOIN required
JOIN public.gfn_metric_entity_daily entity
  ON entity.site_id = requested.site_id
 AND entity.metric_key = required.metric_key
 AND entity.metric_version = required.metric_version
 AND entity.fact_date = sqlc.arg(fact_date)::date
ORDER BY requested.position, required.position;

-- name: ListNavInsightCompareCertificates :many
WITH requested AS (
    SELECT site_id, position
    FROM unnest(sqlc.arg(site_ids)::bigint[]) WITH ORDINALITY AS selected(site_id, position)
), reference AS (
    SELECT sqlc.arg(fact_date)::date AS fact_date,
           ((sqlc.arg(fact_date)::date + 1)::timestamp AT TIME ZONE 'UTC')::timestamptz AS reference_at,
           registry.freshness_seconds
    FROM public.gfn_metric_registry registry
    WHERE registry.metric_key = 'tls_certificate_verification'
      AND registry.metric_version = 1
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
FROM requested
CROSS JOIN reference
JOIN public.gfn_metric_entity_daily entity
  ON entity.site_id = requested.site_id
 AND entity.metric_key = 'tls_certificate_verification'
 AND entity.metric_version = 1
 AND entity.fact_date = reference.fact_date
JOIN public.gfn_site_daily site
  ON site.site_id = requested.site_id
 AND site.fact_date = reference.fact_date
 AND site.finalized_at IS NOT NULL
JOIN public.gfn_site_target_daily target
  ON target.target_tracking_period_id = site.primary_target_tracking_period_id
 AND target.fact_date = reference.fact_date
 AND target.finalized_at IS NOT NULL
WHERE site.primary_target_tracking_period_id IS NOT NULL
  AND target.tls_state_observed_at IS NOT NULL
  AND target.tls_state_observed_at <= reference.reference_at
  AND target.tls_state_observed_at >= reference.reference_at
      - make_interval(secs => reference.freshness_seconds::double precision)
  AND lower(COALESCE(target.tls_handshake, '')) = 'collected'
ORDER BY requested.position;
