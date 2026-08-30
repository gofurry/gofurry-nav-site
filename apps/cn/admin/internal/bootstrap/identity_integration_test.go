package bootstrap_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GoFurry/easyhash"
	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
	authcontroller "github.com/gofurry/gofurry-admin/internal/app/auth/controller"
	authmw "github.com/gofurry/gofurry-admin/internal/app/auth/middleware"
	authmodels "github.com/gofurry/gofurry-admin/internal/app/auth/models"
	authservice "github.com/gofurry/gofurry-admin/internal/app/auth/service"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	"github.com/golang-jwt/jwt/v5"
)

func TestAdminIdentityAuthorizationPersistence(t *testing.T) {
	configPath := requireAdminIntegrationConfig(t)
	baseDSN := adminIntegrationDSN(t, configPath)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	adminDB := adminIntegrationSQLDB(t, baseDSN, "postgres")
	defer adminDB.Close()
	name := integrationDatabaseName("gfa_identity")
	createIntegrationDatabase(t, ctx, adminDB, name)
	defer dropIntegrationDatabase(t, adminDB, name)
	dsn := integrationDatabaseDSN(baseDSN, name)
	applyIntegrationBaseline(t, "admin", dsn)
	pool := integrationPool(t, ctx, dsn)
	defer pool.Close()

	auditLogger := audit.New(pool)
	authService := authservice.New(pool, auditLogger)
	authAPI := authcontroller.New(authService, auditLogger)
	app := identityTestApp(authService, authAPI)

	initialized, initErr := authService.IsInitialized(ctx)
	if initErr != nil || initialized {
		t.Fatalf("fresh identity initialized=%v err=%v", initialized, initErr)
	}
	ownerID := responseID(t, requestJSON(t, app, http.MethodPost, "/auth/bootstrap", `{"username":" OWNER ","display_name":"Owner","password":"owner-password"}`, nil, http.StatusOK))
	closeResponse(requestJSON(t, app, http.MethodPost, "/auth/bootstrap", `{"username":"other","display_name":"Other","password":"other-password"}`, nil, http.StatusConflict))
	closeResponse(requestJSON(t, app, http.MethodPost, "/auth/login", `{"username":"unknown","password":"owner-password"}`, nil, http.StatusUnauthorized))
	closeResponse(requestJSON(t, app, http.MethodPost, "/auth/login", `{"username":"owner","password":"wrong-password"}`, nil, http.StatusUnauthorized))
	ownerCookie := loginCookie(t, app, "owner", "owner-password")
	if count := queryInt64(t, ctx, pool, `SELECT count(*) FROM gfa_admin_account WHERE id=$1 AND username='owner' AND role='owner' AND status='active' AND last_login_at IS NOT NULL`, ownerID); count != 1 {
		t.Fatal("bootstrap/login did not create and update the initial Owner")
	}

	developerID := responseID(t, requestJSON(t, app, http.MethodPost, "/auth/accounts/", `{"username":"developer","display_name":"Developer","role":"developer","password":"developer-password"}`, ownerCookie, http.StatusOK))
	operatorID := responseID(t, requestJSON(t, app, http.MethodPost, "/auth/accounts/", `{"username":"operator","display_name":"Operator","role":"operator","password":"operator-password"}`, ownerCookie, http.StatusOK))
	secondOwnerID := responseID(t, requestJSON(t, app, http.MethodPost, "/auth/accounts/", `{"username":"owner-two","display_name":"Owner Two","role":"owner","password":"owner-two-password"}`, ownerCookie, http.StatusOK))

	developerCookie := loginCookie(t, app, "developer", "developer-password")
	operatorCookie := loginCookie(t, app, "operator", "operator-password")
	closeResponse(requestJSON(t, app, http.MethodPost, "/auth/accounts/", `{"username":"denied","display_name":"Denied","role":"operator","password":"denied-password"}`, developerCookie, http.StatusForbidden))
	closeResponse(requestJSON(t, app, http.MethodPost, "/auth/accounts/", `{"username":"denied","display_name":"Denied","role":"operator","password":"denied-password"}`, operatorCookie, http.StatusForbidden))
	assertRoleRouteMatrix(t, app, ownerCookie, developerCookie, operatorCookie)

	// Role changes are database-backed and revoke the old token.
	closeResponse(requestJSON(t, app, http.MethodPut, fmt.Sprintf("/auth/accounts/%d/role", developerID), `{"role":"operator"}`, ownerCookie, http.StatusOK))
	closeResponse(requestJSON(t, app, http.MethodGet, "/auth/me", "", developerCookie, http.StatusUnauthorized))
	developerAsOperator := loginCookie(t, app, "developer", "developer-password")
	closeResponse(requestJSON(t, app, http.MethodPut, fmt.Sprintf("/auth/accounts/%d/role", developerID), `{"role":"developer"}`, ownerCookie, http.StatusOK))
	closeResponse(requestJSON(t, app, http.MethodGet, "/auth/me", "", developerAsOperator, http.StatusUnauthorized))
	developerCookie = loginCookie(t, app, "developer", "developer-password")

	// Disabled accounts immediately lose access and cannot log in.
	closeResponse(requestJSON(t, app, http.MethodPut, fmt.Sprintf("/auth/accounts/%d/status", operatorID), `{"status":"disabled"}`, ownerCookie, http.StatusOK))
	closeResponse(requestJSON(t, app, http.MethodGet, "/auth/me", "", operatorCookie, http.StatusUnauthorized))
	closeResponse(requestJSON(t, app, http.MethodPost, "/auth/login", `{"username":"operator","password":"operator-password"}`, nil, http.StatusUnauthorized))
	closeResponse(requestJSON(t, app, http.MethodPut, fmt.Sprintf("/auth/accounts/%d/status", operatorID), `{"status":"active"}`, ownerCookie, http.StatusOK))
	operatorCookie = loginCookie(t, app, "operator", "operator-password")

	// Password reset and explicit revocation both invalidate existing sessions.
	closeResponse(requestJSON(t, app, http.MethodPost, fmt.Sprintf("/auth/accounts/%d/password", operatorID), `{"password":"operator-new-password"}`, ownerCookie, http.StatusOK))
	closeResponse(requestJSON(t, app, http.MethodGet, "/auth/me", "", operatorCookie, http.StatusUnauthorized))
	closeResponse(requestJSON(t, app, http.MethodPost, "/auth/login", `{"username":"operator","password":"operator-password"}`, nil, http.StatusUnauthorized))
	operatorCookie = loginCookie(t, app, "operator", "operator-new-password")
	closeResponse(requestJSON(t, app, http.MethodPost, fmt.Sprintf("/auth/accounts/%d/revoke-sessions", operatorID), `{}`, ownerCookie, http.StatusOK))
	closeResponse(requestJSON(t, app, http.MethodGet, "/auth/me", "", operatorCookie, http.StatusUnauthorized))

	// Expired JWTs fail before database authorization.
	expired := signedIdentityToken(t, ownerID, 1, time.Now().Add(-2*time.Hour), time.Now().Add(-time.Hour))
	closeResponse(requestJSON(t, app, http.MethodGet, "/auth/me", "", &http.Cookie{Name: env.GetServerConfig().Auth.CookieName, Value: expired}, http.StatusUnauthorized))

	// With two Owners, one can be demoted. The remaining active Owner cannot be
	// disabled or demoted afterward.
	closeResponse(requestJSON(t, app, http.MethodPut, fmt.Sprintf("/auth/accounts/%d/role", secondOwnerID), `{"role":"developer"}`, ownerCookie, http.StatusOK))
	closeResponse(requestJSON(t, app, http.MethodPut, fmt.Sprintf("/auth/accounts/%d/status", ownerID), `{"status":"disabled"}`, ownerCookie, http.StatusConflict))
	closeResponse(requestJSON(t, app, http.MethodPut, fmt.Sprintf("/auth/accounts/%d/role", ownerID), `{"role":"developer"}`, ownerCookie, http.StatusConflict))

	// Recreate a second Owner, then race two disables. Row locking must allow
	// exactly one request and leave exactly one active Owner.
	thirdOwnerID := responseID(t, requestJSON(t, app, http.MethodPost, "/auth/accounts/", `{"username":"owner-three","display_name":"Owner Three","role":"owner","password":"owner-three-password"}`, ownerCookie, http.StatusOK))
	meta := audit.SystemMeta("identity-concurrency-test")
	results := make(chan error, 2)
	var start sync.WaitGroup
	start.Add(1)
	for _, accountID := range []int64{ownerID, thirdOwnerID} {
		go func(id int64) {
			start.Wait()
			_, err := authService.ChangeStatus(context.Background(), id, authorization.StatusDisabled, meta)
			results <- err
		}(accountID)
	}
	start.Done()
	var successes, conflicts int
	for range 2 {
		err := <-results
		if err == nil {
			successes++
		} else if appErr, ok := err.(interface{ GetHTTPStatus() int }); ok && appErr.GetHTTPStatus() == http.StatusConflict {
			conflicts++
		} else {
			t.Fatalf("concurrent last-Owner result: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent disables success=%d conflict=%d", successes, conflicts)
	}
	if count := queryInt64(t, ctx, pool, `SELECT count(*) FROM gfa_admin_account WHERE role='owner' AND status='active'`); count != 1 {
		t.Fatalf("active Owner count=%d, want 1", count)
	}
	if count := queryInt64(t, ctx, pool, `SELECT count(*) FROM gfa_admin_audit_log WHERE operator_account_id IS NOT NULL AND operator_name <> '' AND operator_role IN ('owner','developer','operator')`); count < 10 {
		t.Fatalf("identity audit snapshot count=%d", count)
	}
	if count := queryInt64(t, ctx, pool, `SELECT count(*) FROM gfa_admin_audit_log WHERE before_data LIKE '%password_hash%' OR after_data LIKE '%password_hash%'`); count != 0 {
		t.Fatal("audit snapshots exposed password hashes")
	}
}

func TestAdminLegacyIdentityUpgrade(t *testing.T) {
	configPath := requireAdminIntegrationConfig(t)
	baseDSN := adminIntegrationDSN(t, configPath)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	adminDB := adminIntegrationSQLDB(t, baseDSN, "postgres")
	defer adminDB.Close()
	name := integrationDatabaseName("gfa_legacy")
	createIntegrationDatabase(t, ctx, adminDB, name)
	defer dropIntegrationDatabase(t, adminDB, name)
	dsn := integrationDatabaseDSN(baseDSN, name)
	applyIntegrationMigrationTo(t, "admin", dsn, "20260823000000")

	hashConfig := easyhash.DefaultPBKDF2()
	hashConfig.PBKDF2Iterations = env.GetServerConfig().Auth.PBKDF2Iterations
	hash, err := easyhash.CreatePBKDF2(hashConfig, "legacy-password")
	if err != nil {
		t.Fatal(err)
	}
	legacyPool := integrationPool(t, ctx, dsn)
	if _, err := legacyPool.Exec(ctx, `INSERT INTO gfa_admin_account(id,password_hash,session_version,created_at,updated_at,password_updated_at) VALUES (1,$1,7,'2026-08-01','2026-08-02','2026-08-03')`, hash); err != nil {
		t.Fatal(err)
	}
	if _, err := legacyPool.Exec(ctx, `INSERT INTO gfa_admin_audit_log(action,resource,target_id,operator,session_version,created_at) VALUES ('legacy','resource','1','admin',7,'2026-08-04')`); err != nil {
		t.Fatal(err)
	}
	legacyPool.Close()
	applyIntegrationBaseline(t, "admin", dsn)

	pool := integrationPool(t, ctx, dsn)
	defer pool.Close()
	var username, displayName, role, status, preservedHash string
	var sessionVersion int64
	var createdAt, updatedAt, passwordUpdatedAt time.Time
	if err := pool.QueryRow(ctx, `SELECT username,display_name,role,status,password_hash,session_version,created_at,updated_at,password_updated_at FROM gfa_admin_account WHERE id=1`).Scan(&username, &displayName, &role, &status, &preservedHash, &sessionVersion, &createdAt, &updatedAt, &passwordUpdatedAt); err != nil {
		t.Fatal(err)
	}
	if username != "owner" || displayName != "Owner" || role != "owner" || status != "active" || preservedHash != hash || sessionVersion != 7 {
		t.Fatalf("legacy account migration mismatch: %s %s %s %s session=%d", username, displayName, role, status, sessionVersion)
	}
	if !createdAt.Equal(time.Date(2026, 8, 1, 0, 0, 0, 0, time.Local)) && createdAt.Format(time.DateOnly) != "2026-08-01" {
		t.Fatalf("legacy created_at changed: %v", createdAt)
	}
	if updatedAt.Format(time.DateOnly) != "2026-08-02" || passwordUpdatedAt.Format(time.DateOnly) != "2026-08-03" {
		t.Fatalf("legacy timestamps changed: updated=%v password=%v", updatedAt, passwordUpdatedAt)
	}
	service := authservice.New(pool, audit.New(pool))
	if _, principal, loginErr := service.Login(ctx, "owner", "legacy-password"); loginErr != nil || principal.Role != authorization.RoleOwner {
		t.Fatalf("legacy password login failed: principal=%v err=%v", principal, loginErr)
	}
	if count := queryInt64(t, ctx, pool, `SELECT count(*) FROM gfa_admin_audit_log WHERE action='legacy' AND operator_account_id=1 AND operator_name='Owner' AND operator_role='owner'`); count != 1 {
		t.Fatal("legacy audit identity was not preserved and backfilled")
	}
}

func identityTestApp(authService *authservice.AuthService, authAPI *authcontroller.AuthAPI) *fiber.App {
	app := fiber.New()
	app.Get("/auth/state", authAPI.State)
	app.Post("/auth/bootstrap", authAPI.Bootstrap)
	app.Post("/auth/login", authAPI.Login)
	protected := app.Group("", authmw.Required(authService))
	protected.Get("/auth/me", authAPI.Me)
	accounts := protected.Group("/auth/accounts", authmw.Require(authorization.AccountManage))
	accounts.Get("/", authAPI.ListAccounts)
	accounts.Post("/", authAPI.CreateAccount)
	accounts.Put("/:id/display-name", authAPI.UpdateDisplayName)
	accounts.Put("/:id/role", authAPI.ChangeRole)
	accounts.Put("/:id/status", authAPI.ChangeStatus)
	accounts.Post("/:id/password", authAPI.ResetAccountPassword)
	accounts.Post("/:id/revoke-sessions", authAPI.RevokeSessions)
	for path, capability := range map[string]authorization.Capability{
		"/test/content-write":      authorization.ContentWrite,
		"/test/collection-execute": authorization.CollectionExecute,
		"/test/collection-control": authorization.CollectionControl,
		"/test/metrics-read":       authorization.MetricsRead,
		"/test/metrics-technical":  authorization.MetricsTechnical,
		"/test/changes-read":       authorization.ChangesRead,
		"/test/changes-technical":  authorization.ChangesTechnical,
		"/test/dataops-read":       authorization.DataOpsRead,
	} {
		protected.Get(path, authmw.Require(capability), func(c fiber.Ctx) error { return c.SendStatus(http.StatusOK) })
	}
	return app
}

func assertRoleRouteMatrix(t *testing.T, app *fiber.App, owner, developer, operator *http.Cookie) {
	t.Helper()
	for path, status := range map[string]int{
		"/test/content-write": http.StatusOK, "/test/collection-execute": http.StatusOK,
		"/test/collection-control": http.StatusForbidden, "/test/metrics-read": http.StatusOK,
		"/test/metrics-technical": http.StatusForbidden, "/test/changes-read": http.StatusOK,
		"/test/changes-technical": http.StatusForbidden, "/test/dataops-read": http.StatusForbidden,
	} {
		closeResponse(requestJSON(t, app, http.MethodGet, path, "", operator, status))
	}
	for _, path := range []string{"/test/content-write", "/test/collection-execute", "/test/collection-control", "/test/metrics-read", "/test/metrics-technical", "/test/changes-read", "/test/changes-technical", "/test/dataops-read"} {
		closeResponse(requestJSON(t, app, http.MethodGet, path, "", developer, http.StatusOK))
		closeResponse(requestJSON(t, app, http.MethodGet, path, "", owner, http.StatusOK))
	}
}

func loginCookie(t *testing.T, app *fiber.App, username, password string) *http.Cookie {
	t.Helper()
	response := requestJSON(t, app, http.MethodPost, "/auth/login", fmt.Sprintf(`{"username":%q,"password":%q}`, username, password), nil, http.StatusOK)
	defer response.Body.Close()
	cookies := response.Cookies()
	if len(cookies) == 0 {
		t.Fatal("login response did not set a cookie")
	}
	return cookies[0]
}

func signedIdentityToken(t *testing.T, accountID, sessionVersion int64, issuedAt, expiresAt time.Time) string {
	t.Helper()
	claims := authmodels.AdminClaims{AccountID: accountID, SessionVersion: sessionVersion, RegisteredClaims: jwt.RegisteredClaims{
		Issuer: env.GetServerConfig().Server.AppID, Subject: fmt.Sprintf("admin:%d", accountID),
		IssuedAt: jwt.NewNumericDate(issuedAt), NotBefore: jwt.NewNumericDate(issuedAt), ExpiresAt: jwt.NewNumericDate(expiresAt),
	}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(env.GetServerConfig().Auth.JWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func requireAdminIntegrationConfig(t *testing.T) string {
	t.Helper()
	path := strings.TrimSpace(os.Getenv("GOFURRY_ADMIN_INTEGRATION_CONFIG"))
	if path == "" {
		t.Skip("set GOFURRY_ADMIN_INTEGRATION_CONFIG for PostgreSQL integration tests")
	}
	return path
}

func applyIntegrationMigrationTo(t *testing.T, domain, dsn, version string) {
	t.Helper()
	repositoryRoot := integrationRepositoryRoot(t)
	command := exec.Command("go", "tool", "goose", "-dir", filepath.Join(repositoryRoot, "db", domain, "migrations"), "postgres", dsn, "up-to", version)
	command.Dir = filepath.Join(repositoryRoot, "tools")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("apply %s migration through %s: %v\n%s", domain, version, err, output)
	}
}

func closeResponse(response *http.Response) {
	if response != nil && response.Body != nil {
		_ = response.Body.Close()
	}
}
