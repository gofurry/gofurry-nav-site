package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofurry/monitor"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const addr = ":18083"

func main() {
	e := echo.New()
	m := monitor.NewMonitor(http.NotFoundHandler(), monitor.Config{
		Path:            "/monitor",
		Title:           "Echo Monitor Demo",
		Description:     "Echo native middleware records requests into monitor.",
		DefaultLanguage: "zh-CN",
		DefaultTheme:    "dark",
		Refresh:         2 * time.Second,
	})
	defer m.Stop()

	e.Use(monitorEcho(m), middleware.Recover())
	e.GET("/monitor", echo.WrapHandler(m))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "echo monitor demo")
	})
	e.GET("/ok", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"framework": "echo", "status": "ok"})
	})
	e.GET("/slow", func(c echo.Context) error {
		time.Sleep(250 * time.Millisecond)
		return c.String(http.StatusOK, "echo slow response")
	})
	e.GET("/error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "echo forced error")
	})

	fmt.Printf("Echo demo listening on http://localhost%s\n", addr)
	fmt.Printf("Monitor: http://localhost%s/monitor\n", addr)
	if err := e.Start(addr); err != nil {
		panic(err)
	}
}

func monitorEcho(m *monitor.Monitor) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if c.Request().URL.Path == "/monitor" {
				return next(c)
			}

			started := time.Now()
			m.RequestStarted()
			err = next(c)

			status := c.Response().Status
			if err != nil {
				status = http.StatusInternalServerError
				if echoErr, ok := err.(*echo.HTTPError); ok {
					status = echoErr.Code
				}
			}
			if status == 0 {
				status = http.StatusOK
			}
			m.RequestFinished(status, time.Since(started))
			return err
		}
	}
}
