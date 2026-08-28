-- V3-P0.2 set-based Game fact projectors.
-- +goose Up

-- +goose StatementBegin
CREATE FUNCTION public.gfg_project_player_fact_day(target_date date)
RETURNS TABLE(hourly_rows bigint, daily_rows bigint)
LANGUAGE plpgsql
AS $function$
DECLARE
    day_start timestamp with time zone := target_date::timestamp AT TIME ZONE 'UTC';
    day_end timestamp with time zone := (target_date + 1)::timestamp AT TIME ZONE 'UTC';
    cutover timestamp with time zone;
BEGIN
    SELECT quality_cutover_at INTO cutover
    FROM public.gfg_fact_rollup_checkpoints
    WHERE pipeline_key = 'game.player_facts';

    DROP TABLE IF EXISTS pg_temp.gfg_player_fact_slots;
    CREATE TEMPORARY TABLE pg_temp.gfg_player_fact_slots ON COMMIT DROP AS
    WITH scheduled_jobs AS (
        SELECT job.id,
               job.scheduled_for,
               job.scope_type,
               job.scope_id,
               job.status AS job_status
        FROM public.gfg_collection_jobs job
        WHERE job.trigger = 'scheduled'
          AND job.job_key = 'game.players'
          AND job.tasks @> ARRAY['players']::text[]
          AND job.scheduled_for >= day_start
          AND job.scheduled_for < day_end
    ), latest_runs AS (
        SELECT DISTINCT ON (run.job_id)
               run.job_id, run.id AS run_id, run.status AS run_status, run.attempt_no
        FROM public.gfg_collection_runs run
        JOIN scheduled_jobs job ON job.id = run.job_id
        ORDER BY run.job_id, run.attempt_no DESC, run.started_at DESC
    ), latest_results AS (
        SELECT DISTINCT ON (run.job_id, result.game_id)
               run.job_id,
               result.game_id,
               result.run_id,
               result.status,
               NULLIF(result.error_kind, '') AS error_kind,
               result.started_at,
               result.ended_at,
               run.attempt_no
        FROM public.gfg_collection_runs run
        JOIN public.gfg_collection_task_results result ON result.run_id = run.id
        JOIN scheduled_jobs job ON job.id = run.job_id
        WHERE result.task_type = 'players'
        ORDER BY run.job_id, result.game_id, run.attempt_no DESC,
                 result.started_at DESC, result.id DESC
    )
    SELECT period.id AS tracking_period_id,
           period.game_id,
           period.appid,
           job.id AS job_id,
           job.scheduled_for,
           date_trunc('hour', job.scheduled_for) AS bucket_start,
           CASE
               WHEN result.status = 'success' THEN 'success'
               WHEN result.status = 'partial' THEN 'partial'
               WHEN result.status = 'failed' THEN 'failure'
               WHEN result.status = 'skipped' THEN 'skipped'
               WHEN job.job_status = 'missed' THEN 'missed'
               WHEN job.job_status = 'skipped' THEN 'skipped'
               WHEN job.job_status = 'canceled' OR run.run_status = 'canceled' THEN 'canceled'
               ELSE 'unattempted'
           END AS outcome,
           result.error_kind,
           raw.status AS raw_status,
           raw.count AS player_count,
           raw.collected_at,
           CASE WHEN raw.collected_at IS NULL THEN NULL
                ELSE (extract(epoch FROM (raw.collected_at - job.scheduled_for)) * 1000)::bigint
           END AS observation_lag_ms
    FROM scheduled_jobs job
    JOIN public.gfg_game_tracking_periods period
      ON period.tracked_from <= job.scheduled_for
     AND (period.tracked_until IS NULL OR job.scheduled_for < period.tracked_until)
     AND (job.scope_type = 'all'
          OR (job.scope_type = 'game' AND job.scope_id = period.game_id))
    LEFT JOIN latest_runs run ON run.job_id = job.id
    LEFT JOIN latest_results result
      ON result.job_id = job.id AND result.game_id = period.game_id
    LEFT JOIN public.gfg_game_player_counts raw
      ON raw.run_id = result.run_id AND raw.game_id = period.game_id;

    DELETE FROM public.gfg_game_player_hourly WHERE bucket_start >= day_start AND bucket_start < day_end;
    DELETE FROM public.gfg_game_player_daily WHERE fact_date = target_date;

    WITH failure_counts AS (
        SELECT tracking_period_id,
               bucket_start,
               jsonb_object_agg(error_kind, failure_count ORDER BY error_kind) AS counts
        FROM (
            SELECT tracking_period_id, bucket_start,
                   COALESCE(error_kind, 'unknown') AS error_kind,
                   count(*)::bigint AS failure_count
            FROM pg_temp.gfg_player_fact_slots
            WHERE outcome = 'failure'
            GROUP BY tracking_period_id, bucket_start, COALESCE(error_kind, 'unknown')
        ) failures
        GROUP BY tracking_period_id, bucket_start
    ), aggregated AS (
        SELECT slot.tracking_period_id,
               slot.game_id,
               slot.appid,
               slot.bucket_start,
               min(slot.player_count) FILTER (WHERE slot.outcome = 'success' AND slot.raw_status = 'success') AS min_players,
               max(slot.player_count) FILTER (WHERE slot.outcome = 'success' AND slot.raw_status = 'success') AS max_players,
               avg(slot.player_count::double precision) FILTER (WHERE slot.outcome = 'success' AND slot.raw_status = 'success') AS avg_players,
               percentile_cont(0.5) WITHIN GROUP (ORDER BY slot.player_count)
                   FILTER (WHERE slot.outcome = 'success' AND slot.raw_status = 'success') AS median_players,
               count(*)::integer AS expected_samples,
               count(*) FILTER (WHERE slot.outcome IN ('success', 'partial', 'failure'))::integer AS attempted_samples,
               count(*) FILTER (WHERE slot.outcome = 'success')::integer AS successful_samples,
               count(*) FILTER (WHERE slot.outcome = 'partial')::integer AS partial_samples,
               count(*) FILTER (WHERE slot.outcome = 'failure')::integer AS failed_samples,
               count(*) FILTER (WHERE slot.outcome = 'skipped')::integer AS skipped_samples,
               count(*) FILTER (WHERE slot.outcome = 'missed')::integer AS missed_samples,
               count(*) FILTER (WHERE slot.outcome = 'canceled')::integer AS canceled_samples,
               count(*) FILTER (WHERE slot.outcome = 'unattempted')::integer AS unattempted_samples,
               min(slot.collected_at) AS first_observed_at,
               max(slot.collected_at) AS last_observed_at,
               avg(slot.observation_lag_ms)::bigint AS avg_observation_lag_ms,
               max(slot.observation_lag_ms) AS max_observation_lag_ms
        FROM pg_temp.gfg_player_fact_slots slot
        GROUP BY slot.tracking_period_id, slot.game_id, slot.appid, slot.bucket_start
    )
    INSERT INTO public.gfg_game_player_hourly (
        tracking_period_id, game_id, appid, bucket_start,
        min_players, max_players, avg_players, median_players,
        expected_samples, attempted_samples, successful_samples, partial_samples,
        failed_samples, skipped_samples, missed_samples, canceled_samples,
        unattempted_samples, failure_kind_counts, first_observed_at,
        last_observed_at, avg_observation_lag_ms, max_observation_lag_ms,
        quality_basis, projection_version, finalized_at, created_at, updated_at
    )
    SELECT aggregate.tracking_period_id, aggregate.game_id, aggregate.appid, aggregate.bucket_start,
           aggregate.min_players, aggregate.max_players, aggregate.avg_players, aggregate.median_players,
           CASE WHEN day_start >= cutover THEN aggregate.expected_samples END,
           aggregate.attempted_samples, aggregate.successful_samples, aggregate.partial_samples,
           aggregate.failed_samples,
           CASE WHEN day_start >= cutover THEN aggregate.skipped_samples END,
           CASE WHEN day_start >= cutover THEN aggregate.missed_samples END,
           CASE WHEN day_start >= cutover THEN aggregate.canceled_samples END,
           CASE WHEN day_start >= cutover THEN aggregate.unattempted_samples END,
           COALESCE(failure.counts, '{}'::jsonb), aggregate.first_observed_at,
           aggregate.last_observed_at, aggregate.avg_observation_lag_ms,
           aggregate.max_observation_lag_ms,
           CASE WHEN day_start >= cutover THEN 'acquisition_ledger' ELSE 'legacy_observed_only' END,
           1, transaction_timestamp(), transaction_timestamp(), transaction_timestamp()
    FROM aggregated aggregate
    LEFT JOIN failure_counts failure
      ON failure.tracking_period_id = aggregate.tracking_period_id
     AND failure.bucket_start = aggregate.bucket_start;

    GET DIAGNOSTICS hourly_rows = ROW_COUNT;

    WITH failure_counts AS (
        SELECT tracking_period_id,
               jsonb_object_agg(error_kind, failure_count ORDER BY error_kind) AS counts
        FROM (
            SELECT tracking_period_id,
                   COALESCE(error_kind, 'unknown') AS error_kind,
                   count(*)::bigint AS failure_count
            FROM pg_temp.gfg_player_fact_slots
            WHERE outcome = 'failure'
            GROUP BY tracking_period_id, COALESCE(error_kind, 'unknown')
        ) failures
        GROUP BY tracking_period_id
    ), aggregated AS (
        SELECT slot.tracking_period_id,
               slot.game_id,
               slot.appid,
               min(slot.player_count) FILTER (WHERE slot.outcome = 'success' AND slot.raw_status = 'success') AS min_players,
               max(slot.player_count) FILTER (WHERE slot.outcome = 'success' AND slot.raw_status = 'success') AS max_players,
               avg(slot.player_count::double precision) FILTER (WHERE slot.outcome = 'success' AND slot.raw_status = 'success') AS avg_players,
               percentile_cont(0.5) WITHIN GROUP (ORDER BY slot.player_count)
                   FILTER (WHERE slot.outcome = 'success' AND slot.raw_status = 'success') AS median_players,
               count(*)::integer AS expected_samples,
               count(*) FILTER (WHERE slot.outcome IN ('success', 'partial', 'failure'))::integer AS attempted_samples,
               count(*) FILTER (WHERE slot.outcome = 'success')::integer AS successful_samples,
               count(*) FILTER (WHERE slot.outcome = 'partial')::integer AS partial_samples,
               count(*) FILTER (WHERE slot.outcome = 'failure')::integer AS failed_samples,
               count(*) FILTER (WHERE slot.outcome = 'skipped')::integer AS skipped_samples,
               count(*) FILTER (WHERE slot.outcome = 'missed')::integer AS missed_samples,
               count(*) FILTER (WHERE slot.outcome = 'canceled')::integer AS canceled_samples,
               count(*) FILTER (WHERE slot.outcome = 'unattempted')::integer AS unattempted_samples,
               min(slot.collected_at) AS first_observed_at,
               max(slot.collected_at) AS last_observed_at,
               avg(slot.observation_lag_ms)::bigint AS avg_observation_lag_ms,
               max(slot.observation_lag_ms) AS max_observation_lag_ms
        FROM pg_temp.gfg_player_fact_slots slot
        GROUP BY slot.tracking_period_id, slot.game_id, slot.appid
    )
    INSERT INTO public.gfg_game_player_daily (
        tracking_period_id, game_id, appid, fact_date,
        min_players, max_players, avg_players, median_players,
        expected_samples, attempted_samples, successful_samples, partial_samples,
        failed_samples, skipped_samples, missed_samples, canceled_samples,
        unattempted_samples, failure_kind_counts, first_observed_at,
        last_observed_at, avg_observation_lag_ms, max_observation_lag_ms,
        quality_basis, projection_version, finalized_at, created_at, updated_at
    )
    SELECT aggregate.tracking_period_id, aggregate.game_id, aggregate.appid, target_date,
           aggregate.min_players, aggregate.max_players, aggregate.avg_players, aggregate.median_players,
           CASE WHEN day_start >= cutover THEN aggregate.expected_samples END,
           aggregate.attempted_samples, aggregate.successful_samples, aggregate.partial_samples,
           aggregate.failed_samples,
           CASE WHEN day_start >= cutover THEN aggregate.skipped_samples END,
           CASE WHEN day_start >= cutover THEN aggregate.missed_samples END,
           CASE WHEN day_start >= cutover THEN aggregate.canceled_samples END,
           CASE WHEN day_start >= cutover THEN aggregate.unattempted_samples END,
           COALESCE(failure.counts, '{}'::jsonb), aggregate.first_observed_at,
           aggregate.last_observed_at, aggregate.avg_observation_lag_ms,
           aggregate.max_observation_lag_ms,
           CASE WHEN day_start >= cutover THEN 'acquisition_ledger' ELSE 'legacy_observed_only' END,
           1, transaction_timestamp(), transaction_timestamp(), transaction_timestamp()
    FROM aggregated aggregate
    LEFT JOIN failure_counts failure USING (tracking_period_id);

    GET DIAGNOSTICS daily_rows = ROW_COUNT;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.gfg_project_state_fact_day(target_date date)
