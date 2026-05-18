package httpapi

import (
	"encoding/json"
	"io/fs"
	"path"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	fibercompress "github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/etag"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/config"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/model"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/security"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/service"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/web"
)

type Server struct {
	cfg       config.Config
	svc       *service.Service
	startedAt time.Time
}

func New(cfg config.Config, svc *service.Service) *fiber.App {
	server := &Server{cfg: cfg, svc: svc, startedAt: time.Now().UTC()}
	app := fiber.New(fiber.Config{
		AppName:      cfg.CenterID,
		ServerHeader: cfg.CenterID,
		ErrorHandler: ErrorHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	})
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://127.0.0.1:5179", "http://127.0.0.1:8080"},
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodDelete, fiber.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-GoFurry-Node-ID", "X-GoFurry-Timestamp", "X-GoFurry-Signature"},
		AllowCredentials: true,
		MaxAge:           86400,
	}))
	app.Use(fibercompress.New())
	app.Use(etag.New())
	app.Get(healthcheck.LivenessEndpoint, healthcheck.New(healthcheck.Config{Probe: func(c fiber.Ctx) bool { return true }}))
	app.Get(healthcheck.ReadinessEndpoint, healthcheck.New(healthcheck.Config{Probe: func(c fiber.Ctx) bool { return server.svc != nil }}))
	app.Get("/healthz", server.health)

	v1 := app.Group("/api/v1")
	v1.Post("/agent/ingest", server.agentIngest)
	v1.Get("/peer/summary", server.requirePeer, server.peerSummary)
	v1.Post("/peer/heartbeat", server.requirePeer, server.peerHeartbeat)
	v1.Post("/events/sync", server.requireEvent, server.createSyncRun)
	v1.Post("/events/deploy", server.requireEvent, server.createDeployEvent)

	auth := v1.Group("/admin/auth")
	auth.Get("/state", server.authState)
	auth.Post("/login", server.authLogin)
	auth.Post("/logout", server.authLogout)
	auth.Get("/me", server.requireAdmin, server.authMe)

	admin := v1.Group("/dashboard", server.requireAdmin)
	admin.Get("/overview", server.overview)
	admin.Get("/metrics/overview", server.metricsOverview)
	admin.Get("/nodes", server.nodes)
	admin.Get("/nodes/:id/metrics", server.nodeMetrics)
	admin.Get("/nodes/:id", server.node)
	admin.Get("/services", server.services)
	admin.Get("/alerts", server.alerts)
	admin.Get("/sync-runs", server.syncRuns)
	admin.Get("/peer/status", server.peerStatus)
	admin.Get("/deployments", server.deployments)
	attachEmbeddedUI(app)
	return app
}

func (s *Server) health(c fiber.Ctx) error {
	return ok(c, fiber.Map{
		"center_id":  s.cfg.CenterID,
		"region":     s.cfg.Region,
		"status":     "ok",
		"started_at": s.startedAt,
	})
}

func (s *Server) agentIngest(c fiber.Ctx) error {
	body := c.Body()
	nodeID := strings.TrimSpace(c.Get("X-GoFurry-Node-ID"))
	timestamp := strings.TrimSpace(c.Get("X-GoFurry-Timestamp"))
	signature := strings.TrimSpace(c.Get("X-GoFurry-Signature"))
	token := bearerToken(c.Get(fiber.HeaderAuthorization))
	expected, okToken := s.cfg.AgentTokenMap()[nodeID]
	if nodeID == "" || !okToken || token == "" || token != expected {
		return fail(c, fiber.StatusUnauthorized, "invalid agent token")
	}
	if err := security.CheckTimestamp(timestamp, s.cfg.Security.SignatureWindow.Duration, time.Now().UTC()); err != nil {
		return fail(c, fiber.StatusUnauthorized, "invalid timestamp")
	}
	if !security.Verify(expected, timestamp, nodeID, body, signature) {
		return fail(c, fiber.StatusUnauthorized, "invalid signature")
	}
	var payload model.AgentPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid json")
	}
	if payload.NodeID != nodeID {
		return fail(c, fiber.StatusBadRequest, "node_id mismatch")
	}
	if err := s.svc.Ingest(requestContext(c), payload); err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	return accepted(c, fiber.Map{"accepted": true, "node_id": nodeID})
}

