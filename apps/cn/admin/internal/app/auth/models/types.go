package models

import "github.com/golang-jwt/jwt/v5"

type PasswordRequest struct {
	Password string `json:"password"`
}

type AuthStateResponse struct {
	Initialized   bool `json:"initialized"`
	Authenticated bool `json:"authenticated"`
}

type MeResponse struct {
	Initialized    bool  `json:"initialized"`
	Authenticated  bool  `json:"authenticated"`
	SessionVersion int64 `json:"session_version"`
}

type AdminClaims struct {
	SessionVersion int64 `json:"session_version"`
	jwt.RegisteredClaims
}
