package main

import (
	"os"
	"path/filepath"
	"testing"

	"go.yaml.in/yaml/v4"
)

func TestTemporaryConfigIsolatesOutput(t *testing.T) {
	root, temp := t.TempDir(), t.TempDir()
	for _, path := range []string{"db/migrations", "app/queries"} {
		if err := os.MkdirAll(filepath.Join(root, path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, path, "query.sql"), []byte("SELECT 1;"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	data, outputs, err := temporaryConfig([]byte("version: '2'\nsql:\n- engine: postgresql\n  schema: [db/migrations]\n  queries: app/queries\n  gen:\n    go:\n      out: app/sqlc\n      sql_package: pgx/v5\n"), root, temp)
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	entry := config["sql"].([]any)[0].(map[string]any)
	goGen := entry["gen"].(map[string]any)["go"].(map[string]any)
	if len(outputs) != 1 || outputs[0] != "app/sqlc" || goGen["out"] != "app/sqlc" || goGen["sql_package"] != "pgx/v5" || entry["queries"] != "app/queries" {
		t.Fatalf("incorrect isolated configuration: %s", data)
	}
	if copied, err := os.ReadFile(filepath.Join(temp, "app/queries/query.sql")); err != nil || string(copied) != "SELECT 1;" {
		t.Fatalf("query input was not copied: %q, %v", copied, err)
	}
}

func TestDriftComparisonNeverWritesSource(t *testing.T) {
	for _, state := range []string{"same", "changed", "missing", "extra"} {
		t.Run(state, func(t *testing.T) {
			current, generated := t.TempDir(), t.TempDir()
			write := func(root, name, body string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			write(generated, "query.go", "package sqlc\n")
			if state != "missing" {
				write(current, "query.go", "package sqlc\r\n")
			}
			if state == "changed" {
				write(current, "query.go", "// preserve local edit\n")
			}
			if state == "extra" {
				write(current, "stale.go", "// stale generated file\n")
			}
			before, _ := os.ReadFile(filepath.Join(current, "query.go"))
			err := compareDirectories(current, generated)
			if (err == nil) != (state == "same") {
				t.Fatalf("comparison for %s: %v", state, err)
			}
			after, _ := os.ReadFile(filepath.Join(current, "query.go"))
			if string(before) != string(after) {
				t.Fatal("comparison changed source")
			}
		})
	}
}

func TestTemporaryConfigRejectsEscapingOutput(t *testing.T) {
	for _, output := range []string{"../escape", ".", ""} {
		root := t.TempDir()
		for _, dir := range []string{"db", "queries"} {
			if err := os.Mkdir(filepath.Join(root, dir), 0o700); err != nil {
				t.Fatal(err)
			}
		}
		_, _, err := temporaryConfig([]byte("sql:\n- schema: db\n  queries: queries\n  gen:\n    go:\n      out: '"+output+"'\n"), root, t.TempDir())
		if err == nil {
			t.Fatalf("output %q was not rejected", output)
		}
	}
}
