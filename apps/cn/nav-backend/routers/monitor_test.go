package routers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func TestMonitorServesDashboardAndCountsDownstreamTraffic(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: func(c fiber.Ctx, err error) error {
		if err == fiber.ErrForbidden {
			return c.SendStatus(fiber.StatusForbidden)
		}
		return c.SendStatus(fiber.StatusInternalServerError)
	}})
	registerMonitor(app)
	app.Use(recover.New())
	app.Get("/ordinary", func(c fiber.Ctx) error { return c.Status(201).SendString("ordinary response") })
	app.Get("/denied", func(c fiber.Ctx) error { return fiber.ErrForbidden })
	app.Get("/panic", func(c fiber.Ctx) error { panic("test panic") })
	for _, tc := range []struct {
		path   string
		status int
		text   string
	}{
		{"/monitor", 200, "GoFurry Nav Monitor"},
		{"/MONITOR/", 200, "GoFurry Nav Monitor"},
		{"/ordinary", 201, "ordinary response"},
		{"/denied", 403, "Forbidden"},
		{"/panic", 500, "Internal Server Error"},
	} {
		resp, err := app.Test(httptest.NewRequest("GET", tc.path, nil))
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil || resp.StatusCode != tc.status || !strings.Contains(string(body), tc.text) {
			t.Fatalf("%s: status=%d body=%s err=%v", tc.path, resp.StatusCode, body, err)
		}
	}
	request := httptest.NewRequest("GET", "/monitor", nil)
	request.Header.Set("Accept", "application/json")
	resp, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var snapshot struct {
		HTTP struct {
			Requests uint64            `json:"requests"`
			InFlight uint64            `json:"in_flight"`
			Status   map[string]uint64 `json:"status"`
		} `json:"http"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 || snapshot.HTTP.Requests != 3 || snapshot.HTTP.InFlight != 0 || snapshot.HTTP.Status["2xx"] != 1 || snapshot.HTTP.Status["4xx"] != 1 || snapshot.HTTP.Status["5xx"] != 1 {
		t.Fatalf("monitor should count completed downstream responses, excluding itself: %+v", snapshot)
	}
}

func TestMonitorStateBelongsToEachApp(t *testing.T) {
	for range 2 {
		app := fiber.New()
		registerMonitor(app)
		app.Get("/ok", func(c fiber.Ctx) error { return c.SendStatus(200) })
		resp, err := app.Test(httptest.NewRequest("GET", "/ok", nil))
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		req := httptest.NewRequest("GET", "/monitor", nil)
		req.Header.Set("Accept", "application/json")
		resp, err = app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		var snapshot struct {
			HTTP struct {
				Requests uint64 `json:"requests"`
			} `json:"http"`
		}
		err = json.NewDecoder(resp.Body).Decode(&snapshot)
		resp.Body.Close()
		if err != nil || snapshot.HTTP.Requests != 1 {
			t.Fatalf("shared or stopped monitor: %+v, %v", snapshot, err)
		}
	}
}
