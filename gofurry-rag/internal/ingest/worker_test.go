package ingest

import (
	"context"
	"strings"
	"testing"

	"github.com/GoFurry/gofurry-rag/config"
	"github.com/GoFurry/gofurry-rag/internal/db"
)

func TestWorkerUsesEmbeddingInputTemplate(t *testing.T) {
	repo := &workerRepo{
		doc: &db.Document{
			ID:         1,
			Title:      "GoFurry",
			SourceType: "site",
			SourceID:   "about",
			URL:        "https://example.com/about",
			Content:    "第一段\n\n第二段",
		},
	}
	embedder := &recordingEmbedder{}
	worker := NewWorker(repo, embedder, config.Config{ChunkSize: 20, ChunkOverlap: 4, EmbedBatchSize: 8})

	processed, err := worker.ProcessOnce(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !processed {
		t.Fatal("expected document to be processed")
	}
	if len(embedder.inputs) != 1 {
		t.Fatalf("inputs = %#v", embedder.inputs)
	}
	input := embedder.inputs[0]
	for _, want := range []string{"Title: GoFurry", "Source Type: site", "Source ID: about", "URL: https://example.com/about", "Content:\n第一段"} {
		if !strings.Contains(input, want) {
			t.Fatalf("missing %q in %q", want, input)
		}
	}
	if len(repo.chunks) != 1 || strings.Contains(repo.chunks[0].Content, "Title:") {
		t.Fatalf("stored chunk should remain raw: %#v", repo.chunks)
	}
}

type workerRepo struct {
	doc    *db.Document
	chunks []db.NewChunk
}

func (r *workerRepo) ClaimPendingDocument(ctx context.Context) (*db.Document, error) {
	if r.doc == nil {
		return nil, nil
	}
	doc := r.doc
	r.doc = nil
	return doc, nil
}

func (r *workerRepo) ReplaceChunks(ctx context.Context, documentID int64, chunks []db.NewChunk) error {
	r.chunks = chunks
	return nil
}

func (r *workerRepo) MarkDocumentFailed(ctx context.Context, id int64, message string) error {
	return nil
}

type recordingEmbedder struct {
	inputs []string
}

func (e *recordingEmbedder) Embed(ctx context.Context, input []string) ([][]float64, error) {
	e.inputs = append(e.inputs, input...)
	embeddings := make([][]float64, len(input))
	for i := range embeddings {
		embeddings[i] = []float64{0.1, 0.2}
	}
	return embeddings, nil
}

func (e *recordingEmbedder) Health(ctx context.Context) error {
	return nil
}

func (e *recordingEmbedder) Model() string {
	return "fake"
}
