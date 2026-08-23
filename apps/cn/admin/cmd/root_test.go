package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandShowsHelpWithoutServing(t *testing.T) {
	cmd := newRootCmd()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"serve", "install", "uninstall", "version", "reset-password"} {
		if !strings.Contains(output.String(), name) {
			t.Fatalf("root help is missing %q: %s", name, output.String())
		}
	}
}

func TestServeRequiresExplicitConfig(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"serve"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "config") {
		t.Fatalf("serve error = %v", err)
	}
}
