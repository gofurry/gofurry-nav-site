package ingest

import (
	"strings"
	"unicode/utf8"
)

type Splitter struct {
	ChunkSize    int
	ChunkOverlap int
}

func NewSplitter(size, overlap int) Splitter {
	if size <= 0 {
		size = 700
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= size {
		overlap = size / 4
	}
	return Splitter{ChunkSize: size, ChunkOverlap: overlap}
}

func (s Splitter) Split(text string) []string {
	text = normalizeText(text)
	if text == "" {
		return nil
	}
	if runeLen(text) <= s.ChunkSize {
		return []string{text}
	}

	paragraphs := strings.Split(text, "\n\n")
	chunks := []string{}
	current := ""
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		if runeLen(paragraph) > s.ChunkSize {
			if current != "" {
				chunks = append(chunks, current)
				current = ""
			}
			chunks = append(chunks, s.hardSplit(paragraph)...)
			continue
		}
		candidate := paragraph
		if current != "" {
			candidate = current + "\n\n" + paragraph
		}
		if runeLen(candidate) <= s.ChunkSize {
			current = candidate
			continue
		}
		if current != "" {
			chunks = append(chunks, current)
		}
		current = withOverlap(lastRunes(current, s.ChunkOverlap), paragraph)
	}
	if strings.TrimSpace(current) != "" {
		chunks = append(chunks, current)
	}
	return chunks
}

func (s Splitter) hardSplit(text string) []string {
	runes := []rune(text)
	chunks := []string{}
	step := s.ChunkSize - s.ChunkOverlap
	if step <= 0 {
		step = s.ChunkSize
	}
	for start := 0; start < len(runes); start += step {
		end := start + s.ChunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, strings.TrimSpace(string(runes[start:end])))
		if end == len(runes) {
			break
		}
	}
	return chunks
}

func normalizeText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func runeLen(text string) int {
	return utf8.RuneCountInString(text)
}

func lastRunes(text string, n int) string {
	if n <= 0 || text == "" {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= n {
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(string(runes[len(runes)-n:]))
}

func withOverlap(overlap, text string) string {
	if strings.TrimSpace(overlap) == "" {
		return text
	}
	return strings.TrimSpace(overlap) + "\n\n" + text
}
