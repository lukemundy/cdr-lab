package cds_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/cds"
)

// Verifies: CDSL-0034, CDSL-0035
func TestNewResponseGivesTheDataWithALinkToTheRequestAndAnEmptyMeta(t *testing.T) {
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/cds-au/v1/things/a%20b?ignored=1", nil)

	body, err := json.Marshal(cds.NewResponse(req, "https://api.example", []string{"a"}))

	require.NoError(t, err)
	assert.JSONEq(t, `{
		"data": ["a"],
		"links": {"self": "https://api.example/cds-au/v1/things/a%20b"},
		"meta": {}
	}`, string(body))
}
