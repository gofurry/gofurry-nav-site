package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/GoFurry/easyhash"
	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
	"github.com/gofurry/gofurry-admin/internal/app/auth/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	adminsqlc "github.com/gofurry/gofurry-admin/internal/db/admin/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{2,63}$`)

type AuthService struct {
	pool  *pgxpool.Pool
	q     *adminsqlc.Queries
	audit *audit.Logger
}

func New(pool *pgxpool.Pool, auditLogger *audit.Logger) *AuthService {
	return &AuthService{pool: pool, q: adminsqlc.New(pool), audit: auditLogger}
}

func (s *AuthService) IsInitialized(ctx context.Context) (bool, common.Error) {
	count, err := s.q.CountAdminAccounts(ctx)
	if err != nil {
		return false, daoError(err)
	}
	return count > 0, nil
}

func (s *AuthService) Bootstrap(ctx context.Context, request models.BootstrapRequest, meta audit.Meta) (models.AccountResponse, common.Error) {
	username, validationErr := validateUsername(request.Username)
	if validationErr != nil {
		return models.AccountResponse{}, validationErr
	}
	displayName, validationErr := validateDisplayName(request.DisplayName)
	if validationErr != nil {
		return models.AccountResponse{}, validationErr
	}
	if validationErr = validatePassword(request.Password); validationErr != nil {
		return models.AccountResponse{}, validationErr
	}
	hash, hashErr := s.createPasswordHash(request.Password)
	if hashErr != nil {
		return models.AccountResponse{}, hashErr
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.AccountResponse{}, daoError(err)
	}
	defer tx.Rollback(ctx)
	queries := s.q.WithTx(tx)
	if _, err := queries.LockAdminBootstrap(ctx); err != nil {
		return models.AccountResponse{}, daoError(err)
	}
	count, err := queries.CountAdminAccounts(ctx)
	if err != nil {
		return models.AccountResponse{}, daoError(err)
	}
	if count > 0 {
		return models.AccountResponse{}, common.NewError(common.RETURN_FAILED, http.StatusConflict, "admin identity has already been initialized")
	}

	now := time.Now()
	row, err := queries.InsertAdminAccount(ctx, adminsqlc.InsertAdminAccountParams{
		Username: username, DisplayName: displayName, Role: string(authorization.RoleOwner),
		Status: string(authorization.StatusActive), PasswordHash: hash, SessionVersion: 1,
		PasswordUpdatedAt: timestamp(now),
	})
	if err != nil {
		return models.AccountResponse{}, accountWriteError(err)
	}
	account, convertErr := accountFromValues(row.ID, row.Username, row.DisplayName, row.Role, row.Status,
		row.PasswordHash, row.SessionVersion, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.PasswordUpdatedAt)
	if convertErr != nil {
		return models.AccountResponse{}, convertErr
	}
	principal := principalForAccount(account)
	if auditErr := s.audit.LogTx(ctx, tx, audit.MetaForPrincipal(meta, principal), "account.bootstrap", "gfa_admin_account", account.ID, nil, accountAuditSnapshot(account)); auditErr != nil {
		return models.AccountResponse{}, auditErr
	}
	if err := tx.Commit(ctx); err != nil {
		return models.AccountResponse{}, daoError(err)
	}
	return models.AccountDTO(*account), nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, *authorization.Principal, common.Error) {
	username = canonicalUsername(username)
	if !usernamePattern.MatchString(username) || password == "" {
		return "", nil, invalidCredentials()
	}
	account, err := s.getAccountByUsername(ctx, username)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil, invalidCredentials()
	}
	if err != nil {
		return "", nil, daoError(err)
	}
	if account.Status != authorization.StatusActive {
		return "", nil, invalidCredentials()
	}
	ok, verifyErr := easyhash.VerifyPBKDF2(password, account.PasswordHash)
	if verifyErr != nil {
		return "", nil, common.NewServiceError("password verification failed")
	}
	if !ok {
		return "", nil, invalidCredentials()
	}
	row, err := s.q.UpdateAdminLastLogin(ctx, account.ID)
	if err != nil {
		return "", nil, daoError(err)
	}
	account, convertErr := accountFromValues(row.ID, row.Username, row.DisplayName, row.Role, row.Status,
		row.PasswordHash, row.SessionVersion, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.PasswordUpdatedAt)
	if convertErr != nil {
		return "", nil, convertErr
	}
	principal := principalForAccount(account)
	token, tokenErr := s.buildToken(principal)
	if tokenErr != nil {
		return "", nil, tokenErr
	}
	return token, principal, nil
}

func (s *AuthService) ParseAndValidateToken(ctx context.Context, tokenValue string) (*models.AdminClaims, *authorization.Principal, common.Error) {
	tokenValue = strings.TrimSpace(tokenValue)
	if tokenValue == "" {
		return nil, nil, unauthorized("not logged in")
	}
	cfg := env.GetServerConfig().Auth
	token, err := jwt.ParseWithClaims(tokenValue, &models.AdminClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method %s", token.Method.Alg())
		}
		return []byte(cfg.JWTSecret), nil
	}, jwt.WithIssuer(env.GetServerConfig().Server.AppID))
	if err != nil {
		return nil, nil, unauthorized("login state is invalid")
	}
	claims, ok := token.Claims.(*models.AdminClaims)
	if !ok || !token.Valid || claims.AccountID <= 0 || claims.Subject != fmt.Sprintf("admin:%d", claims.AccountID) {
		return nil, nil, unauthorized("login state is invalid")
	}
	account, accountErr := s.getAccountByID(ctx, claims.AccountID)
	if errors.Is(accountErr, pgx.ErrNoRows) {
		return nil, nil, unauthorized("login state is invalid")
	}
	if accountErr != nil {
		return nil, nil, daoError(accountErr)
	}
	if account.Status != authorization.StatusActive {
		return nil, nil, unauthorized("login state is invalid")
	}
	if claims.SessionVersion != account.SessionVersion {
		return nil, nil, unauthorized("login state has expired")
	}
	return claims, principalForAccount(account), nil
}

func (s *AuthService) ResetPasswordByUsername(ctx context.Context, username, password string, meta audit.Meta) common.Error {
	username, validationErr := validateUsername(username)
	if validationErr != nil {
		return validationErr
	}
	if validationErr = validatePassword(password); validationErr != nil {
		return validationErr
	}
	hash, hashErr := s.createPasswordHash(password)
	if hashErr != nil {
		return hashErr
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return daoError(err)
	}
	defer tx.Rollback(ctx)
	queries := s.q.WithTx(tx)
	row, err := queries.GetAdminAccountByUsername(ctx, username)
	if errors.Is(err, pgx.ErrNoRows) {
		return accountNotFound()
	}
	if err != nil {
		return daoError(err)
	}
	before, convertErr := accountFromValues(row.ID, row.Username, row.DisplayName, row.Role, row.Status,
		row.PasswordHash, row.SessionVersion, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.PasswordUpdatedAt)
	if convertErr != nil {
		return convertErr
	}
	now := time.Now()
	updated, err := queries.UpdateAdminAccountPassword(ctx, adminsqlc.UpdateAdminAccountPasswordParams{
		PasswordHash: hash, PasswordUpdatedAt: timestamp(now), AccountID: before.ID,
	})
	if err != nil {
		return accountWriteError(err)
	}
	after, convertErr := accountFromValues(updated.ID, updated.Username, updated.DisplayName, updated.Role, updated.Status,
		updated.PasswordHash, updated.SessionVersion, updated.LastLoginAt, updated.CreatedAt, updated.UpdatedAt, updated.PasswordUpdatedAt)
	if convertErr != nil {
		return convertErr
	}
	if auditErr := s.audit.LogTx(ctx, tx, meta, "account.password_reset", "gfa_admin_account", after.ID, accountAuditSnapshot(before), accountAuditSnapshot(after)); auditErr != nil {
		return auditErr
	}
	if err := tx.Commit(ctx); err != nil {
		return daoError(err)
	}
	return nil
}

func (s *AuthService) BuildAuthCookie(token string) *fiber.Cookie {
	cfg := env.GetServerConfig().Auth
	return &fiber.Cookie{Name: cfg.CookieName, Value: token, Path: "/", Domain: cfg.CookieDomain, SameSite: cfg.SameSite, MaxAge: cfg.CookieMaxAgeSecs, Secure: cfg.CookieSecure, HTTPOnly: true}
}

func (s *AuthService) BuildLogoutCookie() *fiber.Cookie {
	cfg := env.GetServerConfig().Auth
	return &fiber.Cookie{Name: cfg.CookieName, Value: "", Path: "/", Domain: cfg.CookieDomain, SameSite: cfg.SameSite, MaxAge: -1, Secure: cfg.CookieSecure, HTTPOnly: true, Expires: time.Unix(0, 0)}
}

func (s *AuthService) buildToken(principal *authorization.Principal) (string, common.Error) {
	cfg := env.GetServerConfig().Auth
	now := time.Now()
	claims := &models.AdminClaims{AccountID: principal.AccountID, SessionVersion: principal.SessionVersion, RegisteredClaims: jwt.RegisteredClaims{
		Issuer: env.GetServerConfig().Server.AppID, Subject: fmt.Sprintf("admin:%d", principal.AccountID), IssuedAt: jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(cfg.SessionTTLHours) * time.Hour)),
	}}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", common.NewServiceError("failed to generate login token")
	}
	return signed, nil
}

func (s *AuthService) createPasswordHash(password string) (string, common.Error) {
	cfg := easyhash.DefaultPBKDF2()
	cfg.PBKDF2Iterations = env.GetServerConfig().Auth.PBKDF2Iterations
	hash, err := easyhash.CreatePBKDF2(cfg, password)
	if err != nil {
		return "", common.NewServiceError("failed to generate password hash")
	}
	return hash, nil
}

func (s *AuthService) getAccountByID(ctx context.Context, accountID int64) (*models.AdminAccount, error) {
	row, err := s.q.GetAdminAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	account, appErr := accountFromValues(row.ID, row.Username, row.DisplayName, row.Role, row.Status,
		row.PasswordHash, row.SessionVersion, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.PasswordUpdatedAt)
	if appErr != nil {
		return nil, appErr
	}
	return account, nil
}

func (s *AuthService) getAccountByUsername(ctx context.Context, username string) (*models.AdminAccount, error) {
	row, err := s.q.GetAdminAccountByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	account, appErr := accountFromValues(row.ID, row.Username, row.DisplayName, row.Role, row.Status,
		row.PasswordHash, row.SessionVersion, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.PasswordUpdatedAt)
	if appErr != nil {
		return nil, appErr
	}
	return account, nil
}

func accountFromValues(id int64, username, displayName, roleValue, statusValue, passwordHash string, sessionVersion int64,
	lastLoginAt, createdAt, updatedAt, passwordUpdatedAt pgtype.Timestamp,
) (*models.AdminAccount, common.Error) {
	role, ok := authorization.ParseRole(roleValue)
	if !ok {
		return nil, common.NewServiceError("admin account has an unsupported role")
	}
	status, ok := authorization.ParseStatus(statusValue)
	if !ok {
		return nil, common.NewServiceError("admin account has an unsupported status")
	}
	return &models.AdminAccount{
		ID: id, Username: username, DisplayName: displayName, Role: role, Status: status,
		PasswordHash: passwordHash, SessionVersion: sessionVersion, LastLoginAt: nullableTime(lastLoginAt),
		CreatedAt: createdAt.Time, UpdatedAt: updatedAt.Time, PasswordUpdatedAt: nullableTime(passwordUpdatedAt),
	}, nil
}

func principalForAccount(account *models.AdminAccount) *authorization.Principal {
	return &authorization.Principal{
		AccountID: account.ID, Username: account.Username, DisplayName: account.DisplayName,
		Role: account.Role, Status: account.Status, SessionVersion: account.SessionVersion,
		Capabilities: authorization.CapabilitiesFor(account.Role),
	}
}

func accountAuditSnapshot(account *models.AdminAccount) map[string]any {
	if account == nil {
		return nil
	}
	return map[string]any{
		"id": account.ID, "username": account.Username, "display_name": account.DisplayName,
		"role": account.Role, "status": account.Status, "session_version": account.SessionVersion,
		"last_login_at": account.LastLoginAt, "password_updated_at": account.PasswordUpdatedAt,
	}
}

func validateUsername(value string) (string, common.Error) {
	username := canonicalUsername(value)
	if !usernamePattern.MatchString(username) {
		return "", common.NewValidationError("username must be 3-64 lower-case letters, numbers, dots, underscores, or hyphens")
	}
	return username, nil
}

func canonicalUsername(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validateDisplayName(value string) (string, common.Error) {
	displayName := strings.TrimSpace(value)
	if displayName == "" || len(displayName) > 128 {
		return "", common.NewValidationError("display_name must be 1-128 characters")
	}
	return displayName, nil
}

func validatePassword(value string) common.Error {
	if value == "" {
		return common.NewValidationError("password must not be empty")
	}
	return nil
}

func timestamp(value time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: value, Valid: !value.IsZero()}
}

func nullableTime(value pgtype.Timestamp) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func invalidCredentials() common.Error {
	return unauthorized("invalid username or password")
}

func unauthorized(message string) common.Error {
	return common.NewError(common.RETURN_FAILED, http.StatusUnauthorized, message)
}

func accountNotFound() common.Error {
	return common.NewError(common.RETURN_FAILED, http.StatusNotFound, "admin account was not found")
}

func accountWriteError(err error) common.Error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return common.NewError(common.RETURN_FAILED, http.StatusConflict, "username is already in use")
		case "23514", "23502":
			return common.NewValidationError("admin account data is invalid")
		}
	}
	return daoError(err)
}

func daoError(err error) common.Error {
	if appErr, ok := err.(common.Error); ok {
		return appErr
	}
	return common.NewDaoError(err.Error())
}
