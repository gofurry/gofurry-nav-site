package api

import (
	"strings"
	"unicode/utf8"

	"github.com/gofurry/gofurry-rag/internal/db"
	"github.com/gofurry/gofurry-rag/internal/service"
)

const publicSourceSnippetRunes = 600

type publicQueryResponse struct {
	Answer  string              `json:"answer"`
	Sources []publicQuerySource `json:"sources"`
	Usage   service.QueryUsage  `json:"usage"`
}

type publicQuerySource struct {
	Title      string  `json:"title"`
	URL        string  `json:"url,omitempty"`
	SourceType string  `json:"source_type"`
	Snippet    string  `json:"snippet"`
	Score      float64 `json:"score"`
	ChunkIndex int     `json:"chunk_index"`
}

func newPublicQueryResponse(response service.QueryResponse) publicQueryResponse {
	return publicQueryResponse{
		Answer:  response.Answer,
		Sources: publicQuerySources(response.Sources),
		Usage:   response.Usage,
	}
}

func publicQuerySources(sources []db.Source) []publicQuerySource {
	if len(sources) == 0 {
		return []publicQuerySource{}
	}
	result := make([]publicQuerySource, 0, len(sources))
	for _, source := range sources {
		result = append(result, publicQuerySource{
			Title:      source.Title,
			URL:        source.URL,
			SourceType: source.SourceType,
			Snippet:    publicSnippet(source.Content),
			Score:      source.Score,
			ChunkIndex: source.ChunkIndex,
		})
	}
	return result
}

func publicSnippet(content string) string {
	content = strings.TrimSpace(content)
	if utf8.RuneCountInString(content) <= publicSourceSnippetRunes {
		return content
	}
	runes := []rune(content)
	return strings.TrimSpace(string(runes[:publicSourceSnippetRunes])) + "..."
}
