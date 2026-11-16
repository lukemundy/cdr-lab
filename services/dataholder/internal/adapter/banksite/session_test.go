package banksite_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/banksite"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	brandID    = "8a6bd0a4-9c3e-4bfb-9a0e-3a64c3ef0d61"
	cookieName = "__Host-bank-session-" + brandID
	// The type of the 403 for a session cookie with no valid session.
	sessionEndedType = "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0010"
)

// customers is a store of customers in memory, with the rules of the data holder for a create.
type customers struct {
	mu        sync.Mutex
	records   map[string]domain.Customer
	createErr error
	closes    []closeCall
	closeErr  error
}

type closeCall struct {
	customerNumber, accountID string
	ends                      []string
}

func newCustomers(existing ...domain.Customer) *customers {
	c := &customers{records: map[string]domain.Customer{}}
	for _, customer := range existing {
		c.records[customer.CustomerNumber] = customer
	}

	return c
}

func (c *customers) Create(_ context.Context, customer domain.Customer) (domain.Customer, error) {
	if err := customer.Validate(); err != nil {
		return domain.Customer{}, err
	}

	if c.createErr != nil {
		return domain.Customer{}, c.createErr
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	customer.CustomerNumber = fmt.Sprintf("%08d", len(c.records)+1)
	c.records[customer.CustomerNumber] = customer

	return customer, nil
}

func (c *customers) Customer(_ context.Context, customerNumber string) (app.CustomerRecord, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	customer, ok := c.records[customerNumber]
	if !ok {
		return app.CustomerRecord{}, app.ErrCustomerNotFound
	}

	return app.CustomerRecord{Customer: customer}, nil
}

// Identify has the rule of app.Customers.Identify, for the customers of the fake.
func (c *customers) Identify(ctx context.Context, identifier string) (app.CustomerRecord, error) {
	identifier = strings.TrimSpace(identifier)
	if !strings.Contains(identifier, "@") {
		return c.Customer(ctx, identifier)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, customer := range c.records {
		if strings.EqualFold(customer.EmailAddress, identifier) {
			return app.CustomerRecord{Customer: customer}, nil
		}
	}

	return app.CustomerRecord{}, app.ErrCustomerNotFound
}

func (c *customers) CloseAccountThatEnds(
	_ context.Context,
	customerNumber, accountID string,
	ends []string,
) (domain.Account, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closeErr != nil {
		return domain.Account{}, c.closeErr
	}

	c.closes = append(c.closes, closeCall{customerNumber: customerNumber, accountID: accountID, ends: ends})

	return domain.Account{AccountID: accountID, OpenStatus: domain.OpenStatusClosed}, nil
}

func (c *customers) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.records)
}

var quoll = domain.Customer{
	CustomerNumber: "00047291", FirstName: "Zebracorn", LastName: "Quoll", EmailAddress: "quoll@mail.test",
}

// site is one instance of the site, so that a test can send more than one request to the same sessions.
type site struct {
	t     *testing.T
	mux   *http.ServeMux
	clock time.Time
}

func newSite(t *testing.T, store banksite.Customers) *site {
	t.Helper()

	return newSiteWith(t, banksite.Ports{Customers: store})
}

// newSiteWith gives the site of the ports, with the brand ID, the products and the clock of the site.
func newSiteWith(t *testing.T, ports banksite.Ports) *site {
	t.Helper()

	s := &site{t: t, mux: http.NewServeMux(), clock: now}
	ports.BrandID = brandID
	ports.Products = &products{}
	ports.Now = func() time.Time { return s.clock }
	require.NoError(t, banksite.Register(s.mux, ui, []string{host}, ports))

	return s
}

type request struct {
	method, target, body string
	cookies              []*http.Cookie
	header               map[string]string
}

// do validates the response against the OpenAPI document.
func (s *site) do(r request) response {
	s.t.Helper()

	var body io.Reader
	if r.body != "" {
		body = strings.NewReader(r.body)
	}

	req := httptest.NewRequestWithContext(s.t.Context(), r.method, "https://"+host+r.target, body)
	for _, cookie := range r.cookies {
		req.AddCookie(cookie)
	}

	for name, value := range r.header {
		req.Header.Set(name, value)
	}

	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() { _ = resp.Body.Close() }()

	resp.Request = req
	require.NoError(s.t, loadSpec(s.t).ValidateResponse(resp))

	data, err := io.ReadAll(resp.Body)
	require.NoError(s.t, err)

	return response{status: resp.StatusCode, header: resp.Header, body: string(data)}
}

