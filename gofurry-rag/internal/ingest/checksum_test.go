package ingest

import "testing"

func TestChecksumStable(t *testing.T) {
	first := Checksum("GoFurry")
	second := Checksum("GoFurry")
	if first != second {
		t.Fatalf("checksum changed: %s != %s", first, second)
	}
	if first == Checksum("gofurry") {
		t.Fatal("checksum should change when content changes")
	}
}
