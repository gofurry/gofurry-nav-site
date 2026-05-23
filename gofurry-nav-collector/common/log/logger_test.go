package log

import (
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

func TestLoggerFormatterHumanReadableLine(t *testing.T) {
	entry := zapcore.Entry{
		Time:    time.Date(2026, 5, 23, 20, 30, 15, 123456000, time.Local),
		Level:   zapcore.InfoLevel,
		Message: "Ping 采集运行完成",
	}
	fields := map[string]interface{}{
		sFunctionEvent: "gf-nav-collector",
		sFunctionName:  "github.com/gofurry/gofurry-nav-collector/collector/ping/service.Ping",
		sFunctionLine:  123,
		"duration":     2 * time.Second,
		"protocol":     "ping",
		"targets":      42,
	}

	got := string(formatLine(entry, fields))

	for _, want := range []string{
		"2026-05-23 20:30:15.123 INFO  [gf-nav-collector] Ping 采集运行完成",
		"caller=service.Ping:123",
		"duration=2s",
		"protocol=ping",
		"targets=42",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("formatLine() missing %q in %q", want, got)
		}
	}
}

func TestLoggerFormatterQuotesHumanTextFields(t *testing.T) {
	entry := zapcore.Entry{
		Time:    time.Date(2026, 5, 23, 20, 30, 15, 0, time.Local),
		Level:   zapcore.WarnLevel,
		Message: "HTTP 探测失败",
	}
	fields := map[string]interface{}{
		"reason": "上一轮仍在运行",
		"url":    "https://example.com/a path",
	}

	got := string(formatLine(entry, fields))

	for _, want := range []string{
		"WARN  [collector] HTTP 探测失败",
		`reason=上一轮仍在运行`,
		`url="https://example.com/a path"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("formatLine() missing %q in %q", want, got)
		}
	}
}

func TestZapFieldsRoundTripToHumanLine(t *testing.T) {
	fields := encodeFields(toZapFields(map[string]interface{}{
		sFunctionEvent: "gf-nav-collector",
		"duration":     3 * time.Second,
		"protocol":     "dns",
	}))

	for _, want := range []string{
		sFunctionEvent,
		"duration",
		"protocol",
	} {
		if _, ok := fields[want]; !ok {
			t.Fatalf("encodeFields() missing %q in %#v", want, fields)
		}
	}
	if fields["duration"] != "3s" {
		t.Fatalf("duration should be human readable, got %#v", fields["duration"])
	}
}
