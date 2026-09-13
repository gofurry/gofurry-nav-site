package assets

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

type memoryStore struct {
	objects map[string]Object
	fail    bool
	puts    int
}

func (s *memoryStore) Put(_ context.Context, o Object) error {
	if s.fail {
		return errors.New("unavailable")
	}
	s.puts++
	s.objects[o.Key] = o
	return nil
}
func (s *memoryStore) Head(_ context.Context, k string) (Info, error) {
	if s.fail {
		return Info{}, errors.New("unavailable")
	}
	o, ok := s.objects[k]
	return Info{Exists: ok, Size: int64(len(o.Data)), ContentType: o.ContentType, SHA256: o.SHA256, Kind: o.Kind, CacheControl: CacheControl}, nil
}
func (s *memoryStore) Get(_ context.Context, k string) ([]byte, error) { return s.objects[k].Data, nil }
func TestPrimaryMirrorPublication(t *testing.T) {
	p, m := &memoryStore{objects: map[string]Object{}}, &memoryStore{objects: map[string]Object{}}
	s := &Service{Primary: p, Mirror: m}
	o, _ := NewObject("site-icon", 123, "ICON.ICO", []byte{0, 0, 1, 0, 1, 0})
	p.fail = true
	if _, err := s.Publish(context.Background(), o); err == nil || m.puts != 0 {
		t.Fatal("primary failure published mirror")
	}
	p.fail = false
	m.fail = true
	r, err := s.Publish(context.Background(), o)
	if err != nil || r.Primary != "ready" || r.Mirror != "failed" || len(r.Warnings) != 1 {
		t.Fatal("mirror failure blocked primary")
	}
	m.fail = false
	r, err = s.Publish(context.Background(), o)
	if err != nil || r.Mirror != "ready" || p.puts != 1 {
		t.Fatal("immutable primary was overwritten")
	}
	if err := s.RepairMirror(context.Background(), o.Key); err != nil {
		t.Fatal(err)
	}
	bad := o
	bad.SHA256 = strings.Repeat("a", 64)
	p.objects[o.Key] = bad
	if _, err := s.Publish(context.Background(), o); err == nil {
		t.Fatal("immutable collision accepted")
	}
}
func TestObjectKeysAndFormats(t *testing.T) {
	avif, err := os.ReadFile("testdata/hero.avif")
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"hero-desktop", "hero-mobile"} {
		o, err := NewObject(kind, 0, "photo.jpg", avif)
		if err != nil || !ValidKey(o.Key) || !strings.Contains(o.Key, strings.TrimPrefix(kind, "hero-")+"/") {
			t.Fatalf("invalid AVIF key: %v", err)
		}
	}
	for _, name := range []string{"site.SVG", "icon.ICO", "favicon"} {
		o, err := NewObject("site-icon", 123, name, []byte("original bytes"))
		if err != nil || !ValidKey(o.Key) || !strings.HasPrefix(o.Key, "nav/sites/123/icon/") {
			t.Fatal("icon format changed")
		}
		if name == "favicon" && strings.Contains(o.Key, ".") {
			t.Fatal("invented icon extension")
		}
	}
	if _, _, err := Validate("hero-desktop", "fake.avif", []byte("not avif")); err == nil {
		t.Fatal("fake AVIF accepted")
	}
	for _, key := range []string{"../x", "https://example.com/x", "nav/sites/0/icon/abc.svg", "nav/hero/desktop/1.avif"} {
		if ValidKey(key) {
			t.Fatal("unsafe key accepted")
		}
	}
}
func TestSVGOriginalContent(t *testing.T) {
	for _, body := range []string{
		`<?xml version="1.0"?><!DOCTYPE svg><!-- exported drawing --><svg xmlns="http://www.w3.org/2000/svg"><style>.shape{fill:black}</style><filter id="blur"><feGaussianBlur stdDeviation="2"/></filter><path class="shape" filter="url(#blur)" d="M0 0h10v10z"/></svg>`,
		`<svg onload="example()"><script/><image href="https://example.com/image.png"/><foreignObject/><animate attributeName="opacity"/></svg>`,
		`<svg><!--` + strings.Repeat("drawing metadata ", 40000) + `--></svg>`,
	} {
		data := []byte(body)
		o, err := NewObject("pattern", 0, "export.SVG", data)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(o.Data, data) || o.ContentType != "image/svg+xml" || !ValidKey(o.Key) {
			t.Fatal("SVG bytes or publication format changed")
		}
	}
	for _, data := range [][]byte{nil, make([]byte, MaxSize+1)} {
		if _, _, err := Validate("pattern", "pattern.svg", data); err == nil {
			t.Fatal("upload size limit bypassed")
		}
	}
	if _, _, err := Validate("pattern", "photo.png", []byte("image")); err == nil {
		t.Fatal("non-SVG pattern accepted")
	}
}
