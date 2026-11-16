package authserver_test

import (
	"context"
	"crypto/rand"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/cds"
)

// arrangementGrants ends an arrangement as the data holder does: it revokes the grant of the arrangement.
type arrangementGrants struct{ grants *grantStore }

func (a arrangementGrants) RevokeArrangement(ctx context.Context, clientID, arrangementID string) error {
	grant, err := a.grants.find(func(g authserver.Grant) bool { return g.ArrangementID == arrangementID })
	if err != nil || grant.ClientID != clientID || !grant.RevokedAt.IsZero() || grant.SharingDuration == 0 {
		return authserver.ErrInvalidArrangement
	}

	grant.RevokedAt = time.Now()

	return a.grants.SaveGrant(ctx, grant)
}

type rawReply struct {
	status int
	header http.Header
	body   string
}

// revokeArrangement sends the form with a new client assertion of the client.
func (b *bank) revokeArrangement(clientID string, form url.Values) rawReply {
	b.t.Helper()

	a := b.assertion()
	a.claims["iss"], a.claims["sub"], a.claims["aud"] = clientID, clientID, b.arrangementURL
	assertionForm := b.form(a)
	assertionForm.Del("scope")
	assertionForm.Del("grant_type")

	maps.Copy(assertionForm, form)

	return b.rawPost(b.arrangementURL, assertionForm)
}

