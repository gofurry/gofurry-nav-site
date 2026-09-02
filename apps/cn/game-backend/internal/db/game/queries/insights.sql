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

-- name: ListGameInsightMetricBreakdown :many
SELECT daily.dimension_value,
       COALESCE(tag.name, '')::text AS label,
       COALESCE(tag.name_en, '')::text AS label_en,
       daily.population_count,
       daily.eligible_count,
       daily.positive_count,
       daily.negative_count
FROM public.gfg_metric_daily daily
LEFT JOIN public.gfg_tag tag
  ON daily.dimension_key IN ('primary_tag_id', 'tag_id')
 AND tag.id::text = daily.dimension_value
WHERE daily.metric_key = sqlc.arg(metric_key)
  AND daily.metric_version = sqlc.arg(metric_version)
  AND daily.fact_date = sqlc.arg(fact_date)::date
  AND daily.dimension_key = sqlc.arg(dimension_key)
ORDER BY daily.eligible_count DESC, daily.dimension_value ASC;

-- name: GetGameInsightMetricSliceAvailability :one
WITH available AS (
    SELECT min(fact_date)::date AS available_from,
           max(fact_date)::date AS available_through
    FROM public.gfg_metric_daily
    WHERE metric_key = sqlc.arg(metric_key)
      AND metric_version = sqlc.arg(metric_version)
      AND dimension_key = sqlc.arg(dimension_key)
      AND dimension_value = sqlc.arg(dimension_value)
)
SELECT available.available_from,
       available.available_through,
       COALESCE(tag.name, '')::text AS label,
       COALESCE(tag.name_en, '')::text AS label_en
FROM available
LEFT JOIN public.gfg_tag tag
  ON sqlc.arg(dimension_key)::text IN ('primary_tag_id', 'tag_id')
 AND tag.id::text = sqlc.arg(dimension_value)::text;

-- name: ListGameInsightMetricSliceTrend :many
SELECT daily.fact_date,
       daily.population_count,
       daily.eligible_count,
       daily.positive_count,
       daily.negative_count
FROM public.gfg_metric_daily daily
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

-- name: GetGameInsightState :one
SELECT game_id, fact_date, tracking_period_id, appid, tracked_at_end,
       is_free, windows, mac, linux, release_availability
FROM public.gfg_game_daily
WHERE game_id = sqlc.arg(game_id)
  AND finalized_at IS NOT NULL
ORDER BY fact_date DESC
LIMIT 1;

-- name: GetGameInsightPlayerSummary :one
WITH player_horizon AS (
    SELECT processed_through
    FROM public.gfg_fact_rollup_checkpoints
    WHERE pipeline_key = 'game.player_facts'
), quality AS (
    SELECT COALESCE(max(daily.max_players), 0)::bigint AS peak_30d,
           COALESCE(sum(daily.avg_players * daily.successful_samples)
             / NULLIF(sum(daily.successful_samples), 0), 0)::double precision AS average_30d,
           GREATEST(horizon.processed_through - 29, (period.tracked_from AT TIME ZONE 'UTC')::date)::date AS eligible_from,
           count(*) FILTER (WHERE daily.successful_samples > 0)::bigint AS observed_days,
           COALESCE(sum(daily.successful_samples), 0)::bigint AS successful_samples,
           CASE
             WHEN count(daily.*) = horizon.processed_through
                    - GREATEST(horizon.processed_through - 29, (period.tracked_from AT TIME ZONE 'UTC')::date) + 1
              AND bool_and(daily.expected_samples IS NOT NULL)
              AND sum(daily.expected_samples) > 0
             THEN sum(daily.successful_samples)::double precision / sum(daily.expected_samples)
             ELSE 0::double precision
           END::double precision AS sample_coverage,
           (count(daily.*) = horizon.processed_through
              - GREATEST(horizon.processed_through - 29, (period.tracked_from AT TIME ZONE 'UTC')::date) + 1
            AND bool_and(daily.expected_samples IS NOT NULL)
            AND sum(daily.expected_samples) > 0) AS has_sample_coverage,
           horizon.processed_through AS fact_through
    FROM player_horizon horizon
    JOIN public.gfg_game_tracking_periods period ON period.id = sqlc.arg(tracking_period_id)
    LEFT JOIN public.gfg_game_player_daily daily
      ON daily.tracking_period_id = period.id
     AND daily.fact_date BETWEEN GREATEST(horizon.processed_through - 29, (period.tracked_from AT TIME ZONE 'UTC')::date)
                             AND horizon.processed_through
     AND daily.finalized_at IS NOT NULL
    GROUP BY horizon.processed_through, period.tracked_from
)
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
       quality.peak_30d,
       quality.average_30d,
       quality.fact_through,
       quality.eligible_from,
       quality.observed_days,
       quality.successful_samples,
       quality.sample_coverage,
       quality.has_sample_coverage
