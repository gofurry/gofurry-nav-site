package middleware

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
	"github.com/gofurry/gofurry-admin/internal/app/auth/service"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

func Required(authService *service.AuthService) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := strings.TrimSpace(c.Cookies(env.GetServerConfig().Auth.CookieName))
		_, principal, err := authService.ParseAndValidateToken(c.Context(), token)
		if err != nil {
			return common.NewResponse(c).ErrorWithCode(err, err.GetHTTPStatus())
		}
		c.Locals(authorization.PrincipalContextKey, principal)
		return c.Next()
	}
}

func Require(capability authorization.Capability) fiber.Handler {
	return func(c fiber.Ctx) error {
		principal, err := CurrentPrincipal(c)
		if err != nil {
			return common.NewResponse(c).ErrorWithCode(err, err.GetHTTPStatus())
		}
		if !principal.Has(capability) {
			denied := common.NewError(common.RETURN_FAILED, http.StatusForbidden, "access is forbidden")
			return common.NewResponse(c).ErrorWithCode(denied, http.StatusForbidden)
		}
		return c.Next()
	}
}

func CurrentPrincipal(c fiber.Ctx) (*authorization.Principal, common.Error) {
	raw := c.Locals(authorization.PrincipalContextKey)
	principal, ok := raw.(*authorization.Principal)
	if !ok || principal == nil {
		return nil, common.NewError(common.RETURN_FAILED, http.StatusUnauthorized, "not logged in")
	}
	return principal, nil
}
