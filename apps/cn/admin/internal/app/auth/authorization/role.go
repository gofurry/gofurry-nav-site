package authorization

import "strings"

type Role string

const (
	RoleOwner     Role = "owner"
	RoleDeveloper Role = "developer"
	RoleOperator  Role = "operator"
)

func ParseRole(value string) (Role, bool) {
	role := Role(strings.ToLower(strings.TrimSpace(value)))
	switch role {
	case RoleOwner, RoleDeveloper, RoleOperator:
		return role, true
	default:
		return "", false
	}
}

type AccountStatus string

const (
	StatusActive   AccountStatus = "active"
	StatusDisabled AccountStatus = "disabled"
)

func ParseStatus(value string) (AccountStatus, bool) {
	status := AccountStatus(strings.ToLower(strings.TrimSpace(value)))
	switch status {
	case StatusActive, StatusDisabled:
		return status, true
	default:
		return "", false
	}
}