FROM quality;

-- name: ListGameInsightRegionalPrices :many
WITH regions AS (SELECT unnest(ARRAY['CN', 'US', 'HK']::text[]) AS region)
SELECT regions.region::text AS region,
       price.fact_date,
       price.price_state,
       price.currency,
       price.initial_amount,
       price.final_amount,
       price.discount_percent
FROM regions
LEFT JOIN public.gfg_game_price_daily price
  ON price.tracking_period_id = sqlc.arg(tracking_period_id)
 AND price.region = regions.region
 AND price.fact_date = sqlc.arg(fact_date)::date
 AND price.finalized_at IS NOT NULL
ORDER BY CASE regions.region WHEN 'CN' THEN 1 WHEN 'US' THEN 2 ELSE 3 END;

-- name: GetGameInsightObservedLow :one
WITH RECURSIVE current_price AS (
    SELECT price.*
    FROM public.gfg_game_price_daily price
    WHERE price.tracking_period_id = sqlc.arg(tracking_period_id)
      AND price.region = sqlc.arg(region)
      AND price.fact_date = sqlc.arg(fact_date)::date
      AND price.finalized_at IS NOT NULL
      AND price.price_state = 'priced'
      AND price.currency IS NOT NULL
), identity AS (
    SELECT current_price.* FROM current_price
    UNION ALL
    SELECT previous.*
    FROM identity next
    JOIN public.gfg_game_price_daily previous
      ON previous.tracking_period_id = next.tracking_period_id
     AND previous.region = next.region
     AND previous.fact_date = next.fact_date - 1
     AND previous.finalized_at IS NOT NULL
     AND previous.price_state = 'priced'
     AND previous.currency = next.currency
), low_amount AS (
    SELECT min(final_amount)::bigint AS amount, min(fact_date)::date AS observed_since
    FROM identity
), first_low AS (
    SELECT identity.* FROM identity CROSS JOIN low_amount
    WHERE identity.final_amount = low_amount.amount
    ORDER BY identity.fact_date ASC
    LIMIT 1
)
SELECT low_amount.amount, first_low.currency, first_low.fact_date AS first_seen,
       low_amount.observed_since, first_low.initial_amount, first_low.discount_percent
FROM low_amount
JOIN first_low ON true;

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
SELECT daily.fact_date,
       daily.price_state,
       daily.currency,
       daily.initial_amount,
       daily.final_amount,
       daily.discount_percent
FROM public.gfg_game_price_daily daily
WHERE daily.tracking_period_id = sqlc.arg(tracking_period_id)
  AND daily.region = sqlc.arg(region)
  AND daily.finalized_at IS NOT NULL
  AND daily.fact_date <= sqlc.arg(through_date)::date
  AND (
      sqlc.arg(range_days)::integer = 0
      OR daily.fact_date >= sqlc.arg(through_date)::date - (sqlc.arg(range_days)::integer - 1)
  )
ORDER BY daily.fact_date;

-- name: GetGameInsightLatestPlayerRankingMeta :one
WITH latest_slot AS (
    SELECT max(scheduled_for) AS scheduled_for
    FROM public.gfg_collection_jobs
    WHERE job_key = 'game.players' AND trigger = 'scheduled' AND scope_type = 'all'
      AND status IN ('success', 'partial', 'failed', 'skipped', 'missed', 'canceled')
), snapshot AS (
    SELECT id, scheduled_for
    FROM public.gfg_collection_jobs
    WHERE job_key = 'game.players' AND trigger = 'scheduled' AND scope_type = 'all'
      AND status IN ('success', 'partial')
    ORDER BY scheduled_for DESC, id DESC LIMIT 1
), observations AS (
    SELECT DISTINCT ON (raw.game_id) raw.game_id, raw.collected_at
    FROM snapshot
    JOIN public.gfg_collection_runs run ON run.job_id = snapshot.id
    JOIN public.gfg_game_player_counts raw ON raw.run_id = run.id AND raw.status = 'success'
    ORDER BY raw.game_id, raw.collected_at DESC, raw.id DESC
)
SELECT snapshot.scheduled_for::timestamptz AS snapshot_scheduled_for,
       latest_slot.scheduled_for::timestamptz AS latest_slot_scheduled_for,
       min(observations.collected_at)::timestamptz AS observed_from,
       max(observations.collected_at)::timestamptz AS observed_through,
       (SELECT count(*)::bigint FROM public.gfg_game_tracking_periods period
        WHERE snapshot.scheduled_for IS NOT NULL AND period.tracking_basis = 'explicit'
          AND period.tracked_from <= snapshot.scheduled_for
          AND (period.tracked_until IS NULL OR snapshot.scheduled_for < period.tracked_until)) AS population,
       count(observations.game_id)::bigint AS ranked
