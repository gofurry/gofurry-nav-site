package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRecommendationsRebuildHelp(t *testing.T) {
	cmd := newRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"recommendations", "rebuild", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "--config") {
		t.Fatal("rebuild must use explicit config")
	}
}
