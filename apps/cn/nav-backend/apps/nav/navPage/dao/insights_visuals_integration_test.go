package dao_test

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	insightsdao "github.com/gofurry/gofurry-nav-backend/apps/nav/insights/dao"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/models"
	insightsservice "github.com/gofurry/gofurry-nav-backend/apps/nav/insights/service"
	navsqlc "github.com/gofurry/gofurry-nav-backend/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.yaml.in/yaml/v4"
)

// The existing CI -run prefix includes this Insights-only test. Reuse its
// isolated database lifecycle without changing any Nav Home/Detail tests.
func TestPostgresNavBackendPersistenceSemanticsInsightVisuals(t *testing.T) {
	configPath := os.Getenv("GOFURRY_NAV_BACKEND_INTEGRATION_CONFIG")
	if configPath == "" {
		t.Skip("set GOFURRY_NAV_BACKEND_INTEGRATION_CONFIG for isolated PostgreSQL")
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Database integrationDatabaseConfig `yaml:"database"`
		DataBase integrationDatabaseConfig `yaml:"data_base"`
	}
	if err = yaml.Unmarshal(content, &cfg); err != nil {
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
	name := integrationDatabaseName()
	createIntegrationDatabase(t, ctx, adminDB, name)
	defer dropIntegrationDatabase(t, adminDB, name)
	dsn := integrationDatabaseDSN(baseDSN, name)
	applyNavBaseline(t, dsn)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	queries := navsqlc.New(pool)
	dao := insightsdao.New(queries)
	svc := insightsservice.New(dao)
	read := func() models.Overview {
		t.Helper()
		got, err := svc.GetOverview(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if got.SiteVisuals == nil || len(got.SiteVisuals) > 5 {
			t.Fatal("array bound", got.SiteVisuals)
		}
		return got
	}
	if len(read().SiteVisuals) != 0 {
		t.Fatal("empty read")
	}
	for id := int64(1082201); id <= 1082209; id++ {
		key := fmt.Sprintf("nav/sites/%d/icon/%s.png", id, strings.Repeat("a", 32))
		exec(`INSERT INTO gfn_site(id,name,name_en,info,info_en,create_time,update_time,country,nsfw,welfare,deleted,view_count,icon) VALUES($1,'站点','Site','','',now(),now(),'CN','0','0',false,0,$2)`, id, key)
		if id == 1082209 {
			continue
		} // Trustworthy but not in the actual event feed.
		exec(`INSERT INTO gfn_change_events(event_key,detector_key,detector_version,site_id,projection_date,time_basis,event_code,scope_kind,scope_key,old_value,new_value,source_event_key,source_before_key,source_after_key,source_versions,materialized_at) VALUES($1,'ipv6_transition',2,$2,'2026-10-09','day','ipv6_enabled','global','all','{}','{}',$1,'before','after','{}',now())`, fmt.Sprintf("visual-%d", id), id)
		if got := read(); len(got.SiteVisuals) != min(int(id-1082200), 5) {
			t.Fatalf("0..5 items: %+v", got.SiteVisuals)
		}
	}
	baseline := read()
	if len(baseline.RecentChanges) != 8 || baseline.SiteVisuals[0].SiteID != 1082208 {
		t.Fatal("event order", baseline)
	}
	for _, tc := range []struct{ Name, Mutation, Restore string }{
		{"adult", `UPDATE gfn_site SET nsfw='1' WHERE id=1082208`, `UPDATE gfn_site SET nsfw='0' WHERE id=1082208`},
		{"unknown", `UPDATE gfn_site SET nsfw='?' WHERE id=1082208`, `UPDATE gfn_site SET nsfw='0' WHERE id=1082208`},
		{"empty", `UPDATE gfn_site SET nsfw='' WHERE id=1082208`, `UPDATE gfn_site SET nsfw='0' WHERE id=1082208`},
		{"literal true", `UPDATE gfn_site SET nsfw='true' WHERE id=1082208`, `UPDATE gfn_site SET nsfw='0' WHERE id=1082208`},
		{"deleted", `UPDATE gfn_site SET deleted=true,deleted_at=now() WHERE id=1082208`, `UPDATE gfn_site SET deleted=false,deleted_at=NULL WHERE id=1082208`},
		{"foreign icon", `UPDATE gfn_site SET icon=replace(icon,'/1082208/','/1082209/') WHERE id=1082208`, `UPDATE gfn_site SET icon=replace(icon,'/1082209/','/1082208/') WHERE id=1082208`},
		{"blank name", `UPDATE gfn_site SET name=E'\t',name_en='' WHERE id=1082208`, `UPDATE gfn_site SET name='站点',name_en='Site' WHERE id=1082208`},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			exec(tc.Mutation)
			got := read()
			if got.SiteVisuals[0].SiteID != 1082207 {
				t.Fatalf("invalid record entered: %+v", got.SiteVisuals)
			}
			for _, visual := range got.SiteVisuals {
				if visual.SiteID == 1082209 {
					t.Fatal("non-event site admitted")
				}
			}
			exec(tc.Restore)
		})
	}
	for _, asset := range []any{nil, "", "https://example.test/logo.png", "nav/sites/1082208/icon/abc.png", "nav/sites/1082208/icon/" + strings.Repeat("a", 32) + ".png\n"} {
		exec(`UPDATE gfn_site SET icon=$1 WHERE id=1082208`, asset)
		if read().SiteVisuals[0].SiteID != 1082207 {
			t.Fatal("invalid icon admitted", asset)
		}
	}
	exec(`UPDATE gfn_site SET icon=$1 WHERE id=1082208`, fmt.Sprintf("nav/sites/1082208/icon/%s.svg", strings.Repeat("b", 32)))
	if !strings.HasSuffix(read().SiteVisuals[0].Visual.Asset, ".svg") {
		t.Fatal("updated asset not visible")
	}
	// The actual NOT NULL varchar(4) also prevents null and these long strings.
	// Unit tests still reject them defensively if a store returns malformed data.
	for _, invalid := range []any{nil, "false", "unknown"} {
		if _, err := pool.Exec(ctx, `UPDATE gfn_site SET nsfw=$1 WHERE id=1082208`, invalid); err == nil {
			t.Fatal("expected NSFW schema constraint", invalid)
		}
	}
	// Optional query timeout: lock only icon's table. Baseline reads also use
	// gfn_site, so call the optional DAO directly and verify cancellation there.
	locker, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer locker.Rollback(context.Background())
	if _, err = locker.Exec(ctx, `LOCK TABLE gfn_site IN ACCESS EXCLUSIVE MODE`); err != nil {
		t.Fatal(err)
	}
	bounded, stop := context.WithTimeout(ctx, 50*time.Millisecond)
	defer stop()
	if _, err = dao.ListOverviewSiteVisuals(bounded, []int64{1082208}); err == nil || bounded.Err() != context.DeadlineExceeded {
		t.Fatal("optional deadline ignored", err)
	}
	_ = locker.Rollback(ctx)
	// Compare events against the public mapper before/after decoration. Returning
	// a new icon must never rewrite the existing event DTO through this projection.
	current := read()
	raw, err := dao.ListOverviewChanges(ctx, []string{"ipv6_transition"}, []string{"ipv6_transition/2/ipv6_enabled"}, 8)
	if err != nil {
		t.Fatal(err)
	}
	for i, change := range current.RecentChanges {
		if change.Entity.ID != raw[i].EntityID || change.OccurredAt != nil || change.Date != "2026-10-09" {
			t.Fatal("event semantics changed")
		}
	}
	if !reflect.DeepEqual(current.Metrics, baseline.Metrics) || current.Changes7D != baseline.Changes7D {
		t.Fatal("decoration changed facts")
	}
}
