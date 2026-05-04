package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/GoFurry/gofurry-rag/internal/config"
	"github.com/GoFurry/gofurry-rag/internal/db"
	"github.com/GoFurry/gofurry-rag/internal/embedder"
	"github.com/GoFurry/gofurry-rag/internal/ingest"
)

var ErrValidation = errors.New("validation failed")

type Repository interface {
	Ping(ctx context.Context) error
	CreateDocument(ctx context.Context, params db.CreateDocumentParams) (db.Document, error)
	ListDocuments(ctx context.Context, filter db.ListDocumentsFilter) (db.PageResult[db.Document], error)
	ListChunks(ctx context.Context, documentID int64, page, pageSize int) (db.PageResult[db.Chunk], error)
	DeleteDocument(ctx context.Context, id int64) error
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
	}
	if err := s.repo.Ping(ctx); err != nil {
		result["status"] = "degraded"
		result["database_error"] = err.Error()
	}
	if err := s.embedder.Health(ctx); err != nil {
		result["status"] = "degraded"
		result["ollama_error"] = err.Error()
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

func (s *Service) DeleteDocument(ctx context.Context, id int64) error {
	if id <= 0 {
		return wrapValidation("document id is required")
	}
	return s.repo.DeleteDocument(ctx, id)
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
		Answer:  "已找到以下相关资料，请参考 sources。",
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