RETURNS TABLE(game_rows bigint, price_rows bigint)
LANGUAGE plpgsql
AS $function$
DECLARE
    day_start timestamp with time zone := target_date::timestamp AT TIME ZONE 'UTC';
    day_end timestamp with time zone := (target_date + 1)::timestamp AT TIME ZONE 'UTC';
    eligible_count bigint;
BEGIN
    WITH eligible AS (
        SELECT period.*,
               LEAST(day_end, COALESCE(period.tracked_until, day_end)) AS terminal_at,
               period.tracked_until IS NULL OR period.tracked_until > day_end AS tracked_at_end
        FROM public.gfg_game_tracking_periods period
        WHERE period.tracked_from < day_end
          AND (period.tracked_until IS NULL OR period.tracked_until > day_start)
          AND period.tracking_basis = 'explicit'
    ), materialized AS (
        SELECT eligible.*,
               prior.game_id AS prior_game_id
        FROM eligible
        LEFT JOIN LATERAL (
            SELECT daily.*
            FROM public.gfg_game_daily daily
            WHERE daily.tracking_period_id = eligible.id
              AND daily.fact_date <= target_date
            ORDER BY daily.fact_date DESC
            LIMIT 1
        ) prior ON true
    )
    SELECT count(*), count(prior_game_id) INTO eligible_count, game_rows
    FROM materialized;

    IF game_rows <> eligible_count THEN
        RAISE EXCEPTION 'Game state source unavailable for %, expected %, materialized %', target_date, eligible_count, game_rows;
    END IF;

    WITH eligible AS (
        SELECT period.*,
               LEAST(day_end, COALESCE(period.tracked_until, day_end)) AS terminal_at,
               period.tracked_until IS NULL OR period.tracked_until > day_end AS tracked_at_end
        FROM public.gfg_game_tracking_periods period
        WHERE period.tracked_from < day_end
          AND (period.tracked_until IS NULL OR period.tracked_until > day_start)
          AND period.tracking_basis = 'explicit'
    ), latest AS (
        SELECT eligible.id AS target_tracking_period_id,
               eligible.appid AS target_appid,
               eligible.terminal_at AS terminal_snapshot_at,
               eligible.tracked_at_end AS terminal_tracked_at_end,
               prior.*
        FROM eligible
        JOIN LATERAL (
            SELECT daily.*
            FROM public.gfg_game_daily daily
            WHERE daily.tracking_period_id = eligible.id
              AND daily.fact_date <= target_date
            ORDER BY daily.fact_date DESC
            LIMIT 1
        ) prior ON true
    )
    INSERT INTO public.gfg_game_daily (
        game_id, fact_date, tracking_period_id, appid, snapshot_at, tracked_at_end,
        name, name_en, view_count, game_type, is_free, windows, mac, linux,
        release_availability, release_precision, release_exact_date, release_year,
        release_month, release_quarter, release_window_start, release_window_end,
        release_observed_at, first_available_precision, first_available_exact_date,
        first_available_year, first_available_month, first_available_quarter,
        first_available_window_start, first_available_window_end,
        first_available_source, first_available_inferred, language_codes,
        unknown_language_names, full_audio_language_codes, languages_observed_at,
        developers, publishers, primary_tag_id, secondary_tag_id, tag_ids,
        details_observed_at, materialization_source, projection_version,
        finalized_at, created_at, updated_at
    )
    SELECT latest.game_id, target_date, latest.target_tracking_period_id, latest.target_appid,
           latest.terminal_snapshot_at, latest.terminal_tracked_at_end,
           latest.name, latest.name_en, latest.view_count, latest.game_type,
           latest.is_free, latest.windows, latest.mac, latest.linux,
           latest.release_availability, latest.release_precision,
           latest.release_exact_date, latest.release_year, latest.release_month,
           latest.release_quarter, latest.release_window_start,
           latest.release_window_end, latest.release_observed_at,
           latest.first_available_precision, latest.first_available_exact_date,
           latest.first_available_year, latest.first_available_month,
           latest.first_available_quarter, latest.first_available_window_start,
           latest.first_available_window_end, latest.first_available_source,
           latest.first_available_inferred, latest.language_codes,
           latest.unknown_language_names, latest.full_audio_language_codes,
           latest.languages_observed_at, latest.developers, latest.publishers,
           latest.primary_tag_id, latest.secondary_tag_id, latest.tag_ids,
           latest.details_observed_at,
           CASE WHEN latest.fact_date = target_date THEN latest.materialization_source
                ELSE 'carried_forward' END,
           1, transaction_timestamp(), latest.created_at, transaction_timestamp()
    FROM latest
    ON CONFLICT (game_id, fact_date) DO UPDATE
    SET tracking_period_id = EXCLUDED.tracking_period_id,
        appid = EXCLUDED.appid,
        snapshot_at = EXCLUDED.snapshot_at,
        tracked_at_end = EXCLUDED.tracked_at_end,
        name = EXCLUDED.name,
        name_en = EXCLUDED.name_en,
        view_count = EXCLUDED.view_count,
        game_type = EXCLUDED.game_type,
        is_free = EXCLUDED.is_free,
        windows = EXCLUDED.windows,
        mac = EXCLUDED.mac,
        linux = EXCLUDED.linux,
        release_availability = EXCLUDED.release_availability,
        release_precision = EXCLUDED.release_precision,
        release_exact_date = EXCLUDED.release_exact_date,
        release_year = EXCLUDED.release_year,
        release_month = EXCLUDED.release_month,
        release_quarter = EXCLUDED.release_quarter,
        release_window_start = EXCLUDED.release_window_start,
        release_window_end = EXCLUDED.release_window_end,
        release_observed_at = EXCLUDED.release_observed_at,
        first_available_precision = EXCLUDED.first_available_precision,
        first_available_exact_date = EXCLUDED.first_available_exact_date,
        first_available_year = EXCLUDED.first_available_year,
        first_available_month = EXCLUDED.first_available_month,
        first_available_quarter = EXCLUDED.first_available_quarter,
        first_available_window_start = EXCLUDED.first_available_window_start,
        first_available_window_end = EXCLUDED.first_available_window_end,
        first_available_source = EXCLUDED.first_available_source,
        first_available_inferred = EXCLUDED.first_available_inferred,
        language_codes = EXCLUDED.language_codes,
        unknown_language_names = EXCLUDED.unknown_language_names,
        full_audio_language_codes = EXCLUDED.full_audio_language_codes,
        languages_observed_at = EXCLUDED.languages_observed_at,
        developers = EXCLUDED.developers,
        publishers = EXCLUDED.publishers,
        primary_tag_id = EXCLUDED.primary_tag_id,
        secondary_tag_id = EXCLUDED.secondary_tag_id,
        tag_ids = EXCLUDED.tag_ids,
        details_observed_at = EXCLUDED.details_observed_at,
        materialization_source = EXCLUDED.materialization_source,
        projection_version = EXCLUDED.projection_version,
        finalized_at = transaction_timestamp(),
        updated_at = transaction_timestamp();

    GET DIAGNOSTICS game_rows = ROW_COUNT;

    WITH eligible AS (
        SELECT period.*
        FROM public.gfg_game_tracking_periods period
        WHERE period.tracked_from < day_end
          AND (period.tracked_until IS NULL OR period.tracked_until > day_start)
          AND period.tracking_basis = 'explicit'
    ), candidates AS (
        SELECT period.id AS tracking_period_id,
               period.game_id,
               period.appid,
               region.region,
               COALESCE(today.price_state, prior.price_state, current.price_state, 'unknown') AS price_state,
               COALESCE(today.currency, prior.currency, NULLIF(current.currency, '')) AS currency,
               COALESCE(today.initial_amount, prior.initial_amount, current.initial_amount) AS initial_amount,
               COALESCE(today.final_amount, prior.final_amount, current.final_amount) AS final_amount,
               COALESCE(today.discount_percent, prior.discount_percent, current.discount_percent::integer) AS discount_percent,
               COALESCE(today.observed_at, prior.observed_at, current.collected_at) AS observed_at,
               CASE WHEN today.tracking_period_id IS NOT NULL THEN today.materialization_source
                    WHEN prior.tracking_period_id IS NOT NULL THEN 'carried_forward'
                    ELSE 'bootstrap' END AS source_kind
        FROM eligible period
        CROSS JOIN (VALUES ('CN'::text), ('US'::text), ('HK'::text)) region(region)
        LEFT JOIN public.gfg_game_price_daily today
          ON today.tracking_period_id = period.id
         AND today.region = region.region
         AND today.fact_date = target_date
        LEFT JOIN LATERAL (
            SELECT daily.*
            FROM public.gfg_game_price_daily daily
            WHERE daily.tracking_period_id = period.id
              AND daily.region = region.region
              AND daily.fact_date < target_date
            ORDER BY daily.fact_date DESC
            LIMIT 1
        ) prior ON today.tracking_period_id IS NULL
        LEFT JOIN public.gfg_game_prices current
          ON current.game_id = period.game_id
         AND current.appid = period.appid
         AND current.region = region.region
         AND today.tracking_period_id IS NULL
         AND prior.tracking_period_id IS NULL
    )
    INSERT INTO public.gfg_game_price_daily (
        tracking_period_id, game_id, appid, region, fact_date, price_state,
        currency, initial_amount, final_amount, discount_percent, observed_at,
        materialization_source, projection_version, finalized_at,
        created_at, updated_at
    )
    SELECT tracking_period_id, game_id, appid, region, target_date, price_state,
           CASE WHEN price_state = 'priced' THEN currency END,
           CASE WHEN price_state = 'priced' THEN initial_amount END,
           CASE WHEN price_state = 'priced' THEN final_amount END,
           CASE WHEN price_state = 'priced' THEN discount_percent END,
           observed_at, source_kind, 1, transaction_timestamp(),
           transaction_timestamp(), transaction_timestamp()
    FROM candidates
    ON CONFLICT (tracking_period_id, region, fact_date) DO UPDATE
    SET game_id = EXCLUDED.game_id,
        appid = EXCLUDED.appid,
        price_state = EXCLUDED.price_state,
        currency = EXCLUDED.currency,
        initial_amount = EXCLUDED.initial_amount,
        final_amount = EXCLUDED.final_amount,
        discount_percent = EXCLUDED.discount_percent,
        observed_at = EXCLUDED.observed_at,
        materialization_source = EXCLUDED.materialization_source,
        projection_version = EXCLUDED.projection_version,
        finalized_at = transaction_timestamp(),
        updated_at = transaction_timestamp();

    GET DIAGNOSTICS price_rows = ROW_COUNT;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

-- +goose Down

DROP FUNCTION public.gfg_project_state_fact_day(date);
DROP FUNCTION public.gfg_project_player_fact_day(date);
