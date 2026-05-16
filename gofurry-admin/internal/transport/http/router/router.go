package router

import (
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	env "github.com/gofurry/awesome-fiber-template/v3/medium/config"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/bootstrap"
	applog "github.com/gofurry/awesome-fiber-template/v3/medium/internal/infra/logging"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/transport/http/webui"
	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/common"
	corazalite "github.com/gofurry/coraza-fiber-lite"
	swagger "github.com/gofiber/contrib/v3/swaggerui"
	"github.com/gofiber/fiber/v3"
	fibercompress "github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	fibercsrf "github.com/gofiber/fiber/v3/middleware/csrf"
	fiberetag "github.com/gofiber/fiber/v3/middleware/etag"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	fiberlogger "github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/pprof"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

type Builder struct{}

func New() *Builder {
	return &Builder{}
}

func (builder *Builder) Init() *fiber.App {
	cfg := env.GetServerConfig()
	appName := cfg.Server.AppName
	if appName == "" {
		appName = common.COMMON_PROJECT_NAME
	}

	app := fiber.New(fiber.Config{
		AppName:      appName,
		ServerHeader: appName,
		ErrorHandler: customErrorHandler,
		TrustProxy:   true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	registerMiddlewares(app)
	registerHealthRoutes(app, appName)
	registerCSRFTokenRoute(app)

	api(wrapTimeoutRouter(app.Group("/api"), cfg.Middleware.Timeout))

	if cfg.Server.IsFullStack {
		attachEmbeddedUI(app)
	}

	return app
}

func attachEmbeddedUI(app *fiber.App) {
	uiFS, err := fs.Sub(webui.FS, "dist")
	if err != nil {
		return
	}
	index, err := fs.ReadFile(uiFS, "index.html")
	if err != nil {
		return
	}

	sendIndex := func(c fiber.Ctx) error {
		c.Type("html", "utf-8")
		return c.Send(index)
	}

	app.Use(func(c fiber.Ctx) error {
		if c.Method() != fiber.MethodGet && c.Method() != fiber.MethodHead {
			return fiber.ErrNotFound
		}

		reqPath := c.Path()
		if reqPath == "/api" || strings.HasPrefix(reqPath, "/api/") || reqPath == "/v1" || strings.HasPrefix(reqPath, "/v1/") {
			return fiber.ErrNotFound
		}

		if reqPath == "/" || reqPath == "" {
			return sendIndex(c)
		}

		cleaned := path.Clean(reqPath)
		cleaned = strings.TrimPrefix(cleaned, "/")
		if cleaned == "." || cleaned == "" {
			return sendIndex(c)
		}

		if stat, err := fs.Stat(uiFS, cleaned); err == nil && !stat.IsDir() {
			return c.SendFile(cleaned, fiber.SendFile{FS: uiFS})
		}

		return sendIndex(c)
	})
}

func registerHealthRoutes(app *fiber.App, appName string) {
	cfg := env.GetServerConfig()
	if cfg.Middleware.Health.Enabled {
		app.Get(healthcheck.LivenessEndpoint, healthcheck.New(healthcheck.Config{
			Probe: func(c fiber.Ctx) bool {
				return bootstrap.Live()
			},
		}))
		app.Get(healthcheck.ReadinessEndpoint, healthcheck.New(healthcheck.Config{
			Probe: func(c fiber.Ctx) bool {
				return bootstrap.Ready()
			},
		}))
		app.Get(healthcheck.StartupEndpoint, healthcheck.New(healthcheck.Config{
			Probe: func(c fiber.Ctx) bool {
				return bootstrap.Started()
			},
		}))
	}

	if cfg.Middleware.Health.IncludeLegacy {
		app.Get("/healthz", func(c fiber.Ctx) error {
			ready := bootstrap.Ready()
			statusCode := fiber.StatusOK
			status := "ok"
			if !ready {
				statusCode = fiber.StatusServiceUnavailable
				status = "not_ready"
			}
			return c.Status(statusCode).JSON(common.ResultData{
				Code:    common.RETURN_SUCCESS,
				Message: "success",
				Data: fiber.Map{
					"name":    appName,
					"version": cfg.Server.AppVersion,
					"status":  status,
					"live":    bootstrap.Live(),
					"ready":   ready,
					"startup": bootstrap.Started(),
				},
			})
		})
	}
}

func registerCSRFTokenRoute(app *fiber.App) {
	cfg := env.GetServerConfig()
	if !cfg.Middleware.CSRF.Enabled {
		return
	}

	app.Get(cfg.Middleware.CSRF.TokenPath, func(c fiber.Ctx) error {
		token := fibercsrf.TokenFromContext(c)
		c.Set(fibercsrf.HeaderName, token)
		return common.NewResponse(c).SuccessWithData(fiber.Map{
			"token":       token,
			"header_name": fibercsrf.HeaderName,
			"cookie_name": cfg.Middleware.CSRF.CookieName,
		})
	})
}

func registerMiddlewares(app *fiber.App) {
	cfg := env.GetServerConfig()

	if cfg.Middleware.RequestID.Enabled {
		app.Use(requestid.New(requestid.Config{
			Header: cfg.Middleware.RequestID.Header,
		}))
	}

	if cfg.Middleware.AccessLog.Enabled {
		app.Use(fiberlogger.New(fiberlogger.Config{
			Format:        cfg.Middleware.AccessLog.Format,
			TimeFormat:    cfg.Middleware.AccessLog.TimeFormat,
			TimeZone:      cfg.Middleware.AccessLog.TimeZone,
			DisableColors: true,
			Stream:        io.Discard,
			Done: func(c fiber.Ctx, logString []byte) {
				line := strings.TrimSpace(string(logString))
				if line != "" {
					applog.Info(line)
				}
			},
		}))
	}

	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.Server.Mode == "debug",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.Middleware.Cors.AllowOrigins,
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodPatch,
			fiber.MethodOptions,
		},
		AllowHeaders: buildHeaderList(
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			cfg.Middleware.RequestID.Header,
			fibercsrf.HeaderName,
			fiber.HeaderIfNoneMatch,
		),
		AllowCredentials: true,
		ExposeHeaders: buildHeaderList(
			"Content-Length",
			cfg.Middleware.RequestID.Header,
			fibercsrf.HeaderName,
			fiber.HeaderETag,
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
			fiber.HeaderRetryAfter,
		),
		MaxAge: 86400,
	}))

	if cfg.Middleware.SecurityHeaders.Enabled {
		app.Use(helmet.New(helmet.Config{
			ContentSecurityPolicy: cfg.Middleware.SecurityHeaders.ContentSecurityPolicy,
			PermissionPolicy:      cfg.Middleware.SecurityHeaders.PermissionPolicy,
			HSTSMaxAge:            cfg.Middleware.SecurityHeaders.HSTSMaxAge,
			HSTSExcludeSubdomains: cfg.Middleware.SecurityHeaders.HSTSExcludeSubdomains,
			HSTSPreloadEnabled:    cfg.Middleware.SecurityHeaders.HSTSPreloadEnabled,
			CSPReportOnly:         cfg.Middleware.SecurityHeaders.CSPReportOnly,
		}))
	}

	if cfg.Middleware.Compression.Enabled {
		app.Use(fibercompress.New(fibercompress.Config{
			Level: compressionLevel(cfg.Middleware.Compression.Level),
		}))
	}

	if cfg.Middleware.ETag.Enabled {
		app.Use(fiberetag.New(fiberetag.Config{
			Weak: cfg.Middleware.ETag.Weak,
		}))
	}

	if cfg.Middleware.Limiter.Enabled {
		app.Use(limiter.New(limiter.Config{
			Max:               cfg.Middleware.Limiter.MaxRequests,
			Expiration:        time.Duration(cfg.Middleware.Limiter.Expiration) * time.Second,
			LimiterMiddleware: limiterStrategy(cfg.Middleware.Limiter.Strategy),
			KeyGenerator: func(c fiber.Ctx) string {
				return limiterKey(c, cfg.Middleware.Limiter)
			},
			Next: func(c fiber.Ctx) bool {
				return pathExcluded(c.Path(), cfg.Middleware.Limiter.ExcludePaths)
			},
			SkipFailedRequests:     cfg.Middleware.Limiter.SkipFailedRequests,
			SkipSuccessfulRequests: cfg.Middleware.Limiter.SkipSuccessfulRequests,
			DisableHeaders:         cfg.Middleware.Limiter.DisableHeaders,
			LimitReached: func(c fiber.Ctx) error {
				return common.NewResponse(c).ErrorWithCode(common.NewValidationError("too many requests"), fiber.StatusTooManyRequests)
			},
		}))
	}

	if cfg.Waf.Enabled {
		app.Use(corazalite.CorazaMiddleware())
	}

	if cfg.Middleware.CSRF.Enabled {
		app.Use(fibercsrf.New(fibercsrf.Config{
			Next: func(c fiber.Ctx) bool {
				return pathExcluded(c.Path(), cfg.Middleware.CSRF.ExcludePaths)
			},
			CookieName:        cfg.Middleware.CSRF.CookieName,
			CookieSameSite:    cfg.Middleware.CSRF.CookieSameSite,
			CookieSecure:      cfg.Middleware.CSRF.CookieSecure,
			CookieHTTPOnly:    cfg.Middleware.CSRF.CookieHTTPOnly,
			CookieSessionOnly: cfg.Middleware.CSRF.CookieSessionOnly,
			IdleTimeout:       time.Duration(cfg.Middleware.CSRF.IdleTimeoutSeconds) * time.Second,
			SingleUseToken:    cfg.Middleware.CSRF.SingleUseToken,
			TrustedOrigins:    cfg.Middleware.CSRF.TrustedOrigins,
			ErrorHandler: func(c fiber.Ctx, err error) error {
				message := "csrf validation failed"
				switch {
				case errors.Is(err, fibercsrf.ErrTokenNotFound):
					message = "csrf token not found"
				case errors.Is(err, fibercsrf.ErrTokenInvalid):
					message = "csrf token invalid"
				case err != nil:
					message = err.Error()
				}
				return common.NewResponse(c).ErrorWithCode(common.NewError(common.RETURN_FAILED, fiber.StatusForbidden, message), fiber.StatusForbidden)
			},
		}))
	}

	if cfg.Server.Mode == "debug" {
		app.Use(pprof.New())

		if cfg.Middleware.Swagger.Enabled {
			if _, err := os.Stat(cfg.Middleware.Swagger.FilePath); os.IsNotExist(err) {
				slog.Warn("swagger file does not exist, skip swagger middleware", "file", cfg.Middleware.Swagger.FilePath)
			} else {
				app.Use(swagger.New(swagger.Config{
					BasePath: cfg.Middleware.Swagger.BasePath,
					FilePath: cfg.Middleware.Swagger.FilePath,
					Path:     cfg.Middleware.Swagger.Path,
					Title:    cfg.Middleware.Swagger.Title,
				}))
			}
		}
	}
}