func (s *Server) peerSummary(c fiber.Ctx) error {
	result, err := s.svc.PeerSummary(requestContext(c))
	if err != nil {
		return fail(c, fiber.StatusInternalServerError, err.Error())
	}
	return ok(c, result)
}

func (s *Server) peerHeartbeat(c fiber.Ctx) error {
	var req model.PeerSummary
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid json")
	}
	if err := s.svc.RecordPeerSummary(requestContext(c), req); err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	return accepted(c, fiber.Map{"accepted": true})
}

func (s *Server) createSyncRun(c fiber.Ctx) error {
	var req model.SyncEventRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid json")
	}
	result, err := s.svc.CreateSyncRun(requestContext(c), req)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	return accepted(c, result)
}

func (s *Server) createDeployEvent(c fiber.Ctx) error {
	var req model.DeployEventRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid json")
	}
	result, err := s.svc.CreateDeployEvent(requestContext(c), req)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	return accepted(c, result)
}

func (s *Server) overview(c fiber.Ctx) error {
	result, err := s.svc.Overview(requestContext(c))
	if err != nil {
		return fail(c, fiber.StatusInternalServerError, err.Error())
	}
	return ok(c, result)
}

func (s *Server) metricsOverview(c fiber.Ctx) error {
	result, err := s.svc.OverviewMetrics(requestContext(c), c.Query("range", "1h"))
	if err != nil {
		return fail(c, fiber.StatusInternalServerError, err.Error())
	}
	return ok(c, result)
}

func (s *Server) nodes(c fiber.Ctx) error {
	result, err := s.svc.Nodes(requestContext(c))
	if err != nil {
		return fail(c, fiber.StatusInternalServerError, err.Error())
	}
	return ok(c, result)
}

func (s *Server) node(c fiber.Ctx) error {
	result, err := s.svc.Node(requestContext(c), c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusNotFound, "node not found")
	}
	return ok(c, result)
}

func (s *Server) nodeMetrics(c fiber.Ctx) error {
	result, err := s.svc.NodeMetrics(requestContext(c), c.Params("id"), c.Query("range", "1h"))
	if err != nil {
		return fail(c, fiber.StatusNotFound, "node not found")
	}
	return ok(c, result)
}

func (s *Server) services(c fiber.Ctx) error {
	result, err := s.svc.Services(requestContext(c))
	if err != nil {
		return fail(c, fiber.StatusInternalServerError, err.Error())
	}
	return ok(c, result)
}

func (s *Server) alerts(c fiber.Ctx) error {
	activeOnly := c.Query("active", "true") != "false"
	result, err := s.svc.Alerts(requestContext(c), activeOnly)
	if err != nil {
		return fail(c, fiber.StatusInternalServerError, err.Error())
	}
	return ok(c, result)
}

func (s *Server) syncRuns(c fiber.Ctx) error {
	result, err := s.svc.SyncRuns(requestContext(c), queryInt(c, "limit", 50))
	if err != nil {
		return fail(c, fiber.StatusInternalServerError, err.Error())
	}
	return ok(c, result)
}

func (s *Server) peerStatus(c fiber.Ctx) error {
	result, err := s.svc.PeerStatus(requestContext(c))
	if err != nil {
		return fail(c, fiber.StatusInternalServerError, err.Error())
	}
	return ok(c, result)
}

func (s *Server) deployments(c fiber.Ctx) error {
	result, err := s.svc.DeployEvents(requestContext(c), queryInt(c, "limit", 50))
	if err != nil {
		return fail(c, fiber.StatusInternalServerError, err.Error())
	}
	return ok(c, result)
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
}
