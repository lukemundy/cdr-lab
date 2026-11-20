//go:build conformance

package dataholder

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

const (
	// sectorProduct is the only Registration Test software product with a sector_identifier_uri in the seed file.
	sectorProduct = "1c4e6a82-9f3b-4d5e-b7a9-7b8c9d0e1fc6"
	// noAccountScope needs no account selection, so a consent with it needs no account.
	noAccountScope = "openid profile"
	ninetyDays     = 7_776_000
	// oneHour is a sharing duration of a one-time collection, which has a refresh token.
	oneHour = 3600
)

func introspectionEndpoint(t *testing.T, client registeredClient) string {
	t.Helper()

	endpoint := openIDConfiguration(t, client.issuer)["introspection_endpoint"]
	require.IsType(t, "", endpoint)

	return endpoint.(string)
}

// authorisation is a consent of a customer for a software product, before the code exchange.
type authorisation struct {
	client   registeredClient
	claims   map[string]any
	code     string
	verifier string
	// before and after are the times around the confirmation of the consumer.
	before, after time.Time
}

// consents gives a new consent of the customer, who authenticates with a new browser and a new OTP each time.
func (a account) consents(
	client registeredClient,
	scope string,
	sharingDuration any,
	accountIDs ...string,
) authorisation {
	a.t.Helper()

	sent := len(emailsTo(a.t, a.customer.EmailAddress))
	verifier := rand.Text() + rand.Text()
	challenge := sha256.Sum256([]byte(verifier))
	claims := client.claimsFor(scope, sharingDuration)
	claims["code_challenge"] = base64.RawURLEncoding.EncodeToString(challenge[:])

	next := account{t: a.t, customer: a.customer, browser: newConsumer(a.t)}
	otpPage := next.identify(client, claims)
	consent := next.browser.submit(otpPage, "continue", url.Values{"otp": {next.otps(sent + 1)[sent]}})
	require.True(a.t, isConsentPage(consent), consent.body)

	before := time.Now()
	response := client.responseClaims(a.t, next.confirm(consent, accountIDs...))
	require.IsType(a.t, "", response["code"], response)

	return authorisation{
		client: client, claims: claims, code: response["code"].(string), verifier: verifier,
		before: before, after: time.Now(),
	}
}

// tokenRequest sends a form with a new client assertion to an endpoint of the authorisation server.
func (c registeredClient) tokenRequest(t *testing.T, endpoint string, form url.Values) labtest.TokenReply {
	t.Helper()

	request := labtest.TokenForm(t, c.clientID, "", c.assertion(t, endpoint))
	request.Del("grant_type")
	request.Del("scope")

	maps.Copy(request, form)

	return labtest.RequestToken(t, endpoint, c.cert, request)
}

func (a authorisation) exchangeForm() url.Values {
	return url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {a.code},
		"redirect_uri":  {registeredRedirectURI},
		"code_verifier": {a.verifier},
	}
}

func (a authorisation) exchange(t *testing.T) labtest.TokenReply {
	t.Helper()

	return a.client.tokenRequest(t, tokenEndpoint(t, a.client), a.exchangeForm())
}

// tokens gives the token response of a code exchange that succeeds.
func (a authorisation) tokens(t *testing.T) map[string]any {
	t.Helper()

	r := a.exchange(t)
	require.Equal(t, http.StatusOK, r.Status, r.Body)

	return r.Body
}

func (c registeredClient) refresh(t *testing.T, refreshToken string) labtest.TokenReply {
	t.Helper()

	form := url.Values{"grant_type": {"refresh_token"}, "refresh_token": {refreshToken}}

	return c.tokenRequest(t, tokenEndpoint(t, c), form)
}

func (c registeredClient) introspect(t *testing.T, token string) labtest.TokenReply {
	t.Helper()

	return c.tokenRequest(t, introspectionEndpoint(t, c), url.Values{"token": {token}})
}

func (c registeredClient) subjectOf(t *testing.T, tokens map[string]any) string {
	t.Helper()
	require.IsType(t, "", tokens["id_token"], tokens)

	sub := c.signedClaims(t, tokens["id_token"].(string))["sub"]
	require.IsType(t, "", sub)

	return sub.(string)
}

func requireUUID(t *testing.T, value any) string {
	t.Helper()
	require.IsType(t, "", value)

	parsed, err := uuid.Parse(value.(string))
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(4), parsed.Version())

	return value.(string)
}

