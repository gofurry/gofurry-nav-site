//go:build linux

package systemd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Install(ctx context.Context, request InstallRequest) (string, error) {
	spec, err := prepareUnit(request)
	if err != nil {
		return "", err
	}
	content, err := renderUnit(spec)
	if err != nil {
		return "", err
	}
	systemctl, err := verifySystemd(ctx)
	if err != nil {
		return "", err
	}
	if os.Geteuid() != 0 {
		return "", errorsNeedRoot()
	}
	unitName := request.ServiceName + ".service"
	unitPath := filepath.Join(unitDirectory, unitName)
	if err = writeUnitAtomically(unitPath, content, request.Force); err != nil {
		return "", err
	}
	if err = runSystemctl(ctx, systemctl, "daemon-reload"); err != nil {
		return "", err
	}
	if err = runSystemctl(ctx, systemctl, "enable", unitName); err != nil {
		return "", err
	}
	return unitPath, nil
}

func Uninstall(ctx context.Context, serviceName string) error {
	if err := validateServiceName(serviceName); err != nil {
		return err
	}
	systemctl, err := verifySystemd(ctx)
	if err != nil {
		return err
	}
	if os.Geteuid() != 0 {
		return errorsNeedRoot()
	}
	unitName := serviceName + ".service"
	if active, statusErr := systemctlState(ctx, systemctl, "is-active", unitName); statusErr != nil {
		return statusErr
	} else if active {
		if err = runSystemctl(ctx, systemctl, "stop", unitName); err != nil {
			return err
		}
	}
	if enabled, statusErr := systemctlState(ctx, systemctl, "is-enabled", unitName); statusErr != nil {
		return statusErr
	} else if enabled {
		if err = runSystemctl(ctx, systemctl, "disable", unitName); err != nil {
			return err
		}
	}
	unitPath := filepath.Join(unitDirectory, unitName)
	if err = os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove systemd unit %s: %w", unitPath, err)
	}
	if err = runSystemctl(ctx, systemctl, "daemon-reload"); err != nil {
		return err
	}
	_ = runSystemctl(ctx, systemctl, "reset-failed", unitName)
	return nil
}

func verifySystemd(ctx context.Context) (string, error) {
	systemctl, err := exec.LookPath("systemctl")
	if err != nil {
		return "", fmt.Errorf("systemd is unavailable: systemctl was not found: %w", err)
	}
	command := exec.CommandContext(ctx, systemctl, "show", "--property=Version", "--value")
	if output, runErr := command.CombinedOutput(); runErr != nil {
		return "", fmt.Errorf("systemd is unavailable or not running: %w%s", runErr, formatCommandOutput(output))
	}
	return systemctl, nil
}

func runSystemctl(ctx context.Context, systemctl string, arguments ...string) error {
	command := exec.CommandContext(ctx, systemctl, arguments...)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl %s failed: %w%s", strings.Join(arguments, " "), err, formatCommandOutput(output))
	}
	return nil
}

func systemctlState(ctx context.Context, systemctl, action, unitName string) (bool, error) {
	command := exec.CommandContext(ctx, systemctl, action, "--quiet", unitName)
	output, err := command.CombinedOutput()
	if err == nil {
		return true, nil
	}
	if _, ok := err.(*exec.ExitError); ok && len(output) == 0 {
		return false, nil
	}
	return false, fmt.Errorf("systemctl %s %s failed: %w%s", action, unitName, err, formatCommandOutput(output))
}

func formatCommandOutput(output []byte) string {
	text := strings.TrimSpace(string(output))
	if text == "" {
		return ""
	}
	return ": " + text
}

func errorsNeedRoot() error {
	return fmt.Errorf("installing or uninstalling a systemd unit requires root privileges; rerun with sudo")
}
