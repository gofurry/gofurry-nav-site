package middleware

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
)

func TestRequireCapability(t *testing.T) {
	tests := []struct {
		name       string
		role       authorization.Role
		capability authorization.Capability
		wantStatus int
	}{
		{name: "operator content mutation", role: authorization.RoleOperator, capability: authorization.ContentWrite, wantStatus: fiber.StatusOK},
		{name: "operator run now", role: authorization.RoleOperator, capability: authorization.CollectionExecute, wantStatus: fiber.StatusOK},
		{name: "operator schedule mutation denied", role: authorization.RoleOperator, capability: authorization.CollectionControl, wantStatus: fiber.StatusForbidden},
		{name: "operator metric read", role: authorization.RoleOperator, capability: authorization.MetricsRead, wantStatus: fiber.StatusOK},
		{name: "operator metric technical denied", role: authorization.RoleOperator, capability: authorization.MetricsTechnical, wantStatus: fiber.StatusForbidden},
		{name: "operator change read", role: authorization.RoleOperator, capability: authorization.ChangesRead, wantStatus: fiber.StatusOK},
		{name: "operator change technical denied", role: authorization.RoleOperator, capability: authorization.ChangesTechnical, wantStatus: fiber.StatusForbidden},
		{name: "developer schedule mutation", role: authorization.RoleDeveloper, capability: authorization.CollectionControl, wantStatus: fiber.StatusOK},
		{name: "developer metric technical", role: authorization.RoleDeveloper, capability: authorization.MetricsTechnical, wantStatus: fiber.StatusOK},
		{name: "developer change technical", role: authorization.RoleDeveloper, capability: authorization.ChangesTechnical, wantStatus: fiber.StatusOK},
		{name: "developer dataops", role: authorization.RoleDeveloper, capability: authorization.DataOpsRead, wantStatus: fiber.StatusOK},
		{name: "developer audit", role: authorization.RoleDeveloper, capability: authorization.AuditRead, wantStatus: fiber.StatusOK},
		{name: "developer account management denied", role: authorization.RoleDeveloper, capability: authorization.AccountManage, wantStatus: fiber.StatusForbidden},
		{name: "operator dataops denied", role: authorization.RoleOperator, capability: authorization.DataOpsRead, wantStatus: fiber.StatusForbidden},
		{name: "operator audit denied", role: authorization.RoleOperator, capability: authorization.AuditRead, wantStatus: fiber.StatusForbidden},
		{name: "owner account management", role: authorization.RoleOwner, capability: authorization.AccountManage, wantStatus: fiber.StatusOK},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c fiber.Ctx) error {
				c.Locals(authorization.PrincipalContextKey, &authorization.Principal{
					AccountID: 1, Role: test.role, Status: authorization.StatusActive,
					Capabilities: authorization.CapabilitiesFor(test.role),
				})
				return c.Next()
			}, Require(test.capability), func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
			response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			_, _ = io.Copy(io.Discard, response.Body)
			if response.StatusCode != test.wantStatus {
				t.Fatalf("status=%d, want %d", response.StatusCode, test.wantStatus)
			}
		})
	}
}

func TestRequireCapabilityWithoutPrincipalIsUnauthorized(t *testing.T) {
	app := fiber.New()
	app.Get("/", Require(authorization.ContentRead), func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status=%d, want %d", response.StatusCode, fiber.StatusUnauthorized)
	}
}
