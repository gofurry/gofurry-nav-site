package retrieval

import (
	"context"

	"github.com/gofurry/gofurry-rag/internal/db"
)

type Repository interface {
	SearchChunks(ctx context.Context, embedding []float64, topK int) ([]db.Source, error)
}

type Retriever struct {
	repo Repository
}

func New(repo Repository) *Retriever {
	return &Retriever{repo: repo}
}

func (r *Retriever) Search(ctx context.Context, embedding []float64, topK int) ([]db.Source, error) {
	return r.repo.SearchChunks(ctx, embedding, topK)
}
