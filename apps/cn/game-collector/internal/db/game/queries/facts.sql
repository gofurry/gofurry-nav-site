-- name: ListGameFactCheckpoints :many
SELECT pipeline_key, projection_version, source_start_date, processed_through,
       quality_cutover_at, created_at, updated_at
FROM gfg_fact_rollup_checkpoints
ORDER BY pipeline_key;

-- name: LockGameFactCheckpoint :one
SELECT pipeline_key, projection_version, source_start_date, processed_through,
       quality_cutover_at, created_at, updated_at
FROM gfg_fact_rollup_checkpoints
WHERE pipeline_key = sqlc.arg(pipeline_key)
FOR UPDATE;

-- name: GameFactDaySettled :one
SELECT NOT EXISTS (
    SELECT 1
    FROM gfg_collection_jobs job
    WHERE job.trigger = 'scheduled'
      AND job.job_key = sqlc.arg(job_key)
      AND job.scheduled_for >= sqlc.arg(day_start)
      AND job.scheduled_for < sqlc.arg(day_end)
      AND job.status IN ('queued', 'running')
)
AND NOT EXISTS (
    SELECT 1
    FROM gfg_collection_schedules schedule
    WHERE schedule.job_key = sqlc.arg(job_key)
      AND schedule.enabled
      AND (schedule.next_scheduled_for IS NULL OR schedule.next_scheduled_for < sqlc.arg(day_end))
) AS settled;

-- name: ProjectGamePlayerFactDay :one
SELECT count(*)::bigint AS projected
FROM gfg_project_player_fact_day(sqlc.arg(fact_date)::date);

-- name: ProjectGameStateFactDay :one
SELECT count(*)::bigint AS projected
FROM gfg_project_state_fact_day(sqlc.arg(fact_date)::date);

-- name: AdvanceGameFactCheckpoint :execrows
UPDATE gfg_fact_rollup_checkpoints
SET processed_through = sqlc.arg(processed_through),
    updated_at = transaction_timestamp()
WHERE pipeline_key = sqlc.arg(pipeline_key)
  AND (processed_through IS NULL OR processed_through < sqlc.arg(processed_through));

-- name: GameFactSourceBounds :one
SELECT min(collected_at AT TIME ZONE 'UTC')::date AS player_raw_start,
       max(collected_at AT TIME ZONE 'UTC')::date AS player_raw_end
FROM gfg_game_player_counts;

-- name: CountGamePlayerFactRows :one
SELECT count(*)::bigint
FROM gfg_game_player_daily
WHERE fact_date >= sqlc.arg(start_date) AND fact_date < sqlc.arg(end_date);

-- name: CountGameStateFactRows :one
SELECT count(*)::bigint
FROM gfg_game_daily
WHERE fact_date >= sqlc.arg(start_date) AND fact_date < sqlc.arg(end_date);

-- name: GamePlayerRebuildSourceAvailable :one
SELECT NOT EXISTS (
    SELECT 1
    FROM gfg_game_player_daily fact
    WHERE fact.fact_date = sqlc.arg(fact_date)
      AND (
          (fact.successful_samples > 0 AND (
              SELECT count(*)
              FROM gfg_game_player_counts raw
              JOIN gfg_collection_runs run ON run.id = raw.run_id
              JOIN gfg_collection_jobs job ON job.id = run.job_id
              WHERE raw.game_id = fact.game_id
                AND raw.appid = fact.appid
                AND raw.status = 'success'
                AND job.trigger = 'scheduled'
                AND job.job_key = 'game.players'
                AND (job.scheduled_for AT TIME ZONE 'UTC')::date = fact.fact_date
          ) < fact.successful_samples)
          OR
          (fact.attempted_samples > 0 AND (
              SELECT count(*)
              FROM gfg_collection_task_results result
              JOIN gfg_collection_runs run ON run.id = result.run_id
              JOIN gfg_collection_jobs job ON job.id = run.job_id
              WHERE result.game_id = fact.game_id
                AND result.appid = fact.appid
                AND result.task_type = 'players'
                AND job.trigger = 'scheduled'
                AND job.job_key = 'game.players'
                AND (job.scheduled_for AT TIME ZONE 'UTC')::date = fact.fact_date
          ) < fact.attempted_samples)
      )
) AS available;

-- name: PruneGamePlayerRawBatch :execrows
WITH eligible AS (
    SELECT raw.id
    FROM gfg_game_player_counts raw
    WHERE raw.collected_at < sqlc.arg(older_than)
      AND EXISTS (
          SELECT 1
          FROM gfg_fact_rollup_checkpoints checkpoint
          WHERE checkpoint.pipeline_key = 'game.player_facts'
            AND checkpoint.processed_through IS NOT NULL
            AND (raw.collected_at AT TIME ZONE 'UTC')::date <= checkpoint.processed_through
      )
    ORDER BY raw.id
    LIMIT sqlc.arg(batch_size)
    FOR UPDATE SKIP LOCKED
)
DELETE FROM gfg_game_player_counts raw
USING eligible
WHERE raw.id = eligible.id;
