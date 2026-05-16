package router

import (
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	env "github.com/gofurry/gofurry-rag/config"
	"github.com/gofurry/gofurry-rag/internal/api"
	"github.com/gofurry/gofurry-rag/internal/bootstrap"
	"github.com/gofurry/gofurry-rag/internal/web"
	"github.com/gofurry/gofurry-rag/pkg/common"
	"github.com/gofiber/fiber/v3"
	fibercompress "github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	fiberetag "github.com/gofiber/fiber/v3/middleware/etag"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	fiberlogger "github.com/gofiber/fiber/v3/middleware/logger"
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
		ErrorHandler: api.ErrorHandler,
		TrustProxy:   true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 90 * time.Second,
	})

	registerMiddlewares(app)
	registerHealthRoutes(app, appName)
	api.NewServer(*cfg, bootstrap.RAGService(), bootstrap.Worker()).RegisterRoutes(wrapTimeoutRouter(app.Group("/api"), cfg.Middleware.Timeout).Group("/v1"))

	if cfg.Server.IsFullStack {
		attachEmbeddedUI(app)
	}
	return app
}

func registerHealthRoutes(app *fiber.App, appName string) {
	cfg := env.GetServerConfig()
	if cfg.Middleware.Health.Enabled {
		app.Get(healthcheck.LivenessEndpoint, healthcheck.New(healthcheck.Config{
			Probe: func(c fiber.Ctx) bool { return bootstrap.Live() },
		}))
		app.Get(healthcheck.ReadinessEndpoint, healthcheck.New(healthcheck.Config{
			Probe: func(c fiber.Ctx) bool { return bootstrap.Ready() },
		}))
		app.Get(healthcheck.StartupEndpoint, healthcheck.New(healthcheck.Config{
			Probe: func(c fiber.Ctx) bool { return bootstrap.Started() },
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
			return c.Status(statusCode).JSON(api.Result{
				Code:    1,
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

func registerMiddlewares(app *fiber.App) {
	cfg := env.GetServerConfig()

	if cfg.Middleware.RequestID.Enabled {
		app.Use(requestid.New(requestid.Config{Header: cfg.Middleware.RequestID.Header}))
	}
	if cfg.Middleware.AccessLog.Enabled {
		app.Use(fiberlogger.New(fiberlogger.Config{
			Format:        cfg.Middleware.AccessLog.Format,
			TimeFormat:    cfg.Middleware.AccessLog.TimeFormat,
			TimeZone:      cfg.Middleware.AccessLog.TimeZone,
			DisableColors: true,
			Stream:        io.Discard,
			Done: func(c fiber.Ctx, logString []byte) {
				if line := strings.TrimSpace(string(logString)); line != "" {
					slog.Info(line)
				}
			},
		}))
	}

	app.Use(recover.New(recover.Config{EnableStackTrace: cfg.Server.Mode == "debug"}))
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
		AllowHeaders:     buildHeaderList("Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", cfg.Middleware.RequestID.Header, fiber.HeaderIfNoneMatch),
		AllowCredentials: true,
		ExposeHeaders:    buildHeaderList("Content-Length", cfg.Middleware.RequestID.Header, fiber.HeaderETag, "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", fiber.HeaderRetryAfter),
		MaxAge:           86400,
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
			Next: func(c fiber.Ctx) bool {
				return c.Path() == "/api/v1/chat/stream"
			},
		}))
	}
	if cfg.Middleware.ETag.Enabled {
		app.Use(fiberetag.New(fiberetag.Config{
			Weak: cfg.Middleware.ETag.Weak,
			Next: func(c fiber.Ctx) bool {
				return c.Path() == "/api/v1/chat/stream"
			},
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
				return api.ErrorWithCode(c, fiber.StatusTooManyRequests, "too many requests")
			},
		}))
	}
}

func attachEmbeddedUI(app *fiber.App) {
	uiFS, err := fs.Sub(web.FS, "dist")
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
	app.Get("/", func(c fiber.Ctx) error {
		return c.Redirect().To("/admin")
	})
	app.Get("/admin", sendIndex)
	app.Get("/admin/*", func(c fiber.Ctx) error {
		asset := strings.TrimPrefix(c.Path(), "/admin/")
		if asset == "" || asset == "." {
			return sendIndex(c)
		}
		cleaned := path.Clean(asset)
		if stat, err := fs.Stat(uiFS, cleaned); err == nil && !stat.IsDir() {
			return c.SendFile(cleaned, fiber.SendFile{FS: uiFS})
		}
		return sendIndex(c)
	})
	app.Use(func(c fiber.Ctx) error {
		if strings.HasPrefix(c.Path(), "/api/") {
			return fiber.ErrNotFound
		}
		return c.Status(http.StatusNotFound).SendString("not found")
	})
}

func buildHeaderList(items ...string) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
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
		if value := strings.TrimSpace(c.Get(cfg.KeyHeader)); value != "" {
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
