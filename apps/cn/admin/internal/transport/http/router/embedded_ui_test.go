package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestEmbeddedReactSPAAndAPIBoundary(t *testing.T) {
	app := fiber.New()
	attachEmbeddedUI(app)

	index := responseBody(t, app, "/")
	if !strings.Contains(index, `<div id="root"></div>`) {
		t.Fatal("embedded index is not the React production entrypoint")
	}

	for _, route := range []string{
		"/login", "/setup", "/nav/sites", "/nav/sites/42", "/game/games", "/game/games/42",
		"/collection", "/metrics", "/changes", "/system/data-operations", "/system/audit", "/system/accounts",
	} {
		if body := responseBody(t, app, route); body != index {
			t.Fatalf("nested SPA route %s did not return the React index", route)
		}
	}

	assetPattern := regexp.MustCompile(`(?:src|href)="(/assets/[^"]+)"`)
	match := assetPattern.FindStringSubmatch(index)
	if len(match) != 2 {
		t.Fatal("React index does not reference a production asset")
	}
	assetResponse, err := app.Test(httptest.NewRequest("GET", match[1], nil))
	if err != nil {
		t.Fatalf("request embedded asset: %v", err)
	}
	defer assetResponse.Body.Close()
	if assetResponse.StatusCode != fiber.StatusOK {
		t.Fatalf("embedded asset status=%d", assetResponse.StatusCode)
	}
	if body := readResponse(t, assetResponse); body == index || len(body) == 0 {
		t.Fatal("asset request returned the SPA index or an empty body")
	}

	apiResponse, err := app.Test(httptest.NewRequest("GET", "/api/v1/not-a-route", nil))
	if err != nil {
		t.Fatalf("request unknown API route: %v", err)
	}
	defer apiResponse.Body.Close()
	if apiResponse.StatusCode != fiber.StatusNotFound {
		t.Fatalf("unknown API status=%d, want 404", apiResponse.StatusCode)
	}
	if body := readResponse(t, apiResponse); body == index || strings.Contains(body, `<div id="root"></div>`) {
		t.Fatal("unknown API route leaked the SPA index")
	}
}

func responseBody(t *testing.T, app *fiber.App, target string) string {
	t.Helper()
	response, err := app.Test(httptest.NewRequest("GET", target, nil))
	if err != nil {
		t.Fatalf("request %s: %v", target, err)
	}
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("request %s status=%d", target, response.StatusCode)
	}
	return readResponse(t, response)
}

func readResponse(t *testing.T, response *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	return string(body)
}