// requireResourceID checks the format of an accountId or a transactionId: a version 8 UUID from a hash.
func requireResourceID(t *testing.T, value any) string {
	t.Helper()
	require.IsType(t, "", value)

	parsed, err := uuid.Parse(value.(string))
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(8), parsed.Version())

	return value.(string)
}

func certificateThumbprint(t *testing.T, cert tls.Certificate) string {
	t.Helper()

	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	require.NoError(t, err)

	sum := sha256.Sum256(leaf.Raw)

	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func keysOf(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	return keys
}

func requireNotActive(t *testing.T, r labtest.TokenReply) {
	t.Helper()
	require.Equal(t, http.StatusOK, r.Status, r.Body)
	assert.Equal(t, map[string]any{"active": false}, r.Body)
}

// This test is the full sequence of a consent: the operator creates a customer and accounts, a software product
// registers and pushes a request, the consumer authenticates and gives consent, and the software product gets tokens,
// refreshes and introspects.
func TestASoftwareProductGetsTokensForAConsentThenRefreshesAndIntrospects(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newAccount(t)
	own := customer.addAccount("Everyday", "20000070")
	customer.addAccount("Savings", "20000071")

	const scope = "openid bank:accounts.basic:read bank:transactions:read"

	consent := customer.consents(client, scope, ninetyDays, own.AccountID)
	exchanged := consent.exchange(t)
	require.Equal(t, http.StatusOK, exchanged.Status, exchanged.Body)

	tokens := exchanged.Body
	require.IsType(t, "", tokens["access_token"])
	require.IsType(t, "", tokens["refresh_token"])
	refreshToken := tokens["refresh_token"].(string)
	arrangementID := requireUUID(t, tokens["cdr_arrangement_id"])

	t.Run("the token response has the tokens of the consent", func(t *testing.T) {
		assert.Equal(t, "Bearer", tokens["token_type"])
		assert.EqualValues(t, 300, tokens["expires_in"])
		assert.Equal(t, scope, tokens["scope"])
		assert.NotEmpty(t, refreshToken)
		assert.Equal(t, "no-store", exchanged.Header.Get("Cache-Control"))
	})

	t.Run("the access token is bound to the client certificate of the token request", func(t *testing.T) {
		// The test reads the binding in the token, which is a JWT of the data holder.
		claims := client.signedClaims(t, tokens["access_token"].(string))

		assert.Equal(t, map[string]any{"x5t#S256": certificateThumbprint(t, client.cert)}, claims["cnf"])
		assert.Equal(t, scope, claims["scope"])
	})

	t.Run("the ID token has the authentication and no personal information", func(t *testing.T) {
		claims := client.signedClaims(t, tokens["id_token"].(string))

		assert.Equal(t, client.issuer, claims["iss"])
		assert.Equal(t, client.clientID, claims["aud"])
		assert.Equal(t, "urn:cds.au:cdr:2", claims["acr"])
		assert.Equal(t, consent.claims["nonce"], claims["nonce"])
		requireUUID(t, claims["sub"])
		assert.NotEqual(t, customer.customer.CustomerNumber, claims["sub"])

		require.IsType(t, float64(0), claims["auth_time"])
		authenticated := time.Unix(int64(claims["auth_time"].(float64)), 0)
		assert.WithinDuration(t, consent.before, authenticated, time.Minute)
		assert.Equal(t, []string{"acr", "aud", "auth_time", "exp", "iat", "iss", "nonce", "sub"}, keysOf(claims))
	})

	t.Run("a refresh gives new tokens for the same arrangement, and the refresh token stays", func(t *testing.T) {
		for range 2 {
			refreshed := client.refresh(t, refreshToken)

			require.Equal(t, http.StatusOK, refreshed.Status, refreshed.Body)
			assert.NotEmpty(t, refreshed.Body["access_token"])
			assert.NotEqual(t, tokens["access_token"], refreshed.Body["access_token"])
			assert.Equal(t, "Bearer", refreshed.Body["token_type"])
			assert.EqualValues(t, 300, refreshed.Body["expires_in"])
			assert.Equal(t, scope, refreshed.Body["scope"])
			assert.Equal(t, arrangementID, refreshed.Body["cdr_arrangement_id"])
			assert.Equal(t, client.subjectOf(t, tokens), client.subjectOf(t, refreshed.Body))

			if issued, ok := refreshed.Body["refresh_token"]; ok {
				assert.Equal(t, refreshToken, issued, "the data holder does not change the refresh token")
			}
		}
	})

	t.Run("introspection gives the arrangement and the end of the sharing", func(t *testing.T) {
		info := client.introspect(t, refreshToken)

		require.Equal(t, http.StatusOK, info.Status, info.Body)
		assert.Equal(t, []string{"active", "cdr_arrangement_id", "exp", "scope"}, keysOf(info.Body))
		assert.Equal(t, true, info.Body["active"])
		assert.Equal(t, scope, info.Body["scope"])
		assert.Equal(t, arrangementID, info.Body["cdr_arrangement_id"])

		require.IsType(t, float64(0), info.Body["exp"])
		sharingEnds := time.Unix(int64(info.Body["exp"].(float64)), 0)
		sharing := ninetyDays * time.Second
		assert.False(t, sharingEnds.Before(consent.before.Add(sharing).Truncate(time.Second)), sharingEnds)
		assert.False(t, sharingEnds.After(consent.after.Add(sharing)), sharingEnds)
	})

	t.Run("introspection of an access token or an ID token is not active", func(t *testing.T) {
		requireNotActive(t, client.introspect(t, tokens["access_token"].(string)))
		requireNotActive(t, client.introspect(t, tokens["id_token"].(string)))
	})

	t.Run("a later use of the code is refused and ends the refresh token", func(t *testing.T) {
		requireTokenError(t, consent.exchange(t), http.StatusBadRequest, "invalid_grant")
		requireTokenError(t, consent.exchange(t), http.StatusBadRequest, "invalid_grant")
		requireTokenError(t, client.refresh(t, refreshToken), http.StatusBadRequest, "invalid_grant")
		requireNotActive(t, client.introspect(t, refreshToken))
	})
}

func TestARefreshTokenIsIssuedOnlyForASharingDurationOfMoreThanZero(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newEligibleAccount(t)

	tests := map[string]struct {
		sharingDuration any
		want            bool
	}{
		"absent":   {sharingDuration: nil, want: false},
		"zero":     {sharingDuration: 0, want: false},
		"one hour": {sharingDuration: 3600, want: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tokens := customer.consents(client, noAccountScope, tt.sharingDuration).tokens(t)

			assert.NotEmpty(t, tokens["access_token"])
			assert.NotEmpty(t, tokens["id_token"])
			assert.Equal(t, noAccountScope, tokens["scope"])
			requireUUID(t, tokens["cdr_arrangement_id"])
			assert.Equal(t, tt.want, tokens["refresh_token"] != nil, tokens)
		})
	}
}

