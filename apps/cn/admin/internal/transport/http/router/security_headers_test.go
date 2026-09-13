package router

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/gofurry-admin/config"
)

func TestManagedAssetPreviewHeaders(t *testing.T) {
	// Use tracked configuration: this test needs no databases or cloud credentials.
	if err := env.MustInitServerConfig("gofurry-admin", "../../../../config/server.example.yaml"); err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	registerMiddlewares(app)
	attachEmbeddedUI(app)
	for _, path := range []string{"/nav/sites/106", "/nav/hero-assets", "/nav/background-patterns"} {
		response, err := app.Test(httptest.NewRequest("GET", path, nil))
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != 200 {
			t.Fatalf("%s: status %d", path, response.StatusCode)
		}
		for name, want := range map[string]string{
			"Cross-Origin-Embedder-Policy": "unsafe-none",
			"Cross-Origin-Opener-Policy":   "same-origin",
			"X-Content-Type-Options":       "nosniff",
			"X-Frame-Options":              "SAMEORIGIN",
		} {
			if got := response.Header.Get(name); got != want {
				t.Errorf("%s: %s = %q, want %q", path, name, got, want)
			}
		}
	}
}
