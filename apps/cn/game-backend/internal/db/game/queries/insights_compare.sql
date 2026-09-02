-- name: GetGameInsightCompareFactHorizon :one
WITH requested AS (
    SELECT game_id
    FROM unnest(sqlc.arg(game_ids)::bigint[]) AS selected(game_id)
), complete_dates AS (
    SELECT daily.fact_date
    FROM public.gfg_game_daily daily
    JOIN requested ON requested.game_id = daily.game_id
    WHERE daily.finalized_at IS NOT NULL
      AND daily.tracked_at_end
    GROUP BY daily.fact_date
    HAVING count(*) = cardinality(sqlc.arg(game_ids)::bigint[])
)
SELECT max(fact_date)::date AS fact_date
FROM complete_dates;

-- name: ListGameInsightCompareFacts :many
WITH requested AS (
    SELECT game_id, position
    FROM unnest(sqlc.arg(game_ids)::bigint[]) WITH ORDINALITY AS selected(game_id, position)
)
SELECT daily.game_id,
       daily.tracking_period_id,
       daily.is_free,
       daily.windows,
       daily.mac,
       daily.linux,
       daily.release_availability,
       CASE
           WHEN daily.languages_observed_at IS NULL THEN 'unobserved'
           WHEN daily.languages_observed_at < ((daily.fact_date + 1)::timestamp AT TIME ZONE 'UTC') - interval '259200 seconds' THEN 'stale'
           ELSE 'fresh'
       END::text AS language_evidence,
       COALESCE(daily.language_codes, ARRAY[]::text[])::text[] AS language_codes,
       COALESCE(daily.full_audio_language_codes, ARRAY[]::text[])::text[] AS full_audio_language_codes,
       COALESCE(daily.unknown_language_names, ARRAY[]::text[])::text[] AS unknown_language_names,
       (price.fact_date IS NOT NULL)::boolean AS price_available,
       price.price_state,
       price.currency,
       price.initial_amount,
       price.final_amount,
       price.discount_percent
FROM requested
JOIN public.gfg_game_daily daily
  ON daily.game_id = requested.game_id
 AND daily.fact_date = sqlc.arg(fact_date)::date
 AND daily.finalized_at IS NOT NULL
 AND daily.tracked_at_end
LEFT JOIN public.gfg_game_price_daily price
  ON price.tracking_period_id = daily.tracking_period_id
 AND price.fact_date = daily.fact_date
 AND price.region = sqlc.arg(region)
 AND price.finalized_at IS NOT NULL
ORDER BY requested.position;

-- name: ListGameInsightCompareCurrentPlayers :many
WITH requested AS (
    SELECT game_id, position
    FROM unnest(sqlc.arg(game_ids)::bigint[]) WITH ORDINALITY AS selected(game_id, position)
), snapshot AS (
    SELECT id, scheduled_for
    FROM public.gfg_collection_jobs
    WHERE job_key = 'game.players'
      AND trigger = 'scheduled'
      AND scope_type = 'all'
      AND status IN ('success', 'partial')
    ORDER BY scheduled_for DESC, id DESC
    LIMIT 1
), observations AS (
    SELECT DISTINCT ON (raw.game_id)
           raw.game_id,
           raw.count AS player_count,
           raw.collected_at
    FROM snapshot
    JOIN public.gfg_collection_runs run ON run.job_id = snapshot.id
    JOIN public.gfg_game_player_counts raw
      ON raw.run_id = run.id
     AND raw.status = 'success'
    ORDER BY raw.game_id, raw.collected_at DESC, raw.id DESC
)
SELECT requested.game_id::bigint AS game_id,
       snapshot.scheduled_for::timestamptz AS snapshot_scheduled_for,
       (observations.game_id IS NOT NULL)::boolean AS available,
       COALESCE(observations.player_count, 0)::bigint AS player_count,
       observations.collected_at
FROM requested
LEFT JOIN snapshot ON true
LEFT JOIN observations ON observations.game_id = requested.game_id
ORDER BY requested.position;

-- name: ListGameInsightComparePlayer30d :many
WITH requested AS (
    SELECT game_id, position
    FROM unnest(sqlc.arg(game_ids)::bigint[]) WITH ORDINALITY AS selected(game_id, position)
), horizon AS (
    SELECT processed_through
    FROM public.gfg_fact_rollup_checkpoints
    WHERE pipeline_key = 'game.player_facts'
), cohort AS (
    SELECT requested.game_id,
           requested.position,
           horizon.processed_through,
           period.id AS tracking_period_id,
           CASE WHEN period.id IS NULL OR horizon.processed_through IS NULL THEN NULL
                ELSE GREATEST(
                    horizon.processed_through - 29,
                    (period.tracked_from AT TIME ZONE 'UTC')::date
                )::date
           END AS eligible_from
    FROM requested
    LEFT JOIN horizon ON true
    LEFT JOIN LATERAL (
        SELECT candidate.id, candidate.tracked_from
        FROM public.gfg_game_tracking_periods candidate
        WHERE candidate.game_id = requested.game_id
          AND candidate.tracking_basis = 'explicit'
          AND candidate.tracked_from < (horizon.processed_through + 1)::timestamp AT TIME ZONE 'UTC'
          AND (
              candidate.tracked_until IS NULL
              OR candidate.tracked_until >= (horizon.processed_through + 1)::timestamp AT TIME ZONE 'UTC'
          )
        ORDER BY candidate.tracked_from DESC, candidate.id DESC
        LIMIT 1
    ) period ON true
)
SELECT cohort.game_id::bigint AS game_id,
       cohort.processed_through AS fact_through,
       cohort.eligible_from,
       COALESCE(max(daily.max_players) FILTER (WHERE daily.successful_samples > 0), 0)::bigint AS peak_30d,
       COALESCE((
           sum(daily.avg_players * daily.successful_samples) FILTER (WHERE daily.successful_samples > 0)
           / NULLIF(sum(daily.successful_samples) FILTER (WHERE daily.successful_samples > 0), 0)
       ), 0)::double precision AS average_30d,
       count(*) FILTER (WHERE daily.successful_samples > 0)::bigint AS observed_days,
       COALESCE(sum(daily.successful_samples), 0)::bigint AS successful_samples,
       CASE
           WHEN cohort.tracking_period_id IS NOT NULL
            AND count(daily.*) = cohort.processed_through - cohort.eligible_from + 1
            AND bool_and(daily.expected_samples IS NOT NULL)
            AND sum(daily.expected_samples) > 0
           THEN sum(daily.successful_samples)::double precision / sum(daily.expected_samples)
           ELSE 0::double precision
       END::double precision AS sample_coverage,
       COALESCE((
           cohort.tracking_period_id IS NOT NULL
           AND count(daily.*) = cohort.processed_through - cohort.eligible_from + 1
           AND bool_and(daily.expected_samples IS NOT NULL)
           AND sum(daily.expected_samples) > 0
       ), false)::boolean AS has_sample_coverage
FROM cohort
LEFT JOIN public.gfg_game_player_daily daily
  ON daily.tracking_period_id = cohort.tracking_period_id
 AND daily.fact_date BETWEEN cohort.eligible_from AND cohort.processed_through
 AND daily.finalized_at IS NOT NULL
GROUP BY cohort.game_id, cohort.position, cohort.processed_through,
         cohort.tracking_period_id, cohort.eligible_from
ORDER BY cohort.position;
