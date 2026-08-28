package bootstrap_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/gofurry-admin/config"
	authcontroller "github.com/gofurry/gofurry-admin/internal/app/auth/controller"
	authmw "github.com/gofurry/gofurry-admin/internal/app/auth/middleware"
	authservice "github.com/gofurry/gofurry-admin/internal/app/auth/service"
	collectioncontroller "github.com/gofurry/gofurry-admin/internal/app/collectionadmin/controller"
	collectionservice "github.com/gofurry/gofurry-admin/internal/app/collectionadmin/service"
	gameadmin "github.com/gofurry/gofurry-admin/internal/app/gameadmin/controller"
	navadmin "github.com/gofurry/gofurry-admin/internal/app/navadmin/controller"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"gopkg.in/yaml.v3"
)

type integrationEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func TestAdminThreeDatabasePersistence(t *testing.T) {
	configPath := strings.TrimSpace(os.Getenv("GOFURRY_ADMIN_INTEGRATION_CONFIG"))
	if configPath == "" {
		t.Skip("set GOFURRY_ADMIN_INTEGRATION_CONFIG for PostgreSQL integration tests")
	}
	baseDSN := adminIntegrationDSN(t, configPath)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	adminDB := adminIntegrationSQLDB(t, baseDSN, "postgres")
	defer adminDB.Close()

	names := map[string]string{"admin": integrationDatabaseName("gfa"), "nav": integrationDatabaseName("gfn"), "game": integrationDatabaseName("gfg")}
	for _, name := range names {
		createIntegrationDatabase(t, ctx, adminDB, name)
		defer dropIntegrationDatabase(t, adminDB, name)
	}
	dsns := map[string]string{}
	for domain, name := range names {
		dsns[domain] = integrationDatabaseDSN(baseDSN, name)
		applyIntegrationBaseline(t, domain, dsns[domain])
	}

	adminPool := integrationPool(t, ctx, dsns["admin"])
	navPool := integrationPool(t, ctx, dsns["nav"])
	gamePool := integrationPool(t, ctx, dsns["game"])
	defer navPool.Close()
	defer gamePool.Close()
	defer adminPool.Close()
	assertCurrentDatabase(t, ctx, adminPool, names["admin"])
	assertCurrentDatabase(t, ctx, navPool, names["nav"])
	assertCurrentDatabase(t, ctx, gamePool, names["game"])

	auditLogger := audit.New(adminPool)
	authService := authservice.New(adminPool, auditLogger)
	authAPI := authcontroller.New(authService, auditLogger)
	navAPI := navadmin.New(navPool, auditLogger)
	gameAPI := gameadmin.New(gamePool, auditLogger)
	collectionAPI := collectioncontroller.New(collectionservice.New(gamePool, navPool, auditLogger))
	app := fiber.New()
	app.Post("/auth/bootstrap", authAPI.Bootstrap)
	app.Post("/auth/login", authAPI.Login)
	app.Post("/nav/audit-failure", navAPI.CreateSite)
	protected := app.Group("", authmw.Required(authService))
	protected.Post("/nav/sites", navAPI.CreateSite)
	protected.Get("/nav/sites/:id", navAPI.GetSite)
	protected.Put("/nav/sites/:id", navAPI.UpdateSite)
	protected.Delete("/nav/sites/:id", navAPI.DeleteSite)
	protected.Post("/nav/collector-domains", navAPI.CreateCollectorDomain)
	protected.Put("/nav/collector-domains/:id", navAPI.UpdateCollectorDomain)
	protected.Delete("/nav/collector-domains/:id", navAPI.DeleteCollectorDomain)
	protected.Post("/nav/collector-domains/:id/primary", navAPI.SetPrimaryCollectorDomain)
	protected.Post("/game/games", gameAPI.CreateGame)
	protected.Get("/game/games/:id", gameAPI.GetGame)
	protected.Put("/game/games/:id", gameAPI.UpdateGame)
	protected.Delete("/game/games/:id", gameAPI.DeleteGame)
	protected.Post("/game/prizes", gameAPI.CreatePrize)
	protected.Put("/game/prizes/:id", gameAPI.UpdatePrize)
	protected.Post("/collection/jobs", collectionAPI.CreateJobs)
	protected.Post("/collection/jobs/:domain/:id/cancel", collectionAPI.CancelJob)
	protected.Post("/collection/jobs/:domain/:id/retry", collectionAPI.RetryJob)

	requestJSON(t, app, http.MethodPost, "/auth/bootstrap", `{"password":"integration-password"}`, nil, http.StatusOK)
	login := requestJSON(t, app, http.MethodPost, "/auth/login", `{"password":"integration-password"}`, nil, http.StatusOK)
	if len(login.Cookies()) == 0 {
		t.Fatal("login response did not set the auth cookie")
	}
	cookie := login.Cookies()[0]

	site := requestJSON(t, app, http.MethodPost, "/nav/sites", `{"name":"站点","name_en":"Site","info":"简介","info_en":"Info","country":"CN","nsfw":"0","welfare":"0"}`, cookie, http.StatusOK)
	siteID := responseID(t, site)
	requestJSON(t, app, http.MethodPut, fmt.Sprintf("/nav/sites/%d", siteID), `{"name":"站点更新","name_en":"Updated Site","info":"简介","info_en":"Updated info","country":null,"nsfw":"0","welfare":"0","icon":null}`, cookie, http.StatusOK)
	requestJSON(t, app, http.MethodGet, fmt.Sprintf("/nav/sites/%d", siteID), "", cookie, http.StatusOK)
	requestJSON(t, app, http.MethodDelete, fmt.Sprintf("/nav/sites/%d", siteID), "", cookie, http.StatusOK)
	if got := queryBool(t, ctx, navPool, `SELECT deleted FROM gfn_site WHERE id=$1`, siteID); !got {
		t.Fatal("Nav delete did not preserve soft-delete behavior")
	}
	if count := queryInt64(t, ctx, navPool, `SELECT COUNT(*) FROM gfn_site WHERE id=$1 AND deleted_at IS NOT NULL`, siteID); count != 1 {
		t.Fatal("Nav soft delete did not record deleted_at")
	}

	trackedSite := requestJSON(t, app, http.MethodPost, "/nav/sites", `{"name":"Tracked","name_en":"Tracked","info":"x","info_en":"x","country":"CN","nsfw":"0","welfare":"0"}`, cookie, http.StatusOK)
	trackedSiteID := responseID(t, trackedSite)
	if count := queryInt64(t, ctx, navPool, `SELECT COUNT(*) FROM gfn_site_daily WHERE site_id=$1 AND fact_date=(transaction_timestamp() AT TIME ZONE 'UTC')::date AND finalized_at IS NULL`, trackedSiteID); count != 1 {
		t.Fatal("Site create did not write through today's Site Daily marker")
	}
	rootDomain := requestJSON(t, app, http.MethodPost, "/nav/collector-domains", fmt.Sprintf(`{"site_id":%d,"name":"example.test","proxy":"0","prefix":null,"tls":"1"}`, trackedSiteID), cookie, http.StatusOK)
	rootDomainID := responseID(t, rootDomain)
	wwwDomain := requestJSON(t, app, http.MethodPost, "/nav/collector-domains", fmt.Sprintf(`{"site_id":%d,"name":"example.test","proxy":"0","prefix":"www.","tls":"1"}`, trackedSiteID), cookie, http.StatusOK)
	wwwDomainID := responseID(t, wwwDomain)
	if count := queryInt64(t, ctx, navPool, `SELECT COUNT(*) FROM gfn_target_tracking_periods WHERE site_id=$1 AND tracked_until IS NULL`, trackedSiteID); count != 2 {
		t.Fatalf("active target period count=%d", count)
	}
	requestJSON(t, app, http.MethodPut, fmt.Sprintf("/nav/collector-domains/%d", rootDomainID), fmt.Sprintf(`{"site_id":%d,"name":"example.test","proxy":"1","prefix":null,"tls":"0"}`, trackedSiteID), cookie, http.StatusOK)
	if count := queryInt64(t, ctx, navPool, `SELECT COUNT(*) FROM gfn_target_tracking_periods WHERE collector_domain_id=$1`, rootDomainID); count != 1 {
		t.Fatalf("proxy/TLS-only update changed Target identity, period count=%d", count)
	}
	requestJSON(t, app, http.MethodPost, fmt.Sprintf("/nav/collector-domains/%d/primary", wwwDomainID), "", cookie, http.StatusOK)
	if count := queryInt64(t, ctx, navPool, `SELECT COUNT(*) FROM gfn_site_daily daily JOIN gfn_target_tracking_periods target ON target.id=daily.primary_target_tracking_period_id WHERE daily.site_id=$1 AND target.collector_domain_id=$2`, trackedSiteID, wwwDomainID); count != 1 {
		t.Fatal("Set Primary did not refresh today's Site Daily marker")
	}
	requestJSON(t, app, http.MethodPut, fmt.Sprintf("/nav/collector-domains/%d", wwwDomainID), fmt.Sprintf(`{"site_id":%d,"name":"example.test","proxy":"1","prefix":"api.","tls":"0"}`, trackedSiteID), cookie, http.StatusOK)
	if count := queryInt64(t, ctx, navPool, `SELECT COUNT(*) FROM gfn_target_tracking_periods WHERE collector_domain_id=$1`, wwwDomainID); count != 2 {
		t.Fatalf("identity change period count=%d", count)
	}
	requestJSON(t, app, http.MethodDelete, fmt.Sprintf("/nav/collector-domains/%d", wwwDomainID), "", cookie, http.StatusOK)
	if count := queryInt64(t, ctx, navPool, `SELECT COUNT(*) FROM gfn_site_primary_target_periods p JOIN gfn_target_tracking_periods t ON t.id=p.target_tracking_period_id WHERE p.site_id=$1 AND p.effective_until IS NULL AND t.collector_domain_id=$2`, trackedSiteID, rootDomainID); count != 1 {
		t.Fatal("deleting Primary Target did not select deterministic replacement")
	}
	requestJSON(t, app, http.MethodDelete, fmt.Sprintf("/nav/sites/%d", trackedSiteID), "", cookie, http.StatusOK)
	if count := queryInt64(t, ctx, navPool, `SELECT COUNT(*) FROM gfn_target_tracking_periods WHERE site_id=$1 AND tracked_until IS NULL`, trackedSiteID); count != 0 {
		t.Fatal("deleting Site left active target periods")
	}

	gamePayload := `{"name":"Game","name_en":"Game EN","info":"Info","info_en":"Info EN","resources":[],"groups":[],"developers":[],"publishers":[],"appid":424242,"header":"","links":[],"weight":5,"primary_tag":0,"secondary_tag":0}`
	game := requestJSON(t, app, http.MethodPost, "/game/games", gamePayload, cookie, http.StatusOK)
	gameID := responseID(t, game)
	if count := queryInt64(t, ctx, gamePool, `SELECT COUNT(*) FROM gfg_game_tracking_periods WHERE game_id=$1 AND tracked_until IS NULL`, gameID); count != 1 {
		t.Fatal("Game create did not open tracking period")
	}
	if count := queryInt64(t, ctx, gamePool, `SELECT COUNT(*) FROM gfg_game_daily WHERE game_id=$1 AND fact_date=(transaction_timestamp() AT TIME ZONE 'UTC')::date AND finalized_at IS NULL`, gameID); count != 1 {
		t.Fatal("Game create did not write through today's Game Daily marker")
	}
	if count := queryInt64(t, ctx, gamePool, `SELECT COUNT(*) FROM gfg_collection_jobs WHERE scope_id=$1 AND trigger='entity_created' AND tasks=ARRAY['details','news']::text[]`, gameID); count != 1 {
		t.Fatalf("game creation durable job count=%d", count)
	}
	manualPayload := fmt.Sprintf(`{"domain":"game","scope_type":"game","scope_id":%d,"tasks":["details"]}`, gameID)
	firstManual := responseFirstID(t, requestJSON(t, app, http.MethodPost, "/collection/jobs", manualPayload, cookie, http.StatusOK))
	secondManual := responseFirstID(t, requestJSON(t, app, http.MethodPost, "/collection/jobs", manualPayload, cookie, http.StatusOK))
	if firstManual != secondManual {
		t.Fatalf("active manual dedupe returned jobs %d and %d", firstManual, secondManual)
	}
	if count := queryInt64(t, ctx, gamePool, `SELECT COUNT(*) FROM gfg_collection_jobs WHERE dedupe_key=$1 AND status IN ('queued','running')`, fmt.Sprintf("game:game:%d:details", gameID)); count != 1 {
		t.Fatalf("active manual dedupe count=%d", count)
	}
	requestJSON(t, app, http.MethodPost, fmt.Sprintf("/collection/jobs/game/%d/cancel", firstManual), "", cookie, http.StatusOK)
	retried := responseID(t, requestJSON(t, app, http.MethodPost, fmt.Sprintf("/collection/jobs/game/%d/retry", firstManual), "", cookie, http.StatusOK))
	if retried != firstManual {
		t.Fatalf("state-refresh retry created job %d, want same job %d", retried, firstManual)
	}
	playersPayload := fmt.Sprintf(`{"domain":"game","scope_type":"game","scope_id":%d,"tasks":["players"]}`, gameID)
	playersJob := responseFirstID(t, requestJSON(t, app, http.MethodPost, "/collection/jobs", playersPayload, cookie, http.StatusOK))
	requestJSON(t, app, http.MethodPost, fmt.Sprintf("/collection/jobs/game/%d/cancel", playersJob), "", cookie, http.StatusOK)
	requestJSON(t, app, http.MethodPost, fmt.Sprintf("/collection/jobs/game/%d/retry", playersJob), "", cookie, http.StatusBadRequest)
	responseFirstID(t, requestJSON(t, app, http.MethodPost, "/collection/jobs", `{"domain":"nav","scope_type":"all","tasks":["ping","robots"]}`, cookie, http.StatusOK))
	if count := queryInt64(t, ctx, navPool, `SELECT COUNT(*) FROM gfn_collection_jobs WHERE trigger='manual' AND scope_type='all' AND job_key IN ('nav.ping','nav.robots')`); count != 2 {
		t.Fatalf("Nav multi-protocol fan-out job count=%d", count)
	}
	requestJSON(t, app, http.MethodPost, "/game/games", gamePayload, cookie, http.StatusBadRequest)
	for _, statement := range []string{
		`INSERT INTO gfg_game_details (game_id,appid,collected_at) VALUES ($1,424242,now())`,
		`INSERT INTO gfg_game_release_state (game_id,availability,precision,source,source_region,source_locale,normalizer_version,observed_at) VALUES ($1,'unknown','none','steam','US','en','steam-go/v1.3.9',now())`,
		`INSERT INTO gfg_game_release_history (game_id,availability,precision,source,source_region,source_locale,normalizer_version,observed_at) VALUES ($1,'unknown','none','steam','US','en','steam-go/v1.3.9',now())`,
		`INSERT INTO gfg_game_first_available (game_id,precision,exact_date,release_year,release_month,window_start,window_end,source,inferred,normalizer_version) VALUES ($1,'day','2020-01-02',2020,1,'2020-01-02','2020-01-02','legacy_manual',false,'gofurry-legacy-release/v1')`,
		`INSERT INTO gfg_game_languages (game_id,language_code,steam_name,tier,sort_order,source,source_region,source_locale,normalizer_version,observed_at) VALUES ($1,'en','English','platform',0,'steam','US','en','steam-go/v1.3.9',now())`,
	} {
		if _, err := gamePool.Exec(ctx, statement, gameID); err != nil {
			t.Fatal(err)
		}
	}
	updatedGamePayload := `{"name":"Game Updated","name_en":"Game EN","info":"Info","info_en":"Info EN","resources":[],"groups":[],"developers":[],"publishers":[],"appid":424242,"header":"","links":[],"weight":0,"primary_tag":0,"secondary_tag":0}`
	requestJSON(t, app, http.MethodPut, fmt.Sprintf("/game/games/%d", gameID), updatedGamePayload, cookie, http.StatusOK)
	if count := queryInt64(t, ctx, gamePool, `SELECT COUNT(*) FROM gfg_game_release_state WHERE game_id=$1`, gameID); count != 1 {
		t.Fatalf("unchanged AppID unexpectedly reset canonical state, count=%d", count)
	}
	changedAppIDPayload := `{"name":"Game Updated","name_en":"Game EN","info":"Info","info_en":"Info EN","resources":[],"groups":[],"developers":[],"publishers":[],"appid":424243,"header":"","links":[],"weight":0,"primary_tag":0,"secondary_tag":0}`
	requestJSON(t, app, http.MethodPut, fmt.Sprintf("/game/games/%d", gameID), changedAppIDPayload, cookie, http.StatusOK)
	if count := queryInt64(t, ctx, gamePool, `SELECT COUNT(*) FROM gfg_collection_jobs WHERE scope_id=$1 AND trigger='entity_changed' AND tasks=ARRAY['details','news']::text[]`, gameID); count != 1 {
		t.Fatalf("AppID change durable recollect job count=%d", count)
	}
	if count := queryInt64(t, ctx, gamePool, `SELECT
  (SELECT COUNT(*) FROM gfg_game_details WHERE game_id=$1) +
  (SELECT COUNT(*) FROM gfg_game_release_state WHERE game_id=$1) +
  (SELECT COUNT(*) FROM gfg_game_release_history WHERE game_id=$1) +
  (SELECT COUNT(*) FROM gfg_game_first_available WHERE game_id=$1) +
	  (SELECT COUNT(*) FROM gfg_game_languages WHERE game_id=$1)`, gameID); count != 1 {
		t.Fatalf("changed AppID current reset/history preservation count=%d", count)
	}
	if count := queryInt64(t, ctx, gamePool, `SELECT COUNT(*) FROM gfg_game_tracking_periods WHERE game_id=$1`, gameID); count != 2 {
		t.Fatalf("AppID change tracking period count=%d", count)
	}
	if count := queryInt64(t, ctx, gamePool, `SELECT COUNT(*) FROM gfg_game_daily WHERE game_id=$1 AND appid=424243 AND fact_date=(transaction_timestamp() AT TIME ZONE 'UTC')::date`, gameID); count != 1 {
		t.Fatal("AppID change did not refresh today's Game Daily marker")
	}
	requestJSON(t, app, http.MethodGet, fmt.Sprintf("/game/games/%d", gameID), "", cookie, http.StatusOK)

	prize := requestJSON(t, app, http.MethodPost, "/game/prizes", `{"title":"Prize","desc":"Desc","prize":{"keys":[],"title":"Reward","platform":"Steam"},"key":"join","start_time":"2026-08-23 10:00:00","end_time":"2026-08-24 10:00:00","status":true}`, cookie, http.StatusOK)
	prizeID := responseID(t, prize)
	requestJSON(t, app, http.MethodPut, fmt.Sprintf("/game/prizes/%d", prizeID), `{"title":"Prize","desc":"Desc","prize":{"keys":[],"title":"Reward","platform":"Steam"},"key":"join","start_time":"2026-08-23 10:00:00","end_time":"2026-08-24 10:00:00","status":false}`, cookie, http.StatusOK)
	if got := queryBool(t, ctx, gamePool, `SELECT status FROM gfg_prize WHERE id=$1`, prizeID); got {
		t.Fatal("prize status=false was not persisted")
	}
	requestJSON(t, app, http.MethodDelete, fmt.Sprintf("/game/games/%d", gameID), "", cookie, http.StatusOK)
	if count := queryInt64(t, ctx, gamePool, `SELECT COUNT(*) FROM gfg_game WHERE id=$1`, gameID); count != 0 {
		t.Fatalf("Game hard delete left %d rows", count)
	}
	if count := queryInt64(t, ctx, gamePool, `SELECT COUNT(*) FROM gfg_game_release_history WHERE game_id=$1`, gameID); count != 1 {
		t.Fatalf("Game delete removed historical release rows, count=%d", count)
	}
	if count := queryInt64(t, ctx, gamePool, `SELECT COUNT(*) FROM gfg_game_tracking_periods WHERE game_id=$1 AND tracked_until IS NULL`, gameID); count != 0 {
		t.Fatal("Game delete left an active tracking period")
	}
	secondGamePayload := `{"name":"Game 2","name_en":"Game 2","info":"Info","info_en":"Info","resources":[],"groups":[],"developers":[],"publishers":[],"appid":424244,"header":"","links":[],"weight":0,"primary_tag":0,"secondary_tag":0}`
	secondGameID := responseID(t, requestJSON(t, app, http.MethodPost, "/game/games", secondGamePayload, cookie, http.StatusOK))
	if secondGameID <= gameID {
		t.Fatalf("Game sequence reused or regressed id: first=%d second=%d", gameID, secondGameID)
	}
	if count := queryInt64(t, ctx, adminPool, `SELECT COUNT(*) FROM gfa_admin_audit_log`); count < 10 {
		t.Fatalf("expected auth and CRUD audit entries, got %d", count)
	}

	// Prove that gfn and gfa are not treated as a distributed transaction: a
	// failed gfa audit must roll back the still-open gfn business transaction.
	adminPool.Close()
	requestJSON(t, app, http.MethodPost, "/nav/audit-failure", `{"name":"Rollback","name_en":"Rollback","info":"x","info_en":"x","nsfw":"0","welfare":"0"}`, nil, http.StatusInternalServerError)
	if count := queryInt64(t, ctx, navPool, `SELECT COUNT(*) FROM gfn_site WHERE name='Rollback'`); count != 0 {
		t.Fatalf("cross-database audit failure left %d Nav rows", count)
	}
}

