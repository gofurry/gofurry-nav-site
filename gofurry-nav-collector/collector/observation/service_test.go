package observation

import "testing"

func TestLatestKey(t *testing.T) {
	got := LatestKey(ProtocolPing, 123)
	want := "collector:v2:latest:ping:123"
	if got != want {
		t.Fatalf("LatestKey() = %q, want %q", got, want)
	}
}