func (b *bank) rawPost(target string, form url.Values) rawReply {
	b.t.Helper()

	req, err := http.NewRequestWithContext(b.t.Context(), http.MethodPost, target, strings.NewReader(form.Encode()))
	require.NoError(b.t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := b.http.Do(req)
	require.NoError(b.t, err)

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	require.NoError(b.t, err)

	return rawReply{status: resp.StatusCode, header: resp.Header, body: string(data)}
}

func requireCDSError(t *testing.T, r rawReply, status int, e cds.Error, detail string) {
	t.Helper()
	require.Equal(t, status, r.status, r.body)
	assert.Equal(t, "application/json", r.header.Get("Content-Type"))
	assert.JSONEq(t, `{"errors": [{"code": "`+e.Code+`", "title": "`+e.Title+`", "detail": "`+detail+`"}]}`, r.body)
}

var (
	invalidArrangement = cds.Error{
		Code:  "urn:au-cds:error:cds-all:Authorisation/InvalidArrangement",
		Title: "Invalid Consent Arrangement",
	}
	fieldMissing = cds.Error{Code: "urn:au-cds:error:cds-all:Field/Missing", Title: "Missing Required Field"}
)

// Verifies: AUTH-0229
func TestARevocationOfAnArrangementOfTheClientGivesNoContentAndEndsEachToken(t *testing.T) {
	b := newBank(t)
	a := b.authorise(hour)
	tokens := b.exchange(a)

	r := b.revokeArrangement(registeredClientID, url.Values{"cdr_arrangement_id": {a.consent.ArrangementID}})

	require.Equal(t, http.StatusNoContent, r.status, r.body)
	assert.Empty(t, r.body)
	assert.False(t, b.consentTokenValid(tokens["access_token"].(string)))
	requireTokenError(t, b.refresh(tokens["refresh_token"].(string)), "invalid_grant")
	assert.Equal(t, map[string]any{"active": false}, b.introspect(tokens["refresh_token"].(string)).body)
}

func TestTheEndpointURLTheIssuerAndTheTokenEndpointAreEachAValidAudience(t *testing.T) {
	b := newBank(t)

	for _, audience := range []string{b.arrangementURL, dataHolderIssuer, b.tokenURL} {
		a := b.authorise(hour)
		assertion := b.assertion()
		assertion.claims["aud"] = audience
		form := b.form(assertion)
		form.Del("scope")
		form.Del("grant_type")
		form.Set("cdr_arrangement_id", a.consent.ArrangementID)

		r := b.rawPost(b.arrangementURL, form)

		assert.Equal(t, http.StatusNoContent, r.status, audience+": "+r.body)
	}
}

// Verifies: AUTH-0230
func TestARevocationThatTheClientCannotMakeGivesInvalidArrangement(t *testing.T) {
	b := newBank(t)
	a := b.authorise(hour)
	tokens := b.exchange(a)
	id := a.consent.ArrangementID

	requireCDSError(t, b.revokeArrangement(secondClientID, url.Values{"cdr_arrangement_id": {id}}),
		http.StatusUnprocessableEntity, invalidArrangement, id)
	assert.True(t, b.consentTokenValid(tokens["access_token"].(string)), "a different client changes nothing")

	unknown := rand.Text()
	requireCDSError(t, b.revokeArrangement(registeredClientID, url.Values{"cdr_arrangement_id": {unknown}}),
		http.StatusUnprocessableEntity, invalidArrangement, unknown)

	require.Equal(t, http.StatusNoContent, b.revokeArrangement(registeredClientID,
		url.Values{"cdr_arrangement_id": {id}}).status)
	requireCDSError(t, b.revokeArrangement(registeredClientID, url.Values{"cdr_arrangement_id": {id}}),
		http.StatusUnprocessableEntity, invalidArrangement, id)
}

// Verifies: AUTH-0228, AUTH-0231
func TestARevocationWithNoArrangementIDGivesFieldMissing(t *testing.T) {
	b := newBank(t)
	b.authorise(hour)

	for name, form := range map[string]url.Values{
		"no cdr_arrangement_id":      {},
		"empty cdr_arrangement_id":   {"cdr_arrangement_id": {""}},
		"only a cdr_arrangement_jwt": {"cdr_arrangement_jwt": {"eyJhbGciOiJQUzI1NiJ9.e30.c2ln"}},
	} {
		requireCDSError(t, b.revokeArrangement(registeredClientID, form), http.StatusBadRequest, fieldMissing,
			"cdr_arrangement_id")
		assert.True(t, b.grants.only(t).RevokedAt.IsZero(), name)
	}
}

// Verifies: AUTH-0032
func TestARevocationWithAClientAuthenticationThatFailsGivesInvalidClient(t *testing.T) {
	b := newBank(t)
	a := b.authorise(hour)
	form := url.Values{"cdr_arrangement_id": {a.consent.ArrangementID}}

	assertion := b.assertion()
	assertion.claims["aud"] = "https://other.example/arrangements/revoke"
	badAudience := b.form(assertion)
	badAudience.Set("cdr_arrangement_id", a.consent.ArrangementID)

	for name, r := range map[string]rawReply{
		"no client assertion": b.rawPost(b.arrangementURL, form),
		"wrong audience":      b.rawPost(b.arrangementURL, badAudience),
	} {
		require.Equal(t, http.StatusUnauthorized, r.status, name+": "+r.body)
		assert.JSONEq(t, `{"error": "invalid_client", "error_description": "invalid client"}`, r.body, name)
	}

	assert.True(t, b.grants.only(t).RevokedAt.IsZero())
}

func TestAClientAssertionOfARevocationIsUsedOneTime(t *testing.T) {
	b := newBank(t)
	first, second := b.authorise(hour), b.authorise(hour)

	assertion := b.assertion()
	assertion.claims["aud"] = b.arrangementURL
	form := b.form(assertion)
	form.Del("scope")
	form.Del("grant_type")

	form.Set("cdr_arrangement_id", first.consent.ArrangementID)
	require.Equal(t, http.StatusNoContent, b.rawPost(b.arrangementURL, form).status)

	form.Set("cdr_arrangement_id", second.consent.ArrangementID)
	assert.Equal(t, http.StatusUnauthorized, b.rawPost(b.arrangementURL, form).status)
}

// Verifies: AUTH-0208
func TestAnAccessTokenExpiresAtTheEndOfTheSharingWhenItIsLessThan5MinutesAway(t *testing.T) {
	b := newBank(t)

	tokens := b.exchange(b.authorise(120))

	require.IsType(t, float64(0), tokens["expires_in"])
	assert.InDelta(t, 120, tokens["expires_in"], 3)
	expiry := b.claims(tokens["access_token"].(string))["exp"].(float64)
	sharingEnds := float64(b.grants.only(t).SharingEndsAt.Unix())
	assert.LessOrEqual(t, expiry, sharingEnds, "the token never expires after the end of the sharing")
	assert.InDelta(t, sharingEnds, expiry, 2)
}

// Verifies: AUTH-0207
func TestAnAccessTokenWithoutASharingDurationHas5Minutes(t *testing.T) {
	b := newBank(t)

	assert.EqualValues(t, 300, b.exchange(b.authorise(noSharing))["expires_in"])
	assert.EqualValues(t, 300, b.token(b.form(b.assertion())).body["expires_in"])
}
