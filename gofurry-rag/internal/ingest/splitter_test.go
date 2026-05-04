package ingest

import (
	"strings"
	"testing"
)

func TestSplitterKeepsSmallText(t *testing.T) {
	got := NewSplitter(20, 5).Split("第一段\n\n第二段")
	if len(got) != 1 {
		t.Fatalf("len = %d, chunks = %#v", len(got), got)
	}
}

func TestSplitterHardSplitsWithOverlap(t *testing.T) {
	got := NewSplitter(10, 3).Split(strings.Repeat("猫", 25))
	if len(got) != 4 {
		t.Fatalf("len = %d, chunks = %#v", len(got), got)
	}
	if runeLen(got[1]) > 10 {
		t.Fatalf("chunk too long: %q", got[1])
	}
	if !strings.HasPrefix(got[1], strings.Repeat("猫", 3)) {
		t.Fatalf("missing overlap: %#v", got)
	}
}
