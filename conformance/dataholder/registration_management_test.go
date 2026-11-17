//go:build conformance

package dataholder

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

// lifecycleProduct is only for the test that updates and deletes its registration, so that the registrations that
// the other tests share do not change.
const lifecycleProduct = "b8d0e2f4-6c8a-4a0b-9c2d-4e5f6a7b8c93"

func signedRegistration(t *testing.T, claims map[string]any) string {
	t.Helper()

	return labtest.SignJWT(t, jose.PS256, recipientKeys(t).RSA, "rsa", claims)
}

func registrationOf(t *testing.T, read registrationRead) map[string]any {
	t.Helper()

	var registration map[string]any
	require.NoError(t, json.Unmarshal(read.body, &registration), string(read.body))

	return registration
}

func requireManagementError(t *testing.T, read registrationRead, code string) {
	t.Helper()

	require.Equal(t, http.StatusBadRequest, read.status, string(read.body))
	assert.Equal(t, code, registrationOf(t, read)["error"], string(read.body))
}

// The steps of this test depend on each other, so a step that fails stops the test.
func TestASoftwareProductUpdatesAndDeletesItsRegistrationAndRegistersAgain(t *testing.T) {
	r := newRegistrant(t, lifecycleProduct)
	created := r.register(t, r.claims())
	require.Equal(t, http.StatusCreated, created.status, created.body)

	client := registeredClient{registrant: r, clientID: created.body["client_id"].(string), reply: created}
	token := client.accessToken(t)
	manage := func(method, requestJWT string) registrationRead {
		return client.manageRegistration(t, method, client.cert, client.clientID, token, requestJWT)
	}

	// The arrangement is before the update, because the update changes the algorithm of the request object.
	customer := newEligibleAccount(t)
	tokens := customer.consents(client, noAccountScope, ninetyDays).tokens(t)
	arrangementID := arrangementIDOf(t, tokens)
	require.IsType(t, "", tokens["refresh_token"], tokens)
	refreshToken := tokens["refresh_token"].(string)

	// A new SSA has a new jti, so it is different from the SSA of the registration.
	newSSA := newRegistrant(t, lifecycleProduct).ssa
	require.NotEqual(t, r.ssa, newSSA)

	t.Run("an update that is not valid gives the error of a registration and changes nothing", func(t *testing.T) {
		claims := labtest.RegistrationClaims(lifecycleProduct, r.issuer, newSSA)
		claims["redirect_uris"] = []string{"https://other.example/callback"}

		requireManagementError(t, manage(http.MethodPut, signedRegistration(t, claims)), "invalid_redirect_uri")

		assert.Equal(t, created.body, registrationOf(t, manage(http.MethodGet, "")))
	})

	t.Run("an update with an SSA of a different software product is refused", func(t *testing.T) {
		other := newRegistrant(t, minimalProduct)
		claims := labtest.RegistrationClaims(minimalProduct, r.issuer, other.ssa)

		requireManagementError(t, manage(http.MethodPut, signedRegistration(t, claims)), "invalid_software_statement")

		assert.Equal(t, created.body, registrationOf(t, manage(http.MethodGet, "")))
	})

	t.Run("an update with the SSA of the registration is refused", func(t *testing.T) {
		claims := labtest.RegistrationClaims(lifecycleProduct, r.issuer, r.ssa)

		requireManagementError(t, manage(http.MethodPut, signedRegistration(t, claims)), "invalid_software_statement")

		assert.Equal(t, created.body, registrationOf(t, manage(http.MethodGet, "")))
	})

	t.Run("an update with a content type that is not a JWT gives 415", func(t *testing.T) {
		status := putWithContentType(t, client, token, "application/json")

		assert.Equal(t, http.StatusUnsupportedMediaType, status)
	})

	require.False(t, t.Failed())

	claims := labtest.RegistrationClaims(lifecycleProduct, r.issuer, newSSA)
	claims["grant_types"] = []string{"client_credentials", "authorization_code"}
	claims["request_object_signing_alg"] = "ES256"
	updated := manage(http.MethodPut, signedRegistration(t, claims))
	require.Equal(t, http.StatusOK, updated.status, string(updated.body))

	t.Run("an update replaces the registration and keeps the client ID", func(t *testing.T) {
		registration := registrationOf(t, updated)

		assert.Equal(t, client.clientID, registration["client_id"])
		assert.Equal(t, created.body["client_id_issued_at"], registration["client_id_issued_at"])
		assert.Equal(t, newSSA, registration["software_statement"])
		assert.Equal(t, "ES256", registration["request_object_signing_alg"])
		assert.Equal(t, []any{"client_credentials", "authorization_code"}, registration["grant_types"],
			"a value of the old registration that the new request does not give is not in the registration")
		assert.Equal(t, "no-store", updated.header.Get("Cache-Control"))

		assert.Equal(t, registration, registrationOf(t, manage(http.MethodGet, "")), "GET gives the same values")
	})

	require.False(t, t.Failed())

	deleting := time.Now().Truncate(time.Second)
	deleted := manage(http.MethodDelete, "")
	deletedAt := time.Now()

	require.Equal(t, http.StatusNoContent, deleted.status, string(deleted.body))
	assert.Empty(t, deleted.body)

	t.Run("after a deletion, the token gives 401", func(t *testing.T) {
		read := manage(http.MethodGet, "")

		assert.Equal(t, http.StatusUnauthorized, read.status)
		assert.Equal(t, `Bearer error="invalid_token"`, read.header.Get("WWW-Authenticate"))
	})

	t.Run("after a deletion, the arrangement ends as a revocation, with no notice", func(t *testing.T) {
		got := arrangementOf(t, customer, arrangementID)

		assert.Equal(t, "REVOKED_BY_DATA_RECIPIENT", got.EndReason)
		require.NotNil(t, got.EndedAt)
		assert.WithinRange(t, *got.EndedAt, deleting, deletedAt, "the end has the time of the request")
		assert.Empty(t, got.NoticeStatus)
		assert.Empty(t, labtest.NoticesFor(arrangementID))
	})

	t.Run("after a deletion, client authentication with the client ID fails", func(t *testing.T) {
		tokenReply := client.requestToken(t, client.assertion(t, tokenEndpoint(t, client)))

		requireTokenError(t, tokenReply, http.StatusUnauthorized, "invalid_client")
		requireTokenError(t, client.refresh(t, refreshToken), http.StatusUnauthorized, "invalid_client")
	})

	t.Run("after a deletion, a registration with an SSA that the data holder accepted is refused", func(t *testing.T) {
		requireRegistrationError(
			t,
			r.register(t, labtest.RegistrationClaims(lifecycleProduct, r.issuer, newSSA)),
			"invalid_software_statement",
		)
	})

	t.Run("after a deletion, the software product registers again and gets a new client ID", func(t *testing.T) {
		again := newRegistrant(t, lifecycleProduct)

		reply := again.register(t, again.claims())

		require.Equal(t, http.StatusCreated, reply.status, reply.body)
		assert.NotEqual(t, client.clientID, reply.body["client_id"])
	})
}

