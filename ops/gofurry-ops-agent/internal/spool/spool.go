package spool

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Store struct {
	dir      string
	maxFiles int
}

func New(dir string, maxFiles int) *Store {
	return &Store{dir: dir, maxFiles: maxFiles}
}

func (s *Store) Append(body []byte) error {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return err
	}
	if err := s.prune(); err != nil {
		return err
	}
	name := "spool-" + time.Now().UTC().Format("20060102T150405.000000000") + ".jsonl"
	path := filepath.Join(s.dir, name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return nil
	}
	if _, err := file.Write(append(body, '\n')); err != nil {
		return err
	}
	return nil
}

func (s *Store) Replay(ctx context.Context, send func(context.Context, []byte) error) error {
	files, err := s.files()
	if err != nil {
		return err
	}
	for _, path := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.replayFile(ctx, path, send); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) replayFile(ctx context.Context, path string, send func(context.Context, []byte) error) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	var remaining [][]byte
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	failed := false
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		copied := append([]byte(nil), line...)
		if failed {
			remaining = append(remaining, copied)
			continue
		}
		if err := send(ctx, copied); err != nil {
			failed = true
			remaining = append(remaining, copied)
		}
	}
	if err := scanner.Err(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if len(remaining) == 0 {
		return os.Remove(path)
	}
	tmp := path + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(out)
	for _, line := range remaining {
		if _, err := writer.Write(append(line, '\n')); err != nil {
			_ = out.Close()
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (s *Store) files() ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".jsonl") {
			files = append(files, filepath.Join(s.dir, name))
		}
	}
	sort.Strings(files)
	return files, nil
}

func (s *Store) prune() error {
	if s.maxFiles <= 0 {
		return nil
	}
	files, err := s.files()
	if err != nil {
		return err
	}
	if len(files) < s.maxFiles {
		return nil
	}
	removeCount := len(files) - s.maxFiles + 1
	for i := 0; i < removeCount; i++ {
		if err := os.Remove(files[i]); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("prune spool %s: %w", files[i], err)
		}
	}
	return nil
}
