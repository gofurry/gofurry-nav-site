-- name: OpenGameTrackingPeriod :one
INSERT INTO gfg_game_tracking_periods (
    game_id, appid, tracked_from, tracking_basis, opened_reason
)
VALUES (
    sqlc.arg(game_id), sqlc.arg(appid), transaction_timestamp(),
    'explicit', sqlc.arg(opened_reason)
)
RETURNING *;

-- name: CloseGameTrackingPeriod :one
UPDATE gfg_game_tracking_periods
SET tracked_until = GREATEST(transaction_timestamp(), tracked_from + interval '1 microsecond'),
    closed_reason = sqlc.arg(closed_reason),
    updated_at = transaction_timestamp()
WHERE game_id = sqlc.arg(game_id)
  AND tracking_basis = 'explicit'
  AND tracked_until IS NULL
RETURNING *;

-- name: RefreshCurrentGameDaily :exec
SELECT gfg_refresh_current_game_daily(
    sqlc.arg(game_id)::bigint,
    sqlc.arg(materialization_source)::text
);
