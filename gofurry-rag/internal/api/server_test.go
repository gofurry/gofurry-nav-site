package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofurry/gofurry-rag/config"
	"github.com/gofurry/gofurry-rag/internal/db"
	"github.com/gofurry/gofurry-rag/internal/service"
	"github.com/gofurry/gofurry-rag/internal/tencentmaas"
	"github.com/gofiber/fiber/v3"
)

func TestAdminRoutesRequireToken(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", bytes.NewBufferString(`{"password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestLoginSetsCookieAndMeWorks(t *testing.T) {
	app := testApp()
	cookie := loginCookie(t, app)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/auth/me", nil)
	req.AddCookie(cookie)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestCreateTextDocument(t *testing.T) {
	app := testApp()
	cookie := loginCookie(t, app)
	reqBody := bytes.NewBufferString(`{"title":"T","content":"hello","source_type":"manual"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents/text", reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var result Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Code != 1 {
		t.Fatalf("result = %+v", result)
	}
}

func TestOverviewRequiresCookie(t *testing.T) {
	app := testApp()
	cookie := loginCookie(t, app)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/overview", nil)
	req.AddCookie(cookie)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestHealthRequiresCookie(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestHealthWithCookieWorks(t *testing.T) {
	app := testApp()
	cookie := loginCookie(t, app)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.AddCookie(cookie)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestChatStatusIsPublic(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/status", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestReindexDocumentRequiresCookie(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents/1/reindex", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestReindexDocumentSetsPending(t *testing.T) {
	app := testApp()
	cookie := loginCookie(t, app)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents/text", bytes.NewBufferString(`{"title":"T","content":"hello","source_type":"manual"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(cookie)
	if resp, err := app.Test(createReq); err != nil {
		t.Fatal(err)
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("create status = %d", resp.StatusCode)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents/1/reindex", nil)
	req.AddCookie(cookie)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var result struct {
		Code int `json:"code"`
		Data struct {
			DocumentID int64  `json:"document_id"`
			Status     string `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Code != 1 || result.Data.DocumentID != 1 || result.Data.Status != db.StatusPending {
		t.Fatalf("result = %+v", result)
	}
}

func TestBatchReindexDocuments(t *testing.T) {
	app, repo := testAppWithRepo()
	cookie := loginCookie(t, app)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents/reindex", bytes.NewBufferString(`{
		"scope":"filters",
		"filters":{
			"source_type":["website"],
			"category":["faq"],
			"language":["zh-CN"],
			"status":["ready"]
		}
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if repo.lastBatchMode != "reindex" {
		t.Fatalf("batch mode = %q", repo.lastBatchMode)
	}
	if len(repo.lastBatchFilter.SourceTypes) != 1 || repo.lastBatchFilter.SourceTypes[0] != "website" {
		t.Fatalf("source types = %+v", repo.lastBatchFilter.SourceTypes)
	}
	if len(repo.lastBatchFilter.Categories) != 1 || repo.lastBatchFilter.Categories[0] != "faq" {
		t.Fatalf("categories = %+v", repo.lastBatchFilter.Categories)
	}
	if len(repo.lastBatchFilter.Languages) != 1 || repo.lastBatchFilter.Languages[0] != "zh-CN" {
		t.Fatalf("languages = %+v", repo.lastBatchFilter.Languages)
	}
	if len(repo.lastBatchFilter.Statuses) != 1 || repo.lastBatchFilter.Statuses[0] != "ready" {
		t.Fatalf("statuses = %+v", repo.lastBatchFilter.Statuses)
	}
}

func TestRetryFailedDocuments(t *testing.T) {
	app, repo := testAppWithRepo()
	cookie := loginCookie(t, app)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents/retry-failed", bytes.NewBufferString(`{
		"scope":"document_ids",
		"document_ids":[1,2,0]
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if repo.lastBatchMode != "retry-failed" {
		t.Fatalf("batch mode = %q", repo.lastBatchMode)
	}
	if len(repo.lastBatchFilter.DocumentIDs) != 2 || repo.lastBatchFilter.DocumentIDs[0] != 1 || repo.lastBatchFilter.DocumentIDs[1] != 2 {
		t.Fatalf("document ids = %+v", repo.lastBatchFilter.DocumentIDs)
	}
}

func TestQueryReturnsSources(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/query", bytes.NewBufferString(`{"question":"gofurry","top_k":1}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var result struct {
		Code int                   `json:"code"`
		Data service.QueryResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Code != 1 {
		t.Fatalf("result = %+v", result)
	}
	if len(result.Data.Sources) != 1 {
		t.Fatalf("sources = %+v", result.Data.Sources)
	}
	source := result.Data.Sources[0]
	if source.SourceType != "manual" || source.SourceID != "about" || source.ChunkIndex != 2 || source.TokenCount != 6 {
		t.Fatalf("source debug fields = %+v", source)
	}
}

func TestQueryIncludeDetailsRequiresAdmin(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/query", bytes.NewBufferString(`{"question":"gofurry","top_k":1,"include_details":true}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestQueryIncludeDetailsWithAdminReturnsCitations(t *testing.T) {
	app := testApp()
	cookie := loginCookie(t, app)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents/text", bytes.NewBufferString(`{"title":"T","content":"hello","source_type":"manual"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(cookie)
	if resp, err := app.Test(createReq); err != nil {
		t.Fatal(err)
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("create status = %d", resp.StatusCode)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/query", bytes.NewBufferString(`{"question":"gofurry","top_k":1,"include_details":true}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var result struct {
		Code int                   `json:"code"`
		Data service.QueryResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Code != 1 || len(result.Data.Citations) == 0 {
		t.Fatalf("result = %+v", result)
	}
}

func TestChatStreamReturnsSSE(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/stream", bytes.NewBufferString(`{"question":"gofurry","top_k":1}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, needle := range []string{"event: status", "event: sources", "event: delta", "event: done"} {
		if !strings.Contains(text, needle) {
			t.Fatalf("missing %q in stream: %s", needle, text)
		}
	}
}

func TestChatStreamIncludeDetailsRequiresAdmin(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/stream", bytes.NewBufferString(`{"question":"gofurry","top_k":1,"include_details":true}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestQueryPassesFilters(t *testing.T) {
	app, repo := testAppWithRepo()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/query", bytes.NewBufferString(`{
		"question":"gofurry",
		"top_k":2,
		"filters":{
			"source_type":["site"],
			"document_ids":[3,4],
			"category":["intro"],
			"language":["zh-CN"]
		}
	}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if len(repo.lastSearchFilter.SourceTypes) != 1 || repo.lastSearchFilter.SourceTypes[0] != "site" {
		t.Fatalf("source types = %+v", repo.lastSearchFilter.SourceTypes)
	}
	if len(repo.lastSearchFilter.DocumentIDs) != 2 || repo.lastSearchFilter.DocumentIDs[0] != 3 || repo.lastSearchFilter.DocumentIDs[1] != 4 {
		t.Fatalf("document ids = %+v", repo.lastSearchFilter.DocumentIDs)
	}
	if len(repo.lastSearchFilter.Categories) != 1 || repo.lastSearchFilter.Categories[0] != "intro" {
		t.Fatalf("categories = %+v", repo.lastSearchFilter.Categories)
	}
	if len(repo.lastSearchFilter.Languages) != 1 || repo.lastSearchFilter.Languages[0] != "zh-CN" {
		t.Fatalf("languages = %+v", repo.lastSearchFilter.Languages)
	}
}

func TestListDocumentsPassesFilters(t *testing.T) {
	app, repo := testAppWithRepo()
	cookie := loginCookie(t, app)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents?page=2&page_size=6&status=failed&source_type=website,nav&category=faq&language=zh-CN&keyword=about", nil)
	req.AddCookie(cookie)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if repo.lastListFilter.Page != 2 || repo.lastListFilter.PageSize != 6 || repo.lastListFilter.Status != "failed" {
		t.Fatalf("list filter = %+v", repo.lastListFilter)
	}
	if len(repo.lastListFilter.SourceTypes) != 2 || repo.lastListFilter.SourceTypes[0] != "website" || repo.lastListFilter.SourceTypes[1] != "nav" {
		t.Fatalf("source types = %+v", repo.lastListFilter.SourceTypes)
	}
	if repo.lastListFilter.Category != "faq" || repo.lastListFilter.Language != "zh-CN" || repo.lastListFilter.Keyword != "about" {
		t.Fatalf("list filter = %+v", repo.lastListFilter)
	}
}

func TestChunkPreviewRequiresCookie(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/debug/chunk-preview", bytes.NewBufferString(`{"text":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestChunkPreviewWithText(t *testing.T) {
	app := testApp()
	cookie := loginCookie(t, app)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/debug/chunk-preview", bytes.NewBufferString(`{"text":"`+strings.Repeat("猫", 25)+`","variants":[{"chunk_size":10,"chunk_overlap":3}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var result struct {
		Code int                          `json:"code"`
		Data service.ChunkPreviewResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Code != 1 || result.Data.Source != "text" || len(result.Data.Variants) != 1 {
		t.Fatalf("result = %+v", result)
	}
	if result.Data.Variants[0].ChunkCount != 4 || len(result.Data.Variants[0].Chunks) != 4 {
		t.Fatalf("variant = %+v", result.Data.Variants[0])
	}
}

func TestChunkPreviewWithDocument(t *testing.T) {
	app := testApp()
	cookie := loginCookie(t, app)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents/text", bytes.NewBufferString(`{"title":"T","content":"hello world","source_type":"manual"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(cookie)
	if resp, err := app.Test(createReq); err != nil {
		t.Fatal(err)
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("create status = %d", resp.StatusCode)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/debug/chunk-preview", bytes.NewBufferString(`{"document_id":1}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var result struct {
		Code int                          `json:"code"`
		Data service.ChunkPreviewResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Code != 1 || result.Data.Source != "document" || result.Data.Title != "T" || len(result.Data.Variants) != 3 {
		t.Fatalf("result = %+v", result)
	}
}

func TestChunkPreviewRejectsInvalidVariant(t *testing.T) {
	app := testApp()
	cookie := loginCookie(t, app)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/debug/chunk-preview", bytes.NewBufferString(`{"text":"hello","variants":[{"chunk_size":10,"chunk_overlap":10}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestUpdateAndDeleteChunkRequireCookie(t *testing.T) {
	app := testApp()
	cookie := loginCookie(t, app)

	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/chunks/1", bytes.NewBufferString(`{"content":"updated chunk"}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.AddCookie(cookie)
	updateResp, err := app.Test(updateReq)
	if err != nil {
		t.Fatal(err)
	}
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("update status = %d", updateResp.StatusCode)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/chunks/1", nil)
	deleteReq.AddCookie(cookie)
	deleteResp, err := app.Test(deleteReq)
	if err != nil {
		t.Fatal(err)
	}
	if deleteResp.StatusCode != http.StatusOK {
		t.Fatalf("delete status = %d", deleteResp.StatusCode)
	}
}

func TestLogoutClearsCookie(t *testing.T) {
	app := testApp()
	cookie := loginCookie(t, app)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/logout", nil)
	req.AddCookie(cookie)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if len(resp.Cookies()) == 0 || resp.Cookies()[0].MaxAge != -1 {
		t.Fatalf("logout cookie not expired: %#v", resp.Cookies())
	}
}

func testApp() *fiber.App {
	app, _ := testAppWithRepo()
	return app
}

func testAppWithRepo() (*fiber.App, *fakeRepo) {
	cfg := config.Config{
		AppName:         "test",
		AdminToken:      "test-token",
		ConsolePasscode: "test-token",
		JWTSecret:       "jwt-secret",
		AuthCookieName:  "gofurry_rag_session",
		SessionTTLHours: 1,
		TopK:            6,
		Auth: config.AuthConfig{
			CookieName:       "gofurry_rag_session",
			CookieMaxAgeSecs: 3600,
			SessionTTLHours:  1,
			SameSite:         "Lax",
		},
	}
	repo := newFakeRepo()
	svc := service.New(repo, fakeEmbedder{}, &fakeChat{configured: true, model: "deepseek-v4-flash", answer: "gofurry is a content discovery website."}, cfg, nil)
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	NewServer(cfg, svc, nil).RegisterRoutes(app.Group("/api/v1"))
	return app, repo
}

func loginCookie(t *testing.T, app *fiber.App) *http.Cookie {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", bytes.NewBufferString(`{"password":"test-token"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login status = %d", resp.StatusCode)
	}
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Fatal("login did not set cookie")
	}
	return cookies[0]
}

type fakeRepo struct {
	mu               sync.Mutex
	next             int64
	docs             []db.Document
	lastListFilter   db.ListDocumentsFilter
	lastSearchFilter db.BatchDocumentFilter
	lastBatchFilter  db.BatchDocumentFilter
	lastBatchMode    string
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{next: 1}
}

func (r *fakeRepo) Ping(ctx context.Context) error {
	return nil
}

func (r *fakeRepo) CreateDocument(ctx context.Context, params db.CreateDocumentParams) (db.Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	doc := db.Document{
		ID:         r.next,
		Title:      params.Title,
		SourceType: params.SourceType,
		SourceID:   params.SourceID,
		URL:        params.URL,
		Content:    params.Content,
		Status:     db.StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	r.next++
	r.docs = append(r.docs, doc)
	return doc, nil
}

func (r *fakeRepo) GetDocument(ctx context.Context, id int64) (db.Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, doc := range r.docs {
		if doc.ID == id {
			return doc, nil
		}
	}
	return db.Document{}, service.ErrValidation
}

func (r *fakeRepo) GetDocumentByChunkID(ctx context.Context, chunkID int64) (db.Document, error) {
	return db.Document{ID: 1, Title: "gofurry", SourceType: "manual", SourceID: "about"}, nil
}

func (r *fakeRepo) ListDocuments(ctx context.Context, filter db.ListDocumentsFilter) (db.PageResult[db.Document], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastListFilter = filter
	items := append([]db.Document(nil), r.docs...)
	return db.PageResult[db.Document]{Items: items, Total: int64(len(items))}, nil
}

func (r *fakeRepo) ListChunks(ctx context.Context, documentID int64, page, pageSize int) (db.PageResult[db.Chunk], error) {
	return db.PageResult[db.Chunk]{Items: []db.Chunk{}, Total: 0}, nil
}

func (r *fakeRepo) ReindexDocument(ctx context.Context, id int64) (db.Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.docs {
		if r.docs[i].ID == id {
			r.docs[i].Status = db.StatusPending
			r.docs[i].ErrorMessage = ""
			r.docs[i].UpdatedAt = time.Now()
			return r.docs[i], nil
		}
	}
	return db.Document{}, service.ErrValidation
}

func (r *fakeRepo) BatchReindexDocuments(ctx context.Context, filter db.BatchDocumentFilter) (db.BatchResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastBatchFilter = filter
	r.lastBatchMode = "reindex"
	return db.BatchResult{AcceptedCount: 2, SkippedCount: 1, Status: db.StatusPending}, nil
}

func (r *fakeRepo) RetryFailedDocuments(ctx context.Context, filter db.BatchDocumentFilter) (db.BatchResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastBatchFilter = filter
	r.lastBatchMode = "retry-failed"
	return db.BatchResult{AcceptedCount: 1, SkippedCount: 0, Status: db.StatusPending}, nil
}

func (r *fakeRepo) UpdateChunkContent(ctx context.Context, id int64, content, contentHash string, tokenCount int, embedding []float64) (db.Chunk, error) {
	return db.Chunk{ID: id, DocumentID: 1, Content: content, ContentHash: contentHash, TokenCount: tokenCount, HasEmbedding: len(embedding) > 0, EmbeddingDim: len(embedding)}, nil
}

func (r *fakeRepo) DeleteChunk(ctx context.Context, id int64) error {
	return nil
}

func (r *fakeRepo) DeleteDocument(ctx context.Context, id int64) error {
	return nil
}

func (r *fakeRepo) Overview(ctx context.Context) (db.Overview, error) {
	return db.Overview{DocumentTotal: int64(len(r.docs)), ChunkTotal: 2, EmbeddedChunkTotal: 2}, nil
}

func (r *fakeRepo) SearchChunks(ctx context.Context, embedding []float64, topK int, filter db.BatchDocumentFilter) ([]db.Source, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastSearchFilter = filter
	return []db.Source{{DocumentID: 1, ChunkID: 1, SourceType: "manual", SourceID: "about", Title: "gofurry", ChunkIndex: 2, TokenCount: 6, Score: 0.9, Content: "source"}}, nil
}

type fakeEmbedder struct{}

func (fakeEmbedder) Embed(ctx context.Context, input []string) ([][]float64, error) {
	return [][]float64{{0.1, 0.2}}, nil
}

func (fakeEmbedder) Health(ctx context.Context) error {
	return nil
}

func (fakeEmbedder) Model() string {
	return "fake"
}

type fakeChat struct {
	configured    bool
	model         string
	answer        string
	completeCalls int
	streamCalls   int
}

func (f *fakeChat) Model() string {
	if f.model != "" {
		return f.model
	}
	return "fake-chat"
}

func (f *fakeChat) Configured() bool {
	return f != nil && f.configured
}

func (f *fakeChat) Health(ctx context.Context) error {
	return nil
}

func (f *fakeChat) Complete(ctx context.Context, _ []tencentmaas.Message) (tencentmaas.CompletionResult, error) {
	f.completeCalls++
	return tencentmaas.CompletionResult{
		Model:            f.Model(),
		Answer:           f.answer,
		PromptTokens:     12,
		CompletionTokens: 34,
		TotalTokens:      46,
		CachedTokens:     2,
		ReasoningTokens:  8,
	}, nil
}

func (f *fakeChat) Stream(ctx context.Context, _ []tencentmaas.Message, onDelta func(string) error) (tencentmaas.CompletionResult, error) {
	f.streamCalls++
	for _, piece := range []string{"gofurry", " is", " a site."} {
		if onDelta != nil {
			if err := onDelta(piece); err != nil {
				return tencentmaas.CompletionResult{}, err
			}
		}
	}
	return tencentmaas.CompletionResult{
		Model:            f.Model(),
		Answer:           f.answer,
		PromptTokens:     12,
		CompletionTokens: 34,
		TotalTokens:      46,
		CachedTokens:     2,
		ReasoningTokens:  8,
	}, nil
}
