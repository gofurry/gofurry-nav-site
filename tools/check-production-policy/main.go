package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var productionModules = []string{
	"apps/cn/admin",
	"apps/cn/game-backend",
	"apps/cn/game-collector",
	"apps/cn/nav-backend",
	"apps/cn/nav-collector",
}

var forbidden = []struct {
	name    string
	pattern *regexp.Regexp
}{
	{"GORM", regexp.MustCompile(`gorm\.io/`)},
	{"application schema migration", regexp.MustCompile(`\bAutoMigrate\b`)},
	{"Logrus", regexp.MustCompile(`github\.com/sirupsen/logrus`)},
	{"pkg/errors", regexp.MustCompile(`github\.com/pkg/errors`)},
	{"YAML v2", regexp.MustCompile(`gopkg\.in/yaml\.v2`)},
	{"Swagger/swag", regexp.MustCompile(`github\.com/swaggo/`)},
	{"kardianos/service", regexp.MustCompile(`github\.com/kardianos/service`)},
	{"Viper AutomaticEnv", regexp.MustCompile(`\.AutomaticEnv\s*\(`)},
	{"legacy dependency", regexp.MustCompile(`\blegacy/`)},
	{"experimental dependency", regexp.MustCompile(`\bexperimental/`)},
}

var productionTooling = []string{
	"build.bat",
	"sqlc.yaml",
	".github/workflows/checks.yml",
	".github/workflows/security.yml",
}

var archiveToolingPath = regexp.MustCompile(`(?i)(^|[/\\])(legacy|experimental|third-party|apps[/\\]intl)([/\\]|$)`)

var adminDatabaseDrivers = []struct {
	name    string
	pattern *regexp.Regexp
}{
	{"MySQL", regexp.MustCompile(`(?i)(gorm\.io/driver/mysql|github\.com/go-sql-driver/mysql)`)},
	{"SQLite", regexp.MustCompile(`(?i)(gorm\.io/driver/sqlite|github\.com/mattn/go-sqlite3|modernc\.org/sqlite)`)},
}

var schemaDDL = regexp.MustCompile(`(?i)\b(CREATE\s+TABLE|ALTER\s+TABLE|DROP\s+TABLE|CREATE\s+(UNIQUE\s+)?INDEX)\b`)

type finding struct{ file, rule, line string }

func main() {
	repositoryRoot, err := findRepositoryRoot()
	if err != nil {
		fatal(err)
	}
	var findings []finding
	for _, module := range productionModules {
		root := filepath.Join(repositoryRoot, module)
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if entry.Name() == "vendor" || entry.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			name := entry.Name()
			if name != "go.mod" && !strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, ".sql") {
				return nil
			}
			lines, err := readLines(path)
			if err != nil {
				return err
			}
			relative, _ := filepath.Rel(repositoryRoot, path)
			for number, line := range lines {
				for _, rule := range forbidden {
					if rule.pattern.MatchString(line) {
						findings = append(findings, finding{relative, rule.name, fmt.Sprintf("%d", number+1)})
					}
				}
				if module == "apps/cn/admin" && name == "go.mod" {
					for _, rule := range adminDatabaseDrivers {
						if rule.pattern.MatchString(line) {
							findings = append(findings, finding{relative, "Admin " + rule.name + " database dependency", fmt.Sprintf("%d", number+1)})
						}
					}
				}
				if !strings.HasSuffix(name, "_test.go") && schemaDDL.MatchString(line) {
					findings = append(findings, finding{relative, "service-side schema DDL", fmt.Sprintf("%d", number+1)})
				}
			}
			return nil
		})
		if err != nil {
			fatal(err)
		}
	}
	toolingFindings, err := checkProductionTooling(repositoryRoot)
	if err != nil {
		fatal(err)
	}
	findings = append(findings, toolingFindings...)
	versionFindings, err := checkGoVersions(repositoryRoot)
	if err != nil {
		fatal(err)
	}
	findings = append(findings, versionFindings...)
	if _, err := os.Stat(filepath.Join(repositoryRoot, "go.work")); err == nil {
		findings = append(findings, finding{"go.work", "root Go workspace", "1"})
	} else if !os.IsNotExist(err) {
		fatal(err)
	}
	if len(findings) > 0 {
		sort.Slice(findings, func(i, j int) bool {
			return findings[i].file+findings[i].line+findings[i].rule < findings[j].file+findings[j].line+findings[j].rule
		})
		for _, item := range findings {
			fmt.Fprintf(os.Stderr, "%s:%s: forbidden %s\n", filepath.ToSlash(item.file), item.line, item.rule)
		}
		os.Exit(1)
	}
	fmt.Println("production repository policy passed for five Go modules")
}

func checkProductionTooling(repositoryRoot string) ([]finding, error) {
	var findings []finding
	for _, relative := range productionTooling {
		lines, err := readLines(filepath.Join(repositoryRoot, filepath.FromSlash(relative)))
		if err != nil {
			return nil, err
		}
		for number, line := range lines {
			if archiveToolingPath.MatchString(line) {
				findings = append(findings, finding{relative, "archive/placeholder path in production tooling", fmt.Sprintf("%d", number+1)})
			}
		}
	}
	return findings, nil
}

func checkGoVersions(repositoryRoot string) ([]finding, error) {
	modules := append(append([]string{}, productionModules...), "tools")
	var findings []finding
	for _, module := range modules {
		relative := filepath.ToSlash(filepath.Join(module, "go.mod"))
		lines, err := readLines(filepath.Join(repositoryRoot, filepath.FromSlash(relative)))
		if err != nil {
			return nil, err
		}
		found := false
		for number, line := range lines {
			if strings.HasPrefix(line, "go ") {
				found = true
				if line != "go 1.26.7" {
					findings = append(findings, finding{relative, "Go version other than 1.26.7", fmt.Sprintf("%d", number+1)})
				}
			}
		}
		if !found {
			findings = append(findings, finding{relative, "missing Go version", "1"})
		}
	}
	return findings, nil
}

func findRepositoryRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(current, "sqlc.yaml")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("repository root containing sqlc.yaml not found")
		}
		current = parent
	}
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
