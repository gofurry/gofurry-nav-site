package models

import "time"

const (
	SearchSuggestionsSchemaVersion = 1

	SearchSuggestionsStateReady = "ready"
	SearchSuggestionsStateEmpty = "empty"
	SearchSuggestionsStateError = "error"

	SearchSuggestionsCacheHit  = "hit"
	SearchSuggestionsCacheMiss = "miss"
)

type SearchSuggestionsResponse struct {
	SchemaVersion  int       `json:"schema_version"`
	GeneratedAt    time.Time `json:"generated_at"`
	State          string    `json:"state"`
	Engine         string    `json:"engine"`
	Query          string    `json:"query"`
	Suggestions    []string  `json:"suggestions"`
	CacheState     string    `json:"cache_state"`
	ReasonMessages []string  `json:"reason_messages,omitempty"`
}
