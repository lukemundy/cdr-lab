package httpapi_test

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/httpapi"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	resourceBaseURL = "https://mtls.bank.example"
	userInfoPath    = "/userinfo"
	customerPath    = "/cds-au/v1/common/customer"
	authDate        = "Thu, 15 Oct 2026 08:00:00 GMT"
)

// consentTokens gives the consent token of each bearer value, and an invalid token error for a value that it does not
// have.
type consentTokens map[string]authserver.ConsentToken

func (v consentTokens) VerifyConsentToken(
	_ context.Context,
	token string,
	_ *x509.Certificate,
) (authserver.ConsentToken, error) {
	if token == "failure" {
		return authserver.ConsentToken{}, errors.New("signing keys not available")
	}

	verified, ok := v[token]
	if !ok {
		return authserver.ConsentToken{}, authserver.ErrInvalidToken
	}

	return verified, nil
}

// arrangementCustomers gives the customer of each arrangement ID.
type arrangementCustomers map[string]domain.Customer

func (a arrangementCustomers) Customer(_ context.Context, arrangementID string) (domain.Customer, error) {
	customer, ok := a[arrangementID]
	if !ok {
		return domain.Customer{}, errors.New("database failure")
	}

	return customer, nil
}

// currentArrangements gives the error of each arrangement in it, and no error for each other arrangement.
type currentArrangements map[string]error

func (c currentArrangements) RequireCurrent(_ context.Context, arrangementID string) error {
	return c[arrangementID]
}

// clientStatuses gives the status of each client, and an error for a client that it does not have.
type clientStatuses map[string]domain.SoftwareProductStatus

func (c clientStatuses) Status(_ context.Context, clientID string) (domain.SoftwareProductStatus, error) {
	status, ok := c[clientID]
	if !ok {
		return "", errors.New("database failure")
	}

	return status, nil
}

var lastUpdated = time.Date(2026, 10, 15, 13, 30, 0, 0, time.UTC)

func consentMux() *http.ServeMux {
	return secureMux(consentSecure())
}

func secureMux(s httpapi.Secure) *http.ServeMux {
	mux := http.NewServeMux()
	httpapi.RegisterSecure(mux, s)

	return mux
}

