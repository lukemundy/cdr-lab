package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixture is a repository of one service and one chart, with a registry of its images and a CI result.
type fixture struct {
	t         *testing.T
	h         history
	published map[string]string
	ci        error
	registry  *httptest.Server
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	f := &fixture{t: t, h: newHistory(t), published: map[string]string{}}
	f.registry = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		digest, ok := f.published[strings.TrimPrefix(r.URL.Path, "/v2/team/register/manifests/")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.Header().Set("Docker-Content-Digest", digest)
	}))
	t.Cleanup(f.registry.Close)

	f.h.change("chore: start", map[string]string{
		"charts/register/Chart.yaml": "name: register\n",
		"charts/README.md":           "The charts.\n",
		"services/register/main.go":  "package main\n",
	})

	return f
}

// image commits a change to an input of the image, and publishes the image of that commit.
func (f *fixture) image(message string) string {
	f.t.Helper()

	commit := f.h.change(message, map[string]string{"services/register/main.go": "package main // " + message})
	sum := sha256.Sum256([]byte(commit))
	f.published[commit] = "sha256:" + hex.EncodeToString(sum[:])

	return commit
}

func (f *fixture) config(mode releaseMode) planConfig {
	f.t.Helper()

	template, err := filepath.Abs("changelog.tera")
	require.NoError(f.t, err)

	return planConfig{
		repo:        f.h.repo,
		mainRef:     "main",
		mode:        mode,
		runID:       "123",
		imagePrefix: strings.TrimPrefix(f.registry.URL, "https://") + "/team",
		githubRepo:  "team/cdr-lab",
		inputs:      map[string][]string{"register": testInputs},
		template:    template,
		ci:          func(context.Context, string) error { return f.ci },
		client:      f.registry.Client(),
	}
}

func (f *fixture) plan(mode releaseMode) (releasePlan, error) {
	f.t.Helper()

	return makePlan(f.t.Context(), f.config(mode), "HEAD")
}

// Verifies: RLS-0006, RLS-0009, RLS-0068
func TestAReleaseWithNoEarlierVersionIsTheFirstVersion(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	image := f.image("feat(register): add a thing")

	got, err := f.plan(modeCheck)

	require.NoError(t, err)
	assert.Equal(t, "v0.1.0", got.Version)
	assert.Empty(t, got.PreviousVersion)
	assert.True(t, got.Fresh.First)
	assert.Equal(t, []planImage{{
		Name: "register", Commit: image, Digest: f.published[image],
		Source: f.config(modeCheck).imagePrefix + "/register", Target: f.config(modeCheck).imagePrefix + "/register",
	}}, got.Images)
	assert.Equal(t, []string{"register"}, got.Charts)
}

// Verifies: RLS-0005
func TestTheVersionComesFromTheCommitsAfterTheEarlierVersion(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		message string
		want    string
	}{
		"a fix gives the next patch version":                {"fix(register): repair a thing", "v0.1.1"},
		"a feat gives the next minor version":               {"feat(register): add a thing", "v0.2.0"},
		"a breaking change on major version 0 gives minor":  {"feat(register)!: remove a thing", "v0.2.0"},
		"a breaking change in the footer gives minor on 0.": {"fix: a thing\n\nBREAKING CHANGE: gone", "v0.2.0"},
		"a breaking fix gives minor on major version 0":     {"fix(register)!: change a thing", "v0.2.0"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newFixture(t)
			f.image("feat(register): add a thing")
			f.h.git("tag", "v0.1.0")
			f.image(tc.message)

			got, err := f.plan(modeCheck)

			require.NoError(t, err)
			assert.Equal(t, tc.want, got.Version)
			assert.Equal(t, "v0.1.0", got.PreviousVersion)
		})
	}
}

// Verifies: RLS-0007, RLS-0024
func TestACommitThatChangesNoVersionIsRefused(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	f.image("feat(register): add a thing")
	f.h.git("tag", "v0.1.0")
	f.h.change("docs: explain a thing", map[string]string{"charts/README.md": "More.\n"})

	_, err := f.plan(modeCheck)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no commit after v0.1.0 changes the version")
}

// Verifies: RLS-0008, RLS-0024
func TestAVersionThatHasATagIsRefused(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	f.image("feat(register): add a thing")
	f.h.git("tag", "v0.1.0")
	release := f.image("feat(register): add a second thing")
	f.h.change("docs: explain a thing", nil)
	f.h.git("tag", "v0.2.0")
	f.h.git("checkout", "--quiet", "--detach", release)

	_, err := f.plan(modeCheck)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already has the git tag")
}

// Verifies: RLS-0002, RLS-0024
func TestACommitThatIsNotInTheFirstParentHistoryOfMainIsRefused(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	f.image("feat(register): add a thing")
	f.h.git("checkout", "--quiet", "-b", "topic")
	f.image("feat(register): add a topic thing")
	f.h.git("checkout", "--quiet", "main")
	f.h.change("fix(register): repair a thing", nil)
	f.h.git("merge", "--quiet", "--no-ff", "--message", "chore: merge", "topic")
	f.h.git("checkout", "--quiet", "--detach", "topic")

	_, err := f.plan(modeCheck)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "first-parent history")
}

// Verifies: RLS-0020, RLS-0024
func TestACommitWithNoPassingCIRunIsRefused(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	f.image("feat(register): add a thing")
	f.ci = errors.New("the CI run failed")

	_, err := f.plan(modeCheck)

	require.ErrorIs(t, err, f.ci)
}

// Verifies: RLS-0022, RLS-0024
func TestACommitWithAnImageThatIsNotUpToDateIsRefused(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	f.image("feat(register): add a thing")
	f.h.change("fix(register): repair a thing", map[string]string{"services/register/main.go": "package main // new"})

	_, err := f.plan(modeCheck)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "register")
}

