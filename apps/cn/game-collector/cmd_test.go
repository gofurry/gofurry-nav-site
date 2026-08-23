package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandShowsHelpAndKeepsManualCommands(t *testing.T) {
	cmd := newRootCommand()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"serve", "collect", "players", "all", "version"} {
		if !strings.Contains(output.String(), name) {
			t.Fatalf("root help is missing %q: %s", name, output.String())
		}
	}
	collect, _, err := cmd.Find([]string{"full"})
	if err != nil || collect.Name() != "collect" {
		t.Fatalf("full alias did not resolve to collect: cmd=%v err=%v", collect, err)
	}
}

func TestCollectorServeRequiresExplicitConfig(t *testing.T) {
	cmd := newRootCommand()
	cmd.SetArgs([]string{"serve"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "--config is required") {
		t.Fatalf("serve error = %v", err)
	}
}
