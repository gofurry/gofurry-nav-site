package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/GoFurry/gofurry-rag/config"
	"github.com/GoFurry/gofurry-rag/internal/db"
	"github.com/GoFurry/gofurry-rag/internal/embedder"
	"github.com/GoFurry/gofurry-rag/internal/ingest"
)

var ErrValidation = errors.New("validation failed")

type Repository interface {
	Ping(ctx context.Context) error
	CreateDocument(ctx context.Context, params db.CreateDocumentParams) (db.Document, error)
	GetDocument(ctx context.Context, id int64) (db.Document, error)
	GetDocumentByChunkID(ctx context.Context, chunkID int64) (db.Document, error)
	ListDocuments(ctx context.Context, filter db.ListDocumentsFilter) (db.PageResult[db.Document], error)
	ListChunks(ctx context.Context, documentID int64, page, pageSize int) (db.PageResult[db.Chunk], error)
	ReindexDocument(ctx context.Context, id int64) (db.Document, error)
	UpdateChunkContent(ctx context.Context, id int64, content, contentHash string, tokenCount int, embedding []float64) (db.Chunk, error)
	DeleteChunk(ctx context.Context, id int64) error
	DeleteDocument(ctx context.Context, id int64) error
	Overview(ctx context.Context) (db.Overview, error)
	SearchChunks(ctx context.Context, embedding []float64, topK int) ([]db.Source, error)
}

type Service struct {
	repo     Repository
	embedder embedder.Client
	cfg      config.Config
}

type TextDocumentRequest struct {
	Title      string          `json:"title"`
	Content    string          `json:"content"`
	SourceType string          `json:"source_type"`
	SourceID   string          `json:"source_id"`
	URL        string          `json:"url"`
	Metadata   json.RawMessage `json:"metadata"`
}

type QueryRequest struct {
	Question string `json:"question"`
	TopK     int    `json:"top_k"`
}

type UpdateChunkRequest struct {
	Content string `json:"content"`
}

type ChunkPreviewRequest struct {
	DocumentID int64                 `json:"document_id"`
	Text       string                `json:"text"`
	Variants   []ChunkPreviewVariant `json:"variants"`
}

type ChunkPreviewVariant struct {
	ChunkSize    int `json:"chunk_size"`
	ChunkOverlap int `json:"chunk_overlap"`
}

type ChunkPreviewResponse struct {
	Source   string                      `json:"source"`
	Title    string                      `json:"title"`
	Variants []ChunkPreviewVariantResult `json:"variants"`
}

type ChunkPreviewVariantResult struct {
	ChunkSize    int                 `json:"chunk_size"`
	ChunkOverlap int                 `json:"chunk_overlap"`
	ChunkCount   int                 `json:"chunk_count"`
	MinChars     int                 `json:"min_chars"`
	MaxChars     int                 `json:"max_chars"`
	AvgChars     float64             `json:"avg_chars"`
	Chunks       []ChunkPreviewChunk `json:"chunks"`
}

type ChunkPreviewChunk struct {
	Index     int    `json:"index"`
	CharCount int    `json:"char_count"`
	Content   string `json:"content"`
}

type QueryResponse struct {
	Answer  string      `json:"answer"`
	Sources []db.Source `json:"sources"`
	Usage   QueryUsage  `json:"usage"`
}

type QueryUsage struct {
	TopK           int    `json:"top_k"`
	EmbeddingModel string `json:"embedding_model"`
}

func New(repo Repository, embedder embedder.Client, cfg config.Config) *Service {
	return &Service{repo: repo, embedder: embedder, cfg: cfg}
}

