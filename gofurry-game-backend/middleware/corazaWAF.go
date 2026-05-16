package middleware

import (
	"log/slog"
	"net/http"
	"sync"

	fibercoraza "github.com/gofiber/contrib/v3/coraza"
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/roof/env"
)

var (
	globalWAF     fiber.Handler
	wafOnce       sync.Once
	globalWAFErr  error
	globalWAFConf env.WafConfig
)

func InitGlobalWAF(cfg env.WafConfig) {
	wafOnce.Do(func() {
		globalWAFConf = cfg

		corazaCfg := buildCorazaConfig(cfg)
		if _, err := fibercoraza.NewEngine(corazaCfg); err != nil {
			globalWAFErr = err
			slog.Error("[CorazaWAF] init failed", "error", err, "directives_files", corazaCfg.DirectivesFile)
			return
		}

		globalWAF = fibercoraza.New(corazaCfg)
	})
}

func CorazaMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		if globalWAFErr != nil {
			slog.Error("[CorazaWAF] unavailable", "error", globalWAFErr)
			return common.NewResponse(c).ErrorWithCode("WAF initialization failed", http.StatusInternalServerError)
		}

		if globalWAF == nil {
			slog.Error("[CorazaWAF] handler not initialized", "waf_enabled", globalWAFConf.WafSwitch)
			return common.NewResponse(c).ErrorWithCode("WAF is not initialized", http.StatusInternalServerError)
		}

		return globalWAF(c)
	}
}

func buildCorazaConfig(cfg env.WafConfig) fibercoraza.Config {
	corazaCfg := fibercoraza.ConfigDefault
	corazaCfg.Next = func(c fiber.Ctx) bool {
		return c.Path() == "/metrics"
	}
	corazaCfg.DirectivesFile = cfg.ResolveDirectivesFiles()
	corazaCfg.BlockMessage = "Request blocked by WAF"
	corazaCfg.BlockHandler = func(c fiber.Ctx, details fibercoraza.InterruptionDetails) error {
		status := details.StatusCode
		if status < http.StatusBadRequest {
			status = http.StatusForbidden
		}
		c.Set("X-WAF-Blocked", "true")
		return common.NewResponse(c).ErrorWithCode("您的请求存在安全风险，已被系统拦截。", status)
	}
	corazaCfg.ErrorHandler = func(c fiber.Ctx, failure fibercoraza.MiddlewareError) error {
		slog.Error("[CorazaWAF] request processing failed", "error", failure.Err, "code", failure.Code)
		status := failure.StatusCode
		if status < http.StatusBadRequest {
			status = http.StatusInternalServerError
		}
		return common.NewResponse(c).ErrorWithCode("WAF request processing failed", status)
	}
	return corazaCfg
}
