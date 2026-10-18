package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
)

func TestPostgresHTTP2MetricV2Upgrade(t *testing.T) {
	dsn := integrationAdminDSN(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	admin := openDatabase(t, dsn, "postgres")
	defer admin.Close()
	var major int
	if err := admin.QueryRowContext(ctx, `SELECT current_setting('server_version_num')::integer/10000`).Scan(&major); err != nil || major != 18 {
		t.Fatalf("isolated PostgreSQL 18 required: major=%d err=%v", major, err)
	}
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	migrations := filepath.Join(root, "db", "nav", "migrations")
	for _, scenario := range []string{"next UTC day", "later Facts start", "Facts start unavailable"} {
		t.Run(scenario, func(t *testing.T) {
			name := temporaryDatabaseName("gfn", "http2_upgrade")
			createDatabase(t, ctx, admin, name)
			defer dropDatabase(t, admin, name)
			db := openDatabase(t, dsn, name)
			defer db.Close()
			if err := goose.UpToContext(ctx, db, migrations, 20261006010000); err != nil {
				t.Fatal(err)
			}
			if _, err := db.ExecContext(ctx, `
UPDATE gfn_fact_rollup_checkpoints SET source_start_date=(now() AT TIME ZONE 'UTC')::date-10;
UPDATE gfn_metric_checkpoints SET source_start_date='2026-09-01',processed_through='2026-09-02' WHERE metric_key='http2_adoption' AND metric_version=1;
UPDATE gfn_change_checkpoints SET source_start_date='2026-09-01',processed_through='2026-09-02' WHERE detector_key='http2_transition' AND detector_version=1;
INSERT INTO gfn_metric_entity_daily(metric_key,metric_version,fact_date,site_id,state,reason_code,source_observed_at,dimension_values,source_projection_versions,evaluated_at)
VALUES('http2_adoption',1,'2026-09-02',920001,'negative','http2_not_negotiated','2026-09-02 12:00Z','{}','{}',now());
INSERT INTO gfn_metric_daily(metric_key,metric_version,fact_date,dimension_key,dimension_value,population_count,eligible_count,not_applicable_count,positive_count,negative_count,stale_count,not_probed_count,probe_failed_count,unknown_count,computed_at)
VALUES('http2_adoption',1,'2026-09-02','global','all',1,1,0,0,1,0,0,0,0,now());
INSERT INTO gfn_change_events(event_key,detector_key,detector_version,site_id,projection_date,event_at,time_basis,event_code,scope_kind,scope_key,source_event_key,source_before_key,source_after_key,materialized_at)
VALUES('http2-v1-retained','http2_transition',1,920001,'2026-09-02','2026-09-02 12:00Z','observed','http2_disabled','global','all','retained-source','before','after',now());`); err != nil {
				t.Fatal(err)
			}
			switch scenario {
			case "later Facts start":
				if _, err := db.ExecContext(ctx, `UPDATE gfn_fact_rollup_checkpoints SET source_start_date=(now() AT TIME ZONE 'UTC')::date+3 WHERE pipeline_key='nav.target_facts'`); err != nil {
					t.Fatal(err)
				}
			case "Facts start unavailable":
				if _, err := db.ExecContext(ctx, `UPDATE gfn_fact_rollup_checkpoints SET source_start_date=NULL WHERE pipeline_key='nav.target_facts'`); err != nil {
					t.Fatal(err)
				}
			}
			// Capture complete rows, not merely counts: a migration may not rewrite history.
			checks := []string{
				`SELECT jsonb_agg(to_jsonb(t) ORDER BY metric_key,metric_version) FROM gfn_metric_checkpoints t`,
				`SELECT jsonb_agg(to_jsonb(t) ORDER BY detector_key,detector_version) FROM gfn_change_checkpoints t`,
				`SELECT jsonb_agg(to_jsonb(t)) FROM gfn_metric_entity_daily t`,
				`SELECT jsonb_agg(to_jsonb(t)) FROM gfn_metric_daily t`,
				`SELECT jsonb_agg(to_jsonb(t)) FROM gfn_change_events t`,
				`SELECT jsonb_agg(to_jsonb(t)-'status'-'retired_at' ORDER BY metric_key,metric_version) FROM gfn_metric_registry t`,
				`SELECT jsonb_agg(to_jsonb(t)-'status'-'retired_at' ORDER BY detector_key,detector_version) FROM gfn_change_registry t`,
				`SELECT jsonb_agg(to_jsonb(t) ORDER BY metric_key,metric_version) FROM gfn_metric_registry t WHERE metric_key<>'http2_adoption'`,
				`SELECT jsonb_agg(to_jsonb(t) ORDER BY detector_key,detector_version) FROM gfn_change_registry t WHERE detector_key<>'http2_transition'`,
				`SELECT to_jsonb(pg_get_functiondef('gfn_project_site_capability_metric_day(text,integer,date)'::regprocedure))`,
				`SELECT to_jsonb(pg_get_functiondef('gfn_project_site_capability_change_day(text,integer,date)'::regprocedure))`,
			}
			before := make([]string, len(checks))
			for i, q := range checks {
				if err := db.QueryRowContext(ctx, q).Scan(&before[i]); err != nil {
					t.Fatal(err)
				}
			}
			if err := goose.UpContext(ctx, db, migrations); err != nil {
				t.Fatal(err)
			}
			for i, q := range checks {
				if i == 0 || i == 5 {
					q += ` WHERE NOT (metric_key='http2_adoption' AND metric_version=2)`
				}
				var after string
				if err := db.QueryRowContext(ctx, q).Scan(&after); err != nil {
					t.Fatal(err)
				}
				if after != before[i] {
					t.Fatalf("historical contract changed at check %d", i)
				}
			}
			var retiredMetric, retiredDetector, activeV2, changeV2 int
			if err := db.QueryRowContext(ctx, `SELECT
 (SELECT count(*) FROM gfn_metric_registry WHERE metric_key='http2_adoption' AND metric_version=1 AND status='retired' AND retired_at IS NOT NULL),
 (SELECT count(*) FROM gfn_change_registry WHERE detector_key='http2_transition' AND detector_version=1 AND status='retired' AND retired_at IS NOT NULL),
 (SELECT count(*) FROM gfn_metric_registry WHERE metric_key='http2_adoption' AND metric_version=2 AND status='active' AND retired_at IS NULL),
 (SELECT count(*) FROM gfn_change_registry WHERE detector_key='http2_transition' AND detector_version=2)`).Scan(&retiredMetric, &retiredDetector, &activeV2, &changeV2); err != nil || retiredMetric != 1 || retiredDetector != 1 || activeV2 != 1 || changeV2 != 0 {
				t.Fatalf("registry retirement %d/%d/%d/%d: %v", retiredMetric, retiredDetector, activeV2, changeV2, err)
			}
			var start, processed sql.NullTime
			if err := db.QueryRowContext(ctx, `SELECT source_start_date,processed_through FROM gfn_metric_checkpoints WHERE metric_key='http2_adoption' AND metric_version=2`).Scan(&start, &processed); err != nil {
				t.Fatal(err)
			}
			var today time.Time
			if err := db.QueryRowContext(ctx, `SELECT (now() AT TIME ZONE 'UTC')::date`).Scan(&today); err != nil {
				t.Fatal(err)
			}
			want := today.AddDate(0, 0, 1)
			if scenario == "later Facts start" {
				want = today.AddDate(0, 0, 3)
			}
			if processed.Valid || scenario == "Facts start unavailable" && start.Valid || scenario != "Facts start unavailable" && (!start.Valid || !start.Time.Equal(want)) {
				t.Fatalf("cutover=%v processed=%v want=%v", start, processed, want)
			}
			if err := goose.DownContext(ctx, db, migrations); err == nil || !strings.Contains(err.Error(), "irreversible") {
				t.Fatalf("destructive Down allowed: %v", err)
			}
			if version, err := goose.GetDBVersionContext(ctx, db); err != nil || version != 20261018010000 {
				t.Fatalf("failed Down changed version=%d err=%v", version, err)
			}
		})
	}
}
