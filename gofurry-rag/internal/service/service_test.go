package service

import (
	"context"
	"errors"
	"testing"

	"github.com/GoFurry/gofurry-rag/config"
	"github.com/GoFurry/gofurry-rag/internal/db"
	"github.com/GoFurry/gofurry-rag/internal/ingest"
)

func TestUpdateChunkUsesEmbeddingInputTemplate(t *testing.T) {
	repo := &serviceRepo{doc: db.Document{ID: 1, Title: "GoFurry", SourceType: "site", SourceID: "about"}}
	embedder := &serviceEmbedder{}
	svc := New(repo, embedder, config.Config{}, nil)

	chunk, err := svc.UpdateChunk(context.Background(), 7, UpdateChunkRequest{Content: "更新后的内容"})
	if err != nil {
		t.Fatal(err)
	}
	if chunk.Content != "更新后的内容" {
		t.Fatalf("chunk = %+v", chunk)
	}
	if len(embedder.inputs) != 1 || embedder.inputs[0] == "更新后的内容" {
		t.Fatalf("embedding input was not templated: %#v", embedder.inputs)
	}
	if want := "Title: GoFurry"; !contains(embedder.inputs[0], want) {
		t.Fatalf("missing %q in %q", want, embedder.inputs[0])
	}
}

func TestQueryRejectsLongQuestion(t *testing.T) {
	svc := New(&serviceRepo{}, &serviceEmbedder{}, config.Config{MaxQueryQuestionRunes: 3}, nil)
	_, err := svc.Query(context.Background(), QueryRequest{Question: "hello"})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v", err)
	}
}

func TestQueryRejectsTooLargeTopK(t *testing.T) {
	svc := New(&serviceRepo{}, &serviceEmbedder{}, config.Config{TopK: 3, MaxQueryTopK: 2}, nil)
	_, err := svc.Query(context.Background(), QueryRequest{Question: "hello", TopK: 3})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v", err)
	}
}

func TestOverviewIncludesWorkerSnapshot(t *testing.T) {
	repo := &serviceRepo{}
	worker := &fakeWorkerStatusProvider{status: ingest.WorkerStatus{
		State:             "processing",
		ActiveWorkers:     2,
		TotalProcessed:    11,
		TotalFailed:       3,
		LastDurationMs:    1200,
		AverageDurationMs: 890.5,
		RecentError:       "timeout",
	}}
	svc := New(repo, &serviceEmbedder{}, config.Config{}, worker)
	overview, err := svc.Overview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if overview.WorkerState != "processing" || overview.WorkerActiveWorkers != 2 || overview.WorkerTotalProcessed != 11 {
		t.Fatalf("overview = %+v", overview)
	}
}

type serviceRepo struct {
	doc db.Document
}

func (r *serviceRepo) Ping(ctx context.Context) error { return nil }

func (r *serviceRepo) CreateDocument(ctx context.Context, params db.CreateDocumentParams) (db.Document, error) {
	return db.Document{}, nil
}

func (r *serviceRepo) GetDocument(ctx context.Context, id int64) (db.Document, error) {
	return r.doc, nil
}

func (r *serviceRepo) GetDocumentByChunkID(ctx context.Context, chunkID int64) (db.Document, error) {
	return r.doc, nil
}

func (r *serviceRepo) ListDocuments(ctx context.Context, filter db.ListDocumentsFilter) (db.PageResult[db.Document], error) {
	return db.PageResult[db.Document]{}, nil
}

func (r *serviceRepo) ListChunks(ctx context.Context, documentID int64, page, pageSize int) (db.PageResult[db.Chunk], error) {
	return db.PageResult[db.Chunk]{}, nil
}

func (r *serviceRepo) ReindexDocument(ctx context.Context, id int64) (db.Document, error) {
	return db.Document{}, nil
}

func (r *serviceRepo) BatchReindexDocuments(ctx context.Context, filter db.BatchDocumentFilter) (db.BatchResult, error) {
	return db.BatchResult{}, nil
}

func (r *serviceRepo) RetryFailedDocuments(ctx context.Context, filter db.BatchDocumentFilter) (db.BatchResult, error) {
	return db.BatchResult{}, nil
}

func (r *serviceRepo) UpdateChunkContent(ctx context.Context, id int64, content, contentHash string, tokenCount int, embedding []float64) (db.Chunk, error) {
	return db.Chunk{ID: id, DocumentID: r.doc.ID, Content: content, TokenCount: tokenCount, HasEmbedding: len(embedding) > 0, EmbeddingDim: len(embedding)}, nil
}

func (r *serviceRepo) DeleteChunk(ctx context.Context, id int64) error { return nil }

func (r *serviceRepo) DeleteDocument(ctx context.Context, id int64) error { return nil }

func (r *serviceRepo) Overview(ctx context.Context) (db.Overview, error) { return db.Overview{}, nil }

func (r *serviceRepo) SearchChunks(ctx context.Context, embedding []float64, topK int, filter db.BatchDocumentFilter) ([]db.Source, error) {
	return nil, nil
}

type fakeWorkerStatusProvider struct {
	status ingest.WorkerStatus
}

func (f *fakeWorkerStatusProvider) Status() ingest.WorkerStatus {
	return f.status
}

type serviceEmbedder struct {
	inputs []string
}

func (e *serviceEmbedder) Embed(ctx context.Context, input []string) ([][]float64, error) {
	e.inputs = append(e.inputs, input...)
	return [][]float64{{0.1, 0.2}}, nil
}

func (e *serviceEmbedder) Health(ctx context.Context) error { return nil }

func (e *serviceEmbedder) Model() string { return "fake" }

func contains(text, needle string) bool {
	for i := 0; i+len(needle) <= len(text); i++ {
		if text[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