func TestACodeExchangeThatIsNotValidGivesInvalidGrantAndEndsTheCode(t *testing.T) {
	client := registered(t, allScopesProduct)
	other := registered(t, registrationTestProduct)
	customer := newEligibleAccount(t)

	tests := map[string]func(authorisation) labtest.TokenReply{
		"redirect_uri that is not that of the request": func(a authorisation) labtest.TokenReply {
			form := a.exchangeForm()
			form.Set("redirect_uri", "https://conformance.example/other")

			return client.tokenRequest(t, tokenEndpoint(t, client), form)
		},
		"code_verifier that does not match": func(a authorisation) labtest.TokenReply {
			form := a.exchangeForm()
			form.Set("code_verifier", rand.Text()+rand.Text())

			return client.tokenRequest(t, tokenEndpoint(t, client), form)
		},
		"code of a different client": func(a authorisation) labtest.TokenReply {
			return other.tokenRequest(t, tokenEndpoint(t, other), a.exchangeForm())
		},
	}
	for name, exchange := range tests {
		t.Run(name, func(t *testing.T) {
			consent := customer.consents(client, noAccountScope, ninetyDays)

			requireTokenError(t, exchange(consent), http.StatusBadRequest, "invalid_grant")
			requireTokenError(t, consent.exchange(t), http.StatusBadRequest, "invalid_grant")
		})
	}

	t.Run("code that does not exist", func(t *testing.T) {
		unknown := authorisation{client: client, code: rand.Text(), verifier: rand.Text() + rand.Text()}

		requireTokenError(t, unknown.exchange(t), http.StatusBadRequest, "invalid_grant")
	})
}

