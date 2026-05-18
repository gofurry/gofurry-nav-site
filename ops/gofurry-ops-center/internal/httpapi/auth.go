package httpapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/security"
)

type loginRequest struct {
	Passcode string `json:"passcode"`
	Password string `json:"password"`
}

func (s *Server) authState(c fiber.Ctx) error {
	return ok(c, fiber.Map{"authenticated": s.validSession(c), "initialized": true})
}

func (s *Server) authLogin(c fiber.Ctx) error {
	var req loginRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid json")
	}
	passcode := strings.TrimSpace(req.Passcode)
	if passcode == "" {
		passcode = strings.TrimSpace(req.Password)
	}
	if passcode == "" || !constantTimeEqual(passcode, strings.TrimSpace(s.cfg.Security.DashboardPasscode)) {
		return fail(c, fiber.StatusUnauthorized, "invalid passcode")
	}
	token, err := security.NewSession(s.cfg.Security.SessionSecret, s.cfg.Security.SessionTTL.Duration)
	if err != nil {
		return fail(c, fiber.StatusInternalServerError, "create session failed")
	}
	c.Cookie(&fiber.Cookie{
		Name:     s.cfg.Security.CookieName,
		Value:    token,
		HTTPOnly: true,
		Secure:   s.cfg.Security.CookieSecure,
		SameSite: "Lax",
		Expires:  time.Now().Add(s.cfg.Security.SessionTTL.Duration),
		Path:     "/",
	})
	return ok(c, fiber.Map{"authenticated": true})
}

func (s *Server) authLogout(c fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     s.cfg.Security.CookieName,
		Value:    "",
		HTTPOnly: true,
		Secure:   s.cfg.Security.CookieSecure,
		SameSite: "Lax",
		Expires:  time.Unix(0, 0),
		Path:     "/",
	})
	return ok(c, fiber.Map{"authenticated": false})
}

func (s *Server) authMe(c fiber.Ctx) error {
	return ok(c, fiber.Map{"authenticated": true})
}

func (s *Server) requireAdmin(c fiber.Ctx) error {
	if !s.validSession(c) {
		return fail(c, fiber.StatusUnauthorized, "not authenticated")
	}
	return c.Next()
}

func (s *Server) requirePeer(c fiber.Ctx) error {
	if s.cfg.Security.PeerToken == "" {
		return fail(c, fiber.StatusUnauthorized, "peer token disabled")
	}
	if bearerToken(c.Get(fiber.HeaderAuthorization)) != s.cfg.Security.PeerToken {
		return fail(c, fiber.StatusUnauthorized, "invalid peer token")
	}
	return c.Next()
}

func (s *Server) requireEvent(c fiber.Ctx) error {
	token := bearerToken(c.Get(fiber.HeaderAuthorization))
	if s.cfg.Security.EventToken != "" && token == s.cfg.Security.EventToken {
		return c.Next()
	}
	if s.validSession(c) {
		return c.Next()
	}
	return fail(c, fiber.StatusUnauthorized, "invalid event token")
}

func (s *Server) validSession(c fiber.Ctx) bool {
	token := strings.TrimSpace(c.Cookies(s.cfg.Security.CookieName))
	return token != "" && security.VerifySession(s.cfg.Security.SessionSecret, token)
}

func constantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		subtle.ConstantTimeCompare([]byte(a), []byte(b))
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func bearerToken(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(strings.ToLower(value), "bearer ") {
		return strings.TrimSpace(value[len("bearer "):])
	}
	return ""
}

func queryInt(c fiber.Ctx, key string, fallback int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func requestContext(c fiber.Ctx) context.Context {
	if c.Context() != nil {
		return c.Context()
	}
	return context.Background()
}
