package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/GoFurry/gofurry-rag/config"
	"github.com/GoFurry/gofurry-rag/internal/db"
	"github.com/GoFurry/gofurry-rag/internal/service"
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

func TestQueryReturnsSources(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/query", bytes.NewBufferString(`{"question":"GoFurry","top_k":1}`))
	req.Header.Set("Content-Type", "application/json")
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
	svc := service.New(newFakeRepo(), fakeEmbedder{}, cfg)
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	NewServer(cfg, svc, nil).RegisterRoutes(app.Group("/api/v1"))
	return app
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
	mu   sync.Mutex
	next int64
	docs []db.Document
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
		Status:     db.StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	r.next++
	r.docs = append(r.docs, doc)
	return doc, nil
}

func (r *fakeRepo) ListDocuments(ctx context.Context, filter db.ListDocumentsFilter) (db.PageResult[db.Document], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := append([]db.Document(nil), r.docs...)
	return db.PageResult[db.Document]{Items: items, Total: int64(len(items))}, nil
}

func (r *fakeRepo) ListChunks(ctx context.Context, documentID int64, page, pageSize int) (db.PageResult[db.Chunk], error) {
	return db.PageResult[db.Chunk]{Items: []db.Chunk{}, Total: 0}, nil
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

func (r *fakeRepo) SearchChunks(ctx context.Context, embedding []float64, topK int) ([]db.Source, error) {
	return []db.Source{{DocumentID: 1, ChunkID: 1, Title: "GoFurry", Score: 0.9, Content: "source"}}, nil
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
