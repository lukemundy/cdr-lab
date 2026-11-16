package authserver_test

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
)

var approved = authserver.Consent{ArrangementID: "arrangement-1", Subject: "subject-1"}

// pages is a ConsumerPages that records each call and answers with next.
type pages struct {
	interactions []authserver.Interaction
	refusals     []int
	// next gives the outcome of each interaction. With no next, an interaction continues with the state "page <n>".
	next func(authserver.Interaction) authserver.Outcome
}

func (p *pages) Interact(w http.ResponseWriter, _ *http.Request, in authserver.Interaction) authserver.Outcome {
	p.interactions = append(p.interactions, in)

	if p.next != nil {
		return p.next(in)
	}

	state := "page " + string(rune('0'+len(p.interactions)))
	_, _ = io.WriteString(w, state)

	return authserver.Continue(state)
}

func (p *pages) Refused(w http.ResponseWriter, _ *http.Request, status int) {
	p.refusals = append(p.refusals, status)
	w.WriteHeader(status)
	_, _ = io.WriteString(w, "refused")
}

type pageReply struct {
	status   int
	body     string
	location string
}

func (b *bank) browse(method, target string) pageReply {
	b.t.Helper()

	req, err := http.NewRequestWithContext(b.t.Context(), method, b.publicURL+target, nil)
	require.NoError(b.t, err)

	resp, err := b.browser.Do(req)
	require.NoError(b.t, err)

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(b.t, err)

	return pageReply{status: resp.StatusCode, body: string(body), location: resp.Header.Get("Location")}
}

// requestURI pushes a valid request object and gives its request_uri.
func (b *bank) requestURI() string {
	b.t.Helper()

	r := b.push(b.requestObject())
	require.Equal(b.t, http.StatusCreated, r.status, r.body)

	return r.body["request_uri"].(string)
}

func authorisePath(clientID, requestURI string) string {
	return "/authorize?" + url.Values{"client_id": {clientID}, "request_uri": {requestURI}}.Encode()
}

// Verifies: AUTH-0096
func TestTheAuthorisationEndpointStartsTheConsumerPagesForAPushedRequest(t *testing.T) {
	b := newBank(t)

	reply := b.browse(http.MethodGet, authorisePath(registeredClientID, b.requestURI()))

	require.Equal(t, http.StatusOK, reply.status, reply.body)
	assert.Equal(t, "page 1", reply.body)
	require.Len(t, b.pages.interactions, 1)

	first := b.pages.interactions[0]
	assert.Equal(t, registeredClientID, first.ClientID)
	assert.Equal(t, []string{"openid", "bank:accounts.basic:read"}, first.Scopes)
	assert.Nil(t, first.State)
	assert.Regexp(t, `^/authorize/[0-9a-f-]{36}$`, first.Path)
}

func TestTheNextRequestOfAFlowGetsTheStateThatTheLastOneKept(t *testing.T) {
	b := newBank(t)
	b.browse(http.MethodGet, authorisePath(registeredClientID, b.requestURI()))
	path := b.pages.interactions[0].Path

	for _, method := range []string{http.MethodPost, http.MethodGet} {
		reply := b.browse(method, path)
		require.Equal(t, http.StatusOK, reply.status, reply.body)
	}

	require.Len(t, b.pages.interactions, 3)
	assert.Equal(t, "page 1", b.pages.interactions[1].State)
	assert.Equal(t, "page 2", b.pages.interactions[2].State)
	assert.Equal(t, path, b.pages.interactions[2].Path)
}

func TestTwoFlowsHaveDifferentPathsAndStates(t *testing.T) {
	b := newBank(t)
	b.browse(http.MethodGet, authorisePath(registeredClientID, b.requestURI()))
	b.browse(http.MethodGet, authorisePath(registeredClientID, b.requestURI()))

	first, second := b.pages.interactions[0], b.pages.interactions[1]

	assert.NotEqual(t, first.Path, second.Path)
	assert.Nil(t, second.State)
}

// Verifies: AUTH-0098, AUTH-0100
func TestARequestURIIsAcceptedOneTimeOnly(t *testing.T) {
	b := newBank(t)
	target := authorisePath(registeredClientID, b.requestURI())

	first := b.browse(http.MethodGet, target)
	second := b.browse(http.MethodGet, target)

	require.Equal(t, http.StatusOK, first.status, first.body)
	assert.Equal(t, http.StatusBadRequest, second.status)
	assert.Equal(t, "refused", second.body)
	assert.Empty(t, second.location)
	assert.Len(t, b.pages.interactions, 1, "the second use does not start the pages")
}

