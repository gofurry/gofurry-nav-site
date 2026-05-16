package service

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/awesome-fiber-template/v3/medium/config"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/auth/models"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/shared/audit"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/infra/db"
	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/common"
	"github.com/gofurry/easyhash"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const ClaimsContextKey = "auth_claims"

type authService struct{}

var authSingleton = new(authService)

func GetAuthService() *authService {
	return authSingleton
}

func (s *authService) IsInitialized() (bool, common.Error) {
	_, err := s.getAccount()
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return false, nil
	default:
		return false, common.NewDaoError(err.Error())
	}
}

func (s *authService) Bootstrap(password string, meta audit.Meta) common.Error {
	password = strings.TrimSpace(password)
	if password == "" {
		return common.NewValidationError("password must not be empty")
	}

	engine, err := adminDB()
	if err != nil {
		return err
	}

	if err := engine.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.AdminAccount{}).Count(&count).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		if count > 0 {
			return common.NewError(common.RETURN_FAILED, http.StatusConflict, "admin password has already been initialized")
		}

		hash, hashErr := s.createPasswordHash(password)
		if hashErr != nil {
			return hashErr
		}

		now := time.Now()
		account := &models.AdminAccount{
			ID:                1,
			PasswordHash:      hash,
			SessionVersion:    1,
			PasswordUpdatedAt: &now,
		}
		if err := tx.Create(account).Error; err != nil {
			return common.NewDaoError(err.Error())
		}
		if err := audit.LogTx(tx, meta, "bootstrap", account.TableName(), account.ID, nil, map[string]any{
			"id":                  account.ID,
			"session_version":     account.SessionVersion,
			"password_updated_at": account.PasswordUpdatedAt,
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return asCommonError(err)
	}
	return nil
}

func (s *authService) Login(password string) (string, *models.AdminClaims, common.Error) {
	password = strings.TrimSpace(password)
	if password == "" {
		return "", nil, common.NewValidationError("password must not be empty")
	}

	account, err := s.getAccount()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, common.NewError(common.RETURN_FAILED, http.StatusBadRequest, "admin password has not been initialized")
		}
		return "", nil, common.NewDaoError(err.Error())
	}

	ok, verifyErr := easyhash.VerifyPBKDF2(password, account.PasswordHash)
	if verifyErr != nil {
		return "", nil, common.NewServiceError("password verification failed")
	}
	if !ok {
		return "", nil, common.NewError(common.RETURN_FAILED, http.StatusUnauthorized, "invalid password")
	}

	claims, token, tokenErr := s.buildToken(account.SessionVersion)
	if tokenErr != nil {
		return "", nil, tokenErr
	}
	return token, claims, nil
}

func (s *authService) ResetPassword(password string, meta audit.Meta) common.Error {
	password = strings.TrimSpace(password)
	if password == "" {
		return common.NewValidationError("password must not be empty")
	}

	engine, err := adminDB()
	if err != nil {
		return err
	}

	if err := engine.Transaction(func(tx *gorm.DB) error {
		hash, hashErr := s.createPasswordHash(password)
		if hashErr != nil {
			return hashErr
		}

		now := time.Now()
		var account models.AdminAccount
		var before map[string]any
		err := tx.First(&account, "id = ?", 1).Error
		switch {
		case err == nil:
			before = map[string]any{
				"id":                  account.ID,
				"session_version":     account.SessionVersion,
				"password_updated_at": account.PasswordUpdatedAt,
			}
			account.PasswordHash = hash
			account.SessionVersion++
			account.PasswordUpdatedAt = &now
			if err := tx.Save(&account).Error; err != nil {
				return common.NewDaoError(err.Error())
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			account = models.AdminAccount{
				ID:                1,
				PasswordHash:      hash,
				SessionVersion:    1,
				PasswordUpdatedAt: &now,
			}
			if err := tx.Create(&account).Error; err != nil {
				return common.NewDaoError(err.Error())
			}
		default:
			return common.NewDaoError(err.Error())
		}

		if err := audit.LogTx(tx, meta, "reset_password", account.TableName(), account.ID, before, map[string]any{
			"id":                  account.ID,
			"session_version":     account.SessionVersion,
			"password_updated_at": account.PasswordUpdatedAt,
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return asCommonError(err)
	}
	return nil
}

func (s *authService) ParseAndValidateToken(tokenValue string) (*models.AdminClaims, common.Error) {
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

	account, accountErr := s.getAccount()
	if accountErr != nil {
		if errors.Is(accountErr, gorm.ErrRecordNotFound) {
			return nil, common.NewError(common.RETURN_FAILED, http.StatusUnauthorized, "admin password has not been initialized")
		}
		return nil, common.NewDaoError(accountErr.Error())
	}
	if claims.SessionVersion != account.SessionVersion {
		return nil, common.NewError(common.RETURN_FAILED, http.StatusUnauthorized, "login state has expired")
	}

	return claims, nil
}

func (s *authService) BuildAuthCookie(token string) *fiber.Cookie {
	cfg := env.GetServerConfig().Auth
	return &fiber.Cookie{
		Name:     cfg.CookieName,
		Value:    token,
		Path:     "/",
		Domain:   cfg.CookieDomain,
		SameSite: cfg.SameSite,
		MaxAge:   cfg.CookieMaxAgeSecs,
		Secure:   cfg.CookieSecure,
		HTTPOnly: true,
	}
}

func (s *authService) BuildLogoutCookie() *fiber.Cookie {
	cfg := env.GetServerConfig().Auth
	return &fiber.Cookie{
		Name:     cfg.CookieName,
		Value:    "",
		Path:     "/",
		Domain:   cfg.CookieDomain,
		SameSite: cfg.SameSite,
		MaxAge:   -1,
		Secure:   cfg.CookieSecure,
		HTTPOnly: true,
		Expires:  time.Unix(0, 0),
	}
}

func (s *authService) buildToken(sessionVersion int64) (*models.AdminClaims, string, common.Error) {
	cfg := env.GetServerConfig().Auth
	now := time.Now()
	claims := &models.AdminClaims{
		SessionVersion: sessionVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    env.GetServerConfig().Server.AppID,
			Subject:   "admin",
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(cfg.SessionTTLHours) * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return nil, "", common.NewServiceError("failed to generate login token")
	}
	return claims, signed, nil
}

func (s *authService) createPasswordHash(password string) (string, common.Error) {
	cfg := easyhash.DefaultPBKDF2()
	cfg.PBKDF2Iterations = env.GetServerConfig().Auth.PBKDF2Iterations
	hash, err := easyhash.CreatePBKDF2(cfg, password)
	if err != nil {
		return "", common.NewServiceError("failed to generate password hash")
	}
	return hash, nil
}

func (s *authService) getAccount() (*models.AdminAccount, error) {
	engine, err := adminDB()
	if err != nil {
		return nil, err
	}

	var account models.AdminAccount
	if err := engine.First(&account, "id = ?", 1).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func adminDB() (*gorm.DB, common.Error) {
	engine := db.Databases.DB(db.Admin)
	if engine == nil {
		return nil, common.NewDaoError("admin database is not initialized")
	}
	return engine, nil
}

func asCommonError(err error) common.Error {
	if err == nil {
		return nil
	}
	if appErr, ok := err.(common.Error); ok {
		return appErr
	}
	return common.NewDaoError(err.Error())
}
