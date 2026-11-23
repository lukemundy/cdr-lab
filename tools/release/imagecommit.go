package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// readInputs reads the git pathspecs of the image inputs of each service, by service directory name.
func readInputs(path string) (map[string][]string, error) {
	data, err := os.ReadFile(path) //nolint:gosec // The path is a flag of a repository tool that a maintainer runs.
	if err != nil {
		return nil, fmt.Errorf("read the image inputs: %w", err)
	}

	var inputs map[string][]string
	if err := json.Unmarshal(data, &inputs); err != nil {
		return nil, fmt.Errorf("read the image inputs %s: %w", path, err)
	}

	return inputs, nil
}

type gitRepo struct {
	dir string
}

func (g gitRepo) run(ctx context.Context, args ...string) ([]string, error) {
	//nolint:gosec // The arguments are a revision and pathspecs after "--", so git cannot read a pathspec as a flag.
	out, err := exec.CommandContext(ctx, "git", append([]string{"-C", g.dir}, args...)...).Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w", args[0], err)
	}

	return strings.Fields(string(out)), nil
}

// imageCommit walks the first-parent history back from rev, and gives the first commit that has a published image.
// It refuses when a commit after the image commit changes an input of the image, because the image of that commit
// is not published.
func imageCommit(ctx context.Context, repo gitRepo, published func(commit string) (bool, error), rev string,
	pathspecs []string,
) (string, error) {
	lastChange, err := repo.run(ctx, append([]string{"log", "-1", "--first-parent", "--format=%H", rev, "--"},
		pathspecs...)...)
	if err != nil {
		return "", err
	}

	if len(lastChange) == 0 {
		return "", fmt.Errorf("no commit at or before %s changes an input of the image", rev)
	}

	commits, err := repo.run(ctx, "rev-list", "--first-parent", rev, "--")
	if err != nil {
		return "", err
	}

	for _, commit := range commits {
		ok, err := published(commit)
		if err != nil {
			return "", err
		}

		if ok {
			return commit, nil
		}

		if commit == lastChange[0] {
			break
		}
	}

	return "", fmt.Errorf("the commit %s changes an input of the image, and no image at or after it is published",
		lastChange[0])
}
