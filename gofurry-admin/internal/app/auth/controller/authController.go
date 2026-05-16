package controller

import (
	"strings"

	env "github.com/gofurry/awesome-fiber-template/v3/medium/config"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/auth/models"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/auth/service"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/shared/adminutil"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/shared/audit"
	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/common"
	"github.com/gofiber/fiber/v3"
)

type authAPI struct{}

var AuthAPI = &authAPI{}

func (api *authAPI) State(c fiber.Ctx) error {
	authService := service.GetAuthService()
	initialized, err := authService.IsInitialized()
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	authenticated := false
	token := strings.TrimSpace(c.Cookies(env.GetServerConfig().Auth.CookieName))
	if token != "" {
		if claims, parseErr := authService.ParseAndValidateToken(token); parseErr == nil {
			authenticated = true
			c.Locals(service.ClaimsContextKey, claims)
		}
	}

	return common.NewResponse(c).SuccessWithData(models.AuthStateResponse{
		Initialized:   initialized,
		Authenticated: authenticated,
	})
}

func (api *authAPI) Bootstrap(c fiber.Ctx) error {
	var req models.PasswordRequest
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}

	if serviceErr := service.GetAuthService().Bootstrap(req.Password, audit.MetaFromFiber(c)); serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}

	return common.NewResponse(c).Success()
}

func (api *authAPI) Login(c fiber.Ctx) error {
	var req models.PasswordRequest
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}

	token, claims, serviceErr := service.GetAuthService().Login(req.Password)
	if serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}
	if auditErr := audit.Log(audit.MetaFromFiber(c), "login", "gfa_admin_account", 1, nil, map[string]any{
		"session_version": claims.SessionVersion,
	}); auditErr != nil {
		return common.NewResponse(c).Error(auditErr)
	}

	c.Cookie(service.GetAuthService().BuildAuthCookie(token))
	return common.NewResponse(c).SuccessWithData(models.MeResponse{
		Initialized:    true,
		Authenticated:  true,
		SessionVersion: claims.SessionVersion,
	})
}

func (api *authAPI) Logout(c fiber.Ctx) error {
	claims, _ := currentClaims(c)
	meta := audit.MetaFromFiber(c)
	if claims != nil {
		meta.SessionVersion = claims.SessionVersion
	}
	if auditErr := audit.Log(meta, "logout", "gfa_admin_account", 1, nil, map[string]any{
		"session_version": meta.SessionVersion,
	}); auditErr != nil {
		return common.NewResponse(c).Error(auditErr)
	}
	c.Cookie(service.GetAuthService().BuildLogoutCookie())
	return common.NewResponse(c).Success()
}

func (api *authAPI) Me(c fiber.Ctx) error {
	claims, err := currentClaims(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	return common.NewResponse(c).SuccessWithData(models.MeResponse{
		Initialized:    true,
		Authenticated:  true,
		SessionVersion: claims.SessionVersion,
	})
}

func currentClaims(c fiber.Ctx) (*models.AdminClaims, common.Error) {
	raw := c.Locals(service.ClaimsContextKey)
	claims, ok := raw.(*models.AdminClaims)
	if !ok || claims == nil {
		return nil, common.NewError(common.RETURN_FAILED, fiber.StatusUnauthorized, "not logged in")
	}
	return claims, nil
}
