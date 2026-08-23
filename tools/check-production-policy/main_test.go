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
