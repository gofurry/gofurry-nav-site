package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofurry/monitor"
)

const addr = ":18082"

func main() {
	r := gin.New()
	m := monitor.NewMonitor(http.NotFoundHandler(), monitor.Config{
		Path:            "/monitor",
		Title:           "Gin Monitor Demo",
		Description:     "Gin native middleware records requests into monitor.",
		DefaultLanguage: "zh-CN",
		DefaultTheme:    "dark",
		Refresh:         2 * time.Second,
	})
	defer m.Stop()

	r.Use(monitorGin(m), gin.Recovery())
	r.GET("/monitor", gin.WrapH(m))
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "gin monitor demo")
	})
	r.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"framework": "gin", "status": "ok"})
	})
	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(250 * time.Millisecond)
		c.String(http.StatusOK, "gin slow response")
	})
	r.GET("/error", func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "gin forced error"})
	})

	fmt.Printf("Gin demo listening on http://localhost%s\n", addr)
	fmt.Printf("Monitor: http://localhost%s/monitor\n", addr)
	if err := r.Run(addr); err != nil {
		panic(err)
	}
}

func monitorGin(m *monitor.Monitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/monitor" {
			c.Next()
			return
		}

		started := time.Now()
		m.RequestStarted()
		c.Next()

		status := c.Writer.Status()
		if status == 0 {
			status = http.StatusOK
		}
		m.RequestFinished(status, time.Since(started))
	}
}