// Verifies: AUTH-0100, AUTH-0101
func TestTheAuthorisationEndpointRefusesARequestThatItCannotTrustWithNoRedirect(t *testing.T) {
	const otherClientID = "client-2"

	tests := map[string]struct {
		target func(b *bank) string
		status int
	}{
		// Verifies: AUTH-0098
		"request_uri that the data holder does not know": {
			target: func(*bank) string { return authorisePath(registeredClientID, requestURIPrefix+"unknown") },
			status: http.StatusBadRequest,
		},
		// Verifies: AUTH-0097
		"no request_uri": {
			target: func(*bank) string { return "/authorize?client_id=" + registeredClientID },
			status: http.StatusBadRequest,
		},
		// Verifies: AUTH-0097
		"request parameter": {
			target: func(b *bank) string {
				object := b.form(b.signedObject(b.requestObject())).Get("client_assertion")
				return "/authorize?" + url.Values{"client_id": {registeredClientID}, "request": {object}}.Encode()
			},
			status: http.StatusBadRequest,
		},
		// Verifies: AUTH-0099
		"request_uri of a different client": {
			target: func(b *bank) string { return authorisePath(otherClientID, b.requestURI()) },
			status: http.StatusUnauthorized,
		},
		"client that is not registered": {
			target: func(b *bank) string { return authorisePath("client-3", b.requestURI()) },
			status: http.StatusUnauthorized,
		},
		"no client_id": {
			target: func(b *bank) string { return "/authorize?request_uri=" + url.QueryEscape(b.requestURI()) },
			status: http.StatusUnauthorized,
		},
		// Verifies: AUTH-0102
		"flow that does not exist": {
			target: func(*bank) string { return "/authorize/5a1bf696-ee03-408b-b315-97955415d1f0" },
			status: http.StatusBadRequest,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := newBank(t)
			other := b.clients[registeredClientID]
			other.ID = otherClientID
			b.clients[otherClientID] = other

			reply := b.browse(http.MethodGet, tt.target(b))

			assert.Equal(t, tt.status, reply.status, reply.body)
			assert.Equal(t, "refused", reply.body)
			assert.Empty(t, reply.location)
			assert.Equal(t, []int{tt.status}, b.pages.refusals)
			assert.Empty(t, b.pages.interactions)
		})
	}
}

// jarmClaims verifies the response JWT of a redirect with the key of the data holder, and gives its claims.
func (b *bank) jarmClaims(location string) map[string]any {
	b.t.Helper()

	redirect, err := url.Parse(location)
	require.NoError(b.t, err)
	assert.True(b.t, strings.HasPrefix(location, registeredRedirectURI+"?"), location)
	assert.Equal(b.t, []string{"response"}, queryNames(redirect.Query()))

	parsed, err := jwt.ParseSigned(redirect.Query().Get("response"), []jose.SignatureAlgorithm{jose.PS256})
	require.NoError(b.t, err)

	keys, err := b.signingKeys(b.t.Context())
	require.NoError(b.t, err)

	var claims map[string]any
	require.NoError(b.t, parsed.Claims(&keys[0].Key.PublicKey, &claims))

	return claims
}

func queryNames(query url.Values) []string {
	names := make([]string, 0, len(query))
	for name := range query {
		names = append(names, name)
	}

	return names
}

// Verifies: AUTH-0104, AUTH-0105, AUTH-0106, AUTH-0107
func TestADenialRedirectsTheBrowserWithASignedErrorForTheSoftwareProduct(t *testing.T) {
	for _, denial := range []authserver.Denial{authserver.AccessDenied, authserver.TemporarilyUnavailable} {
		t.Run(string(denial), func(t *testing.T) {
			b := newBank(t)
			b.browse(http.MethodGet, authorisePath(registeredClientID, b.requestURI()))
			b.pages.next = func(authserver.Interaction) authserver.Outcome { return authserver.Deny(denial) }

			reply := b.browse(http.MethodPost, b.pages.interactions[0].Path)

			require.Equal(t, http.StatusSeeOther, reply.status, reply.body)

			claims := b.jarmClaims(reply.location)
			assert.Equal(t, dataHolderIssuer, claims["iss"])
			assert.Equal(t, registeredClientID, claims["aud"])
			assert.Equal(t, string(denial), claims["error"])
			assert.Equal(t, "state-1", claims["state"])
			assert.NotContains(t, claims, "code")
			assert.InDelta(t, float64(time.Now().Add(10*time.Minute).Unix()), claims["exp"], 30)
		})
	}
}

