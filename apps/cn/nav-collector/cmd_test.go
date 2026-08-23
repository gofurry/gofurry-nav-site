package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandShowsHelpWithoutServing(t *testing.T) {
	cmd := newRootCommand()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "serve") || !strings.Contains(output.String(), "version") {
		t.Fatalf("unexpected root help: %s", output.String())
	}
}

func TestServeRequiresExplicitConfig(t *testing.T) {
	cmd := newRootCommand()
	cmd.SetArgs([]string{"serve"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "--config is required") {
		t.Fatalf("serve error = %v", err)
	}
}
