package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/gofurry-rag/config"
)

func TestProxyHeaderIgnoredWhenTrustProxyDisabled(t *testing.T) {
	app := fiber.New(newFiberConfig(env.Config{
		Server: env.ServerConfig{
			ProxyHeader: "X-Forwarded-For",
		},
	}, "test"))
	app.Get("/ip", func(c fiber.Ctx) error {
		return c.SendString(c.IP())
	})

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.10")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) == "203.0.113.10" {
		t.Fatalf("spoofed proxy header was trusted: %q", body)
	}
}

func TestProxyHeaderUsedWhenProxyIsTrusted(t *testing.T) {
	app := fiber.New(newFiberConfig(env.Config{
		Server: env.ServerConfig{
			TrustProxy:     true,
			ProxyHeader:    "X-Forwarded-For",
			TrustedProxies: []string{"0.0.0.0"},
		},
	}, "test"))
	app.Get("/ip", func(c fiber.Ctx) error {
		return c.SendString(c.IP())
	})

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.10")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "203.0.113.10" {
		t.Fatalf("trusted proxy header ignored: %q", body)
	}
}
