package main

import (
	"os"
	"path/filepath"
	"strings"
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
