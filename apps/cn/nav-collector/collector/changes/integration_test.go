package changes

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-collector/roof/env"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestNavChangeProjectorsIntegration(t *testing.T) {
	config := os.Getenv("GOFURRY_NAV_COLLECTOR_INTEGRATION_CONFIG")
	if config == "" {
		t.Skip("GOFURRY_NAV_COLLECTOR_INTEGRATION_CONFIG is not set")
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
	var target1, target2, target3 int64
	for _, fixture := range []struct {
		target      string
		from, until time.Time
		id          *int64
	}{{"a.change.test", day1, day4, &target1}, {"b.change.test", day1, day4, &target2}, {"c.change.test", day4, day4.AddDate(0, 0, 2), &target3}} {
		if err := tx.QueryRow(ctx, `INSERT INTO gfn_target_tracking_periods(site_id,target,tracked_from,tracked_until,tracking_basis,opened_reason,closed_reason) VALUES(870001,$1,$2,$3,'legacy_observed','change_test','change_test') RETURNING id`, fixture.target, fixture.from, fixture.until).Scan(fixture.id); err != nil {
			t.Fatal(err)
		}
	}
	for _, fixture := range []struct {
		day     time.Time
		primary int64
	}{{day1, target1}, {day2, target1}, {day3, target1}, {day4, target3}} {
		if _, err := tx.Exec(ctx, `INSERT INTO gfn_site_daily(site_id,fact_date,snapshot_at,tracked_at_end,name,name_en,view_count,group_ids,primary_target_tracking_period_id,primary_target,primary_basis,active_target_count,projection_version,finalized_at) VALUES(870001,$1::date,$1::date+interval '23 hours',true,'Historical Change Site','Historical Change Site',0,ARRAY[]::bigint[],$2,'target','explicit',1,1,$1::date+interval '24 hours')`, fixture.day, fixture.primary); err != nil {
			t.Fatal(err)
		}
	}
	for _, metric := range []string{"ipv6_adoption", "tls13_adoption", "security_txt_adoption"} {
		for _, fixture := range []struct {
			day           time.Time
			state, reason string
		}{{day1, "negative", "before"}, {day2, "probe_failed", "gap"}, {day3, "positive", "after"}, {day4, "negative", "new_identity"}} {
			if _, err := tx.Exec(ctx, `INSERT INTO gfn_metric_entity_daily(metric_key,metric_version,fact_date,site_id,state,reason_code,source_observed_at,dimension_values,source_projection_versions,evaluated_at) VALUES($1,1,$2::date,870001,$3,$4,$2::date+interval '12 hours','{}','{"gfn_site_daily":1}',$2::date+interval '24 hours')`, metric, fixture.day, fixture.state, fixture.reason); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, detector := range []string{"ipv6_transition", "tls13_transition", "security_txt_transition"} {
		assertNavProjectCount(t, ctx, tx, detector, day3, 1)
		assertNavProjectCount(t, ctx, tx, detector, day4, 0)
	}
	// A -> B -> A within one day is two effective events, not a collapsed diff.
	t0 := day3.Add(2 * time.Hour)
	t1 := t0.Add(time.Hour)
	t2 := t1.Add(time.Hour)
	var primary1, primary2, primary3 int64
	if err := tx.QueryRow(ctx, `INSERT INTO gfn_site_primary_target_periods(site_id,target_tracking_period_id,effective_from,effective_until,basis) VALUES(870001,$1,$2,$3,'explicit') RETURNING id`, target1, t0, t1).Scan(&primary1); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, `INSERT INTO gfn_site_primary_target_periods(site_id,target_tracking_period_id,effective_from,effective_until,basis) VALUES(870001,$1,$2,$3,'explicit') RETURNING id`, target2, t1, t2).Scan(&primary2); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, `INSERT INTO gfn_site_primary_target_periods(site_id,target_tracking_period_id,effective_from,effective_until,basis) VALUES(870001,$1,$2,$3,'explicit') RETURNING id`, target1, t2, t2.Add(time.Hour)).Scan(&primary3); err != nil {
		t.Fatal(err)
	}
	var gapTarget1, gapTarget2 int64
	if err := tx.QueryRow(ctx, `INSERT INTO gfn_target_tracking_periods(site_id,target,tracked_from,tracked_until,tracking_basis,opened_reason,closed_reason) VALUES(870002,'gap-a.change.test',$1,$2,'legacy_observed','change_test','change_test') RETURNING id`, day1, day4).Scan(&gapTarget1); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, `INSERT INTO gfn_target_tracking_periods(site_id,target,tracked_from,tracked_until,tracking_basis,opened_reason,closed_reason) VALUES(870002,'gap-b.change.test',$1,$2,'legacy_observed','change_test','change_test') RETURNING id`, day1, day4).Scan(&gapTarget2); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO gfn_site_primary_target_periods(site_id,target_tracking_period_id,effective_from,effective_until,basis) VALUES(870002,$1,$2,$3,'explicit'),(870002,$4,$5,$6,'explicit')`, gapTarget1, t0, t1, gapTarget2, t2, t2.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	assertNavProjectCount(t, ctx, tx, "primary_target_transition", day3, 2)
	for _, fixture := range []struct {
		day         time.Time
		fingerprint string
		observed    *time.Time
	}{{day1, "aaa", timePtr(day1.Add(12 * time.Hour))}, {day2, "", nil}, {day3, "bbb", timePtr(day3.Add(12 * time.Hour))}} {
		if _, err := tx.Exec(ctx, `INSERT INTO gfn_site_target_daily(target_tracking_period_id,site_id,target,fact_date,snapshot_at,tracked_at_end,tls_state_observed_at,tls_cert_fingerprint_sha256,tls_cert_spki_sha256,tls_cert_issuer,tls_cert_not_before,tls_cert_not_after,tls_cert_verified,projection_version,finalized_at) VALUES($1,870001,'a.change.test',$2::date,$2::date+interval '23 hours',true,$3,NULLIF($4,''),'spki','issuer',$2::date,$2::date+interval '90 days',true,1,$2::date+interval '24 hours')`, target1, fixture.day, fixture.observed, fixture.fingerprint); err != nil {
			t.Fatal(err)
		}
	}
	assertNavProjectCount(t, ctx, tx, "tls_certificate_transition", day3, 1)
	assertNavProjectCount(t, ctx, tx, "tls_certificate_transition", day3, 1)
	if _, err := tx.Exec(ctx, `INSERT INTO gfn_site_target_daily(target_tracking_period_id,site_id,target,fact_date,snapshot_at,tracked_at_end,tls_state_observed_at,tls_cert_fingerprint_sha256,projection_version,finalized_at) VALUES($1,870001,'c.change.test',$2::date,$2::date+interval '23 hours',true,$2::date+interval '12 hours','ccc',1,$2::date+interval '24 hours')`, target3, day4); err != nil {
		t.Fatal(err)
	}
	assertNavProjectCount(t, ctx, tx, "tls_certificate_transition", day4, 0)
	var basis, eventAt, afterAt string
	if err := tx.QueryRow(ctx, `SELECT time_basis,event_at::text,source_after_at::text FROM gfn_change_events WHERE detector_key='tls_certificate_transition' AND projection_date=$1`, day3).Scan(&basis, &eventAt, &afterAt); err != nil {
		t.Fatal(err)
	}
	if basis != "observed" || eventAt != afterAt {
		t.Fatalf("certificate provenance basis=%s event=%s after=%s", basis, eventAt, afterAt)
	}
}
func assertNavProjectCount(t *testing.T, ctx context.Context, tx pgx.Tx, key string, day time.Time, want int64) {
	t.Helper()
	var got int64
	if err := tx.QueryRow(ctx, `SELECT gfn_project_change_day($1,1,$2::date)`, key, day).Scan(&got); err != nil {
		t.Fatalf("project %s: %v", key, err)
	}
	if got != want {
		t.Fatalf("project %s count=%d want=%d", key, got, want)
	}
}
func timePtr(value time.Time) *time.Time { return &value }

func TestNavChangeEngineBackfillRebuildIntegration(t *testing.T) {
	config := os.Getenv("GOFURRY_NAV_COLLECTOR_INTEGRATION_CONFIG")
	if config == "" {
		t.Skip("GOFURRY_NAV_COLLECTOR_INTEGRATION_CONFIG is not set")
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
	if err := pool.QueryRow(ctx, `SELECT source_start_date,processed_through FROM gfn_metric_checkpoints WHERE metric_key='ipv6_adoption' AND metric_version=1`).Scan(&originalMetricSource, &originalMetricProcessed); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT source_start_date,processed_through FROM gfn_change_checkpoints WHERE detector_key='ipv6_transition' AND detector_version=1`).Scan(&originalChangeSource, &originalChangeProcessed); err != nil {
		t.Fatal(err)
	}
	day1 := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	day3 := day1.AddDate(0, 0, 2)
	var target int64
	if err := pool.QueryRow(ctx, `INSERT INTO gfn_target_tracking_periods(site_id,target,tracked_from,tracked_until,tracking_basis,opened_reason,closed_reason) VALUES(870101,'engine.change.test',$1,$2,'legacy_observed','engine_smoke','engine_smoke') RETURNING id`, day1, day3.AddDate(0, 0, 1)).Scan(&target); err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM gfn_change_events WHERE detector_key='ipv6_transition' AND site_id=870101;DELETE FROM gfn_metric_entity_daily WHERE site_id=870101;DELETE FROM gfn_site_daily WHERE site_id=870101;DELETE FROM gfn_target_tracking_periods WHERE id=$1;UPDATE gfn_metric_checkpoints SET source_start_date=$2,processed_through=$3 WHERE metric_key='ipv6_adoption' AND metric_version=1;UPDATE gfn_change_checkpoints SET source_start_date=$4,processed_through=$5 WHERE detector_key='ipv6_transition' AND detector_version=1`, pgx.QueryExecModeSimpleProtocol, target, originalMetricSource, originalMetricProcessed, originalChangeSource, originalChangeProcessed)
	}
	defer cleanup()
	for _, fixture := range []struct {
		day   time.Time
		state string
	}{{day1, "negative"}, {day3, "positive"}} {
		if _, err := pool.Exec(ctx, `INSERT INTO gfn_site_daily(site_id,fact_date,snapshot_at,tracked_at_end,name,name_en,view_count,group_ids,primary_target_tracking_period_id,primary_target,primary_basis,active_target_count,projection_version,finalized_at) VALUES(870101,$1::date,$1::date+interval '23 hours',true,'Engine Smoke Site','Engine Smoke Site',0,'{}',$2,'engine.change.test','explicit',1,1,$1::date+interval '24 hours');INSERT INTO gfn_metric_entity_daily(metric_key,metric_version,fact_date,site_id,state,reason_code,source_observed_at,dimension_values,source_projection_versions,evaluated_at) VALUES('ipv6_adoption',1,$1::date,870101,$3,'fixture',$1::date+interval '12 hours','{}','{"gfn_site_daily":1}',$1::date+interval '24 hours')`, pgx.QueryExecModeSimpleProtocol, fixture.day, target, fixture.state); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := pool.Exec(ctx, `UPDATE gfn_metric_checkpoints SET source_start_date=$1::date,processed_through=$2::date WHERE metric_key='ipv6_adoption' AND metric_version=1;UPDATE gfn_change_checkpoints SET source_start_date=$1::date,processed_through=$3::date WHERE detector_key='ipv6_transition' AND detector_version=1`, pgx.QueryExecModeSimpleProtocol, day1, day3, day1.AddDate(0, 0, 1)); err != nil {
		t.Fatal(err)
	}
	engine := New(pool, Options{})
	through := day3
	summary, err := engine.Backfill(ctx, BackfillOptions{Detector: "ipv6_transition", Version: 1, Through: &through, MaxDays: 1})
	if err != nil || summary.Processed != 1 {
		t.Fatalf("backfill summary=%+v err=%v", summary, err)
	}
	rebuilt, err := engine.Rebuild(ctx, "ipv6_transition", 1, day3, nil, 0, false)
	if err != nil || rebuilt.Processed != 1 {
		t.Fatalf("rebuild summary=%+v err=%v", rebuilt, err)
	}
	var events int64
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM gfn_change_events WHERE detector_key='ipv6_transition' AND site_id=870101`).Scan(&events); err != nil || events != 1 {
		t.Fatalf("events=%d err=%v", events, err)
	}
}
