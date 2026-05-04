package embedder

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type OllamaClient struct {
	baseURL string
	model   string
	dim     int
	client  *http.Client
}

func NewOllamaClient(baseURL, model string, dim int) *OllamaClient {
	return &OllamaClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		dim:     dim,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *OllamaClient) Model() string {
	return c.model
}

func (c *OllamaClient) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/tags", nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ollama tags returned %s", resp.Status)
	}
	return nil
}

func (c *OllamaClient) Embed(ctx context.Context, input []string) ([][]float64, error) {
	if len(input) == 0 {
		return [][]float64{}, nil
	}
	payload := map[string]any{
		"model": c.model,
		"input": input,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ollama embed returned %s", resp.Status)
	}

	var result struct {
		Embeddings [][]float64 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Embeddings) != len(input) {
		return nil, fmt.Errorf("ollama returned %d embeddings for %d inputs", len(result.Embeddings), len(input))
	}
	for i, embedding := range result.Embeddings {
		if len(embedding) != c.dim {
			return nil, fmt.Errorf("embedding %d has dimension %d, want %d", i, len(embedding), c.dim)
		}
	}
	return result.Embeddings, nil
}

var ErrModelUnavailable = errors.New("embedding model unavailable")