func consentSecure() httpapi.Secure {
	return httpapi.Secure{
		ResourceBaseURL:  resourceBaseURL + "/",
		RegistrationPath: registrationPath,
		UserInfoPath:     userInfoPath,
		Registrar:        &registrar{},
		Tokens:           tokens{},
		Registrations:    registrations{},
		ConsentTokens: consentTokens{
			"full": {
				ClientID:      "client-1",
				Scopes:        []string{"openid", "profile", "common:customer.basic:read"},
				Subject:       "subject-1",
				ArrangementID: "arrangement-1",
			},
			"no-profile": {
				ClientID:      "client-1",
				Scopes:        []string{"openid", "common:customer.basic:read"},
				Subject:       "subject-1",
				ArrangementID: "arrangement-1",
			},
			"no-first-name": {
				ClientID:      "client-1",
				Scopes:        []string{"openid", "profile", "common:customer.basic:read"},
				Subject:       "subject-2",
				ArrangementID: "arrangement-2",
			},
			"bank-only": {
				ClientID:      "client-1",
				Scopes:        []string{"openid", "bank:accounts.basic:read"},
				Subject:       "subject-1",
				ArrangementID: "arrangement-1",
			},
			"transactions": {
				ClientID:      "client-1",
				Scopes:        []string{"openid", "bank:transactions:read"},
				Subject:       "subject-1",
				ArrangementID: "arrangement-1",
			},
			"account-detail": {
				ClientID:      "client-1",
				Scopes:        []string{"openid", "bank:accounts.detail:read"},
				Subject:       "subject-1",
				ArrangementID: "arrangement-1",
			},
			"no-openid": {
				ClientID:      "client-1",
				Scopes:        []string{"common:customer.basic:read"},
				Subject:       "subject-1",
				ArrangementID: "arrangement-1",
			},
			"inactive": {
				ClientID:      "client-inactive",
				Scopes:        []string{"openid", "profile", "common:customer.basic:read"},
				Subject:       "subject-1",
				ArrangementID: "arrangement-1",
			},
			"inactive-bank-only": {
				ClientID:      "client-inactive",
				Scopes:        []string{"openid", "bank:accounts.basic:read"},
				Subject:       "subject-1",
				ArrangementID: "arrangement-1",
			},
			"status-failure": {
				ClientID:      "client-unknown",
				Scopes:        []string{"openid", "profile", "common:customer.basic:read"},
				Subject:       "subject-1",
				ArrangementID: "arrangement-1",
			},
			"no-arrangement": {
				ClientID:      "client-1",
				Scopes:        []string{"openid", "profile", "common:customer.basic:read"},
				Subject:       "subject-3",
				ArrangementID: "arrangement-3",
			},
			"ended": {
				ClientID:      "client-1",
				Scopes:        []string{"openid", "profile", "common:customer.basic:read"},
				Subject:       "subject-1",
				ArrangementID: "arrangement-ended",
			},
			"arrangement-failure": {
				ClientID:      "client-1",
				Scopes:        []string{"openid", "profile", "common:customer.basic:read"},
				Subject:       "subject-1",
				ArrangementID: "arrangement-failure",
			},
		},
		Arrangements: currentArrangements{
			"arrangement-ended":   app.ErrArrangementEnded,
			"arrangement-failure": errors.New("database failure"),
		},
		Customers: arrangementCustomers{
			"arrangement-1": {
				CustomerNumber: "00000001",
				FirstName:      "Alex",
				LastName:       "Example",
				EmailAddress:   "alex@example.com",
				LastUpdated:    lastUpdated,
			},
			"arrangement-2": {
				CustomerNumber: "00000002",
				LastName:       "Sample",
				EmailAddress:   "sample@example.com",
				LastUpdated:    lastUpdated,
			},
		},
		Accounts:     &arrangementAccounts{},
		Transactions: &accountTransactions{},
		Disclosures:  &disclosureRecords{},
		ClientStatuses: clientStatuses{
			"client-1": domain.SoftwareProductActive, "client-inactive": domain.SoftwareProductInactive,
		},
	}
}

func serve(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	consentMux().ServeHTTP(rec, req)

	return rec
}

// resourceRequest has each header that a request on the resource base URI needs.
func resourceRequest(t *testing.T, method, target, token string) *http.Request {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, target, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("x-v", "1")
	req.Header.Set("x-fapi-auth-date", authDate)

	return req
}

