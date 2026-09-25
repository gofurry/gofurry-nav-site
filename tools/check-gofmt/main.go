// Command check-gofmt checks Go sources without rewriting CRLF worktrees.
package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: check-gofmt <module directory>...")
		os.Exit(1)
	}
	failed := false
	for _, root := range os.Args[1:] {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "node_modules" || entry.Name() == "vendor" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if err := checkFormat(data); err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", filepath.ToSlash(path), err)
				failed = true
			}
			return nil
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			failed = true
		}
	}
	if failed {
		os.Exit(1)
	}
	fmt.Println("Go formatting passed (LF and CRLF checkouts)")
}

func checkFormat(source []byte) error {
	normalized := bytes.ReplaceAll(source, []byte("\r\n"), []byte("\n"))
	formatted, err := format.Source(normalized)
	if err != nil {
		return err
	}
	if !bytes.Equal(normalized, formatted) {
		return fmt.Errorf("not gofmt formatted; run task fmt")
	}
	return nil
}
