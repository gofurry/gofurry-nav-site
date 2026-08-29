package observation_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	dnsdao "github.com/gofurry/gofurry-nav-collector/collector/dns/dao"
	dnsmodels "github.com/gofurry/gofurry-nav-collector/collector/dns/models"
	"github.com/gofurry/gofurry-nav-collector/collector/facts"
	httpdao "github.com/gofurry/gofurry-nav-collector/collector/http/dao"
	httpmodels "github.com/gofurry/gofurry-nav-collector/collector/http/models"
	lightdao "github.com/gofurry/gofurry-nav-collector/collector/lightprobe/dao"
	metricengine "github.com/gofurry/gofurry-nav-collector/collector/metrics"
	"github.com/gofurry/gofurry-nav-collector/collector/observation"
	pingdao "github.com/gofurry/gofurry-nav-collector/collector/ping/dao"
	pingmodels "github.com/gofurry/gofurry-nav-collector/collector/ping/models"
	cm "github.com/gofurry/gofurry-nav-collector/common/models"
	navsqlc "github.com/gofurry/gofurry-nav-collector/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"gopkg.in/yaml.v3"
)

type integrationDatabaseConfig struct {
	Name     string `yaml:"db_name"`
	Username string `yaml:"db_username"`
	Password string `yaml:"db_password"`
	Host     string `yaml:"db_host"`
	Port     string `yaml:"db_port"`
}

func TestPostgresCollectorPersistenceSemantics(t *testing.T) {
	configPath := os.Getenv("GOFURRY_NAV_COLLECTOR_INTEGRATION_CONFIG")
	if configPath == "" {
		t.Skip("set GOFURRY_NAV_COLLECTOR_INTEGRATION_CONFIG for PostgreSQL integration tests")
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Database integrationDatabaseConfig `yaml:"data_base"`
	}
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		t.Fatal(err)
	}
	baseDSN := integrationDSN(cfg.Database)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	adminDB := integrationSQLDB(t, baseDSN, "postgres")
	defer adminDB.Close()
	databaseName := integrationDatabaseName()
	createIntegrationDatabase(t, ctx, adminDB, databaseName)
	defer dropIntegrationDatabase(t, adminDB, databaseName)
	testDSN := integrationDatabaseDSN(baseDSN, databaseName)
	applyNavBaseline(t, testDSN)
	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	now := time.Now().UTC().Truncate(time.Second)
	seedTargets(t, ctx, pool, now)
	pingDAO := pingdao.New(pool)
	httpDAO := httpdao.New(pool)
	dnsDAO := dnsdao.New(pool)
	lightDAO := lightdao.New(pool)
	for label, load := range map[string]func() (int, error){
		"ping": func() (int, error) { rows, e := pingDAO.GetList(); return len(rows), gfError(e) },
		"http": func() (int, error) { rows, e := httpDAO.GetList(); return len(rows), gfError(e) },
		"dns":  func() (int, error) { rows, e := dnsDAO.GetList(); return len(rows), gfError(e) },
		"light": func() (int, error) {
			rows, e := lightDAO.GetList(observation.ProtocolRDAP)
			return len(rows), gfError(e)
		},
	} {
		count, err := load()
		if err != nil || count != 1 {
			t.Fatalf("%s targets: count=%d err=%v", label, count, err)
		}
	}

	for i := 0; i < 3; i++ {
		when := now.Add(time.Duration(i) * time.Second)
		if err := pingDAO.Add(&pingmodels.GfnCollectorLogPing{ID: int64(100 + i), Name: "example.test", Delay: "1ms", Loss: "0%", Status: "up", CreateTime: cm.LocalTime(when)}); err != nil {
			t.Fatalf("ping write: %s", err.GetMsg())
		}
		if err := httpDAO.Add(&httpmodels.GfnCollectorLogHTTP{ID: int64(200 + i), Name: "example.test", Info: `{"statusCode":200}`, Status: "success", CreateTime: cm.LocalTime(when)}); err != nil {
			t.Fatalf("http write: %s", err.GetMsg())
		}
		a := `[{"value":"203.0.113.1"}]`
		if err := dnsDAO.Add(&dnsmodels.GfnCollectorLogDn{ID: int64(300 + i), Name: "example.test", A: &a, Status: "success", CreateTime: when}); err != nil {
			t.Fatalf("dns write: %s", err.GetMsg())
		}
	}
	for label, prune := range map[string]func(string) (int64, error){
		"ping": func(keep string) (int64, error) { count, e := pingDAO.DeleteByNum(keep); return count, gfError(e) },
		"http": func(keep string) (int64, error) { count, e := httpDAO.DeleteByNum(keep); return count, gfError(e) },
		"dns":  func(keep string) (int64, error) { count, e := dnsDAO.DeleteByNum(keep); return count, gfError(e) },
	} {
		deleted, err := prune("1")
		if err != nil || deleted != 2 {
			t.Fatalf("%s retention: deleted=%d err=%v", label, deleted, err)
		}
	}

	observationDAO := observation.NewDAO(pool)
	protocols := []string{observation.ProtocolPing, observation.ProtocolPing, observation.ProtocolPing, observation.ProtocolHTTP, observation.ProtocolDNS, observation.ProtocolPortCheck, observation.ProtocolRDAP}
	for i, protocol := range protocols {
		record := observation.GfnCollectorObservation{ID: int64(400 + i), SiteID: 1, Target: "example.test", Protocol: protocol,
			Status: observation.StatusSuccess, ObservedAt: now.Add(time.Duration(i) * time.Second), DurationMS: int64(i + 1),
			Payload: `{"ok":true}`, SchemaVersion: 1, CreateTime: now.Add(time.Duration(i) * time.Second)}
		if err := observationDAO.AddObservation(&record); err != nil {
			t.Fatalf("observation write: %s", err.GetMsg())
		}
	}
	trend, gfErr := observationDAO.ListTrendRows(ctx, 1, "example.test", now.Add(-time.Hour), 20)
	if gfErr != nil || len(trend) != 5 {
		t.Fatalf("trend rows=%d err=%v", len(trend), gfErr)
	}
	changes, gfErr := observationDAO.ListChangeRows(ctx, 1, "example.test", now.Add(-time.Hour), 20)
	if gfErr != nil || len(changes) != 4 {
		t.Fatalf("change rows=%d err=%v", len(changes), gfErr)
	}
	deleted, gfErr := observationDAO.DeleteByProtocolLimit(observation.ProtocolPing, "1")
	if gfErr != nil || deleted != 0 {
		t.Fatalf("observation retention: deleted=%d err=%v", deleted, gfErr)
	}
	var remaining int64
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM gfn_collector_observation WHERE protocol='ping'`).Scan(&remaining); err != nil || remaining != 3 {
		t.Fatalf("remaining ping observations=%d err=%v", remaining, err)
	}

	var controlNow time.Time
	if err := pool.QueryRow(ctx, `SELECT now()`).Scan(&controlNow); err != nil {
		t.Fatal(err)
	}
	seedTaskResultRetention(t, ctx, pool, controlNow)
	pruned, err := navsqlc.New(pool).DeleteNavCollectionTaskResultsOlderThan(ctx, 90)
	if err != nil || pruned != 1 {
		t.Fatalf("task-result retention: deleted=%d err=%v", pruned, err)
	}
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM gfn_collection_task_results`).Scan(&remaining); err != nil || remaining != 1 {
		t.Fatalf("remaining task results=%d err=%v", remaining, err)
	}
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM gfn_collector_observation`).Scan(&remaining); err != nil || remaining != int64(len(protocols)) {
		t.Fatalf("retention changed observations: remaining=%d err=%v", remaining, err)
	}
	testHistoricalNavFacts(t, ctx, pool)
	testHistoricalNavMetrics(t, ctx, pool)
}

func testHistoricalNavMetrics(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	day := time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC)
	dayEnd := day.Add(24 * time.Hour)
	if _, err := pool.Exec(ctx, `
