package console_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/console"
)

// Verifies: DH-1621
func TestABodyLargerThanTheLimitOfItsOperationGives413(t *testing.T) {
	tests := map[string]struct {
		target string
		limit  int
		detail string
	}{
		"customer create":    {"/api/customers", 64 << 10, "64 KiB"},
		"account create":     {"/api/customers/00000001/accounts", 64 << 10, "64 KiB"},
		"transaction create": {"/api/customers/00000001/accounts/account-1/transactions", 64 << 10, "64 KiB"},
		"close":              {closeEveryday, 16 << 10, "16 KiB"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resp := post(t, console.Ports{}, tt.target, `{"padding": "`+strings.Repeat("x", tt.limit)+`"}`)

			require.Equal(t, http.StatusRequestEntityTooLarge, resp.status, resp.body)

			var body struct {
				Detail string `json:"detail"`
			}
			require.NoError(t, json.Unmarshal([]byte(resp.body), &body))
			assert.Equal(t, "The request body is larger than "+tt.detail+".", body.Detail)
		})
	}
}