FROM latest_slot
LEFT JOIN snapshot ON true
LEFT JOIN observations ON true
GROUP BY snapshot.scheduled_for, latest_slot.scheduled_for;

-- name: ListGameInsightLatestPlayerRanking :many
WITH snapshot AS (
    SELECT id, scheduled_for
    FROM public.gfg_collection_jobs
    WHERE job_key = 'game.players' AND trigger = 'scheduled' AND scope_type = 'all'
      AND status IN ('success', 'partial')
    ORDER BY scheduled_for DESC, id DESC LIMIT 1
), observations AS (
    SELECT DISTINCT ON (raw.game_id) raw.game_id, raw.count AS player_count, raw.collected_at
    FROM snapshot
    JOIN public.gfg_collection_runs run ON run.job_id = snapshot.id
    JOIN public.gfg_game_player_counts raw ON raw.run_id = run.id AND raw.status = 'success'
    JOIN public.gfg_game_tracking_periods period
      ON period.game_id = raw.game_id AND period.tracking_basis = 'explicit'
     AND period.tracked_from <= snapshot.scheduled_for
     AND (period.tracked_until IS NULL OR snapshot.scheduled_for < period.tracked_until)
    ORDER BY raw.game_id, raw.collected_at DESC, raw.id DESC
)
SELECT observations.game_id, COALESCE(NULLIF(game.name, ''), game.name_en, '')::text AS game_name,
       observations.player_count, observations.collected_at
FROM observations
JOIN public.gfg_game game ON game.id = observations.game_id
ORDER BY observations.player_count DESC, observations.game_id ASC
LIMIT sqlc.arg(limit_count);

-- name: GetGameInsightPlayer30dMeta :one
WITH horizon AS (
    SELECT processed_through FROM public.gfg_fact_rollup_checkpoints WHERE pipeline_key = 'game.player_facts'
), population AS (
    SELECT count(*)::bigint AS count
    FROM horizon
    JOIN public.gfg_game_tracking_periods period ON period.tracking_basis = 'explicit'
     AND period.tracked_from < (horizon.processed_through + 1)::timestamp AT TIME ZONE 'UTC'
     AND (period.tracked_until IS NULL OR period.tracked_until >= (horizon.processed_through + 1)::timestamp AT TIME ZONE 'UTC')
), ranked AS (
    SELECT count(DISTINCT daily.tracking_period_id)::bigint AS count
    FROM horizon
    JOIN public.gfg_game_tracking_periods period ON period.tracking_basis = 'explicit'
     AND period.tracked_from < (horizon.processed_through + 1)::timestamp AT TIME ZONE 'UTC'
     AND (period.tracked_until IS NULL OR period.tracked_until >= (horizon.processed_through + 1)::timestamp AT TIME ZONE 'UTC')
    JOIN public.gfg_game_player_daily daily ON daily.tracking_period_id = period.id
     AND daily.fact_date BETWEEN GREATEST(horizon.processed_through - 29, (period.tracked_from AT TIME ZONE 'UTC')::date) AND horizon.processed_through
     AND daily.finalized_at IS NOT NULL AND daily.successful_samples > 0
)
SELECT horizon.processed_through AS window_through,
       (horizon.processed_through - 29)::date AS window_from,
       population.count AS population, ranked.count AS ranked
FROM horizon CROSS JOIN population CROSS JOIN ranked;