INSERT INTO gfn_target_tracking_periods
    (id,collector_domain_id,site_id,target,tracked_from,tracking_basis,opened_reason)
SELECT site_id + 10000,NULL,site_id,'metric-' || site_id || '.example',$1::timestamptz - interval '1 day',
       'legacy_observed','metric_integration'
FROM unnest(ARRAY[99200,99201,99202,99204,99205,99206,99207,99208]::bigint[]) site_id;

INSERT INTO gfn_site_daily (
    site_id,fact_date,snapshot_at,tracked_at_end,name,name_en,site_country,nsfw,welfare,
    view_count,group_ids,primary_target_tracking_period_id,primary_target,primary_basis,
    active_target_count,projection_version,finalized_at,created_at,updated_at
)
SELECT input.site_id,$1::date,$2,true,'metric site ' || input.site_id,'metric site ' || input.site_id,
       input.country,input.nsfw,input.welfare,0,input.group_ids,
       CASE WHEN input.site_id=99203 THEN NULL ELSE input.site_id+10000 END,
       CASE WHEN input.site_id=99203 THEN NULL ELSE 'metric-' || input.site_id || '.example' END,
       CASE WHEN input.site_id=99203 THEN NULL ELSE 'legacy_backfill' END,
       CASE WHEN input.site_id=99203 THEN 0 ELSE 1 END,1,$2,$2,$2
FROM (VALUES
    (99200::bigint,'CN'::text,true::boolean,NULL::boolean,ARRAY[2,5,5]::bigint[]),
    (99201::bigint,NULL::text,false::boolean,true::boolean,ARRAY[]::bigint[]),
    (99202::bigint,'JP'::text,NULL::boolean,false::boolean,ARRAY[2]::bigint[]),
    (99203::bigint,'US'::text,false::boolean,false::boolean,ARRAY[]::bigint[]),
    (99204::bigint,'JP'::text,false::boolean,false::boolean,ARRAY[]::bigint[]),
    (99205::bigint,'JP'::text,false::boolean,false::boolean,ARRAY[]::bigint[]),
    (99206::bigint,'JP'::text,false::boolean,false::boolean,ARRAY[]::bigint[]),
    (99207::bigint,'JP'::text,false::boolean,false::boolean,ARRAY[]::bigint[]),
    (99208::bigint,'JP'::text,false::boolean,false::boolean,ARRAY[]::bigint[])
) input(site_id,country,nsfw,welfare,group_ids);

