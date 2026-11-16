package cdsrules_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Verifies: AUTH-0104, AUTH-0105, AUTH-0106, AUTH-0107
func TestTheAuthorisationResponseIsASignedJARMResponse(t *testing.T) {
	server, recipient := newAuthServer(t)

	a := recipient.authorise(nil)

	require.Equal(t, http.StatusCreated, a.par.status, a.par.body)
	assert.Equal(t, server.issuer, a.jarm["iss"])
	assert.Equal(t, clientID, a.jarm["aud"])
	assert.Equal(t, "state-1", a.jarm["state"])
	assert.NotEmpty(t, a.jarm["code"])
	exp, ok := a.jarm["exp"].(float64)
	require.True(t, ok, "the JARM response has exp")
	assert.WithinDuration(t, time.Now(), time.Unix(int64(exp), 0), 10*time.Minute)
}

func TestTheJARMCodeGivesACertificateBoundAccessToken(t *testing.T) {
	server, recipient := newAuthServer(t)

	tokens := recipient.exchange(recipient.authorise(nil))

	info, _, err := server.op.Introspect(t.Context(), tokens["access_token"].(string))
	require.NoError(t, err)
	assert.True(t, info.IsActive)
	require.NotNil(t, info.Confirmation)
	assert.NotEmpty(t, info.Confirmation.CertThumbprint, "the token is bound to the client certificate")
}
