//go:build conformance

package dataholder

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

// All customer data in this file is invented.

const (
	customerPath  = "/cds-au/v1/common/customer"
	customerScope = "openid profile common:customer.basic:read"
)

func customerURL(t *testing.T) string {
	t.Helper()

	return secureBaseURI(t) + customerPath
}

func userInfoURL(t *testing.T, client registeredClient) string {
	t.Helper()

	endpoint := openIDConfiguration(t, client.issuer)["userinfo_endpoint"]
	require.IsType(t, "", endpoint)

	return endpoint.(string)
}

// resourceHeader has each header that a request on the resource base URI needs.
func resourceHeader(token string) http.Header {
	return http.Header{
		"Authorization":    {"Bearer " + token},
		"X-V":              {"1"},
		"X-Fapi-Auth-Date": {time.Now().UTC().Format(http.TimeFormat)},
	}
}

// secureRequest sends a request to the mTLS listener with the client certificate cert.
func secureRequest(
	t *testing.T,
	cert tls.Certificate,
	method, target string,
	header http.Header,
	body io.Reader,
) response {
	t.Helper()

	client := &http.Client{Timeout: labtest.RequestTimeout, Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      labtest.CDRCABundle(t),
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		},
	}}

	req, err := http.NewRequestWithContext(t.Context(), method, target, body)
	require.NoError(t, err)

	req.Header = header

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return response{Response: resp, body: data}
}

// getCustomer validates each response with a status that the pinned CDS common OpenAPI specification documents.
func getCustomer(t *testing.T, cert tls.Certificate, header http.Header) response {
	t.Helper()

	resp := secureRequest(t, cert, http.MethodGet, customerURL(t), header, nil)

	spec, err := LoadCommonSpec()
	require.NoError(t, err)

	resp.Body = io.NopCloser(bytes.NewReader(resp.body))
	require.NoError(t, spec.ValidateDocumentedResponse(resp.Response))

	return resp
}

// userInfo sends a POST with no body, because the token is in the Authorization header.
func userInfo(t *testing.T, client registeredClient, cert tls.Certificate, method, token string) response {
	t.Helper()

	header := http.Header{}
	if token != "" {
		header.Set("Authorization", "Bearer "+token)
	}

	return secureRequest(t, cert, method, userInfoURL(t, client), header, nil)
}

func decodeJSON(t *testing.T, resp response) map[string]any {
	t.Helper()
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(resp.body, &body), string(resp.body))

	return body
}

// errorTitles has the title of the CDS error list for each code of these tests.
var errorTitles = map[string]string{
	"urn:au-cds:error:cds-all:Authorisation/InvalidArrangement":        "Invalid Consent Arrangement",
	"urn:au-cds:error:cds-all:Field/Missing":                           "Missing Required Field",
	"urn:au-cds:error:cds-all:Authorisation/InvalidConsent":            "Consent Is Invalid",
	"urn:au-cds:error:cds-all:Header/Missing":                          "Missing Required Header",
	"urn:au-cds:error:cds-all:Header/Invalid":                          "Invalid Header",
	"urn:au-cds:error:cds-all:Header/InvalidVersion":                   "Invalid Version",
	"urn:au-cds:error:cds-all:Header/UnsupportedVersion":               "Unsupported Version",
	"urn:au-cds:error:cds-all:Resource/Invalid":                        "Invalid Resource",
	"urn:au-cds:error:cds-all:Resource/NotImplemented":                 "Resource Not Implemented",
	"urn:au-cds:error:cds-banking:Authorisation/InvalidBankingAccount": "Invalid Banking Account",
}

func requireResourceError(t *testing.T, resp response, status int, code, detail string) {
	t.Helper()

	assert.Equal(t, detail, requireCDSError(t, resp, status, code, errorTitles[code]))
}

func requireRefusedToken(t *testing.T, resp response, wwwAuthenticate string) {
	t.Helper()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode, string(resp.body))
	assert.Equal(t, wwwAuthenticate, resp.Header.Get("WWW-Authenticate"))
	assert.Empty(t, resp.body)
}