INSERT INTO gfn_site_target_daily (
    target_tracking_period_id,site_id,target,fact_date,snapshot_at,tracked_at_end,
    tls_state_observed_at,tls_handshake,tls_version,
    dns_state_observed_at,dns_has_aaaa,projection_version,finalized_at,created_at,updated_at
)
SELECT input.site_id+10000,input.site_id,'metric-' || input.site_id || '.example',$1::date,$2,true,
       input.tls_observed,input.tls_handshake,input.tls_version,
       input.dns_observed,input.dns_has_aaaa,1,$2,$2,$2
FROM (VALUES
    (99200::bigint,$2::timestamptz-interval '1 hour','tls'::text,'TLS1.3'::text,$2::timestamptz-interval '1 hour',true::boolean),
    (99201::bigint,$2::timestamptz-interval '1 hour','tls'::text,'TLS1.2'::text,$2::timestamptz-interval '1 hour',false::boolean),
    (99202::bigint,$2::timestamptz-interval '4 days','not_tls'::text,NULL::text,$2::timestamptz-interval '4 days',true::boolean),
    (99204::bigint,NULL::timestamptz,NULL::text,NULL::text,NULL::timestamptz,NULL::boolean),
    (99205::bigint,NULL::timestamptz,NULL::text,NULL::text,NULL::timestamptz,NULL::boolean),
    (99206::bigint,$2::timestamptz-interval '1 hour','tls'::text,NULL::text,$2::timestamptz-interval '1 hour',NULL::boolean),
    (99207::bigint,NULL::timestamptz,NULL::text,NULL::text,NULL::timestamptz,NULL::boolean),
    (99208::bigint,$2::timestamptz-interval '1 hour','not_tls'::text,NULL::text,$2::timestamptz-interval '1 hour',true::boolean)
) input(site_id,tls_observed,tls_handshake,tls_version,dns_observed,dns_has_aaaa);

INSERT INTO gfn_site_target_protocol_daily (
    target_tracking_period_id,site_id,target,protocol,fact_date,
    expected_count,attempted_count,success_count,partial_count,failure_count,
    skipped_count,missed_count,canceled_count,unattempted_count,failure_kind_counts,
    quality_basis,known_state_observed_at,known_state,projection_version,finalized_at,created_at,updated_at
)
SELECT site.site_id+10000,site.site_id,'metric-' || site.site_id || '.example',protocol.protocol,$1::date,
       CASE WHEN site.site_id=99207 THEN NULL ELSE CASE WHEN site.site_id=99205 THEN 0 ELSE 1 END END,
       CASE WHEN site.site_id IN (99205,99207) THEN 0 ELSE 1 END,
       CASE WHEN site.site_id IN (99204,99205,99207) THEN 0 ELSE 1 END,
       0,
       CASE WHEN site.site_id=99204 THEN 1 ELSE 0 END,
       CASE WHEN site.site_id=99207 THEN NULL ELSE 0 END,
       CASE WHEN site.site_id=99207 THEN NULL ELSE 0 END,
       CASE WHEN site.site_id=99207 THEN NULL ELSE 0 END,
       CASE WHEN site.site_id=99207 THEN NULL ELSE 0 END,
       '{}'::jsonb,
       CASE WHEN site.site_id=99207 THEN 'legacy_observed_only' ELSE 'acquisition_ledger' END,
       CASE WHEN protocol.protocol='security_txt' AND site.site_id IN (99200,99201,99202,99206,99208)
            THEN CASE WHEN site.site_id=99202 THEN $2::timestamptz-interval '22 days' ELSE $2::timestamptz-interval '1 hour' END
            ELSE NULL END,
       CASE WHEN protocol.protocol='security_txt' THEN CASE site.site_id
            WHEN 99200 THEN '{"exists":true}'::jsonb
            WHEN 99201 THEN '{"exists":false}'::jsonb
            WHEN 99202 THEN '{"exists":true}'::jsonb
            WHEN 99206 THEN '{}'::jsonb
            WHEN 99208 THEN '{"exists":true}'::jsonb
            ELSE NULL END ELSE NULL END,
       1,$2,$2,$2
FROM unnest(ARRAY[99200,99201,99202,99204,99205,99206,99207,99208]::bigint[]) site(site_id)
CROSS JOIN unnest(ARRAY['dns','http','security_txt']::text[]) protocol(protocol);

