package metrics

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

func TestPostgresMacMetricProjection(t *testing.T) {
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
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	day := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	var period int64
	if err := tx.QueryRow(ctx, `INSERT INTO gfg_game_tracking_periods(game_id,appid,tracked_from,tracking_basis,opened_reason) VALUES(880001,880001,$1,'explicit','mac_metric_test') RETURNING id`, day).Scan(&period); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO gfg_game_daily(game_id,fact_date,tracking_period_id,appid,snapshot_at,tracked_at_end,name,name_en,view_count,mac,details_observed_at,primary_tag_id,tag_ids,developers,publishers,materialization_source,projection_version,finalized_at) VALUES(880001,$1,$2,880001,$1::date+interval '20 hours',true,'Mac Fixture','Mac Fixture',0,true,$1::date+interval '12 hours',123,ARRAY[123,456]::bigint[],ARRAY[]::text[],ARRAY[]::text[],'observed',1,$1::date+interval '24 hours')`, day, period); err != nil {
		t.Fatal(err)
	}
	for index, fixture := range []struct {
		mac        *bool
		observedAt *time.Time
	}{{boolPointer(false), timePointer(day.Add(12 * time.Hour))}, {boolPointer(true), timePointer(day.AddDate(0, 0, -4))}, {nil, nil}} {
		gameID := int64(880002 + index)
		var fixturePeriod int64
		if err := tx.QueryRow(ctx, `INSERT INTO gfg_game_tracking_periods(game_id,appid,tracked_from,tracking_basis,opened_reason) VALUES($1,$1,$2,'explicit','mac_metric_test') RETURNING id`, gameID, day).Scan(&fixturePeriod); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO gfg_game_daily(game_id,fact_date,tracking_period_id,appid,snapshot_at,tracked_at_end,name,name_en,view_count,mac,details_observed_at,tag_ids,developers,publishers,materialization_source,projection_version,finalized_at) VALUES($1,$2,$3,$1,$2::date+interval '20 hours',true,'Mac Fixture','Mac Fixture',0,$4,$5,ARRAY[]::bigint[],ARRAY[]::text[],ARRAY[]::text[],'observed',1,$2::date+interval '24 hours')`, gameID, day, fixturePeriod, fixture.mac, fixture.observedAt); err != nil {
			t.Fatal(err)
		}
	}
	var count int64
	if err := tx.QueryRow(ctx, `SELECT gfg_project_mac_metric_day($1::date)`, day).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 4 {
		t.Fatalf("projected entities=%d", count)
	}
	var state string
	if err := tx.QueryRow(ctx, `SELECT state FROM gfg_metric_entity_daily WHERE metric_key='mac_support' AND metric_version=1 AND fact_date=$1 AND game_id=880001`, day).Scan(&state); err != nil || state != "positive" {
		t.Fatalf("state=%q err=%v", state, err)
	}
	var dimensions int64
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM gfg_metric_daily WHERE metric_key='mac_support' AND metric_version=1 AND fact_date=$1`, day).Scan(&dimensions); err != nil || dimensions != 5 {
		t.Fatalf("dimensions=%d err=%v", dimensions, err)
	}
	var population, positive, negative, stale, unknown int64
	if err := tx.QueryRow(ctx, `SELECT population_count,positive_count,negative_count,stale_count,unknown_count FROM gfg_metric_daily WHERE metric_key='mac_support' AND metric_version=1 AND fact_date=$1 AND dimension_key='global' AND dimension_value='all'`, day).Scan(&population, &positive, &negative, &stale, &unknown); err != nil || population != 4 || positive != 1 || negative != 1 || stale != 1 || unknown != 1 {
		t.Fatalf("global states population=%d positive=%d negative=%d stale=%d unknown=%d err=%v", population, positive, negative, stale, unknown, err)
	}
}

func TestPostgresMacMetricBackfillAndRebuild(t *testing.T) {
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
	var originalMetricSource, originalMetricProcessed, originalStateProcessed pgtype.Date
	if err := pool.QueryRow(ctx, `SELECT source_start_date,processed_through FROM gfg_metric_checkpoints WHERE metric_key='mac_support' AND metric_version=1`).Scan(&originalMetricSource, &originalMetricProcessed); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT processed_through FROM gfg_fact_rollup_checkpoints WHERE pipeline_key='game.state_facts'`).Scan(&originalStateProcessed); err != nil {
		t.Fatal(err)
	}
	day := time.Date(2040, 1, 1, 0, 0, 0, 0, time.UTC)
	var period int64
	if err := pool.QueryRow(ctx, `INSERT INTO gfg_game_tracking_periods(game_id,appid,tracked_from,tracking_basis,opened_reason) VALUES(880101,880101,$1,'explicit','mac_backfill_test') RETURNING id`, day).Scan(&period); err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM gfg_metric_daily WHERE metric_key='mac_support' AND metric_version=1 AND fact_date=$1;DELETE FROM gfg_metric_entity_daily WHERE metric_key='mac_support' AND metric_version=1 AND game_id=880101;DELETE FROM gfg_game_daily WHERE game_id=880101;DELETE FROM gfg_game_tracking_periods WHERE id=$2;UPDATE gfg_metric_checkpoints SET source_start_date=$3,processed_through=$4 WHERE metric_key='mac_support' AND metric_version=1;UPDATE gfg_fact_rollup_checkpoints SET processed_through=$5 WHERE pipeline_key='game.state_facts'`, pgx.QueryExecModeSimpleProtocol, day, period, originalMetricSource, originalMetricProcessed, originalStateProcessed)
	}
	defer cleanup()
	if _, err := pool.Exec(ctx, `INSERT INTO gfg_game_daily(game_id,fact_date,tracking_period_id,appid,snapshot_at,tracked_at_end,name,name_en,view_count,mac,details_observed_at,tag_ids,developers,publishers,materialization_source,projection_version,finalized_at) VALUES(880101,$1,$2,880101,$1::date+interval '20 hours',true,'Mac Backfill','Mac Backfill',0,true,$1::date+interval '12 hours',ARRAY[]::bigint[],ARRAY[]::text[],ARRAY[]::text[],'observed',1,$1::date+interval '24 hours');UPDATE gfg_fact_rollup_checkpoints SET processed_through=$1::date WHERE pipeline_key='game.state_facts';UPDATE gfg_metric_checkpoints SET source_start_date=$1::date,processed_through=NULL WHERE metric_key='mac_support' AND metric_version=1`, pgx.QueryExecModeSimpleProtocol, day, period); err != nil {
		t.Fatal(err)
	}
	engine := New(pool, Options{})
	summary, err := engine.Backfill(ctx, BackfillOptions{Metric: "mac_support", Version: 1, Through: &day, MaxDays: 1})
	if err != nil || summary.Processed != 1 {
		t.Fatalf("Mac Metric backfill summary=%+v err=%v", summary, err)
	}
	rebuilt, err := engine.Rebuild(ctx, "mac_support", 1, day, day, 1, false)
	if err != nil || rebuilt.Processed != 1 {
		t.Fatalf("Mac Metric rebuild summary=%+v err=%v", rebuilt, err)
	}
	var state string
	if err := pool.QueryRow(ctx, `SELECT state FROM gfg_metric_entity_daily WHERE metric_key='mac_support' AND metric_version=1 AND fact_date=$1 AND game_id=880101`, day).Scan(&state); err != nil || state != "positive" {
		t.Fatalf("Mac Metric backfill state=%q err=%v", state, err)
	}
}

func boolPointer(value bool) *bool           { return &value }
func timePointer(value time.Time) *time.Time { return &value }
