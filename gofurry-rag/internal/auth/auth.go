package auth

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofurry/gofurry-rag/config"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

const ClaimsContextKey = "auth_claims"

var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrNotLoggedIn     = errors.New("not logged in")
	ErrInvalidSession  = errors.New("login state is invalid")
)

type Claims struct {
	SessionVersion int64 `json:"session_version"`
	jwt.RegisteredClaims
}

type Service struct {
	cfg config.Config
}

func New(cfg config.Config) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) Login(password string) (string, *Claims, error) {
	password = strings.TrimSpace(password)
	passcode := strings.TrimSpace(s.cfg.ConsolePasscode)
	if password == "" || passcode == "" {
		return "", nil, ErrInvalidPassword
	}
	if subtle.ConstantTimeCompare([]byte(password), []byte(passcode)) != 1 {
		return "", nil, ErrInvalidPassword
	}
	return s.buildToken()
}

func (s *Service) ParseAndValidateToken(tokenValue string) (*Claims, error) {
	tokenValue = strings.TrimSpace(tokenValue)
	if tokenValue == "" {
		return nil, ErrNotLoggedIn
	}
	token, err := jwt.ParseWithClaims(tokenValue, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method %s", token.Method.Alg())
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, ErrInvalidSession
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidSession
	}
	if claims.SessionVersion != 1 {
		return nil, ErrInvalidSession
	}
	return claims, nil
}

func (s *Service) BuildAuthCookie(token string) *fiber.Cookie {
	maxAge := s.cfg.Auth.CookieMaxAgeSecs
	if maxAge <= 0 {
		maxAge = s.cfg.SessionTTLHours * 3600
	}
	return &fiber.Cookie{
		Name:     s.cfg.AuthCookieName,
		Value:    token,
		Path:     "/",
		Domain:   s.cfg.Auth.CookieDomain,
		SameSite: s.cfg.Auth.SameSite,
		Secure:   s.cfg.Auth.CookieSecure,
		MaxAge:   maxAge,
		HTTPOnly: true,
	}
}

func (s *Service) BuildLogoutCookie() *fiber.Cookie {
	return &fiber.Cookie{
		Name:     s.cfg.AuthCookieName,
		Value:    "",
		Path:     "/",
		Domain:   s.cfg.Auth.CookieDomain,
		SameSite: s.cfg.Auth.SameSite,
		Secure:   s.cfg.Auth.CookieSecure,
		MaxAge:   -1,
		HTTPOnly: true,
		Expires:  time.Unix(0, 0),
	}
}

func (s *Service) buildToken() (string, *Claims, error) {
	now := time.Now()
	claims := &Claims{
		SessionVersion: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.AppName,
			Subject:   "admin",
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.cfg.SessionTTLHours) * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", nil, err
	}
	return signed, claims, nil
}
