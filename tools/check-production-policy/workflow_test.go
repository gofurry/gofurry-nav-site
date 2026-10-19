package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	for _, name := range []string{"detect-changes", "production-go", "nav-web", "nav-web-fast", "repository-policy", "active-vulnerability", "foundation", "postgres-integration"} {
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
		for _, required := range []string{"Taskfile.yml", ".go-version", ".node-version"} {
			found := false
			for _, path := range paths {
				found = found || path == required
			}
			if !found {
				t.Fatalf("%s ignores %s changes", event, required)
			}
		}
	}
	for _, command := range []string{"node .github/scripts/detect-changes.mjs", "node --test .github/scripts/detect-changes.test.mjs"} {
		if !strings.Contains(string(data), command) {
			t.Fatalf("change selection and its Taskfile regression tests must run: %s", command)
		}
	}
}

func readWorkflowForTest(t *testing.T, name string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repositoryRootForTest(t), ".github/workflows", name))
	if err != nil {
		t.Fatal(err)
	}
	var workflow map[string]any
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return workflow
}

func TestWorkflowToolchainSources(t *testing.T) {
	for _, name := range []string{"checks.yml", "security.yml", "nav-web-full.yml", "nav-web-visual.yml"} {
		workflow := readWorkflowForTest(t, name)
		for _, value := range workflow["jobs"].(map[string]any) {
			job := value.(map[string]any)
			for _, value := range job["steps"].([]any) {
				step := value.(map[string]any)
				uses, _ := step["uses"].(string)
				if strings.HasPrefix(uses, "actions/setup-node@") {
					with := step["with"].(map[string]any)
					if with["node-version-file"] != ".node-version" || with["node-version"] != nil {
						t.Fatalf("%s must discover Node from the root declaration", name)
					}
				}
				if strings.HasPrefix(uses, "actions/setup-go@") {
					with := step["with"].(map[string]any)
					file, _ := with["go-version-file"].(string)
					if !strings.HasSuffix(file, "/go.mod") || with["go-version"] != nil {
						t.Fatalf("%s must retain module-owned Go versions", name)
					}
				}
			}
		}
	}
}

func TestSecurityScansEveryModuleBeforeFailing(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is required to exercise the actual Actions scan loop")
	}
	for _, owner := range []struct{ file, job string }{{"security.yml", "govulncheck"}, {"checks.yml", "active-vulnerability"}} {
		workflow := readWorkflowForTest(t, owner.file)
		job := workflow["jobs"].(map[string]any)[owner.job].(map[string]any)
		script := ""
		for _, value := range job["steps"].([]any) {
			step := value.(map[string]any)
			if run, ok := step["run"].(string); ok && strings.Contains(run, "go tool govulncheck") {
				script = run
			}
		}
		if script == "" {
			t.Fatal("missing scan script", owner.file)
		}
		for _, fail := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/fail=%t", owner.file, fail), func(t *testing.T) {
				stub := "go() { printf 'VISIT:%s\\n' \"$4\"; "
				if fail {
					stub += "case \"$4\" in */admin) return 3;; */game-collector) return 2;; */nav-collector) return 1;; esac; "
				}
				stub += "return 0; }\n"
				command := exec.Command(bash, "-s")
				command.Stdin = strings.NewReader(stub + script)
				output, err := command.CombinedOutput()
				if (err != nil) != fail {
					t.Fatalf("unexpected scan exit %v: %s", err, output)
				}
				if fail && command.ProcessState.ExitCode() != 1 {
					t.Fatalf("aggregate must exit 1: %s", output)
				}
				text := string(output)
				if strings.Count(text, "VISIT:") != len(productionModules) {
					t.Fatalf("incomplete scan: %s", text)
				}
				for _, module := range productionModules {
					if strings.Count(text, "VISIT:../"+module+"\n") != 1 {
						t.Fatalf("module not scanned exactly once: %s", text)
					}
					name := filepath.Base(module)
					result := "PASS"
					if fail && (name == "admin" || name == "game-collector" || name == "nav-collector") {
						result = "FAIL"
					}
					if !strings.Contains(text, name+": "+result) {
						t.Fatalf("missing summary: %s", text)
					}
				}
			})
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
		t.Fatalf("active modules do not use Go %s: %+v", requiredGoVersion, findings)
	}
}

func TestToolchainDeclarationDrift(t *testing.T) {
	if findings, err := checkToolchainDeclarations(repositoryRootForTest(t)); err != nil || len(findings) != 0 {
		t.Fatalf("root toolchain declarations: %v %+v", err, findings)
	}
	for _, file := range []string{".go-version", ".node-version"} {
		for _, value := range []string{"", "1.26.7", "24.1", "invalid"} {
			t.Run(file+"/"+value, func(t *testing.T) {
				root := t.TempDir()
				for path, content := range map[string]string{".go-version": requiredGoVersion, ".node-version": "24"} {
					if path == file {
						content = value
					}
					if err := os.WriteFile(filepath.Join(root, path), []byte(content+"\n"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				findings, err := checkToolchainDeclarations(root)
				if err != nil || len(findings) != 1 || findings[0].file != file {
					t.Fatalf("declaration drift not isolated: %v %+v", err, findings)
				}
				if err := os.Remove(filepath.Join(root, file)); err != nil {
					t.Fatal(err)
				}
				if _, err := checkToolchainDeclarations(root); !os.IsNotExist(err) {
					t.Fatalf("missing declaration accepted: %v", err)
				}
			})
		}
	}
}
