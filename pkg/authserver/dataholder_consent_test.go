package authserver_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
)

type grantStore struct {
	mu     sync.Mutex
	grants map[string]authserver.Grant
}

func (s *grantStore) SaveGrant(_ context.Context, grant authserver.Grant) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.grants[grant.ID] = grant

	return nil
}

func (s *grantStore) find(match func(authserver.Grant) bool) (authserver.Grant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, grant := range s.grants {
		if match(grant) {
			return grant, nil
		}
	}

	return authserver.Grant{}, authserver.ErrGrantNotFound
}

func (s *grantStore) Grant(_ context.Context, id string) (authserver.Grant, error) {
	return s.find(func(g authserver.Grant) bool { return g.ID == id })
}

func (s *grantStore) GrantByCode(_ context.Context, code string) (authserver.Grant, error) {
	return s.find(func(g authserver.Grant) bool { return g.Code == code })
}

func (s *grantStore) GrantByRefreshToken(_ context.Context, token string) (authserver.Grant, error) {
	return s.find(func(g authserver.Grant) bool { return g.RefreshToken != "" && g.RefreshToken == token })
}

// only gives the one grant of a test.
func (s *grantStore) only(t *testing.T) authserver.Grant {
	t.Helper()
	require.Len(t, s.grants, 1)

	for _, grant := range s.grants {
		return grant
	}

	return authserver.Grant{}
}

// authorisation is the result of a flow in which the consumer gave consent.
type authorisation struct {
	consent  authserver.Consent
	code     string
	verifier string
}

// Sharing durations in seconds.
const (
	noSharing = 0
	hour      = 3600
)

// authorise pushes a request with the sharing duration in seconds, and gives consent to it.
func (b *bank) authorise(sharingDuration int) authorisation {
	b.t.Helper()

	a := authorisation{verifier: rand.Text() + rand.Text()}
	challenge := sha256.Sum256([]byte(a.verifier))
	object := b.requestObject()
	object["code_challenge"] = base64.RawURLEncoding.EncodeToString(challenge[:])
	object["claims"] = map[string]any{"sharing_duration": sharingDuration}

	pushed := b.push(object)
	require.Equal(b.t, http.StatusCreated, pushed.status, pushed.body)

	b.pages.next = func(in authserver.Interaction) authserver.Outcome {
		now := time.Now().Truncate(time.Second)
		a.consent = authserver.Consent{
			ArrangementID:   rand.Text(),
			Subject:         "0e0d7ac1-2a0f-4f4c-9c0c-6f2f6c7d3b11",
			AuthenticatedAt: now.Add(-time.Minute),
			SharingEndsAt:   now.Add(in.SharingDuration),
		}

		return authserver.Approve(a.consent)
	}
	redirect := b.browse(http.MethodGet, authorisePath(registeredClientID, pushed.body["request_uri"].(string)))
	require.Equal(b.t, http.StatusSeeOther, redirect.status, redirect.body)

	a.code = b.jarmClaims(redirect.location)["code"].(string)

	return a
}

// assertionForm gives a form with only the client assertion parameters.
func (b *bank) assertionForm() url.Values {
	form := b.form(b.assertion())
	form.Del("scope")
	form.Del("grant_type")

	return form
}

func (b *bank) tokenForm(grantType string) url.Values {
	form := b.assertionForm()
	form.Set("grant_type", grantType)

	return form
}

func (b *bank) exchangeForm(a authorisation) url.Values {
	form := b.tokenForm("authorization_code")
	form.Set("code", a.code)
	form.Set("redirect_uri", registeredRedirectURI)
	form.Set("code_verifier", a.verifier)

	return form
}

func (b *bank) exchange(a authorisation) map[string]any {
	b.t.Helper()

	r := b.token(b.exchangeForm(a))
	require.Equal(b.t, http.StatusOK, r.status, r.body)

	return r.body
}

func (b *bank) refresh(refreshToken string) reply {
	b.t.Helper()

	form := b.tokenForm("refresh_token")
	form.Set("refresh_token", refreshToken)

	return b.token(form)
}

func (b *bank) introspect(token string) reply {
	b.t.Helper()

	form := b.assertionForm()
	form.Set("token", token)

	return b.post(b.introspectionURL, form)
}

func (b *bank) claims(token string) map[string]any {
	b.t.Helper()

	parsed, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.PS256})
	require.NoError(b.t, err)

	keys, err := b.signingKeys(b.t.Context())
	require.NoError(b.t, err)

	var claims map[string]any
	require.NoError(b.t, parsed.Claims(&keys[0].Key.PublicKey, &claims))

	return claims
}

