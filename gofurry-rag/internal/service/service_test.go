package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/GoFurry/gofurry-rag/config"
	"github.com/GoFurry/gofurry-rag/internal/db"
	"github.com/GoFurry/gofurry-rag/internal/ingest"
	"github.com/GoFurry/gofurry-rag/internal/tencentmaas"
)

func TestUpdateChunkUsesEmbeddingInputTemplate(t *testing.T) {
	repo := &serviceRepo{doc: db.Document{ID: 1, Title: "GoFurry", SourceType: "site", SourceID: "about"}}
	embedder := &serviceEmbedder{}
	svc := New(repo, embedder, &fakeChat{configured: false}, config.Config{}, nil)

	chunk, err := svc.UpdateChunk(context.Background(), 7, UpdateChunkRequest{Content: "updated content"})
	if err != nil {
		t.Fatal(err)
	}
	if chunk.Content != "updated content" {
		t.Fatalf("chunk = %+v", chunk)
	}
	if len(embedder.inputs) != 1 || embedder.inputs[0] == "updated content" {
		t.Fatalf("embedding input was not templated: %#v", embedder.inputs)
	}
	if want := "Title: GoFurry"; !strings.Contains(embedder.inputs[0], want) {
		t.Fatalf("missing %q in %q", want, embedder.inputs[0])
	}
}

func TestQueryRejectsLongQuestion(t *testing.T) {
	svc := New(&serviceRepo{}, &serviceEmbedder{}, &fakeChat{configured: true}, config.Config{MaxQueryQuestionRunes: 3}, nil)
	_, err := svc.Query(context.Background(), QueryRequest{Question: "hello"})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v", err)
	}
}

func TestQueryRejectsTooLargeTopK(t *testing.T) {
	svc := New(&serviceRepo{}, &serviceEmbedder{}, &fakeChat{configured: true}, config.Config{TopK: 3, MaxQueryTopK: 2}, nil)
	_, err := svc.Query(context.Background(), QueryRequest{Question: "hello", TopK: 3})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v", err)
	}
}

