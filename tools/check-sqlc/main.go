// Command check-sqlc compares temporary sqlc output with the working-tree contract.
// It never generates into the repository, including when drift is detected.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"go.yaml.in/yaml/v4"
)

func main() {
	rootFlag := flag.String("root", "..", "repository root")
	flag.Parse()
	if err := run(*rootFlag); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("sqlc generated code matches the working-tree contract (repository unchanged)")
}

func run(root string) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	temporary, err := os.MkdirTemp("", "gofurry-check-sqlc-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)
	data, err := os.ReadFile(filepath.Join(root, "sqlc.yaml"))
	if err != nil {
		return err
	}
	config, outputs, err := temporaryConfig(data, root, temporary)
	if err != nil {
		return err
	}
	configPath := filepath.Join(temporary, "sqlc.yaml")
	if err := os.WriteFile(configPath, config, 0o600); err != nil {
		return err
	}
	generate := exec.Command("go", "tool", "sqlc", "generate", "-f", configPath)
	generate.Dir = filepath.Join(root, "tools")
	generate.Stdout, generate.Stderr = os.Stdout, os.Stderr
	if err := generate.Run(); err != nil {
		return err
	}
	for _, output := range outputs {
		if err := compareDirectories(filepath.Join(root, output), filepath.Join(temporary, output)); err != nil {
			return fmt.Errorf("%s: %w; run task generate:sqlc", output, err)
		}
	}
	return nil
}

// Retain relative paths and copy only schema/query inputs. sqlc's path handling
// cannot use Windows drive-qualified input paths under a temporary config.
func temporaryConfig(data []byte, root, temporary string) ([]byte, []string, error) {
	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, nil, err
	}
	queries, ok := config["sql"].([]any)
	if !ok || len(queries) == 0 {
		return nil, nil, fmt.Errorf("sqlc config has no SQL entries")
	}
	var outputs []string
	copied := map[string]bool{}
	copyInput := func(name string) error {
		if !filepath.IsLocal(name) || filepath.Clean(name) == "." {
			return fmt.Errorf("sqlc input must be inside the repository: %q", name)
		}
		if copied[name] {
			return nil
		}
		source, target := filepath.Join(root, name), filepath.Join(temporary, name)
		info, err := os.Stat(source)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		if info.IsDir() {
			err = os.CopyFS(target, os.DirFS(source))
		} else {
			var content []byte
			content, err = os.ReadFile(source)
			if err == nil {
				err = os.WriteFile(target, content, 0o600)
			}
		}
		copied[name] = err == nil
		return err
	}
	for _, value := range queries {
		entry, ok := value.(map[string]any)
		if !ok {
			return nil, nil, fmt.Errorf("invalid sqlc SQL entry")
		}
		for _, key := range []string{"schema", "queries"} {
			switch paths := entry[key].(type) {
			case string:
				if err := copyInput(paths); err != nil {
					return nil, nil, err
				}
			case []any:
				for _, path := range paths {
					name, ok := path.(string)
					if !ok {
						return nil, nil, fmt.Errorf("invalid sqlc %s path", key)
					}
					if err := copyInput(name); err != nil {
						return nil, nil, err
					}
				}
			default:
				return nil, nil, fmt.Errorf("invalid sqlc %s paths", key)
			}
		}
		gen, _ := entry["gen"].(map[string]any)
		goGen, _ := gen["go"].(map[string]any)
		output, _ := goGen["out"].(string)
		if output == "" || !filepath.IsLocal(output) || filepath.Clean(output) == "." {
			return nil, nil, fmt.Errorf("sqlc Go output must be inside the repository: %q", output)
		}
		outputs = append(outputs, output)
	}
	encoded, err := yaml.Marshal(config)
	return encoded, outputs, err
}

func compareDirectories(current, generated string) error {
	read := func(root string) (map[string][]byte, error) {
		files := map[string][]byte{}
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if os.IsNotExist(err) && path == root {
				return nil
			}
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			name, err := filepath.Rel(root, path)
			files[name] = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
			return err
		})
		return files, err
	}
	want, err := read(generated)
	if err != nil {
		return err
	}
	got, err := read(current)
	if err != nil {
		return err
	}
	var changed []string
	for name, content := range want {
		if other, exists := got[name]; !exists || !bytes.Equal(content, other) {
			changed = append(changed, name)
		}
	}
	for name := range got {
		if _, exists := want[name]; !exists {
			changed = append(changed, name)
		}
	}
	if len(changed) != 0 {
		sort.Strings(changed)
		return fmt.Errorf("generated code drift: %s", strings.Join(changed, ", "))
	}
	return nil
}