UPDATE gfn_fact_rollup_checkpoints
SET source_start_date=$1::date,processed_through=$1::date,updated_at=$2
WHERE pipeline_key IN ('nav.target_facts','nav.site_facts');
UPDATE gfn_metric_checkpoints
SET source_start_date=$1::date,processed_through=NULL,updated_at=$2;
`, pgx.QueryExecModeSimpleProtocol, day, dayEnd); err != nil {
		t.Fatal(err)
	}

	engine := metricengine.New(pool, metricengine.Options{})
	if err := engine.ValidateCatalog(ctx); err != nil {
		t.Fatalf("Nav registry/evaluator drift: %v", err)
	}
	assertNavCount(t, ctx, pool, `SELECT count(*) FROM gfn_metric_registry`, 3)
	assertNavCount(t, ctx, pool, `SELECT count(*) FROM gfn_metric_checkpoints`, 3)
	for _, key := range []string{"ipv6_adoption", "tls13_adoption", "security_txt_adoption"} {
		result, err := engine.RunNext(ctx, key, 1, false)
		if err != nil || !result.Processed {
			t.Fatalf("run Nav metric %s: result=%+v err=%v", key, result, err)
		}
	}

	wantIPv6 := map[int64]string{
		99200: "positive/aaaa_present", 99201: "negative/aaaa_absent", 99202: "stale/dns_state_stale",
		99203: "unknown/primary_target_unknown", 99204: "probe_failed/dns_probe_failed",
		99205: "not_probed/dns_not_probed", 99206: "unknown/dns_metric_field_unknown",
		99207: "unknown/historical_probe_state_unknown", 99208: "positive/aaaa_present",
	}
	for siteID, expected := range wantIPv6 {
		var state, reason string
		if err := pool.QueryRow(ctx, `SELECT state,reason_code FROM gfn_metric_entity_daily WHERE metric_key='ipv6_adoption' AND metric_version=1 AND fact_date=$1 AND site_id=$2`, day, siteID).Scan(&state, &reason); err != nil {
			t.Fatal(err)
		}
		if state+"/"+reason != expected {
			t.Fatalf("ipv6_adoption site=%d got=%s/%s want=%s", siteID, state, reason, expected)
		}
	}

	wantTLS := map[int64]string{
		99200: "positive/tls13_negotiated", 99201: "negative/tls_version_other", 99202: "stale/tls_state_stale",
		99203: "unknown/primary_target_unknown", 99204: "probe_failed/tls_probe_failed",
		99205: "not_probed/tls_not_probed", 99206: "unknown/tls_version_unknown",
		99207: "unknown/historical_probe_state_unknown", 99208: "not_applicable/target_not_tls",
	}
	for siteID, expected := range wantTLS {
		var state, reason string
		if err := pool.QueryRow(ctx, `SELECT state,reason_code FROM gfn_metric_entity_daily WHERE metric_key='tls13_adoption' AND metric_version=1 AND fact_date=$1 AND site_id=$2`, day, siteID).Scan(&state, &reason); err != nil {
			t.Fatal(err)
		}
		if state+"/"+reason != expected {
			t.Fatalf("tls13_adoption site=%d got=%s/%s want=%s", siteID, state, reason, expected)
		}
	}

	wantSecurity := map[int64]string{
		99200: "positive/security_txt_present", 99201: "negative/security_txt_absent",
		99202: "stale/security_txt_state_stale", 99203: "unknown/primary_target_unknown",
		99204: "probe_failed/security_txt_probe_failed", 99205: "not_probed/security_txt_not_probed",
		99206: "unknown/security_txt_field_unknown", 99207: "unknown/historical_probe_state_unknown",
		99208: "positive/security_txt_present",
	}
	for siteID, expected := range wantSecurity {
		var state, reason string
		if err := pool.QueryRow(ctx, `SELECT state,reason_code FROM gfn_metric_entity_daily WHERE metric_key='security_txt_adoption' AND metric_version=1 AND fact_date=$1 AND site_id=$2`, day, siteID).Scan(&state, &reason); err != nil {
			t.Fatal(err)
		}
		if state+"/"+reason != expected {
			t.Fatalf("security_txt_adoption site=%d got=%s/%s want=%s", siteID, state, reason, expected)
		}
	}

	var population, eligible, notApplicable int64
	if err := pool.QueryRow(ctx, `SELECT population_count,eligible_count,not_applicable_count FROM gfn_metric_daily WHERE metric_key='tls13_adoption' AND fact_date=$1 AND dimension_key='global' AND dimension_value='all'`, day).Scan(&population, &eligible, &notApplicable); err != nil {
		t.Fatal(err)
	}
	if population != 9 || eligible != 8 || notApplicable != 1 {
		t.Fatalf("unexpected Nav TLS counts population=%d eligible=%d not_applicable=%d", population, eligible, notApplicable)
	}
	assertNavCount(t, ctx, pool, `SELECT population_count FROM gfn_metric_daily WHERE metric_key='ipv6_adoption' AND fact_date=$1 AND dimension_key='group_id' AND dimension_value='5'`, 1, day)
	assertNavCount(t, ctx, pool, `SELECT population_count FROM gfn_metric_daily WHERE metric_key='ipv6_adoption' AND fact_date=$1 AND dimension_key='site_country' AND dimension_value='unknown'`, 1, day)

	if _, err := engine.Rebuild(ctx, "ipv6_adoption", 1, day, day, 1, false); err != nil {
		t.Fatalf("Nav metric rebuild: %v", err)
	}
	assertNavCount(t, ctx, pool, `SELECT count(*) FROM gfn_metric_entity_daily WHERE metric_key='ipv6_adoption' AND fact_date=$1`, 9, day)

	// Retired versions do not advance at runtime but remain explicitly rebuildable.
	nextDay := day.AddDate(0, 0, 1)
	if _, err := pool.Exec(ctx, `