func (s *Service) Health(ctx context.Context) map[string]any {
	result := map[string]any{
		"status":          "ok",
		"app_name":        s.cfg.AppName,
		"embedding_model": s.embedder.Model(),
		"database": map[string]any{
			"type":      s.cfg.Database.DBType,
			"name":      s.cfg.Database.Postgres.DBName,
			"host":      s.cfg.Database.Postgres.DBHost,
			"port":      s.cfg.Database.Postgres.DBPort,
			"connected": true,
		},
		"ollama": map[string]any{
			"base_url":  s.cfg.OllamaBaseURL,
			"model":     s.embedder.Model(),
			"embed_dim": s.cfg.EmbedDim,
			"healthy":   true,
		},
	}
	if err := s.repo.Ping(ctx); err != nil {
		result["status"] = "degraded"
		result["database_error"] = err.Error()
		result["database"].(map[string]any)["connected"] = false
		result["database"].(map[string]any)["error"] = err.Error()
	}
	if err := s.embedder.Health(ctx); err != nil {
		result["status"] = "degraded"
		result["ollama_error"] = err.Error()
		result["ollama"].(map[string]any)["healthy"] = false
		result["ollama"].(map[string]any)["error"] = err.Error()
	}
	return result
}

func (s *Service) CreateTextDocument(ctx context.Context, req TextDocumentRequest) (db.Document, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	req.SourceType = strings.TrimSpace(req.SourceType)
	if req.SourceType == "" {
		req.SourceType = "manual"
	}
	if req.Content == "" {
		return db.Document{}, wrapValidation("content is required")
	}
	if len(req.Metadata) == 0 {
		req.Metadata = json.RawMessage(`{}`)
	}
	if !json.Valid(req.Metadata) {
		return db.Document{}, wrapValidation("metadata must be valid JSON")
	}
	return s.repo.CreateDocument(ctx, db.CreateDocumentParams{
		Title:      req.Title,
		Content:    req.Content,
		SourceType: req.SourceType,
		SourceID:   strings.TrimSpace(req.SourceID),
		URL:        strings.TrimSpace(req.URL),
		Checksum:   ingest.Checksum(req.Content),
		Metadata:   req.Metadata,
	})
}

func (s *Service) ListDocuments(ctx context.Context, filter db.ListDocumentsFilter) (db.PageResult[db.Document], error) {
	filter.Keyword = strings.TrimSpace(filter.Keyword)
	filter.Status = strings.TrimSpace(filter.Status)
	filter.SourceType = strings.TrimSpace(filter.SourceType)
	return s.repo.ListDocuments(ctx, filter)
}

func (s *Service) ListChunks(ctx context.Context, documentID int64, page, pageSize int) (db.PageResult[db.Chunk], error) {
	if documentID <= 0 {
		return db.PageResult[db.Chunk]{}, wrapValidation("document id is required")
	}
	return s.repo.ListChunks(ctx, documentID, page, pageSize)
}

func (s *Service) ReindexDocument(ctx context.Context, id int64) (db.Document, error) {
	if id <= 0 {
		return db.Document{}, wrapValidation("document id is required")
	}
	return s.repo.ReindexDocument(ctx, id)
}

func (s *Service) UpdateChunk(ctx context.Context, id int64, req UpdateChunkRequest) (db.Chunk, error) {
	if id <= 0 {
		return db.Chunk{}, wrapValidation("chunk id is required")
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return db.Chunk{}, wrapValidation("content is required")
	}
	doc, err := s.repo.GetDocumentByChunkID(ctx, id)
	if err != nil {
		return db.Chunk{}, err
	}
	embeddings, err := s.embedder.Embed(ctx, []string{ingest.BuildEmbeddingInput(doc, content)})
	if err != nil {
		return db.Chunk{}, err
	}
	return s.repo.UpdateChunkContent(ctx, id, content, ingest.Checksum(content), len([]rune(content)), embeddings[0])
}