func TestTheReleaseCommitMustBeCheckedOut(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	old := f.image("feat(register): add a thing")
	f.image("feat(register): add a second thing")

	_, err := makePlan(t.Context(), f.config(modeCheck), old)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "check out")
}

// Verifies: RLS-0030, RLS-0040, RLS-0041
func TestAReleaseGivesTheVersionToTheImagesAndTheCharts(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	f.image("feat(register): add a thing")

	got, err := f.plan(modeRelease)

	require.NoError(t, err)

	prefix := f.config(modeRelease).imagePrefix

	assert.Equal(t, "v0.1.0", got.ImageTag)
	assert.Equal(t, "0.1.0", got.ChartVersion)
	assert.Equal(t, "oci://"+prefix+"/charts", got.ChartRepository)
	assert.Equal(t, prefix+"/register", got.Images[0].Target)
	assert.Equal(t, prefix, got.ImagePrefix)
}

// Verifies: RLS-0014, RLS-0015
func TestARehearsalUsesTheRehearsalNamesOnly(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	f.image("feat(register): add a thing")

	got, err := f.plan(modeRehearse)

	require.NoError(t, err)

	prefix := f.config(modeRehearse).imagePrefix

	assert.Equal(t, "v0.1.0", got.Version)
	assert.Equal(t, "rehearsal-123", got.ImageTag)
	assert.Equal(t, "0.0.0-rehearsal.123", got.ChartVersion)
	assert.Equal(t, "oci://"+prefix+"/rehearsal/charts", got.ChartRepository)
	assert.Equal(t, prefix+"/register", got.Images[0].Source)
	assert.Equal(t, prefix+"/rehearsal/register", got.Images[0].Target)
	assert.Equal(t, prefix+"/rehearsal", got.ImagePrefix)
}

// Verifies: RLS-0012
func TestAnUnknownModeOrARehearsalWithNoRunIDIsRefused(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	f.image("feat(register): add a thing")

	_, err := f.plan("publish")
	require.Error(t, err)

	cfg := f.config(modeRehearse)
	cfg.runID = ""
	_, err = makePlan(t.Context(), cfg, "HEAD")
	require.Error(t, err)
}

// Verifies: RLS-0067
func TestTheChangelogGroupsTheCommitsByTypeThenScope(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	f.image("feat(register): add a thing")
	f.h.git("tag", "v0.1.0")
	f.h.change("fix(release): repair the notes", nil)
	f.h.change("feat(register)!: remove a thing", nil)
	f.h.change("feat(dataholder): add a bank", nil)
	f.image("fix: repair a thing")

	got, err := f.plan(modeCheck)

	require.NoError(t, err)
	assert.NotContains(t, got.Changelog, "add a thing", "a commit before the earlier version")
	assert.Regexp(t, `(?s)### Bug Fixes\n- \*\*release:\*\* repair the notes \([0-9a-f]{7}\)\n- repair a thing .*`+
		`### Features\n- \*\*dataholder:\*\* add a bank .*\n- \*\*register:\*\* \*\*BREAKING\*\* remove a thing`,
		got.Changelog)
}

func TestTheFreshDeployment(t *testing.T) {
	t.Parallel()

	const (
		migration = "services/register/internal/migrations/00001_create.sql"
		golden    = "services/register/internal/app/testdata/events.golden.json"
	)

	for name, tc := range map[string]struct {
		change map[string]string
		want   []string
	}{
		// Verifies: RLS-0065, RLS-0066
		"a new migration and a new entry": {
			map[string]string{
				"services/register/internal/migrations/00002_more.sql": "CREATE TABLE b ();",
				golden: `{"s/A": {"schemaVersion": 1, "payload": {"a": 1}}, "s/B": {"schemaVersion": 1}}`,
			},
			nil,
		},
		// Verifies: RLS-0066
		"a new format of the same entry": {
			map[string]string{golden: "{\n  \"s/A\": {\"payload\": {\"a\": 1}, \"schemaVersion\": 1}\n}"},
			nil,
		},
		// Verifies: RLS-0064
		"a changed migration": {map[string]string{migration: "CREATE TABLE a (b int);"}, []string{migration}},
		// Verifies: RLS-0064
		"a deleted migration": {map[string]string{migration: ""}, []string{migration}},
		// Verifies: RLS-0064
		"a renamed migration": {
			map[string]string{
				migration: "", "services/register/internal/migrations/00009_create.sql": "CREATE TABLE a ();",
			},
			[]string{migration},
		},
		// Verifies: RLS-0064
		"a changed entry": {
			map[string]string{golden: `{"s/A": {"schemaVersion": 1, "payload": {"a": 2}}}`},
			[]string{golden + ": s/A"},
		},
		// Verifies: RLS-0064
		"a removed entry": {map[string]string{golden: `{}`}, []string{golden + ": s/A"}},
		// Verifies: RLS-0064
		"a deleted event golden file": {map[string]string{golden: ""}, []string{golden}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newFixture(t)
			f.h.change("feat(register): add the store", map[string]string{
				migration: "CREATE TABLE a ();",
				golden:    `{"s/A": {"schemaVersion": 1, "payload": {"a": 1}}}`,
			})
			f.image("feat(register): add a thing")
			f.h.git("tag", "v0.1.0")
			f.h.change("feat(register): change the store", tc.change)
			f.image("fix(register): repair a thing")

			got, err := f.plan(modeCheck)

			require.NoError(t, err)
			assert.False(t, got.Fresh.First)
			assert.Equal(t, tc.want, got.Fresh.Files)
		})
	}
}
