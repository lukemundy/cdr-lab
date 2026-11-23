package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ciCommit = "0123456789abcdef0123456789abcdef01234567"

func checkCI(t *testing.T, runs string) error {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/team/cdr-lab/actions/workflows/ci.yml/runs", r.URL.Path)
		assert.Equal(t, ciCommit, r.URL.Query().Get("head_sha"))
		assert.Equal(t, "main", r.URL.Query().Get("branch"))
		assert.Equal(t, "Bearer a-token", r.Header.Get("Authorization"))

		_, _ = w.Write([]byte(`{"workflow_runs": [` + runs + `]}`))
	}))
	t.Cleanup(srv.Close)

	c := ciChecker{client: srv.Client(), api: srv.URL, repo: "team/cdr-lab", workflow: "ci.yml", token: "a-token"}

	return c.check(t.Context(), ciCommit)
}

const (
	succeededRun = `{"event": "push", "status": "completed", "conclusion": "success", ` +
		`"created_at": "2026-11-20T01:00:00Z"}`
	dispatched = `{"event": "workflow_dispatch", "status": "completed", "conclusion": "success", ` +
		`"created_at": "2026-11-20T01:00:00Z"}`
	failedLater = `{"event": "push", "status": "completed", "conclusion": "failure", ` +
		`"created_at": "2026-11-20T02:00:00Z"}`
	runningLater = `{"event": "push", "status": "in_progress", "conclusion": null, ` +
		`"created_at": "2026-11-20T02:00:00Z"}`
	pullRequestLater = `{"event": "pull_request", "status": "completed", "conclusion": "failure", ` +
		`"created_at": "2026-11-20T02:00:00Z"}`
)

// Verifies: RLS-0020
func TestANewestCIRunOnMainThatPassedIsAccepted(t *testing.T) {
	t.Parallel()

	for name, runs := range map[string]string{
		"a passed push run":                 succeededRun,
		"a passed run from the auto-merge":  dispatched,
		"a pull request run does not count": succeededRun + "," + pullRequestLater,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.NoError(t, checkCI(t, runs))
		})
	}
}

// Verifies: RLS-0020
func TestACommitWithNoPassedNewestCIRunOnMainIsRefused(t *testing.T) {
	t.Parallel()

	for name, runs := range map[string]string{
		"no run":                            "",
		"a failed run after the passed run": failedLater + "," + succeededRun,
		"a run that is not complete":        runningLater + "," + succeededRun,
		"a pull request run only":           pullRequestLater,
		"a cancelled run": `{"event": "push", "status": "completed", "conclusion": "cancelled", ` +
			`"created_at": "2026-11-20T01:00:00Z"}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := checkCI(t, runs)

			require.Error(t, err)
			assert.Contains(t, err.Error(), ciCommit)
		})
	}
}
