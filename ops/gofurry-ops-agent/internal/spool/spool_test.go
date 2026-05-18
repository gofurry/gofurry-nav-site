package spool

import (
	"context"
	"errors"
	"os"
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
