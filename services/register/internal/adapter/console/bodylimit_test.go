package console_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/console"
)

// Verifies: REG-1110, REG-1493
func TestABodyLargerThanTheLimitOfItsActionGives413(t *testing.T) {
	tests := map[string]struct {
		target string
		limit  int
		detail string
	}{
		"create":                   {createDataRecipient, 1 << 20, "1 MiB"},
		"add":                      {addBrand, 1 << 20, "1 MiB"},
		"change":                   {changeDetails, 1 << 20, "1 MiB"},
		"status action":            {"/api/data-recipients/dr-1/suspend", 1 << 10, "1 KiB"},
		"data holder action":       {"/api/data-holders/dh-1/remove", 1 << 10, "1 KiB"},
		"key set action":           {"/api/key-sets/ssa/publish-next", 1 << 10, "1 KiB"},
		"certificate registration": {certificates, 32 << 10, "32 KiB"},
		"issue":                    {issueRoutes["client"].path, 32 << 10, "32 KiB"},
		"revocation check":         {checkRevocation, 32 << 10, "32 KiB"},
		"revocation":               {revoke, 32 << 10, "32 KiB"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			body := `{"version": 1, "padding": "` + strings.Repeat("x", tt.limit) + `"}`

			resp := post(t, console.Ports{}, tt.target, body)

			require.Equal(t, http.StatusRequestEntityTooLarge, resp.status, resp.body)
			assert.Equal(t, "The request body is larger than "+tt.detail+".", problemDetail(t, resp))
		})
	}
}
