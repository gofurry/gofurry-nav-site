package ingest

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/GoFurry/gofurry-rag/config"
	"github.com/GoFurry/gofurry-rag/internal/db"
	"github.com/GoFurry/gofurry-rag/internal/embedder"
)

type Repository interface {
	ClaimPendingDocument(ctx context.Context) (*db.Document, error)
	ReplaceChunks(ctx context.Context, documentID int64, chunks []db.NewChunk) error
	MarkDocumentFailed(ctx context.Context, id int64, message string) error
}

type Worker struct {
	repo     Repository
	embedder embedder.Client
	cfg      config.Config
	splitter Splitter
	once     sync.Once
	stats    workerStats
}

type WorkerStatus struct {
	State                 string     `json:"state"`
	ActiveWorkers         int        `json:"active_workers"`
	CurrentDocumentID     *int64     `json:"current_document_id,omitempty"`
	LastDocumentID        *int64     `json:"last_document_id,omitempty"`
	TotalProcessed        int64      `json:"total_processed"`
	TotalFailed           int64      `json:"total_failed"`
	LastDurationMs        int64      `json:"last_duration_ms"`
	AverageDurationMs     float64    `json:"average_duration_ms"`
	RecentError           string     `json:"recent_error,omitempty"`
	RecentErrorAt         *time.Time `json:"recent_error_at,omitempty"`
	LastSuccessAt         *time.Time `json:"last_success_at,omitempty"`
	LastStartedAt         *time.Time `json:"last_started_at,omitempty"`
	LastCompletedAt       *time.Time `json:"last_completed_at,omitempty"`
	CurrentBatchCount     int        `json:"current_batch_count"`
	CurrentEmbeddingModel string     `json:"current_embedding_model,omitempty"`
}

type workerStats struct {
	mu                    sync.RWMutex
	state                 string
	activeWorkers         int
	currentDocumentID     *int64
	lastDocumentID        *int64
	totalProcessed        int64
	totalFailed           int64
	totalDuration         time.Duration
	lastDuration          time.Duration
	recentError           string
	recentErrorAt         *time.Time
	lastSuccessAt         *time.Time
	lastStartedAt         *time.Time
	lastCompletedAt       *time.Time
	currentBatchCount     int
	currentEmbeddingModel string
}

func NewWorker(repo Repository, embedder embedder.Client, cfg config.Config) *Worker {
	return &Worker{
		repo:     repo,
		embedder: embedder,
		cfg:      cfg,
		splitter: NewSplitter(cfg.ChunkSize, cfg.ChunkOverlap),
	}
}

func (w *Worker) Start(ctx context.Context) {
	workers := w.cfg.IngestWorkers
	if workers <= 0 {
		workers = 1
	}
	w.once.Do(func() {
		for i := 0; i < workers; i++ {
			go w.loop(ctx)
		}
	})
}

func (w *Worker) Status() WorkerStatus {
	return w.stats.snapshot()
}

func (w *Worker) ProcessOnce(ctx context.Context) (bool, error) {
	doc, err := w.repo.ClaimPendingDocument(ctx)
	if err != nil || doc == nil {
		if doc == nil {
			w.stats.markIdle()
		}
		return false, err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(w.cfg.IngestTimeoutSeconds)*time.Second)
	defer cancel()

	startedAt := time.Now()
	w.stats.markProcessing(doc.ID, startedAt, w.embedder.Model())
	slog.InfoContext(timeoutCtx, "ingest document claimed",
		"document_id", doc.ID,
		"source_type", doc.SourceType,
		"status", doc.Status,
		"timeout_seconds", w.cfg.IngestTimeoutSeconds,
	)

	if err := w.process(timeoutCtx, *doc); err != nil {
		if markErr := w.repo.MarkDocumentFailed(timeoutCtx, doc.ID, err.Error()); markErr != nil {
			w.stats.markFailed(doc.ID, time.Since(startedAt), markErr.Error())
			return true, markErr
		}
		w.stats.markFailed(doc.ID, time.Since(startedAt), err.Error())
		return true, err
	}

	w.stats.markSuccess(doc.ID, time.Since(startedAt))
	return true, nil
}

