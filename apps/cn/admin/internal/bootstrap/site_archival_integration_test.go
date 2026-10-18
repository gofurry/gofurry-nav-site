package bootstrap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
	authmw "github.com/gofurry/gofurry-admin/internal/app/auth/middleware"
	collectioncontroller "github.com/gofurry/gofurry-admin/internal/app/collectionadmin/controller"
	collectionservice "github.com/gofurry/gofurry-admin/internal/app/collectionadmin/service"
	navadmin "github.com/gofurry/gofurry-admin/internal/app/navadmin/controller"
	"github.com/gofurry/gofurry-admin/internal/app/navadmin/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestAdminSiteArchival(t *testing.T) {
	base := adminIntegrationDSN(t, requireAdminIntegrationConfig(t))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	root := adminIntegrationSQLDB(t, base, "postgres")
	defer root.Close()
	var major int
	if err := root.QueryRowContext(ctx, `SELECT current_setting('server_version_num')::int/10000`).Scan(&major); err != nil || major != 18 {
		t.Fatalf("isolated PG18 required: %d %v", major, err)
	}
	pools := map[string]*pgxpool.Pool{}
	for _, domain := range []string{"admin", "nav"} {
		name := integrationDatabaseName("site_archival_" + domain)
		createIntegrationDatabase(t, ctx, root, name)
		defer dropIntegrationDatabase(t, root, name)
		dsn := integrationDatabaseDSN(base, name)
		applyIntegrationBaseline(t, domain, dsn)
		pools[domain] = integrationPool(t, ctx, dsn)
		defer pools[domain].Close()
	}
	nav, admin := pools["nav"], pools["admin"]
	if _, err := admin.Exec(ctx, `INSERT INTO gfa_admin_account(id,username,display_name,password_hash,role,status,session_version,created_at,updated_at) VALUES(140,'archival-test','Archive Tester','test','operator','active',1,now(),now())`); err != nil {
		t.Fatal(err)
	}
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := nav.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	principal := &authorization.Principal{AccountID: 140, Username: "archival-test", Role: authorization.RoleOperator, Capabilities: authorization.CapabilitiesFor(authorization.RoleOperator)}
	api := navadmin.New(nav, audit.New(admin))
	collection := collectioncontroller.New(collectionservice.New(nil, nav, audit.New(admin)))
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error { c.Locals(authorization.PrincipalContextKey, principal); return c.Next() })
	read, write := authmw.Require(authorization.ContentRead), authmw.Require(authorization.ContentWrite)
	app.Post("/sites", write, api.CreateSite)
	app.Get("/sites", read, api.ListSites)
	app.Get("/summaries", read, api.ListSiteWorkspaceSummaries)
	app.Get("/sites/:id/workspace", read, api.GetSiteWorkspace)
	app.Get("/sites/:id", read, api.GetSite)
	app.Put("/sites/:id", write, api.UpdateSite)
	app.Delete("/sites/:id", write, api.DeleteSite)
	app.Delete("/sites/:id/icon", write, api.ClearSiteIcon)
	app.Post("/targets", write, api.CreateCollectorDomain)
	app.Put("/targets/:id", write, api.UpdateCollectorDomain)
	app.Delete("/targets/:id", write, api.DeleteCollectorDomain)
	app.Post("/targets/:id/primary", write, api.SetPrimaryCollectorDomain)
	app.Put("/groups/bulk", write, api.BulkReplaceSiteGroupMaps)
	app.Post("/groups", write, api.CreateSiteGroup)
	app.Post("/maps", write, api.CreateSiteGroupMap)
	app.Put("/maps/:id", write, api.UpdateSiteGroupMap)
	app.Delete("/maps/:id", write, api.DeleteSiteGroupMap)
	app.Post("/featured", write, api.CreateFeaturedSite)
	app.Put("/featured/:id", write, api.UpdateFeaturedSite)
	app.Delete("/featured/:id", write, api.DeleteFeaturedSite)
	app.Post("/jobs", authmw.Require(authorization.CollectionExecute), collection.CreateJobs)
	call := func(method, path, body string, status int) json.RawMessage {
		t.Helper()
		resp := requestJSON(t, app, method, path, body, nil, status)
		defer resp.Body.Close()
		var envelope integrationEnvelope
		if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
			t.Fatal(err)
		}
		return envelope.Data
	}
	create := func(name string) int64 {
		var site models.SiteDTO
		if err := json.Unmarshal(call("POST", "/sites", fmt.Sprintf(`{"name":%q,"name_en":"Fixture","info":"保留简介","nsfw":"0","welfare":"0"}`, name), 200), &site); err != nil {
			t.Fatal(err)
		}
		return site.ID
	}
	id, activeID, legacyID := create("Archival fixture"), create("Active fixture"), create("Legacy fixture")
	path := fmt.Sprintf("/sites/%d", id)
	decodeID := func(raw json.RawMessage) int64 {
		t.Helper()
		var result struct{ ID int64 }
		if err := json.Unmarshal(raw, &result); err != nil {
			t.Fatal(err)
		}
		return result.ID
	}
	groupID := decodeID(call("POST", "/groups", `{"name":"Archive Group","name_en":"Archive Group"}`, 200))
	mapPayload := fmt.Sprintf(`{"site_id":%d,"group_id":%d,"weight":5}`, id, groupID)
	mapID := decodeID(call("POST", "/maps", mapPayload, 200))
	featuredID := decodeID(call("POST", "/featured", fmt.Sprintf(`{"site_id":%d,"weight":7}`, id), 200))
	targetPayload := func(owner int64, name string) string {
		return fmt.Sprintf(`{"site_id":%d,"name":%q,"prefix":null,"tls":"1","proxy":"0"}`, owner, name)
	}
	var target models.CollectorDomain
	if err := json.Unmarshal(call("POST", "/targets", targetPayload(id, "historical.example.test"), 200), &target); err != nil {
		t.Fatal(err)
	}
	targetPath := fmt.Sprintf("/targets/%d", target.ID)
	call("PUT", targetPath, targetPayload(id, "current.example.test"), 200)
	call("POST", targetPath+"/primary", "{}", 200)
	exec(`INSERT INTO gfn_collector_observation(id,site_id,target,protocol,status,observed_at,payload) VALUES(1,$1,'historical.example.test','http','success','2026-09-02','{"protocol":"HTTP/1.1"}')`, id)
	exec(`INSERT INTO gfn_site_target_daily(target_tracking_period_id,site_id,target,fact_date,snapshot_at,tracked_at_end,projection_version,finalized_at) SELECT id,site_id,target,'2026-09-02','2026-09-02 23:59Z',true,1,'2026-09-03' FROM gfn_target_tracking_periods WHERE site_id=$1`, id)
	exec(`INSERT INTO gfn_site_daily(site_id,fact_date,snapshot_at,tracked_at_end,name,name_en,view_count,group_ids,active_target_count,projection_version,finalized_at) VALUES($1,'2026-09-02','2026-09-02 23:59Z',true,'Historical name','Historical',1,'{}',1,1,'2026-09-03')`, id)
	// Seed retained historical Metric and Change rows, independent of current content.
	exec(`INSERT INTO gfn_metric_entity_daily(metric_key,metric_version,fact_date,site_id,state,reason_code,source_observed_at,dimension_values,source_projection_versions,evaluated_at) VALUES('http2_adoption',1,'2026-09-02',$1,'negative','http2_not_negotiated','2026-09-02 12:00Z','{}','{}',now())`, id)
	exec(`INSERT INTO gfn_change_events(event_key,detector_key,detector_version,site_id,projection_date,event_at,time_basis,event_code,scope_kind,scope_key,source_event_key,source_before_key,source_after_key,materialized_at) VALUES('archival-history','http2_transition',1,$1,'2026-09-02','2026-09-02 12:00Z','observed','http2_disabled','global','all','fixture','before','after',now())`, id)
	retained := func(table string) string {
		t.Helper()
		var result string
		if err := nav.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]'::jsonb)::text FROM `+table+` t WHERE site_id=$1`, id).Scan(&result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	metricBefore, changesBefore, domainsBefore := retained("gfn_metric_entity_daily"), retained("gfn_change_events"), retained("gfn_collector_domain")
	observationsBefore, factsBefore := retained("gfn_collector_observation"), retained("gfn_site_target_daily")
	var siteFactBefore string
	if err := nav.QueryRow(ctx, `SELECT to_jsonb(s)::text FROM gfn_site_daily s WHERE site_id=$1 AND fact_date='2026-09-02'`, id).Scan(&siteFactBefore); err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{"", `{}`, `{"deletion_reason_code":" "}`, `{"deletion_reason_code":"invalid"}`, `{"deletion_reason_code":"other","deletion_reason":"` + strings.Repeat("字", 1001) + `"}`} {
		call("DELETE", path, body, 400)
	}
	if queryBool(t, ctx, nav, `SELECT deleted FROM gfn_site WHERE id=$1`, id) {
		t.Fatal("invalid request archived Site")
	}
	// Audit failure must roll back metadata and both active periods.
	if _, err := admin.Exec(ctx, `ALTER TABLE gfa_admin_audit_log ADD CONSTRAINT reject_archival_test CHECK (action <> 'delete')`); err != nil {
		t.Fatal(err)
	}
	payload := `{"deletion_reason_code":"unreachable","deletion_reason":"  核对记录  "}`
	call("DELETE", path, payload, 500)
	if queryBool(t, ctx, nav, `SELECT deleted FROM gfn_site WHERE id=$1`, id) || queryInt64(t, ctx, nav, `SELECT count(*) FROM gfn_target_tracking_periods WHERE site_id=$1 AND tracked_until IS NULL`, id) != 1 {
		t.Fatal("audit failure did not roll back Site/periods")
	}
	if _, err := admin.Exec(ctx, `ALTER TABLE gfa_admin_audit_log DROP CONSTRAINT reject_archival_test`); err != nil {
		t.Fatal(err)
	}
	call("DELETE", path, payload, 200)
	var snapshot string
	if err := nav.QueryRow(ctx, `SELECT to_jsonb(s)::text FROM gfn_site s WHERE id=$1`, id).Scan(&snapshot); err != nil {
		t.Fatal(err)
	}
	call("DELETE", path, `{"deletion_reason_code":"closed"}`, 400)
	var after string
	if err := nav.QueryRow(ctx, `SELECT to_jsonb(s)::text FROM gfn_site s WHERE id=$1`, id).Scan(&after); err != nil || snapshot != after {
		t.Fatal("repeat archive overwrote metadata", err)
	}
	if metricBefore != retained("gfn_metric_entity_daily") || changesBefore != retained("gfn_change_events") || domainsBefore != retained("gfn_collector_domain") {
		t.Fatal("archive rewrote retained history/config")
	}
	if observationsBefore != retained("gfn_collector_observation") || factsBefore != retained("gfn_site_target_daily") {
		t.Fatal("archive rewrote observations/target Facts")
	}
	var siteFactAfter string
	if err := nav.QueryRow(ctx, `SELECT to_jsonb(s)::text FROM gfn_site_daily s WHERE site_id=$1 AND fact_date='2026-09-02'`, id).Scan(&siteFactAfter); err != nil || siteFactBefore != siteFactAfter {
		t.Fatal("archive rewrote finalized Site Fact", err)
	}
	for _, sql := range []string{`SELECT count(*) FROM gfn_target_tracking_periods WHERE site_id=$1 AND tracked_until IS NULL`, `SELECT count(*) FROM gfn_site_primary_target_periods WHERE site_id=$1 AND effective_until IS NULL`} {
		if queryInt64(t, ctx, nav, sql, id) != 0 {
			t.Fatal("archive left an active period")
		}
	}
	if queryInt64(t, ctx, nav, `SELECT count(*) FROM gfn_target_tracking_periods WHERE site_id=$1`, id) != 2 {
		t.Fatal("lost historical identity")
	}
	var workspace models.SiteWorkspace
	if err := json.Unmarshal(call("GET", path+"/workspace", "", 200), &workspace); err != nil {
		t.Fatal(err)
	}
	if !workspace.Site.Deleted || workspace.Site.DeletedAt == nil || workspace.Site.DeletionReasonCode == nil || *workspace.Site.DeletionReasonCode != "unreachable" || *workspace.Site.DeletionReason != "核对记录" || workspace.FormerPrimaryTarget != "current.example.test" || len(workspace.Targets) != 1 || workspace.Targets[0].Primary {
		t.Fatalf("archived workspace: %+v", workspace)
	}
	// Current default APIs still exclude archived rows. Only the Admin workspace opens them.
	call("GET", path, "", 500)
	exec(`UPDATE gfn_site SET deleted=true,deleted_at='2026-01-01' WHERE id=$1`, legacyID)
	assertPage(t, requestJSON(t, app, "GET", "/sites", "", nil, 200), 1, 1)
	for _, url := range []string{"/summaries?keyword=historical.example.test", "/summaries?status=archived&keyword=historical.example.test", "/summaries?status=archived&page_size=1&page_num=1", "/summaries?status=archived&page_size=1&page_num=2"} {
		raw := call("GET", url, "", 200)
		var page struct {
			Total int
			List  []models.SiteWorkspaceSummary
		}
		if err := json.Unmarshal(raw, &page); err != nil {
			t.Fatal(err)
		}
		wantTotal, wantLen := 2, 1
		if strings.Contains(url, "keyword") {
			wantTotal = 1
		}
		if !strings.Contains(url, "archived") {
			wantTotal, wantLen = 0, 0
		}
		if page.Total != wantTotal || len(page.List) != wantLen {
			t.Fatalf("%s: %s", url, raw)
		}
		if strings.Contains(url, "page_num=1") && page.List[0].ID != id {
			t.Fatal("archived order not deleted_at descending")
		}
		if strings.Contains(url, "page_num=2") && (page.List[0].ID != legacyID || page.List[0].DeletionReasonCode != nil || page.List[0].PrimaryTarget != "") {
			t.Fatal("legacy archive misrepresented")
		}
	}
	call("GET", "/summaries?status=all", "", 400)
	call("PUT", path, `{"name":"changed","name_en":"Changed"}`, 400)
	call("POST", "/targets", targetPayload(id, "new.example.test"), 400)
	call("PUT", targetPath, targetPayload(id, "changed.example.test"), 400)
	call("PUT", targetPath, targetPayload(activeID, "moved.example.test"), 400)
	call("DELETE", targetPath, "", 400)
	call("POST", targetPath+"/primary", "{}", 400)
	call("PUT", "/groups/bulk", fmt.Sprintf(`{"owner_id":%d,"ids":[]}`, id), 400)
	call("POST", "/maps", mapPayload, 400)
	call("PUT", fmt.Sprintf("/maps/%d", mapID), mapPayload, 400)
	call("DELETE", fmt.Sprintf("/maps/%d", mapID), "", 400)
	call("POST", "/featured", fmt.Sprintf(`{"site_id":%d}`, id), 400)
	call("PUT", fmt.Sprintf("/featured/%d", featuredID), fmt.Sprintf(`{"site_id":%d}`, activeID), 400)
	call("DELETE", fmt.Sprintf("/featured/%d", featuredID), "", 400)
	if queryInt64(t, ctx, nav, `SELECT count(*) FROM gfn_site_group_map WHERE id=$1 AND weight=5`, mapID) != 1 || queryInt64(t, ctx, nav, `SELECT count(*) FROM gfn_featured_site WHERE id=$1 AND weight=7`, featuredID) != 1 {
		t.Fatal("archived relationships changed")
	}
	call("DELETE", path+"/icon", "", 500)
	for _, scope := range []string{"site", "target"} {
		call("POST", "/jobs", fmt.Sprintf(`{"domain":"nav","scope_type":%q,"scope_id":%d,"target":"current.example.test","tasks":["http"]}`, scope, id), 400)
	}
	if queryInt64(t, ctx, nav, `SELECT count(*) FROM gfn_collection_jobs WHERE scope_id=$1`, id) != 0 {
		t.Fatal("archived Site enqueued")
	}
	var auditCount int
	if err := admin.QueryRow(ctx, `SELECT count(*) FROM gfa_admin_audit_log WHERE resource='gfn_site' AND target_id=$1 AND action='delete' AND before_data::jsonb->>'deleted'='false' AND after_data::jsonb->>'deletion_reason_code'='unreachable' AND after_data::jsonb->>'deletion_reason'='核对记录'`, fmt.Sprint(id)).Scan(&auditCount); err != nil || auditCount != 1 {
		t.Fatalf("audit metadata count=%d err=%v", auditCount, err)
	}
	principal.Capabilities = nil
	call("GET", path+"/workspace", "", 403)
	call("DELETE", fmt.Sprintf("/sites/%d", activeID), payload, 403)
}
