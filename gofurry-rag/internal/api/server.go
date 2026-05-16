package api

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gofurry/gofurry-rag/config"
	"github.com/gofurry/gofurry-rag/internal/auth"
	"github.com/gofurry/gofurry-rag/internal/db"
	"github.com/gofurry/gofurry-rag/internal/ingest"
	"github.com/gofurry/gofurry-rag/internal/service"
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
	v1.Get("/health", s.requireAdmin, s.health)
	admin := v1.Group("/admin")
	admin.Get("/auth/state", s.authState)
	admin.Post("/auth/login", s.authLogin)
	admin.Post("/auth/logout", s.authLogout)
	protected := admin.Group("", s.requireAdmin)
	protected.Get("/auth/me", s.authMe)
	protected.Get("/overview", s.overview)
	protected.Post("/documents/text", s.createTextDocument)
	protected.Get("/documents", s.listDocuments)
	protected.Post("/documents/reindex", s.batchReindexDocuments)
	protected.Post("/documents/retry-failed", s.retryFailedDocuments)
	protected.Get("/documents/:id/chunks", s.listChunks)
	protected.Post("/documents/:id/reindex", s.reindexDocument)
	protected.Delete("/documents/:id", s.deleteDocument)
	protected.Patch("/chunks/:id", s.updateChunk)
	protected.Delete("/chunks/:id", s.deleteChunk)
	protected.Post("/debug/chunk-preview", s.chunkPreview)
	v1.Get("/chat/status", s.chatStatus)
	v1.Post("/chat/query", s.query)
	v1.Post("/chat/stream", s.chatStream)
}

func requestContext(c fiber.Ctx) context.Context {
	ctx := c.Context()
	if ctx == nil {
		return context.Background()
	}
	return ctx
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

func (s *Server) requireDetailedQueryAdmin(c fiber.Ctx, includeDetails bool) error {
	if !includeDetails {
		return nil
	}
	token := strings.TrimSpace(c.Cookies(s.cfg.AuthCookieName))
	claims, err := s.authService.ParseAndValidateToken(token)
	if err != nil {
		return err
	}
	c.Locals(auth.ClaimsContextKey, claims)
	return nil
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
	return ok(c, s.service.Health(requestContext(c)))
}

func (s *Server) chatStatus(c fiber.Ctx) error {
	if s.service == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}
	return ok(c, s.service.ChatStatus())
}

func (s *Server) overview(c fiber.Ctx) error {
	if s.service == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}
	result, err := s.service.Overview(requestContext(c))
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
	doc, err := s.service.CreateTextDocument(requestContext(c), req)
	if err != nil {
		return fail(c, err)
	}
	return ok(c, fiber.Map{"document_id": doc.ID, "status": doc.Status})
}

func (s *Server) listDocuments(c fiber.Ctx) error {
	if s.service == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}
	result, err := s.service.ListDocuments(requestContext(c), db.ListDocumentsFilter{
		Page:        queryInt(c, "page", 1),
		PageSize:    queryInt(c, "page_size", 20),
		Status:      c.Query("status"),
		SourceTypes: queryCSV(c, "source_type"),
		Category:    c.Query("category"),
		Language:    c.Query("language"),
		Keyword:     c.Query("keyword"),
	})
	if err != nil {
		return fail(c, err)
	}
	return ok(c, result)
}

func (s *Server) batchReindexDocuments(c fiber.Ctx) error {
	if s.service == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}
	var req service.BatchDocumentsRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fail(c, err)
	}
	result, err := s.service.BatchReindexDocuments(requestContext(c), req)
	if err != nil {
		return fail(c, err)
	}
	return ok(c, result)
}

func (s *Server) retryFailedDocuments(c fiber.Ctx) error {
	if s.service == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}
	var req service.BatchDocumentsRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fail(c, err)
	}
	result, err := s.service.RetryFailedDocuments(requestContext(c), req)
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
	result, err := s.service.ListChunks(requestContext(c), id, queryInt(c, "page", 1), queryInt(c, "page_size", 20))
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
	if err := s.service.DeleteDocument(requestContext(c), id); err != nil {
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
	doc, err := s.service.ReindexDocument(requestContext(c), id)
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
	chunk, err := s.service.UpdateChunk(requestContext(c), id, req)
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
	if err := s.service.DeleteChunk(requestContext(c), id); err != nil {
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
	result, err := s.service.ChunkPreview(requestContext(c), req)
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
	if err := s.requireDetailedQueryAdmin(c, req.IncludeDetails); err != nil {
		return fail(c, err)
	}
	result, err := s.service.Query(requestContext(c), req)
	if err != nil {
		return fail(c, err)
	}
	return ok(c, result)
}

func queryCSV(c fiber.Ctx, key string) []string {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func queryInt(c fiber.Ctx, key string, fallback int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
