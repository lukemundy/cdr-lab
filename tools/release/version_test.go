package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Verifies: RLS-0005
func TestTheBumpOfACommit(t *testing.T) {
	t.Parallel()

	for message, want := range map[string]bump{
		"feat(register): add a thing":               bumpMinor,
		"fix: repair a thing":                       bumpPatch,
		"docs: explain a thing":                     bumpNone,
		"chore(deps)!: drop a thing":                bumpMajor,
		"refactor: a thing\n\nBREAKING-CHANGE: x":   bumpMajor,
		"Merge branch 'topic'":                      bumpNone,
		"feat:missing space":                        bumpNone,
		"fix: a thing\n\nThe BREAKING CHANGE: text": bumpPatch,
	} {
		assert.Equal(t, want, commitBump(message), message)
	}
}

// Verifies: RLS-0005, RLS-0006
func TestTheVersionOnMajorVersionOne(t *testing.T) {
	t.Parallel()

	h := newHistory(t)
	h.change("feat: a thing", map[string]string{"a.txt": "a"})
	h.git("tag", "v1.2.3")
	h.change("fix!: change a thing", nil)
	head := h.git("rev-parse", "HEAD")

	got, err := nextVersion(t.Context(), h.repo, "v1.2.3", head)

	require.NoError(t, err)
	assert.Equal(t, "v2.0.0", got)
}

// Verifies: RLS-0004
func TestTheEarlierVersionIsTheHighestVersionTag(t *testing.T) {
	t.Parallel()

	h := newHistory(t)
	h.change("feat: a thing", map[string]string{"a.txt": "a"})
	h.git("tag", "v0.9.0")
	h.git("tag", "v0.10.0")
	h.git("tag", "v1.0.0-rc.1")
	h.git("tag", "v2")
	h.git("tag", "1.5.0")
	head := h.change("fix: repair a thing", nil)

	got, err := previousVersion(t.Context(), h.repo, head)

	require.NoError(t, err)
	assert.Equal(t, "v0.10.0", got)
}
