package collaboration

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gofurry/gofurry-admin/pkg/common"
)

var digits = regexp.MustCompile(`^[0-9]+$`)
var hostname = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9.-]*[a-z0-9])?$`)

// CanonicalSource is deliberately offline. Unknown Game/Other sources retain
// their text and have no deduplication key; no guessed matches or HTTP fetches.
func CanonicalSource(kind, source string) *string {
	source = strings.TrimSpace(source)
	if source == "" {
		return nil
	}
	if kind == "game" {
		candidate := source
		if strings.HasPrefix(strings.ToLower(source), "steam:") {
			candidate = source[6:]
		}
		if u, err := url.Parse(source); err == nil && (u.Scheme == "https" || u.Scheme == "http") && strings.EqualFold(u.Hostname(), "store.steampowered.com") && u.User == nil {
			parts := strings.Split(strings.Trim(u.Path, "/"), "/")
			if len(parts) >= 2 && parts[0] == "app" {
				candidate = parts[1]
			}
		}
		if digits.MatchString(candidate) {
			id, err := strconv.ParseInt(candidate, 10, 64)
			if err == nil && id > 0 && id <= 9007199254740991 {
				return ptr("steam:" + strconv.FormatInt(id, 10))
			}
		}
	}
	if kind == "site" {
		raw := source
		if !strings.Contains(raw, "://") {
			raw = "https://" + raw
		}
		u, err := url.Parse(raw)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil {
			return nil
		}
		host := strings.TrimPrefix(strings.TrimRight(strings.ToLower(u.Hostname()), "."), "www.")
		if len(host) <= 253 && strings.Contains(host, ".") && !strings.Contains(host, "..") && hostname.MatchString(host) {
			return ptr("host:" + host)
		}
	}
	return nil
}

func normalize(in Input) (Input, *string, common.Error) {
	in.Title, in.Source, in.Note = strings.TrimSpace(in.Title), strings.TrimSpace(in.Source), strings.TrimSpace(in.Note)
	if in.Priority == "" {
		in.Priority = "normal"
	}
	if !oneOf(in.Kind, "game", "site", "other") {
		return in, nil, common.NewValidationError("kind must be game, site or other")
	}
	if !oneOf(in.Priority, "normal", "high") {
		return in, nil, common.NewValidationError("priority must be normal or high")
	}
	if in.Title == "" && in.Source == "" {
		return in, nil, common.NewValidationError("title or source is required")
	}
	if strings.ContainsRune(in.Title+in.Source+in.Note, 0) || utf8.RuneCountInString(in.Title) > 500 || utf8.RuneCountInString(in.Source) > 2048 || utf8.RuneCountInString(in.Note) > 10000 {
		return in, nil, common.NewValidationError("title/source/note exceeds input limits")
	}
	return in, CanonicalSource(in.Kind, in.Source), nil
}
func ptr[T any](value T) *T { return &value }
func nullable(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
func text(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
func oneOf(value string, options ...string) bool {
	for _, option := range options {
		if value == option {
			return true
		}
	}
	return false
}
