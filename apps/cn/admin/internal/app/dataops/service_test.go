package dataops

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestMigrationStatus(t *testing.T) {
	tests := []struct {
		name              string
		current, expected int64
		pending           int
		want              string
	}{
		{name: "current", current: 3, expected: 3, want: "current"},
		{name: "missing migration", current: 3, expected: 3, pending: 1, want: "behind"},
		{name: "behind", current: 2, expected: 3, pending: 1, want: "behind"},
		{name: "ahead", current: 4, expected: 3, want: "ahead"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := migrationStatus(test.current, test.expected, test.pending); got != test.want {
				t.Fatalf("status=%q, want %q", got, test.want)
			}
		})
	}
}

func TestExpectedRepositoryMigrationVersions(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source")
	}
	repositoryRoot := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(repositoryRoot, "sqlc.yaml")); err == nil {
			break
		}
		parent := filepath.Dir(repositoryRoot)
		if parent == repositoryRoot {
			t.Fatal("repository root not found")
		}
		repositoryRoot = parent
	}
	owners := map[string]string{"gfa": "admin", "gfn": "nav", "gfg": "game"}
	for key, owner := range owners {
		entries, err := os.ReadDir(filepath.Join(repositoryRoot, "db", owner, "migrations"))
		if err != nil {
			t.Fatal(err)
		}
		actual := make([]int64, 0, len(entries))
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
				continue
			}
			version, parseErr := strconv.ParseInt(strings.SplitN(entry.Name(), "_", 2)[0], 10, 64)
			if parseErr != nil {
				t.Fatalf("parse migration %s: %v", entry.Name(), parseErr)
			}
			actual = append(actual, version)
		}
		sort.Slice(actual, func(i, j int) bool { return actual[i] < actual[j] })
		if !reflect.DeepEqual(expectedRepoMigrations[key], actual) {
			t.Fatalf("%s compiled migrations=%v, repository=%v", key, expectedRepoMigrations[key], actual)
		}
	}
}
