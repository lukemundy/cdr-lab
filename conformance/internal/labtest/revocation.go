package labtest

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Revoke(t *testing.T, certVariable, reason string) {
	t.Helper()

	status, body := revoke(t, certVariable, reason)
	require.Equal(t, http.StatusNoContent, status, body)
}

// RevokedClientCertificate revokes the fixture first, and a later test gets 409 because the fixture is already revoked.
func RevokedClientCertificate(t *testing.T) tls.Certificate {
	t.Helper()

	status, body := revoke(t, "PKI_REVOKED_CLIENT_CERT", "keyCompromise")
	require.Contains(t, []int{http.StatusNoContent, http.StatusConflict}, status, body)

	return ClientCertificate(t, "PKI_REVOKED_CLIENT_CERT", "PKI_REVOKED_CLIENT_KEY")
}

func revoke(t *testing.T, certVariable, reason string) (status int, body string) {
	t.Helper()

	certificate, err := os.ReadFile(Env(t, certVariable))
	require.NoError(t, err)

	reqBody, err := json.Marshal(map[string]string{"certificate": string(certificate), "reason": reason})
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(
		t.Context(), http.MethodPost, Env(t, "REGISTER_OPERATOR_URL")+"/v1/cdr-certificates/revoke",
		bytes.NewReader(reqBody),
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+Env(t, "REGISTER_OPERATOR_TOKEN"))

	resp, err := Client(t).Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, string(data)
}
