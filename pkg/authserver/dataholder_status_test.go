package authserver_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
)

type clientStatuses struct {
	mu       sync.Mutex
	inactive map[string]bool
}

func (c *clientStatuses) Active(_ context.Context, clientID string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return !c.inactive[clientID], nil
}

func (c *clientStatuses) set(clientID string, active bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.inactive == nil {
		c.inactive = map[string]bool{}
	}

	c.inactive[clientID] = !active
}

// Verifies: AUTH-0110
func TestAPushedRequestOfAClientThatIsNotActiveIsRefused(t *testing.T) {
	b := newBank(t)
	b.statuses.set(registeredClientID, false)

	requirePushRefused(t, b.push(b.requestObject()), http.StatusBadRequest, "unauthorized_client")
}

// Verifies: AUTH-0114
func TestAPushedRequestWithAClientAssertionThatIsNotValidGetsInvalidClientBeforeTheStatus(t *testing.T) {
	b := newBank(t)
	b.statuses.set(registeredClientID, false)
	a := b.signedObject(b.requestObject())
	form := b.parForm(a)
	form.Set("client_assertion", "not-a-jwt")

	requirePushRefused(t, b.par(form), http.StatusUnauthorized, "invalid_client")
}

// Verifies: AUTH-0111
func TestAnAuthorisationOfAClientThatIsNotActiveEndsWithUnauthorizedClient(t *testing.T) {
	t.Run("at the authorisation endpoint", func(t *testing.T) {
		b := newBank(t)
		requestURI := b.requestURI()
		b.statuses.set(registeredClientID, false)

		reply := b.browse(http.MethodGet, authorisePath(registeredClientID, requestURI))

		require.Equal(t, http.StatusSeeOther, reply.status, reply.body)
		assert.Equal(t, "unauthorized_client", b.jarmClaims(reply.location)["error"])
		assert.Empty(t, b.pages.interactions, "the consumer sees no page")
	})

	t.Run("at the confirmation", func(t *testing.T) {
		b := newBank(t)
		b.browse(http.MethodGet, authorisePath(registeredClientID, b.requestURI()))
		b.pages.next = func(authserver.Interaction) authserver.Outcome { return authserver.Approve(approved) }
		b.statuses.set(registeredClientID, false)

		reply := b.browse(http.MethodPost, b.pages.interactions[0].Path)

		require.Equal(t, http.StatusSeeOther, reply.status, reply.body)
		claims := b.jarmClaims(reply.location)
		assert.Equal(t, "unauthorized_client", claims["error"])
		assert.NotContains(t, claims, "code")
		assert.Len(t, b.pages.interactions, 1, "the pages do not approve the consent")
	})
}

// Verifies: AUTH-0112
func TestTheCodeAndRefreshGrantsOfAClientThatIsNotActiveAreRefused(t *testing.T) {
	b := newBank(t)
	tokens := b.exchange(b.authorise(hour))
	refreshToken, ok := tokens["refresh_token"].(string)
	require.True(t, ok, tokens)

	code := b.authorise(hour)
	b.statuses.set(registeredClientID, false)

	requireError(t, b.token(b.exchangeForm(code)), http.StatusBadRequest, "unauthorized_client")
	requireError(t, b.refresh(refreshToken), http.StatusBadRequest, "unauthorized_client")

	b.statuses.set(registeredClientID, true)

	refreshed := b.refresh(refreshToken)
	require.Equal(t, http.StatusOK, refreshed.status, refreshed.body)
	assert.NotContains(t, refreshed.body, "refresh_token", "the refresh token does not change")
}

// The status check comes before the library checks that the code belongs to the client.
func TestTheCodeOfAClientThatIsNotActiveIsNotRevokedWhenADifferentClientSendsIt(t *testing.T) {
	b := newBank(t)
	code := b.authorise(hour)
	b.statuses.set(registeredClientID, false)

	a := b.assertion()
	a.claims["iss"], a.claims["sub"] = secondClientID, secondClientID
	form := b.exchangeForm(code)
	form.Set("client_assertion", b.form(a).Get("client_assertion"))

	requireError(t, b.token(form), http.StatusBadRequest, "unauthorized_client")

	b.statuses.set(registeredClientID, true)

	b.exchange(code)
}

// Verifies: AUTH-0113
func TestAClientThatIsNotActiveGetsAClientCredentialsToken(t *testing.T) {
	b := newBank(t)
	b.statuses.set(registeredClientID, false)

	reply := b.token(b.form(b.assertion()))

	require.Equal(t, http.StatusOK, reply.status, reply.body)
}

// Verifies: AUTH-0114
func TestATokenRequestWithAClientAssertionThatIsNotValidGetsInvalidClientBeforeTheStatus(t *testing.T) {
	b := newBank(t)
	code := b.authorise(hour)
	b.statuses.set(registeredClientID, false)

	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	a := b.assertion()
	a.key = otherKey
	form := b.exchangeForm(code)
	form.Set("client_assertion", b.form(a).Get("client_assertion"))

	requireError(t, b.token(form), http.StatusUnauthorized, "invalid_client")
}
