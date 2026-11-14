//go:build conformance

package register

import (
	"crypto/tls"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

func keySetAction(keySet, action string) string {
	return "/v1/key-sets/" + keySet + "/" + action
}

func keyIDs(set *jose.JSONWebKeySet) []string {
	ids := make([]string, 0, len(set.Keys))
	for _, k := range set.Keys {
		ids = append(ids, k.KeyID)
	}

	return ids
}

// signedBy verifies the JWT with the key in the JWKS that has its kid, and gives the kid.
func signedBy(t *testing.T, token string, set *jose.JSONWebKeySet) string {
	t.Helper()

	parsed, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.PS256})
	require.NoError(t, err)

	kid := parsed.Headers[0].KeyID
	keys := set.Key(kid)
	require.Len(t, keys, 1, "the JWKS has the key %s", kid)
	require.NoError(
		t,
		parsed.Claims(keys[0].Key, &map[string]any{}),
		"the signature verifies with the key from the JWKS",
	)

	return kid
}

// publishNextKey completes a key change that an earlier run did not complete, so that the test can run again.
func publishNextKey(t *testing.T, keySet string) {
	t.Helper()

	status, p := operate(t, keySetAction(keySet, "publish-next"))
	if status == http.StatusConflict {
		status, p = operate(t, keySetAction(keySet, "activate-next"))
		require.Equal(t, http.StatusNoContent, status, p.Detail)
		status, p = operate(t, keySetAction(keySet, "publish-next"))
	}

	require.Equal(t, http.StatusNoContent, status, p.Detail)
}

// A data holder gets the Register JWKS again not more than one time each second. The new key is thus published for
// longer than that before it signs, so that a data holder that got the JWKS immediately before the publication can
// get it again for the first JWT of the new key.
const publishedBeforeActive = 1100 * time.Millisecond

// activateNextKey is for a key that the test published immediately before.
func activateNextKey(t *testing.T, keySet string) {
	t.Helper()

	time.Sleep(publishedBeforeActive)

	status, p := operate(t, keySetAction(keySet, "activate-next"))
	require.Equal(t, http.StatusNoContent, status, p.Detail)
}

// keyChange does a full key change, where issue gives a new JWT that the Register signs with the key set at jwksURL.
func keyChange(t *testing.T, keySet, jwksURL string, issue func() string) (before, after string) {
	t.Helper()

	before = issue()
	oldKey := signedBy(t, before, jwksAt(t, jwksURL))

	publishNextKey(t, keySet)
	published := jwksAt(t, jwksURL)
	duringChange := issue()

	require.Len(t, published.Keys, 2, "the JWKS has the old and the new key")
	require.Contains(t, keyIDs(published), oldKey)
	assert.Equal(t, oldKey, signedBy(t, duringChange, published), "the old key signs until the new key is active")

	activateNextKey(t, keySet)

	activated := jwksAt(t, jwksURL)
	after = issue()

	assert.ElementsMatch(t, keyIDs(published), keyIDs(activated), "the JWKS still has the old and the new key")
	newKey := signedBy(t, after, activated)
	assert.NotEqual(t, oldKey, newKey, "the new key signs after its activation")
	assert.Equal(t, oldKey, signedBy(t, before, activated), "a JWT of the old key still verifies")
	assert.Equal(t, oldKey, signedBy(t, duringChange, activated))

	return before, after
}

func TestAChangeOfTheSSAKeyDoesNotBreakVerification(t *testing.T) {
	cert := activeTokenClient(t)
	token := accessToken(t, idpTestActiveProduct, cert)
	jwksURL := labtest.Env(t, "REGISTER_PUBLIC_URL") + registerJWKSPath
	etagBefore := call(t, registerJWKSPath, http.Header{}).Header.Get("ETag")

	_, after := keyChange(t, "ssa", jwksURL, func() string {
		r := secureRequest(
			t,
			cert,
			ssaPath("all", idpTestBrand, idpTestActiveProduct),
			bearerV4(token),
			validatedSSAStatuses,
		)
		require.Equal(t, http.StatusOK, r.status, string(r.body))

		return string(r.body)
	})

	assert.NotEqual(
		t,
		etagBefore,
		call(t, registerJWKSPath, http.Header{}).Header.Get("ETag"),
		"the ETag of Get JWKS changes with the key set",
	)

	requireTheDataHolderAcceptsAnSSAOfTheKeyOf(t, after)
}

