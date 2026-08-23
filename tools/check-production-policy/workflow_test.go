package main

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func repositoryRootForTest(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestEngineeringFoundationWorkflowParses(t *testing.T) {
	path := filepath.Join("..", "..", ".github", "workflows", "checks.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var workflow map[string]any
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		t.Fatalf("parse checks workflow: %v", err)
	}
	jobs, ok := workflow["jobs"].(map[string]any)
	if !ok {
		t.Fatal("checks workflow has no jobs mapping")
	}
	for _, name := range []string{"detect-changes", "production-go", "nav-web", "repository-policy", "active-vulnerability", "foundation", "postgres-integration"} {
		if _, ok := jobs[name]; !ok {
			t.Fatalf("checks workflow is missing %s", name)
		}
	}
}

func TestSecurityWorkflowParses(t *testing.T) {
	path := filepath.Join("..", "..", ".github", "workflows", "security.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var workflow map[string]any
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		t.Fatalf("parse security workflow: %v", err)
	}
	jobs, ok := workflow["jobs"].(map[string]any)
	if !ok {
		t.Fatal("security workflow has no jobs mapping")
	}
	if _, ok := jobs["govulncheck"]; !ok {
		t.Fatal("security workflow is missing govulncheck")
	}
}

func TestProductionToolingExcludesArchiveAndPlaceholderTrees(t *testing.T) {
	findings, err := checkProductionTooling(repositoryRootForTest(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("production tooling contains archive paths: %+v", findings)
	}
}

func TestActiveModulesUseRequiredGoVersion(t *testing.T) {
	findings, err := checkGoVersions(repositoryRootForTest(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("active modules do not use Go 1.26.7: %+v", findings)
	}
}
