package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	imageDigest = "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	chartDigest = "sha256:2222222222222222222222222222222222222222222222222222222222222222"
)

func samplePlan() releasePlan {
	return releasePlan{
		Mode:            modeRelease,
		GitHubRepo:      "team/cdr-lab",
		Commit:          "0123456789abcdef0123456789abcdef01234567",
		Version:         "v0.2.0",
		PreviousVersion: "v0.1.0",
		ImageTag:        "v0.2.0",
		ChartVersion:    "0.2.0",
		ChartRepository: "oci://registry.test/team/charts",
		Images: []planImage{{
			Name: "register", Commit: "89abcdef0123456789abcdef0123456789abcdef", Digest: imageDigest,
			Source: "registry.test/team/register", Target: "registry.test/team/register",
		}},
		Charts:    []string{"register"},
		Fresh:     freshDeployment{Files: []string{"services/register/internal/migrations/00001_create.sql"}},
		Changelog: "### Features\n- **register:** add a thing (89abcde)\n",
	}
}

func notes(t *testing.T, p releasePlan, digests map[string]string) string {
	t.Helper()

	var out strings.Builder
	require.NoError(t, writeNotes(&out, p, digests))

	return out.String()
}

// Verifies: RLS-0063, RLS-0064, RLS-0067, RLS-0068, RLS-0069, RLS-0070, RLS-0071
func TestTheReleaseNotes(t *testing.T) {
	t.Parallel()

	got := notes(t, samplePlan(), map[string]string{"register": chartDigest})

	assert.True(t, strings.HasPrefix(got, "## Needs a fresh deployment\n"), got)
	assert.Contains(t, got, "- `services/register/internal/migrations/00001_create.sql`")
	assert.Contains(t, got, "- **register:** add a thing (89abcde)")
	assert.Contains(t, got,
		"| `registry.test/team/register:v0.2.0` | `89abcdef0123456789abcdef0123456789abcdef` | `"+imageDigest+"` |")
	assert.Contains(t, got, "| `oci://registry.test/team/charts/register` | `0.2.0` | `"+chartDigest+"` |")
	assert.Contains(t, got, "does not list the npm packages")
	assert.Contains(t, got, "https://github.com/team/cdr-lab/blob/0123456789abcdef0123456789abcdef01234567/"+
		"charts/register/README.md#verify-a-release")
	assert.NotContains(t, got, "published nothing")
}

// Verifies: RLS-0066
func TestTheFreshDeploymentSectionWithNoChange(t *testing.T) {
	t.Parallel()

	p := samplePlan()
	p.Fresh = freshDeployment{}
	assert.Contains(t, notes(t, p, map[string]string{"register": chartDigest}), "does not need a fresh deployment")

	p.Fresh = freshDeployment{First: true}
	p.PreviousVersion = ""
	assert.Contains(t, notes(t, p, map[string]string{"register": chartDigest}), "is the first release")
}

// Verifies: RLS-0013
func TestTheNotesOfACheckSayThatTheRunPublishedNothing(t *testing.T) {
	t.Parallel()

	p := samplePlan()
	p.Mode = modeCheck

	got := notes(t, p, nil)

	assert.Contains(t, got, "published nothing")
	assert.Contains(t, got, "| `oci://registry.test/team/charts/register` | `0.2.0` | not published |")
}

func TestTheNotesOfAReleaseNeedTheDigestOfEachChart(t *testing.T) {
	t.Parallel()

	var out strings.Builder

	err := writeNotes(&out, samplePlan(), map[string]string{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "register")
}
