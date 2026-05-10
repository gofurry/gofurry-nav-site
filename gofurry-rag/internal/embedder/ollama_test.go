package embedder

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOllamaEmbed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var req struct {
			Model string   `json:"model"`
			Input []string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Model != "qwen3-embedding:0.6b" || len(req.Input) != 2 {
			t.Fatalf("unexpected request: %+v", req)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"embeddings": [][]float64{{0.1, 0.2}, {0.3, 0.4}},
		})
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "qwen3-embedding:0.6b", 2, nil)
	embeddings, err := client.Embed(context.Background(), []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(embeddings) != 2 || len(embeddings[0]) != 2 {
		t.Fatalf("unexpected embeddings: %#v", embeddings)
	}
}

func TestOllamaEmbedRejectsWrongDim(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"embeddings": [][]float64{{0.1}},
		})
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "model", 2, nil)
	if _, err := client.Embed(context.Background(), []string{"a"}); err == nil {
		t.Fatal("expected dimension error")
	}
}
