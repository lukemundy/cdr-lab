//go:build conformance

package register

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

// The suite calls the operator interface only to set up a state for a test of the CDR, and the Register tests the
// interface itself.

// propagation is the maximum time from the end of a status change to its visibility on the APIs.
const propagation = time.Second

type problem struct {
	Detail string `json:"detail"`
}

func operate(t *testing.T, path string) (int, problem) {
	t.Helper()

	req, err := http.NewRequestWithContext(
		t.Context(), http.MethodPost, labtest.Env(t, "REGISTER_INTERNAL_URL")+path, http.NoBody,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+labtest.Env(t, "REGISTER_OPERATOR_TOKEN"))
	resp, err := labtest.Client(t).Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var p problem

	if resp.StatusCode != http.StatusNoContent {
		require.NoError(t, json.Unmarshal(body, &p), string(body))
	}

	return resp.StatusCode, p
}

func recipientAction(id, action string) string {
	return "/v1/data-recipients/" + id + "/" + action
}

func productAction(recipient, product, action string) string {
	return "/v1/data-recipients/" + recipient + "/software-products/" + product + "/" + action
}
