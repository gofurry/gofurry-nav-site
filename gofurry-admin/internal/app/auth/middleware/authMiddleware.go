package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/app/auth/service"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

func Required() fiber.Handler {
	return func(c fiber.Ctx) error {
		token := strings.TrimSpace(c.Cookies(env.GetServerConfig().Auth.CookieName))
		claims, err := service.GetAuthService().ParseAndValidateToken(token)
		if err != nil {
			return common.NewResponse(c).ErrorWithCode(err, err.GetHTTPStatus())
		}
		c.Locals(service.ClaimsContextKey, claims)
		return c.Next()
	}
}
