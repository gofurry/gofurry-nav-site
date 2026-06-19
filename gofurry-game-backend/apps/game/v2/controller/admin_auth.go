package controller

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-game-backend/common"
	env "github.com/gofurry/gofurry-game-backend/roof/env"
)

const defaultAdminTokenHeader = "X-GoFurry-Admin-Token"

func RequireAdminToken() fiber.Handler {
	return func(c fiber.Ctx) error {
		cfg := env.GetServerConfig().Admin
		expected := strings.TrimSpace(cfg.Token)
		if expected == "" {
			return common.NewResponse(c).ErrorWithCode("admin token is not configured", http.StatusServiceUnavailable)
		}

		header := strings.TrimSpace(cfg.Header)
		if header == "" {
			header = defaultAdminTokenHeader
		}

		provided := strings.TrimSpace(c.Get(header))
		if provided == "" {
			auth := strings.TrimSpace(c.Get(fiber.HeaderAuthorization))
			if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
				provided = strings.TrimSpace(auth[7:])
			}
		}
		if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			return common.NewResponse(c).ErrorWithCode("invalid admin token", http.StatusUnauthorized)
		}
		return c.Next()
	}
}
