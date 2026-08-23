//go:build !linux

package systemd

import (
	"context"
	"strings"
	"testing"
)

func TestNonLinuxOperationsReturnClearErrors(t *testing.T) {
	if _, err := Install(context.Background(), InstallRequest{}); err == nil || !strings.Contains(err.Error(), "only on Linux") {
		t.Fatalf("Install() error = %v, want unsupported-platform error", err)
	}
	if err := Uninstall(context.Background(), "gf-nav"); err == nil || !strings.Contains(err.Error(), "only on Linux") {
		t.Fatalf("Uninstall() error = %v, want unsupported-platform error", err)
	}
}
