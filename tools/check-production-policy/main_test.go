package main

import (
	"slices"
	"testing"
)

func TestProductionModulesMatchActiveServices(t *testing.T) {
	want := []string{
		"apps/cn/admin",
		"apps/cn/game-backend",
		"apps/cn/game-collector",
		"apps/cn/nav-backend",
		"apps/cn/nav-collector",
	}
	if !slices.Equal(productionModules, want) {
		t.Fatalf("production modules = %v, want %v", productionModules, want)
	}
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
	for _, line := range []string{"dir: apps/cn/web", "cmd: check-legacy-compatibility", "dir: tools"} {
		if archiveToolingPath.MatchString(line) {
			t.Errorf("non-archive path rejected: %s", line)
		}
	}
}
