package main

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	versionPattern = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
	headerPattern  = regexp.MustCompile(`^([a-zA-Z]+)(\([^()]*\))?(!)?: \S`)
	breakingFooter = regexp.MustCompile(`(?m)^BREAKING[ -]CHANGE: `)
)

type bump int

const (
	bumpNone bump = iota
	bumpPatch
	bumpMinor
	bumpMajor
)

// commitBump gives the part of the version that a Conventional Commit message changes. A message that is not a
// Conventional Commit, for example a merge commit, changes no part.
func commitBump(message string) bump {
	header, _, _ := strings.Cut(message, "\n")

	m := headerPattern.FindStringSubmatch(header)

	switch {
	case m == nil:
		return bumpNone
	case m[3] == "!" || breakingFooter.MatchString(message):
		return bumpMajor
	case strings.EqualFold(m[1], "feat"):
		return bumpMinor
	case strings.EqualFold(m[1], "fix"):
		return bumpPatch
	default:
		return bumpNone
	}
}

// previousVersion gives the highest version tag that is an ancestor of the commit, or "" when there is none.
func previousVersion(ctx context.Context, repo gitRepo, commit string) (string, error) {
	tags, err := repo.run(ctx, "tag", "--list", "--merged", commit, "v*")
	if err != nil {
		return "", err
	}

	var best string

	for _, tag := range tags {
		if versionPattern.MatchString(tag) && (best == "" || compareVersions(tag, best) > 0) {
			best = tag
		}
	}

	return best, nil
}

func parseVersion(v string) [3]int {
	var parts [3]int

	for i, s := range versionPattern.FindStringSubmatch(v)[1:] {
		parts[i], _ = strconv.Atoi(s)
	}

	return parts
}

func compareVersions(a, b string) int {
	pa, pb := parseVersion(a), parseVersion(b)

	for i := range pa {
		if pa[i] != pb[i] {
			return pa[i] - pb[i]
		}
	}

	return 0
}

// nextVersion calculates the version from the Conventional Commits after the earlier version tag, up to the commit.
func nextVersion(ctx context.Context, repo gitRepo, previous, commit string) (string, error) {
	version, err := calculateVersion(ctx, repo, previous, commit)
	if err != nil {
		return "", err
	}

	if _, err := repo.commitID(ctx, "refs/tags/"+version); err == nil {
		return "", fmt.Errorf("the version %s already has the git tag %s", version, version)
	}

	return version, nil
}

func calculateVersion(ctx context.Context, repo gitRepo, previous, commit string) (string, error) {
	if previous == "" {
		return "v0.1.0", nil
	}

	out, err := repo.output(ctx, "log", "--first-parent", "--format=%B%x00", previous+".."+commit, "--")
	if err != nil {
		return "", err
	}

	level := bumpNone
	for message := range strings.SplitSeq(string(out), "\x00") {
		level = max(level, commitBump(strings.TrimSpace(message)))
	}

	v := parseVersion(previous)

	if level == bumpMajor && v[0] == 0 {
		level = bumpMinor
	}

	switch level {
	case bumpMajor:
		v = [3]int{v[0] + 1, 0, 0}
	case bumpMinor:
		v = [3]int{v[0], v[1] + 1, 0}
	case bumpPatch:
		v[2]++
	case bumpNone:
		return "", fmt.Errorf("no commit after %s changes the version", previous)
	}

	return fmt.Sprintf("v%d.%d.%d", v[0], v[1], v[2]), nil
}