// These participants of services/register/seed/example.json are for the registration at the data holder.
const (
	registrationTestRecipient = "6f1d2a57-0c5e-4b0e-9d0a-3f6c1b7e8a41"
	registrationTestBrand     = "b2c4e6a8-1d3f-4a5b-8c7d-9e0f1a2b3c4d"
	keyChangeProduct          = "f6b8c0d2-4a6e-4e8f-9a0b-2c3d4e5f6a71"
)

// requireTheDataHolderAcceptsAnSSAOfTheKeyOf is in this test because only the test that changes the SSA key knows when
// the change is complete: a data holder that keeps the Register JWKS must get it again for the new key, with no
// restart.
func requireTheDataHolderAcceptsAnSSAOfTheKeyOf(t *testing.T, ssaOfNewKey string) {
	t.Helper()

	cert := tokenClient(
		t, registrationTestRecipient, keyChangeProduct, "PKI_REGISTRATION_CLIENT_CERT", "PKI_REGISTRATION_CLIENT_KEY",
	)
	token := accessToken(t, keyChangeProduct, cert)
	ssa := labtest.SoftwareStatement(t, token, cert, registrationTestBrand, keyChangeProduct)
	require.Equal(t, keyID(t, ssaOfNewKey), keyID(t, ssa), "the SSA has the new key")

	brand := labtest.RegisteredDataHolderBrand(t, token, cert)
	configuration := labtest.OpenIDConfiguration(t, brand.EndpointDetail.InfosecBaseURI)
	endpoint, issuer := configuration["registration_endpoint"].(string), configuration["issuer"].(string)

	// The data holder gets the Register JWKS again not more than one time each second.
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		request := labtest.SignJWT(
			t, jose.PS256, recipientKeys(t).RSA, "rsa", labtest.RegistrationClaims(keyChangeProduct, issuer, ssa),
		)
		resp := labtest.PostRegistration(t, endpoint, cert, "application/jwt", request)

		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		require.NoError(c, err)
		assert.Equal(c, http.StatusCreated, resp.StatusCode, string(body))
	}, labtest.PollLimit, 500*time.Millisecond)
}

func keyID(t *testing.T, token string) string {
	t.Helper()

	parsed, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.PS256})
	require.NoError(t, err)

	return parsed.Headers[0].KeyID
}

func TestAChangeOfTheAccessTokenKeyDoesNotBreakVerification(t *testing.T) {
	cert := activeTokenClient(t)
	cfg, _ := openIDConfiguration(t)
	accepted := func(token string) int { return brandsStatus(t, cert, token) }

	before, after := keyChange(
		t,
		"access-token",
		cfg.JWKSURI,
		func() string { return accessToken(t, idpTestActiveProduct, cert) },
	)

	assert.Equal(t, http.StatusOK, accepted(before), "the Register accepts a token of the old key after the key change")
	assert.Equal(t, http.StatusOK, accepted(after), "the Register accepts a token of the new key")
}

func brandsStatus(t *testing.T, cert tls.Certificate, token string) int {
	t.Helper()
	return secureCall(t, cert, dataHolderBrandsPath, bearerV3(token)).status
}

func TestAKeyChangeOfOneKeySetDoesNotChangeTheOtherKeySet(t *testing.T) {
	cfg, _ := openIDConfiguration(t)
	jwksURL := labtest.Env(t, "REGISTER_PUBLIC_URL") + registerJWKSPath
	accessTokenKeys := keyIDs(jwksAt(t, cfg.JWKSURI))

	publishNextKey(t, "ssa")
	activateNextKey(t, "ssa")

	assert.ElementsMatch(t, accessTokenKeys, keyIDs(jwksAt(t, cfg.JWKSURI)))
	assert.Len(t, jwksAt(t, jwksURL).Keys, 2)
}