func (s *site) signIn(identifier string, cookies ...*http.Cookie) response {
	s.t.Helper()

	return s.do(request{
		method: http.MethodPost, target: "/api/session", body: `{"identifier":"` + identifier + `"}`,
		cookies: cookies,
	})
}

func (s *site) readSession(cookies ...*http.Cookie) response {
	s.t.Helper()

	return s.do(request{method: http.MethodGet, target: "/api/session", cookies: cookies})
}

// requestCookie is a cookie that a browser sends, with the attributes of the session cookie.
func requestCookie(name, value string) *http.Cookie {
	return &http.Cookie{Name: name, Value: value, Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode}
}

// sessionCookie gives the session cookie of the response, and fails when the response has a different cookie.
func sessionCookie(t *testing.T, resp response) *http.Cookie {
	t.Helper()

	cookies, err := http.ParseSetCookie(strings.Join(resp.header.Values("Set-Cookie"), ""))
	require.NoError(t, err)
	require.Len(t, resp.header.Values("Set-Cookie"), 1)
	require.Equal(t, cookieName, cookies.Name)

	return cookies
}

// assertExpired checks the attributes of the session cookie, because a browser removes a __Host- cookie only with
// them.
func assertExpired(t *testing.T, resp response) {
	t.Helper()

	cookie := sessionCookie(t, resp)
	assert.Equal(t, -1, cookie.MaxAge, "Max-Age=0 parses as -1")
	assert.True(t, cookie.Secure)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.Equal(t, "/", cookie.Path)
	assert.Empty(t, cookie.Domain)
}

// Verifies: DH-1741, DH-1742
func TestASignUpCreatesTheCustomerAndStartsASession(t *testing.T) {
	store := newCustomers()
	s := newSite(t, store)

	resp := s.do(request{
		method: http.MethodPost, target: "/api/customers",
		body: `{"firstName": "Zebracorn", "lastName": "Quoll", "emailAddress": "quoll@mail.test"}`,
	})

	require.Equal(t, http.StatusCreated, resp.status, resp.body)
	assert.JSONEq(t, `{"customerNumber": "00000001", "firstName": "Zebracorn", "lastName": "Quoll"}`, resp.body)
	assert.Empty(t, resp.header.Get("Location"))
	assert.Equal(t, 1, store.count())

	session := s.readSession(sessionCookie(t, resp))
	require.Equal(t, http.StatusOK, session.status)
	assert.JSONEq(t, `{"customerNumber": "00000001", "firstName": "Zebracorn", "lastName": "Quoll"}`, session.body)
}

func TestTheSessionCookieHasOnlyTheIDWithTheHostPrefixAndNoExpiry(t *testing.T) {
	s := newSite(t, newCustomers(quoll))

	resp := s.signIn(quoll.CustomerNumber)

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)
	cookie := sessionCookie(t, resp)
	assert.True(t, cookie.Secure)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.Equal(t, "/", cookie.Path)
	assert.Empty(t, cookie.Domain)
	assert.Zero(t, cookie.MaxAge)
	assert.True(t, cookie.Expires.IsZero())
	assert.NotContains(t, cookie.Value, quoll.CustomerNumber)
	assert.GreaterOrEqual(t, len(cookie.Value), 22, "at least 128 bits in base64")
}