func requestJSON(t *testing.T, app *fiber.App, method, path, body string, cookie *http.Cookie, expectedStatus int) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = bytes.NewBufferString(body)
	}
	req := httptest.NewRequest(method, "http://admin.test"+path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != expectedStatus {
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("%s %s status=%d want=%d body=%s", method, path, resp.StatusCode, expectedStatus, data)
	}
	return resp
}

func responseID(t *testing.T, resp *http.Response) int64 {
	t.Helper()
	defer resp.Body.Close()
	var envelope integrationEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	var data struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.ID <= 0 {
		t.Fatalf("invalid response id: %s", envelope.Data)
	}
	return data.ID
}

func responseFirstID(t *testing.T, resp *http.Response) int64 {
	t.Helper()
	defer resp.Body.Close()
	var envelope integrationEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	var data []struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 || data[0].ID <= 0 {
		t.Fatalf("invalid job response: %s", envelope.Data)
	}
	return data[0].ID
}

func adminIntegrationDSN(t *testing.T, configPath string) string {
	t.Helper()
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Database env.DataBaseConfig `yaml:"database"`
	}
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		t.Fatal(err)
	}
	database := cfg.Database.Postgres
	if database.DBHost == "" && database.DSN == "" {
		database = env.SQLDataBaseConfig{DSN: cfg.Database.DSN, DBName: cfg.Database.DBName, DBHost: cfg.Database.DBHost, DBPort: cfg.Database.DBPort, DBUser: cfg.Database.DBUser, DBPass: cfg.Database.DBPass}
	}
	return database.ConnectionString()
}