// Verifies: AUTH-0102
func TestAFlowThatEndedWithADenialDoesNotContinue(t *testing.T) {
	b := newBank(t)
	b.browse(http.MethodGet, authorisePath(registeredClientID, b.requestURI()))
	path := b.pages.interactions[0].Path
	b.pages.next = func(authserver.Interaction) authserver.Outcome { return authserver.Deny(authserver.AccessDenied) }
	b.browse(http.MethodPost, path)
	b.pages.next = nil

	reply := b.browse(http.MethodPost, path)

	assert.Equal(t, http.StatusBadRequest, reply.status)
	assert.Empty(t, reply.location)
	assert.Len(t, b.pages.interactions, 2, "the pages are not called for a flow that ended")
}

// Verifies: AUTH-0090
func TestAnInteractionHasTheSharingDurationOfTheRequest(t *testing.T) {
	tests := map[string]struct {
		claims map[string]any
		want   time.Duration
	}{
		"no sharing_duration": {claims: nil, want: 0},
		"90 days":             {claims: map[string]any{"sharing_duration": 7_776_000}, want: 90 * 24 * time.Hour},
		"more than one year":  {claims: map[string]any{"sharing_duration": 99_999_999}, want: 365 * 24 * time.Hour},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := newBank(t)
			object := b.requestObject()

			if tt.claims != nil {
				object["claims"] = tt.claims
			}

			pushed := b.push(object)
			require.Equal(t, http.StatusCreated, pushed.status, pushed.body)

			b.browse(http.MethodGet, authorisePath(registeredClientID, pushed.body["request_uri"].(string)))

			require.Len(t, b.pages.interactions, 1)
			assert.Equal(t, tt.want, b.pages.interactions[0].SharingDuration)
		})
	}
}

// Verifies: AUTH-0104, AUTH-0105, AUTH-0106, AUTH-0107
func TestAnApprovalRedirectsTheBrowserWithASignedCodeForTheSoftwareProduct(t *testing.T) {
	b := newBank(t)
	b.browse(http.MethodGet, authorisePath(registeredClientID, b.requestURI()))
	b.pages.next = func(authserver.Interaction) authserver.Outcome { return authserver.Approve(approved) }

	reply := b.browse(http.MethodPost, b.pages.interactions[0].Path)

	require.Equal(t, http.StatusSeeOther, reply.status, reply.body)

	claims := b.jarmClaims(reply.location)
	assert.Equal(t, dataHolderIssuer, claims["iss"])
	assert.Equal(t, registeredClientID, claims["aud"])
	assert.Equal(t, "state-1", claims["state"])
	assert.NotEmpty(t, claims["code"])
	assert.NotContains(t, claims, "error")
	assert.InDelta(t, float64(time.Now().Add(10*time.Minute).Unix()), claims["exp"], 30)
}

func TestTwoApprovalsGiveDifferentCodes(t *testing.T) {
	b := newBank(t)
	b.pages.next = func(in authserver.Interaction) authserver.Outcome {
		if in.State == nil {
			return authserver.Continue("consent")
		}

		return authserver.Approve(approved)
	}

	codes := make([]any, 0, 2)

	for range 2 {
		b.browse(http.MethodGet, authorisePath(registeredClientID, b.requestURI()))
		path := b.pages.interactions[len(b.pages.interactions)-1].Path
		codes = append(codes, b.jarmClaims(b.browse(http.MethodPost, path).location)["code"])
	}

	assert.NotEqual(t, codes[0], codes[1])
}

// Verifies: AUTH-0102
func TestAFlowThatEndedWithAnApprovalDoesNotContinue(t *testing.T) {
	b := newBank(t)
	b.browse(http.MethodGet, authorisePath(registeredClientID, b.requestURI()))
	path := b.pages.interactions[0].Path
	b.pages.next = func(authserver.Interaction) authserver.Outcome { return authserver.Approve(approved) }
	b.browse(http.MethodPost, path)
	b.pages.next = nil

	reply := b.browse(http.MethodPost, path)

	assert.Equal(t, http.StatusBadRequest, reply.status)
	assert.Empty(t, reply.location)
	assert.Len(t, b.pages.interactions, 2, "the pages are not called for a flow that ended")
}

func TestAnOutcomeGivesItsKind(t *testing.T) {
	state, continues := authserver.Continue("page").Continues()
	assert.True(t, continues)
	assert.Equal(t, "page", state)

	consent, ok := authserver.Approve(approved).Approved()
	assert.True(t, ok)
	assert.Equal(t, approved, consent)

	_, continues = authserver.Approve(approved).Continues()
	assert.False(t, continues)

	_, continues = authserver.Deny(authserver.AccessDenied).Continues()
	assert.False(t, continues)

	_, ok = authserver.Deny(authserver.AccessDenied).Approved()
	assert.False(t, ok)
}
