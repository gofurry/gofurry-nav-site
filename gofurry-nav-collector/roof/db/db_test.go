package db

import (
	"net/url"
	"testing"
)

func TestBuildPostgresDSNEscapesCredentials(t *testing.T) {
	dsn := buildPostgresDSN("127.0.0.1", "5432", "post gres", "pa ss!@#", "gfn")

	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("buildPostgresDSN() returned invalid URL: %v", err)
	}
	if parsed.Scheme != "postgres" {
		t.Fatalf("scheme = %q, want postgres", parsed.Scheme)
	}
	if parsed.Host != "127.0.0.1:5432" {
		t.Fatalf("host = %q, want 127.0.0.1:5432", parsed.Host)
	}
	if parsed.Path != "/gfn" {
		t.Fatalf("path = %q, want /gfn", parsed.Path)
	}
	if got := parsed.User.Username(); got != "post gres" {
		t.Fatalf("username = %q, want post gres", got)
	}
	if got, _ := parsed.User.Password(); got != "pa ss!@#" {
		t.Fatalf("password = %q, want pa ss!@#", got)
	}
}
