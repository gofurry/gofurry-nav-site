package api

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/GoFurry/gofurry-rag/config"
	"github.com/GoFurry/gofurry-rag/internal/auth"
	"github.com/GoFurry/gofurry-rag/internal/db"
	"github.com/GoFurry/gofurry-rag/internal/ingest"
	"github.com/GoFurry/gofurry-rag/internal/service"
	"github.com/gofiber/fiber/v3"
)

type Server struct {
	cfg         config.Config
	service     *service.Service
	authService *auth.Service
	worker      *ingest.Worker
}

func NewServer(cfg config.Config, svc *service.Service, worker *ingest.Worker) *Server {
	return &Server{cfg: cfg, service: svc, authService: auth.New(cfg), worker: worker}
}

func (s *Server) RegisterRoutes(v1 fiber.Router) {
	v1.Get("/health", s.health)
	admin := v1.Group("/admin")
	admin.Get("/auth/state", s.authState)
	admin.Post("/auth/login", s.authLogin)
	admin.Post("/auth/logout", s.authLogout)
	protected := admin.Group("", s.requireAdmin)
	protected.Get("/auth/me", s.authMe)
	protected.Get("/overview", s.overview)
	protected.Post("/documents/text", s.createTextDocument)
	protected.Get("/documents", s.listDocuments)
	protected.Get("/documents/:id/chunks", s.listChunks)
	protected.Post("/documents/:id/reindex", s.reindexDocument)
	protected.Delete("/documents/:id", s.deleteDocument)
	protected.Patch("/chunks/:id", s.updateChunk)
	protected.Delete("/chunks/:id", s.deleteChunk)
	protected.Post("/debug/chunk-preview", s.chunkPreview)
	v1.Post("/chat/query", s.query)
}

func (s *Server) requireAdmin(c fiber.Ctx) error {
	token := strings.TrimSpace(c.Cookies(s.cfg.AuthCookieName))
	claims, err := s.authService.ParseAndValidateToken(token)
	if err != nil {
		return fail(c, err)
	}
	c.Locals(auth.ClaimsContextKey, claims)
	return c.Next()
}

type passwordRequest struct {
	Password string `json:"password"`
}

func (s *Server) authState(c fiber.Ctx) error {
	authenticated := false
	token := strings.TrimSpace(c.Cookies(s.cfg.AuthCookieName))
	if token != "" {
		if claims, err := s.authService.ParseAndValidateToken(token); err == nil {
			authenticated = true
			c.Locals(auth.ClaimsContextKey, claims)
		}
	}
	return ok(c, fiber.Map{"initialized": true, "authenticated": authenticated})
}

func (s *Server) authLogin(c fiber.Ctx) error {
	var req passwordRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fail(c, err)
	}
	token, claims, err := s.authService.Login(req.Password)
	if err != nil {
		return fail(c, err)
	}
	c.Cookie(s.authService.BuildAuthCookie(token))
	return ok(c, fiber.Map{
		"initialized":     true,
		"authenticated":   true,
		"session_version": claims.SessionVersion,
	})
}

func (s *Server) authLogout(c fiber.Ctx) error {
	c.Cookie(s.authService.BuildLogoutCookie())
	return ok(c, fiber.Map{"authenticated": false})
}

func (s *Server) authMe(c fiber.Ctx) error {
	claims, _ := c.Locals(auth.ClaimsContextKey).(*auth.Claims)
	version := int64(0)
	if claims != nil {
		version = claims.SessionVersion
	}
	return ok(c, fiber.Map{"initialized": true, "authenticated": true, "session_version": version})
}

func (s *Server) health(c fiber.Ctx) error {
	if s.service == nil {
		return ok(c, fiber.Map{"status": "not_ready"})
	}
	ctx := context.Background()
	return ok(c, s.service.Health(ctx))
}

func (s *Server) overview(c fiber.Ctx) error {
	if s.service == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}
	result, err := s.service.Overview(context.Background())
	if err != nil {
		return fail(c, err)
	}
	return ok(c, result)
}

func (s *Server) createTextDocument(c fiber.Ctx) error {
	if s.service == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}
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
	if s.service == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}
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
	if s.service == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}
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
	if s.service == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fail(c, service.ErrValidation)
	}
	if err := s.service.DeleteDocument(context.Background(), id); err != nil {
		return fail(c, err)
	}
	return ok(c, fiber.Map{"deleted": true})
}

func (s *Server) reindexDocument(c fiber.Ctx) error {
	if s.service == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fail(c, service.ErrValidation)
	}
	doc, err := s.service.ReindexDocument(context.Background(), id)
	if err != nil {
		return fail(c, err)
	}
	return ok(c, fiber.Map{"document_id": doc.ID, "status": doc.Status})
}

func (s *Server) updateChunk(c fiber.Ctx) error {
	if s.service == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fail(c, service.ErrValidation)
	}
	var req service.UpdateChunkRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fail(c, err)
	}
	chunk, err := s.service.UpdateChunk(context.Background(), id, req)
	if err != nil {
		return fail(c, err)
	}
	return ok(c, chunk)
}

func (s *Server) deleteChunk(c fiber.Ctx) error {
	if s.service == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fail(c, service.ErrValidation)
	}
	if err := s.service.DeleteChunk(context.Background(), id); err != nil {
		return fail(c, err)
	}
	return ok(c, fiber.Map{"deleted": true})
}

func (s *Server) chunkPreview(c fiber.Ctx) error {
	if s.service == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}
	var req service.ChunkPreviewRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fail(c, err)
	}
	result, err := s.service.ChunkPreview(context.Background(), req)
	if err != nil {
		return fail(c, err)
	}
	return ok(c, result)
}

func (s *Server) query(c fiber.Ctx) error {
	if s.service == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}
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
