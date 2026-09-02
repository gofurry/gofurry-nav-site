package metrics

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-collector/roof/env"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestP23SiteCapabilityMetricProjectorsIntegration(t *testing.T) {
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

	day := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	targets := map[int64]int64{}
	for siteID := int64(871001); siteID <= 871008; siteID++ {
		var targetID int64
		if err := tx.QueryRow(ctx, `
INSERT INTO gfn_target_tracking_periods(
    site_id,target,tracked_from,tracked_until,tracking_basis,opened_reason,closed_reason
) VALUES($1,$2,$3,$4,'legacy_observed','p23_metric_test','p23_metric_test') RETURNING id`,
			siteID, "p23-"+time.Unix(siteID, 0).UTC().Format("150405")+".example", day, day.AddDate(0, 0, 2)).Scan(&targetID); err != nil {
			t.Fatal(err)
		}
		targets[siteID] = targetID
	}
	for siteID := int64(871001); siteID <= 871009; siteID++ {
		var primary any
		var primaryTarget, primaryBasis any
		if targetID, ok := targets[siteID]; ok {
			primary = targetID
			primaryTarget = "p23.example"
			primaryBasis = "explicit"
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO gfn_site_daily(
    site_id,fact_date,snapshot_at,tracked_at_end,name,name_en,site_country,nsfw,welfare,
    view_count,group_ids,primary_target_tracking_period_id,primary_target,primary_basis,
    active_target_count,projection_version,finalized_at
) VALUES($1,$2::date,$2::date+interval '23 hours',true,$3,$3,'CN',false,false,
         0,ARRAY[23]::bigint[],$4,$5,$6,CASE WHEN $4::bigint IS NULL THEN 0 ELSE 1 END,1,
         $2::date+interval '24 hours')`, siteID, day, "P2.3 Metric Site", primary,
			primaryTarget, primaryBasis); err != nil {
			t.Fatal(err)
		}
	}

	type targetFixture struct {
		siteID       int64
		httpObserved string
		httpProtocol *string
		headers      string
		tlsObserved  string
		handshake    string
		verified     *bool
	}
	protocol := func(value string) *string { return &value }
	boolean := func(value bool) *bool { return &value }
	fixtures := []targetFixture{
		{871001, "12 hours", protocol("HTTP/2"), `{"strict_transport_security":true,"content_security_policy":true}`, "12 hours", "collected", boolean(true)},
		{871002, "12 hours", protocol("HTTP/1.1"), `{"strict_transport_security":false,"content_security_policy":false}`, "12 hours", "collected", boolean(false)},
		{871003, "12 hours", protocol("HTTP/3"), `{"strict_transport_security":false,"content_security_policy":true}`, "12 hours", "not_tls", nil},
		{871004, "12 hours", nil, `{}`, "12 hours", "failed", boolean(false)},
		{871005, "-48 hours", protocol("HTTP/2"), `{"strict_transport_security":true,"content_security_policy":true}`, "-48 hours", "collected", boolean(true)},
		{871008, "12 hours", protocol("HTTP/1.1"), `{"content_security_policy_report_only":true}`, "12 hours", "collected", nil},
	}
	for _, fixture := range fixtures {
		if _, err := tx.Exec(ctx, `
INSERT INTO gfn_site_target_daily(
    target_tracking_period_id,site_id,target,fact_date,snapshot_at,tracked_at_end,
    http_state_observed_at,http_protocol,http_security_headers,tls_state_observed_at,
    tls_handshake,tls_cert_verified,projection_version,finalized_at
) VALUES($1,$2,'p23.example',$3::date,$3::date+interval '23 hours',true,
         $3::date+$4::interval,$5,$6::jsonb,$3::date+$7::interval,$8,$9,1,$3::date+interval '24 hours')`,
			targets[fixture.siteID], fixture.siteID, day, fixture.httpObserved, fixture.httpProtocol,
			fixture.headers, fixture.tlsObserved, fixture.handshake, fixture.verified); err != nil {
			t.Fatal(err)
		}
	}
	for _, fixture := range []struct {
		siteID                                int64
		expected, attempted, success, failure int
	}{{871006, 1, 1, 0, 1}, {871007, 1, 0, 0, 0}} {
		if _, err := tx.Exec(ctx, `
INSERT INTO gfn_site_target_protocol_daily(
    target_tracking_period_id,site_id,target,protocol,fact_date,expected_count,
    attempted_count,success_count,partial_count,failure_count,skipped_count,missed_count,
    canceled_count,unattempted_count,quality_basis,projection_version,finalized_at
) VALUES($1,$2,'p23.example','http',$3::date,$4,$5,$6,0,$7,0,0,0,$4::integer-$5::integer,
         'acquisition_ledger',1,$3::date+interval '24 hours')`, targets[fixture.siteID], fixture.siteID,
			day, fixture.expected, fixture.attempted, fixture.success, fixture.failure); err != nil {
			t.Fatal(err)
		}
	}

	assertMetric := func(key string, want map[int64]string, population, eligible, positive, negative, stale, notProbed, probeFailed, unknown, notApplicable int64) {
		t.Helper()
		var projected int64
		if err := tx.QueryRow(ctx, `SELECT gfn_project_site_capability_metric_day($1,1,$2::date)`, key, day).Scan(&projected); err != nil {
			t.Fatalf("project %s: %v", key, err)
		}
		if projected != 9 {
			t.Fatalf("project %s rows=%d", key, projected)
		}
		for siteID, state := range want {
			var actual string
			if err := tx.QueryRow(ctx, `SELECT state FROM gfn_metric_entity_daily WHERE metric_key=$1 AND metric_version=1 AND fact_date=$2 AND site_id=$3`, key, day, siteID).Scan(&actual); err != nil || actual != state {
				t.Fatalf("%s site=%d state=%q want=%q err=%v", key, siteID, actual, state, err)
			}
		}
		var counts [9]int64
		if err := tx.QueryRow(ctx, `SELECT population_count,eligible_count,positive_count,negative_count,stale_count,not_probed_count,probe_failed_count,unknown_count,not_applicable_count FROM gfn_metric_daily WHERE metric_key=$1 AND metric_version=1 AND fact_date=$2 AND dimension_key='global' AND dimension_value='all'`, key, day).Scan(&counts[0], &counts[1], &counts[2], &counts[3], &counts[4], &counts[5], &counts[6], &counts[7], &counts[8]); err != nil {
			t.Fatal(err)
		}
		wantCounts := [9]int64{population, eligible, positive, negative, stale, notProbed, probeFailed, unknown, notApplicable}
		if counts != wantCounts {
			t.Fatalf("%s counts=%v want=%v", key, counts, wantCounts)
		}
	}
	assertMetric("http2_adoption", map[int64]string{871001: "positive", 871002: "negative", 871003: "negative", 871004: "unknown", 871005: "stale", 871006: "probe_failed", 871007: "not_probed", 871008: "negative", 871009: "unknown"}, 9, 9, 1, 3, 1, 1, 1, 2, 0)
	assertMetric("hsts_adoption", map[int64]string{871001: "positive", 871002: "negative", 871003: "not_applicable", 871004: "unknown", 871005: "stale", 871006: "probe_failed", 871007: "not_probed", 871008: "unknown", 871009: "unknown"}, 9, 8, 1, 1, 1, 1, 1, 3, 1)
	assertMetric("csp_adoption", map[int64]string{871001: "positive", 871002: "negative", 871003: "positive", 871004: "unknown", 871005: "stale", 871006: "probe_failed", 871007: "not_probed", 871008: "unknown", 871009: "unknown"}, 9, 9, 2, 1, 1, 1, 1, 3, 0)
	assertMetric("tls_certificate_verification", map[int64]string{871001: "positive", 871002: "negative", 871003: "not_applicable", 871004: "unknown", 871005: "stale", 871006: "probe_failed", 871007: "not_probed", 871008: "unknown", 871009: "unknown"}, 9, 8, 1, 1, 1, 1, 1, 3, 1)
	assertMetricExplainAnalyze(t, ctx, tx, day)
}

func assertMetricExplainAnalyze(t *testing.T, ctx context.Context, tx pgx.Tx, day time.Time) {
	t.Helper()
	rows, err := tx.Query(ctx, `EXPLAIN (ANALYZE, BUFFERS) SELECT gfn_project_site_capability_metric_day('http2_adoption',1,$1::date)`, day)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var plan []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatal(err)
		}
		plan = append(plan, line)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(plan, "\n")
	if !strings.Contains(joined, "Execution Time") {
		t.Fatalf("Metric EXPLAIN ANALYZE did not execute: %s", joined)
	}
	t.Logf("P2.3 Metric projector query plan:\n%s", joined)
}

func TestP23SiteCapabilityMetricBackfillAndRebuildIntegration(t *testing.T) {
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
	var originalSource, originalProcessed, originalTargetFacts, originalSiteFacts pgtype.Date
	if err := pool.QueryRow(ctx, `SELECT source_start_date,processed_through FROM gfn_metric_checkpoints WHERE metric_key='http2_adoption' AND metric_version=1`).Scan(&originalSource, &originalProcessed); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT processed_through FROM gfn_fact_rollup_checkpoints WHERE pipeline_key='nav.target_facts'`).Scan(&originalTargetFacts); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT processed_through FROM gfn_fact_rollup_checkpoints WHERE pipeline_key='nav.site_facts'`).Scan(&originalSiteFacts); err != nil {
		t.Fatal(err)
	}
	day := time.Date(2041, 1, 1, 0, 0, 0, 0, time.UTC)
	var target int64
	if err := pool.QueryRow(ctx, `INSERT INTO gfn_target_tracking_periods(site_id,target,tracked_from,tracked_until,tracking_basis,opened_reason,closed_reason) VALUES(871101,'metric-backfill.p23.example',$1::date,$1::date+interval '2 days','legacy_observed','p23_backfill_test','p23_backfill_test') RETURNING id`, day).Scan(&target); err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM gfn_metric_daily WHERE metric_key='http2_adoption' AND metric_version=1 AND fact_date=$1;DELETE FROM gfn_metric_entity_daily WHERE metric_key='http2_adoption' AND metric_version=1 AND site_id=871101;DELETE FROM gfn_site_target_daily WHERE target_tracking_period_id=$2;DELETE FROM gfn_site_daily WHERE site_id=871101;DELETE FROM gfn_target_tracking_periods WHERE id=$2;UPDATE gfn_metric_checkpoints SET source_start_date=$3,processed_through=$4 WHERE metric_key='http2_adoption' AND metric_version=1;UPDATE gfn_fact_rollup_checkpoints SET processed_through=$5 WHERE pipeline_key='nav.target_facts';UPDATE gfn_fact_rollup_checkpoints SET processed_through=$6 WHERE pipeline_key='nav.site_facts'`, pgx.QueryExecModeSimpleProtocol, day, target, originalSource, originalProcessed, originalTargetFacts, originalSiteFacts)
	}
	defer cleanup()
	if _, err := pool.Exec(ctx, `INSERT INTO gfn_site_daily(site_id,fact_date,snapshot_at,tracked_at_end,name,name_en,view_count,group_ids,primary_target_tracking_period_id,primary_target,primary_basis,active_target_count,projection_version,finalized_at) VALUES(871101,$1::date,$1::date+interval '23 hours',true,'Metric Backfill','Metric Backfill',0,ARRAY[]::bigint[],$2,'metric-backfill.p23.example','explicit',1,1,$1::date+interval '24 hours');INSERT INTO gfn_site_target_daily(target_tracking_period_id,site_id,target,fact_date,snapshot_at,tracked_at_end,http_state_observed_at,http_protocol,projection_version,finalized_at) VALUES($2,871101,'metric-backfill.p23.example',$1::date,$1::date+interval '23 hours',true,$1::date+interval '12 hours','HTTP/2',1,$1::date+interval '24 hours');UPDATE gfn_fact_rollup_checkpoints SET processed_through=$1::date WHERE pipeline_key IN ('nav.target_facts','nav.site_facts');UPDATE gfn_metric_checkpoints SET source_start_date=$1::date,processed_through=NULL WHERE metric_key='http2_adoption' AND metric_version=1`, pgx.QueryExecModeSimpleProtocol, day, target); err != nil {
		t.Fatal(err)
	}
	engine := New(pool, Options{})
	summary, err := engine.Backfill(ctx, BackfillOptions{Metric: "http2_adoption", Version: 1, Through: &day, MaxDays: 1})
	if err != nil || summary.Processed != 1 {
		t.Fatalf("HTTP/2 Metric backfill summary=%+v err=%v", summary, err)
	}
	rebuilt, err := engine.Rebuild(ctx, "http2_adoption", 1, day, day, 1, false)
	if err != nil || rebuilt.Processed != 1 {
		t.Fatalf("HTTP/2 Metric rebuild summary=%+v err=%v", rebuilt, err)
	}
	var state string
	if err := pool.QueryRow(ctx, `SELECT state FROM gfn_metric_entity_daily WHERE metric_key='http2_adoption' AND metric_version=1 AND fact_date=$1 AND site_id=871101`, day).Scan(&state); err != nil || state != "positive" {
		t.Fatalf("HTTP/2 Metric backfill state=%q err=%v", state, err)
	}
}
