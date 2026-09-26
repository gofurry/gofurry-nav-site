package router

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
	"github.com/gofurry/gofurry-admin/internal/app/collaboration"
	"github.com/gofurry/gofurry-admin/internal/bootstrap"
)

// Exercise the production registration, not a test-only copy of its handlers.
// An invalid body reaches the controller without needing a database; a missing
// route returns 404, while capability checks must still run before decoding.
func TestCollaborationTransitionRoutes(t *testing.T) {
	for _, access := range []struct {
		name      string
		principal *authorization.Principal
		status    int
	}{
		{"owner", &authorization.Principal{Capabilities: authorization.CapabilitiesFor(authorization.RoleOwner)}, fiber.StatusBadRequest},
		{"developer", &authorization.Principal{Capabilities: authorization.CapabilitiesFor(authorization.RoleDeveloper)}, fiber.StatusBadRequest},
		{"operator", &authorization.Principal{Capabilities: authorization.CapabilitiesFor(authorization.RoleOperator)}, fiber.StatusBadRequest},
		{"read-only", &authorization.Principal{Capabilities: []authorization.Capability{authorization.CollaborationRead}}, fiber.StatusForbidden},
		{"content-write-only", &authorization.Principal{Capabilities: []authorization.Capability{authorization.ContentWrite}}, fiber.StatusForbidden},
		{"anonymous", nil, fiber.StatusUnauthorized},
	} {
		t.Run(access.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c fiber.Ctx) error {
				if access.principal != nil {
					c.Locals(authorization.PrincipalContextKey, access.principal)
				}
				return c.Next()
			})
			collaborationRoutes(app.Group("/api/v1/collaboration"), &bootstrap.Runtime{CollaborationAPI: collaboration.NewAPI(nil)})
			for _, action := range []string{"research", "release", "shelve", "restore", "link", "land", "reopen"} {
				t.Run(action, func(t *testing.T) {
					path := "/api/v1/collaboration/ideas/1/" + action
					req := httptest.NewRequest(fiber.MethodPost, path, strings.NewReader(`{"version":`))
					req.Header.Set("Content-Type", "application/json")
					response, err := app.Test(req)
					if err != nil {
						t.Fatal(err)
					}
					defer response.Body.Close()
					body, _ := io.ReadAll(response.Body)
					if response.StatusCode != access.status {
						t.Fatalf("POST %s: status=%d want=%d body=%s", path, response.StatusCode, access.status, body)
					}
				})
			}
		})
	}
}
