//go:build !linux

package systemd

import (
	"context"
	"strings"
	"testing"
)

func TestSystemdCommandsRejectNonLinux(t *testing.T) {
	if _, err := Install(context.Background(), InstallRequest{}); err == nil || !strings.Contains(err.Error(), "only on Linux") {
		t.Fatalf("Install() error = %v", err)
	}
	if err := Uninstall(context.Background(), "gf-uptime"); err == nil || !strings.Contains(err.Error(), "only on Linux") {
		t.Fatalf("Uninstall() error = %v", err)
	}
}
