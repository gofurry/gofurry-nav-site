package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"
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
	BatchReindexDocuments(ctx context.Context, filter db.BatchDocumentFilter) (db.BatchResult, error)
	RetryFailedDocuments(ctx context.Context, filter db.BatchDocumentFilter) (db.BatchResult, error)
	UpdateChunkContent(ctx context.Context, id int64, content, contentHash string, tokenCount int, embedding []float64) (db.Chunk, error)
	DeleteChunk(ctx context.Context, id int64) error
	DeleteDocument(ctx context.Context, id int64) error
	Overview(ctx context.Context) (db.Overview, error)
	SearchChunks(ctx context.Context, embedding []float64, topK int, filter db.BatchDocumentFilter) ([]db.Source, error)
}

type Service struct {
	repo     Repository
	embedder embedder.Client
	chat     chatClient
	cfg      config.Config
	worker   workerStatusProvider
}

type workerStatusProvider interface {
	Status() ingest.WorkerStatus
}

type ollamaQueueStatuser interface {
	QueueStatus() embedder.OllamaQueueStatus
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
	Question       string       `json:"question"`
	TopK           int          `json:"top_k"`
	Filters        QueryFilters `json:"filters"`
	IncludeDetails bool         `json:"include_details"`
}

type UpdateChunkRequest struct {
	Content string `json:"content"`
}

type MetadataFilters struct {
	SourceType []string `json:"source_type"`
	Category   []string `json:"category"`
	Language   []string `json:"language"`
	Status     []string `json:"status"`
}

type BatchDocumentsRequest struct {
	Scope       string          `json:"scope"`
	DocumentIDs []int64         `json:"document_ids"`
	Filters     MetadataFilters `json:"filters"`
}

type QueryFilters struct {
	SourceType  []string `json:"source_type"`
	DocumentIDs []int64  `json:"document_ids"`
	Category    []string `json:"category"`
	Language    []string `json:"language"`
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
	Answer    string          `json:"answer"`
	Sources   []db.Source     `json:"sources"`
	Citations []QueryCitation `json:"citations,omitempty"`
	Usage     QueryUsage      `json:"usage"`
}

type QueryCitation struct {
	Rank         int                   `json:"rank"`
	UsedInPrompt bool                  `json:"used_in_prompt"`
	Source       db.Source             `json:"source"`
	Lineage      QueryCitationLineage  `json:"lineage"`
	Document     QueryCitationDocument `json:"document"`
	Chunk        QueryCitationChunk    `json:"chunk"`
}

type QueryCitationLineage struct {
	DocumentID int64   `json:"document_id"`
	ChunkID    int64   `json:"chunk_id"`
	ChunkIndex int     `json:"chunk_index"`
	SourceType string  `json:"source_type"`
	SourceID   string  `json:"source_id,omitempty"`
	Title      string  `json:"title"`
	URL        string  `json:"url,omitempty"`
	Score      float64 `json:"score"`
	TokenCount int     `json:"token_count"`
}

type QueryCitationDocument struct {
	ID                 int64           `json:"id"`
	SourceType         string          `json:"source_type"`
	SourceID           string          `json:"source_id,omitempty"`
	Title              string          `json:"title"`
	URL                string          `json:"url,omitempty"`
	Checksum           string          `json:"checksum,omitempty"`
	Content            string          `json:"content"`
	Status             string          `json:"status"`
	ErrorMessage       string          `json:"error_message"`
	Metadata           json.RawMessage `json:"metadata,omitempty"`
	ChunkCount         int             `json:"chunk_count"`
	RetryCount         int             `json:"retry_count"`
	LastErrorAt        *time.Time      `json:"last_error_at,omitempty"`
	ProcessedAt        *time.Time      `json:"processed_at,omitempty"`
	ReindexRequestedAt *time.Time      `json:"reindex_requested_at,omitempty"`
	LastIndexedAt      *time.Time      `json:"last_indexed_at,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type QueryCitationChunk struct {
	ID          int64      `json:"id"`
	DocumentID  int64      `json:"document_id"`
	ChunkIndex  int        `json:"chunk_index"`
	Content     string     `json:"content"`
	ContentHash string     `json:"content_hash"`
	TokenCount  int        `json:"token_count"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
}

