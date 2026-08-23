package domain

import "testing"

func TestSteamNewsURL(t *testing.T) {
	t.Parallel()

	got := SteamNewsURL(550, "1234567890")
	want := "https://store.steampowered.com/news/app/550/view/1234567890"
	if got != want {
		t.Fatalf("unexpected news URL: got %q want %q", got, want)
	}

	if got := SteamNewsURL(0, "123"); got != "" {
		t.Fatalf("expected empty URL for missing appid, got %q", got)
	}
	if got := SteamNewsURL(550, ""); got != "" {
		t.Fatalf("expected empty URL for missing gid, got %q", got)
	}
}
