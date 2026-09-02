package changes

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gofurry/gofurry-game-collector/roof/env"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestGameChangeProjectorsIntegration(t *testing.T) {
	config := os.Getenv("GOFURRY_GAME_COLLECTOR_INTEGRATION_CONFIG")
	if config == "" {
		t.Skip("GOFURRY_GAME_COLLECTOR_INTEGRATION_CONFIG is not set")
	}
	if err := env.LoadServerConfig(config); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, env.GetServerConfig().DataBase.ConnectionString())
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	day1 := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	day2 := day1.AddDate(0, 0, 1)
	day3 := day2.AddDate(0, 0, 1)
	day4 := day3.AddDate(0, 0, 1)
	day5 := day4.AddDate(0, 0, 1)
	day6 := day5.AddDate(0, 0, 1)
	var period1, period2 int64
	if err := tx.QueryRow(ctx, `INSERT INTO gfg_game_tracking_periods(game_id,appid,tracked_from,tracked_until,tracking_basis,opened_reason,closed_reason) VALUES(870001,870001,$1,$2,'legacy_observed','change_test','change_test') RETURNING id`, day1, day4).Scan(&period1); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, `INSERT INTO gfg_game_tracking_periods(game_id,appid,tracked_from,tracked_until,tracking_basis,opened_reason,closed_reason) VALUES(870001,870002,$1,$2,'legacy_observed','change_test','change_test') RETURNING id`, day4, day4.AddDate(0, 0, 3)).Scan(&period2); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []struct {
		day    time.Time
		period int64
	}{{day1, period1}, {day2, period1}, {day3, period1}, {day4, period2}, {day5, period2}, {day6, period2}} {
		if _, err := tx.Exec(ctx, `INSERT INTO gfg_game_daily(game_id,fact_date,tracking_period_id,appid,snapshot_at,tracked_at_end,name,name_en,view_count,developers,publishers,tag_ids,materialization_source,projection_version,finalized_at) VALUES(870001,$1::date,$2,870001,$1::date+interval '23 hours',true,'Historical Change Game','Historical Change Game',0,ARRAY[]::text[],ARRAY[]::text[],ARRAY[]::bigint[],'observed',1,$1::date+interval '24 hours')`, fixture.day, fixture.period); err != nil {
			t.Fatal(err)
		}
	}
	for _, metric := range []string{"free_game_share", "windows_support", "linux_support", "mac_support"} {
		for _, fixture := range []struct {
			day           time.Time
			state, reason string
		}{{day1, "negative", "before"}, {day2, "unknown", "gap"}, {day3, "positive", "after"}, {day4, "negative", "new_identity"}} {
			if _, err := tx.Exec(ctx, `INSERT INTO gfg_metric_entity_daily(metric_key,metric_version,fact_date,game_id,state,reason_code,source_observed_at,dimension_values,source_projection_versions,evaluated_at) VALUES($1,1,$2::date,870001,$3,$4,$2::date+interval '12 hours','{}','{"gfg_game_daily":1}',$2::date+interval '24 hours')`, metric, fixture.day, fixture.state, fixture.reason); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, detector := range []string{"free_game_transition", "windows_support_transition", "linux_support_transition"} {
		assertProjectCount(t, ctx, tx, "gfg_project_change_day", detector, day3, 1)
		assertProjectCount(t, ctx, tx, "gfg_project_change_day", detector, day4, 0)
	}
	var macEvents int64
	if err := tx.QueryRow(ctx, `SELECT gfg_project_mac_change_day($1::date)`, day3).Scan(&macEvents); err != nil || macEvents != 0 {
		t.Fatalf("mac unknown-to-supported false positive events=%d err=%v", macEvents, err)
	}
	if err := tx.QueryRow(ctx, `SELECT gfg_project_mac_change_day($1::date)`, day4).Scan(&macEvents); err != nil || macEvents != 0 {
		t.Fatalf("mac tracking-period break events=%d err=%v", macEvents, err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO gfg_metric_entity_daily(metric_key,metric_version,fact_date,game_id,state,reason_code,source_observed_at,dimension_values,source_projection_versions,evaluated_at) VALUES('mac_support',1,$1::date,870001,'positive','reliable_transition',$1::date+interval '12 hours','{}','{"gfg_game_daily":1}',$1::date+interval '24 hours')`, day5); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, `SELECT gfg_project_mac_change_day($1::date)`, day5).Scan(&macEvents); err != nil || macEvents != 1 {
		t.Fatalf("mac reliable same-period transition events=%d err=%v", macEvents, err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO gfg_metric_entity_daily(metric_key,metric_version,fact_date,game_id,state,reason_code,source_observed_at,dimension_values,source_projection_versions,evaluated_at) VALUES('mac_support',1,$1::date,870001,'negative','reliable_transition',$1::date+interval '12 hours','{}','{"gfg_game_daily":1}',$1::date+interval '24 hours')`, day6); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, `SELECT gfg_project_mac_change_day($1::date)`, day6).Scan(&macEvents); err != nil || macEvents != 1 {
		t.Fatalf("mac reliable removal events=%d err=%v", macEvents, err)
	}
	var macRemoval string
	if err := tx.QueryRow(ctx, `SELECT event_code FROM gfg_change_events WHERE detector_key='mac_support_transition' AND projection_date=$1`, day6).Scan(&macRemoval); err != nil || macRemoval != "mac_support_removed" {
		t.Fatalf("mac removal code=%q err=%v", macRemoval, err)
	}
	for _, fixture := range []struct {
		day      time.Time
		state    string
		final    *int64
		discount *int32
	}{{day1, "priced", ptr64(1000), ptr32(0)}, {day2, "unknown", nil, nil}, {day3, "priced", ptr64(800), ptr32(20)}} {
		if _, err := tx.Exec(ctx, `INSERT INTO gfg_game_price_daily(tracking_period_id,game_id,appid,region,fact_date,price_state,currency,initial_amount,final_amount,discount_percent,observed_at,materialization_source,projection_version,finalized_at) VALUES($1,870001,870001,'CN',$2::date,$3,CASE WHEN $3='priced' THEN 'CNY' END,CASE WHEN $3='priced' THEN 1000 END,$4,$5,$2::date+interval '12 hours','observed',1,$2::date+interval '24 hours')`, period1, fixture.day, fixture.state, fixture.final, fixture.discount); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO gfg_game_price_daily(tracking_period_id,game_id,appid,region,fact_date,price_state,currency,initial_amount,final_amount,discount_percent,observed_at,materialization_source,projection_version,finalized_at) VALUES($1,870001,870001,'US',$2::date,'priced','USD',1200,1000,10,$2::date+interval '12 hours','observed',1,$2::date+interval '24 hours'),($1,870001,870001,'US',$3::date,'free',NULL,NULL,NULL,NULL,$3::date+interval '12 hours','observed',1,$3::date+interval '24 hours')`, period1, day1, day3); err != nil {
		t.Fatal(err)
	}
	assertProjectCount(t, ctx, tx, "gfg_project_change_day", "game_price_transition", day3, 3)
	var usEvents int64
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM gfg_change_events WHERE detector_key='game_price_transition' AND projection_date=$1 AND scope_key='US' AND event_code='game_price_state_changed'`, day3).Scan(&usEvents); err != nil || usEvents != 1 {
		t.Fatalf("price priority events=%d err=%v", usEvents, err)
	}
	for index, availability := range []string{"upcoming", "available", "upcoming"} {
		observed := day3.Add(time.Duration(index+1) * time.Hour)
		if _, err := tx.Exec(ctx, `INSERT INTO gfg_game_release_history(game_id,availability,precision,raw_text,source,source_region,source_locale,normalizer_version,observed_at) VALUES(870001,$1,'unknown','ignored','test','US','english','v1',$2)`, availability, observed); err != nil {
			t.Fatal(err)
		}
	}
	for index, availability := range []string{"upcoming", "available"} {
		if _, err := tx.Exec(ctx, `INSERT INTO gfg_game_release_history(game_id,availability,precision,raw_text,source,source_region,source_locale,normalizer_version,observed_at) VALUES(870999,$1,'unknown','ignored','test','US','english','v1',$2)`, availability, day3.Add(time.Duration(index+1)*time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	assertProjectCount(t, ctx, tx, "gfg_project_change_day", "game_release_transition", day3, 2)
	var firstKeys, secondKeys []string
	if err := rowsToStrings(ctx, tx, `SELECT event_key FROM gfg_change_events WHERE detector_key='game_release_transition' ORDER BY event_key`, &firstKeys); err != nil {
		t.Fatal(err)
	}
	assertProjectCount(t, ctx, tx, "gfg_project_change_day", "game_release_transition", day3, 2)
	if err := rowsToStrings(ctx, tx, `SELECT event_key FROM gfg_change_events WHERE detector_key='game_release_transition' ORDER BY event_key`, &secondKeys); err != nil {
		t.Fatal(err)
	}
	if len(firstKeys) != len(secondKeys) {
		t.Fatalf("idempotent keys length %d != %d", len(firstKeys), len(secondKeys))
	}
	for i := range firstKeys {
		if firstKeys[i] != secondKeys[i] {
			t.Fatalf("event key changed: %q != %q", firstKeys[i], secondKeys[i])
		}
	}
	if _, err := tx.Exec(ctx, "SAVEPOINT undeclared_code"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE gfg_change_registry SET event_codes=ARRAY['game_release_plan_changed'] WHERE detector_key='game_release_transition' AND detector_version=1`); err != nil {
		t.Fatal(err)
	}
	var ignored int64
	if err := tx.QueryRow(ctx, `SELECT gfg_project_change_day('game_release_transition',1,$1::date)`, day3).Scan(&ignored); err == nil {
		t.Fatal("undeclared release event code was accepted")
	}
	if _, err := tx.Exec(ctx, "ROLLBACK TO SAVEPOINT undeclared_code"); err != nil {
		t.Fatal(err)
	}
}

func TestGameChangeCheckpointLockSerializes(t *testing.T) {
	config := os.Getenv("GOFURRY_GAME_COLLECTOR_INTEGRATION_CONFIG")
	if config == "" {
		t.Skip("GOFURRY_GAME_COLLECTOR_INTEGRATION_CONFIG is not set")
	}
	if err := env.LoadServerConfig(config); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, env.GetServerConfig().DataBase.ConnectionString())
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	first, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Rollback(ctx)
	if _, err := first.Exec(ctx, `SELECT detector_key FROM gfg_change_checkpoints WHERE detector_key='free_game_transition' AND detector_version=1 FOR UPDATE`); err != nil {
		t.Fatal(err)
	}
	waitCtx, cancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer cancel()
	second, err := pool.Begin(waitCtx)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Rollback(context.Background())
	if _, err := second.Exec(waitCtx, `SELECT detector_key FROM gfg_change_checkpoints WHERE detector_key='free_game_transition' AND detector_version=1 FOR UPDATE`); err == nil {
		t.Fatal("second checkpoint lock did not wait for the first transaction")
	}
}

func TestGameChangeEngineBackfillRebuildIntegration(t *testing.T) {
	config := os.Getenv("GOFURRY_GAME_COLLECTOR_INTEGRATION_CONFIG")
	if config == "" {
		t.Skip("GOFURRY_GAME_COLLECTOR_INTEGRATION_CONFIG is not set")
	}
	if err := env.LoadServerConfig(config); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, env.GetServerConfig().DataBase.ConnectionString())
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var originalMetricSource, originalMetricProcessed, originalChangeSource, originalChangeProcessed pgtype.Date
	if err := pool.QueryRow(ctx, `SELECT source_start_date,processed_through FROM gfg_metric_checkpoints WHERE metric_key='free_game_share' AND metric_version=1`).Scan(&originalMetricSource, &originalMetricProcessed); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT source_start_date,processed_through FROM gfg_change_checkpoints WHERE detector_key='free_game_transition' AND detector_version=1`).Scan(&originalChangeSource, &originalChangeProcessed); err != nil {
		t.Fatal(err)
	}
	day1 := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	day3 := day1.AddDate(0, 0, 2)
	var period int64
	if err := pool.QueryRow(ctx, `INSERT INTO gfg_game_tracking_periods(game_id,appid,tracked_from,tracked_until,tracking_basis,opened_reason,closed_reason) VALUES(870101,870101,$1,$2,'legacy_observed','engine_smoke','engine_smoke') RETURNING id`, day1, day3.AddDate(0, 0, 1)).Scan(&period); err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM gfg_change_events WHERE detector_key='free_game_transition' AND game_id=870101;DELETE FROM gfg_metric_entity_daily WHERE game_id=870101;DELETE FROM gfg_game_daily WHERE game_id=870101;DELETE FROM gfg_game_tracking_periods WHERE id=$1;UPDATE gfg_metric_checkpoints SET source_start_date=$2,processed_through=$3 WHERE metric_key='free_game_share' AND metric_version=1;UPDATE gfg_change_checkpoints SET source_start_date=$4,processed_through=$5 WHERE detector_key='free_game_transition' AND detector_version=1`, pgx.QueryExecModeSimpleProtocol, period, originalMetricSource, originalMetricProcessed, originalChangeSource, originalChangeProcessed)
	}
	defer cleanup()
	for _, fixture := range []struct {
		day   time.Time
		state string
	}{{day1, "negative"}, {day3, "positive"}} {
		if _, err := pool.Exec(ctx, `INSERT INTO gfg_game_daily(game_id,fact_date,tracking_period_id,appid,snapshot_at,tracked_at_end,name,name_en,view_count,developers,publishers,tag_ids,materialization_source,projection_version,finalized_at) VALUES(870101,$1::date,$2,870101,$1::date+interval '23 hours',true,'Engine Smoke Game','Engine Smoke Game',0,'{}','{}','{}','observed',1,$1::date+interval '24 hours');INSERT INTO gfg_metric_entity_daily(metric_key,metric_version,fact_date,game_id,state,reason_code,source_observed_at,dimension_values,source_projection_versions,evaluated_at) VALUES('free_game_share',1,$1::date,870101,$3,'fixture',$1::date+interval '12 hours','{}','{"gfg_game_daily":1}',$1::date+interval '24 hours')`, pgx.QueryExecModeSimpleProtocol, fixture.day, period, fixture.state); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := pool.Exec(ctx, `UPDATE gfg_metric_checkpoints SET source_start_date=$1::date,processed_through=$2::date WHERE metric_key='free_game_share' AND metric_version=1;UPDATE gfg_change_checkpoints SET source_start_date=$1::date,processed_through=$3::date WHERE detector_key='free_game_transition' AND detector_version=1`, pgx.QueryExecModeSimpleProtocol, day1, day3, day1.AddDate(0, 0, 1)); err != nil {
		t.Fatal(err)
	}
	engine := New(pool, Options{})
	through := day3
	summary, err := engine.Backfill(ctx, BackfillOptions{Detector: "free_game_transition", Version: 1, Through: &through, MaxDays: 1})
	if err != nil || summary.Processed != 1 {
		t.Fatalf("backfill summary=%+v err=%v", summary, err)
	}
	rebuilt, err := engine.Rebuild(ctx, "free_game_transition", 1, day3, nil, 0, false)
	if err != nil || rebuilt.Processed != 1 {
		t.Fatalf("rebuild summary=%+v err=%v", rebuilt, err)
	}
	var events int64
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM gfg_change_events WHERE detector_key='free_game_transition' AND game_id=870101`).Scan(&events); err != nil || events != 1 {
		t.Fatalf("events=%d err=%v", events, err)
	}
}

func TestMacChangeBackfillAndRebuildIntegration(t *testing.T) {
	config := os.Getenv("GOFURRY_GAME_COLLECTOR_INTEGRATION_CONFIG")
	if config == "" {
		t.Skip("GOFURRY_GAME_COLLECTOR_INTEGRATION_CONFIG is not set")
	}
	if err := env.LoadServerConfig(config); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, env.GetServerConfig().DataBase.ConnectionString())
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var originalMetricSource, originalMetricProcessed, originalChangeSource, originalChangeProcessed pgtype.Date
	if err := pool.QueryRow(ctx, `SELECT source_start_date,processed_through FROM gfg_metric_checkpoints WHERE metric_key='mac_support' AND metric_version=1`).Scan(&originalMetricSource, &originalMetricProcessed); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT source_start_date,processed_through FROM gfg_change_checkpoints WHERE detector_key='mac_support_transition' AND detector_version=1`).Scan(&originalChangeSource, &originalChangeProcessed); err != nil {
		t.Fatal(err)
	}
	day1 := time.Date(2040, 2, 1, 0, 0, 0, 0, time.UTC)
	day2 := day1.AddDate(0, 0, 1)
	var period int64
	if err := pool.QueryRow(ctx, `INSERT INTO gfg_game_tracking_periods(game_id,appid,tracked_from,tracking_basis,opened_reason) VALUES(870201,870201,$1,'explicit','mac_change_backfill_test') RETURNING id`, day1).Scan(&period); err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM gfg_change_events WHERE detector_key='mac_support_transition' AND game_id=870201;DELETE FROM gfg_metric_entity_daily WHERE metric_key='mac_support' AND metric_version=1 AND game_id=870201;DELETE FROM gfg_game_daily WHERE game_id=870201;DELETE FROM gfg_game_tracking_periods WHERE id=$1;UPDATE gfg_metric_checkpoints SET source_start_date=$2,processed_through=$3 WHERE metric_key='mac_support' AND metric_version=1;UPDATE gfg_change_checkpoints SET source_start_date=$4,processed_through=$5 WHERE detector_key='mac_support_transition' AND detector_version=1`, pgx.QueryExecModeSimpleProtocol, period, originalMetricSource, originalMetricProcessed, originalChangeSource, originalChangeProcessed)
	}
	defer cleanup()
	for _, fixture := range []struct {
		day   time.Time
		state string
	}{{day1, "negative"}, {day2, "positive"}} {
		if _, err := pool.Exec(ctx, `INSERT INTO gfg_game_daily(game_id,fact_date,tracking_period_id,appid,snapshot_at,tracked_at_end,name,name_en,view_count,mac,details_observed_at,tag_ids,developers,publishers,materialization_source,projection_version,finalized_at) VALUES(870201,$1,$2,870201,$1::date+interval '20 hours',true,'Mac Change Backfill','Mac Change Backfill',0,$3='positive',$1::date+interval '12 hours',ARRAY[]::bigint[],ARRAY[]::text[],ARRAY[]::text[],'observed',1,$1::date+interval '24 hours');INSERT INTO gfg_metric_entity_daily(metric_key,metric_version,fact_date,game_id,state,reason_code,source_observed_at,dimension_values,source_projection_versions,evaluated_at) VALUES('mac_support',1,$1,870201,$3,'fixture',$1::date+interval '12 hours','{}','{"gfg_game_daily":1}',$1::date+interval '24 hours')`, pgx.QueryExecModeSimpleProtocol, fixture.day, period, fixture.state); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := pool.Exec(ctx, `UPDATE gfg_metric_checkpoints SET source_start_date=$1::date,processed_through=$2::date WHERE metric_key='mac_support' AND metric_version=1;UPDATE gfg_change_checkpoints SET source_start_date=$1::date,processed_through=NULL WHERE detector_key='mac_support_transition' AND detector_version=1`, pgx.QueryExecModeSimpleProtocol, day1, day2); err != nil {
		t.Fatal(err)
	}
	engine := New(pool, Options{})
	summary, err := engine.Backfill(ctx, BackfillOptions{Detector: "mac_support_transition", Version: 1, Through: &day2, MaxDays: 2})
	if err != nil || summary.Processed != 2 {
		t.Fatalf("Mac Change backfill summary=%+v err=%v", summary, err)
	}
	rebuilt, err := engine.Rebuild(ctx, "mac_support_transition", 1, day2, &day2, 0, false)
	if err != nil || rebuilt.Processed != 1 {
		t.Fatalf("Mac Change rebuild summary=%+v err=%v", rebuilt, err)
	}
	var eventCode string
	if err := pool.QueryRow(ctx, `SELECT event_code FROM gfg_change_events WHERE detector_key='mac_support_transition' AND game_id=870201 AND projection_date=$1`, day2).Scan(&eventCode); err != nil || eventCode != "mac_support_added" {
		t.Fatalf("Mac Change backfill event=%q err=%v", eventCode, err)
	}
}

func assertProjectCount(t *testing.T, ctx context.Context, tx pgx.Tx, function, key string, day time.Time, want int64) {
	t.Helper()
	var got int64
	if err := tx.QueryRow(ctx, `SELECT `+function+`($1,1,$2::date)`, key, day).Scan(&got); err != nil {
		t.Fatalf("project %s: %v", key, err)
	}
	if got != want {
		t.Fatalf("project %s count=%d want=%d", key, got, want)
	}
}
func rowsToStrings(ctx context.Context, tx pgx.Tx, query string, target *[]string) error {
	rows, err := tx.Query(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return err
		}
		*target = append(*target, value)
	}
	return rows.Err()
}
func ptr64(value int64) *int64 { return &value }
func ptr32(value int32) *int32 { return &value }
