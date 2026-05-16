package middleware

import (
	"strings"

	env "github.com/gofurry/awesome-fiber-template/v3/medium/config"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/auth/service"
	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/common"
	"github.com/gofiber/fiber/v3"
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
