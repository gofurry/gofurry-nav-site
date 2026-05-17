package contentsync

import (
	"fmt"
	"io"
	"strings"
)

const (
	maxMarkdownBytes  = 2 * 1024 * 1024
	maxErrorBodyBytes = 64 * 1024
)

func readLimitedString(body io.Reader, limit int64) (string, error) {
	if body == nil {
		return "", nil
	}
	limited := &io.LimitedReader{R: body, N: limit + 1}
	data, err := io.ReadAll(limited)
	if err != nil {
		return "", err
	}
	if int64(len(data)) > limit {
		return "", fmt.Errorf("response body exceeds %d bytes", limit)
	}
	return string(data), nil
}

func readErrorBody(body io.Reader) string {
	text, err := readLimitedString(body, maxErrorBodyBytes)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(text)
}
