package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"go.yaml.in/yaml/v4"
)

type taskDefinition struct {
	Dir      string            `yaml:"dir"`
	Internal bool              `yaml:"internal"`
	Env      map[string]string `yaml:"env"`
	Cmds     []any             `yaml:"cmds"`
}

type taskDocument struct {
	Version string                    `yaml:"version"`
	Dotenv  []string                  `yaml:"dotenv"`
	Vars    map[string][]string       `yaml:"vars"`
	Tasks   map[string]taskDefinition `yaml:"tasks"`
}

func checkTaskfile(root string) ([]finding, error) {
	data, err := os.ReadFile(filepath.Join(root, "Taskfile.yml"))
	if err != nil {
		return nil, err
	}
	var findings []finding
	for _, problem := range validateTaskfile(data) {
		findings = append(findings, finding{"Taskfile.yml", problem, "1"})
	}
	return findings, nil
}

// Check semantic release boundaries, not a snapshot of the Taskfile's formatting.
// CI also parses it with the minimum supported Task executable.
func validateTaskfile(data []byte) []string {
	var document taskDocument
	if err := yaml.Unmarshal(data, &document); err != nil {
		return []string{fmt.Sprintf("invalid Taskfile: %v", err)}
	}
	var problems []string
	if document.Version != "3.45.3" {
		problems = append(problems, "unsupported Task schema/minimum version")
	}
	if len(document.Dotenv) != 0 {
		problems = append(problems, "automatic dotenv loading")
	}
	modules := slices.Clone(document.Vars["GO_MODULES"])
	wantModules := append(slices.Clone(productionModules), "tools")
	slices.Sort(modules)
	slices.Sort(wantModules)
	if !slices.Equal(modules, wantModules) {
		problems = append(problems, "incorrect active Go module set")
	}
	expected := map[string]string{
		"nav-backend": "gf-nav", "nav-collector": "gf-nav-collector",
		"game-backend": "gf-game", "game-collector": "gf-game-collector",
		"admin": "gofurry-admin", "uptime": "gf-uptime",
	}
	var aggregate []string
	for _, command := range document.Tasks["build"].Cmds {
		call, _ := command.(map[string]any)
		name, _ := call["task"].(string)
		aggregate = append(aggregate, strings.TrimPrefix(name, "build:"))
	}
	if len(aggregate) != len(expected) {
		problems = append(problems, "release aggregate must contain six Go targets")
	}
	for module, artifact := range expected {
		if !slices.Contains(aggregate, module) {
			problems = append(problems, "missing release target "+module)
		}
		found := false
		for _, command := range document.Tasks["build:"+module].Cmds {
			call, _ := command.(map[string]any)
			vars, _ := call["vars"].(map[string]any)
			if call["task"] == "_release" && vars["MODULE"] == module && vars["ARTIFACT"] == artifact {
				found = true
			}
		}
		if !found {
			problems = append(problems, "incorrect release artifact mapping for "+module)
		}
	}
	release := document.Tasks["_release"]
	if !release.Internal {
		problems = append(problems, "incorrect release helper visibility")
	}
	var buildCommand string
	for _, command := range release.Cmds {
		if value, ok := command.(string); ok && strings.Contains(value, " build ") {
			buildCommand = value
		}
	}
	for _, argument := range []string{"GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go", "-trimpath", `-ldflags="-s -w"`, "/build/{{.ARTIFACT}}/{{.ARTIFACT}}"} {
		if !strings.Contains(buildCommand, argument) {
			problems = append(problems, "release build missing "+argument)
		}
	}
	admin := document.Tasks["build:admin"].Cmds
	if len(admin) < 6 {
		problems = append(problems, "incomplete Admin frontend/embed/companion build")
	} else {
		first, _ := admin[0].(map[string]any)
		last, _ := admin[len(admin)-1].(string)
		if first["task"] != "deps:admin" || !callsFrontendBuild(admin[1], "admin/react") || !strings.Contains(last, "webui/dist/. build/gofurry-admin/dist/") {
			problems = append(problems, "Admin must install, build and preserve the dist companion")
		}
	}
	// Corepack resolves packageManager from cwd before pnpm processes --dir.
	frontendBuild := document.Tasks["_frontend-build"]
	if !frontendBuild.Internal || frontendBuild.Dir != "apps/cn/{{.FRONTEND}}" || len(frontendBuild.Cmds) != 1 || frontendBuild.Cmds[0] != "pnpm run build" {
		problems = append(problems, "frontend builds must run in the owning package directory")
	}
	verifyBuilds := document.Tasks["verify:frontend-build"].Cmds
	if len(verifyBuilds) != 2 || !callsFrontendBuild(verifyBuilds[0], "admin/react") || !callsFrontendBuild(verifyBuilds[1], "nav-web") {
		problems = append(problems, "frontend verification must build both owning packages")
	}
	navBuild := document.Tasks["build:nav-web"].Cmds
	if len(navBuild) != 2 || !callsFrontendBuild(navBuild[1], "nav-web") {
		problems = append(problems, "Nav Web must build in its owning package directory")
	}
	for _, frontend := range []string{"admin", "nav-web"} {
		commands := document.Tasks["deps:"+frontend].Cmds
		if len(commands) != 1 || commands[0] != "pnpm install --frozen-lockfile" {
			problems = append(problems, "non-frozen frontend install for "+frontend)
		}
	}
	return problems
}

func callsFrontendBuild(command any, frontend string) bool {
	call, _ := command.(map[string]any)
	vars, _ := call["vars"].(map[string]any)
	return call["task"] == "_frontend-build" && vars["FRONTEND"] == frontend
}
