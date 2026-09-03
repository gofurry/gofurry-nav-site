package service

import (
	"net/http"
	"strings"
	"testing"

	"github.com/GoFurry/easyhash"
	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/app/auth/models"
)

func TestCreatePasswordHashUsesConfiguredIterations(t *testing.T) {
	t.Parallel()

	hash, err := (&AuthService{}).createPasswordHash("abc123")
	if err != nil {
		t.Fatalf("createPasswordHash returned error: %v", err)
	}

	parts := strings.Split(hash, ":")
	if len(parts) != 3 {
		t.Fatalf("unexpected hash format: %q", hash)
	}
	if parts[1] != "300000" {
		t.Fatalf("expected 300000 iterations, got %s", parts[1])
	}

	ok, verifyErr := easyhash.VerifyPBKDF2("abc123", hash)
	if verifyErr != nil {
		t.Fatalf("verify returned error: %v", verifyErr)
	}
	if !ok {
		t.Fatalf("expected password verification to succeed")
	}
	if env.GetServerConfig().Auth.PBKDF2Iterations != 300000 {
		t.Fatalf("unexpected config iteration count: %d", env.GetServerConfig().Auth.PBKDF2Iterations)
	}
}

func TestVerifyCurrentPassword(t *testing.T) {
	t.Parallel()

	service := &AuthService{}
	hash, err := service.createPasswordHash("current-password")
	if err != nil {
		t.Fatal(err)
	}
	account := &models.AdminAccount{PasswordHash: hash}
	if verifyErr := verifyCurrentPassword(account, "current-password"); verifyErr != nil {
		t.Fatalf("correct current password rejected: %v", verifyErr)
	}
	if verifyErr := verifyCurrentPassword(account, "wrong-password"); verifyErr == nil || verifyErr.GetHTTPStatus() != http.StatusUnauthorized {
		t.Fatalf("wrong current password result=%v, want 401", verifyErr)
	}
}