func requireCDSError(t *testing.T, rec *httptest.ResponseRecorder, status int, e cds.Error, detail string) {
	t.Helper()
	require.Equal(t, status, rec.Code, rec.Body.String())

	var body struct {
		Errors []map[string]string `json:"errors"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), rec.Body.String())
	assert.Equal(t, []map[string]string{{"code": e.Code, "title": e.Title, "detail": detail}}, body.Errors)
}

// Verifies: DH-1031, DH-1050, DH-1051
func TestGetCustomerGivesTheCustomerOfTheArrangementAsAPerson(t *testing.T) {
	rec := serve(resourceRequest(t, http.MethodGet, customerPath, "full"))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "1", rec.Header().Get("x-v"))
	assert.JSONEq(t, `{
		"data": {
			"customerUType": "person",
			"person": {
				"lastUpdateTime": "2026-10-15T13:30:00Z",
				"firstName": "Alex",
				"lastName": "Example",
				"middleNames": []
			}
		},
		"links": {"self": "https://mtls.bank.example/cds-au/v1/common/customer"},
		"meta": {}
	}`, rec.Body.String())
}

// Verifies: DH-1051
func TestGetCustomerOmitsTheFirstNameOfACustomerWhoHasNone(t *testing.T) {
	rec := serve(resourceRequest(t, http.MethodGet, customerPath, "no-first-name"))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.NotContains(t, rec.Body.String(), "firstName")
}

// Each request below has no other fault, unless its name says so.
// Verifies: DH-1020, DH-1026, DH-1027, DH-1030
func TestARefusedTokenGets401WithNoBody(t *testing.T) {
	cases := map[string]func(*http.Request){
		"no Authorization header": func(r *http.Request) { r.Header.Del("Authorization") },
		"a scheme that is not Bearer": func(r *http.Request) {
			r.Header.Set("Authorization", "Basic full")
		},
		"the token in the query only": func(r *http.Request) {
			r.Header.Del("Authorization")
			r.URL.RawQuery = url.Values{"access_token": {"full"}}.Encode()
		},
	}

	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			for _, target := range []string{customerPath, userInfoPath} {
				req := resourceRequest(t, http.MethodGet, target, "full")
				change(req)

				rec := serve(req)

				assert.Equal(t, http.StatusUnauthorized, rec.Code, target)
				assert.Equal(t, "Bearer", rec.Header().Get("WWW-Authenticate"), target)
				assert.Empty(t, rec.Body.String(), target)
			}
		})
	}

	t.Run("a token that the data holder does not accept", func(t *testing.T) {
		for _, target := range []string{customerPath, userInfoPath} {
			rec := serve(resourceRequest(t, http.MethodGet, target, "unknown"))

			assert.Equal(t, http.StatusUnauthorized, rec.Code, target)
			assert.Equal(t, `Bearer error="invalid_token"`, rec.Header().Get("WWW-Authenticate"), target)
			assert.Empty(t, rec.Body.String(), target)
		}
	})
}

// Verifies: DH-1021, DH-1027
func TestAValidTokenOfAnArrangementThatEndedGetsInvalidToken(t *testing.T) {
	for _, target := range []string{customerPath, userInfoPath} {
		rec := serve(resourceRequest(t, http.MethodGet, target, "ended"))

		assert.Equal(t, http.StatusUnauthorized, rec.Code, target)
		assert.Equal(t, `Bearer error="invalid_token"`, rec.Header().Get("WWW-Authenticate"), target)
		assert.Empty(t, rec.Body.String(), target)
	}
}

func TestATokenWhoseArrangementTheDataHolderCannotReadGets500(t *testing.T) {
	rec := serve(resourceRequest(t, http.MethodGet, customerPath, "arrangement-failure"))

	assert.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
}

// Verifies: DH-1024
func TestATokenCheckComesBeforeTheOtherChecksOfTheRequest(t *testing.T) {
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, customerPath, nil)
	req.Header.Set("Authorization", "Bearer unknown")
	req.Header.Set("x-fapi-customer-ip-address", "not-an-address")

	rec := serve(req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	req.Header.Set("Authorization", "Bearer bank-only")

	rec = serve(req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// Verifies: DH-1028
func TestATokenWithoutTheScopeOfTheAPIGets403(t *testing.T) {
	rec := serve(resourceRequest(t, http.MethodGet, customerPath, "bank-only"))

	requireCDSError(
		t, rec, http.StatusForbidden, cds.InvalidConsent,
		"The consent does not have the scope common:customer.basic:read.",
	)
	assert.Equal(
		t,
		`Bearer error="insufficient_scope", scope="common:customer.basic:read"`,
		rec.Header().Get("WWW-Authenticate"),
	)
}

// Verifies: DH-0434, DH-0435
func TestAResourceRequestNeedsAnAuthDateThatIsAnHTTPDate(t *testing.T) {
	req := resourceRequest(t, http.MethodGet, customerPath, "full")
	req.Header.Del("x-fapi-auth-date")
	requireCDSError(t, serve(req), http.StatusBadRequest, cds.HeaderMissing, "x-fapi-auth-date")

	req.Header.Set("x-fapi-auth-date", "2026-10-15T08:00:00Z")
	requireCDSError(t, serve(req), http.StatusBadRequest, cds.HeaderInvalid, "x-fapi-auth-date")

	// RFC 9110, section 5.6.7: a recipient accepts the two obsolete formats.
	for _, date := range []string{authDate, "Thursday, 15-Oct-26 08:00:00 GMT", "Thu Oct 15 08:00:00 2026"} {
		req.Header.Set("x-fapi-auth-date", date)
		assert.Equal(t, http.StatusOK, serve(req).Code, date)
	}
}

// Verifies: DH-0436, DH-0437
func TestAResourceRequestAcceptsOnlyAnIPAddressAsTheCustomerIPAddress(t *testing.T) {
	req := resourceRequest(t, http.MethodGet, customerPath, "full")

	for _, address := range []string{"203.0.113.7", "2001:db8::7"} {
		req.Header.Set("x-fapi-customer-ip-address", address)
		assert.Equal(t, http.StatusOK, serve(req).Code, address)
	}

	req.Header.Set("x-fapi-customer-ip-address", "203.0.113.300")
	requireCDSError(t, serve(req), http.StatusBadRequest, cds.HeaderInvalid, "x-fapi-customer-ip-address")
}

// Verifies: DH-0438
func TestAResourceRequestAcceptsAnyClientHeaders(t *testing.T) {
	req := resourceRequest(t, http.MethodGet, customerPath, "full")
	req.Header.Set("x-cds-client-headers", "not base64 {")

	assert.Equal(t, http.StatusOK, serve(req).Code)
}

func TestAResourceRequestUsesTheVersionRules(t *testing.T) {
	req := resourceRequest(t, http.MethodGet, customerPath, "full")
	req.Header.Del("x-v")
	requireCDSError(t, serve(req), http.StatusBadRequest, cds.HeaderMissing, "x-v")

	req.Header.Set("x-v", "2")
	req.Header.Set("x-min-v", "2")
	requireCDSError(
		t, serve(req), http.StatusNotAcceptable, cds.UnsupportedVersion, "The supported versions are 1 to 1.",
	)
}

func TestAFailureOfTheDataHolderGivesUnexpectedError(t *testing.T) {
	for _, token := range []string{"failure", "no-arrangement"} {
		rec := serve(resourceRequest(t, http.MethodGet, customerPath, token))

		assert.Equal(t, http.StatusInternalServerError, rec.Code, token)
		assert.Contains(t, rec.Body.String(), cds.Unexpected.Code, token)
	}
}

func userInfo(t *testing.T, method, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, userInfoPath, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	return serve(req)
}

// Verifies: DH-1101, DH-1102
func TestUserInfoGivesTheSubjectAndWithProfileTheNameOfTheCustomer(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		rec := userInfo(t, method, "full")

		require.Equal(t, http.StatusOK, rec.Code, method)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.JSONEq(t, `{
			"sub": "subject-1",
			"name": "Alex Example",
			"given_name": "Alex",
			"family_name": "Example",
			"updated_at": 1792071000
		}`, rec.Body.String(), method)
	}
}

// Verifies: DH-1101
func TestUserInfoGivesOnlyTheSubjectWithoutProfile(t *testing.T) {
	rec := userInfo(t, http.MethodGet, "no-profile")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"sub": "subject-1"}`, rec.Body.String())
}

