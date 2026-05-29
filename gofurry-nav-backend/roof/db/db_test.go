package db

import (
	"net/url"
	"strings"
	"testing"

	"github.com/gofurry/gofurry-nav-backend/roof/env"
)

func TestBuildPostgresDSNEscapesSpecialPassword(t *testing.T) {
	dsn := buildPostgresDSN(env.DataBaseConfig{
		DBHost:     "127.0.0.1",
		DBPort:     "5432",
		DBUsername: "postgres",
		DBPassword: "p@ss word:with/slash",
		DBName:     "gfn",
	})
	if strings.Contains(dsn, "p@ss word") {
		t.Fatalf("dsn leaked unescaped password: %s", dsn)
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse dsn error = %v", err)
	}
	password, _ := parsed.User.Password()
	if password != "p@ss word:with/slash" {
		t.Fatalf("password = %q", password)
	}
	if parsed.Query().Get("sslmode") != "disable" {
		t.Fatalf("sslmode = %q", parsed.Query().Get("sslmode"))
	}
}