UPDATE gfn_metric_registry SET status='retired',retired_at=transaction_timestamp()
WHERE metric_key='tls13_adoption' AND metric_version=1;
DELETE FROM gfn_metric_checkpoints
WHERE metric_key='security_txt_adoption' AND metric_version=1;
UPDATE gfn_fact_rollup_checkpoints SET processed_through=$1::date
WHERE pipeline_key IN ('nav.target_facts','nav.site_facts');
`, pgx.QueryExecModeSimpleProtocol, nextDay); err != nil {
		t.Fatal(err)
	}
	if err := engine.Reconcile(ctx); err == nil {
		t.Fatal("Nav reconcile should report the missing security.txt checkpoint")
	}
	assertNavCount(t, ctx, pool, `SELECT count(*) FROM gfn_metric_checkpoints WHERE metric_key='ipv6_adoption' AND processed_through=$1::date`, 1, nextDay)
	assertNavCount(t, ctx, pool, `SELECT population_count FROM gfn_metric_daily WHERE metric_key='ipv6_adoption' AND fact_date=$1 AND dimension_key='global' AND dimension_value='all'`, 0, nextDay)
	assertNavCount(t, ctx, pool, `SELECT count(*) FROM gfn_metric_checkpoints WHERE metric_key='tls13_adoption' AND processed_through=$1::date`, 1, day)
	if _, err := engine.Rebuild(ctx, "tls13_adoption", 1, day, day, 1, true); err != nil {
		t.Fatalf("retired Nav metric explicit rebuild: %v", err)
	}

	// Future evidence rejects the atomic day and preserves its prior checkpoint.
	futureDay := nextDay.AddDate(0, 0, 1)
	if _, err := pool.Exec(ctx, `
INSERT INTO gfn_site_daily (
 site_id,fact_date,snapshot_at,tracked_at_end,name,name_en,view_count,group_ids,
 primary_target_tracking_period_id,primary_target,primary_basis,active_target_count,
 projection_version,finalized_at,created_at,updated_at
)
VALUES (99200,$1::date,$2,true,'metric site 99200','metric site 99200',0,ARRAY[]::bigint[],
        109200,'metric-99200.example','legacy_backfill',1,1,$2,$2,$2);
