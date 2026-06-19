package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTestExampleConfigName(t *testing.T) {
	got := testExampleConfigName("server.yaml")
	if got != "server.example.yaml" {
		t.Fatalf("example config name = %q, want %q", got, "server.example.yaml")
	}
}

func TestLocalConfigCandidatesWalksParents(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "apps", "game", "v2", "service")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})

	if err := os.Chdir(nested); err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(root, "conf", "server.yaml")
	for _, got := range localConfigCandidates("server.yaml") {
		if got == want {
			return
		}
	}

	t.Fatalf("local config candidates did not include parent config %q", want)
}

func TestTestFallbackLoadsParentExampleConfig(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "apps", "game", "v2", "service")
	confDir := filepath.Join(root, "conf")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(confDir, 0o755); err != nil {
		t.Fatal(err)
	}

	exampleConfig := strings.Join([]string{
		"cluster_id: 9",
		"server:",
		"  memory_limit: 1",
		"  gc_percent: 1000",
		"thread:",
		"  event_publish_thread: 1",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(confDir, "server.example.yaml"), []byte(exampleConfig), 0o644); err != nil {
		t.Fatal(err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})

	if err := os.Chdir(nested); err != nil {
		t.Fatal(err)
	}

	cfg := new(serverConfig)
	for _, file := range localConfigCandidates(testExampleConfigName("server.yaml")) {
		if tryLoadConfig(file, cfg) {
			if cfg.ClusterId != 9 {
				t.Fatalf("cluster_id = %d, want 9", cfg.ClusterId)
			}
			return
		}
	}

	t.Fatal("example config fallback was not loaded")
}
