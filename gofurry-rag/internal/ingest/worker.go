package ingest

import (
	"context"
	"log"
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

func (w *Worker) ProcessOnce(ctx context.Context) (bool, error) {
	doc, err := w.repo.ClaimPendingDocument(ctx)
	if err != nil || doc == nil {
		return false, err
	}
	if err := w.process(ctx, *doc); err != nil {
		if markErr := w.repo.MarkDocumentFailed(ctx, doc.ID, err.Error()); markErr != nil {
			return true, markErr
		}
		return true, err
	}
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
			log.Printf("ingest worker error: %v", err)
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
		embeddings, err := w.embedder.Embed(ctx, batch)
		if err != nil {
			return err
		}
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
	return w.repo.ReplaceChunks(ctx, doc.ID, newChunks)
}