// newAccountWithoutFirstName is a customer with a last name only.
func newAccountWithoutFirstName(t *testing.T) account {
	t.Helper()

	reply := operate(t, http.MethodPost, "/v1/customers", customerRequest{
		LastName:     "Sample",
		EmailAddress: strings.ToLower(rand.Text()) + "@mail.test",
	})
	require.Equal(t, http.StatusCreated, reply.status, string(reply.body))

	var created operatorCustomer
	reply.decode(t, &created)

	customer := account{t: t, customer: created, browser: newConsumer(t)}
	customer.addAccount("Everyday", newAccountNumber())

	return customer
}

func TestASoftwareProductReadsTheCustomerWhoGaveConsent(t *testing.T) {
	client := registered(t, allScopesProduct)
	created := time.Now().Truncate(time.Second)
	customer := newEligibleAccount(t)
	consent := customer.consents(client, customerScope, ninetyDays)
	tokens := consent.tokens(t)
	accessToken := tokens["access_token"].(string)
	subject := client.subjectOf(t, tokens)

	var lastUpdate time.Time

	t.Run("Get Customer gives the customer as a person", func(t *testing.T) {
		header := resourceHeader(accessToken)
		interactionID := uuid.NewString()
		header.Set("x-fapi-interaction-id", interactionID)
		header.Set("x-fapi-customer-ip-address", "2001:db8::7")
		header.Set("x-cds-client-headers", "not a value that the data holder reads")

		resp := getCustomer(t, client.cert, header)

		require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
		assert.Equal(t, "1", resp.Header.Get("x-v"))
		assert.Equal(t, interactionID, resp.Header.Get("x-fapi-interaction-id"))
		_, err := http.ParseTime(resp.Header.Get("Date"))
		require.NoError(t, err, "the response has a Date header")

		var body envelope[struct {
			CustomerUType string `json:"customerUType"`
			Person        struct {
				LastUpdateTime time.Time `json:"lastUpdateTime"`
				FirstName      string    `json:"firstName"`
				LastName       string    `json:"lastName"`
				MiddleNames    []string  `json:"middleNames"`
			} `json:"person"`
		}]
		require.NoError(t, json.Unmarshal(resp.body, &body))

		assert.Equal(t, customerURL(t), body.Links.Self)
		assert.Equal(t, "person", body.Data.CustomerUType)
		assert.Equal(t, customer.customer.FirstName, body.Data.Person.FirstName)
		assert.Equal(t, customer.customer.LastName, body.Data.Person.LastName)
		assert.Equal(t, []string{}, body.Data.Person.MiddleNames)
		assert.WithinRange(t, body.Data.Person.LastUpdateTime, created, time.Now())
		assert.NotContains(t, string(resp.body), "prefix")
		assert.NotContains(t, string(resp.body), "suffix")
		assert.NotContains(t, string(resp.body), "occupation")

		lastUpdate = body.Data.Person.LastUpdateTime
	})

	t.Run("UserInfo gives the sub of the ID token and the name of the customer", func(t *testing.T) {
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			resp := userInfo(t, client, client.cert, method, accessToken)

			require.Equal(t, http.StatusOK, resp.StatusCode, method, string(resp.body))
			claims := decodeJSON(t, resp)
			assert.Equal(t, []string{"family_name", "given_name", "name", "sub", "updated_at"}, keysOf(claims), method)
			assert.Equal(t, subject, claims["sub"], method)
			assert.Equal(t, "Alex Example", claims["name"], method)
			assert.Equal(t, customer.customer.FirstName, claims["given_name"], method)
			assert.Equal(t, customer.customer.LastName, claims["family_name"], method)
			assert.EqualValues(t, lastUpdate.Unix(), claims["updated_at"], method)
		}
	})

	t.Run("an access token from a refresh gives the same data", func(t *testing.T) {
		refreshed := client.refresh(t, tokens["refresh_token"].(string))
		require.Equal(t, http.StatusOK, refreshed.Status, refreshed.Body)
		token := refreshed.Body["access_token"].(string)

		assert.Equal(t, http.StatusOK, getCustomer(t, client.cert, resourceHeader(token)).StatusCode)

		claims := decodeJSON(t, userInfo(t, client, client.cert, http.MethodGet, token))
		assert.Equal(t, subject, claims["sub"])
	})

	t.Run("a token of a grant that a second use of the code revoked is refused", func(t *testing.T) {
		requireTokenError(t, consent.exchange(t), http.StatusBadRequest, "invalid_grant")

		const invalid = `Bearer error="invalid_token"`

		requireRefusedToken(t, getCustomer(t, client.cert, resourceHeader(accessToken)), invalid)
		requireRefusedToken(t, userInfo(t, client, client.cert, http.MethodGet, accessToken), invalid)
	})
}