func (w *Worker) loop(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		processed, err := w.ProcessOnce(ctx)
		if err != nil {
			slog.Error("ingest worker error", "error", err)
		}
		if processed {
			continue
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (w *Worker) process(ctx context.Context, doc db.Document) error {
	textChunks := w.splitter.Split(doc.Content)
	slog.InfoContext(ctx, "ingest document split",
		"document_id", doc.ID,
		"chunk_count", len(textChunks),
	)

	newChunks := make([]db.NewChunk, 0, len(textChunks))
	batchSize := w.cfg.EmbedBatchSize
	if batchSize <= 0 {
		batchSize = 8
	}

	for start := 0; start < len(textChunks); start += batchSize {
		end := start + batchSize
		if end > len(textChunks) {
			end = len(textChunks)
		}
		batch := textChunks[start:end]
		inputs := make([]string, 0, len(batch))
		for _, content := range batch {
			inputs = append(inputs, BuildEmbeddingInput(doc, content))
		}

		batchStartedAt := time.Now()
		slog.InfoContext(ctx, "ollama embed batch start",
			"document_id", doc.ID,
			"chunk_start", start,
			"chunk_end", end,
			"input_count", len(inputs),
			"model", w.embedder.Model(),
		)
		embedCtx := embedder.WithPriority(ctx, embedder.PriorityIngest)
		embeddings, err := w.embedder.Embed(embedCtx, inputs)
		if err != nil {
			slog.ErrorContext(embedCtx, "ollama embed batch failed",
				"document_id", doc.ID,
				"chunk_start", start,
				"chunk_end", end,
				"input_count", len(inputs),
				"elapsed_ms", time.Since(batchStartedAt).Milliseconds(),
				"model", w.embedder.Model(),
				"error", err,
			)
			return err
		}
		slog.InfoContext(ctx, "ollama embed batch complete",
			"document_id", doc.ID,
			"chunk_start", start,
			"chunk_end", end,
			"input_count", len(inputs),
			"elapsed_ms", time.Since(batchStartedAt).Milliseconds(),
			"model", w.embedder.Model(),
		)

		for i, content := range batch {
			newChunks = append(newChunks, db.NewChunk{
				ChunkIndex:  start + i,
				Content:     content,
				ContentHash: Checksum(content),
				TokenCount:  runeLen(content),
				Embedding:   embeddings[i],
			})
		}
	}

	if err := w.repo.ReplaceChunks(ctx, doc.ID, newChunks); err != nil {
		slog.ErrorContext(ctx, "ingest document persist failed",
			"document_id", doc.ID,
			"chunk_count", len(newChunks),
			"error", err,
		)
		return err
	}

	slog.InfoContext(ctx, "ingest document complete",
		"document_id", doc.ID,
		"chunk_count", len(newChunks),
	)
	return nil
}

func (s *workerStats) snapshot() WorkerStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	avg := 0.0
	if s.totalProcessed > 0 {
		avg = float64(s.totalDuration.Milliseconds()) / float64(s.totalProcessed)
	}

	return WorkerStatus{
		State:                 s.state,
		ActiveWorkers:         s.activeWorkers,
		CurrentDocumentID:     cloneInt64Ptr(s.currentDocumentID),
		LastDocumentID:        cloneInt64Ptr(s.lastDocumentID),
		TotalProcessed:        s.totalProcessed,
		TotalFailed:           s.totalFailed,
		LastDurationMs:        s.lastDuration.Milliseconds(),
		AverageDurationMs:     avg,
		RecentError:           s.recentError,
		RecentErrorAt:         cloneTimePtr(s.recentErrorAt),
		LastSuccessAt:         cloneTimePtr(s.lastSuccessAt),
		LastStartedAt:         cloneTimePtr(s.lastStartedAt),
		LastCompletedAt:       cloneTimePtr(s.lastCompletedAt),
		CurrentBatchCount:     s.currentBatchCount,
		CurrentEmbeddingModel: s.currentEmbeddingModel,
	}
}

func (s *workerStats) markIdle() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeWorkers = 0
	s.state = "idle"
	s.currentDocumentID = nil
	s.currentBatchCount = 0
}

func (s *workerStats) markProcessing(documentID int64, startedAt time.Time, model string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = "processing"
	s.activeWorkers++
	s.currentDocumentID = ptrInt64(documentID)
	s.lastStartedAt = ptrTime(startedAt)
	s.currentEmbeddingModel = model
}

func (s *workerStats) markSuccess(documentID int64, duration time.Duration) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activeWorkers > 0 {
		s.activeWorkers--
	}
	if s.activeWorkers > 0 {
		s.state = "processing"
	} else {
		s.state = "idle"
	}
	s.currentDocumentID = nil
	s.lastDocumentID = ptrInt64(documentID)
	s.totalProcessed++
	s.totalDuration += duration
	s.lastDuration = duration
	s.lastSuccessAt = ptrTime(now)
	s.lastCompletedAt = ptrTime(now)
	s.currentBatchCount = 0
}

func (s *workerStats) markFailed(documentID int64, duration time.Duration, message string) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activeWorkers > 0 {
		s.activeWorkers--
	}
	if s.activeWorkers > 0 {
		s.state = "processing"
	} else {
		s.state = "failed"
	}
	s.currentDocumentID = nil
	s.lastDocumentID = ptrInt64(documentID)
	s.totalFailed++
	s.lastDuration = duration
	s.recentError = message
	s.recentErrorAt = ptrTime(now)
	s.lastCompletedAt = ptrTime(now)
	s.currentBatchCount = 0
}

func ptrInt64(value int64) *int64 {
	v := value
	return &v
}

func ptrTime(value time.Time) *time.Time {
	v := value
	return &v
}

func cloneInt64Ptr(value *int64) *int64 {
	if value == nil {
		return nil
	}
	v := *value
	return &v
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	v := *value
	return &v
}
