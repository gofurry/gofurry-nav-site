package service

import (
	"testing"

	"github.com/gofurry/gofurry-nav-backend/common/util"
)

func TestLegacyPingValueParsingIsDefensive(t *testing.T) {
	if got := util.ExtractSuffix2Int("", "ms"); got != 0 {
		t.Fatalf("empty delay parsed as %d", got)
	}
	if got := util.ExtractSuffix2Int("bad", "ms"); got != 0 {
		t.Fatalf("bad delay parsed as %d", got)
	}
	if got := util.ExtractSuffix2Int("12ms", "ms"); got != 12 {
		t.Fatalf("valid delay parsed as %d", got)
	}
	if avgWithUnit(30, 0, "ms") != "0ms" {
		t.Fatalf("zero-count average should be safe")
	}
	if avgWithUnit(30, 3, "ms") != "10ms" {
		t.Fatalf("average mismatch")
	}
}
