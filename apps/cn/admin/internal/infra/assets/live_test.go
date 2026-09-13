package assets

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	env "github.com/gofurry/gofurry-admin/config"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// Explicit dev-only acceptance. The application never loads a root .env.
func TestRealDevStorage(t *testing.T) {
	config := os.Getenv("GOFURRY_ASSET_DEV_CONFIG")
	if config == "" {
		t.Skip("set GOFURRY_ASSET_DEV_CONFIG to ignored Admin dev server.yaml")
	}
	if err := env.MustInitServerConfig("gofurry-admin", config); err != nil {
		t.Fatal("dev config failed to load")
	}
	cfg := env.GetServerConfig().ExternalServices.AssetStorage
	if !strings.Contains(cfg.Primary.Bucket, "-dev-") || !strings.HasSuffix(cfg.Mirror.Bucket, "-dev") || !strings.Contains(cfg.Primary.PublicBaseURL, "assets-dev.") || !strings.Contains(cfg.Mirror.PublicBaseURL, "assets-dev.") {
		t.Fatal("real acceptance requires development buckets and CDNs")
	}
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	avif, err := os.ReadFile("testdata/hero.avif")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		kind, name string
		data       []byte
	}{
		{"site-icon", "fixture.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16"><circle cx="8" cy="8" r="5" fill="#765c43"/></svg>`)},
		{"hero-desktop", "desktop.avif", avif}, {"hero-mobile", "mobile.avif", avif},
		{"pattern", "pattern.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24"><path d="M1 1h3v3H1z"/></svg>`)},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			o, err := NewObject(tc.kind, 9223372036854775000, tc.name, tc.data)
			if err != nil {
				t.Fatal(err)
			}
			result, err := s.Publish(ctx, o)
			if err != nil || result.Mirror != "ready" {
				t.Fatalf("real publication failed: %v %+v", err, result)
			}
			for name, store := range map[string]ObjectStore{"COS": s.Primary, "R2": s.Mirror} {
				info, err := store.Head(ctx, o.Key)
				if err != nil || verify(info, o) != nil {
					t.Fatalf("%s HEAD mismatch", name)
				}
				got, err := store.Get(ctx, o.Key)
				if err != nil || !bytes.Equal(got, o.Data) {
					t.Fatalf("%s GET mismatch: %v", name, err)
				}
			}
			for _, base := range []string{cfg.Primary.PublicBaseURL, cfg.Mirror.PublicBaseURL} {
				req, _ := http.NewRequestWithContext(ctx, http.MethodGet, URL(base, o.Key), nil)
				r, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Fatal("CDN GET failed")
				}
				got, readErr := readBounded(r.Body)
				r.Body.Close()
				if readErr != nil || r.StatusCode != 200 || !bytes.Equal(got, o.Data) {
					t.Fatalf("CDN byte mismatch (HTTP %d)", r.StatusCode)
				}
			}
			if err := s.RepairMirror(ctx, o.Key); err != nil {
				t.Fatal(err)
			}
			t.Logf("real COS/R2 Put/Head/Get, both CDN GET and mirror repair PASS: %s", o.Key)
		})
	}
	for _, base := range []string{cfg.Primary.PublicBaseURL, cfg.Mirror.PublicBaseURL} {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		req, _ := http.NewRequestWithContext(ctx, "GET", URL(base, ProbeKey)+"?probe=asset-acceptance", nil)
		r, err := http.DefaultClient.Do(req)
		if err != nil {
			cancel()
			t.Fatal("probe request failed")
		}
		b, err := readBounded(r.Body)
		r.Body.Close()
		cancel()
		h := sha256.Sum256(b)
		if err != nil || r.StatusCode != 200 || len(b) != 8192 || hex.EncodeToString(h[:]) != "a2dacdf8cbd7f21e14efa73f83610bc1f40225cae90e691788000ca82cd8004e" {
			t.Fatal("real probe hash mismatch")
		}
	}
}

// Image GET alone cannot prove that a browser can use a cross-origin SVG mask.
func TestRealDevStorageCORS(t *testing.T) {
	config := os.Getenv("GOFURRY_ASSET_DEV_CONFIG")
	if config == "" {
		t.Skip("set GOFURRY_ASSET_DEV_CONFIG for real browser-origin CORS acceptance")
	}
	if err := env.MustInitServerConfig("gofurry-admin", config); err != nil {
		t.Fatal("dev config unavailable")
	}
	cfg := env.GetServerConfig().ExternalServices.AssetStorage
	for _, base := range []string{cfg.Primary.PublicBaseURL, cfg.Mirror.PublicBaseURL} {
		if !strings.Contains(base, "assets-dev.") {
			t.Fatal("requires development CDN")
		}
		for _, key := range []string{ProbeKey, "nav/patterns/e3f8e3ed70a58cfdf4625b74f1e19563.svg"} {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			req, _ := http.NewRequestWithContext(ctx, "GET", URL(base, key), nil)
			req.Header.Set("Origin", "http://localhost:3000")
			r, err := http.DefaultClient.Do(req)
			if err != nil {
				cancel()
				t.Fatal("CORS request failed")
			}
			_, readErr := readBounded(r.Body)
			r.Body.Close()
			cancel()
			origin := r.Header.Get("Access-Control-Allow-Origin")
			if r.StatusCode != 200 || readErr != nil || (origin != "*" && origin != "http://localhost:3000") {
				t.Errorf("%s/%s: HTTP %d, missing/mismatched CORS origin %q", base, key, r.StatusCode, origin)
			}
		}
	}
}
