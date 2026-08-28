-- name: RefreshCollectedGameDaily :exec
SELECT gfg_refresh_current_game_daily(
    sqlc.arg(game_id)::bigint,
    sqlc.arg(materialization_source)::text
);

-- name: UpsertPriceDailyObserved :exec
INSERT INTO gfg_game_price_daily (
    tracking_period_id, game_id, appid, region, fact_date, price_state,
    currency, initial_amount, final_amount, discount_percent, observed_at,
    materialization_source, projection_version, finalized_at,
    created_at, updated_at
)
SELECT p.id,
       sqlc.arg(game_id),
       sqlc.arg(appid),
       sqlc.arg(region),
       (sqlc.arg(observed_at)::timestamptz AT TIME ZONE 'UTC')::date,
       sqlc.arg(price_state),
       CASE WHEN sqlc.arg(price_state)::text = 'priced' THEN NULLIF(sqlc.arg(currency)::text, '') END,
       CASE WHEN sqlc.arg(price_state)::text = 'priced' THEN sqlc.arg(initial_amount)::bigint END,
       CASE WHEN sqlc.arg(price_state)::text = 'priced' THEN sqlc.arg(final_amount)::bigint END,
       CASE WHEN sqlc.arg(price_state)::text = 'priced' THEN sqlc.arg(discount_percent)::integer END,
       sqlc.arg(observed_at),
       'observed',
       1,
       NULL,
       transaction_timestamp(),
       transaction_timestamp()
FROM gfg_game_tracking_periods p
WHERE p.game_id = sqlc.arg(game_id)
  AND p.appid = sqlc.arg(appid)
  AND p.tracking_basis = 'explicit'
  AND p.tracked_until IS NULL
ON CONFLICT (tracking_period_id, region, fact_date) DO UPDATE
SET appid = EXCLUDED.appid,
    price_state = EXCLUDED.price_state,
    currency = EXCLUDED.currency,
    initial_amount = EXCLUDED.initial_amount,
    final_amount = EXCLUDED.final_amount,
    discount_percent = EXCLUDED.discount_percent,
    observed_at = EXCLUDED.observed_at,
    materialization_source = EXCLUDED.materialization_source,
    projection_version = EXCLUDED.projection_version,
    updated_at = transaction_timestamp()
WHERE gfg_game_price_daily.finalized_at IS NULL
  AND (gfg_game_price_daily.observed_at IS NULL
       OR EXCLUDED.observed_at >= gfg_game_price_daily.observed_at);
