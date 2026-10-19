//go:build !linux

package systemd

import (
	"context"
	"fmt"
)

func Install(context.Context, InstallRequest) (string, error) {
	return "", fmt.Errorf("systemd installation is supported only on Linux")
}

func Uninstall(context.Context, string) error {
	return fmt.Errorf("systemd uninstallation is supported only on Linux")
}
