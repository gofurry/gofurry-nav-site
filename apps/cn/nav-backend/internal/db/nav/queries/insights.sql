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
