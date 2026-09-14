package main

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.yaml.in/yaml/v4"
)

func TestYAMLConfigurationCompatibility(t *testing.T) {
	type configuration struct {
		Port    string            `yaml:"port"`
		Enabled bool              `yaml:"enabled"`
		Timeout time.Duration     `yaml:"timeout"`
		Headers map[string]string `yaml:"headers"`
		Missing *bool             `yaml:"missing,omitempty"`
		Zero    int               `yaml:"zero,omitempty"`
	}
	data := []byte("port: 5432\nenabled: yes\ntimeout: 5s\nheaders:\n  language: 中文\n  toggle: on\n")
	var got configuration
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Port != "5432" || !got.Enabled || got.Timeout != 5*time.Second || got.Headers["toggle"] != "on" || got.Headers["language"] != "中文" || got.Missing != nil || got.Zero != 0 {
		t.Fatalf("typed configuration semantics changed: %+v", got)
	}
	encoded, err := yaml.Marshal(got)
	if err != nil || strings.Contains(string(encoded), "missing:") || strings.Contains(string(encoded), "zero:") {
		t.Fatalf("omitempty semantics changed: %s, %v", encoded, err)
	}
	var roundTrip configuration
	if err := yaml.Unmarshal(encoded, &roundTrip); err != nil || !reflect.DeepEqual(roundTrip, got) {
		t.Fatalf("config round trip changed: %v", err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(append(encoded, []byte("unknown: true\n")...)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&roundTrip); err == nil {
		t.Fatal("strict unknown-field rejection was lost")
	}
	if err := yaml.Unmarshal([]byte("port: 1\nport: 2\n"), &roundTrip); err == nil {
		t.Fatal("duplicate-key rejection was lost")
	}
}