func TestGetCustomerAndUserInfoOmitTheFirstNameOfACustomerWhoHasNone(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newAccountWithoutFirstName(t)
	accessToken := customer.consents(client, customerScope, ninetyDays).tokens(t)["access_token"].(string)

	resp := getCustomer(t, client.cert, resourceHeader(accessToken))
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
	assert.NotContains(t, string(resp.body), "firstName")

	claims := decodeJSON(t, userInfo(t, client, client.cert, http.MethodGet, accessToken))
	assert.Equal(t, []string{"family_name", "name", "sub", "updated_at"}, keysOf(claims))
	assert.Equal(t, "Sample", claims["name"])
}

func TestUserInfoGivesOnlyTheSubWithoutTheScopeProfile(t *testing.T) {
	client := registered(t, allScopesProduct)
	tokens := newEligibleAccount(t).consents(client, "openid common:customer.basic:read", ninetyDays).tokens(t)

	resp := userInfo(t, client, client.cert, http.MethodGet, tokens["access_token"].(string))

	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
	assert.Equal(t, map[string]any{"sub": client.subjectOf(t, tokens)}, decodeJSON(t, resp))
}

func TestARequestWithoutATokenThatGivesAccessIsRefusedWith401AndNoBody(t *testing.T) {
	client := registered(t, allScopesProduct)
	accessToken := newEligibleAccount(t).consents(client, customerScope, ninetyDays).tokens(t)["access_token"].(string)
	// The CDR CA issued it, so the TLS handshake succeeds, but the token is not bound to it.
	other := labtest.ClientCertificate(t, "PKI_CLIENT_CERT", "PKI_CLIENT_KEY")

	const invalid = `Bearer error="invalid_token"`

	t.Run("no Authorization header", func(t *testing.T) {
		header := resourceHeader(accessToken)
		header.Del("Authorization")

		requireRefusedToken(t, getCustomer(t, client.cert, header), "Bearer")
		requireRefusedToken(t, userInfo(t, client, client.cert, http.MethodGet, ""), "Bearer")
	})

	t.Run("the token in the query or in the body only", func(t *testing.T) {
		header := resourceHeader(accessToken)
		header.Del("Authorization")

		query := "?" + url.Values{"access_token": {accessToken}}.Encode()

		resp := secureRequest(t, client.cert, http.MethodGet, customerURL(t)+query, header, nil)
		requireRefusedToken(t, resp, "Bearer")

		resp = secureRequest(t, client.cert, http.MethodGet, userInfoURL(t, client)+query, http.Header{}, nil)
		requireRefusedToken(t, resp, "Bearer")

		resp = secureRequest(
			t, client.cert, http.MethodPost, userInfoURL(t, client),
			http.Header{"Content-Type": {"application/x-www-form-urlencoded"}},
			strings.NewReader(url.Values{"access_token": {accessToken}}.Encode()),
		)
		requireRefusedToken(t, resp, "Bearer")
	})

	tokens := map[string]struct {
		token string
		cert  tls.Certificate
	}{
		"a token that the data holder did not issue":         {token: rand.Text(), cert: client.cert},
		"a token with a different client certificate":        {token: accessToken, cert: other},
		"a client credentials token of the software product": {token: client.accessToken(t), cert: client.cert},
	}
	for name, tt := range tokens {
		t.Run(name, func(t *testing.T) {
			requireRefusedToken(t, getCustomer(t, tt.cert, resourceHeader(tt.token)), invalid)
			requireRefusedToken(t, userInfo(t, client, tt.cert, http.MethodGet, tt.token), invalid)
		})
	}

	t.Run("a token that is not valid, in a request with other faults", func(t *testing.T) {
		header := http.Header{
			"Authorization":              {"Bearer " + rand.Text()},
			"X-Fapi-Customer-Ip-Address": {"not an address"},
		}

		requireRefusedToken(t, getCustomer(t, client.cert, header), invalid)
	})
}

