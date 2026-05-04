package auth

import (
	"testing"

	"github.com/GoFurry/gofurry-rag/internal/config"
)

func TestLoginAndValidateToken(t *testing.T) {
	svc := New(config.Config{
		AppName:         "test",
		ConsolePasscode: "secret",
		JWTSecret:       "jwt-secret",
		AuthCookieName:  "session",
		SessionTTLHours: 1,
	})
	token, claims, err := svc.Login("secret")
	if err != nil {
		t.Fatal(err)
	}
	if token == "" || claims.SessionVersion != 1 {
		t.Fatalf("unexpected token or claims: %q %#v", token, claims)
	}
	parsed, err := svc.ParseAndValidateToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Subject != "admin" {
		t.Fatalf("subject = %q", parsed.Subject)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	svc := New(config.Config{ConsolePasscode: "secret", JWTSecret: "jwt-secret", SessionTTLHours: 1})
	if _, _, err := svc.Login("nope"); err != ErrInvalidPassword {
		t.Fatalf("err = %v", err)
	}
}
