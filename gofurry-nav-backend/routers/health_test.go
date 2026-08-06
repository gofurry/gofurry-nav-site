package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
)

func TestHealthEndpoints(t *testing.T) {
	app := fiber.New()
	registerHealthChecks(app)

	liveness, err := app.Test(httptest.NewRequest(http.MethodGet, healthcheck.LivenessEndpoint, http.NoBody))
	if err != nil {
		t.Fatalf("request liveness endpoint: %v", err)
	}
	defer liveness.Body.Close()
	if liveness.StatusCode != http.StatusOK {
		t.Fatalf("liveness status = %d", liveness.StatusCode)
	}

	readiness, err := app.Test(httptest.NewRequest(http.MethodGet, healthcheck.ReadinessEndpoint, http.NoBody))
	if err != nil {
		t.Fatalf("request readiness endpoint: %v", err)
	}
	defer readiness.Body.Close()
	if readiness.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("readiness without initialized dependencies = %d", readiness.StatusCode)
	}
}
