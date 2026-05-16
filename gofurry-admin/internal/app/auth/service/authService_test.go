package service

import (
	"strings"
	"testing"

	"github.com/GoFurry/easyhash"
	env "github.com/gofurry/awesome-fiber-template/v3/medium/config"
)

func TestCreatePasswordHashUsesConfiguredIterations(t *testing.T) {
	t.Parallel()

	hash, err := GetAuthService().createPasswordHash("abc123")
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
