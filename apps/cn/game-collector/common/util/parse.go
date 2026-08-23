package util

import (
	"bytes"

	"github.com/yuin/goldmark"
)

// Markdown 转 HTML
func MarkdownToHTML(markdownContent string) (string, error) {
	var buf bytes.Buffer
	md := goldmark.New()
	err := md.Convert([]byte(markdownContent), &buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
