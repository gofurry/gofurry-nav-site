package assets

import (
	"encoding/binary"
	"errors"
	"mime"
	"net/http"
	"path"
	"regexp"
	"strings"
)

var objectKeyPattern = regexp.MustCompile(`^(nav/sites/[1-9][0-9]*/icon/[a-f0-9]{32}(\.[a-z0-9]{1,16})?|nav/hero/(desktop|mobile)/[a-f0-9]{32}\.avif|nav/patterns/[a-f0-9]{32}\.svg)$`)

func ValidKey(key string) bool { return key == ProbeKey || objectKeyPattern.MatchString(key) }

func Validate(kind, filename string, data []byte) (string, string, error) {
	if len(data) == 0 || len(data) > MaxSize {
		return "", "", errors.New("empty file or upload exceeds 5 MiB")
	}
	switch kind {
	case "site-icon":
		if len(data) > 2<<20 {
			return "", "", errors.New("site icon exceeds 2 MiB")
		}
		ext := strings.ToLower(path.Ext(strings.ReplaceAll(filename, `\`, "/")))
		if ext != "" && !regexp.MustCompile(`^\.[a-z0-9]{1,16}$`).MatchString(ext) {
			return "", "", errors.New("invalid file extension")
		}
		ct := http.DetectContentType(data)
		if ct == "application/octet-stream" || strings.HasPrefix(ct, "text/plain") {
			if guessed := mime.TypeByExtension(ext); guessed != "" {
				ct = guessed
			}
		}
		return ct, ext, nil
	case "hero-desktop", "hero-mobile":
		if len(data) < 24 || string(data[4:8]) != "ftyp" {
			return "", "", errors.New("hero must be AVIF")
		}
		n := int(binary.BigEndian.Uint32(data[:4]))
		if n < 16 || n > len(data) {
			return "", "", errors.New("invalid AVIF file type box")
		}
		avif := string(data[8:12]) == "avif" || string(data[8:12]) == "avis"
		for i := 16; i+4 <= n; i += 4 {
			avif = avif || string(data[i:i+4]) == "avif" || string(data[i:i+4]) == "avis"
		}
		if !avif {
			return "", "", errors.New("hero must be AVIF")
		}
		return "image/avif", ".avif", nil
	case "pattern":
		if !strings.EqualFold(path.Ext(filename), ".svg") {
			return "", "", errors.New("pattern must be an SVG file")
		}
		// Preserve administrator-supplied SVG bytes without content filtering.
		return "image/svg+xml", ".svg", nil
	}
	return "", "", errors.New("unsupported asset kind")
}
