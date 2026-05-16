package api

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-rag/config"
	"github.com/gofurry/gofurry-rag/internal/auth"
	"github.com/gofurry/gofurry-rag/internal/contentsync"
	"github.com/gofurry/gofurry-rag/internal/db"
	"github.com/gofurry/gofurry-rag/internal/ingest"
	"github.com/gofurry/gofurry-rag/internal/service"
)

type Server struct {
	cfg         config.Config
	service     *service.Service
	authService *auth.Service
	worker      *ingest.Worker
	syncManager syncManager
	chatLimiter *publicChatLimiter
}

type syncManager interface {
	Status(ctx context.Context) (contentsync.StatusResponse, error)
	Trigger(ctx context.Context, source, trigger string) error
}

func NewServer(cfg config.Config, svc *service.Service, worker *ingest.Worker, syncManager syncManager) *Server {
	limitRequests := cfg.PublicQueryRateLimitRequests
	limitWindow := cfg.PublicQueryRateLimitWindowSec
	if svc != nil {
		limits := svc.ChatStatus().Limits
		limitRequests = limits.PublicQueryRateLimitRequests
		limitWindow = limits.PublicQueryRateLimitWindowSeconds
	}
	return &Server{
		cfg:         cfg,
		service:     svc,
		authService: auth.New(cfg),
		worker:      worker,
		syncManager: syncManager,
		chatLimiter: newPublicChatLimiter(limitRequests, time.Duration(limitWindow)*time.Second),
	}
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
	protected.Get("/sync/status", s.syncStatus)
	protected.Post("/sync/run", s.syncRun)
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

func (s *Server) queryAdminState(c fiber.Ctx) bool {
	if !strings.Contains(c.Get(fiber.HeaderCookie), s.cfg.AuthCookieName+"=") {
		return false
	}
	token := strings.TrimSpace(c.Cookies(s.cfg.AuthCookieName))
	if token == "" {
		return false
	}
	claims, err := s.authService.ParseAndValidateToken(token)
	if err != nil {
		return false
	}
	c.Locals(auth.ClaimsContextKey, claims)
	return true
}

func (s *Server) preparePublicQuery(c fiber.Ctx, req *service.QueryRequest, admin bool) error {
	if admin {
		return nil
	}
	req.IncludeDetails = false
	questionRunes := utf8.RuneCountInString(strings.TrimSpace(req.Question))
	limits := s.publicChatLimits()
	if limit := limits.PublicQueryMaxQuestionRunes; limit > 0 && questionRunes > limit {
		return publicQueryError{status: fiber.StatusBadRequest, message: "question exceeds the public maximum length"}
	}
	if req.TopK <= 0 {
		req.TopK = minPositive(s.cfg.TopK, limits.PublicQueryMaxTopK)
	}
	if maxTopK := limits.PublicQueryMaxTopK; maxTopK > 0 && req.TopK > maxTopK {
		return publicQueryError{status: fiber.StatusBadRequest, message: "top_k exceeds the public maximum limit"}
	}
	if s.chatLimiter != nil && !s.chatLimiter.Allow(c.IP()) {
		return publicQueryError{status: fiber.StatusTooManyRequests, message: "too many public chat requests"}
	}
	return nil
}

type publicQueryError struct {
	status  int
	message string
}

func (e publicQueryError) Error() string {
	return e.message
}

func (e publicQueryError) HTTPStatus() int {
	return e.status
}

func (s *Server) publicChatLimits() service.ChatLimits {
	limits := service.ChatLimits{
		PublicQueryMaxQuestionRunes:       s.cfg.PublicQueryMaxQuestionRunes,
		PublicQueryMaxTopK:                s.cfg.PublicQueryMaxTopK,
		PublicQueryRateLimitRequests:      s.cfg.PublicQueryRateLimitRequests,
		PublicQueryRateLimitWindowSeconds: s.cfg.PublicQueryRateLimitWindowSec,
	}
	if s.service != nil {
		serviceLimits := s.service.ChatStatus().Limits
		if limits.PublicQueryMaxQuestionRunes <= 0 {
			limits.PublicQueryMaxQuestionRunes = serviceLimits.PublicQueryMaxQuestionRunes
		}
		if limits.PublicQueryMaxTopK <= 0 {
			limits.PublicQueryMaxTopK = serviceLimits.PublicQueryMaxTopK
		}
		if limits.PublicQueryRateLimitRequests <= 0 {
			limits.PublicQueryRateLimitRequests = serviceLimits.PublicQueryRateLimitRequests
		}
		if limits.PublicQueryRateLimitWindowSeconds <= 0 {
			limits.PublicQueryRateLimitWindowSeconds = serviceLimits.PublicQueryRateLimitWindowSeconds
		}
	}
	return limits
}

func minPositive(left, right int) int {
	switch {
	case left <= 0 && right <= 0:
		return 0
	case left <= 0:
		return right
	case right <= 0:
		return left
	case left < right:
		return left
	default:
		return right
	}
}

type publicChatLimiter struct {
	mu       sync.Mutex
	max      int
	window   time.Duration
	requests map[string]publicChatWindow
}

type publicChatWindow struct {
	resetAt time.Time
	count   int
}

func newPublicChatLimiter(max int, window time.Duration) *publicChatLimiter {
	if max <= 0 {
		max = 10
	}
	if window <= 0 {
		window = time.Minute
	}
	return &publicChatLimiter{max: max, window: window, requests: make(map[string]publicChatWindow)}
}

func (l *publicChatLimiter) Allow(key string) bool {
	if l == nil {
		return true
	}
	key = strings.TrimSpace(key)
	if key == "" {
		key = "unknown"
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	window := l.requests[key]
	if window.resetAt.IsZero() || !now.Before(window.resetAt) {
		l.requests[key] = publicChatWindow{resetAt: now.Add(l.window), count: 1}
		return true
	}
	if window.count >= l.max {
		return false
	}
	window.count++
	l.requests[key] = window
	return true
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

func (s *Server) syncStatus(c fiber.Ctx) error {
	if s.syncManager == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}
	result, err := s.syncManager.Status(requestContext(c))
	if err != nil {
		return fail(c, err)
	}
	return ok(c, result)
}

type syncRunRequest struct {
	Source string `json:"source"`
}

func (s *Server) syncRun(c fiber.Ctx) error {
	if s.syncManager == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}
	var req syncRunRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fail(c, err)
	}
	if err := s.syncManager.Trigger(requestContext(c), req.Source, contentsync.TriggerManual); err != nil {
		return fail(c, err)
	}
	return c.Status(fiber.StatusAccepted).JSON(Result{
		Code:    1,
		Message: "accepted",
		Data: fiber.Map{
			"accepted": true,
			"source":   strings.TrimSpace(req.Source),
		},
	})
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
	admin := s.queryAdminState(c)
	if req.IncludeDetails && !admin {
		return fail(c, auth.ErrNotLoggedIn)
	}
	if err := s.preparePublicQuery(c, &req, admin); err != nil {
		return fail(c, err)
	}
	result, err := s.service.Query(requestContext(c), req)
	if err != nil {
		return fail(c, err)
	}
	if !admin {
		return ok(c, newPublicQueryResponse(result))
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
