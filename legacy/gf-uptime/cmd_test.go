package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandLifecycle(t *testing.T) {
	command := newRootCommand()
	for _, name := range []string{"serve", "install", "uninstall", "version"} {
		if found, _, err := command.Find([]string{name}); err != nil || found == nil {
			t.Fatalf("command %q unavailable: %v", name, err)
		}
	}
	install, _, err := command.Find([]string{"install"})
	if err != nil || install.Flags().Lookup("force") == nil {
		t.Fatalf("install --force unavailable: %v", err)
	}
}

func TestVersionDoesNotRequireConfig(t *testing.T) {
	command := newRootCommand()
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetArgs([]string{"version"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "gf-uptime V1.0.0") {
		t.Fatalf("version output = %q", output.String())
	}
}

func TestServeRequiresConfig(t *testing.T) {
	command := newRootCommand()
	command.SetArgs([]string{"serve"})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "--config is required") {
		t.Fatalf("serve error = %v", err)
	}
}
