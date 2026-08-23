package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/GoFurry/easyhash"
	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/app/auth/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	adminsqlc "github.com/gofurry/gofurry-admin/internal/db/admin/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const ClaimsContextKey = "auth_claims"

type AuthService struct {
	pool  *pgxpool.Pool
	q     *adminsqlc.Queries
	audit *audit.Logger
}

func New(pool *pgxpool.Pool, auditLogger *audit.Logger) *AuthService {
	return &AuthService{pool: pool, q: adminsqlc.New(pool), audit: auditLogger}
}

func (s *AuthService) IsInitialized() (bool, common.Error) {
	_, err := s.getAccount(context.Background())
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, pgx.ErrNoRows):
		return false, nil
	default:
		return false, daoError(err)
	}
}

func (s *AuthService) Bootstrap(password string, meta audit.Meta) common.Error {
	password = strings.TrimSpace(password)
	if password == "" {
		return common.NewValidationError("password must not be empty")
	}
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return daoError(err)
	}
	defer tx.Rollback(ctx)
	queries := s.q.WithTx(tx)
	count, err := queries.CountAdminAccounts(ctx)
	if err != nil {
		return daoError(err)
	}
	if count > 0 {
		return common.NewError(common.RETURN_FAILED, http.StatusConflict, "admin password has already been initialized")
	}
	hash, hashErr := s.createPasswordHash(password)
	if hashErr != nil {
		return hashErr
	}
	now := time.Now()
	account, err := queries.InsertAdminAccount(ctx, adminsqlc.InsertAdminAccountParams{
		PasswordHash: hash, SessionVersion: 1, PasswordUpdatedAt: timestamp(now),
	})
	if err != nil {
		return daoError(err)
	}
	if auditErr := s.audit.LogTx(ctx, tx, meta, "bootstrap", "gfa_admin_account", account.ID, nil, map[string]any{
		"id": account.ID, "session_version": account.SessionVersion, "password_updated_at": now,
	}); auditErr != nil {
		return auditErr
	}
	if err := tx.Commit(ctx); err != nil {
		return daoError(err)
	}
	return nil
}

func (s *AuthService) Login(password string) (string, *models.AdminClaims, common.Error) {
	password = strings.TrimSpace(password)
	if password == "" {
		return "", nil, common.NewValidationError("password must not be empty")
	}
	account, err := s.getAccount(context.Background())
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil, common.NewError(common.RETURN_FAILED, http.StatusBadRequest, "admin password has not been initialized")
	}
	if err != nil {
		return "", nil, daoError(err)
	}
	ok, verifyErr := easyhash.VerifyPBKDF2(password, account.PasswordHash)
	if verifyErr != nil {
		return "", nil, common.NewServiceError("password verification failed")
	}
	if !ok {
		return "", nil, common.NewError(common.RETURN_FAILED, http.StatusUnauthorized, "invalid password")
	}
	claims, token, tokenErr := s.buildToken(account.SessionVersion)
	return token, claims, tokenErr
}

func (s *AuthService) ResetPassword(password string, meta audit.Meta) common.Error {
	password = strings.TrimSpace(password)
	if password == "" {
		return common.NewValidationError("password must not be empty")
	}
	hash, hashErr := s.createPasswordHash(password)
	if hashErr != nil {
		return hashErr
	}
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return daoError(err)
	}
	defer tx.Rollback(ctx)
	queries := s.q.WithTx(tx)
	now := time.Now()
	account, err := queries.GetAdminAccount(ctx)
	var before any
	if err == nil {
		before = map[string]any{"id": account.ID, "session_version": account.SessionVersion, "password_updated_at": nullableTime(account.PasswordUpdatedAt)}
		account, err = queries.UpdateAdminAccountPassword(ctx, adminsqlc.UpdateAdminAccountPasswordParams{
			PasswordHash: hash, SessionVersion: account.SessionVersion + 1, PasswordUpdatedAt: timestamp(now),
		})
	} else if errors.Is(err, pgx.ErrNoRows) {
		account, err = queries.InsertAdminAccount(ctx, adminsqlc.InsertAdminAccountParams{
			PasswordHash: hash, SessionVersion: 1, PasswordUpdatedAt: timestamp(now),
		})
	}
	if err != nil {
		return daoError(err)
	}
	if auditErr := s.audit.LogTx(ctx, tx, meta, "reset_password", "gfa_admin_account", account.ID, before, map[string]any{
		"id": account.ID, "session_version": account.SessionVersion, "password_updated_at": now,
	}); auditErr != nil {
		return auditErr
	}
	if err := tx.Commit(ctx); err != nil {
		return daoError(err)
	}
	return nil
}

