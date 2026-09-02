package models

import (
	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
	"github.com/golang-jwt/jwt/v5"
)

type BootstrapRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AccountCreateRequest struct {
	Username    string             `json:"username"`
	DisplayName string             `json:"display_name"`
	Role        authorization.Role `json:"role"`
	Password    string             `json:"password"`
}

type DisplayNameRequest struct {
	DisplayName string `json:"display_name"`
}

type RoleRequest struct {
	Role authorization.Role `json:"role"`
}

type StatusRequest struct {
	Status authorization.AccountStatus `json:"status"`
}

type PasswordRequest struct {
	Password string `json:"password"`
}

type IdentityResponse struct {
	AccountID      int64                       `json:"account_id"`
	Username       string                      `json:"username"`
	DisplayName    string                      `json:"display_name"`
	Role           authorization.Role          `json:"role"`
	Status         authorization.AccountStatus `json:"status"`
	SessionVersion int64                       `json:"session_version"`
	Capabilities   []authorization.Capability  `json:"capabilities"`
}

func IdentityDTO(principal *authorization.Principal) *IdentityResponse {
	if principal == nil {
		return nil
	}
	return &IdentityResponse{
		AccountID: principal.AccountID, Username: principal.Username, DisplayName: principal.DisplayName,
		Role: principal.Role, Status: principal.Status, SessionVersion: principal.SessionVersion,
		Capabilities: append([]authorization.Capability(nil), principal.Capabilities...),
	}
}

type AuthStateResponse struct {
	Initialized   bool              `json:"initialized"`
	Authenticated bool              `json:"authenticated"`
	Identity      *IdentityResponse `json:"identity,omitempty"`
}

type MeResponse = AuthStateResponse

type AdminClaims struct {
	AccountID      int64 `json:"account_id"`
	SessionVersion int64 `json:"session_version"`
	jwt.RegisteredClaims
}
