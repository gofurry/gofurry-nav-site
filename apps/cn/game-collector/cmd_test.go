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
	for _, name := range []string{"serve", "install", "uninstall", "collect", "players", "all", "backfill-first-available", "version"} {
		if !strings.Contains(output.String(), name) {
			t.Fatalf("root help is missing %q: %s", name, output.String())
		}
	}
	collect, _, err := cmd.Find([]string{"full"})
	if err != nil || collect.Name() != "collect" {
		t.Fatalf("full alias did not resolve to collect: cmd=%v err=%v", collect, err)
	}
}

func TestInstallRequiresExplicitConfigAndExposesForce(t *testing.T) {
	cmd := newRootCommand()
	install, _, err := cmd.Find([]string{"install"})
	if err != nil || install.Flags().Lookup("force") == nil {
		t.Fatalf("install --force is unavailable: command=%v err=%v", install, err)
	}
	cmd.SetArgs([]string{"install"})
	if err = cmd.Execute(); err == nil || !strings.Contains(err.Error(), "--config is required") {
		t.Fatalf("install error = %v", err)
	}
}

func TestCollectorServeRequiresExplicitConfig(t *testing.T) {
	cmd := newRootCommand()
	cmd.SetArgs([]string{"serve"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "--config is required") {
		t.Fatalf("serve error = %v", err)
	}
}