// Verifies: DH-1735, DH-1736
func TestASignUpWithValuesThatAreNotValidGivesAFieldErrorForEachValueInPlainLanguage(t *testing.T) {
	store := newCustomers()
	s := newSite(t, store)

	resp := s.do(request{
		method: http.MethodPost, target: "/api/customers",
		body: `{"firstName": "` + strings.Repeat("a", domain.MaxNameCharacters+1) + `", "lastName": "", ` +
			`"emailAddress": "quoll"}`,
	})

	require.Equal(t, http.StatusBadRequest, resp.status)
	assert.Empty(t, resp.header.Values("Set-Cookie"))
	assert.Zero(t, store.count())

	var body struct {
		Errors []struct {
			Pointer string `json:"pointer"`
			Detail  string `json:"detail"`
		} `json:"errors"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))

	details := map[string]string{}

	for _, e := range body.Errors {
		details[e.Pointer] = e.Detail
	}

	assert.Equal(t, map[string]string{
		"#/firstName":    "Enter a first name of 100 characters or fewer, or leave the first name empty.",
		"#/lastName":     "Enter a last name of 1 to 100 characters.",
		"#/emailAddress": "Enter an email address, for example name@example.test.",
	}, details)
}

// Verifies: DH-1719
func TestASignUpOrASignInWithABodyThatIsNotValidGives400(t *testing.T) {
	s := newSite(t, newCustomers(quoll))

	for target, body := range map[string]string{
		"/api/customers": `{"lastName": "Quoll", "emailAddress": "quoll@mail.test", "password": "hunter2"}`,
		"/api/session":   `null`,
	} {
		t.Run(target, func(t *testing.T) {
			resp := s.do(request{method: http.MethodPost, target: target, body: body})

			assert.Equal(t, http.StatusBadRequest, resp.status)
			assert.Empty(t, resp.header.Values("Set-Cookie"))
		})
	}
}

// Verifies: DH-1739
func TestASignInWithAPasswordFieldIsRefused(t *testing.T) {
	s := newSite(t, newCustomers(quoll))

	resp := s.do(request{
		method: http.MethodPost, target: "/api/session",
		body: `{"identifier": "00047291", "password": "correct horse battery staple"}`,
	})

	require.Equal(t, http.StatusBadRequest, resp.status)
	assert.Empty(t, resp.header.Values("Set-Cookie"))
	assert.NotContains(t, resp.body, "correct horse")
}

// Verifies: DH-1740
func TestASignInWithAnIdentifierThatNoCustomerHasGivesAFieldError(t *testing.T) {
	s := newSite(t, newCustomers(quoll))

	for _, identifier := range []string{"00000000", "", "quokka@mail.test"} {
		resp := s.signIn(identifier)

		require.Equal(t, http.StatusBadRequest, resp.status, identifier)
		assert.Empty(t, resp.header.Values("Set-Cookie"))
		assert.JSONEq(t, `[{
			"pointer": "#/identifier",
			"detail": "We could not sign you in. Check your email address or customer number."
		}]`, errorsOf(t, resp))
		assert.NotContains(t, resp.body, "quokka", "the response does not repeat the entry")
	}
}

func TestASignInWithTheEmailAddressInADifferentCaseStartsASessionOfTheCustomer(t *testing.T) {
	s := newSite(t, newCustomers(quoll))

	resp := s.signIn(" Quoll@Mail.TEST ")

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)

	session := s.readSession(sessionCookie(t, resp))
	require.Equal(t, http.StatusOK, session.status)
	assert.JSONEq(t, `{"customerNumber": "00047291", "firstName": "Zebracorn", "lastName": "Quoll"}`, session.body)
}

// Verifies: DH-1737
func TestASignUpWithAnEmailAddressInUseGives409WithAFieldErrorInPlainLanguage(t *testing.T) {
	store := newCustomers(quoll)
	store.createErr = fmt.Errorf("add the customer: %w", app.ErrEmailAddressUsed)
	s := newSite(t, store)

	resp := s.do(request{
		method: http.MethodPost, target: "/api/customers",
		body: `{"lastName": "Quoll", "emailAddress": "QUOLL@mail.test"}`,
	})

	require.Equal(t, http.StatusConflict, resp.status, resp.body)
	assert.Empty(t, resp.header.Values("Set-Cookie"))
	assert.JSONEq(t, `[{
		"pointer": "#/emailAddress",
		"detail": "A customer of the bank has this email address. Sign in, or enter a different email address."
	}]`, errorsOf(t, resp))
	assert.NotContains(t, resp.body, "QUOLL")
}

func errorsOf(t *testing.T, resp response) string {
	t.Helper()

	var body struct {
		Errors json.RawMessage `json:"errors"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))

	return string(body.Errors)
}

func TestEachSignInStartsANewSessionAndEndsTheSessionOfTheRequest(t *testing.T) {
	s := newSite(t, newCustomers(quoll))

	first := sessionCookie(t, s.signIn(quoll.CustomerNumber))
	second := sessionCookie(t, s.signIn(quoll.CustomerNumber, first))

	assert.NotEqual(t, first.Value, second.Value)
	assert.Equal(t, http.StatusForbidden, s.readSession(first).status)
	assert.Equal(t, http.StatusOK, s.readSession(second).status)
}

