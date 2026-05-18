package spool

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplayKeepsFailedAndRemovesSent(t *testing.T) {
	dir := t.TempDir()
	store := New(dir, 10)
	if err := store.Append([]byte(`{"n":1}`)); err != nil {
		t.Fatal(err)
	}
	if err := store.Append([]byte(`{"n":2}`)); err != nil {
		t.Fatal(err)
	}
	calls := 0
	err := store.Replay(context.Background(), func(ctx context.Context, body []byte) error {
		calls++
		if calls == 1 {
			return errors.New("temporary")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("expected remaining spool file")
	}

	if err := store.Replay(context.Background(), func(ctx context.Context, body []byte) error {
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	files, err = os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatalf("expected empty spool, got %d files", len(files))
	}
}

func TestReplayQuarantinesInvalidJSONAndContinues(t *testing.T) {
	dir := t.TempDir()
	store := New(dir, 10)
	bad := filepath.Join(dir, "spool-20260519T010000.000000000.jsonl")
	good := filepath.Join(dir, "spool-20260519T010001.000000000.jsonl")
	if err := os.WriteFile(bad, []byte("{bad-json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(good, []byte("{\"n\":1}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var sent [][]byte
	if err := store.Replay(context.Background(), func(ctx context.Context, body []byte) error {
		sent = append(sent, append([]byte(nil), body...))
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(sent) != 1 || string(sent[0]) != "{\"n\":1}" {
		t.Fatalf("expected later valid spool line to send, got %#v", sent)
	}
	assertBadFile(t, dir)
}

func TestReplayQuarantinesOversizedLineAndContinues(t *testing.T) {
	dir := t.TempDir()
	store := New(dir, 10)
	bad := filepath.Join(dir, "spool-20260519T020000.000000000.jsonl")
	good := filepath.Join(dir, "spool-20260519T020001.000000000.jsonl")
	oversized := append(bytes.Repeat([]byte("x"), maxReplayLineBytes+1), '\n')
	if err := os.WriteFile(bad, oversized, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(good, []byte("{\"n\":2}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var sent [][]byte
	if err := store.Replay(context.Background(), func(ctx context.Context, body []byte) error {
		sent = append(sent, append([]byte(nil), body...))
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(sent) != 1 || string(sent[0]) != "{\"n\":2}" {
		t.Fatalf("expected later valid spool line to send, got %#v", sent)
	}
	assertBadFile(t, dir)
}

func assertBadFile(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	badFiles := 0
	jsonlFiles := 0
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".bad") {
			badFiles++
		}
		if strings.HasSuffix(name, ".jsonl") {
			jsonlFiles++
		}
	}
	if badFiles == 0 {
		t.Fatalf("expected quarantined .bad file, got %#v", entries)
	}
	if jsonlFiles != 0 {
		t.Fatalf("expected no remaining jsonl files, got %#v", entries)
	}
}
