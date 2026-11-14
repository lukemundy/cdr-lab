//go:build conformance

package dataholder

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

// The brand name of the Registration Test software products in services/register/seed/example.json.
const recipientBrandName = "Registration Test"

// consumer is the browser of one consumer. It keeps cookies, and it does not follow a redirect, so that a test can
// read the redirect to the software product.
type consumer struct {
	t      *testing.T
	client *http.Client
}

func newConsumer(t *testing.T) *consumer {
	t.Helper()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	return &consumer{t: t, client: &http.Client{
		Timeout: labtest.RequestTimeout,
		Jar:     jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: labtest.PublicRoot(t), MinVersion: tls.VersionTLS12},
		},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

// webPage is a response to the browser of a consumer.
type webPage struct {
	status int
	header http.Header
	body   string
	// url is the URL of the request.
	url *url.URL
}

func (c *consumer) do(req *http.Request) webPage {
	c.t.Helper()

	resp, err := c.client.Do(req)
	require.NoError(c.t, err)

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(c.t, err)

	return webPage{status: resp.StatusCode, header: resp.Header, body: string(body), url: req.URL}
}

func (c *consumer) open(target string) webPage {
	c.t.Helper()

	req, err := http.NewRequestWithContext(c.t.Context(), http.MethodGet, target, nil)
	require.NoError(c.t, err)

	return c.do(req)
}

var (
	formAction = regexp.MustCompile(`<form method="post" action="([^"]+)">`)
	formToken  = regexp.MustCompile(`<input type="hidden" name="csrf_token" value="([^"]+)">`)
)

// form gives the URL and the token of the forms of a page, which all have the same two values.
func (p webPage) form(t *testing.T) (action *url.URL, token string) {
	t.Helper()

	actionMatch := formAction.FindStringSubmatch(p.body)
	require.Len(t, actionMatch, 2, "the page has a form: %s", p.body)
	tokenMatch := formToken.FindStringSubmatch(p.body)
	require.Len(t, tokenMatch, 2, "the form has a token: %s", p.body)

	action, err := p.url.Parse(actionMatch[1])
	require.NoError(t, err)

	return action, tokenMatch[1]
}

// submit sends a form of the page with the fields, as the button with the action does.
func (c *consumer) submit(from webPage, action string, fields url.Values) webPage {
	c.t.Helper()

	target, token := from.form(c.t)

	form := url.Values{"csrf_token": {token}, "action": {action}}
	maps.Copy(form, fields)

	return c.post(target, form)
}

func (c *consumer) post(target *url.URL, form url.Values) webPage {
	c.t.Helper()

	req, err := http.NewRequestWithContext(
		c.t.Context(), http.MethodPost, target.String(), strings.NewReader(form.Encode()),
	)
	require.NoError(c.t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.do(req)
}

func authorisationEndpoint(t *testing.T, client registeredClient) string {
	t.Helper()

	endpoint := openIDConfiguration(t, client.issuer)["authorization_endpoint"]
	require.IsType(t, "", endpoint)

	return endpoint.(string)
}

// authorisationURL pushes a valid request object, and gives the URL that the software product sends the browser to.
func (c registeredClient) authorisationURL(t *testing.T, claims map[string]any) string {
	t.Helper()

	pushed := c.push(t, claims)
	require.Equal(t, http.StatusCreated, pushed.Status, pushed.Body)

	query := url.Values{"client_id": {c.clientID}, "request_uri": {pushed.Body["request_uri"].(string)}}

	return authorisationEndpoint(t, c) + "?" + query.Encode()
}

// startFlow gives a consumer at the first page of a new flow.
func (c registeredClient) startFlow(t *testing.T) (*consumer, webPage) {
	t.Helper()

	browser := newConsumer(t)
	first := browser.open(c.authorisationURL(t, c.requestObjectClaims()))
	require.Equal(t, http.StatusOK, first.status, first.body)

	return browser, first
}

// requireRefusedPage checks a page for a request that the data holder refuses with no redirect.
func requireRefusedPage(t *testing.T, p webPage, status int) {
	t.Helper()
	require.Equal(t, status, p.status, p.body)
	assert.Contains(t, p.header.Get("Content-Type"), "text/html")
	assert.Empty(t, p.header.Get("Location"), "the data holder does not redirect the browser")
	assert.NotContains(t, p.body, "<form", "the page has no form")
}

// responseClaims verifies the response JWT of a redirect to the software product with the keys of the data holder.
func (c registeredClient) responseClaims(t *testing.T, redirect webPage) map[string]any {
	t.Helper()
	require.Equal(t, http.StatusSeeOther, redirect.status, redirect.body)

	location, err := url.Parse(redirect.header.Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, registeredRedirectURI, location.Scheme+"://"+location.Host+location.Path)
	require.Len(t, location.Query(), 1, "the redirect has only the response parameter")

	return c.signedClaims(t, location.Query().Get("response"))
}

// signedClaims verifies a PS256 JWT of the data holder with the keys at its jwks_uri, and gives the claims.
func (c registeredClient) signedClaims(t *testing.T, token string) map[string]any {
	t.Helper()

	parsed, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.PS256})
	require.NoError(t, err)

	jwksURI := openIDConfiguration(t, c.issuer)["jwks_uri"]
	require.IsType(t, "", jwksURI)

	var keys jose.JSONWebKeySet
	require.NoError(t, json.Unmarshal(get(t, jwksURI.(string), http.Header{}).body, &keys))

	found := keys.Key(parsed.Headers[0].KeyID)
	require.Len(t, found, 1, "the JWKS of the data holder has the key of the JWT")

	var claims map[string]any
	require.NoError(t, parsed.Claims(found[0].Key, &claims))

	return claims
}

