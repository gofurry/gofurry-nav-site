package main

import (
	"os"
	"path/filepath"
	"slices"
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
	for _, name := range []string{"detect-changes", "production-go", "nav-web", "nav-web-build", "nav-web-browser", "nav-web-visual", "nav-web-image", "repository-policy", "active-vulnerability", "foundation", "postgres-integration"} {
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
	for _, command := range []string{"node .github/scripts/detect-changes.mjs", "node --test .github/scripts/detect-changes.test.mjs"} {
		if !strings.Contains(string(data), command) {
			t.Fatalf("change selection and its Taskfile regression tests must run: %s", command)
		}
	}
}

func TestFrontendCIPreservesAllGatesAndPinnedBuildProvenance(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repositoryRootForTest(t), ".github/workflows/checks.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var workflow struct {
		Jobs map[string]struct {
			Needs     any
			If        string
			Container struct{ Image string }
			Strategy  struct {
				Matrix struct{ Shard []int }
			}
			Steps []struct {
				Run  string
				Uses string
				With map[string]any
			}
		}
	}
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		t.Fatal(err)
	}
	jobs := workflow.Jobs
	image := jobs["nav-web-build"].Container.Image
	if !strings.Contains(image, "playwright:v1.60.0-noble@sha256:") {
		t.Fatal("the shared frontend build must use the pinned Visual environment")
	}
	for _, name := range []string{"nav-web-browser", "nav-web-visual", "nav-web-image"} {
		if jobs[name].Needs != "nav-web-build" {
			t.Fatalf("%s must start after the common build, not another test job", name)
		}
	}
	for _, name := range []string{"nav-web-browser", "nav-web-visual"} {
		job := jobs[name]
		if job.Container.Image != image {
			t.Fatalf("%s differs from the build environment", name)
		}
		download := false
		for _, step := range job.Steps {
			if strings.HasPrefix(step.Uses, "actions/download-artifact@") && step.With["name"] == "nav-web-output-${{ github.sha }}" {
				download = true
			}
			if strings.Contains(step.Run, "pnpm run build") || strings.Contains(step.Run, "--update-snapshots") {
				t.Fatalf("%s must compare the shared build without accepting baselines", name)
			}
		}
		if !download {
			t.Fatalf("%s must consume this commit's build", name)
		}
	}
	if !slices.Equal(jobs["nav-web-browser"].Strategy.Matrix.Shard, []int{1, 2, 3}) {
		t.Fatal("Browser CI must run every shard")
	}
	gate := jobs["nav-web"]
	needs, ok := gate.Needs.([]any)
	if !ok || !strings.Contains(gate.If, "always()") {
		t.Fatal("the stable nav-web gate must report upstream failures")
	}
	for _, name := range []string{"nav-web-build", "nav-web-browser", "nav-web-visual", "nav-web-image"} {
		if !slices.Contains(needs, any(name)) {
			t.Fatalf("nav-web no longer requires %s", name)
		}
	}
	if strings.Contains(string(data), "task --dry build") {
		t.Fatal("policy inspection must not evaluate embed preconditions on a clean checkout")
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
