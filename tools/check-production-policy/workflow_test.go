package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"go.yaml.in/yaml/v4"
)

func repositoryRootForTest(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestPostgresIntegrationUsesSchemaReadyContractSpecificConfigs(t *testing.T) {
	path := filepath.Join("..", "..", ".github", "workflows", "checks.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(data)
	for _, expected := range []string{
		"GOFURRY_GAME_COLLECTOR_INTEGRATION_CONFIG: /tmp/gofurry-game-collector-integration.yaml",
		"GOFURRY_NAV_COLLECTOR_INTEGRATION_CONFIG: /tmp/gofurry-nav-collector-integration.yaml",
		"GOFURRY_GAME_BACKEND_INTEGRATION_CONFIG: /tmp/gofurry-backend-integration.yaml",
		"GOFURRY_NAV_BACKEND_INTEGRATION_CONFIG: /tmp/gofurry-backend-integration.yaml",
		"GOFURRY_ADMIN_INTEGRATION_CONFIG: /tmp/gofurry-admin-integration.yaml",
		"db_name: gfg_integration",
		"db_name: gfn_integration",
		"go tool goose -dir ../db/game/migrations",
		"go tool goose -dir ../db/nav/migrations",
		"TestGameExpiredLeaseRecoveryIntegration",
		"TestNavExpiredLeaseRecoveryIntegration",
	} {
		if !strings.Contains(workflow, expected) {
			t.Fatalf("checks workflow is missing isolated integration contract %q", expected)
		}
	}
	if strings.Contains(workflow, "/tmp/gofurry-integration.yaml") {
		t.Fatal("checks workflow still shares one incompatible integration config across applications")
	}
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
	for _, name := range []string{"detect-changes", "production-go", "nav-web", "nav-web-visual", "repository-policy", "active-vulnerability", "foundation", "postgres-integration"} {
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

func TestTaskfileChangesReachQualityGates(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repositoryRootForTest(t), ".github/workflows/checks.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var workflow map[string]any
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		t.Fatal(err)
	}
	triggers := workflow["on"].(map[string]any)
	for _, event := range []string{"push", "pull_request"} {
		paths := triggers[event].(map[string]any)["paths"].([]any)
		found := false
		for _, path := range paths {
			found = found || path == "Taskfile.yml"
		}
		if !found {
			t.Fatalf("%s ignores Taskfile changes", event)
		}
	}
	for _, expression := range []string{`shared='([^']+)'`, `set_bool policy '([^']+)'`, `set_bool nav_web '([^']+)'`} {
		match := regexp.MustCompile(expression).FindSubmatch(data)
		if len(match) != 2 || !regexp.MustCompile(string(match[1])).MatchString("Taskfile.yml") {
			t.Fatalf("Taskfile changes do not reach matrix rule %s", expression)
		}
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
