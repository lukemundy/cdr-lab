package authserver_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
)

func TestAConsentTokenGivesTheClientTheScopesTheSubjectAndTheArrangement(t *testing.T) {
	b := newBank(t)
	a := b.authorise(hour)
	tokens := b.exchange(a)

	for _, token := range []string{
		tokens["access_token"].(string),
		b.refresh(tokens["refresh_token"].(string)).body["access_token"].(string),
	} {
		verified, err := b.dataHolder.VerifyConsentToken(t.Context(), token, leaf(t, b.cert))

		require.NoError(t, err)
		assert.Equal(t, authserver.ConsentToken{
			ClientID:      registeredClientID,
			Scopes:        strings.Fields(tokens["scope"].(string)),
			Subject:       a.consent.Subject,
			ArrangementID: a.consent.ArrangementID,
		}, verified)
	}
}

func TestVerifyConsentTokenRefusesATokenThatGivesNoAccessToTheDataOfAConsent(t *testing.T) {
	t.Run("a token of a revoked grant", func(t *testing.T) {
		b := newBank(t)
		a := b.authorise(hour)
		token := b.exchange(a)["access_token"].(string)
		// A second use of the code revokes the grant.
		requireTokenError(t, b.token(b.exchangeForm(a)), "invalid_grant")

		_, err := b.dataHolder.VerifyConsentToken(t.Context(), token, leaf(t, b.cert))

		assert.ErrorIs(t, err, authserver.ErrInvalidToken)
	})

	t.Run("a client credentials token", func(t *testing.T) {
		b := newBank(t)

		_, err := b.dataHolder.VerifyConsentToken(t.Context(), b.accessToken(), leaf(t, b.cert))

		assert.ErrorIs(t, err, authserver.ErrInvalidToken)
	})

	t.Run("a different client certificate", func(t *testing.T) {
		b := newBank(t)
		token := b.exchange(b.authorise(hour))["access_token"].(string)

		_, err := b.dataHolder.VerifyConsentToken(
			t.Context(), token, leaf(t, clientCertificate(t, "Other App")),
		)

		assert.ErrorIs(t, err, authserver.ErrInvalidToken)
	})

	t.Run("a client that is not registered now", func(t *testing.T) {
		b := newBank(t)
		token := b.exchange(b.authorise(hour))["access_token"].(string)
		delete(b.clients, registeredClientID)

		_, err := b.dataHolder.VerifyConsentToken(t.Context(), token, leaf(t, b.cert))

		assert.ErrorIs(t, err, authserver.ErrInvalidToken)
	})

	t.Run("not a token of the data holder", func(t *testing.T) {
		b := newBank(t)

		_, err := b.dataHolder.VerifyConsentToken(t.Context(), "not-a-token", leaf(t, b.cert))

		assert.ErrorIs(t, err, authserver.ErrInvalidToken)
	})
}

func TestTheUserInfoPathIsThePathOfTheUserInfoEndpointInTheConfiguration(t *testing.T) {
	cfg := dataHolderConfig(t)
	dataHolder, err := authserver.NewDataHolder(cfg)
	require.NoError(t, err)

	rec := getDataHolder(t, cfg, "/.well-known/openid-configuration")

	var configuration map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &configuration))
	assert.Equal(t, dataHolderSecure+dataHolder.UserInfoPath(), configuration["userinfo_endpoint"])
}