func TestATokenWithoutTheScopeOfGetCustomerIsRefusedWith403(t *testing.T) {
	client := registered(t, allScopesProduct)
	accessToken := newEligibleAccount(t).consents(client, noAccountScope, ninetyDays).tokens(t)["access_token"].(string)

	for name, header := range map[string]http.Header{
		"a request with no other fault": resourceHeader(accessToken),
		"a request with other faults":   {"Authorization": {"Bearer " + accessToken}},
	} {
		t.Run(name, func(t *testing.T) {
			resp := getCustomer(t, client.cert, header)

			requireResourceError(
				t, resp, http.StatusForbidden, "urn:au-cds:error:cds-all:Authorisation/InvalidConsent",
				"The consent does not have the scope common:customer.basic:read.",
			)
			assert.Equal(
				t, `Bearer error="insufficient_scope", scope="common:customer.basic:read"`,
				resp.Header.Get("WWW-Authenticate"),
			)
		})
	}
}

func TestAResourceRequestNeedsTheHeadersOfTheCDS(t *testing.T) {
	client := registered(t, allScopesProduct)
	accessToken := newEligibleAccount(t).consents(client, customerScope, ninetyDays).tokens(t)["access_token"].(string)

	const (
		missing = "urn:au-cds:error:cds-all:Header/Missing"
		invalid = "urn:au-cds:error:cds-all:Header/Invalid"
	)

	tests := map[string]struct {
		change       func(http.Header)
		status       int
		code, detail string
	}{
		"no x-fapi-auth-date": {
			change: func(h http.Header) { h.Del("x-fapi-auth-date") },
			status: http.StatusBadRequest, code: missing, detail: "x-fapi-auth-date",
		},
		"an x-fapi-auth-date that is not an HTTP-date": {
			change: func(h http.Header) { h.Set("x-fapi-auth-date", time.Now().UTC().Format(time.RFC3339)) },
			status: http.StatusBadRequest, code: invalid, detail: "x-fapi-auth-date",
		},
		"an x-fapi-customer-ip-address that is not an IP address": {
			change: func(h http.Header) { h.Set("x-fapi-customer-ip-address", "203.0.113.300") },
			status: http.StatusBadRequest, code: invalid, detail: "x-fapi-customer-ip-address",
		},
		"no x-v": {
			change: func(h http.Header) { h.Del("x-v") },
			status: http.StatusBadRequest, code: missing, detail: "x-v",
		},
		"an x-v that is not a positive integer": {
			change: func(h http.Header) { h.Set("x-v", "one") },
			status: http.StatusBadRequest, code: "urn:au-cds:error:cds-all:Header/InvalidVersion", detail: "x-v",
		},
		"no version that the data holder supports": {
			change: func(h http.Header) { h.Set("x-v", "3"); h.Set("x-min-v", "2") },
			status: http.StatusNotAcceptable, code: "urn:au-cds:error:cds-all:Header/UnsupportedVersion",
			detail: "The supported versions are 1 to 1.",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			header := resourceHeader(accessToken)
			tt.change(header)

			resp := getCustomer(t, client.cert, header)

			requireResourceError(t, resp, tt.status, tt.code, tt.detail)
			assert.NotEmpty(t, resp.Header.Get("x-fapi-interaction-id"))
		})
	}

	t.Run("an x-min-v and a higher x-v get the highest supported version", func(t *testing.T) {
		header := resourceHeader(accessToken)
		header.Set("x-v", "9")
		header.Set("x-min-v", "1")

		resp := getCustomer(t, client.cert, header)

		require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
		assert.Equal(t, "1", resp.Header.Get("x-v"))
	})
}

func TestAnAPIOnTheResourceBaseURIThatIsOutOfScopeGivesNotImplemented(t *testing.T) {
	cert := softwareProductOf(t, allScopesProduct)

	for _, target := range []string{
		"GET /cds-au/v1/common/customer/detail",
		"GET /cds-au/v1/banking/payees",
		"GET /cds-au/v1/banking/accounts/" + uuid.NewString() + "/direct-debits",
		"POST /cds-au/v1/banking/payments/scheduled",
	} {
		method, path, _ := strings.Cut(target, " ")
		resp := secureRequest(t, cert, method, secureBaseURI(t)+path, http.Header{"X-V": {"1"}}, nil)

		requireResourceError(
			t, resp, http.StatusNotFound, "urn:au-cds:error:cds-all:Resource/NotImplemented",
			"The data holder does not implement this API.",
		)
	}
}
