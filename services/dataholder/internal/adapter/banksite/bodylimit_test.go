package banksite_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func paddedBody(size int) string {
	return `{"padding": "` + strings.Repeat("x", size) + `"}`
}

// Verifies: DH-1720
func TestABodyLargerThanTheLimitOfItsOperationGives413(t *testing.T) {
	tests := map[string]struct {
		send   func(t *testing.T) response
		detail string
	}{
		"sign-up": {func(t *testing.T) response {
			return newSite(t, newCustomers()).do(request{
				method: http.MethodPost, target: "/api/customers", body: paddedBody(4 << 10),
			})
		}, "4 KiB"},
		"sign-in": {func(t *testing.T) response {
			return newSite(t, newCustomers(quoll)).do(request{
				method: http.MethodPost, target: "/api/session", body: paddedBody(4 << 10),
			})
		}, "4 KiB"},
		"open": {func(t *testing.T) response {
			return signedInOpen(t, &opener{}, paddedBody(4<<10))
		}, "4 KiB"},
		"close": {func(t *testing.T) response {
			return signedInClose(t, customerOfSession(), "account-1", paddedBody(16<<10))
		}, "16 KiB"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resp := tt.send(t)

			require.Equal(t, http.StatusRequestEntityTooLarge, resp.status, resp.body)

			var body struct {
				Detail string `json:"detail"`
			}
			require.NoError(t, json.Unmarshal([]byte(resp.body), &body))
			assert.Equal(t, "The request body is larger than "+tt.detail+".", body.Detail)
		})
	}
}
