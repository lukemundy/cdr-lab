package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// ciChecker reads the runs of the CI workflow from the GitHub REST API. The token is optional for a public repository.
type ciChecker struct {
	client   *http.Client
	api      string
	repo     string
	workflow string
	token    string
}

type workflowRun struct {
	Event      string    `json:"event"`
	Status     string    `json:"status"`
	Conclusion string    `json:"conclusion"`
	CreatedAt  time.Time `json:"created_at"`
	HTMLURL    string    `json:"html_url"`
}

// check refuses the commit unless the newest CI run on main for the commit passed. A run on main starts from a push,
// or from the auto-merge workflow with workflow_dispatch.
func (c ciChecker) check(ctx context.Context, commit string) error {
	u, err := url.Parse(c.api + "/repos/" + c.repo + "/actions/workflows/" + c.workflow + "/runs")
	if err != nil {
		return err
	}

	u.RawQuery = url.Values{"head_sha": {commit}, "branch": {"main"}, "per_page": {"100"}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("get the CI runs of %s: %w", commit, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get the CI runs of %s: %s", commit, resp.Status)
	}

	var body struct {
		Runs []workflowRun `json:"workflow_runs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("read the CI runs of %s: %w", commit, err)
	}

	var newest *workflowRun

	for i, run := range body.Runs {
		if run.Event != "push" && run.Event != "workflow_dispatch" {
			continue
		}

		if newest == nil || run.CreatedAt.After(newest.CreatedAt) {
			newest = &body.Runs[i]
		}
	}

	switch {
	case newest == nil:
		return fmt.Errorf("the commit %s has no CI run on main", commit)
	case newest.Status != "completed":
		return fmt.Errorf("the newest CI run on main of the commit %s is not complete: %s", commit, newest.HTMLURL)
	case newest.Conclusion != "success":
		return fmt.Errorf("the newest CI run on main of the commit %s did not pass (%s): %s", commit,
			newest.Conclusion, newest.HTMLURL)
	default:
		return nil
	}
}
