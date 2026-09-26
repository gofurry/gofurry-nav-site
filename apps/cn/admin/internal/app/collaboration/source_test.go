package collaboration

import (
	"strings"
	"testing"
)

func TestCanonicalSource(t *testing.T) {
	cases := []struct{ kind, source, want string }{
		{"game", " https://store.steampowered.com/app/000123/A_Name/?x=1 ", "steam:123"},
		{"game", "StEaM:00123", "steam:123"}, {"game", "123", "steam:123"},
		{"game", "0", ""}, {"game", "-3", ""}, {"game", "9223372036854775808", ""},
		{"game", "https://store.steampowered.com.evil.test/app/123", ""},
		{"game", "https://else.test/app/123", ""}, {"other", "123", ""}, {"other", "steam:123", ""},
		{"site", "HTTPS://WWW.Example.COM.:8443/path?q=1", "host:example.com"},
		{"site", "www.Example.com/path", "host:example.com"}, {"site", "a.example.com:443", "host:a.example.com"},
		{"site", "example.com.", "host:example.com"}, {"site", "example.com..evil.test", ""},
		{"site", "not a url", ""}, {"site", "javascript:alert(1)", ""}, {"site", "https://user:password@example.com/", ""},
	}
	for _, c := range cases {
		t.Run(c.source, func(t *testing.T) {
			if got := text(CanonicalSource(c.kind, c.source)); got != c.want {
				t.Fatalf("got %q want %q", got, c.want)
			}
		})
	}
}
func TestInputLimits(t *testing.T) {
	for _, in := range []Input{{Kind: "draft", Title: "x"}, {Kind: "game"}, {Kind: "game", Title: "x", Priority: "urgent"}, {Kind: "game", Title: strings.Repeat("a", 501)}} {
		if _, _, err := normalize(in); err == nil {
			t.Fatalf("accepted %+v", in)
		}
	}
	in, key, err := normalize(Input{Kind: "game", Source: " 123 "})
	if err != nil || in.Priority != "normal" || text(key) != "steam:123" {
		t.Fatalf("normalize=%+v %v", in, err)
	}
}
