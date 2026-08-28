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
`, pgx.QueryExecModeSimpleProtocol, day.Add(time.Hour), `{"status_code":200,"response_time_ms":10,"body":"must-not-persist","tls_version":"TLS1.3","cert_verified":true}`); err != nil {
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