func TestQueryReturnsAnswerAndSources(t *testing.T) {
	repo := &serviceRepo{
		sources: []db.Source{{
			DocumentID: 1,
			ChunkID:    2,
			SourceType: "manual",
			SourceID:   "about",
			Title:      "GoFurry",
			ChunkIndex: 2,
			TokenCount: 6,
			Score:      0.91,
			Content:    "GoFurry is a content discovery website.",
		}},
	}
	chat := &fakeChat{
		configured: true,
		model:      "deepseek-v4-flash",
		answer:     "GoFurry is a content discovery website.",
	}
	svc := New(repo, &serviceEmbedder{}, chat, config.Config{TopK: 6}, nil)
	result, err := svc.Query(context.Background(), QueryRequest{Question: "What is GoFurry?"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Answer, "答案：") || !strings.Contains(result.Answer, "引用：") {
		t.Fatalf("result = %+v", result)
	}
	if !strings.Contains(result.Answer, "[1]") {
		t.Fatalf("result = %+v", result)
	}
	if result.Usage.AnswerModel != "deepseek-v4-flash" {
		t.Fatalf("result = %+v", result)
	}
	if len(result.Sources) != 1 || result.Sources[0].ChunkID != 2 {
		t.Fatalf("sources = %+v", result.Sources)
	}
	if chat.completeCalls != 1 {
		t.Fatalf("completeCalls = %d", chat.completeCalls)
	}
	if len(chat.lastMessages) != 2 || !strings.Contains(chat.lastMessages[1].Content, "资料：") {
		t.Fatalf("messages = %+v", chat.lastMessages)
	}
}

func TestQueryReturnsCitationDetails(t *testing.T) {
	repo := &serviceRepo{
		doc: db.Document{
			ID:           1,
			SourceType:   "manual",
			SourceID:     "about",
			Title:        "GoFurry",
			URL:          "https://example.com/about",
			Content:      "GoFurry is a content discovery website.",
			Status:       db.StatusReady,
			ErrorMessage: "",
			Metadata:     json.RawMessage(`{"category":"intro","language":"zh-CN"}`),
			ChunkCount:   3,
		},
		sources: []db.Source{{
			DocumentID: 1,
			ChunkID:    2,
			SourceType: "manual",
			SourceID:   "about",
			Title:      "GoFurry",
			URL:        "https://example.com/about",
			ChunkIndex: 2,
			TokenCount: 6,
			Score:      0.91,
			Content:    "GoFurry is a content discovery website.",
		}},
	}
	chat := &fakeChat{
		configured: true,
		model:      "deepseek-v4-flash",
		answer:     "GoFurry is a content discovery website.",
	}
	svc := New(repo, &serviceEmbedder{}, chat, config.Config{TopK: 6}, nil)
	result, err := svc.Query(context.Background(), QueryRequest{Question: "What is GoFurry?", IncludeDetails: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Citations) != 1 {
		t.Fatalf("citations = %+v", result.Citations)
	}
	citation := result.Citations[0]
	if !citation.UsedInPrompt || citation.Lineage.ChunkID != 2 || citation.Lineage.DocumentID != 1 {
		t.Fatalf("citation lineage = %+v", citation)
	}
	if citation.Document.Content != "GoFurry is a content discovery website." {
		t.Fatalf("document = %+v", citation.Document)
	}
	if string(citation.Document.Metadata) != `{"category":"intro","language":"zh-CN"}` {
		t.Fatalf("metadata = %s", string(citation.Document.Metadata))
	}
	if citation.Chunk.Content != "GoFurry is a content discovery website." {
		t.Fatalf("chunk = %+v", citation.Chunk)
	}
}

func TestQueryReturnsNoSourcesMessage(t *testing.T) {
	chat := &fakeChat{configured: true, model: "deepseek-v4-flash"}
	svc := New(&serviceRepo{}, &serviceEmbedder{}, chat, config.Config{TopK: 6}, nil)
	result, err := svc.Query(context.Background(), QueryRequest{Question: "What is GoFurry?"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Answer, "答案：") || !strings.Contains(result.Answer, "引用：") || !strings.Contains(result.Answer, "无") {
		t.Fatalf("answer = %q", result.Answer)
	}
	if chat.completeCalls != 0 {
		t.Fatalf("completeCalls = %d", chat.completeCalls)
	}
}

func TestBuildChatMessagesTruncatesSources(t *testing.T) {
	sources := []db.Source{
		{
			DocumentID: 1,
			ChunkID:    11,
			SourceType: "manual",
			SourceID:   "about",
			Title:      "Alpha",
			ChunkIndex: 0,
			TokenCount: 20,
			Score:      0.99,
			Content:    strings.Repeat("甲", 600),
		},
		{
			DocumentID: 2,
			ChunkID:    22,
			SourceType: "manual",
			SourceID:   "faq",
			Title:      "Beta",
			ChunkIndex: 1,
			TokenCount: 20,
			Score:      0.88,
			Content:    strings.Repeat("乙", 600),
		},
	}

	messages, usedSources := buildChatMessages("What is GoFurry?", sources, 220)
	if len(messages) != 2 {
		t.Fatalf("messages = %+v", messages)
	}
	if !strings.Contains(messages[1].Content, "问题：") || !strings.Contains(messages[1].Content, "资料：") {
		t.Fatalf("prompt = %q", messages[1].Content)
	}
	if len(usedSources) == len(sources) {
		t.Fatalf("expected truncation, usedSources = %+v", usedSources)
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
	svc := New(repo, &serviceEmbedder{}, &fakeChat{configured: false}, config.Config{}, worker)
	overview, err := svc.Overview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if overview.WorkerState != "processing" || overview.WorkerActiveWorkers != 2 || overview.WorkerTotalProcessed != 11 {
		t.Fatalf("overview = %+v", overview)
	}
}

type serviceRepo struct {
	doc     db.Document
	sources []db.Source
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
	return append([]db.Source(nil), r.sources...), nil
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

type fakeChat struct {
	configured    bool
	model         string
	answer        string
	streamPieces  []string
	lastMessages  []tencentmaas.Message
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

func (f *fakeChat) Complete(ctx context.Context, messages []tencentmaas.Message) (tencentmaas.CompletionResult, error) {
	f.completeCalls++
	f.lastMessages = append([]tencentmaas.Message(nil), messages...)
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

func (f *fakeChat) Stream(ctx context.Context, messages []tencentmaas.Message, onDelta func(string) error) (tencentmaas.CompletionResult, error) {
	f.streamCalls++
	f.lastMessages = append([]tencentmaas.Message(nil), messages...)
	pieces := f.streamPieces
	if len(pieces) == 0 {
		pieces = []string{"GoFurry", " is", " a site."}
	}
	for _, piece := range pieces {
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
