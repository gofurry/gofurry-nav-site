package ingest

import (
	"strings"
	"testing"

	"github.com/gofurry/gofurry-rag/internal/db"
)

func TestBuildEmbeddingInputAddsDocumentContext(t *testing.T) {
	got := BuildEmbeddingInput(db.Document{
		Title:      "gofurry",
		SourceType: "site",
		SourceID:   "about",
		URL:        "https://example.com/about",
	}, "公益导航")
	for _, want := range []string{"Title: gofurry", "Source Type: site", "Source ID: about", "URL: https://example.com/about", "Content:\n公益导航"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func TestBuildEmbeddingInputKeepsPlainChunkWithoutContext(t *testing.T) {
	got := BuildEmbeddingInput(db.Document{}, "  plain chunk  ")
	if got != "plain chunk" {
		t.Fatalf("got %q", got)
	}
}
