package main

import (
	"context"
	"log"
	"time"

	"github.com/GoFurry/gofurry-rag/internal/api"
	"github.com/GoFurry/gofurry-rag/internal/config"
	"github.com/GoFurry/gofurry-rag/internal/db"
	"github.com/GoFurry/gofurry-rag/internal/embedder"
	"github.com/GoFurry/gofurry-rag/internal/ingest"
	"github.com/GoFurry/gofurry-rag/internal/service"
)

func main() {
	config.LoadDotEnv(".env")
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	repo := db.NewRepository(pool)
	if err := repo.Migrate(ctx); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	embedClient := embedder.NewOllamaClient(cfg.OllamaBaseURL, cfg.EmbedModel, cfg.EmbedDim)
	ragService := service.New(repo, embedClient, cfg)
	worker := ingest.NewWorker(repo, embedClient, cfg)
	worker.Start(context.Background())

	app := api.NewServer(cfg, ragService, worker).App()
	log.Printf("gofurry-rag listening on %s", cfg.AppAddr)
	if err := app.Listen(cfg.AppAddr); err != nil {
		log.Fatal(err)
	}
}
