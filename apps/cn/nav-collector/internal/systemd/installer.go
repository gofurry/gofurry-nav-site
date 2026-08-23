package systemd

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

const unitDirectory = "/etc/systemd/system"

var (
	ErrUnitExists    = errors.New("systemd unit already exists")
	validServiceName = regexp.MustCompile(`^[A-Za-z0-9_.@-]+$`)
	validUserName    = regexp.MustCompile(`^[A-Za-z0-9_.@-]+$`)
)

type InstallRequest struct {
	ServiceName string
	Description string
	ConfigFile  string
	Force       bool
}

type unitSpec struct {
	ServiceName      string
	Description      string
	RuntimeUser      string
	WorkingDirectory string
	Executable       string
	ConfigFile       string
}

func validateInstallRequest(request InstallRequest) error {
	if err := validateServiceName(request.ServiceName); err != nil {
		return err
	}
	if strings.TrimSpace(request.Description) == "" {
		return errors.New("service description is required")
	}
	if strings.TrimSpace(request.ConfigFile) == "" {
		return errors.New("--config is required")
	}
	return nil
}

func prepareUnit(request InstallRequest) (unitSpec, error) {
	if err := validateInstallRequest(request); err != nil {
		return unitSpec{}, err
	}

	configFile, err := canonicalExistingPath(request.ConfigFile, false)
	if err != nil {
		return unitSpec{}, fmt.Errorf("resolve config file: %w", err)
	}
	executable, err := os.Executable()
	if err != nil {
		return unitSpec{}, fmt.Errorf("resolve current executable: %w", err)
	}
	executable, err = canonicalExistingPath(executable, false)
	if err != nil {
		return unitSpec{}, fmt.Errorf("resolve current executable: %w", err)
	}
	info, err := os.Stat(executable)
	if err != nil {
		return unitSpec{}, fmt.Errorf("stat current executable: %w", err)
	}
	if info.Mode()&0o111 == 0 {
		return unitSpec{}, fmt.Errorf("current executable is not executable: %s", executable)
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		return unitSpec{}, fmt.Errorf("resolve current working directory: %w", err)
	}
	workingDirectory, err = canonicalExistingPath(workingDirectory, true)
	if err != nil {
		return unitSpec{}, fmt.Errorf("resolve current working directory: %w", err)
	}

	runtimeUser, err := resolveRuntimeUser()
	if err != nil {
		return unitSpec{}, err
	}

	return unitSpec{
		ServiceName:      request.ServiceName,
		Description:      request.Description,
		RuntimeUser:      runtimeUser,
		WorkingDirectory: workingDirectory,
		Executable:       executable,
		ConfigFile:       configFile,
	}, nil
}

func canonicalExistingPath(path string, wantDirectory bool) (string, error) {
	absolute, err := filepath.Abs(strings.TrimSpace(path))
	if err != nil {
		return "", err
	}
	canonical, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(canonical)
	if err != nil {
		return "", err
	}
	if wantDirectory && !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", canonical)
	}
	if !wantDirectory && !info.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file: %s", canonical)
	}
	return filepath.Clean(canonical), nil
}

func resolveRuntimeUser() (string, error) {
	name := strings.TrimSpace(os.Getenv("SUDO_USER"))
	if name != "" {
		resolved, err := user.Lookup(name)
		if err != nil {
			return "", fmt.Errorf("resolve SUDO_USER %q: %w", name, err)
		}
		name = resolved.Username
	} else {
		current, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("resolve current OS user: %w", err)
		}
		name = current.Username
	}
	if !validUserName.MatchString(name) {
		return "", fmt.Errorf("resolved runtime user %q is not safe for a systemd unit", name)
	}
	return name, nil
}

func validateServiceName(name string) error {
	if !validServiceName.MatchString(name) {
		return fmt.Errorf("invalid systemd service name %q", name)
	}
	return nil
}

func renderUnit(spec unitSpec) ([]byte, error) {
	if err := validateServiceName(spec.ServiceName); err != nil {
		return nil, err
	}
	if !validUserName.MatchString(spec.RuntimeUser) {
		return nil, fmt.Errorf("invalid runtime user %q", spec.RuntimeUser)
	}
	if !filepath.IsAbs(spec.WorkingDirectory) || !filepath.IsAbs(spec.Executable) || !filepath.IsAbs(spec.ConfigFile) {
		return nil, errors.New("working directory, executable, and config file must be absolute")
	}
	description, err := quoteUnitValue(spec.Description)
	if err != nil {
		return nil, fmt.Errorf("escape description: %w", err)
	}
	workingDirectory, err := quoteUnitValue(spec.WorkingDirectory)
	if err != nil {
		return nil, fmt.Errorf("escape working directory: %w", err)
	}
	executable, err := quoteUnitValue(spec.Executable)
	if err != nil {
		return nil, fmt.Errorf("escape executable: %w", err)
	}
	configFile, err := quoteUnitValue(spec.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("escape config file: %w", err)
	}

	unit := fmt.Sprintf(`[Unit]
Description=%s
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=%s
WorkingDirectory=%s
ExecStart=%s serve --config %s
Restart=on-failure
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
`, description, spec.RuntimeUser, workingDirectory, executable, configFile)
	return []byte(unit), nil
}

func quoteUnitValue(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", errors.New("value must not be empty")
	}
	var escaped strings.Builder
	escaped.Grow(len(value) + 2)
	escaped.WriteByte('"')
	for _, r := range value {
		switch r {
		case '\x00', '\n', '\r':
			return "", errors.New("value contains a forbidden control character")
		case '\t':
			escaped.WriteString(`\t`)
		case '\\':
			escaped.WriteString(`\\`)
		case '"':
			escaped.WriteString(`\"`)
		case '%':
			escaped.WriteString("%%")
		default:
			if r < 0x20 || r == 0x7f {
				return "", errors.New("value contains a forbidden control character")
			}
			escaped.WriteRune(r)
		}
	}
	escaped.WriteByte('"')
	return escaped.String(), nil
}

func writeUnitAtomically(path string, content []byte, force bool) error {
	directory := filepath.Dir(path)
	if !force {
		if _, err := os.Lstat(path); err == nil {
			return fmt.Errorf("%w: %s (use --force to replace it)", ErrUnitExists, path)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("inspect existing unit: %w", err)
		}
	}

	temporary, err := os.CreateTemp(directory, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary unit: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if err = temporary.Chmod(0o644); err == nil {
		_, err = temporary.Write(content)
	}
	if err == nil {
		err = temporary.Sync()
	}
	closeErr := temporary.Close()
	if err != nil {
		return fmt.Errorf("write temporary unit: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("close temporary unit: %w", closeErr)
	}

	if force {
		err = os.Rename(temporaryPath, path)
	} else {
		err = os.Link(temporaryPath, path)
	}
	if err != nil {
		if !force && os.IsExist(err) {
			return fmt.Errorf("%w: %s (use --force to replace it)", ErrUnitExists, path)
		}
		return fmt.Errorf("publish systemd unit: %w", err)
	}
	if directoryHandle, openErr := os.Open(directory); openErr == nil {
		_ = directoryHandle.Sync()
		_ = directoryHandle.Close()
	}
	return nil
}
