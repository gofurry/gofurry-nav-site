package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppName        string
	AppEnv         string
	AppAddr        string
	AdminToken     string
	DatabaseDSN    string
	OllamaBaseURL  string
	EmbedModel     string
	EmbedDim       int
	ChunkSize      int
	ChunkOverlap   int
	TopK           int
	IngestWorkers  int
	EmbedBatchSize int
}

func Load() Config {
	return Config{
		AppName:        envString("APP_NAME", "gofurry-rag"),
		AppEnv:         envString("APP_ENV", "dev"),
		AppAddr:        envString("APP_ADDR", "127.0.0.1:8080"),
		AdminToken:     envString("RAG_ADMIN_TOKEN", "change-me"),
		DatabaseDSN:    envString("RAG_DB_DSN", "postgres://postgres:your_password@192.168.153.121:5432/postgres?sslmode=disable"),
		OllamaBaseURL:  envString("RAG_OLLAMA_BASE_URL", "http://148.70.18.111:43434"),
		EmbedModel:     envString("RAG_EMBED_MODEL", "qwen3-embedding:0.6b"),
		EmbedDim:       envInt("RAG_EMBED_DIM", 1024),
		ChunkSize:      envInt("RAG_CHUNK_SIZE", 700),
		ChunkOverlap:   envInt("RAG_CHUNK_OVERLAP", 120),
		TopK:           envInt("RAG_TOP_K", 6),
		IngestWorkers:  envInt("RAG_INGEST_WORKERS", 1),
		EmbedBatchSize: envInt("RAG_EMBED_BATCH_SIZE", 8),
	}
}

func LoadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key != "" {
			_ = os.Setenv(key, value)
		}
	}
}

func envString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
