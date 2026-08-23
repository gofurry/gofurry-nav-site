package systemd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderUnitUsesForegroundServe(t *testing.T) {
	workingDirectory := filepath.Join(t.TempDir(), "gf-uptime")
	executable := filepath.Join(workingDirectory, "gf-uptime")
	configFile := filepath.Join(t.TempDir(), "server.yaml")
	content, err := renderUnit(unitSpec{
		ServiceName:      "gf-uptime",
		Description:      "GoFurry uptime status service",
		RuntimeUser:      "gofurry",
		WorkingDirectory: workingDirectory,
		Executable:       executable,
		ConfigFile:       configFile,
	})
	if err != nil {
		t.Fatal(err)
	}
	unit := string(content)
	directiveWorkingDirectory, _ := unitDirectiveValue(workingDirectory)
	quotedExecutable, _ := quoteCommandArgument(executable)
	quotedConfigFile, _ := quoteCommandArgument(configFile)
	for _, expected := range []string{
		"WorkingDirectory=" + directiveWorkingDirectory,
		"ExecStart=" + quotedExecutable + " serve --config " + quotedConfigFile,
		"Restart=on-failure",
	} {
		if !strings.Contains(unit, expected) {
			t.Fatalf("unit does not contain %q:\n%s", expected, unit)
		}
	}
	if strings.Contains(unit, `WorkingDirectory="`) || strings.Contains(unit, "--now") {
		t.Fatalf("unit contains invalid lifecycle syntax:\n%s", unit)
	}
}

func TestWriteUnitAtomicallyRequiresForce(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gf-uptime.service")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeUnitAtomically(path, []byte("new"), false); !errors.Is(err, ErrUnitExists) {
		t.Fatalf("writeUnitAtomically() error = %v", err)
	}
	if err := writeUnitAtomically(path, []byte("new"), true); err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(path)
	if string(content) != "new" {
		t.Fatalf("content = %q", content)
	}
}
