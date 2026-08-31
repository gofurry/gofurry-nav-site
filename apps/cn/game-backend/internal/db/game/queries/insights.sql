-- name: CountGameInsightGames :one
SELECT count(*)::bigint
FROM public.gfg_game;

-- name: GetGameInsightGame :one
SELECT id, name, name_en
FROM public.gfg_game
WHERE id = sqlc.arg(game_id);

-- name: GetGameInsightMetricSummary :one
WITH latest AS (
    SELECT fact_date, population_count, eligible_count, not_applicable_count,
           positive_count, negative_count, stale_count, not_probed_count,
           probe_failed_count, unknown_count
    FROM public.gfg_metric_daily
    WHERE metric_key = sqlc.arg(metric_key)
      AND metric_version = sqlc.arg(metric_version)
      AND dimension_key = 'global'
      AND dimension_value = 'all'
    ORDER BY fact_date DESC
    LIMIT 1
), available AS (
    SELECT min(fact_date)::date AS available_from
    FROM public.gfg_metric_daily
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
LEFT JOIN public.gfg_metric_daily previous
  ON previous.metric_key = sqlc.arg(metric_key)
 AND previous.metric_version = sqlc.arg(metric_version)
 AND previous.dimension_key = 'global'
 AND previous.dimension_value = 'all'
 AND previous.fact_date = latest.fact_date - 30;

-- name: ListGameInsightMetricTrend :many
WITH latest AS (
    SELECT max(fact_date)::date AS fact_date
    FROM public.gfg_metric_daily
    WHERE metric_key = sqlc.arg(metric_key)
      AND metric_version = sqlc.arg(metric_version)
      AND dimension_key = 'global'
      AND dimension_value = 'all'
)
SELECT daily.fact_date,
       daily.eligible_count,
       daily.positive_count,
       daily.negative_count
FROM public.gfg_metric_daily daily
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

-- name: GetGameInsightState :one
SELECT game_id, fact_date, tracking_period_id, appid, tracked_at_end,
       is_free, windows, linux, release_availability
FROM public.gfg_game_daily
WHERE game_id = sqlc.arg(game_id)
  AND finalized_at IS NOT NULL
ORDER BY fact_date DESC
LIMIT 1;

-- name: GetGameInsightPlayerSummary :one
SELECT EXISTS (
           SELECT 1
           FROM public.gfg_game_player_counts raw
	       JOIN public.gfg_game_tracking_periods period
	         ON period.id = sqlc.arg(tracking_period_id)
           WHERE raw.game_id = sqlc.arg(game_id)
             AND raw.appid = sqlc.arg(appid)
             AND raw.status = 'success'
	         AND raw.collected_at >= period.tracked_from
	         AND (period.tracked_until IS NULL OR raw.collected_at < period.tracked_until)
       ) AS has_current,
       COALESCE((
           SELECT raw.count
           FROM public.gfg_game_player_counts raw
	       JOIN public.gfg_game_tracking_periods period
	         ON period.id = sqlc.arg(tracking_period_id)
           WHERE raw.game_id = sqlc.arg(game_id)
             AND raw.appid = sqlc.arg(appid)
             AND raw.status = 'success'
	         AND raw.collected_at >= period.tracked_from
	         AND (period.tracked_until IS NULL OR raw.collected_at < period.tracked_until)
           ORDER BY raw.collected_at DESC, raw.id DESC
           LIMIT 1
       ), 0)::bigint AS current_players,
       (
           SELECT raw.collected_at
           FROM public.gfg_game_player_counts raw
	       JOIN public.gfg_game_tracking_periods period
	         ON period.id = sqlc.arg(tracking_period_id)
           WHERE raw.game_id = sqlc.arg(game_id)
             AND raw.appid = sqlc.arg(appid)
             AND raw.status = 'success'
	         AND raw.collected_at >= period.tracked_from
	         AND (period.tracked_until IS NULL OR raw.collected_at < period.tracked_until)
           ORDER BY raw.collected_at DESC, raw.id DESC
           LIMIT 1
       ) AS current_as_of,
       EXISTS (
           SELECT 1
           FROM public.gfg_game_player_daily daily
           WHERE daily.tracking_period_id = sqlc.arg(tracking_period_id)
             AND daily.fact_date BETWEEN sqlc.arg(through_date)::date - 29
                                     AND sqlc.arg(through_date)::date
             AND daily.successful_samples > 0
             AND daily.max_players IS NOT NULL
       ) AS has_peak_30d,
       COALESCE((
           SELECT max(daily.max_players)::bigint
           FROM public.gfg_game_player_daily daily
           WHERE daily.tracking_period_id = sqlc.arg(tracking_period_id)
             AND daily.fact_date BETWEEN sqlc.arg(through_date)::date - 29
                                     AND sqlc.arg(through_date)::date
             AND daily.successful_samples > 0
             AND daily.max_players IS NOT NULL
       ), 0)::bigint AS peak_30d;

-- name: GetGameInsightPriceSummary :one
SELECT fact_date, price_state, currency, initial_amount,
       final_amount, discount_percent
FROM public.gfg_game_price_daily
WHERE tracking_period_id = sqlc.arg(tracking_period_id)
  AND region = 'CN'
  AND finalized_at IS NOT NULL
ORDER BY fact_date DESC
LIMIT 1;

-- name: ListGameInsightPlayerHistory :many
WITH latest AS (
    SELECT max(fact_date)::date AS fact_date
    FROM public.gfg_game_player_daily
    WHERE tracking_period_id = sqlc.arg(tracking_period_id)
      AND finalized_at IS NOT NULL
      AND successful_samples > 0
      AND max_players IS NOT NULL
)
SELECT daily.fact_date,
       daily.min_players,
       daily.max_players,
       daily.avg_players
FROM public.gfg_game_player_daily daily
CROSS JOIN latest
WHERE daily.tracking_period_id = sqlc.arg(tracking_period_id)
  AND daily.finalized_at IS NOT NULL
  AND daily.successful_samples > 0
  AND daily.max_players IS NOT NULL
  AND (
      sqlc.arg(range_days)::integer = 0
      OR daily.fact_date >= latest.fact_date - (sqlc.arg(range_days)::integer - 1)
  )
ORDER BY daily.fact_date;

-- name: ListGameInsightPriceHistory :many
WITH latest AS (
    SELECT max(fact_date)::date AS fact_date
    FROM public.gfg_game_price_daily
    WHERE tracking_period_id = sqlc.arg(tracking_period_id)
      AND region = 'CN'
      AND finalized_at IS NOT NULL
)
SELECT daily.fact_date,
       daily.price_state,
       daily.currency,
       daily.initial_amount,
       daily.final_amount,
       daily.discount_percent
FROM public.gfg_game_price_daily daily
CROSS JOIN latest
WHERE daily.tracking_period_id = sqlc.arg(tracking_period_id)
  AND daily.region = 'CN'
  AND daily.finalized_at IS NOT NULL
  AND (
      sqlc.arg(range_days)::integer = 0
      OR daily.fact_date >= latest.fact_date - (sqlc.arg(range_days)::integer - 1)
  )
ORDER BY daily.fact_date;

-- name: CountGameInsightOverviewChanges :one
SELECT count(*)::bigint
FROM public.gfg_change_events event
WHERE event.projection_date >= (now() AT TIME ZONE 'UTC')::date - 6
  AND event.detector_key = ANY(sqlc.arg(detector_keys)::text[])
  AND (event.detector_key <> 'game_price_transition'
       OR (event.scope_kind = 'region' AND event.scope_key = 'CN'))
  AND event.detector_key || '/' || event.detector_version::text || '/' || event.event_code
      = ANY(sqlc.arg(contract_ids)::text[]);

-- name: ListGameInsightOverviewChanges :many
WITH newest AS (
    SELECT DISTINCT ON (event.game_id)
           event.game_id,
           event.detector_key,
           event.detector_version,
           event.event_code,
           event.projection_date,
           event.time_basis,
           event.event_at,
           event.event_key
    FROM public.gfg_change_events event
    WHERE event.detector_key = ANY(sqlc.arg(detector_keys)::text[])
      AND (event.detector_key <> 'game_price_transition'
           OR (event.scope_kind = 'region' AND event.scope_key = 'CN'))
      AND event.detector_key || '/' || event.detector_version::text || '/' || event.event_code
          = ANY(sqlc.arg(contract_ids)::text[])
    ORDER BY event.game_id, event.projection_date DESC,
             event.event_at DESC NULLS LAST, event.event_key DESC
)
SELECT newest.game_id,
       COALESCE(NULLIF(history.name, ''), NULLIF(game.name, ''), '')::text AS game_name,
       newest.detector_key,
       newest.detector_version,
       newest.event_code,
       newest.projection_date,
       newest.time_basis,
       newest.event_at
FROM newest
LEFT JOIN public.gfg_game_daily history
  ON history.game_id = newest.game_id
 AND history.fact_date = newest.projection_date
LEFT JOIN public.gfg_game game ON game.id = newest.game_id
ORDER BY newest.projection_date DESC,
         newest.event_at DESC NULLS LAST,
         newest.game_id DESC
LIMIT sqlc.arg(limit_count);

-- name: ListGameInsightGameChanges :many
SELECT event.detector_key,
       event.detector_version,
       event.event_code,
       event.projection_date,
       event.time_basis,
       event.event_at
FROM public.gfg_change_events event
WHERE event.game_id = sqlc.arg(game_id)
  AND event.detector_key = ANY(sqlc.arg(detector_keys)::text[])
  AND (event.detector_key <> 'game_price_transition'
       OR (event.scope_kind = 'region' AND event.scope_key = 'CN'))
  AND event.detector_key || '/' || event.detector_version::text || '/' || event.event_code
      = ANY(sqlc.arg(contract_ids)::text[])
ORDER BY event.projection_date DESC,
         event.event_at DESC NULLS LAST,
         event.event_key DESC
LIMIT sqlc.arg(limit_count);