-- name: ListGameInsightPlayer30dRanking :many
WITH horizon AS (
    SELECT processed_through FROM public.gfg_fact_rollup_checkpoints WHERE pipeline_key = 'game.player_facts'
), cohort AS (
    SELECT period.id, period.game_id,
           GREATEST(horizon.processed_through - 29, (period.tracked_from AT TIME ZONE 'UTC')::date)::date AS eligible_from,
           horizon.processed_through
    FROM horizon
    JOIN public.gfg_game_tracking_periods period ON period.tracking_basis = 'explicit'
     AND period.tracked_from < (horizon.processed_through + 1)::timestamp AT TIME ZONE 'UTC'
     AND (period.tracked_until IS NULL OR period.tracked_until >= (horizon.processed_through + 1)::timestamp AT TIME ZONE 'UTC')
), aggregate AS (
    SELECT cohort.game_id, cohort.eligible_from,
           max(daily.max_players)::bigint AS peak_30d,
           (sum(daily.avg_players * daily.successful_samples) / NULLIF(sum(daily.successful_samples), 0))::double precision AS average_30d,
           count(*) FILTER (WHERE daily.successful_samples > 0)::bigint AS observed_days,
           sum(daily.successful_samples)::bigint AS successful_samples,
           CASE WHEN count(*) = cohort.processed_through - cohort.eligible_from + 1
                  AND bool_and(daily.expected_samples IS NOT NULL) AND sum(daily.expected_samples) > 0
                THEN sum(daily.successful_samples)::double precision / sum(daily.expected_samples)
                ELSE 0::double precision END::double precision AS sample_coverage,
           (count(*) = cohort.processed_through - cohort.eligible_from + 1
             AND bool_and(daily.expected_samples IS NOT NULL) AND sum(daily.expected_samples) > 0) AS has_sample_coverage
    FROM cohort
    JOIN public.gfg_game_player_daily daily ON daily.tracking_period_id = cohort.id
     AND daily.fact_date BETWEEN cohort.eligible_from AND cohort.processed_through
     AND daily.finalized_at IS NOT NULL
    GROUP BY cohort.game_id, cohort.eligible_from, cohort.processed_through
    HAVING sum(daily.successful_samples) > 0
)
SELECT aggregate.game_id, COALESCE(NULLIF(game.name, ''), game.name_en, '')::text AS game_name,
       aggregate.peak_30d, aggregate.average_30d, aggregate.eligible_from,
       aggregate.observed_days, aggregate.successful_samples, aggregate.sample_coverage, aggregate.has_sample_coverage
FROM aggregate JOIN public.gfg_game game ON game.id = aggregate.game_id
ORDER BY CASE WHEN sqlc.arg(use_average)::boolean THEN aggregate.average_30d ELSE aggregate.peak_30d END DESC,
         aggregate.game_id ASC
LIMIT sqlc.arg(limit_count);

-- name: GetGameInsightPriceOverview :one
WITH horizon AS (
    SELECT max(fact_date)::date AS as_of FROM public.gfg_game_daily
    WHERE finalized_at IS NOT NULL AND tracked_at_end
), cohort AS (
    SELECT fact.game_id, price.price_state, price.discount_percent
    FROM horizon JOIN public.gfg_game_daily fact ON fact.fact_date = horizon.as_of
      AND fact.finalized_at IS NOT NULL AND fact.tracked_at_end
    LEFT JOIN public.gfg_game_price_daily price
      ON price.tracking_period_id = fact.tracking_period_id AND price.fact_date = horizon.as_of
     AND price.region = sqlc.arg(region) AND price.finalized_at IS NOT NULL
)
SELECT horizon.as_of, count(cohort.game_id)::bigint AS population,
       count(*) FILTER (WHERE price_state = 'priced')::bigint AS priced,
       count(*) FILTER (WHERE price_state = 'free')::bigint AS free,
       count(*) FILTER (WHERE price_state = 'unpriced')::bigint AS unpriced,
       count(*) FILTER (WHERE price_state = 'unknown')::bigint AS unknown,
       count(cohort.game_id) FILTER (WHERE price_state IS NULL)::bigint AS unavailable,
       count(*) FILTER (WHERE price_state = 'priced' AND discount_percent > 0)::bigint AS discounted
FROM horizon LEFT JOIN cohort ON true
GROUP BY horizon.as_of;

-- name: ListGameInsightCurrentDiscounts :many
WITH horizon AS (
    SELECT max(fact_date)::date AS as_of FROM public.gfg_game_daily
    WHERE finalized_at IS NOT NULL AND tracked_at_end
)
SELECT horizon.as_of, fact.game_id, fact.tracking_period_id,
       COALESCE(NULLIF(fact.name, ''), fact.name_en, '')::text AS game_name,
       price.currency, price.initial_amount, price.final_amount, price.discount_percent
FROM horizon
JOIN public.gfg_game_daily fact ON fact.fact_date = horizon.as_of
 AND fact.finalized_at IS NOT NULL AND fact.tracked_at_end
JOIN public.gfg_game_price_daily price ON price.tracking_period_id = fact.tracking_period_id
 AND price.fact_date = horizon.as_of AND price.region = sqlc.arg(region)
 AND price.finalized_at IS NOT NULL AND price.price_state = 'priced' AND price.discount_percent > 0
