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

	"github.com/GoFurry/gofurry-rag/internal/config"
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

func TestCreateTextDocument(t *testing.T) {
	app := testApp()
	reqBody := bytes.NewBufferString(`{"title":"T","content":"hello","source_type":"manual"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents/text", reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
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

func testApp() *fiber.App {
	cfg := config.Config{
		AppName:    "test",
		AdminToken: "test-token",
		TopK:       6,
	}
	svc := service.New(newFakeRepo(), fakeEmbedder{}, cfg)
	return NewServer(cfg, svc, nil).App()
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

func (r *fakeRepo) DeleteDocument(ctx context.Context, id int64) error {
	return nil
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
