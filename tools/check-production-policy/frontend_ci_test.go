package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"go.yaml.in/yaml/v4"
)

const playwrightImage = "mcr.microsoft.com/playwright:v1.60.0-noble@sha256:9bd26ad900bb5e0f4dee75839e957a89ae89c2b7ab1e76050e559790e946b948"

type ciStep struct {
	Name string
	ID   string `yaml:"id"`
	Uses string
	Run  string
	If   string
	With map[string]any
}

type ciJob struct {
	Needs     any
	If        string
	RunsOn    string `yaml:"runs-on"`
	Timeout   int    `yaml:"timeout-minutes"`
	Env       map[string]string
	Outputs   map[string]string
	Container struct {
		Image   string
		Options string
	}
	Strategy struct {
		FailFast bool `yaml:"fail-fast"`
		Matrix   struct{ Group []string }
	}
	Steps []ciStep
}

type ciWorkflow struct {
	On   map[string]any
	Jobs map[string]ciJob
}

func readCIWorkflow(t *testing.T, name string) ciWorkflow {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repositoryRootForTest(t), ".github/workflows", name))
	if err != nil {
		t.Fatal(err)
	}
	var workflow ciWorkflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return workflow
}

func ciCommands(job ciJob) string {
	var commands []string
	for _, step := range job.Steps {
		commands = append(commands, step.Run)
	}
	return strings.Join(commands, "\n")
}

func requireCICommands(t *testing.T, job ciJob, commands ...string) {
	t.Helper()
	all := ciCommands(job)
	for _, command := range commands {
		if !strings.Contains(all, command) {
			t.Errorf("missing command %q", command)
		}
	}
}

func requireCIAction(t *testing.T, job ciJob, prefix, key string, value any) {
	t.Helper()
	for _, step := range job.Steps {
		if strings.HasPrefix(step.Uses, prefix) && step.With[key] == value {
			return
		}
	}
	t.Fatalf("missing %s with %s=%v", prefix, key, value)
}

func requireCINeeds(t *testing.T, job ciJob, names ...string) {
	t.Helper()
	actual, ok := job.Needs.([]any)
	if !ok || len(actual) != len(names) {
		t.Fatalf("unexpected gate dependencies: %v", job.Needs)
	}
	for _, name := range names {
		if !slices.Contains(actual, any(name)) {
			t.Errorf("missing dependency %s", name)
		}
	}
}

func TestFrontendFastCIHasOneSetupBuildAndSmokeOwner(t *testing.T) {
	w := readCIWorkflow(t, "checks.yml")
	fast := w.Jobs["nav-web-fast"]
	if fast.Needs != "detect-changes" || fast.Timeout != 10 || fast.Container.Image != playwrightImage {
		t.Fatal("Fast must start after selection in the pinned container with its bounded timeout")
	}
	if !strings.Contains(fast.If, "github.event_name == 'pull_request'") || !strings.Contains(fast.If, "github.ref == 'refs/heads/dev'") {
		t.Fatal("Fast must be selected for dev/PR rather than main Full runs")
	}
	requireCICommands(t, fast, "pnpm install --frozen-lockfile", "pnpm run lint", "pnpm run stylelint", "pnpm run style:policy:test", "pnpm run style:policy", "pnpm run test:unit", "pnpm run test:nuxt", "pnpm run typecheck", "pnpm run insights:semantics", "pnpm run seo:recovery:test", "pnpm run build", "pnpm run test:browser:smoke --workers=1")
	commands := ciCommands(fast)
	if strings.Count(commands, "pnpm install") != 1 || strings.Count(commands, "pnpm run build") != 1 || strings.Index(commands, "pnpm run build") > strings.Index(commands, "pnpm run test:browser:smoke") {
		t.Fatal("Smoke must use the build in the same job with no repeated install/build")
	}
	for _, forbidden := range []string{"--shard", "test:visual", "test:browser:regression", "pnpm run test:browser "} {
		if strings.Contains(commands, forbidden) {
			t.Errorf("expensive gate in Fast: %s", forbidden)
		}
	}
	for _, name := range []string{"nav-web-build", "nav-web-browser", "nav-web-visual", "nav-web-image"} {
		if _, ok := w.Jobs[name]; ok {
			t.Errorf("retired daily job %s", name)
		}
	}
	for _, step := range fast.Steps {
		if strings.HasPrefix(step.Uses, "docker/") || strings.HasPrefix(step.Uses, "actions/download-artifact@") {
			t.Fatal("Fast must not build an image or restore a separate build")
		}
	}
	requireCIAction(t, fast, "actions/upload-artifact@", "name", "nav-web-smoke-failure")
	gate := w.Jobs["nav-web"]
	requireCINeeds(t, gate, "detect-changes", "nav-web-fast")
	if !strings.Contains(gate.If, "always()") {
		t.Fatal("the stable Fast gate must report upstream failures")
	}
	requireCICommands(t, gate, `test "$DETECT" = success`, `test "$FAST" = success`)
}

