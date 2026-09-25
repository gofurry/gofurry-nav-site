package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskfileReleaseContract(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repositoryRootForTest(t), "Taskfile.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if problems := validateTaskfile(data); len(problems) != 0 {
		t.Fatal(problems)
	}
	for _, mutation := range []struct{ name, from, to string }{
		{"syntax", "tasks:", "tasks: ["},
		{"dotenv", "tasks:", "dotenv: [.env]\ntasks:"},
		{"target removed", "- task: build:uptime", "- task: build:nav-web"},
		{"wrong architecture", "GOARCH=amd64", "GOARCH=arm64"},
		{"flags lost", "-trimpath", ""},
		{"artifact renamed", "ARTIFACT: gf-nav}", "ARTIFACT: other}"},
		{"unfrozen install", "pnpm install --frozen-lockfile", "pnpm install"},
		{"frontend build in root", "dir: 'apps/cn/{{.FRONTEND}}'", "dir: '.'"},
		{"frontend verify bypasses package cwd", "- task: _frontend-build", "- cmd: pnpm --dir apps/cn/admin/react run build"},
		{"wrong frontend build target", "vars: {FRONTEND: nav-web}", "vars: {FRONTEND: admin/react}"},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			changed := strings.Replace(string(data), mutation.from, mutation.to, 1)
			if changed == string(data) {
				t.Fatal("mutation no longer applies")
			}
			if len(validateTaskfile([]byte(changed))) == 0 {
				t.Fatal("broken contract passed")
			}
		})
	}
}

func TestTaskfileArchivePathIsRejected(t *testing.T) {
	root := t.TempDir()
	for _, name := range productionTooling {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		data := []byte("# fixture\n")
		if name == "Taskfile.yml" {
			data = []byte("cmd: go -C ./legacy/service build .\n")
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	findings, err := checkProductionTooling(root)
	if err != nil || len(findings) != 1 || findings[0].file != "Taskfile.yml" {
		t.Fatalf("archive dependency was not rejected: %+v, %v", findings, err)
	}
}
