package ingest

import "testing"

func TestChecksumStable(t *testing.T) {
	first := Checksum("gofurry")
	second := Checksum("gofurry")
	if first != second {
		t.Fatalf("checksum changed: %s != %s", first, second)
	}
	if first == Checksum("gofurry updated") {
		t.Fatal("checksum should change when content changes")
	}
}
