package main

import "testing"

func TestProductionModulesIncludeStandaloneUptime(t *testing.T) {
	if len(productionModules) != 6 {
		t.Fatalf("production module count = %d, want 6", len(productionModules))
	}
	for _, module := range productionModules {
		if module == "apps/cn/uptime" {
			return
		}
	}
	t.Fatal("apps/cn/uptime is missing from the production policy")
}

func TestArchiveToolingPathBoundaries(t *testing.T) {
	for _, line := range []string{
		"dir: legacy/service", "cmd: go -C ./experimental/demo build .",
		"dir: 'third-party/lib'", `dir: apps\intl\web`, "dir: legacy",
	} {
		if !archiveToolingPath.MatchString(line) {
			t.Errorf("archive path escaped policy: %s", line)
		}
	}
	for _, line := range []string{"dir: apps/cn/nav-web", "cmd: check-legacy-compatibility", "dir: tools"} {
		if archiveToolingPath.MatchString(line) {
			t.Errorf("non-archive path rejected: %s", line)
		}
	}
}
