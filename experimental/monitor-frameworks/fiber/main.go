package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofurry/monitor"
)

const addr = ":18081"

func main() {
	app := fiber.New()
	m := monitor.NewMonitor(http.NotFoundHandler(), monitor.Config{
		Path:            "/monitor",
		Title:           "Fiber Monitor Demo",
		Description:     "Fiber native middleware records requests into monitor.",
		DefaultLanguage: "zh-CN",
		DefaultTheme:    "dark",
		Refresh:         2 * time.Second,
	})
	defer m.Stop()

	app.Use(func(c fiber.Ctx) error {
		if c.Path() == "/monitor" {
			return c.Next()
		}

		started := time.Now()
		m.RequestStarted()
		err := c.Next()

		status := c.Response().StatusCode()
		if err != nil {
			status = fiber.StatusInternalServerError
			if fiberErr, ok := err.(*fiber.Error); ok {
				status = fiberErr.Code
			}
		}
		m.RequestFinished(status, time.Since(started))
		return err
	})

	app.All("/monitor", adaptor.HTTPHandler(m))
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("fiber monitor demo")
	})
	app.Get("/ok", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"framework": "fiber", "status": "ok"})
	})
	app.Get("/slow", func(c fiber.Ctx) error {
		time.Sleep(250 * time.Millisecond)
		return c.SendString("fiber slow response")
	})
	app.Get("/error", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusInternalServerError, "fiber forced error")
	})

	fmt.Printf("Fiber demo listening on http://localhost%s\n", addr)
	fmt.Printf("Monitor: http://localhost%s/monitor\n", addr)
	if err := app.Listen(addr); err != nil {
		panic(err)
	}
}