func (c registeredClient) requireAccessDenied(t *testing.T, redirect webPage, state any) {
	t.Helper()

	claims := c.responseClaims(t, redirect)
	assert.Equal(t, c.issuer, claims["iss"])
	assert.Equal(t, c.clientID, claims["aud"])
	assert.Equal(t, "access_denied", claims["error"])
	assert.Equal(t, state, claims["state"])
	assert.NotContains(t, claims, "code")

	require.IsType(t, float64(0), claims["exp"])
	expires := time.Unix(int64(claims["exp"].(float64)), 0)
	assert.WithinRange(t, expires, time.Now(), time.Now().Add(10*time.Minute+time.Minute))
}

func TestTheAuthorisationEndpointGivesAPageThatAsksOnlyForTheCustomerNumber(t *testing.T) {
	client := registered(t, registrationTestProduct)

	_, first := client.startFlow(t)

	assert.Contains(t, first.header.Get("Content-Type"), "text/html")
	assert.Contains(t, first.body, `name="identifier"`)
	assert.Equal(t, 1, strings.Count(first.body, `<input type="text"`), "the page asks for one value")
	assert.NotContains(t, first.body, `type="password"`)
	assert.Contains(t, first.body, recipientBrandName)
	assert.Contains(t, text(first), "asks for your data from "+bankName(t)+".")
	assert.Contains(t, first.body, "does not need your password")
	assert.NotContains(t, first.body, "<a ", "the page has no link, also not for forgotten details")
	assert.NotContains(t, first.body, "<script", "the page works with no JavaScript")
	assert.NotContains(t, first.body, "http", "the page loads no content from other hosts")
}

func TestThePagesOfAFlowAreNotCachedAndDoNotPermitCrossOriginRequests(t *testing.T) {
	client := registered(t, registrationTestProduct)
	browser := newConsumer(t)

	req, err := http.NewRequestWithContext(
		t.Context(), http.MethodGet, client.authorisationURL(t, client.requestObjectClaims()), nil,
	)
	require.NoError(t, err)
	req.Header.Set("Origin", "https://app.example")

	first := browser.do(req)

	require.Equal(t, http.StatusOK, first.status, first.body)
	assert.Empty(t, first.header.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, first.header.Get("Cache-Control"), "no-store")
	assert.Equal(t, "no-referrer", first.header.Get("Referrer-Policy"))
	assert.Contains(t, first.header.Get("Content-Security-Policy"), "default-src 'none'")
}