func putWithContentType(t *testing.T, client registeredClient, token, contentType string) int {
	t.Helper()

	httpClient := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{
		RootCAs: labtest.CDRCABundle(t), Certificates: []tls.Certificate{client.cert}, MinVersion: tls.VersionTLS12,
	}}}
	req, err := http.NewRequestWithContext(
		t.Context(), http.MethodPut, client.endpoint+"/"+client.clientID, http.NoBody,
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", contentType)

	resp, err := httpClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode
}

func TestAnUpdateOrADeletionWithNoTokenOfTheClientIsRefusedAndChangesNothing(t *testing.T) {
	client := registered(t, registrationTestProduct)
	other := registered(t, minimalProduct)
	token := client.accessToken(t)
	otherCert := labtest.ClientCertificate(t, "PKI_CLIENT_CERT", "PKI_CLIENT_KEY")

	tests := map[string]struct {
		cert            tls.Certificate
		clientID, token string
		status          int
		challenge       string
	}{
		"no token": {client.cert, client.clientID, "", http.StatusUnauthorized, "Bearer"},
		"a token that the data holder did not issue": {
			client.cert, client.clientID, "not-a-token", http.StatusUnauthorized, `Bearer error="invalid_token"`,
		},
		"a different client certificate": {
			otherCert, client.clientID, token, http.StatusUnauthorized, `Bearer error="invalid_token"`,
		},
		"a client ID that does not exist": {
			client.cert, uuid.NewString(), token, http.StatusUnauthorized, `Bearer error="invalid_token"`,
		},
		"the client ID of a different registration": {
			client.cert, other.clientID, token, http.StatusForbidden, "",
		},
	}

	for name, tc := range tests {
		for _, method := range []string{http.MethodPut, http.MethodDelete} {
			t.Run(method+" with "+name, func(t *testing.T) {
				requestJWT := ""
				if method == http.MethodPut {
					requestJWT = signedRegistration(t, client.claims())
				}

				refused := client.manageRegistration(t, method, tc.cert, tc.clientID, tc.token, requestJWT)

				assert.Equal(t, tc.status, refused.status, string(refused.body))
				assert.Equal(t, tc.challenge, refused.header.Get("WWW-Authenticate"))
				assert.Empty(t, refused.body)
			})
		}
	}

	t.Run("the two registrations are the same as before", func(t *testing.T) {
		for _, c := range []registeredClient{client, other} {
			read := c.readRegistration(t, c.cert, c.clientID, c.accessToken(t))

			require.Equal(t, http.StatusOK, read.status, string(read.body))
			assert.Equal(t, c.reply.body, registrationOf(t, read))
		}
	})
}