func TestFrontendFullCIKeepsCompleteGroupsAndExactBuildSource(t *testing.T) {
	w := readCIWorkflow(t, "nav-web-full.yml")
	if len(w.On) != 3 || w.On["push"] == nil || w.On["schedule"] == nil {
		t.Fatal("Full requires main/manual/nightly triggers")
	}
	if _, ok := w.On["workflow_dispatch"]; !ok {
		t.Fatal("Full must be manually dispatchable")
	}
	branches := w.On["push"].(map[string]any)["branches"].([]any)
	if !slices.Equal(branches, []any{"main"}) {
		t.Fatal("Full push must target only main")
	}
	build := w.Jobs["nav-web-build"]
	if build.Container.Image != playwrightImage || build.Timeout != 15 {
		t.Fatal("Full build must use the pinned environment")
	}
	requireCIAction(t, build, "actions/checkout@", "ref", "${{ github.event_name == 'schedule' && 'dev' || github.ref }}")
	requireCICommands(t, build, "git rev-parse HEAD", "node --test .github/scripts/nav-web-regression-groups.test.mjs", "pnpm install --frozen-lockfile", "pnpm run build", "tar -czf")
	if build.Outputs["source_sha"] != "${{ steps.source.outputs.sha }}" {
		t.Fatal("nightly source must resolve once and be exported")
	}
	requireCIAction(t, build, "actions/upload-artifact@", "name", "nav-web-full-output-${{ steps.source.outputs.sha }}")
	requireCIAction(t, build, "actions/upload-artifact@", "retention-days", 1)
	group := w.Jobs["nav-web-full"]
	if group.Needs != "nav-web-build" || group.Container.Image != playwrightImage || group.Timeout != 10 || group.Strategy.FailFast {
		t.Fatal("every Full group must consume the common pinned build and run independently")
	}
	if !slices.Equal(group.Strategy.Matrix.Group, []string{"games", "sites-other", "insights-core", "insights-nav"}) {
		t.Fatal("Full must run all fixed groups")
	}
	requireCIAction(t, group, "actions/checkout@", "ref", "${{ needs.nav-web-build.outputs.source_sha }}")
	requireCIAction(t, group, "actions/download-artifact@", "name", "nav-web-full-output-${{ needs.nav-web-build.outputs.source_sha }}")
	requireCICommands(t, group, "set -euo pipefail", "nav-web-regression-groups.mjs", `mapfile -t specs < "$RUNNER_TEMP/regression-files"`, `pnpm exec playwright test "${specs[@]}" --workers=1`)
	for _, forbidden := range []string{"pnpm run build", "--shard", "--update-snapshots"} {
		if strings.Contains(ciCommands(group), forbidden) {
			t.Errorf("Full group must not use %s", forbidden)
		}
	}
	requireCIAction(t, group, "actions/upload-artifact@", "name", "nav-web-full-failure-${{ matrix.group }}")
	image := w.Jobs["nav-web-image"]
	if image.If != "github.event_name != 'schedule'" || image.Needs != "nav-web-build" {
		t.Fatal("image is main/manual only")
	}
	requireCIAction(t, image, "actions/checkout@", "ref", "${{ needs.nav-web-build.outputs.source_sha }}")
	requireCIAction(t, image, "docker/build-push-action@", "cache-from", "type=gha,scope=nav-web")
	requireCIAction(t, image, "docker/build-push-action@", "cache-to", "type=gha,scope=nav-web,mode=max")
	gate := w.Jobs["nav-web-full-gate"]
	requireCINeeds(t, gate, "nav-web-build", "nav-web-full", "nav-web-image")
	if !strings.Contains(gate.If, "always()") {
		t.Fatal("Full gate must report failures")
	}
	requireCICommands(t, gate, `test "$BUILD" = success`, `test "$REGRESSION" = success`, `if [ "$EVENT" = schedule ]`, `test "$IMAGE" = skipped`, `test "$IMAGE" = success`)
}

func TestFrontendVisualCIIsManualComparisonOnly(t *testing.T) {
	w := readCIWorkflow(t, "nav-web-visual.yml")
	if _, ok := w.On["workflow_dispatch"]; !ok || len(w.On) != 1 {
		t.Fatal("Visual must be manual-only")
	}
	if len(w.Jobs) != 1 {
		t.Fatal("Visual build and compare share one job")
	}
	job := w.Jobs["nav-web-visual"]
	if job.Container.Image != playwrightImage || job.Env["GOFURRY_VISUAL_ENV"] != "pinned" || job.Timeout != 15 {
		t.Fatal("Visual environment changed")
	}
	requireCICommands(t, job, "pnpm install --frozen-lockfile", "pnpm run build", "pnpm run test:visual")
	if strings.Contains(ciCommands(job), "--update-snapshots") || strings.Contains(ciCommands(job), "test:visual:update") {
		t.Fatal("CI cannot accept golden updates")
	}
	requireCIAction(t, job, "actions/upload-artifact@", "name", "nav-web-visual-failure")
}

func TestFrontendCIRunnerAndSmokeBoundaries(t *testing.T) {
	for _, file := range []string{"checks.yml", "security.yml", "nav-web-full.yml", "nav-web-visual.yml"} {
		for name, job := range readCIWorkflow(t, file).Jobs {
			if job.RunsOn != "ubuntu-24.04" {
				t.Errorf("%s/%s has an unpinned hosted runner", file, name)
			}
		}
	}
	root := repositoryRootForTest(t)
	entries, err := os.ReadDir(filepath.Join(root, "apps/cn/web/tests/browser/smoke"))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".spec.ts") {
			names = append(names, entry.Name())
		}
	}
	expected := []string{"game-detail.spec.ts", "games-search.spec.ts", "hero.spec.ts", "nav-home-locales.spec.ts", "resource-routing.spec.ts", "site-detail.spec.ts", "static-locales.spec.ts"}
	if !slices.Equal(names, expected) {
		t.Fatalf("Smoke owners changed: %v", names)
	}
	for _, file := range []string{"playwright.config.ts", "playwright.visual.config.ts"} {
		data, err := os.ReadFile(filepath.Join(root, "apps/cn/web", file))
		if err != nil {
			t.Fatal(err)
		}
		for _, invariant := range []string{"fullyParallel: false", "retries: 0"} {
			if !strings.Contains(string(data), invariant) {
				t.Errorf("%s lost %s", file, invariant)
			}
		}
	}
}
