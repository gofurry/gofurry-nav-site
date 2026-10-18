package metrics

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	navsqlc "github.com/gofurry/gofurry-nav-collector/internal/db/nav/sqlc"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHTTP2MetricV2Integration(t *testing.T) {
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
	if err := New(pool, Options{}).ValidateCatalog(ctx); err != nil {
		t.Fatal(err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	day := time.Date(2042, 1, 3, 0, 0, 0, 0, time.UTC)
	if _, err := tx.Exec(ctx, `UPDATE gfn_metric_checkpoints SET source_start_date=$1,processed_through=NULL WHERE metric_key='http2_adoption' AND metric_version=2`, day); err != nil {
		t.Fatal(err)
	}
	type fixture struct {
		name, state, reason                                                                 string
		change                                                                              map[string]any
		age                                                                                 time.Duration
		noObserved, failed, notProbed, noLedger, noPrimary, noTarget, wrongSite, oldPrimary bool
	}
	fixtures := []fixture{
		{name: "H2", state: "positive", reason: "http2_negotiated"},
		{name: "H1 empty ALPN", state: "negative", reason: "http1_negotiated_h2_enabled", change: map[string]any{"http_protocol": "HTTP/1.1", "tls_alpn": ""}},
		{name: "H1 ALPN", state: "negative", reason: "http1_negotiated_h2_enabled", change: map[string]any{"http_protocol": "HTTP/1.1", "tls_alpn": "http/1.1"}},
		{name: "H1.0", state: "negative", reason: "http1_negotiated_h2_enabled", change: map[string]any{"http_protocol": "HTTP/1.0", "tls_alpn": ""}},
		{name: "plaintext", state: "not_applicable", reason: "http_plaintext_not_applicable", change: map[string]any{"final_url": "http://final.example/", "final_scheme": "http", "http_protocol": "HTTP/1.1", "tls_alpn": ""}},
		{name: "legacy H1", state: "unknown", reason: "http2_probe_profile_unknown", change: map[string]any{"http2_probe_profile": nil, "http_protocol": "HTTP/1.1", "tls_alpn": ""}},
		{name: "legacy H2", state: "unknown", reason: "http2_probe_profile_unknown", change: map[string]any{"http2_probe_profile": nil}},
		{name: "wrong profile", state: "unknown", reason: "http2_probe_profile_unknown", change: map[string]any{"http2_probe_profile": "h1_only"}},
		{name: "missing ALPN", state: "unknown", reason: "http2_evidence_incomplete", change: map[string]any{"tls_alpn": nil}},
		{name: "null ALPN", state: "unknown", reason: "http2_evidence_incomplete", change: map[string]any{"tls_alpn": json.RawMessage("null")}},
		{name: "missing URL", state: "unknown", reason: "http2_evidence_incomplete", change: map[string]any{"final_url": nil}},
		{name: "missing scheme", state: "unknown", reason: "http2_evidence_incomplete", change: map[string]any{"final_scheme": nil}},
		{name: "missing protocol", state: "unknown", reason: "http2_evidence_incomplete", change: map[string]any{"http_protocol": nil}},
		{name: "invalid URL", state: "unknown", reason: "http2_endpoint_inconsistent", change: map[string]any{"final_url": "https:///no-host"}},
		{name: "scheme disagreement", state: "unknown", reason: "http2_endpoint_inconsistent", change: map[string]any{"final_scheme": "http"}},
		{name: "H2 without ALPN", state: "unknown", reason: "http2_protocol_alpn_inconsistent", change: map[string]any{"tls_alpn": ""}},
		{name: "H1 with h2 ALPN", state: "unknown", reason: "http2_protocol_alpn_inconsistent", change: map[string]any{"http_protocol": "HTTP/1.1"}},
		{name: "plaintext H2", state: "unknown", reason: "http2_protocol_alpn_inconsistent", change: map[string]any{"final_url": "http://final.example/", "final_scheme": "http"}},
		{name: "unknown protocol", state: "unknown", reason: "http2_protocol_unknown", change: map[string]any{"http_protocol": "HTTP/3"}},
		{name: "stale", state: "stale", reason: "http_state_stale", age: 48*time.Hour + time.Second},
		{name: "freshness boundary", state: "positive", reason: "http2_negotiated", age: 48 * time.Hour},
		{name: "failed", state: "probe_failed", reason: "http_probe_failed", noObserved: true, failed: true},
		{name: "not probed", state: "not_probed", reason: "http_not_probed", noObserved: true, notProbed: true},
		{name: "missing ledger", state: "unknown", reason: "historical_probe_state_unknown", noObserved: true, noLedger: true},
		{name: "carry after failure", state: "positive", reason: "http2_negotiated", age: 30 * time.Hour, failed: true},
		{name: "carry after no attempt", state: "positive", reason: "http2_negotiated", age: 30 * time.Hour, notProbed: true},
		{name: "certificate invalid 403", state: "positive", reason: "http2_negotiated", change: map[string]any{"status_code": 403, "tls": map[string]any{"cert_verified": false}}},
		{name: "primary unknown", state: "unknown", reason: "primary_target_unknown", noPrimary: true},
		{name: "target fact missing", state: "unknown", reason: "primary_target_fact_unknown", noTarget: true},
		{name: "different site evidence", state: "unknown", reason: "historical_probe_state_unknown", wrongSite: true},
		{name: "previous primary evidence", state: "unknown", reason: "historical_probe_state_unknown", oldPrimary: true},
	}
	wantCounts := map[string]int64{}
	observed := map[int64]*time.Time{}
	for i, f := range fixtures {
		site := int64(920000 + i)
		var target int64
		if err := tx.QueryRow(ctx, `INSERT INTO gfn_target_tracking_periods(site_id,target,tracked_from,tracking_basis,opened_reason) VALUES($1,'h2.example',$2,'legacy_observed','http2_test') RETURNING id`, site, day.AddDate(0, 0, -4)).Scan(&target); err != nil {
			t.Fatal(err)
		}
		var primary any = target
		if f.noPrimary {
			primary = nil
		}
		if _, err := tx.Exec(ctx, `INSERT INTO gfn_site_daily(site_id,fact_date,snapshot_at,tracked_at_end,name,name_en,view_count,group_ids,primary_target_tracking_period_id,primary_target,primary_basis,active_target_count,projection_version,finalized_at,site_country,nsfw,welfare) VALUES($1,$2::date,$2::date+interval '23 hours',true,'H2','H2',0,ARRAY[92,92,93]::bigint[],$3,CASE WHEN $3::bigint IS NULL THEN NULL ELSE 'h2.example' END,CASE WHEN $3::bigint IS NULL THEN NULL ELSE 'explicit' END,CASE WHEN $3::bigint IS NULL THEN 0 ELSE 1 END,1,$2::date+interval '24 hours','CN',false,true)`, site, day, primary); err != nil {
			t.Fatal(err)
		}
		if !f.noTarget {
			if _, err := tx.Exec(ctx, `INSERT INTO gfn_site_target_daily(target_tracking_period_id,site_id,target,fact_date,snapshot_at,tracked_at_end,projection_version,finalized_at) VALUES($1,$2,'h2.example',$3::date,$3::date+interval '23 hours',true,1,$3::date+interval '24 hours')`, target, site, day); err != nil {
				t.Fatal(err)
			}
		}
		payload := map[string]any{"http2_probe_profile": "go_h2_enabled_v1", "final_url": "https://final.example/", "final_scheme": "https", "http_protocol": "HTTP/2.0", "tls_alpn": "h2", "status_code": 200}
		for key, value := range f.change {
			if value == nil {
				delete(payload, key)
			} else {
				payload[key] = value
			}
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		age := f.age
		if age == 0 {
			age = 12 * time.Hour
		}
		stamp := day.Add(24 * time.Hour).Add(-age)
		var observedAt any = stamp
		observed[site] = &stamp
		if f.noObserved {
			observedAt = nil
			observed[site] = nil
		}
		attempted, success, failure, unattempted := 1, 1, 0, 0
		if f.failed {
			success = 0
			failure = 1
		}
		if f.notProbed {
			attempted = 0
			success = 0
			unattempted = 1
		}
		quality := "acquisition_ledger"
		var expected, zero, unattemptedValue any = 1, 0, unattempted
		if f.noLedger {
			quality = "legacy_observed_only"
			expected = nil
			zero = nil
			unattemptedValue = nil
		}
		protocolSite := site
		if f.wrongSite {
			protocolSite = site + 100000
			observed[site] = nil
		}
		if f.oldPrimary {
			if err := tx.QueryRow(ctx, `INSERT INTO gfn_target_tracking_periods(site_id,target,tracked_from,tracked_until,tracking_basis,opened_reason,closed_reason) VALUES($1,'old-h2.example',$2::date-interval '4 days',$2::date,'legacy_observed','http2_test','primary_replaced') RETURNING id`, site, day).Scan(&target); err != nil {
				t.Fatal(err)
			}
			observed[site] = nil
		}
		if _, err := tx.Exec(ctx, `INSERT INTO gfn_site_target_protocol_daily(target_tracking_period_id,site_id,target,protocol,fact_date,expected_count,attempted_count,success_count,partial_count,failure_count,skipped_count,missed_count,canceled_count,unattempted_count,quality_basis,projection_version,finalized_at,known_state_observed_at,known_state) VALUES($1,$2,'h2.example','http',$3::date,$4,$5,$6,0,$7,$8,$8,$8,$9,$10,1,$3::date+interval '24 hours',$11,$12::jsonb)`, target, protocolSite, day, expected, attempted, success, failure, zero, unattemptedValue, quality, observedAt, encoded); err != nil {
			t.Fatal(err)
		}
		if f.noPrimary {
			observed[site] = nil
		}
		wantCounts[f.state]++
	}
	queries := navsqlc.New(tx)
	params := navsqlc.ProjectNavMetricDayParams{MetricKey: "http2_adoption", MetricVersion: 2, FactDate: pgDate(day)}
	for pass := 0; pass < 2; pass++ { // deterministic replacement, never duplicate entities or dimensions
		if count, err := queries.ProjectNavMetricDay(ctx, params); err != nil || count != int64(len(fixtures)) {
			t.Fatalf("project rows=%d err=%v", count, err)
		}
	}
	for i, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			var state, reason string
			var stamp *time.Time
			var versions string
			site := int64(920000 + i)
			if err := tx.QueryRow(ctx, `SELECT state,reason_code,source_observed_at,source_projection_versions::text FROM gfn_metric_entity_daily WHERE metric_key='http2_adoption' AND metric_version=2 AND fact_date=$1 AND site_id=$2`, day, site).Scan(&state, &reason, &stamp, &versions); err != nil {
				t.Fatal(err)
			}
			if state != f.state || reason != f.reason {
				t.Fatalf("got %s/%s want %s/%s", state, reason, f.state, f.reason)
			}
			want := observed[site]
			if (stamp == nil) != (want == nil) || stamp != nil && !stamp.Equal(*want) {
				t.Fatalf("source time changed: %v want %v", stamp, want)
			}
			var provenance map[string]int
			if err := json.Unmarshal([]byte(versions), &provenance); err != nil || provenance["gfn_site_daily"] != 1 {
				t.Fatalf("provenance=%s err=%v", versions, err)
			}
		})
	}
	global, err := queries.GetNavMetricGlobalDaily(ctx, navsqlc.GetNavMetricGlobalDailyParams{MetricKey: "http2_adoption", MetricVersion: 2, FactDate: pgDate(day)})
	if err != nil {
		t.Fatal(err)
	}
	if global.PopulationCount != int64(len(fixtures)) || global.EligibleCount != int64(len(fixtures))-wantCounts["not_applicable"] || global.PositiveCount != wantCounts["positive"] || global.NegativeCount != wantCounts["negative"] || global.StaleCount != wantCounts["stale"] || global.UnknownCount != wantCounts["unknown"] || global.ProbeFailedCount != wantCounts["probe_failed"] || global.NotProbedCount != wantCounts["not_probed"] {
		t.Fatalf("global counts=%+v expected=%v", global, wantCounts)
	}
	var slices int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM gfn_metric_daily WHERE metric_key='http2_adoption' AND metric_version=2 AND fact_date=$1 AND population_count=$2`, day, len(fixtures)).Scan(&slices); err != nil || slices != 6 {
		t.Fatalf("global/country/nsfw/welfare/two deduped groups=%d err=%v", slices, err)
	}
	for _, bad := range []string{"before cutover", "future evidence", "wrong metric"} {
		t.Run(bad, func(t *testing.T) {
			nested, err := tx.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer nested.Rollback(ctx)
			p := params
			switch bad {
			case "before cutover":
				p.FactDate = pgDate(day.AddDate(0, 0, -1))
			case "wrong metric":
				p.MetricKey = "ipv6_adoption"
			case "future evidence":
				if _, err := nested.Exec(ctx, `UPDATE gfn_site_target_protocol_daily SET known_state_observed_at=$1::date+interval '24 hours 1 second' WHERE site_id=920000`, day); err != nil {
					t.Fatal(err)
				}
			}
			var count int64
			err = nested.QueryRow(ctx, `SELECT gfn_project_http2_metric_day_v2($1,$2,$3)`, p.MetricKey, p.MetricVersion, p.FactDate).Scan(&count)
			if err == nil {
				t.Fatal("invalid input accepted")
			}
		})
	}
	// Nonempty eligible population with no known states must not report 0% adoption.
	if _, err := tx.Exec(ctx, `UPDATE gfn_site_target_protocol_daily SET known_state=known_state-'http2_probe_profile' WHERE fact_date=$1`, day); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.ProjectNavMetricDay(ctx, params); err != nil {
		t.Fatal(err)
	}
	var noKnownAdoption *float64
	var noKnownCoverage float64
	if err := tx.QueryRow(ctx, `SELECT positive_count::double precision/NULLIF(positive_count+negative_count,0),(positive_count+negative_count)::double precision/NULLIF(eligible_count,0) FROM gfn_metric_daily WHERE metric_key='http2_adoption' AND metric_version=2 AND fact_date=$1 AND dimension_key='global'`, day).Scan(&noKnownAdoption, &noKnownCoverage); err != nil || noKnownAdoption != nil || noKnownCoverage != 0 {
		t.Fatalf("no-known ratios=%v/%v err=%v", noKnownAdoption, noKnownCoverage, err)
	}
	params.FactDate = pgDate(day.AddDate(0, 0, 1))
	if count, err := queries.ProjectNavMetricDay(ctx, params); err != nil || count != 0 {
		t.Fatalf("empty projection %d: %v", count, err)
	}
	var population int64
	var adoption, coverage *float64
	if err := tx.QueryRow(ctx, `SELECT population_count,positive_count::double precision/NULLIF(positive_count+negative_count,0), (positive_count+negative_count)::double precision/NULLIF(eligible_count,0) FROM gfn_metric_daily WHERE metric_key='http2_adoption' AND metric_version=2 AND fact_date=$1 AND dimension_key='global'`, day.AddDate(0, 0, 1)).Scan(&population, &adoption, &coverage); err != nil || population != 0 || adoption != nil || coverage != nil {
		t.Fatalf("empty result=%d/%v/%v err=%v", population, adoption, coverage, err)
	}
	// Precise dispatch must continue reaching the existing v2 evaluator.
	for _, key := range []string{"ipv6_adoption", "security_txt_adoption"} {
		params.MetricKey = key
		if _, err := queries.ProjectNavMetricDay(ctx, params); err != nil {
			t.Fatalf("%s dispatch: %v", key, err)
		}
	}
}