func TestARefreshTokenThatTheSoftwareProductCannotUseIsRefusedAndIsNotActive(t *testing.T) {
	client := registered(t, allScopesProduct)
	other := registered(t, registrationTestProduct)
	customer := newEligibleAccount(t)

	t.Run("refresh token of a different client", func(t *testing.T) {
		refreshToken := customer.consents(client, noAccountScope, ninetyDays).tokens(t)["refresh_token"].(string)

		requireTokenError(t, other.refresh(t, refreshToken), http.StatusBadRequest, "invalid_grant")
		requireNotActive(t, other.introspect(t, refreshToken))
		assert.Equal(t, true, client.introspect(t, refreshToken).Body["active"])
	})

	t.Run("refresh token after the end of the sharing duration", func(t *testing.T) {
		refreshToken := customer.consents(client, noAccountScope, 1).tokens(t)["refresh_token"].(string)

		time.Sleep(2 * time.Second)

		requireTokenError(t, client.refresh(t, refreshToken), http.StatusBadRequest, "invalid_grant")
		requireNotActive(t, client.introspect(t, refreshToken))
	})

	t.Run("value that is not a token", func(t *testing.T) {
		requireTokenError(t, client.refresh(t, rand.Text()), http.StatusBadRequest, "invalid_grant")
		requireNotActive(t, client.introspect(t, rand.Text()))
	})
}

func TestIntrospectionRefusesARequestThatIsNotValid(t *testing.T) {
	client := registered(t, allScopesProduct)
	endpoint := introspectionEndpoint(t, client)

	t.Run("no token", func(t *testing.T) {
		requireTokenError(t, client.tokenRequest(t, endpoint, url.Values{}), http.StatusBadRequest, "invalid_request")
	})

	t.Run("client assertion that is not valid", func(t *testing.T) {
		form := url.Values{"token": {rand.Text()}, "client_assertion": {"not-a-jwt"}}

		requireTokenError(t, client.tokenRequest(t, endpoint, form), http.StatusUnauthorized, "invalid_client")
	})
}

func TestTheSubjectOfACustomerIsPairwiseAndStaysTheSame(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newEligibleAccount(t)

	first := customer.consents(client, noAccountScope, ninetyDays).tokens(t)
	second := customer.consents(client, noAccountScope, ninetyDays).tokens(t)
	sub := requireUUID(t, client.subjectOf(t, first))

	t.Run("two consents with one software product have one sub and two arrangements", func(t *testing.T) {
		assert.Equal(t, sub, client.subjectOf(t, second))
		assert.NotEqual(t, first["cdr_arrangement_id"], second["cdr_arrangement_id"])
	})

	t.Run("a different customer has a different sub", func(t *testing.T) {
		otherCustomer := newEligibleAccount(t).consents(client, noAccountScope, ninetyDays).tokens(t)

		assert.NotEqual(t, sub, client.subjectOf(t, otherCustomer))
	})

	// This test registers the software product with the sector identifier, and no other test uses it. The data holder
	// gets the sector identifier document from the JWKS server of this process.
	sector := newRegistrant(t, sectorProduct)

	labtest.ServeSectorIdentifierDocument(http.StatusOK, `["https://conformance.example/other"]`)

	withoutRedirectURI := sector.register(t, sector.claims())

	labtest.ServeSectorIdentifierDocument(http.StatusNotFound, "")

	notAvailable := sector.register(t, sector.claims())

	t.Run(
		"a registration is refused when the sector identifier document does not have its redirect URI",
		func(t *testing.T) {
			assert.Equal(t, http.StatusBadRequest, withoutRedirectURI.status)
			assert.Equal(t, "invalid_redirect_uri", withoutRedirectURI.body["error"])
		},
	)

	t.Run(
		"a registration is refused when the data holder cannot get the sector identifier document",
		func(t *testing.T) {
			assert.Equal(t, http.StatusBadRequest, notAvailable.status)
			assert.Equal(t, "invalid_client_metadata", notAvailable.body["error"])
		},
	)

	labtest.ServeSectorIdentifierDocument(http.StatusOK, `["https://conformance.example/callback"]`)

	sectorClient := registerNew(t, sectorProduct)
	sectorSub := sectorClient.subjectOf(t, customer.consents(sectorClient, noAccountScope, ninetyDays).tokens(t))

	t.Run("a software product with a different sector identifier gets a different sub", func(t *testing.T) {
		requireUUID(t, sectorSub)
		assert.NotEqual(t, sub, sectorSub)
	})

	t.Run("the sub stays the same after the software product registers again", func(t *testing.T) {
		deleted := sectorClient.manageRegistration(
			t, http.MethodDelete, sectorClient.cert, sectorClient.clientID, sectorClient.accessToken(t), "",
		)
		require.Equal(t, http.StatusNoContent, deleted.status, string(deleted.body))

		again := registerNew(t, sectorProduct)
		require.NotEqual(t, sectorClient.clientID, again.clientID)

		tokens := customer.consents(again, noAccountScope, ninetyDays).tokens(t)

		assert.Equal(t, sectorSub, again.subjectOf(t, tokens))
	})
}
