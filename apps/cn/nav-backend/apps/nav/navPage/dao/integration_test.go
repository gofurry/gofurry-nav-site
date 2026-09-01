package dao_test

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

	detaildao "github.com/gofurry/gofurry-nav-backend/apps/nav/detail/dao"
	insightsdao "github.com/gofurry/gofurry-nav-backend/apps/nav/insights/dao"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/models"
	insightsservice "github.com/gofurry/gofurry-nav-backend/apps/nav/insights/service"
	navdao "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/dao"
	observationdao "github.com/gofurry/gofurry-nav-backend/apps/nav/readmodel/dao"
	sitepagedao "github.com/gofurry/gofurry-nav-backend/apps/nav/sitePage/dao"
	siteindexservice "github.com/gofurry/gofurry-nav-backend/apps/nav/siteindex/service"
	updatesservice "github.com/gofurry/gofurry-nav-backend/apps/nav/updates/service"
	navsqlc "github.com/gofurry/gofurry-nav-backend/internal/db/nav/sqlc"
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

func TestPostgresNavBackendPersistenceSemantics(t *testing.T) {
	configPath := os.Getenv("GOFURRY_NAV_BACKEND_INTEGRATION_CONFIG")
	if configPath == "" {
		t.Skip("set GOFURRY_NAV_BACKEND_INTEGRATION_CONFIG for PostgreSQL integration tests")
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Database integrationDatabaseConfig `yaml:"database"`
		DataBase integrationDatabaseConfig `yaml:"data_base"`
	}
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Database.Host == "" {
		cfg.Database = cfg.DataBase
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
	seedNavReadModel(t, ctx, pool, now)
	queries := navsqlc.New(pool)
	seedNavInsights(t, ctx, pool, now)
	assertNavInsights(t, ctx, queries)
	nav := navdao.New(queries)

	sites, gfErr := nav.GetSiteList()
	if gfErr != nil || len(sites) != 1 || sites[0].ID != 1 || !strings.Contains(sites[0].Domain, "www.example.test") || strings.Contains(sites[0].Domain, "deleted.example.test") {
		t.Fatalf("public sites: %+v err=%v", sites, gfErr)
	}
	index, gfErr := nav.GetSiteIndexList()
	if gfErr != nil || len(index) != 1 || index[0].ID != 1 {
		t.Fatalf("site index: %+v err=%v", index, gfErr)
	}
	groups, gfErr := nav.GetGroupList()
	if gfErr != nil || len(groups) != 2 || groups[0].ID != 20 {
		t.Fatalf("groups preserve priority ordering: %+v err=%v", groups, gfErr)
	}
	mappings, gfErr := nav.GetGroupMapList()
	if gfErr != nil || len(mappings) != 2 || mappings[0].Weight != 10 {
		t.Fatalf("group mappings preserve weight ordering: %+v err=%v", mappings, gfErr)
	}
	featured, gfErr := nav.GetFeaturedSiteList()
	if gfErr != nil || len(featured) != 1 || featured[0].SiteID != 1 {
		t.Fatalf("featured excludes deleted sites: %+v err=%v", featured, gfErr)
	}
	saying, gfErr := nav.GetSayingByRandom("fr")
	if gfErr != nil || saying.Language != "zh" {
		t.Fatalf("saying language fallback: %+v err=%v", saying, gfErr)
	}

	sitePages := sitepagedao.New(queries)
	deletedSite, gfErr := sitePages.GetSiteById(2)
	if gfErr != nil || !deletedSite.Deleted {
		t.Fatalf("legacy site lookup must preserve deleted-row behavior: %+v err=%v", deletedSite, gfErr)
	}
	if gfErr := sitePages.UpdateViewCount(1, 41); gfErr != nil {
		t.Fatalf("update view count: %s", gfErr.GetMsg())
	}
	if got := queryInt64(t, ctx, pool, `SELECT view_count FROM gfn_site WHERE id=1`); got != 41 {
		t.Fatalf("view count = %d", got)
	}

	details := detaildao.New(queries)
	if _, gfErr := details.GetSiteByID(2); gfErr == nil || gfErr.GetMsg() != "404" {
		t.Fatalf("deleted detail must remain not found: %v", gfErr)
	}
	domains, gfErr := details.ListCollectorDomains(1)
	if gfErr != nil || len(domains) != 1 || domains[0].TargetName() != "www.example.test" {
		t.Fatalf("detail domains: %+v err=%v", domains, gfErr)
	}

	observations := observationdao.New(queries)
	history, gfErr := observations.ListObservations(1, " www.example.test ", "http", 100)
	if gfErr != nil || len(history) != 2 || history[0].ID != 102 || history[0].Payload == "" {
		t.Fatalf("target observation history: %+v err=%v", history, gfErr)
	}

	updates := updatesservice.New(queries).GetUpdates("en")
	if updates.State != "ready" || len(updates.Items) != 1 || updates.Items[0].Title != "English notice" {
		t.Fatalf("public updates: %+v", updates)
	}
	siteIndex := siteindexservice.New(nav).GetSiteIndex()
	if siteIndex.State != "ready" || len(siteIndex.Items) != 1 || len(siteIndex.Items[0].Domains) != 1 {
		t.Fatalf("site index response: %+v", siteIndex)
	}
}

func seedNavInsights(t *testing.T, ctx context.Context, pool *pgxpool.Pool, now time.Time) {
	t.Helper()
	entityDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
	latestDay := entityDay.AddDate(0, 0, 1)
	_, err := pool.Exec(ctx, `
INSERT INTO gfn_metric_daily
    (metric_key,metric_version,fact_date,dimension_key,dimension_value,population_count,eligible_count,
     not_applicable_count,positive_count,negative_count,stale_count,not_probed_count,probe_failed_count,unknown_count,computed_at)
VALUES
    ('ipv6_adoption',2,$1,'global','all',10,10,0,2,6,0,0,1,1,$4),
    ('ipv6_adoption',2,$2,'global','all',10,10,0,6,2,0,0,1,1,$4),
    ('ipv6_adoption',2,$2,'site_country','CN',5,5,0,3,1,0,0,0,1,$4),
    ('ipv6_adoption',2,$2,'site_country','unknown',2,0,2,0,0,0,0,0,0,$4),
    ('ipv6_adoption',2,$2,'group_id','999',4,4,0,2,2,0,0,0,0,$4),
    ('ipv6_adoption',2,$3::date + 28,'site_country','CN',4,4,0,2,2,0,0,0,0,$4),
    ('ipv6_adoption',2,$3,'global','all',10,10,0,2,6,0,0,1,1,$4),
    ('ipv6_adoption',1,$5,'global','all',10,10,0,10,0,0,0,0,0,$4),
    ('tls13_adoption',1,$1,'global','all',10,10,0,4,4,0,0,1,1,$4);
INSERT INTO gfn_metric_entity_daily
    (metric_key,metric_version,fact_date,site_id,state,reason_code,dimension_values,source_projection_versions,evaluated_at)
VALUES
    ('ipv6_adoption',2,$1,1,'probe_failed','probe_failed','{}','{}',$4),
    ('tls13_adoption',1,$1,1,'unknown','unknown','{}','{}',$4);
INSERT INTO gfn_site_daily
    (site_id,fact_date,snapshot_at,tracked_at_end,name,name_en,view_count,group_ids,active_target_count,projection_version,finalized_at)
VALUES (1,$1,$4,true,'站点','Site',7,'{}',1,1,$4)
ON CONFLICT (site_id,fact_date) DO NOTHING;
INSERT INTO gfn_change_events
    (event_key,detector_key,detector_version,site_id,projection_date,event_at,time_basis,event_code,scope_kind,scope_key,
     old_value,new_value,source_event_key,source_before_key,source_after_key,source_versions,materialized_at)
VALUES
    ('insights-ipv6-old','ipv6_transition',2,1,$6,NULL,'day','ipv6_disabled','global','all','{}','{}','src-old','before-old','after-old','{}',$4),
    ('insights-ipv6-new','ipv6_transition',2,1,$1,NULL,'day','ipv6_enabled','global','all','{}','{}','src-new','before-new','after-new','{}',$4),
    ('insights-tls-exact','tls13_transition',1,1,$1,$4,'observed','tls13_enabled','global','all','{}','{}','src-exact','before-exact','after-exact','{}',$4),
    ('insights-primary','primary_target_transition',1,1,$1,NULL,'day','primary_target_changed','global','all','{}','{}','src-primary','before-primary','after-primary','{}',$4);
`, pgx.QueryExecModeSimpleProtocol, entityDay, latestDay, latestDay.AddDate(0, 0, -30), now, latestDay.AddDate(0, 0, 1), entityDay.AddDate(0, 0, -1))
	if err != nil {
		t.Fatal(err)
	}
}

func assertNavInsights(t *testing.T, ctx context.Context, queries *navsqlc.Queries) {
	t.Helper()
	svc := insightsservice.New(insightsdao.New(queries))
	overview, err := svc.GetOverview(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(overview.Metrics) != 2 || overview.Metrics[0].Key != "ipv6" || overview.Metrics[0].AsOf == "" || overview.Metrics[0].Delta30D == nil {
		t.Fatalf("public metric mapping/math: %#v", overview.Metrics)
	}
	if overview.Metrics[0].Value == nil || *overview.Metrics[0].Value != .75 || *overview.Metrics[0].Delta30D != .5 {
		t.Fatalf("public metric values: %#v", overview.Metrics[0])
	}
	if len(overview.RecentChanges) != 1 || overview.RecentChanges[0].Type != "site.tls13.enabled" {
		t.Fatalf("overview newest-one-per-entity/whitelist: %#v", overview.RecentChanges)
	}
	site, err := svc.GetSiteInsights(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(site.Capabilities) != 2 || site.Capabilities[0].State != "unavailable" || site.Capabilities[0].Ecosystem.Value == nil || *site.Capabilities[0].Ecosystem.Value != .25 || site.Capabilities[1].State != "unknown" {
		t.Fatalf("same-day comparison/uncertainty semantics: %#v", site.Capabilities)
	}
	if len(site.RecentChanges) != 4 || site.RecentChanges[0].OccurredAt == nil {
		t.Fatalf("entity timeline/time precision: %#v", site.RecentChanges)
	}
	breakdown, err := svc.GetMetricBreakdown(ctx, "ipv6", "country")
	if err != nil || breakdown.AsOf == nil || len(breakdown.Items) != 2 || breakdown.Items[1].Value != "unknown" || breakdown.Items[1].MetricValue != nil {
		t.Fatalf("dimension breakdown/global horizon/unknown: %#v err=%v", breakdown, err)
	}
	trend, err := svc.GetMetricSliceTrend(ctx, "ipv6", "country", "CN", "90d")
	if err != nil || trend.AsOf == nil || len(trend.Points) != 2 || trend.AvailableThrough == nil || *trend.AvailableThrough != *breakdown.AsOf {
		t.Fatalf("slice trend/global anchor: %#v err=%v", trend, err)
	}
	group, err := svc.GetMetricBreakdown(ctx, "ipv6", "group")
	if err != nil || len(group.Items) != 1 || group.Items[0].LabelEn == nil || *group.Items[0].LabelEn != "Group #999" || group.SliceMode != "overlapping" {
		t.Fatalf("group fallback/overlap: %#v err=%v", group, err)
	}
	first, err := svc.GetChanges(ctx, models.ChangeExplorerQuery{Range: "30d", Limit: 2})
	if err != nil || len(first.Items) != 2 || first.NextCursor == nil || first.Items[0].OccurredAt == nil {
		t.Fatalf("change explorer first page/order: %#v err=%v", first, err)
	}
	second, err := svc.GetChanges(ctx, models.ChangeExplorerQuery{Range: "30d", Cursor: *first.NextCursor, Limit: 2})
	if err != nil || len(second.Items) == 0 || second.Items[0].Type == first.Items[len(first.Items)-1].Type && second.Items[0].Date == first.Items[len(first.Items)-1].Date {
		t.Fatalf("change explorer cursor page: %#v err=%v", second, err)
	}
}

func seedNavReadModel(t *testing.T, ctx context.Context, pool *pgxpool.Pool, now time.Time) {
	t.Helper()
	_, err := pool.Exec(ctx, `
INSERT INTO gfn_site
    (id,name,name_en,info,info_en,create_time,update_time,country,nsfw,welfare,icon,deleted,deleted_at,view_count)
VALUES
    (1,'站点','Site','简介','Info',$1,$1,'CN','0','0',NULL,false,NULL,7),
    (2,'隐藏','Hidden','隐藏','Hidden',$1,$1,NULL,'0','0',NULL,true,$1,3);
INSERT INTO gfn_collector_domain (id,name,proxy,prefix,tls,site_id,deleted) VALUES
    (1,'example.test','0','www.','1',1,false),
    (2,'deleted.example.test','0',NULL,'1',1,true),
    (3,'hidden.example.test','0',NULL,'1',2,false);
INSERT INTO gfn_site_group (id,name,name_en,info,info_en,priority,create_time,update_time) VALUES
    (10,'二组','Second','二组','Second',2,$1,$1),
    (20,'一组','First','一组','First',1,$1,$1);
INSERT INTO gfn_site_group_map (id,site_id,group_id,create_time,update_time,weight) VALUES
    (1,1,20,$1,$1,5), (2,1,20,$1,$1,10);
INSERT INTO gfn_featured_site (id,site_id,weight,create_time,update_time) VALUES
    (1,1,10,$1,$1), (2,2,20,$1,$1);
INSERT INTO gfn_saying (id,author,saying,create_time,update_time,language) VALUES
    (1,NULL,'你好',$1,$1,'zh'), (2,NULL,'Hello',$1,$1,'en');
INSERT INTO gfn_nav_update_notice
    (id,title,title_en,body,body_en,published_at,create_time,update_time,deleted)
VALUES
    (1,'公告','English notice','正文','English body',$1,$1,$1,false),
    (2,'隐藏','Hidden','隐藏','Hidden',$1,$1,$1,true);
INSERT INTO gfn_collector_observation
    (id,site_id,target,protocol,status,observed_at,duration_ms,error_code,error_message,payload,schema_version,create_time)
VALUES
    (100,1,'www.example.test','ping','success',$2,1,NULL,NULL,'{}',1,$1),
    (101,1,'www.example.test','http','failure',$3,25,'timeout','late','{"collector_id":"collector-old"}',1,$1),
    (102,1,'www.example.test','http','success',$1,NULL,NULL,NULL,'{"collector_id":"collector-new","job_id":"job-new"}',1,$1);
`, pgx.QueryExecModeSimpleProtocol, now, now.Add(-8*24*time.Hour), now.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
}

func queryInt64(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string) int64 {
	t.Helper()
	var value int64
	if err := pool.QueryRow(ctx, query).Scan(&value); err != nil {
		t.Fatal(err)
	}
	return value
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
	return "gofurry_nav_backend_it_" + hex.EncodeToString(value[:])
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
	if !strings.HasPrefix(name, "gofurry_nav_backend_it_") || strings.ContainsAny(name, `"'; \\`) {
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