func TestASignUpEndsTheSessionOfTheRequest(t *testing.T) {
	s := newSite(t, newCustomers(quoll))
	old := sessionCookie(t, s.signIn(quoll.CustomerNumber))

	resp := s.do(request{
		method: http.MethodPost, target: "/api/customers", cookies: []*http.Cookie{old},
		body: `{"lastName": "Numbat", "emailAddress": "numbat@mail.test"}`,
	})

	require.Equal(t, http.StatusCreated, resp.status)
	assert.NotEqual(t, old.Value, sessionCookie(t, resp).Value)
	assert.Equal(t, http.StatusForbidden, s.readSession(old).status)
}

func TestACustomerCanHaveMoreThanOneSession(t *testing.T) {
	s := newSite(t, newCustomers(quoll))

	first := sessionCookie(t, s.signIn(quoll.CustomerNumber))
	second := sessionCookie(t, s.signIn(quoll.CustomerNumber))

	assert.Equal(t, http.StatusOK, s.readSession(first).status)
	assert.Equal(t, http.StatusOK, s.readSession(second).status)
}

// Verifies: DH-1743
func TestWithNoSessionCookieTheSessionIsForbidden(t *testing.T) {
	s := newSite(t, newCustomers(quoll))

	resp := s.readSession()

	require.Equal(t, http.StatusForbidden, resp.status)
	assert.Contains(t, resp.body, `"type":"about:blank"`)
	assert.Empty(t, resp.header.Values("Set-Cookie"))
}

// Verifies: DH-1744
func TestASessionCookieWithNoValidSessionSaysThatTheSessionEndedAndExpiresTheCookie(t *testing.T) {
	s := newSite(t, newCustomers(quoll))

	resp := s.readSession(requestCookie(cookieName, "not-a-session"))

	require.Equal(t, http.StatusForbidden, resp.status)
	assert.Contains(t, resp.body, `"type":"`+sessionEndedType+`"`)
	assertExpired(t, resp)
}

func TestASessionEnds24HoursAfterItsStart(t *testing.T) {
	s := newSite(t, newCustomers(quoll))
	cookie := sessionCookie(t, s.signIn(quoll.CustomerNumber))

	s.clock = now.Add(24*time.Hour - time.Second)
	assert.Equal(t, http.StatusOK, s.readSession(cookie).status)

	s.clock = now.Add(24 * time.Hour)
	assert.Equal(t, http.StatusForbidden, s.readSession(cookie).status)
}

func TestASignOutEndsTheSessionAndExpiresTheCookie(t *testing.T) {
	s := newSite(t, newCustomers(quoll))
	cookie := sessionCookie(t, s.signIn(quoll.CustomerNumber))

	resp := s.do(request{method: http.MethodDelete, target: "/api/session", cookies: []*http.Cookie{cookie}})

	require.Equal(t, http.StatusNoContent, resp.status)
	assertExpired(t, resp)
	assert.Empty(t, resp.header.Get("Clear-Site-Data"))
	assert.Equal(t, http.StatusForbidden, s.readSession(cookie).status)
}

// Verifies: DH-1706
func TestASignOutWithNoValidSessionAlsoExpiresTheCookie(t *testing.T) {
	s := newSite(t, newCustomers(quoll))

	resp := s.do(request{method: http.MethodDelete, target: "/api/session"})

	require.Equal(t, http.StatusNoContent, resp.status)
	assertExpired(t, resp)
}

func TestTheSiteDoesNotUseTheSessionCookieOfADifferentBank(t *testing.T) {
	s := newSite(t, newCustomers(quoll))
	other := requestCookie("__Host-bank-session-0d2f0bde-3c5c-4b51-9a7e-16a3d4d4f001", "other-bank")

	assert.Equal(t, http.StatusForbidden, s.readSession(other).status)
	assert.Empty(t, s.readSession(other).header.Values("Set-Cookie"))

	resp := s.signIn(quoll.CustomerNumber, other)
	assert.Equal(t, cookieName, sessionCookie(t, resp).Name)

	signOut := s.do(request{method: http.MethodDelete, target: "/api/session", cookies: []*http.Cookie{other}})
	assert.Equal(t, cookieName, sessionCookie(t, signOut).Name)
}

