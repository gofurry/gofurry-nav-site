package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
	gameadmin "github.com/gofurry/gofurry-admin/internal/app/gameadmin/controller"
	"github.com/gofurry/gofurry-admin/internal/bootstrap"
	"net/http/httptest"
	"testing"
)

func TestGameCollectionRouteCapabilities(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		caps                    []authorization.Capability
		writeStatus, readStatus int
	}{
		{"writer", []authorization.Capability{authorization.ContentRead, authorization.ContentWrite}, 400, 400},
		{"reader", []authorization.Capability{authorization.ContentRead}, 403, 400},
		{"other", []authorization.Capability{authorization.CollectionRead}, 403, 403},
		{"anonymous", nil, 401, 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c fiber.Ctx) error {
				if tc.caps != nil {
					c.Locals(authorization.PrincipalContextKey, &authorization.Principal{Capabilities: tc.caps})
				}
				return c.Next()
			})
			gameCollectionRoutes(app.Group("/api/v1/game/collections"), &bootstrap.Runtime{GameAPI: gameadmin.New(nil, nil)})
			for _, r := range []struct {
				method, path string
				want         int
			}{
				{"GET", "?status=invalid", tc.readStatus}, {"POST", "", tc.writeStatus},
				{"GET", "/invalid", tc.readStatus}, {"PUT", "/invalid", tc.writeStatus},
				{"GET", "/invalid/composition", tc.readStatus}, {"PUT", "/invalid/composition", tc.writeStatus},
				{"GET", "/invalid/members", tc.readStatus}, {"PUT", "/invalid/members", tc.writeStatus},
				{"PUT", "/home-curation", tc.writeStatus}, {"POST", "/invalid/publish", tc.writeStatus}, {"POST", "/invalid/unpublish", tc.writeStatus}, {"POST", "/invalid/archive", tc.writeStatus}, {"POST", "/invalid/restore", tc.writeStatus},
			} {
				res, e := app.Test(httptest.NewRequest(r.method, "/api/v1/game/collections"+r.path, nil))
				if e != nil {
					t.Fatal(e)
				}
				res.Body.Close()
				if res.StatusCode != r.want {
					t.Fatalf("%s %s: %d != %d", r.method, r.path, res.StatusCode, r.want)
				}
			}
		})
	}
}
