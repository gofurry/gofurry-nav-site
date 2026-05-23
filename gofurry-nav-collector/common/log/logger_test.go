package log

import (
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestLoggerFormatterHumanReadableLine(t *testing.T) {
	entry := &logrus.Entry{
		Time:    time.Date(2026, 5, 23, 20, 30, 15, 123456000, time.Local),
		Level:   logrus.InfoLevel,
		Message: "Ping collection run completed",
		Data: logrus.Fields{
			sFunctionEvent: "gf-nav-collector",
			sFunctionName:  "github.com/gofurry/gofurry-nav-collector/collector/ping/service.Ping",
			sFunctionLine:  123,
			"duration":     2 * time.Second,
			"protocol":     "ping",
			"targets":      42,
		},
	}

	gotBytes, err := (&LoggerFormatter{}).Format(entry)
	if err != nil {
		t.Fatalf("Format() returned error: %v", err)
	}
	got := string(gotBytes)

	for _, want := range []string{
		"2026-05-23 20:30:15.123 INFO  [gf-nav-collector] Ping collection run completed",
		"caller=service.Ping:123",
		"duration=2s",
		"protocol=ping",
		"targets=42",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("Format() missing %q in %q", want, got)
		}
	}
}

func TestLoggerFormatterQuotesHumanTextFields(t *testing.T) {
	entry := &logrus.Entry{
		Time:    time.Date(2026, 5, 23, 20, 30, 15, 0, time.Local),
		Level:   logrus.WarnLevel,
		Message: "HTTP probe failed",
		Data: logrus.Fields{
			"reason": "previous run running",
			"url":    "https://example.com/a path",
		},
	}

	gotBytes, err := (&LoggerFormatter{}).Format(entry)
	if err != nil {
		t.Fatalf("Format() returned error: %v", err)
	}
	got := string(gotBytes)

	for _, want := range []string{
		"WARN  [collector] HTTP probe failed",
		`reason="previous run running"`,
		`url="https://example.com/a path"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("Format() missing %q in %q", want, got)
		}
	}
}
