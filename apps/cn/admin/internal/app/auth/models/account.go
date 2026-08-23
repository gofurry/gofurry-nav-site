package models

import "time"

type AdminAccount struct {
	ID                int64      `json:"id"`
	PasswordHash      string     `json:"password_hash"`
	SessionVersion    int64      `json:"session_version"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	PasswordUpdatedAt *time.Time `json:"password_updated_at"`
}

func (AdminAccount) TableName() string {
	return "gfa_admin_account"
}
