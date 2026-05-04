package ingest

import (
	"strings"

	"github.com/GoFurry/gofurry-rag/internal/db"
)

func BuildEmbeddingInput(doc db.Document, chunk string) string {
	chunk = strings.TrimSpace(chunk)
	parts := make([]string, 0, 5)
	if title := strings.TrimSpace(doc.Title); title != "" {
		parts = append(parts, "Title: "+title)
	}
	if sourceType := strings.TrimSpace(doc.SourceType); sourceType != "" {
		parts = append(parts, "Source Type: "+sourceType)
	}
	if sourceID := strings.TrimSpace(doc.SourceID); sourceID != "" {
		parts = append(parts, "Source ID: "+sourceID)
	}
	if url := strings.TrimSpace(doc.URL); url != "" {
		parts = append(parts, "URL: "+url)
	}
	if len(parts) == 0 {
		return chunk
	}
	parts = append(parts, "Content:\n"+chunk)
	return strings.Join(parts, "\n")
}
