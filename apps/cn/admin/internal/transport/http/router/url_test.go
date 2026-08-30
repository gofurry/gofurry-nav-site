package router

import (
	"bufio"
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestEveryBusinessRouteDeclaresCapability(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source path")
	}
	content, err := os.Open(strings.TrimSuffix(file, "_test.go") + ".go")
	if err != nil {
		t.Fatal(err)
	}
	defer content.Close()

	protectedFunctions := map[string]bool{
		"changeRoutes": true, "metricRoutes": true, "collectionRoutes": true,
		"accountRoutes": true, "optionsRoutes": true, "navRoutes": true, "gameRoutes": true,
	}
	current := ""
	scanner := bufio.NewScanner(content)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "func ") {
			current = strings.Fields(strings.TrimPrefix(line, "func "))[0]
			current = strings.SplitN(current, "(", 2)[0]
		}
		if !protectedFunctions[current] || !strings.HasPrefix(line, "root.") {
			continue
		}
		if !strings.Contains(line, "authmw.Require(authorization.") {
			t.Fatalf("protected route has no explicit capability: %s", line)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestRepresentativeRouteCapabilityMatrix(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	data, err := os.ReadFile(strings.TrimSuffix(file, "_test.go") + ".go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	expected := []string{
		`root.Post("/sites", authmw.Require(authorization.ContentWrite)`,
		`root.Post("/schedules/:domain/:id/run", authmw.Require(authorization.CollectionExecute)`,
		`root.Put("/schedules/:domain/:id", authmw.Require(authorization.CollectionControl)`,
		`root.Get("/overview", authmw.Require(authorization.MetricsRead)`,
		`root.Get("/registry", authmw.Require(authorization.MetricsTechnical)`,
		`root.Get("/overview", authmw.Require(authorization.ChangesRead)`,
		`root.Get("/registry", authmw.Require(authorization.ChangesTechnical)`,
		`root.Post("/", authmw.Require(authorization.AccountManage)`,
	}
	for _, fragment := range expected {
		if !strings.Contains(source, fragment) {
			t.Errorf("route capability fragment is missing: %s", fragment)
		}
	}
}