// Verifies: DH-1102
func TestUserInfoOmitsTheGivenNameOfACustomerWhoHasNoFirstName(t *testing.T) {
	rec := userInfo(t, http.MethodGet, "no-first-name")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{
		"sub": "subject-2",
		"name": "Sample",
		"family_name": "Sample",
		"updated_at": 1792071000
	}`, rec.Body.String())
}

// Verifies: DH-1029, DH-1105
func TestUserInfoNeedsTheScopeOpenIDAndGives403WithTheCDSError(t *testing.T) {
	rec := userInfo(t, http.MethodGet, "no-openid")

	requireCDSError(t, rec, http.StatusForbidden, cds.InvalidConsent, "The consent does not have the scope openid.")
	assert.Equal(t, `Bearer error="insufficient_scope", scope="openid"`, rec.Header().Get("WWW-Authenticate"))
}

// Verifies: DH-1020
func TestUserInfoIgnoresATokenInTheBody(t *testing.T) {
	req := httptest.NewRequestWithContext(
		t.Context(), http.MethodPost, userInfoPath, strings.NewReader("access_token=full"),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rec := serve(req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "Bearer", rec.Header().Get("WWW-Authenticate"))
}

func TestUserInfoDoesNotUseTheVersionAndAuthDateHeaders(t *testing.T) {
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, userInfoPath, nil)
	req.Header.Set("Authorization", "Bearer full")
	req.Header.Set("x-v", "99")

	rec := serve(req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("x-v"))
}

// Verifies: DH-0428, DH-0429
func TestAnAPIThatIsOutOfScopeGivesNotImplemented(t *testing.T) {
	for _, target := range []string{
		"GET /cds-au/v1/common/customer/detail",
		"GET /cds-au/v1/banking/accounts/a1/direct-debits",
		"GET /cds-au/v1/banking/accounts/direct-debits",
		"POST /cds-au/v1/banking/accounts/direct-debits",
		"GET /cds-au/v1/banking/accounts/a1/payments/scheduled",
		"GET /cds-au/v1/banking/payments/scheduled",
		"POST /cds-au/v1/banking/payments/scheduled",
		"GET /cds-au/v1/banking/accounts/a1/payments/plans",
		"GET /cds-au/v1/banking/accounts/payments/plans",
		"GET /cds-au/v1/banking/payees",
		"GET /cds-au/v1/banking/payees/p1",
	} {
		method, path, _ := strings.Cut(target, " ")
		req := httptest.NewRequestWithContext(t.Context(), method, path, nil)

		requireCDSError(
			t, serve(req), http.StatusNotFound, cds.ResourceNotImplemented,
			"The data holder does not implement this API.",
		)
	}
}

// Verifies: DH-1033
func TestEachRequestOfTheDataOfAConsentIsLoggedWithItsInteractionIDAndNoCustomerData(t *testing.T) {
	var logs bytes.Buffer

	previous := slog.Default()

	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	for _, req := range []*http.Request{
		resourceRequest(t, http.MethodGet, customerPath, "full"),
		resourceRequest(t, http.MethodGet, customerPath, "unknown"),
		resourceRequest(t, http.MethodGet, userInfoPath, "full"),
	} {
		rec := httptest.NewRecorder()
		rec.Header().Set("x-fapi-interaction-id", "interaction-"+req.Header.Get("Authorization"))
		consentMux().ServeHTTP(rec, req)
	}

	lines := strings.Split(strings.TrimSpace(logs.String()), "\n")
	require.Len(t, lines, 3, logs.String())

	var records []map[string]any

	for _, line := range lines {
		var record map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &record))
		records = append(records, record)
	}

	assert.Equal(t, []any{"interaction-Bearer full"}, records[0]["http.response.header.x-fapi-interaction-id"])
	assert.Equal(t, "GET "+customerPath, records[0]["http.route"])
	assert.EqualValues(t, http.StatusOK, records[0]["http.response.status_code"])
	assert.EqualValues(t, http.StatusUnauthorized, records[1]["http.response.status_code"])
	assert.Equal(t, "GET "+userInfoPath, records[2]["http.route"])

	for _, value := range []string{"Alex", "Example", "00000001", "subject-1", "alex@example.com"} {
		assert.NotContains(t, logs.String(), value)
	}
}

func TestAResourceRequestOfASoftwareProductThatIsNotActiveGets403WithTheStatus(t *testing.T) {
	rec := serve(resourceRequest(t, http.MethodGet, customerPath, "inactive"))

	requireCDSError(t, rec, http.StatusForbidden, cds.ADRStatusNotActive, "INACTIVE")
}

func TestUserInfoOfASoftwareProductThatIsNotActiveGets401InvalidTokenWithNoBody(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		rec := serve(resourceRequest(t, method, userInfoPath, "inactive"))

		assert.Equal(t, http.StatusUnauthorized, rec.Code, method)
		assert.Equal(t, `Bearer error="invalid_token"`, rec.Header().Get("WWW-Authenticate"), method)
		assert.Empty(t, rec.Body.String(), method)
	}
}

func TestTheStatusCheckComesAfterTheTokenCheckAndBeforeTheOtherChecks(t *testing.T) {
	scope := serve(resourceRequest(t, http.MethodGet, customerPath, "inactive-bank-only"))
	requireCDSError(t, scope, http.StatusForbidden, cds.ADRStatusNotActive, "INACTIVE")

	req := resourceRequest(t, http.MethodGet, customerPath, "inactive")
	req.Header.Del("x-fapi-auth-date")
	requireCDSError(t, serve(req), http.StatusForbidden, cds.ADRStatusNotActive, "INACTIVE")
}

func TestAStatusThatTheDataHolderCannotReadGivesUnexpectedError(t *testing.T) {
	rec := serve(resourceRequest(t, http.MethodGet, customerPath, "status-failure"))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