func requireTokenError(t *testing.T, r reply, code string) {
	t.Helper()
	require.Equal(t, http.StatusBadRequest, r.status, r.body)
	assert.Equal(t, code, r.body["error"], r.body)
	assert.Equal(t, "no-store", r.header.Get("Cache-Control"))
}

// Verifies: AUTH-0205, AUTH-0207
func TestACodeOfAConsentGivesTheTokensOfTheConsent(t *testing.T) {
	b := newBank(t)
	a := b.authorise(hour)

	tokens := b.exchange(a)

	assert.Equal(t, "Bearer", tokens["token_type"])
	assert.EqualValues(t, 300, tokens["expires_in"])
	assert.Equal(t, "openid bank:accounts.basic:read", tokens["scope"])
	assert.Equal(t, a.consent.ArrangementID, tokens["cdr_arrangement_id"])
	assert.NotEmpty(t, tokens["refresh_token"])

	verified, err := b.dataHolder.VerifyAccessToken(t.Context(), tokens["access_token"].(string), leaf(t, b.cert))
	require.NoError(t, err)
	assert.Equal(t, registeredClientID, verified.ClientID)

	other := clientCertificate(t, "Other App")
	_, err = b.dataHolder.VerifyAccessToken(t.Context(), tokens["access_token"].(string), leaf(t, other))
	require.ErrorIs(t, err, authserver.ErrInvalidToken)
}

