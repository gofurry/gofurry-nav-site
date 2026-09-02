package controller

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/app/auth/middleware"
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
	initialized, err := api.service.IsInitialized(c.Context())
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var identity *models.IdentityResponse
	token := strings.TrimSpace(c.Cookies(env.GetServerConfig().Auth.CookieName))
	if token != "" {
		if _, principal, parseErr := api.service.ParseAndValidateToken(c.Context(), token); parseErr == nil {
			identity = models.IdentityDTO(principal)
		}
	}
	return common.NewResponse(c).SuccessWithData(models.AuthStateResponse{
		Initialized: initialized, Authenticated: identity != nil, Identity: identity,
	})
}

func (api *AuthAPI) Bootstrap(c fiber.Ctx) error {
	var request models.BootstrapRequest
	if err := adminutil.DecodeBody(c, &request); err != nil {
		return common.NewResponse(c).Error(err)
	}
	account, serviceErr := api.service.Bootstrap(c.Context(), request, audit.MetaFromFiber(c))
	if serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}
	return common.NewResponse(c).SuccessWithData(account)
}

func (api *AuthAPI) Login(c fiber.Ctx) error {
	var request models.LoginRequest
	if err := adminutil.DecodeBody(c, &request); err != nil {
		return common.NewResponse(c).Error(err)
	}
	token, principal, serviceErr := api.service.Login(c.Context(), request.Username, request.Password)
	if serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}
	meta := audit.MetaForPrincipal(audit.MetaFromFiber(c), principal)
	if auditErr := api.audit.Log(c.Context(), meta, "login", "gfa_admin_account", principal.AccountID, nil, map[string]any{
		"account_id": principal.AccountID, "username": principal.Username, "role": principal.Role,
		"session_version": principal.SessionVersion,
	}); auditErr != nil {
		return common.NewResponse(c).Error(auditErr)
	}
	c.Cookie(api.service.BuildAuthCookie(token))
	return common.NewResponse(c).SuccessWithData(models.MeResponse{
		Initialized: true, Authenticated: true, Identity: models.IdentityDTO(principal),
	})
}

func (api *AuthAPI) Logout(c fiber.Ctx) error {
	principal, _ := middleware.CurrentPrincipal(c)
	if principal == nil {
		token := strings.TrimSpace(c.Cookies(env.GetServerConfig().Auth.CookieName))
		if token != "" {
			_, principal, _ = api.service.ParseAndValidateToken(c.Context(), token)
		}
	}
	if principal != nil {
		meta := audit.MetaForPrincipal(audit.MetaFromFiber(c), principal)
		if auditErr := api.audit.Log(c.Context(), meta, "logout", "gfa_admin_account", principal.AccountID, nil, map[string]any{
			"session_version": principal.SessionVersion,
		}); auditErr != nil {
			return common.NewResponse(c).Error(auditErr)
		}
	}
	c.Cookie(api.service.BuildLogoutCookie())
	return common.NewResponse(c).Success()
}

func (api *AuthAPI) Me(c fiber.Ctx) error {
	principal, err := middleware.CurrentPrincipal(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(models.MeResponse{
		Initialized: true, Authenticated: true, Identity: models.IdentityDTO(principal),
	})
}

func (api *AuthAPI) ListAccounts(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	result, err := api.service.ListAccounts(c.Context(), page.Keyword, int32(page.PageSize), int32((page.PageNum-1)*page.PageSize))
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(result)
}

func (api *AuthAPI) CreateAccount(c fiber.Ctx) error {
	var request models.AccountCreateRequest
	if err := adminutil.DecodeBody(c, &request); err != nil {
		return common.NewResponse(c).Error(err)
	}
	account, serviceErr := api.service.CreateAccount(c.Context(), request, audit.MetaFromFiber(c))
	if serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}
	return common.NewResponse(c).SuccessWithData(account)
}

func (api *AuthAPI) UpdateDisplayName(c fiber.Ctx) error {
	accountID, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var request models.DisplayNameRequest
	if err := adminutil.DecodeBody(c, &request); err != nil {
		return common.NewResponse(c).Error(err)
	}
	account, serviceErr := api.service.UpdateDisplayName(c.Context(), accountID, request.DisplayName, audit.MetaFromFiber(c))
	if serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}
	return common.NewResponse(c).SuccessWithData(account)
}

func (api *AuthAPI) ChangeRole(c fiber.Ctx) error {
	accountID, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var request models.RoleRequest
	if err := adminutil.DecodeBody(c, &request); err != nil {
		return common.NewResponse(c).Error(err)
	}
	account, serviceErr := api.service.ChangeRole(c.Context(), accountID, request.Role, audit.MetaFromFiber(c))
	if serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}
	return common.NewResponse(c).SuccessWithData(account)
}

func (api *AuthAPI) ChangeStatus(c fiber.Ctx) error {
	accountID, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var request models.StatusRequest
	if err := adminutil.DecodeBody(c, &request); err != nil {
		return common.NewResponse(c).Error(err)
	}
	account, serviceErr := api.service.ChangeStatus(c.Context(), accountID, request.Status, audit.MetaFromFiber(c))
	if serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}
	return common.NewResponse(c).SuccessWithData(account)
}

func (api *AuthAPI) ResetAccountPassword(c fiber.Ctx) error {
	accountID, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var request models.PasswordRequest
	if err := adminutil.DecodeBody(c, &request); err != nil {
		return common.NewResponse(c).Error(err)
	}
	account, serviceErr := api.service.ResetAccountPassword(c.Context(), accountID, request.Password, audit.MetaFromFiber(c))
	if serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}
	return common.NewResponse(c).SuccessWithData(account)
}

func (api *AuthAPI) RevokeSessions(c fiber.Ctx) error {
	accountID, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	account, serviceErr := api.service.RevokeSessions(c.Context(), accountID, audit.MetaFromFiber(c))
	if serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}
	return common.NewResponse(c).SuccessWithData(account)
}
