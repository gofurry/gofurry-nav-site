package systemd

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRenderUnitEscapesDeploymentValues(t *testing.T) {
	workingDirectory := filepath.Join(t.TempDir(), "go furry", "100%")
	executable := filepath.Join(workingDirectory, `gf"nav%`)
	configFile := filepath.Join(t.TempDir(), "go furry", "server%.yaml")
	content, err := renderUnit(unitSpec{
		ServiceName:      "gf-nav",
		Description:      `GoFurry "Nav" 100%`,
		RuntimeUser:      "deploy-user",
		WorkingDirectory: workingDirectory,
		Executable:       executable,
		ConfigFile:       configFile,
	})
	if err != nil {
		t.Fatalf("renderUnit() error = %v", err)
	}
	unit := string(content)
	directiveWorkingDirectory, _ := unitDirectiveValue(workingDirectory)
	quotedExecutable, _ := quoteCommandArgument(executable)
	quotedConfigFile, _ := quoteCommandArgument(configFile)
	for _, expected := range []string{
		`Description=GoFurry "Nav" 100%%`,
		`User=deploy-user`,
		`WorkingDirectory=` + directiveWorkingDirectory,
		`ExecStart=` + quotedExecutable + ` serve --config ` + quotedConfigFile,
		`Restart=on-failure`,
	} {
		if !strings.Contains(unit, expected) {
			t.Errorf("unit does not contain %q:\n%s", expected, unit)
		}
	}
	if strings.Contains(unit, "ExecStart=/bin/sh") || strings.Contains(unit, "--now") {
		t.Fatalf("unit must execute the binary directly and must not start it during install:\n%s", unit)
	}
	if strings.Contains(unit, `WorkingDirectory="`) {
		t.Fatalf("WorkingDirectory must not use ExecStart-style argument quoting:\n%s", unit)
	}
}

func TestUnitDirectiveValueDoesNotQuoteAbsolutePath(t *testing.T) {
	value := `/home/gofurry/gfs/backend/gf admin/"100%"`
	got, err := unitDirectiveValue(value)
	if err != nil {
		t.Fatal(err)
	}
	if want := `/home/gofurry/gfs/backend/gf admin/"100%%"`; got != want {
		t.Fatalf("unitDirectiveValue() = %q, want %q", got, want)
	}
	if strings.HasPrefix(got, `"`) && strings.HasSuffix(got, `"`) {
		t.Fatalf("unit directive value was incorrectly wrapped in quotes: %q", got)
	}
}

func TestRenderUnitRejectsUnsafeValues(t *testing.T) {
	tests := []unitSpec{
		{ServiceName: "../bad", Description: "bad", RuntimeUser: "deploy", WorkingDirectory: "/srv", Executable: "/srv/app", ConfigFile: "/etc/app.yaml"},
		{ServiceName: "good", Description: "bad\nvalue", RuntimeUser: "deploy", WorkingDirectory: "/srv", Executable: "/srv/app", ConfigFile: "/etc/app.yaml"},
		{ServiceName: "good", Description: "good", RuntimeUser: "bad user", WorkingDirectory: "/srv", Executable: "/srv/app", ConfigFile: "/etc/app.yaml"},
		{ServiceName: "good", Description: "good", RuntimeUser: "deploy", WorkingDirectory: "relative", Executable: "/srv/app", ConfigFile: "/etc/app.yaml"},
	}
	for _, spec := range tests {
		if _, err := renderUnit(spec); err == nil {
			t.Errorf("renderUnit(%+v) succeeded, want error", spec)
		}
	}
}

func TestWriteUnitAtomicallyRefusesExistingUnit(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "app.service")
	if err := os.WriteFile(path, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeUnitAtomically(path, []byte("replacement"), false); !errors.Is(err, ErrUnitExists) {
		t.Fatalf("writeUnitAtomically() error = %v, want ErrUnitExists", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "original" {
		t.Fatalf("existing unit changed to %q", content)
	}
}

func TestWriteUnitAtomicallyForceReplacesUnit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows os.Rename does not replace an existing destination")
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "app.service")
	if err := os.WriteFile(path, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeUnitAtomically(path, []byte("replacement"), true); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "replacement" {
		t.Fatalf("unit content = %q, want replacement", content)
	}
}

func TestValidateInstallRequestRequiresExplicitConfig(t *testing.T) {
	err := validateInstallRequest(InstallRequest{ServiceName: "gf-nav", Description: "GoFurry Nav"})
	if err == nil || !strings.Contains(err.Error(), "--config") {
		t.Fatalf("validateInstallRequest() error = %v, want explicit config error", err)
	}
}

func TestPrepareUnitRejectsMissingConfigBeforeOtherResolution(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.yaml")
	_, err := prepareUnit(InstallRequest{
		ServiceName: "gf-nav",
		Description: "GoFurry Nav",
		ConfigFile:  missing,
	})
	if err == nil || !strings.Contains(err.Error(), "resolve config file") {
		t.Fatalf("prepareUnit() error = %v, want missing config error", err)
	}
}