func (s *Service) ChunkPreview(ctx context.Context, req ChunkPreviewRequest) (ChunkPreviewResponse, error) {
	req.Text = strings.TrimSpace(req.Text)
	if (req.DocumentID <= 0 && req.Text == "") || (req.DocumentID > 0 && req.Text != "") {
		return ChunkPreviewResponse{}, wrapValidation("provide exactly one of document_id or text")
	}

	source := "text"
	title := "临时文本"
	text := req.Text
	if req.DocumentID > 0 {
		doc, err := s.repo.GetDocument(ctx, req.DocumentID)
		if err != nil {
			return ChunkPreviewResponse{}, err
		}
		source = "document"
		title = doc.Title
		text = strings.TrimSpace(doc.Content)
	}
	if text == "" {
		return ChunkPreviewResponse{}, wrapValidation("text is required")
	}

	variants := req.Variants
	if len(variants) == 0 {
		variants = []ChunkPreviewVariant{
			{ChunkSize: 500, ChunkOverlap: 80},
			{ChunkSize: 700, ChunkOverlap: 120},
			{ChunkSize: 900, ChunkOverlap: 150},
		}
	}
	if len(variants) > 10 {
		return ChunkPreviewResponse{}, wrapValidation("variants cannot exceed 10")
	}

	response := ChunkPreviewResponse{Source: source, Title: title, Variants: make([]ChunkPreviewVariantResult, 0, len(variants))}
	for _, variant := range variants {
		if variant.ChunkSize <= 0 {
			return ChunkPreviewResponse{}, wrapValidation("chunk_size must be positive")
		}
		if variant.ChunkOverlap < 0 {
			return ChunkPreviewResponse{}, wrapValidation("chunk_overlap cannot be negative")
		}
		if variant.ChunkOverlap >= variant.ChunkSize {
			return ChunkPreviewResponse{}, wrapValidation("chunk_overlap must be smaller than chunk_size")
		}
		chunks := ingest.NewSplitter(variant.ChunkSize, variant.ChunkOverlap).Split(text)
		response.Variants = append(response.Variants, buildChunkPreviewVariant(variant, chunks))
	}
	return response, nil
}

func (s *Service) DeleteChunk(ctx context.Context, id int64) error {
	if id <= 0 {
		return wrapValidation("chunk id is required")
	}
	return s.repo.DeleteChunk(ctx, id)
}

func (s *Service) DeleteDocument(ctx context.Context, id int64) error {
	if id <= 0 {
		return wrapValidation("document id is required")
	}
	return s.repo.DeleteDocument(ctx, id)
}

func (s *Service) Overview(ctx context.Context) (db.Overview, error) {
	return s.repo.Overview(ctx)
}

func (s *Service) Query(ctx context.Context, req QueryRequest) (QueryResponse, error) {
	req.Question = strings.TrimSpace(req.Question)
	if req.Question == "" {
		return QueryResponse{}, wrapValidation("question is required")
	}
	topK := req.TopK
	if topK <= 0 {
		topK = s.cfg.TopK
	}
	embeddings, err := s.embedder.Embed(ctx, []string{req.Question})
	if err != nil {
		return QueryResponse{}, err
	}
	sources, err := s.repo.SearchChunks(ctx, embeddings[0], topK)
	if err != nil {
		return QueryResponse{}, err
	}
	return QueryResponse{
		Answer:  "Relevant sources were found. Please review the sources field.",
		Sources: sources,
		Usage: QueryUsage{
			TopK:           topK,
			EmbeddingModel: s.embedder.Model(),
		},
	}, nil
}

func wrapValidation(message string) error {
	return errors.Join(ErrValidation, errors.New(message))
}

func buildChunkPreviewVariant(variant ChunkPreviewVariant, chunks []string) ChunkPreviewVariantResult {
	result := ChunkPreviewVariantResult{
		ChunkSize:    variant.ChunkSize,
		ChunkOverlap: variant.ChunkOverlap,
		ChunkCount:   len(chunks),
		Chunks:       make([]ChunkPreviewChunk, 0, min(len(chunks), 20)),
	}
	if len(chunks) == 0 {
		return result
	}
	total := 0
	for i, chunk := range chunks {
		count := utf8.RuneCountInString(chunk)
		total += count
		if i == 0 || count < result.MinChars {
			result.MinChars = count
		}
		if count > result.MaxChars {
			result.MaxChars = count
		}
		if i < 20 {
			result.Chunks = append(result.Chunks, ChunkPreviewChunk{Index: i, CharCount: count, Content: chunk})
		}
	}
	result.AvgChars = float64(total) / float64(len(chunks))
	return result
}
