//go:build openid

package openid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	// moduleTimeout is longer than the longest wait of a module of the plan for a callback.
	moduleTimeout = 10 * time.Minute
	pollInterval  = time.Second
)

// suite is a client of the API of the OpenID conformance suite, through its TLS front. The dev profile of the suite
// permits the API with no login.
type suite struct {
	t       *testing.T
	baseURL string
	http    *http.Client
}

type planModule struct {
	TestModule string            `json:"testModule"`
	Variant    map[string]string `json:"variant,omitempty"`
}

type plan struct {
	ID      string       `json:"id"`
	Modules []planModule `json:"modules"`
}

type moduleInfo struct {
	Status string `json:"status"`
	Result string `json:"result"`
}

// The final statuses and the results of a test module of the suite.
const (
	statusConfigured  = "CONFIGURED"
	statusFinished    = "FINISHED"
	statusInterrupted = "INTERRUPTED"
	resultFailed      = "FAILED"
	resultWarning     = "WARNING"
	resultReview      = "REVIEW"
)

func (s suite) request(method, path string, query url.Values, body any, want int, into any) {
	s.t.Helper()

	var reader io.Reader

	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(s.t, err)

		reader = bytes.NewReader(data)
	}

	target := s.baseURL + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(s.t.Context(), method, target, reader)
	require.NoError(s.t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.http.Do(req)
	require.NoError(s.t, err)

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	require.NoError(s.t, err)
	require.Equal(s.t, want, resp.StatusCode, "%s %s: %s", method, path, data)

	if into != nil {
		require.NoError(s.t, json.Unmarshal(data, into), string(data))
	}
}

func (s suite) createPlan(name string, variant map[string]string, configuration any) plan {
	s.t.Helper()

	encoded, err := json.Marshal(variant)
	require.NoError(s.t, err)

	var created plan
	s.request(http.MethodPost, "/api/plan", url.Values{"planName": {name}, "variant": {string(encoded)}},
		configuration, http.StatusCreated, &created)

	return created
}

func (s suite) run(planID string, module planModule) (string, moduleInfo) {
	s.t.Helper()

	query := url.Values{"test": {module.TestModule}, "plan": {planID}}

	if module.Variant != nil {
		encoded, err := json.Marshal(module.Variant)
		require.NoError(s.t, err)
		query.Set("variant", string(encoded))
	}

	var created struct {
		ID string `json:"id"`
	}
	s.request(http.MethodPost, "/api/runner", query, nil, http.StatusCreated, &created)

	started := false
	deadline := time.Now().Add(moduleTimeout)

	for time.Now().Before(deadline) {
		var info moduleInfo
		s.request(http.MethodGet, "/api/info/"+created.ID, nil, nil, http.StatusOK, &info)

		switch info.Status {
		case statusFinished, statusInterrupted:
			return created.ID, info
		case statusConfigured:
			if !started {
				s.request(http.MethodPost, "/api/runner/"+created.ID, nil, nil, http.StatusOK, nil)

				started = true
			}
		}

		time.Sleep(pollInterval)
	}

	return created.ID, moduleInfo{Status: fmt.Sprintf("not finished after %s", moduleTimeout)}
}

// finding is a failure or a warning in the log of a module.
type finding struct {
	Src    string `json:"src"`
	Result string `json:"result"`
	Msg    string `json:"msg"`
}

func (s suite) findings(moduleID string) []finding {
	s.t.Helper()

	var entries []finding
	s.request(http.MethodGet, "/api/log/"+moduleID, nil, nil, http.StatusOK, &entries)

	var found []finding

	for _, entry := range entries {
		if entry.Result == "FAILURE" || entry.Result == resultWarning {
			found = append(found, entry)
		}
	}

	return found
}
