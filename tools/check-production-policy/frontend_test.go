package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"go.yaml.in/yaml/v4"
)

func TestIndependentFrontendPackageManagers(t *testing.T) {
	root := repositoryRootForTest(t)
	for _, dir := range []string{"apps/cn/admin/react", "apps/cn/nav-web"} {
		t.Run(dir, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, dir, "package.json"))
			if err != nil {
				t.Fatal(err)
			}
			var manifest struct {
				PackageManager string            `json:"packageManager"`
				Scripts        map[string]string `json:"scripts"`
			}
			if err := json.Unmarshal(data, &manifest); err != nil {
				t.Fatal(err)
			}
			if manifest.PackageManager != "pnpm@12.6.0" {
				t.Fatalf("unexpected package manager: %s", manifest.PackageManager)
			}
			for name, script := range manifest.Scripts {
				if regexp.MustCompile(`\b(npm|npx)\b`).MatchString(script) {
					t.Fatalf("%s still invokes npm/npx", name)
				}
			}
			if _, err := os.Stat(filepath.Join(root, dir, "pnpm-lock.yaml")); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(root, dir, "package-lock.json")); !os.IsNotExist(err) {
				t.Fatal("active frontend retains an npm lockfile")
			}
			data, err = os.ReadFile(filepath.Join(root, dir, "pnpm-workspace.yaml"))
			if err != nil {
				t.Fatal(err)
			}
			var settings map[string]any
			if err := yaml.Unmarshal(data, &settings); err != nil {
				t.Fatal(err)
			}
			if settings["packages"] != nil || settings["dangerouslyAllowAllBuilds"] != nil || settings["strictDepBuilds"] == false || settings["verifyDepsBeforeRun"] != "error" {
				t.Fatalf("frontend boundary/script safety relaxed: %+v", settings)
			}
			builds, ok := settings["allowBuilds"].(map[string]any)
			if !ok {
				t.Fatal("missing dependency script review")
			}
			for name, allow := range builds {
				if allow == false {
					continue
				}
				if dir != "apps/cn/nav-web" || (name != "esbuild" && name != "vue-demi") || allow != true {
					t.Fatalf("unreviewed dependency script permission: %s=%v", name, allow)
				}
			}
		})
	}
	for _, name := range []string{"pnpm-workspace.yaml", "pnpm-lock.yaml", "build.bat"} {
		if _, err := os.Stat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Fatalf("unexpected root tooling file %s", name)
		}
	}
}

// Two Vue runtime versions can compile successfully but make Nitro return SSR 500.
// pnpm 12 lockfiles have a package-manager document before the application graph.
func TestNavLockKeepsOneVueSSRRuntime(t *testing.T) {
	file, err := os.Open(filepath.Join(repositoryRootForTest(t), "apps/cn/nav-web/pnpm-lock.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoder := yaml.NewDecoder(file)
	found := false
	for {
		var document struct {
			Importers map[string]struct {
				Dependencies map[string]struct{ Version string }
			}
			Packages map[string]any
		}
		err := decoder.Decode(&document)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		version := strings.Split(document.Importers["."].Dependencies["vue"].Version, "(")[0]
		if version == "" {
			continue
		}
		found = true
		for _, prefix := range []string{"vue@", "@vue/server-renderer@"} {
			count := 0
			for name := range document.Packages {
				if !strings.HasPrefix(name, prefix) {
					continue
				}
				count++
				if name != prefix+version {
					t.Fatalf("%s conflicts with the root Vue runtime %s", name, version)
				}
			}
			if count != 1 {
				t.Fatalf("expected one %s package, found %d", prefix, count)
			}
		}
	}
	if !found {
		t.Fatal("Nav Web lock is missing the Vue application graph")
	}
}
