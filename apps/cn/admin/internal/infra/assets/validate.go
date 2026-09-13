package assets

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"io"
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
		if len(data) > 512<<10 {
			return "", "", errors.New("pattern exceeds 512 KiB")
		}
		if err := ValidateSVG(data); err != nil {
			return "", "", err
		}
		return "image/svg+xml", ".svg", nil
	}
	return "", "", errors.New("unsupported asset kind")
}

// Parse XML and accept a self-contained geometry subset suitable for CSS masks.
// CSS and animation are excluded, so escapes cannot hide external references.
func ValidateSVG(data []byte) error {
	allowed := map[string]bool{}
	for _, name := range strings.Fields("svg g path rect circle ellipse line polyline polygon defs use symbol pattern mask clipPath title desc") {
		allowed[name] = true
	}
	d := xml.NewDecoder(bytes.NewReader(data))
	depth, roots, tokens := 0, 0, 0
	for {
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.New("invalid SVG XML")
		}
		tokens++
		if tokens > 50000 {
			return errors.New("SVG too complex")
		}
		switch t := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
				if t.Name.Local != "svg" {
					return errors.New("SVG root required")
				}
			}
			depth++
			if depth > 64 || !allowed[t.Name.Local] || (t.Name.Space != "" && t.Name.Space != "http://www.w3.org/2000/svg") {
				return errors.New("unsafe SVG element")
			}
			for _, a := range t.Attr {
				name := strings.ToLower(a.Name.Local)
				value := strings.ToLower(strings.TrimSpace(a.Value))
				if a.Name.Space == "xmlns" || name == "xmlns" {
					continue
				}
				if strings.HasPrefix(name, "on") || name == "style" || name == "base" || strings.ContainsAny(value, `\`) || strings.Contains(value, "@import") || strings.Contains(value, ":") {
					return errors.New("unsafe SVG attribute")
				}
				if name == "href" && !strings.HasPrefix(value, "#") {
					return errors.New("SVG external reference is forbidden")
				}
				if strings.Contains(value, "url(") && !regexp.MustCompile(`^url\(#[a-z0-9_-]+\)$`).MatchString(value) {
					return errors.New("SVG external paint is forbidden")
				}
			}
		case xml.EndElement:
			depth--
		case xml.Directive:
			return errors.New("SVG directives are forbidden")
		case xml.ProcInst:
			if t.Target != "xml" {
				return errors.New("SVG processing instruction is forbidden")
			}
		case xml.CharData:
			if depth == 0 && strings.TrimSpace(string(t)) != "" {
				return errors.New("invalid SVG text")
			}
		}
	}
	if roots != 1 || depth != 0 {
		return errors.New("one complete SVG is required")
	}
	return nil
}
