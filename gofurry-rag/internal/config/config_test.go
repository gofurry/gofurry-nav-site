package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_NAME", "")
	t.Setenv("APP_ADDR", "")
	t.Setenv("RAG_EMBED_DIM", "")

	cfg := Load()
	if cfg.AppName != "gofurry-rag" {
		t.Fatalf("AppName = %q", cfg.AppName)
	}
	if cfg.AppAddr != "127.0.0.1:8080" {
		t.Fatalf("AppAddr = %q", cfg.AppAddr)
	}
	if cfg.EmbedDim != 1024 {
		t.Fatalf("EmbedDim = %d", cfg.EmbedDim)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("APP_ADDR", "0.0.0.0:9000")
	t.Setenv("RAG_CHUNK_SIZE", "300")

	cfg := Load()
	if cfg.AppAddr != "0.0.0.0:9000" {
		t.Fatalf("AppAddr = %q", cfg.AppAddr)
	}
	if cfg.ChunkSize != 300 {
		t.Fatalf("ChunkSize = %d", cfg.ChunkSize)
	}
}