type QueryUsage struct {
	TopK             int    `json:"top_k"`
	EmbeddingModel   string `json:"embedding_model"`
	AnswerModel      string `json:"answer_model,omitempty"`
	PromptTokens     int    `json:"prompt_tokens,omitempty"`
	CompletionTokens int    `json:"completion_tokens,omitempty"`
	TotalTokens      int    `json:"total_tokens,omitempty"`
	CachedTokens     int    `json:"cached_tokens,omitempty"`
	ReasoningTokens  int    `json:"reasoning_tokens,omitempty"`
}

func New(repo Repository, embedder embedder.Client, chat chatClient, cfg config.Config, worker workerStatusProvider) *Service {
	return &Service{repo: repo, embedder: embedder, chat: chat, cfg: cfg, worker: worker}
}

func (s *Service) Health(ctx context.Context) map[string]any {
	ollamaQueue := s.ollamaQueueStatus()
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
			"queue":     ollamaQueue,
		},
		"tencent": map[string]any{
			"base_url":   s.cfg.TencentBaseURL,
			"model":      s.cfg.TencentModel,
			"configured": s.chat != nil && s.chat.Configured(),
			"healthy":    false,
		},
	}
	if s.worker != nil {
		status := s.worker.Status()
		result["worker"] = map[string]any{
			"state":                   status.State,
			"active_workers":          status.ActiveWorkers,
			"total_processed":         status.TotalProcessed,
			"total_failed":            status.TotalFailed,
			"last_duration_ms":        status.LastDurationMs,
			"average_duration_ms":     status.AverageDurationMs,
			"recent_error":            status.RecentError,
			"recent_error_at":         status.RecentErrorAt,
			"last_success_at":         status.LastSuccessAt,
			"last_started_at":         status.LastStartedAt,
			"last_completed_at":       status.LastCompletedAt,
			"current_document_id":     status.CurrentDocumentID,
			"current_embedding_model": status.CurrentEmbeddingModel,
		}
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
	if s.chat != nil && s.chat.Configured() {
		if err := s.chat.Health(ctx); err != nil {
			result["status"] = "degraded"
			result["tencent_error"] = err.Error()
			result["tencent"].(map[string]any)["error"] = err.Error()
		} else {
			result["tencent"].(map[string]any)["healthy"] = true
		}
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
	filter.Category = strings.TrimSpace(filter.Category)
	filter.Language = strings.TrimSpace(filter.Language)
	filter.SourceTypes = cleanStrings(filter.SourceTypes)
	filter.DocumentIDs = cleanDocumentIDs(filter.DocumentIDs)
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

func (s *Service) BatchReindexDocuments(ctx context.Context, req BatchDocumentsRequest) (db.BatchResult, error) {
	filter, err := buildBatchDocumentFilter(req)
	if err != nil {
		return db.BatchResult{}, err
	}
	return s.repo.BatchReindexDocuments(ctx, filter)
}

func (s *Service) RetryFailedDocuments(ctx context.Context, req BatchDocumentsRequest) (db.BatchResult, error) {
	filter, err := buildBatchDocumentFilter(req)
	if err != nil {
		return db.BatchResult{}, err
	}
	return s.repo.RetryFailedDocuments(ctx, filter)
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
	embedCtx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.EmbedTimeoutSeconds)*time.Second)
	defer cancel()
	embedCtx = embedder.WithPriority(embedCtx, embedder.PriorityIngest)
	slog.InfoContext(embedCtx, "chunk embedding start",
		"document_id", doc.ID,
		"chunk_id", id,
		"model", s.embedder.Model(),
	)
	embeddings, err := s.embedder.Embed(embedCtx, []string{ingest.BuildEmbeddingInput(doc, content)})
	if err != nil {
		slog.ErrorContext(embedCtx, "chunk embedding failed",
			"document_id", doc.ID,
			"chunk_id", id,
			"model", s.embedder.Model(),
			"error", err,
		)
		return db.Chunk{}, err
	}
	slog.InfoContext(embedCtx, "chunk embedding complete",
		"document_id", doc.ID,
		"chunk_id", id,
		"model", s.embedder.Model(),
	)
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
	overview, err := s.repo.Overview(ctx)
	if err != nil {
		return db.Overview{}, err
	}
	overview.OllamaQueue = s.ollamaQueueStatus()
	if s.worker != nil {
		status := s.worker.Status()
		overview.WorkerState = status.State
		overview.WorkerActiveWorkers = status.ActiveWorkers
		overview.WorkerCurrentDocumentID = status.CurrentDocumentID
		overview.WorkerLastDocumentID = status.LastDocumentID
		overview.WorkerTotalProcessed = status.TotalProcessed
		overview.WorkerTotalFailed = status.TotalFailed
		overview.WorkerLastDurationMs = status.LastDurationMs
		overview.WorkerAverageDurationMs = status.AverageDurationMs
		overview.WorkerRecentError = status.RecentError
		overview.WorkerRecentErrorAt = status.RecentErrorAt
		overview.WorkerLastSuccessAt = status.LastSuccessAt
		overview.WorkerLastStartedAt = status.LastStartedAt
		overview.WorkerLastCompletedAt = status.LastCompletedAt
	}
	return overview, nil
}

