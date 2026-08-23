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
	for _, name := range []string{"serve", "install", "uninstall", "version"} {
		if !strings.Contains(output.String(), name) {
			t.Fatalf("root help is missing %q: %s", name, output.String())
		}
	}
}

func TestServeRequiresExplicitConfig(t *testing.T) {
	cmd := newRootCommand()
	cmd.SetArgs([]string{"serve"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "--config is required") {
		t.Fatalf("serve error = %v", err)
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
