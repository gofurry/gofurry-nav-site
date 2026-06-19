package embedder

import "context"

type Client interface {
	Embed(ctx context.Context, input []string) ([][]float64, error)
	Health(ctx context.Context) error
	Model() string
}