func customErrorHandler(c fiber.Ctx, err error) error {
	if appErr, ok := errors.AsType[common.Error](err); ok {
		return common.NewResponse(c).ErrorWithCode(appErr, appErr.GetHTTPStatus())
	}

	code := fiber.StatusInternalServerError
	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		code = fiberErr.Code
	}

	response := common.NewResponse(c)
	switch code {
	case fiber.StatusNotFound:
		return response.ErrorWithCode(common.NewError(common.RETURN_FAILED, code, "resource not found"), code)
	case fiber.StatusMethodNotAllowed:
		return response.ErrorWithCode(common.NewError(common.RETURN_FAILED, code, "method not allowed"), code)
	case fiber.StatusRequestTimeout:
		return response.ErrorWithCode(common.NewError(common.RETURN_FAILED, code, "request timeout"), code)
	default:
		if env.GetServerConfig().Server.Mode != "debug" {
			return response.ErrorWithCode(common.NewError(common.RETURN_FAILED, code, "internal server error"), code)
		}
		return response.ErrorWithCode(common.NewError(common.RETURN_FAILED, code, err.Error()), code)
	}
}

func buildHeaderList(items ...string) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[strings.ToLower(item)]; ok {
			continue
		}
		seen[strings.ToLower(item)] = struct{}{}
		result = append(result, item)
	}
	return result
}

func compressionLevel(level string) fibercompress.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "best_speed":
		return fibercompress.LevelBestSpeed
	case "best_compression":
		return fibercompress.LevelBestCompression
	default:
		return fibercompress.LevelDefault
	}
}

func limiterStrategy(strategy string) limiter.Handler {
	switch strings.ToLower(strings.TrimSpace(strategy)) {
	case "sliding":
		return limiter.SlidingWindow{}
	default:
		return limiter.FixedWindow{}
	}
}

func limiterKey(c fiber.Ctx, cfg env.LimiterConfig) string {
	switch cfg.KeySource {
	case "path":
		return c.Path()
	case "ip_path":
		return c.IP() + ":" + c.Path()
	case "header":
		value := strings.TrimSpace(c.Get(cfg.KeyHeader))
		if value != "" {
			return value
		}
		return c.IP()
	default:
		return c.IP()
	}
}

func pathExcluded(current string, paths []string) bool {
	current = normalizePath(current)
	for _, item := range paths {
		if current == normalizePath(item) {
			return true
		}
	}
	return false
}

func normalizePath(path string) string {
	normalized := strings.TrimRight(strings.TrimSpace(path), "/")
	if normalized == "" {
		return "/"
	}
	return normalized
}
