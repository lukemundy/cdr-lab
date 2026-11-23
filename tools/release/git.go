package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type gitRepo struct {
	dir string
}

func (g gitRepo) output(ctx context.Context, args ...string) ([]byte, error) {
	//nolint:gosec // The arguments are revisions, and pathspecs after "--", so git cannot read a pathspec as a flag.
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", g.dir}, args...)...)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}

	return out, nil
}

func (g gitRepo) run(ctx context.Context, args ...string) ([]string, error) {
	out, err := g.output(ctx, args...)
	if err != nil {
		return nil, err
	}

	return strings.Fields(string(out)), nil
}

func (g gitRepo) commitID(ctx context.Context, rev string) (string, error) {
	out, err := g.run(ctx, "rev-parse", "--verify", "--end-of-options", rev+"^{commit}")
	if err != nil {
		return "", err
	}

	return out[0], nil
}

// file gives the content of a file at a commit, and false when the commit does not have the file.
func (g gitRepo) file(ctx context.Context, commit, path string) ([]byte, bool, error) {
	//nolint:gosec // The object name is a commit ID and a path of the repository.
	err := exec.CommandContext(ctx, "git", "-C", g.dir, "cat-file", "-e", commit+":"+path).Run()

	if _, ok := errors.AsType[*exec.ExitError](err); ok {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf("git cat-file: %w", err)
	}

	out, err := g.output(ctx, "show", commit+":"+path)

	return out, err == nil, err
}
