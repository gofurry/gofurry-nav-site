package models

import (
	"time"

	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
)

type AdminAccount struct {
	ID                int64
	Username          string
	DisplayName       string
	Role              authorization.Role
	Status            authorization.AccountStatus
	PasswordHash      string
	SessionVersion    int64
	LastLoginAt       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	PasswordUpdatedAt *time.Time
}

type AccountResponse struct {
	ID                int64                       `json:"id"`
	Username          string                      `json:"username"`
	DisplayName       string                      `json:"display_name"`
	Role              authorization.Role          `json:"role"`
	Status            authorization.AccountStatus `json:"status"`
	SessionVersion    int64                       `json:"session_version"`
	LastLoginAt       *time.Time                  `json:"last_login_at"`
	CreatedAt         time.Time                   `json:"created_at"`
	UpdatedAt         time.Time                   `json:"updated_at"`
	PasswordUpdatedAt *time.Time                  `json:"password_updated_at"`
}

func AccountDTO(account AdminAccount) AccountResponse {
	return AccountResponse{
		ID: account.ID, Username: account.Username, DisplayName: account.DisplayName,
		Role: account.Role, Status: account.Status, SessionVersion: account.SessionVersion,
		LastLoginAt: account.LastLoginAt, CreatedAt: account.CreatedAt, UpdatedAt: account.UpdatedAt,
		PasswordUpdatedAt: account.PasswordUpdatedAt,
	}
}

type AccountPage struct {
	Total int64             `json:"total"`
	List  []AccountResponse `json:"list"`
}
