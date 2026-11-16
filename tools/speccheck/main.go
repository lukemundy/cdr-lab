// Command speccheck checks the structure of the specifications that follow ADR 0024, and the Verifies comments of
// the tests. Run it from the repository root.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	os.Exit(run(os.Stdout, os.Stderr))
}

func run(stdout, stderr io.Writer) int {
	// The file list includes new files that are not added yet, and excludes the files that git ignores.
	out, err := exec.CommandContext(context.Background(), "git", "ls-files", "-z", "--cached", "--others",
		"--exclude-standard").Output()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "speccheck: cannot list the files: %v\n", err)

		return 2
	}

	paths := filterPaths(existing(strings.Split(strings.TrimSuffix(string(out), "\x00"), "\x00")))
	res := check(os.DirFS("."), paths)

	for _, f := range res.warnings {
		_, _ = fmt.Fprintf(stdout, "%s:%d: warning: %s\n", f.path, f.line, f.msg)
	}

	for _, f := range res.errors {
		_, _ = fmt.Fprintln(stderr, f)
	}

	if len(res.errors) > 0 {
		_, _ = fmt.Fprintf(stderr, "speccheck: %d errors, %d warnings\n", len(res.errors), len(res.warnings))

		return 1
	}

	return 0
}

// existing removes the files that are deleted in the working tree but not in the index.
func existing(paths []string) []string {
	kept := paths[:0]

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			kept = append(kept, p)
		}
	}

	return kept
}
