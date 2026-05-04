package api

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GoFurry/gofurry-rag/internal/config"
	"github.com/GoFurry/gofurry-rag/internal/db"
	"github.com/GoFurry/gofurry-rag/internal/ingest"
	"github.com/GoFurry/gofurry-rag/internal/service"
	"github.com/GoFurry/gofurry-rag/internal/web"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

type Server struct {
	cfg     config.Config
	service *service.Service
	worker  *ingest.Worker
	app     *fiber.App
}

func NewServer(cfg config.Config, svc *service.Service, worker *ingest.Worker) *Server {
	server := &Server{cfg: cfg, service: svc, worker: worker}
	server.app = server.build()
	return server
}

func (s *Server) App() *fiber.App {
	return s.app
}

func (s *Server) build() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      s.cfg.AppName,
		ServerHeader: s.cfg.AppName,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return fail(c, err)
		},
	})
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods: []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodDelete, fiber.MethodOptions},
	}))

	api := app.Group("/api/v1")
	api.Get("/health", s.health)
	admin := api.Group("/admin", s.requireAdmin)
	admin.Post("/documents/text", s.createTextDocument)
	admin.Get("/documents", s.listDocuments)
	admin.Get("/documents/:id/chunks", s.listChunks)
	admin.Delete("/documents/:id", s.deleteDocument)
	api.Post("/chat/query", s.query)

	attachAdminUI(app)
	return app
}

func (s *Server) requireAdmin(c fiber.Ctx) error {
	token := strings.TrimSpace(strings.TrimPrefix(c.Get("Authorization"), "Bearer "))
	if token == "" || token != s.cfg.AdminToken {
		return fiber.ErrUnauthorized
	}
	return c.Next()
}

func (s *Server) health(c fiber.Ctx) error {
	ctx := context.Background()
	return ok(c, s.service.Health(ctx))
}

func (s *Server) createTextDocument(c fiber.Ctx) error {
	var req service.TextDocumentRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fail(c, err)
	}
	doc, err := s.service.CreateTextDocument(context.Background(), req)
	if err != nil {
		return fail(c, err)
	}
	return ok(c, fiber.Map{"document_id": doc.ID, "status": doc.Status})
}

func (s *Server) listDocuments(c fiber.Ctx) error {
	result, err := s.service.ListDocuments(context.Background(), db.ListDocumentsFilter{
		Page:       queryInt(c, "page", 1),
		PageSize:   queryInt(c, "page_size", 20),
		Status:     c.Query("status"),
		SourceType: c.Query("source_type"),
		Keyword:    c.Query("keyword"),
	})
	if err != nil {
		return fail(c, err)
	}
	return ok(c, result)
}

func (s *Server) listChunks(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fail(c, service.ErrValidation)
	}
	result, err := s.service.ListChunks(context.Background(), id, queryInt(c, "page", 1), queryInt(c, "page_size", 20))
	if err != nil {
		return fail(c, err)
	}
	return ok(c, result)
}

func (s *Server) deleteDocument(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fail(c, service.ErrValidation)
	}
	if err := s.service.DeleteDocument(context.Background(), id); err != nil {
		return fail(c, err)
	}
	return ok(c, fiber.Map{"deleted": true})
}

func (s *Server) query(c fiber.Ctx) error {
	var req service.QueryRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fail(c, err)
	}
	result, err := s.service.Query(context.Background(), req)
	if err != nil {
		return fail(c, err)
	}
	return ok(c, result)
}

func queryInt(c fiber.Ctx, key string, fallback int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func attachAdminUI(app *fiber.App) {
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
		if stat, err := fs.Stat(uiFS, asset); err == nil && !stat.IsDir() {
			return c.SendFile(asset, fiber.SendFile{FS: uiFS})
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