// Verifies: AUTH-0215, AUTH-0216, AUTH-0217
func TestTheIDTokenOfAConsentHasNoPersonalInformation(t *testing.T) {
	b := newBank(t)
	a := b.authorise(hour)

	claims := b.claims(b.exchange(a)["id_token"].(string))

	assert.Equal(t, dataHolderIssuer, claims["iss"])
	assert.Equal(t, a.consent.Subject, claims["sub"])
	assert.Equal(t, registeredClientID, claims["aud"])
	assert.Equal(t, "urn:cds.au:cdr:2", claims["acr"])
	assert.Equal(t, "nonce-1", claims["nonce"])
	assert.EqualValues(t, a.consent.AuthenticatedAt.Unix(), claims["auth_time"])
	assert.ElementsMatch(
		t,
		[]string{"iss", "sub", "aud", "exp", "iat", "auth_time", "acr", "nonce"},
		mapKeys(claims),
	)
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

// Verifies: AUTH-0206
func TestARefreshTokenIsIssuedOnlyForASharingDuration(t *testing.T) {
	b := newBank(t)

	assert.NotContains(t, b.exchange(b.authorise(noSharing)), "refresh_token")
	assert.Contains(t, b.exchange(b.authorise(1)), "refresh_token")
}

// Verifies: AUTH-0109
func TestAnAuthorisationCodeIsValidFor60Seconds(t *testing.T) {
	b := newBank(t)
	a := b.authorise(hour)
	grant := b.grants.only(t)

	assert.Equal(t, 60*time.Second, grant.CodeExpiresAt.Sub(grant.CreatedAt))

	grant.CodeExpiresAt = time.Now().Add(-time.Second)
	require.NoError(t, b.grants.SaveGrant(t.Context(), grant))

	requireTokenError(t, b.token(b.exchangeForm(a)), "invalid_grant")
}

// Verifies: AUTH-0202
func TestACodeExchangeThatIsNotValidGivesInvalidGrant(t *testing.T) {
	tests := map[string]func(url.Values){
		"wrong redirect_uri":  func(f url.Values) { f.Set("redirect_uri", "https://recipient.example/other") },
		"wrong code_verifier": func(f url.Values) { f.Set("code_verifier", rand.Text()+rand.Text()) },
		"unknown code":        func(f url.Values) { f.Set("code", rand.Text()) },
	}
	for name, change := range tests {
		t.Run(name, func(t *testing.T) {
			b := newBank(t)
			form := b.exchangeForm(b.authorise(hour))
			change(form)

			requireTokenError(t, b.token(form), "invalid_grant")
		})
	}
}

// Verifies: AUTH-0202, AUTH-0203
func TestEachLaterUseOfACodeGivesInvalidGrantAndEndsTheRefreshToken(t *testing.T) {
	b := newBank(t)
	a := b.authorise(hour)
	tokens := b.exchange(a)

	requireTokenError(t, b.token(b.exchangeForm(a)), "invalid_grant")
	requireTokenError(t, b.token(b.exchangeForm(a)), "invalid_grant")
	requireTokenError(t, b.refresh(tokens["refresh_token"].(string)), "invalid_grant")
}

// Verifies: AUTH-0210, AUTH-0211
func TestARefreshGivesNewTokensForTheSameArrangement(t *testing.T) {
	b := newBank(t)
	a := b.authorise(hour)
	tokens := b.exchange(a)
	refreshToken := tokens["refresh_token"].(string)

	for range 2 {
		r := b.refresh(refreshToken)

		require.Equal(t, http.StatusOK, r.status, r.body)
		assert.NotEqual(t, tokens["access_token"], r.body["access_token"])
		assert.EqualValues(t, 300, r.body["expires_in"])
		assert.Equal(t, a.consent.ArrangementID, r.body["cdr_arrangement_id"])
		assert.Equal(t, a.consent.Subject, b.claims(r.body["id_token"].(string))["sub"])

		if issued, ok := r.body["refresh_token"]; ok {
			assert.Equal(t, refreshToken, issued)
		}
	}
}

// Verifies: AUTH-0212
func TestARefreshTokenThatIsNotValidGivesInvalidGrant(t *testing.T) {
	b := newBank(t)
	refreshToken := b.exchange(b.authorise(hour))["refresh_token"].(string)

	requireTokenError(t, b.refresh(rand.Text()), "invalid_grant")

	grant := b.grants.only(t)
	grant.SharingEndsAt = time.Now().Add(-time.Second)
	require.NoError(t, b.grants.SaveGrant(t.Context(), grant))

	requireTokenError(t, b.refresh(refreshToken), "invalid_grant")
}

// Verifies: AUTH-0213
func TestARefreshAfterTheClientReplacesItsCertificateGivesATokenForTheNewCertificate(t *testing.T) {
	b := newBank(t)
	refreshToken := b.exchange(b.authorise(hour))["refresh_token"].(string)
	b.cert = clientCertificate(t, "Recipient App Web")
	b.http.Transport.(*http.Transport).TLSClientConfig.Certificates = []tls.Certificate{b.cert}
	b.http.CloseIdleConnections()

	r := b.refresh(refreshToken)

	require.Equal(t, http.StatusOK, r.status, r.body)

	_, err := b.dataHolder.VerifyAccessToken(t.Context(), r.body["access_token"].(string), leaf(t, b.cert))
	require.NoError(t, err)
}

func TestTheTokensOfAConsentStayValidAfterARestart(t *testing.T) {
	b := newBank(t)
	a := b.authorise(hour)
	refreshToken := b.exchange(a)["refresh_token"].(string)

	b.restart()

	r := b.refresh(refreshToken)
	require.Equal(t, http.StatusOK, r.status, r.body)
	assert.Equal(t, a.consent.ArrangementID, r.body["cdr_arrangement_id"])
	assert.Equal(t, true, b.introspect(refreshToken).body["active"])
}

func TestAClientCredentialsGrantIsNotStored(t *testing.T) {
	b := newBank(t)

	b.accessToken()

	assert.Empty(t, b.grants.grants)
}

// Verifies: AUTH-0219
func TestIntrospectionOfARefreshTokenGivesTheArrangementAndTheEndOfTheSharing(t *testing.T) {
	b := newBank(t)
	a := b.authorise(hour)

	r := b.introspect(b.exchange(a)["refresh_token"].(string))

	require.Equal(t, http.StatusOK, r.status, r.body)
	assert.Equal(t, map[string]any{
		"active":             true,
		"exp":                float64(a.consent.SharingEndsAt.Unix()),
		"scope":              "openid bank:accounts.basic:read",
		"cdr_arrangement_id": a.consent.ArrangementID,
	}, r.body)
	assert.Equal(t, "no-store", r.header.Get("Cache-Control"))
}

// Verifies: AUTH-0220
func TestIntrospectionOfATokenThatIsNotARefreshTokenGivesOnlyNotActive(t *testing.T) {
	b := newBank(t)
	tokens := b.exchange(b.authorise(hour))

	for _, name := range []string{"access_token", "id_token"} {
		r := b.introspect(tokens[name].(string))

		require.Equal(t, http.StatusOK, r.status, r.body)
		assert.Equal(t, map[string]any{"active": false}, r.body, name)
	}

	assert.Equal(t, map[string]any{"active": false}, b.introspect(b.accessToken()).body)
	assert.Equal(t, map[string]any{"active": false}, b.introspect(rand.Text()).body)
}

// Verifies: AUTH-0032, AUTH-0221
func TestIntrospectionRefusesARequestThatIsNotValid(t *testing.T) {
	b := newBank(t)

	noClient := url.Values{"token": {rand.Text()}}

	requireTokenError(t, b.post(b.introspectionURL, b.assertionForm()), "invalid_request")

	r := b.post(b.introspectionURL, noClient)
	assert.Equal(t, http.StatusUnauthorized, r.status, r.body)
	assert.Equal(t, "invalid_client", r.body["error"])
}