ORDER BY price.discount_percent DESC, fact.game_id ASC
LIMIT sqlc.arg(limit_count);

-- name: GetGameInsightLanguageOverview :one
WITH horizon AS (
    SELECT max(fact_date)::date AS as_of FROM public.gfg_game_daily
    WHERE finalized_at IS NOT NULL AND tracked_at_end
), cohort AS (
    SELECT fact.*,
           CASE WHEN fact.languages_observed_at IS NULL THEN 'unobserved'
                WHEN fact.languages_observed_at < ((horizon.as_of + 1)::timestamp AT TIME ZONE 'UTC') - interval '259200 seconds' THEN 'stale'
                ELSE 'fresh' END AS evidence
    FROM horizon JOIN public.gfg_game_daily fact ON fact.fact_date = horizon.as_of
      AND fact.finalized_at IS NOT NULL AND fact.tracked_at_end
)
SELECT horizon.as_of, count(cohort.game_id)::bigint AS population,
       count(*) FILTER (WHERE evidence = 'fresh')::bigint AS fresh,
       count(*) FILTER (WHERE evidence = 'stale')::bigint AS stale,
       count(*) FILTER (WHERE evidence = 'unobserved')::bigint AS unobserved,
       count(*) FILTER (WHERE evidence = 'fresh' AND COALESCE(cardinality(unknown_language_names), 0) = 0)::bigint AS fully_normalized_games,
       count(*) FILTER (WHERE evidence = 'fresh' AND COALESCE(cardinality(unknown_language_names), 0) > 0)::bigint AS unmapped_games,
       COALESCE(sum(cardinality(unknown_language_names)) FILTER (WHERE evidence = 'fresh'), 0)::bigint AS unmapped_entries
FROM horizon LEFT JOIN cohort ON true GROUP BY horizon.as_of;

-- name: ListGameInsightLanguages :many
WITH horizon AS (
    SELECT max(fact_date)::date AS as_of FROM public.gfg_game_daily
    WHERE finalized_at IS NOT NULL AND tracked_at_end
), fresh AS (
    SELECT fact.* FROM horizon JOIN public.gfg_game_daily fact ON fact.fact_date = horizon.as_of
    WHERE fact.finalized_at IS NOT NULL AND fact.tracked_at_end
      AND fact.languages_observed_at IS NOT NULL
      AND fact.languages_observed_at >= ((horizon.as_of + 1)::timestamp AT TIME ZONE 'UTC') - interval '259200 seconds'
), supported AS (
    SELECT DISTINCT fresh.game_id, code FROM fresh CROSS JOIN LATERAL unnest(COALESCE(fresh.language_codes, ARRAY[]::text[])) code
), audio AS (
    SELECT DISTINCT fresh.game_id, code FROM fresh CROSS JOIN LATERAL unnest(COALESCE(fresh.full_audio_language_codes, ARRAY[]::text[])) code
), names AS (
    SELECT language_code AS code, min(steam_name)::text AS steam_name
    FROM public.gfg_game_languages WHERE language_code IS NOT NULL GROUP BY language_code
), codes AS (
    SELECT code FROM supported UNION SELECT code FROM audio
)
SELECT codes.code::text AS code, COALESCE(names.steam_name, codes.code)::text AS steam_name,
       count(DISTINCT supported.game_id)::bigint AS supported_games,
       count(DISTINCT audio.game_id)::bigint AS explicit_full_audio_games
FROM codes LEFT JOIN supported ON supported.code = codes.code
LEFT JOIN audio ON audio.code = codes.code
LEFT JOIN names ON names.code = codes.code
GROUP BY codes.code, names.steam_name
ORDER BY supported_games DESC, codes.code ASC;

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

-- name: ListGameInsightExplorerChanges :many
SELECT event.game_id,
       COALESCE(NULLIF(history.name, ''), NULLIF(game.name, ''), '')::text AS game_name,
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
FROM public.gfg_change_events event
LEFT JOIN public.gfg_game_daily history
  ON history.game_id = event.game_id
 AND history.fact_date = event.projection_date
LEFT JOIN public.gfg_game game ON game.id = event.game_id
WHERE event.detector_key = ANY(sqlc.arg(detector_keys)::text[])
  AND event.detector_key || '/' || event.detector_version::text || '/' || event.event_code
      = ANY(sqlc.arg(contract_ids)::text[])
  AND (event.detector_key <> 'game_price_transition'
       OR (event.scope_kind = 'region' AND event.scope_key = 'CN'))
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