INSERT INTO gfn_site_target_daily (
 target_tracking_period_id,site_id,target,fact_date,snapshot_at,tracked_at_end,
 dns_state_observed_at,dns_has_aaaa,projection_version,finalized_at,created_at,updated_at
)
VALUES (109200,99200,'metric-99200.example',$1::date,$2,true,$2::timestamptz + interval '1 second',true,1,$2,$2,$2);
INSERT INTO gfn_site_target_protocol_daily (
 target_tracking_period_id,site_id,target,protocol,fact_date,expected_count,attempted_count,
 success_count,partial_count,failure_count,skipped_count,missed_count,canceled_count,
 unattempted_count,failure_kind_counts,quality_basis,projection_version,finalized_at,created_at,updated_at
)
VALUES (109200,99200,'metric-99200.example','dns',$1::date,1,1,1,0,0,0,0,0,0,'{}','acquisition_ledger',1,$2,$2,$2);
UPDATE gfn_fact_rollup_checkpoints SET processed_through=$1::date
WHERE pipeline_key IN ('nav.target_facts','nav.site_facts');
`, pgx.QueryExecModeSimpleProtocol, futureDay, futureDay.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.RunNext(ctx, "ipv6_adoption", 1, false); err == nil {
		t.Fatal("future Nav metric evidence did not fail")
	}
	assertNavCount(t, ctx, pool, `SELECT count(*) FROM gfn_metric_entity_daily WHERE metric_key='ipv6_adoption' AND fact_date=$1`, 0, futureDay)
	assertNavCount(t, ctx, pool, `SELECT count(*) FROM gfn_metric_checkpoints WHERE metric_key='ipv6_adoption' AND processed_through=$1::date`, 1, nextDay)
}

func assertNavCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, want int64, args ...any) {
	t.Helper()
	var got int64
	if err := pool.QueryRow(ctx, query, args...).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("count=%d want=%d query=%s", got, want, query)
	}
}

func testHistoricalNavFacts(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	day := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	if _, err := pool.Exec(ctx, `
INSERT INTO gfn_site (id,name,name_en,info,info_en,create_time,update_time,nsfw,welfare,deleted,deleted_at)
VALUES (10,'facts site','facts site','x','x',$1,$1,'0','0',false,NULL);
INSERT INTO gfn_collector_domain (id,name,proxy,prefix,tls,site_id,deleted)
VALUES (10,'facts.example','0',NULL,'1',10,false);
INSERT INTO gfn_target_tracking_periods (collector_domain_id,site_id,target,tracked_from,tracked_until,tracking_basis,opened_reason,closed_reason)
VALUES (10,10,'facts.example',$1,$2,'explicit','integration','integration');
INSERT INTO gfn_site_daily (site_id,fact_date,snapshot_at,tracked_at_end,name,name_en,view_count,group_ids,active_target_count,projection_version,created_at,updated_at)
VALUES (10,$1::date,$2::timestamptz - interval '1 second',true,'facts site','facts site',0,ARRAY[]::bigint[],1,1,$1,$1);
INSERT INTO gfn_collection_schedules (job_key,name,enabled,schedule_kind,interval_seconds,anchor_at,timezone,misfire_policy,misfire_grace_seconds,priority,concurrency_key,next_scheduled_for)
VALUES ('nav.http','facts fixture',true,'interval',3600,$1,'UTC','skip',0,1,'http',$2)
ON CONFLICT (job_key) DO UPDATE SET enabled=true,next_scheduled_for=EXCLUDED.next_scheduled_for;
INSERT INTO gfn_collector_instances (instance_id,collector_id,hostname,version,commit_sha,capabilities,started_at,last_heartbeat_at)
VALUES ('nav-facts-fixture','facts','localhost','test','',ARRAY['nav.http'],$1,$2)
ON CONFLICT (instance_id) DO NOTHING;
UPDATE gfn_fact_rollup_checkpoints SET source_start_date=$1::date,processed_through=NULL,quality_cutover_at=$1 WHERE pipeline_key IN ('nav.target_facts','nav.site_facts');
`, pgx.QueryExecModeSimpleProtocol, day, day.Add(25*time.Hour)); err != nil {
		t.Fatal(err)
	}

	var periodID int64
	if err := pool.QueryRow(ctx, `SELECT id FROM gfn_target_tracking_periods WHERE collector_domain_id=10`).Scan(&periodID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO gfn_collection_jobs (id,schedule_id,schedule_version,job_key,trigger,scope_type,tasks,priority,concurrency_key,scheduled_for,status,requested_by,created_at,updated_at,completed_at)
SELECT 99100,id,version,'nav.http','scheduled','all',ARRAY['http'],1,'http',$1,'failed','fixture',$1,$1,$1 FROM gfn_collection_schedules WHERE job_key='nav.http';
INSERT INTO gfn_collection_runs (id,job_id,attempt_no,collector_instance_id,status,scheduled_for,started_at,ended_at,expected_count,attempted_count,success_count,partial_count,failure_count,skipped_count)
VALUES ('nav-facts-run',99100,1,'nav-facts-fixture','failed',$1,$1,$1,1,1,0,0,1,0);
INSERT INTO gfn_collector_observation (id,site_id,target,protocol,status,observed_at,duration_ms,error_code,payload,schema_version,job_id,run_id,collector_instance_id)
VALUES (99100,10,'facts.example','http','failure',$1,20,'timeout','{}',1,99100,'nav-facts-run','nav-facts-fixture'),
       (99101,10,'facts.example','http','success',$1::timestamptz + interval '20 minutes',10,NULL,$2::jsonb,1,NULL,NULL,NULL),
       (99102,10,'other.example','http','success',$1,10,NULL,'{"status_code":201}',1,NULL,NULL,NULL),
       (99103,10,'other.example','http','success',$1::timestamptz + interval '1 minute',10,NULL,'{"status_code":202}',1,NULL,NULL,NULL);
INSERT INTO gfn_collection_task_results (run_id,protocol,site_id,target,status,observation_id,duration_ms,error_kind,started_at,ended_at)
VALUES ('nav-facts-run','http',10,'facts.example','failed',99100,20,'upstream',$1,$1);
`, pgx.QueryExecModeSimpleProtocol, day.Add(time.Hour), `{"status_code":200,"response_time_ms":10,"body":"must-not-persist","tls_version":"TLS1.3","cert_verified":true,"cert_not_before":"","cert_not_after":""}`); err != nil {
		t.Fatal(err)
	}

	retentionEngine := facts.New(pool, facts.Options{Now: func() time.Time { return day.Add(26 * time.Hour) }, FinalizationGrace: time.Minute, RetentionEnabled: true, ObservationKeep: 1, RetentionBatch: 1000})
	if deleted, err := retentionEngine.PruneObservations(ctx); err != nil || deleted != 0 {
		t.Fatalf("retention without checkpoint deleted=%d err=%v", deleted, err)
	}
	targetResult, err := retentionEngine.RunNext(ctx, facts.TargetPipeline, false)
	if err != nil || !targetResult.Processed {
		t.Fatalf("target facts result=%+v err=%v", targetResult, err)
	}
	if _, err := pool.Exec(ctx, `
UPDATE gfn_site_target_protocol_daily
SET known_state = jsonb_set(
    known_state,
    '{tls}',
    '{"cert_not_before":"","cert_not_after":""}'::jsonb,
    true
)
WHERE target_tracking_period_id=$1 AND protocol='http' AND fact_date=$2`, periodID, day); err != nil {
		t.Fatalf("seed legacy empty TLS timestamps: %v", err)
	}
	var projectedRows int64
	if err := pool.QueryRow(ctx, `SELECT gfn_project_target_fact_day($1::date)`, day).Scan(&projectedRows); err != nil {
		t.Fatalf("project target facts with legacy empty TLS timestamps: %v", err)
	}
	if projectedRows == 0 {
		t.Fatal("project target facts with legacy empty TLS timestamps returned no rows")
	}
	var emptyTLSIsNull bool
	if err := pool.QueryRow(ctx, `
SELECT tls_cert_not_before IS NULL AND tls_cert_not_after IS NULL
FROM gfn_site_target_daily
WHERE target_tracking_period_id=$1 AND fact_date=$2`, periodID, day).Scan(&emptyTLSIsNull); err != nil || !emptyTLSIsNull {
		t.Fatalf("legacy empty TLS timestamps were not projected as unknown: null=%v err=%v", emptyTLSIsNull, err)
	}
	siteResult, err := retentionEngine.RunNext(ctx, facts.SitePipeline, false)
	if err != nil || !siteResult.Processed {
		t.Fatalf("site facts result=%+v err=%v", siteResult, err)
	}
	var failureCount, expectedCount int
	var latestObservation string
	var knownState []byte
	if err := pool.QueryRow(ctx, `SELECT expected_count,failure_count,latest_observation_status,known_state FROM gfn_site_target_protocol_daily WHERE target_tracking_period_id=$1 AND protocol='http' AND fact_date=$2`, periodID, day).Scan(&expectedCount, &failureCount, &latestObservation, &knownState); err != nil {
		t.Fatal(err)
	}
	if expectedCount != 1 || failureCount != 1 || latestObservation != "success" || strings.Contains(string(knownState), "must-not-persist") || !strings.Contains(string(knownState), `"status_code": 200`) && !strings.Contains(string(knownState), `"status_code":200`) {
		t.Fatalf("unexpected protocol fact expected=%d failure=%d latest=%s state=%s", expectedCount, failureCount, latestObservation, knownState)
	}
	var typedStatus int
	if err := pool.QueryRow(ctx, `SELECT http_status_code FROM gfn_site_target_daily WHERE target_tracking_period_id=$1 AND fact_date=$2`, periodID, day).Scan(&typedStatus); err != nil || typedStatus != 200 {
		t.Fatalf("typed target HTTP status=%d err=%v", typedStatus, err)
	}
	for _, pipeline := range []string{facts.TargetPipeline, facts.SitePipeline} {
		if _, err := retentionEngine.Rebuild(ctx, pipeline, day, day, true); err != nil {
			t.Fatalf("Nav rebuild dry-run pipeline=%s: %v", pipeline, err)
		}
		if _, err := retentionEngine.Rebuild(ctx, pipeline, day, day, false); err != nil {
			t.Fatalf("idempotent Nav rebuild pipeline=%s: %v", pipeline, err)
		}
	}
	if deleted, err := retentionEngine.PruneObservations(ctx); err != nil || deleted < 2 {
		t.Fatalf("checkpoint-safe target-aware retention deleted=%d err=%v", deleted, err)
	}
	for _, target := range []string{"facts.example", "other.example"} {
		var count int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM gfn_collector_observation WHERE site_id=10 AND target=$1 AND protocol='http'`, target).Scan(&count); err != nil || count != 1 {
			t.Fatalf("target-aware keep_count target=%s count=%d err=%v", target, count, err)
		}
	}
}

func seedTaskResultRetention(t *testing.T, ctx context.Context, pool *pgxpool.Pool, now time.Time) {
	t.Helper()
	_, err := pool.Exec(ctx, `
INSERT INTO gfn_collector_instances
(instance_id,collector_id,hostname,version,commit_sha,capabilities,started_at,last_heartbeat_at)
VALUES ('retention-test','retention-test','localhost','test','',ARRAY['nav.ping']::text[],$1::timestamptz,$1::timestamptz);

WITH jobs AS (
  INSERT INTO gfn_collection_jobs
  (job_key,trigger,scope_type,tasks,priority,concurrency_key,status,requested_by,created_at,updated_at,completed_at)
  VALUES
    ('nav.ping','manual','all',ARRAY['ping']::text[],200,'ping','success','test',$1::timestamptz - interval '91 days',$1::timestamptz - interval '91 days',$1::timestamptz - interval '91 days'),
    ('nav.ping','manual','all',ARRAY['ping']::text[],200,'ping','success','test',$1::timestamptz,$1::timestamptz,$1::timestamptz)
  RETURNING id, created_at
), runs AS (
  INSERT INTO gfn_collection_runs
  (id,job_id,attempt_no,collector_instance_id,status,started_at,ended_at,expected_count,attempted_count,success_count)
  SELECT CASE WHEN created_at < $1::timestamptz - interval '90 days' THEN 'retention-old' ELSE 'retention-new' END,
         id,1,'retention-test','success',created_at,created_at,1,1,1
  FROM jobs
  RETURNING id, started_at
)
INSERT INTO gfn_collection_task_results
(run_id,protocol,site_id,target,status,duration_ms,started_at,ended_at)
SELECT id,'ping',1,'example.test','success',1,started_at,started_at FROM runs`, pgx.QueryExecModeSimpleProtocol, now)
	if err != nil {
		t.Fatal(err)
	}
}

func gfError(err interface{ GetMsg() string }) error {
	if err == nil {
		return nil
	}
	return &integrationError{message: err.GetMsg()}
}

type integrationError struct{ message string }

func (err *integrationError) Error() string { return err.message }

func seedTargets(t *testing.T, ctx context.Context, pool *pgxpool.Pool, now time.Time) {
	t.Helper()
	_, err := pool.Exec(ctx, `INSERT INTO gfn_site
(id,name,name_en,info,info_en,create_time,update_time,nsfw,welfare,deleted,deleted_at)
VALUES (1,'site','site','info','info',$1,$1,'0','0',false,NULL),
       (2,'deleted','deleted','info','info',$1,$1,'0','0',true,$1);
INSERT INTO gfn_collector_domain (id,name,proxy,prefix,tls,site_id,deleted)
VALUES (1,'example.test','0',NULL,'1',1,false),
       (2,'hidden.test','0',NULL,'1',2,false),
       (3,'deleted.test','0',NULL,'1',1,true)`, pgx.QueryExecModeSimpleProtocol, now)
	if err != nil {
		t.Fatal(err)
	}
}

func integrationDSN(cfg integrationDatabaseConfig) string {
	u := &url.URL{Scheme: "postgres", Host: cfg.Host + ":" + cfg.Port, Path: cfg.Name}
	u.User = url.UserPassword(cfg.Username, cfg.Password)
	query := u.Query()
	query.Set("sslmode", "prefer")
	u.RawQuery = query.Encode()
	return u.String()
}

func integrationDatabaseDSN(baseDSN, databaseName string) string {
	u, err := url.Parse(baseDSN)
	if err != nil {
		panic(err)
	}
	u.Path = databaseName
	return u.String()
}

func integrationSQLDB(t *testing.T, dsn, databaseName string) *sql.DB {
	t.Helper()
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.Database = databaseName
	return stdlib.OpenDB(*config)
}

func integrationDatabaseName() string {
	var value [5]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic(err)
	}
	return "gofurry_nav_collector_it_" + hex.EncodeToString(value[:])
}

func createIntegrationDatabase(t *testing.T, ctx context.Context, adminDB *sql.DB, name string) {
	t.Helper()
	validateIntegrationDatabaseName(t, name)
	if _, err := adminDB.ExecContext(ctx, `CREATE DATABASE `+quoteIntegrationIdentifier(name)); err != nil {
		t.Fatal(err)
	}
}

func dropIntegrationDatabase(t *testing.T, adminDB *sql.DB, name string) {
	t.Helper()
	validateIntegrationDatabaseName(t, name)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := adminDB.ExecContext(ctx, `DROP DATABASE IF EXISTS `+quoteIntegrationIdentifier(name)+` WITH (FORCE)`); err != nil {
		t.Errorf("drop temporary database: %v", err)
	}
}

func validateIntegrationDatabaseName(t *testing.T, name string) {
	t.Helper()
	if !strings.HasPrefix(name, "gofurry_nav_collector_it_") || strings.ContainsAny(name, `"'; \`) {
		t.Fatalf("unsafe temporary database name %q", name)
	}
}

func quoteIntegrationIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func applyNavBaseline(t *testing.T, dsn string) {
	t.Helper()
	repositoryRoot := integrationRepositoryRoot(t)
	command := exec.Command("go", "tool", "goose", "-dir", filepath.Join(repositoryRoot, "db", "nav", "migrations"), "postgres", dsn, "up")
	command.Dir = filepath.Join(repositoryRoot, "tools")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("apply Nav baseline: %v\n%s", err, output)
	}
}

func integrationRepositoryRoot(t *testing.T) string {
	t.Helper()
	current, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(current, "sqlc.yaml")); err == nil {
			if _, err := os.Stat(filepath.Join(current, "tools", "go.mod")); err == nil {
				return current
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			t.Fatal("repository root containing sqlc.yaml and tools/go.mod not found")
		}
		current = parent
	}
}
