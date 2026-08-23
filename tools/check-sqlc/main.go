// Command check-sqlc regenerates every service-local sqlc package and fails if
// generated files differ from the checked-in contract.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	rootFlag := flag.String("root", "..", "repository root")
	flag.Parse()
	root, err := filepath.Abs(*rootFlag)
	check(err)
	toolsDir := filepath.Join(root, "tools")

	generate := exec.Command("go", "tool", "sqlc", "generate", "-f", filepath.Join(root, "sqlc.yaml"))
	generate.Dir = toolsDir
	generate.Stdout = os.Stdout
	generate.Stderr = os.Stderr
	check(generate.Run())

	generated := []string{
		"apps/cn/game-collector/internal/db/game/sqlc",
		"apps/cn/game-backend/internal/db/game/sqlc",
		"apps/cn/nav-collector/internal/db/nav/sqlc",
		"apps/cn/nav-backend/internal/db/nav/sqlc",
		"apps/cn/admin/internal/db/game/sqlc",
		"apps/cn/admin/internal/db/nav/sqlc",
		"apps/cn/admin/internal/db/admin/sqlc",
	}
	args := append([]string{"status", "--porcelain", "--"}, generated...)
	status := exec.Command("git", args...)
	status.Dir = root
	output, err := status.Output()
	check(err)
	if strings.TrimSpace(string(output)) != "" {
		fmt.Fprintln(os.Stderr, string(output))
		fmt.Fprintln(os.Stderr, "sqlc generated code drift detected; run from tools/: go tool sqlc generate -f ../sqlc.yaml")
		os.Exit(1)
	}
	fmt.Println("sqlc generated code matches the checked-in contract")
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
