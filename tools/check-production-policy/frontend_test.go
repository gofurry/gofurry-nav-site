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
	for _, dir := range []string{"apps/cn/admin/react", "apps/cn/web"} {
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
				if dir != "apps/cn/web" || (name != "esbuild" && name != "vue-demi") || allow != true {
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

func TestFrontendWorkspaceKeepsDeploymentIdentity(t *testing.T) {
	root := repositoryRootForTest(t)
	read := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(root, "apps/cn/web", name))
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	var manifest struct{ Name string }
	if err := json.Unmarshal(read("package.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Name != "gofurry-nav-web" {
		t.Fatal("workspace rename changed package identity")
	}
	var compose struct {
		Name     string
		Services map[string]struct {
			ContainerName string `yaml:"container_name"`
			Image         string
			Ports         []string
			Environment   map[string]string
			Build         struct {
				Context, Dockerfile string
				Args                map[string]string
			}
		}
	}
	if err := yaml.Unmarshal(read("docker-compose.prod.yml"), &compose); err != nil {
		t.Fatal(err)
	}
	service, ok := compose.Services["gofurry-nav-web"]
	if compose.Name != "nav-web" || len(compose.Services) != 1 || !ok || service.ContainerName != "gofurry-nav-web" || service.Image != "gofurry-nav-web:latest" || len(service.Ports) != 1 || service.Ports[0] != "127.0.0.1:3000:3000" {
		t.Fatal("workspace rename changed production Compose identity/port")
	}
	if service.Build.Context != "." || service.Build.Dockerfile != "Dockerfile" || service.Build.Args["NODE_BUILD_HEAP_MB"] != "${NODE_BUILD_HEAP_MB:-4096}" {
		t.Fatal("missing build-only heap argument")
	}
	for _, key := range []string{"NODE_OPTIONS", "NODE_BUILD_HEAP_MB"} {
		if _, exists := service.Environment[key]; exists {
			t.Fatalf("build setting leaked into runtime: %s", key)
		}
	}
	dockerfile := string(read("Dockerfile"))
	before, buildAndRunner, ok := strings.Cut(dockerfile, "FROM deps AS build")
	if !ok || strings.Contains(before, "NODE_OPTIONS") || strings.Contains(before, "NODE_BUILD_HEAP_MB") {
		t.Fatal("build heap must not enter shared dependency stages")
	}
	build, runner, ok := strings.Cut(buildAndRunner, "FROM deps AS runner")
	if !ok || !strings.Contains(build, "ARG NODE_BUILD_HEAP_MB=4096") || !strings.Contains(build, "ENV NODE_OPTIONS=--max-old-space-size=${NODE_BUILD_HEAP_MB}") || strings.Contains(runner, "NODE_OPTIONS") || strings.Contains(runner, "NODE_BUILD_HEAP_MB") {
		t.Fatal("heap override must remain in the build stage; runner must inherit deps")
	}
}

// Two Vue runtime versions can compile successfully but make Nitro return SSR 500.
// pnpm 12 lockfiles have a package-manager document before the application graph.
func TestNavLockKeepsOneVueSSRRuntime(t *testing.T) {
	file, err := os.Open(filepath.Join(repositoryRootForTest(t), "apps/cn/web/pnpm-lock.yaml"))
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