func adminIntegrationSQLDB(t *testing.T, dsn, databaseName string) *sql.DB {
	t.Helper()
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.Database = databaseName
	return stdlib.OpenDB(*config)
}

func integrationDatabaseDSN(baseDSN, databaseName string) string {
	u, err := url.Parse(baseDSN)
	if err != nil {
		panic(err)
	}
	u.Path = databaseName
	return u.String()
}

func integrationDatabaseName(domain string) string {
	var value [5]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic(err)
	}
	return "gofurry_admin_" + domain + "_it_" + hex.EncodeToString(value[:])
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
	if !strings.HasPrefix(name, "gofurry_admin_") || !strings.Contains(name, "_it_") || strings.ContainsAny(name, `"'; \\`) {
		t.Fatalf("unsafe temporary database name %q", name)
	}
}

func quoteIntegrationIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func applyIntegrationBaseline(t *testing.T, domain, dsn string) {
	t.Helper()
	repositoryRoot := integrationRepositoryRoot(t)
	command := exec.Command("go", "tool", "goose", "-dir", filepath.Join(repositoryRoot, "db", domain, "migrations"), "postgres", dsn, "up")
	command.Dir = filepath.Join(repositoryRoot, "tools")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("apply %s baseline: %v\n%s", domain, err, output)
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

func integrationPool(t *testing.T, ctx context.Context, dsn string) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	return pool
}

func assertCurrentDatabase(t *testing.T, ctx context.Context, pool *pgxpool.Pool, expected string) {
	t.Helper()
	var actual string
	if err := pool.QueryRow(ctx, `SELECT current_database()`).Scan(&actual); err != nil {
		t.Fatal(err)
	}
	if actual != expected {
		t.Fatalf("connected to %q, want %q", actual, expected)
	}
}

func queryBool(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, args ...any) bool {
	t.Helper()
	var value bool
	if err := pool.QueryRow(ctx, query, args...).Scan(&value); err != nil {
		t.Fatal(err)
	}
	return value
}

func queryInt64(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, args ...any) int64 {
	t.Helper()
	var value int64
	if err := pool.QueryRow(ctx, query, args...).Scan(&value); err != nil {
		t.Fatal(err)
	}
	return value
}
