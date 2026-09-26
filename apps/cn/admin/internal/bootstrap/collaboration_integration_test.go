package bootstrap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
	"github.com/gofurry/gofurry-admin/internal/app/collaboration"
	gameadmin "github.com/gofurry/gofurry-admin/internal/app/gameadmin/controller"
	navadmin "github.com/gofurry/gofurry-admin/internal/app/navadmin/controller"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	"github.com/gofurry/gofurry-admin/internal/app/workbench"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type collaborationTrace struct{ queries atomic.Int64 }

func (t *collaborationTrace) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	t.queries.Add(1)
	return ctx
}
func (t *collaborationTrace) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func TestAdminCollaborationThreeDatabase(t *testing.T) {
	baseDSN := adminIntegrationDSN(t, requireAdminIntegrationConfig(t))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	root := adminIntegrationSQLDB(t, baseDSN, "postgres")
	defer root.Close()
	pools := map[string]*pgxpool.Pool{}
	dsns := map[string]string{}
	for _, domain := range []string{"admin", "game", "nav"} {
		name := integrationDatabaseName("collaboration_" + domain)
		createIntegrationDatabase(t, ctx, root, name)
		defer dropIntegrationDatabase(t, root, name)
		dsns[domain] = integrationDatabaseDSN(baseDSN, name)
		applyIntegrationBaseline(t, domain, dsns[domain])
		pools[domain] = integrationPool(t, ctx, dsns[domain])
		defer pools[domain].Close()
	}
	admin, game, nav := pools["admin"], pools["game"], pools["nav"]
	if _, err := admin.Exec(ctx, `INSERT INTO gfa_admin_account(id,password_hash,session_version,created_at,updated_at,username,display_name,role,status) VALUES (1,'test',1,now(),now(),'member-one','Member One','operator','active'),(2,'test',1,now(),now(),'member-two','Member Two','developer','active')`); err != nil {
		t.Fatal(err)
	}
	principal := &authorization.Principal{AccountID: 1, Username: "member-one", DisplayName: "Member One", Role: authorization.RoleOperator, Capabilities: authorization.CapabilitiesFor(authorization.RoleOperator)}
	meta := audit.MetaForPrincipal(audit.Meta{}, principal)
	other := audit.MetaForPrincipal(audit.Meta{}, &authorization.Principal{AccountID: 2, Username: "member-two", DisplayName: "Member Two", Role: authorization.RoleDeveloper})
	logger := audit.New(admin)
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error { c.Locals(authorization.PrincipalContextKey, principal); return c.Next() })
	gameAPI, navAPI := gameadmin.New(game, logger), navadmin.New(nav, logger)
	app.Post("/games", gameAPI.CreateGame)
	app.Post("/sites", navAPI.CreateSite)
	app.Post("/targets", navAPI.CreateCollectorDomain)
	gameID := responseID(t, requestJSON(t, app, "POST", "/games", `{"name":"Existing Game","name_en":"Game","info":"Info","info_en":"Info","resources":[],"groups":[],"developers":[],"publishers":[],"appid":424242,"header":"","links":[],"weight":0,"primary_tag":0,"secondary_tag":0}`, nil, 200))
	siteID := responseID(t, requestJSON(t, app, "POST", "/sites", `{"name":"Existing Site","name_en":"Site","info":"Info","info_en":"Info","country":null,"nsfw":"0","welfare":"0"}`, nil, 200))
	closeResponse(requestJSON(t, app, "POST", "/targets", fmt.Sprintf(`{"site_id":%d,"name":"example.test","prefix":"www.","proxy":"0","tls":"1"}`, siteID), nil, 200))
	// Even the test connections used by Collaboration refuse business writes.
	trace := &collaborationTrace{}
	readonly := func(domain string) *pgxpool.Pool {
		cfg, err := pgxpool.ParseConfig(dsns[domain])
		if err != nil {
			t.Fatal(err)
		}
		cfg.ConnConfig.RuntimeParams["default_transaction_read_only"] = "on"
		cfg.ConnConfig.Tracer = trace
		pool, err := pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(pool.Close)
		return pool
	}
	readGame, readNav := readonly("game"), readonly("nav")
	s := collaboration.New(admin, readGame, readNav, logger)
	api := collaboration.NewAPI(s)
	app.Post("/ideas", api.Create)
	app.Put("/ideas/:id", api.Update)
	app.Delete("/ideas/:id", api.Delete)
	app.Post("/ideas/:id/research", api.Transition("research"))
	app.Post("/ideas/batch", api.Batch)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	idea, e := s.Create(ctx, meta, collaboration.Input{Kind: "game", Source: "steam:424242"})
	must(e)
	siteIdea, e := s.Create(ctx, meta, collaboration.Input{Kind: "site", Source: "https://WWW.example.test.:443/a"})
	must(e)
	beforeGame := queryInt64(t, ctx, game, `SELECT count(*) FROM gfg_game`)
	beforeSite := queryInt64(t, ctx, nav, `SELECT count(*) FROM gfn_site`)
	trace.queries.Store(0)
	preview, e := s.Preview(ctx, []collaboration.Input{{Kind: "game", Source: "424242"}, {Kind: "game", Source: "Steam:424242"}, {Kind: "site", Source: "example.test"}, {Kind: "other", Title: "manual"}, {Kind: "game"}})
	must(e)
	if trace.queries.Load() != 2 {
		t.Fatalf("business duplicate reads=%d want 2", trace.queries.Load())
	}
	if len(preview[0].Warnings) != 2 || len(preview[1].Warnings) != 3 || preview[2].ResourceMatch == nil || !preview[3].Valid || preview[4].Valid {
		t.Fatalf("preview=%+v", preview)
	}
	inputs := make([]collaboration.Input, 100)
	for i := range inputs {
		inputs[i] = collaboration.Input{Kind: "game", Source: fmt.Sprint(800000 + i)}
	}
	before := queryInt64(t, ctx, admin, `SELECT count(*) FROM gfa_content_idea`)
	result, e := s.Batch(ctx, meta, collaboration.BatchInput{Items: inputs})
	must(e)
	if result.InsertedCount != 100 || queryInt64(t, ctx, admin, `SELECT count(*) FROM gfa_content_idea`) != before+100 {
		t.Fatalf("batch=%+v", result)
	}
	result, e = s.Batch(ctx, meta, collaboration.BatchInput{Items: inputs})
	must(e)
	if result.InsertedCount != 0 || result.SkippedCount != 100 {
		t.Fatalf("skip known=%+v", result)
	}
	allow := false
	result, e = s.Batch(ctx, meta, collaboration.BatchInput{Items: inputs[:2], SkipKnown: &allow})
	must(e)
	if result.InsertedCount != 2 {
		t.Fatal("duplicates were not soft warnings")
	}
	if _, e = s.Preview(ctx, make([]collaboration.Input, 501)); e == nil || e.GetHTTPStatus() != 400 {
		t.Fatal("501 candidates accepted")
	}
	maxRows := make([]collaboration.Input, 500)
	for i := range maxRows {
		maxRows[i] = collaboration.Input{Kind: "other", Title: "inventory"}
	}
	if rows, e := s.Preview(ctx, maxRows); e != nil || len(rows) != 500 {
		t.Fatalf("500 preview: %d %v", len(rows), e)
	}
	if queryInt64(t, ctx, admin, `SELECT count(*) FROM gfa_admin_audit_log WHERE action='collaboration.idea.batch_create'`) != 2 {
		t.Fatal("bulk did not use one summary audit")
	}
	closeResponse(requestJSON(t, app, "POST", "/ideas", `{"kind":"game","title":"spoof","created_by_account_id":2}`, nil, 400))
	closeResponse(requestJSON(t, app, "PUT", fmt.Sprintf("/ideas/%d", idea.ID), `{"kind":"game","title":"spoof","version":1,"status":"landed"}`, nil, 400))
	researched, e := s.Transition(ctx, meta, idea.ID, "research", collaboration.Transition{Version: idea.Version})
	must(e)
	again, e := s.Transition(ctx, meta, idea.ID, "research", collaboration.Transition{Version: researched.Version})
	must(e)
	if again.Version != researched.Version {
		t.Fatal("same user research not idempotent")
	}
	if _, e = s.Transition(ctx, other, idea.ID, "research", collaboration.Transition{Version: researched.Version}); e == nil || e.GetHTTPStatus() != 409 || !strings.Contains(e.Error(), "Member One") {
		t.Fatal("research ownership conflict missing")
	}
	closeResponse(requestJSON(t, app, "POST", fmt.Sprintf("/ideas/%d/research", idea.ID), `{"version":1}`, nil, http.StatusConflict))
	released, e := s.Transition(ctx, other, idea.ID, "release", collaboration.Transition{Version: researched.Version})
	must(e)
	shelved, e := s.Transition(ctx, meta, idea.ID, "shelve", collaboration.Transition{Version: released.Version})
	must(e)
	restored, e := s.Transition(ctx, meta, idea.ID, "restore", collaboration.Transition{Version: shelved.Version})
	must(e)
	if _, e = s.Transition(ctx, meta, idea.ID, "land", collaboration.Transition{Version: restored.Version}); e == nil || e.GetHTTPStatus() != 409 {
		t.Fatal("Game landed without link")
	}
	if _, e = s.Transition(ctx, meta, idea.ID, "link", collaboration.Transition{Version: restored.Version, Kind: "game", ResourceID: 99999999}); e == nil || e.GetHTTPStatus() != 400 {
		t.Fatal("nonexistent link accepted")
	}
	landed, e := s.Transition(ctx, meta, idea.ID, "link", collaboration.Transition{Version: restored.Version, Kind: "game", ResourceID: gameID})
	must(e)
	if landed.Status != "landed" || *landed.LinkedResourceID != gameID {
		t.Fatal("Game link missing")
	}
	_, e = s.Transition(ctx, meta, siteIdea.ID, "link", collaboration.Transition{Version: siteIdea.Version, Kind: "site", ResourceID: siteID})
	must(e)
	reopened, e := s.Transition(ctx, meta, idea.ID, "reopen", collaboration.Transition{Version: landed.Version})
	must(e)
	if reopened.LinkedKind != nil || reopened.LandedAt.Valid {
		t.Fatal("reopen retained link")
	}
	free, e := s.Create(ctx, meta, collaboration.Input{Kind: "other", Title: "Other"})
	must(e)
	free, e = s.Transition(ctx, meta, free.ID, "land", collaboration.Transition{Version: free.Version})
	must(e)
	if free.Status != "landed" || free.LinkedResourceID != nil {
		t.Fatal("Other land failed")
	}
	// Concurrent edits must have exactly one winner, with the other returning 409.
	var winners, conflicts atomic.Int64
	var wg sync.WaitGroup
	for _, actor := range []audit.Meta{meta, other} {
		wg.Add(1)
		go func(m audit.Meta) {
			defer wg.Done()
			_, e := s.Update(ctx, m, reopened.ID, collaboration.Input{Kind: "game", Title: m.OperatorName, Version: reopened.Version})
			if e == nil {
				winners.Add(1)
			} else if e.GetHTTPStatus() == 409 {
				conflicts.Add(1)
			} else {
				t.Error(e)
			}
		}(actor)
	}
	wg.Wait()
	if winners.Load() != 1 || conflicts.Load() != 1 {
		t.Fatal("concurrent update silently overwrote")
	}
	testCollaborationCanvas(t, ctx, admin, s, app, meta, other, idea.ID, gameID, siteID, trace)
	// Force audit failure and prove all writes roll back on the GFA transaction.
	current, e := s.Get(ctx, idea.ID)
	must(e)
	_, err := admin.Exec(ctx, `ALTER TABLE gfa_admin_audit_log ADD CONSTRAINT collaboration_test_audit_failure CHECK (action NOT LIKE 'collaboration.%') NOT VALID`)
	must(err)
	before = queryInt64(t, ctx, admin, `SELECT count(*) FROM gfa_content_idea`)
	if _, e = s.Create(ctx, meta, collaboration.Input{Kind: "other", Title: "must rollback"}); e == nil {
		t.Fatal("audit failure ignored")
	}
	if _, e = s.Batch(ctx, meta, collaboration.BatchInput{Items: []collaboration.Input{{Kind: "other", Title: "batch rollback"}}}); e == nil {
		t.Fatal("batch audit failure ignored")
	}
	if _, e = s.Update(ctx, meta, current.ID, collaboration.Input{Kind: "game", Title: "must rollback update", Version: current.Version}); e == nil {
		t.Fatal("update audit failure ignored")
	}
	if _, e = s.Transition(ctx, meta, current.ID, "link", collaboration.Transition{Version: current.Version, Kind: "game", ResourceID: gameID}); e == nil {
		t.Fatal("link audit failure ignored")
	}
	if e = s.Delete(ctx, meta, current.ID, current.Version); e == nil {
		t.Fatal("idea delete audit failure ignored")
	}
	if queryInt64(t, ctx, admin, `SELECT version FROM gfa_content_idea WHERE id=$1`, current.ID) != current.Version {
		t.Fatal("audit failure changed an existing row")
	}
	if queryInt64(t, ctx, admin, `SELECT count(*) FROM gfa_content_idea`) != before {
		t.Fatal("audit failure left partial writes")
	}
	_, err = admin.Exec(ctx, `ALTER TABLE gfa_admin_audit_log DROP CONSTRAINT collaboration_test_audit_failure`)
	must(err)
	_, e = s.Transition(ctx, meta, current.ID, "link", collaboration.Transition{Version: current.Version, Kind: "game", ResourceID: gameID})
	must(e) // Recover using the same formal Game after the GFA-only failure.
	page, e := s.List(ctx, collaboration.Filters{Kind: "game", Status: "active", Keyword: "800", PageSize: 100}, 1)
	must(e)
	if page.Total != 102 || len(page.List) != 100 || page.List[0].CreatorName != "Member One" {
		t.Fatalf("filtered/paged inventory=%+v", page)
	}
	if _, e = s.List(ctx, collaboration.Filters{Researcher: "any-account"}, 1); e == nil || e.GetHTTPStatus() != 400 {
		t.Fatal("arbitrary account filter accepted")
	}
	inventory := workbench.New(nil, nil, nil, nil, nil, nil).WithCollaboration(s)
	summary := inventory.Summary(ctx, &authorization.Principal{Capabilities: []authorization.Capability{authorization.CollaborationRead}})
	if summary.Collaboration == nil || summary.Collaboration.ReserveCount < 100 || len(summary.Attention) != 0 {
		t.Fatalf("neutral inventory missing: %+v", summary)
	}
	hidden := inventory.Summary(ctx, &authorization.Principal{})
	if hidden.Collaboration != nil {
		t.Fatal("summary leaked without capability")
	}
	if queryInt64(t, ctx, game, `SELECT count(*) FROM gfg_game`) != beforeGame || queryInt64(t, ctx, nav, `SELECT count(*) FROM gfn_site`) != beforeSite {
		t.Fatal("Collaboration changed formal resources")
	}
	// Deleting a landed idea requires its current version and never deletes its formal resource.
	current, e = s.Get(ctx, idea.ID)
	must(e)
	deletePath := fmt.Sprintf("/ideas/%d", idea.ID)
	closeResponse(requestJSON(t, app, "DELETE", deletePath, `{}`, nil, 400))
	closeResponse(requestJSON(t, app, "DELETE", deletePath, `{"version":1}`, nil, 409))
	closeResponse(requestJSON(t, app, "DELETE", deletePath, fmt.Sprintf(`{"version":%d}`, current.Version), nil, 200))
	if queryInt64(t, ctx, admin, `SELECT count(*) FROM gfa_content_idea WHERE id=$1`, idea.ID) != 0 {
		t.Fatal("idea delete did not remove row")
	}
	if queryInt64(t, ctx, admin, `SELECT count(*) FROM gfa_admin_audit_log WHERE action='collaboration.idea.delete' AND target_id=$1 AND before_data <> ''`, fmt.Sprint(idea.ID)) != 1 {
		t.Fatal("idea delete audit missing")
	}
	if queryInt64(t, ctx, game, `SELECT count(*) FROM gfg_game WHERE id=$1`, gameID) != 1 || queryInt64(t, ctx, nav, `SELECT count(*) FROM gfn_site WHERE id=$1`, siteID) != 1 {
		t.Fatal("idea deletion removed a formal resource")
	}
	closeResponse(requestJSON(t, app, "DELETE", deletePath, fmt.Sprintf(`{"version":%d}`, current.Version), nil, 409))
	// Ensure HTTP response actors come from the principal, not a client field.
	resp := requestJSON(t, app, "POST", "/ideas", `{"kind":"other","title":"HTTP creation"}`, nil, 200)
	defer resp.Body.Close()
	var envelope integrationEnvelope
	must(json.NewDecoder(resp.Body).Decode(&envelope))
	var saved collaboration.Idea
	must(json.Unmarshal(envelope.Data, &saved))
	if saved.CreatedByAccountID != 1 {
		t.Fatal("wrong actor")
	}
}
