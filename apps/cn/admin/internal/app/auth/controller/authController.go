package controller

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/app/auth/models"
	"github.com/gofurry/gofurry-admin/internal/app/auth/service"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

type AuthAPI struct {
	service *service.AuthService
	audit   *audit.Logger
}

func New(authService *service.AuthService, auditLogger *audit.Logger) *AuthAPI {
	return &AuthAPI{service: authService, audit: auditLogger}
}

func (api *AuthAPI) State(c fiber.Ctx) error {
	initialized, err := api.service.IsInitialized()
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	authenticated := false
	token := strings.TrimSpace(c.Cookies(env.GetServerConfig().Auth.CookieName))
	if token != "" {
		if claims, parseErr := api.service.ParseAndValidateToken(token); parseErr == nil {
			authenticated = true
			c.Locals(service.ClaimsContextKey, claims)
		}
	}

	return common.NewResponse(c).SuccessWithData(models.AuthStateResponse{
		Initialized:   initialized,
		Authenticated: authenticated,
	})
}

func (api *AuthAPI) Bootstrap(c fiber.Ctx) error {
	var req models.PasswordRequest
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}

	if serviceErr := api.service.Bootstrap(req.Password, audit.MetaFromFiber(c)); serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}

	return common.NewResponse(c).Success()
}

func (api *AuthAPI) Login(c fiber.Ctx) error {
	var req models.PasswordRequest
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}

	token, claims, serviceErr := api.service.Login(req.Password)
	if serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}
	if auditErr := api.audit.Log(c.Context(), audit.MetaFromFiber(c), "login", "gfa_admin_account", 1, nil, map[string]any{
		"session_version": claims.SessionVersion,
	}); auditErr != nil {
		return common.NewResponse(c).Error(auditErr)
	}

	c.Cookie(api.service.BuildAuthCookie(token))
	return common.NewResponse(c).SuccessWithData(models.MeResponse{
		Initialized:    true,
		Authenticated:  true,
		SessionVersion: claims.SessionVersion,
	})
}

func (api *AuthAPI) Logout(c fiber.Ctx) error {
	claims, _ := currentClaims(c)
	meta := audit.MetaFromFiber(c)
	if claims != nil {
		meta.SessionVersion = claims.SessionVersion
	}
	if auditErr := api.audit.Log(c.Context(), meta, "logout", "gfa_admin_account", 1, nil, map[string]any{
		"session_version": meta.SessionVersion,
	}); auditErr != nil {
		return common.NewResponse(c).Error(auditErr)
	}
	c.Cookie(api.service.BuildLogoutCookie())
	return common.NewResponse(c).Success()
}

func (api *AuthAPI) Me(c fiber.Ctx) error {
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