func (s *AuthService) ParseAndValidateToken(tokenValue string) (*models.AdminClaims, common.Error) {
	tokenValue = strings.TrimSpace(tokenValue)
	if tokenValue == "" {
		return nil, common.NewError(common.RETURN_FAILED, http.StatusUnauthorized, "not logged in")
	}
	cfg := env.GetServerConfig().Auth
	token, err := jwt.ParseWithClaims(tokenValue, &models.AdminClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method %s", token.Method.Alg())
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, common.NewError(common.RETURN_FAILED, http.StatusUnauthorized, "login state is invalid")
	}
	claims, ok := token.Claims.(*models.AdminClaims)
	if !ok || !token.Valid {
		return nil, common.NewError(common.RETURN_FAILED, http.StatusUnauthorized, "login state is invalid")
	}
	account, accountErr := s.getAccount(context.Background())
	if errors.Is(accountErr, pgx.ErrNoRows) {
		return nil, common.NewError(common.RETURN_FAILED, http.StatusUnauthorized, "admin password has not been initialized")
	}
	if accountErr != nil {
		return nil, daoError(accountErr)
	}
	if claims.SessionVersion != account.SessionVersion {
		return nil, common.NewError(common.RETURN_FAILED, http.StatusUnauthorized, "login state has expired")
	}
	return claims, nil
}

func (s *AuthService) BuildAuthCookie(token string) *fiber.Cookie {
	cfg := env.GetServerConfig().Auth
	return &fiber.Cookie{Name: cfg.CookieName, Value: token, Path: "/", Domain: cfg.CookieDomain, SameSite: cfg.SameSite, MaxAge: cfg.CookieMaxAgeSecs, Secure: cfg.CookieSecure, HTTPOnly: true}
}

func (s *AuthService) BuildLogoutCookie() *fiber.Cookie {
	cfg := env.GetServerConfig().Auth
	return &fiber.Cookie{Name: cfg.CookieName, Value: "", Path: "/", Domain: cfg.CookieDomain, SameSite: cfg.SameSite, MaxAge: -1, Secure: cfg.CookieSecure, HTTPOnly: true, Expires: time.Unix(0, 0)}
}

func (s *AuthService) buildToken(sessionVersion int64) (*models.AdminClaims, string, common.Error) {
	cfg := env.GetServerConfig().Auth
	now := time.Now()
	claims := &models.AdminClaims{SessionVersion: sessionVersion, RegisteredClaims: jwt.RegisteredClaims{
		Issuer: env.GetServerConfig().Server.AppID, Subject: "admin", IssuedAt: jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(cfg.SessionTTLHours) * time.Hour)),
	}}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return nil, "", common.NewServiceError("failed to generate login token")
	}
	return claims, signed, nil
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

func (s *AuthService) getAccount(ctx context.Context) (*models.AdminAccount, error) {
	row, err := s.q.GetAdminAccount(ctx)
	if err != nil {
		return nil, err
	}
	return &models.AdminAccount{ID: row.ID, PasswordHash: row.PasswordHash, SessionVersion: row.SessionVersion,
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time, PasswordUpdatedAt: nullableTime(row.PasswordUpdatedAt)}, nil
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

func daoError(err error) common.Error {
	if appErr, ok := err.(common.Error); ok {
		return appErr
	}
	return common.NewDaoError(err.Error())
}