func TestARequestURIStartsOneFlowOnly(t *testing.T) {
	client := registered(t, registrationTestProduct)
	target := client.authorisationURL(t, client.requestObjectClaims())

	first := newConsumer(t).open(target)
	second := newConsumer(t).open(target)

	require.Equal(t, http.StatusOK, first.status, first.body)
	requireRefusedPage(t, second, http.StatusBadRequest)
}

func TestTheAuthorisationEndpointRefusesARequestThatItCannotTrustWithNoRedirect(t *testing.T) {
	client := registered(t, registrationTestProduct)
	other := registered(t, minimalProduct)
	endpoint := authorisationEndpoint(t, client)

	tests := map[string]struct {
		query  func() url.Values
		status int
	}{
		"request_uri that the data holder does not know": {
			query: func() url.Values {
				return url.Values{"client_id": {client.clientID}, "request_uri": {requestURIPrefix + "unknown"}}
			},
			status: http.StatusBadRequest,
		},
		"request parameter": {
			query: func() url.Values {
				object := client.signRequestObject(t, client.requestObjectClaims())
				return url.Values{"client_id": {client.clientID}, "request": {object}}
			},
			status: http.StatusBadRequest,
		},
		"request_uri of a different software product": {
			query: func() url.Values {
				pushed := client.push(t, client.requestObjectClaims())
				require.Equal(t, http.StatusCreated, pushed.Status, pushed.Body)

				return url.Values{
					"client_id":   {other.clientID},
					"request_uri": {pushed.Body["request_uri"].(string)},
				}
			},
			status: http.StatusUnauthorized,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			refused := newConsumer(t).open(endpoint + "?" + tt.query().Encode())

			requireRefusedPage(t, refused, tt.status)
		})
	}
}

func TestACancelOnEachPageGivesTheSoftwareProductASignedAccessDeniedResponse(t *testing.T) {
	client := registered(t, registrationTestProduct)

	t.Run("customer number page", func(t *testing.T) {
		claims := client.requestObjectClaims()
		browser := newConsumer(t)
		first := browser.open(client.authorisationURL(t, claims))

		cancelled := browser.submit(first, "cancel", nil)

		client.requireAccessDenied(t, cancelled, claims["state"])
	})

	t.Run("page after the customer number", func(t *testing.T) {
		claims := client.requestObjectClaims()
		browser := newConsumer(t)
		first := browser.open(client.authorisationURL(t, claims))
		second := browser.submit(first, "continue", url.Values{"identifier": {"00000000"}})
		require.Equal(t, http.StatusOK, second.status, second.body)

		cancelled := browser.submit(second, "cancel", nil)

		client.requireAccessDenied(t, cancelled, claims["state"])
	})
}

func TestAFlowThatEndedDoesNotContinue(t *testing.T) {
	client := registered(t, registrationTestProduct)
	browser, first := client.startFlow(t)
	browser.submit(first, "cancel", nil)

	again := browser.submit(first, "cancel", nil)

	requireRefusedPage(t, again, http.StatusBadRequest)
}

func TestAFormWithoutItsProtectionAgainstCrossSiteRequestForgeryIsRefused(t *testing.T) {
	client := registered(t, registrationTestProduct)

	t.Run("no form token", func(t *testing.T) {
		browser, first := client.startFlow(t)
		target, _ := first.form(t)

		refused := browser.post(target, url.Values{"action": {"cancel"}})

		assert.Equal(t, http.StatusForbidden, refused.status, refused.body)
		assert.Empty(t, refused.header.Get("Location"))
	})

	t.Run("form from a different browser", func(t *testing.T) {
		_, first := client.startFlow(t)
		target, token := first.form(t)

		refused := newConsumer(t).post(target, url.Values{"csrf_token": {token}, "action": {"cancel"}})

		assert.Equal(t, http.StatusForbidden, refused.status, refused.body)
		assert.Empty(t, refused.header.Get("Location"))
	})

	t.Run("the flow continues for the browser that started it", func(t *testing.T) {
		browser, first := client.startFlow(t)
		target, _ := first.form(t)
		browser.post(target, url.Values{"action": {"cancel"}})

		cancelled := browser.submit(first, "cancel", nil)

		assert.Equal(t, http.StatusSeeOther, cancelled.status, cancelled.body)
	})
}
