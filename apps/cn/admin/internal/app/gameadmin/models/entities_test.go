package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGameCommentIDMarshalsWithoutJavaScriptPrecisionLoss(t *testing.T) {
	t.Parallel()

	const id int64 = 9007199254740993
	raw, err := json.Marshal(GameComment{ID: id})
	if err != nil {
		t.Fatalf("marshal comment: %v", err)
	}
	if !strings.Contains(string(raw), `"id":"9007199254740993"`) {
		t.Fatalf("comment id must cross JSON as a string, got %s", raw)
	}
}