func TestAFullStoreRefusesASignInAndASignUpWith503AndCreatesNoCustomer(t *testing.T) {
	store := newCustomers(quoll)
	s := newSite(t, store)

	for range 10_000 {
		require.Equal(t, http.StatusNoContent, s.signIn(quoll.CustomerNumber).status)
	}

	assert.Equal(t, http.StatusServiceUnavailable, s.signIn(quoll.CustomerNumber).status)

	resp := s.do(request{
		method: http.MethodPost, target: "/api/customers",
		body: `{"lastName": "Numbat", "emailAddress": "numbat@mail.test"}`,
	})
	assert.Equal(t, http.StatusServiceUnavailable, resp.status)
	assert.Equal(t, 1, store.count())

	s.clock = now.Add(24 * time.Hour)
	assert.Equal(t, http.StatusNoContent, s.signIn(quoll.CustomerNumber).status)
}

func TestAFullStoreAcceptsASignInThatEndsAValidSessionOfTheRequest(t *testing.T) {
	store := newCustomers(quoll)
	s := newSite(t, store)

	var last *http.Cookie
	for range 10_000 {
		last = sessionCookie(t, s.signIn(quoll.CustomerNumber))
	}

	again := s.signIn(quoll.CustomerNumber, last)

	require.Equal(t, http.StatusNoContent, again.status)
	assert.Equal(t, http.StatusForbidden, s.readSession(last).status)
	assert.Equal(t, http.StatusOK, s.readSession(sessionCookie(t, again)).status)
	assert.Equal(t, http.StatusServiceUnavailable, s.signIn(quoll.CustomerNumber).status)

	signUp := s.do(request{
		method: http.MethodPost, target: "/api/customers", cookies: []*http.Cookie{sessionCookie(t, again)},
		body: `{"lastName": "Numbat", "emailAddress": "numbat@mail.test"}`,
	})
	assert.Equal(t, http.StatusServiceUnavailable, signUp.status)
	assert.Equal(t, 1, store.count())
}

func TestASignUpThatFailsGives500AndNoSession(t *testing.T) {
	store := newCustomers()
	store.createErr = errors.New("connection to 10.0.0.1 failed")
	s := newSite(t, store)

	resp := s.do(request{
		method: http.MethodPost, target: "/api/customers",
		body: `{"lastName": "Numbat", "emailAddress": "numbat@mail.test"}`,
	})

	assert.Equal(t, http.StatusInternalServerError, resp.status)
	assert.Empty(t, resp.header.Values("Set-Cookie"))
}

func TestTheSiteRefusesAChangeFromADifferentOrigin(t *testing.T) {
	for _, fetchSite := range []string{"cross-site", "same-site"} {
		for _, r := range []request{
			{
				method: http.MethodPost, target: "/api/customers",
				body: `{"lastName":"Numbat","emailAddress":"n@mail.test"}`,
			},
			{method: http.MethodPost, target: "/api/session", body: `{"identifier":"00047291"}`},
			{method: http.MethodDelete, target: "/api/session"},
		} {
			t.Run(fetchSite+" "+r.method+" "+r.target, func(t *testing.T) {
				store := newCustomers(quoll)
				s := newSite(t, store)
				cookie := sessionCookie(t, s.signIn(quoll.CustomerNumber))

				r.cookies = []*http.Cookie{cookie}
				r.header = map[string]string{"Sec-Fetch-Site": fetchSite}
				resp := s.do(r)

				assert.Equal(t, http.StatusForbidden, resp.status)
				assert.Empty(t, resp.header.Values("Set-Cookie"))
				assert.Equal(t, 1, store.count())
				assert.Equal(t, http.StatusOK, s.readSession(cookie).status)
			})
		}
	}
}

// DESIGN.md, Session: the rule of the account routes, also at a read of the session.
// Verifies: DH-1744
func TestAReadOfTheSessionOfACustomerThatTheDataHolderNoLongerHasEndsTheSession(t *testing.T) {
	store := newCustomers(quoll)
	s := newSite(t, store)
	cookie := requestCookie(cookieName, sessionCookie(t, s.signIn(quoll.CustomerNumber)).Value)

	store.mu.Lock()
	delete(store.records, quoll.CustomerNumber)
	store.mu.Unlock()

	resp := s.readSession(cookie)

	require.Equal(t, http.StatusForbidden, resp.status)
	assert.Contains(t, resp.body, `"type":"`+sessionEndedType+`"`)
	assertExpired(t, resp)

	store.mu.Lock()
	store.records[quoll.CustomerNumber] = quoll
	store.mu.Unlock()

	assert.Equal(t, http.StatusForbidden, s.readSession(cookie).status, "the session ended")
}