func (s *Service) ollamaQueueStatus() db.OllamaQueueStatus {
	if statuser, ok := s.embedder.(ollamaQueueStatuser); ok {
		queue := statuser.QueueStatus()
		return db.OllamaQueueStatus{
			MaxConcurrency:     queue.MaxConcurrency,
			QueryQueueSize:     queue.QueryQueueSize,
			IngestQueueSize:    queue.IngestQueueSize,
			Active:             queue.Active,
			QueuedQuery:        queue.QueuedQuery,
			QueuedIngest:       queue.QueuedIngest,
			Rejected:           queue.Rejected,
			OldestWaitMs:       queue.OldestWaitMs,
			WaitTimeoutSeconds: queue.WaitTimeoutSeconds,
		}
	}
	return db.OllamaQueueStatus{}
}

func (s *Service) Query(ctx context.Context, req QueryRequest) (QueryResponse, error) {
	return s.executeQuery(ctx, req, QueryCallbacks{})
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

func buildBatchDocumentFilter(req BatchDocumentsRequest) (db.BatchDocumentFilter, error) {
	scope := strings.TrimSpace(req.Scope)
	filter := db.BatchDocumentFilter{
		DocumentIDs: cleanDocumentIDs(req.DocumentIDs),
		Statuses:    cleanStrings(req.Filters.Status),
		SourceTypes: cleanStrings(req.Filters.SourceType),
		Categories:  cleanStrings(req.Filters.Category),
		Languages:   cleanStrings(req.Filters.Language),
	}
	switch scope {
	case "all":
		return db.BatchDocumentFilter{}, nil
	case "filters":
		if len(filter.Statuses) == 0 && len(filter.SourceTypes) == 0 && len(filter.Categories) == 0 && len(filter.Languages) == 0 {
			return db.BatchDocumentFilter{}, wrapValidation("filters scope requires at least one filter")
		}
		return filter, nil
	case "document_ids":
		if len(filter.DocumentIDs) == 0 {
			return db.BatchDocumentFilter{}, wrapValidation("document_ids scope requires document_ids")
		}
		return filter, nil
	default:
		return db.BatchDocumentFilter{}, wrapValidation("scope must be one of all, filters, document_ids")
	}
}

func cleanStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func cleanDocumentIDs(values []int64) []int64 {
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value > 0 {
			result = append(result, value)
		}
	}
	return result
}
