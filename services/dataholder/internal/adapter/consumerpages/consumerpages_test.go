package consumerpages_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/consumerpages"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	clientID = "client-1"
	flowPath = "/authorize/flow-1"
)

// subjects gives one sub, and records the customer number and the sector identifier of each call.
type subjects struct {
	calls [][2]string
}

func (s *subjects) Subject(_ context.Context, customerNumber, sectorIdentifier, _ string) (string, error) {
	s.calls = append(s.calls, [2]string{customerNumber, sectorIdentifier})

	return "subject-1", nil
}

type registrations struct {
	err             error
	legalEntityName string
	// orgName is the brand name of the data recipient, or the default name when it is empty.
	orgName string
}

func (r *registrations) Registration(_ context.Context, id string) (domain.Registration, error) {
	if r.err != nil {
		return domain.Registration{}, r.err
	}

	if id != clientID {
		return domain.Registration{}, app.ErrRegistrationNotFound
	}

	orgName := r.orgName
	if orgName == "" {
		orgName = "Budget <App> & Co"
	}

	return domain.Registration{
		ClientID:        clientID,
		LegalEntityName: r.legalEntityName,
		OrgName:         orgName,
		ClientName:      "Budget Tracker",
		RedirectURIs:    []string{"https://recipient.example/callback"},
		SoftwareID:      "software-1",
	}, nil
}

const (
	knownCustomer = "00000001"
	knownEmail    = "alex@example.com"
)

var browsers atomic.Uint64

type customers struct {
	app.CustomerStore

	err      error
	accounts []domain.Account
}

func ownAccounts() []domain.Account {
	return []domain.Account{
		{
			AccountID: "account-1", DisplayName: "Everyday", BSB: "123456", AccountNumber: "12345678",
			OpenStatus: domain.OpenStatusOpen,
		},
		{
			AccountID: "account-2", DisplayName: "Holiday <Saver>", BSB: "123456", AccountNumber: "87654321",
			OpenStatus: domain.OpenStatusOpen,
		},
		{
			AccountID: "account-3", DisplayName: "Closed", BSB: "123456", AccountNumber: "11223344",
			OpenStatus: domain.OpenStatusClosed,
		},
	}
}

// arrangements keeps each append of the pages, as the event store does.
type arrangements struct {
	appends []eventsourcing.Append
	err     error
}

func (a *arrangements) Read(context.Context, string, string) ([]eventsourcing.RecordedEvent, error) {
	return nil, nil
}

func (a *arrangements) Append(_ context.Context, appends ...eventsourcing.Append) error {
	if a.err != nil {
		return a.err
	}

	a.appends = append(a.appends, appends...)

	return nil
}

func (c *customers) Customer(_ context.Context, number string) (app.CustomerRecord, error) {
	if c.err != nil {
		return app.CustomerRecord{}, c.err
	}

	if number != knownCustomer {
		return app.CustomerRecord{}, app.ErrCustomerNotFound
	}

	return app.CustomerRecord{
		Customer: domain.Customer{CustomerNumber: number, EmailAddress: knownEmail},
		Accounts: c.accounts,
	}, nil
}

func (c *customers) CustomerNumberOf(_ context.Context, emailAddress string) (string, error) {
	if c.err != nil {
		return "", c.err
	}

	if !strings.EqualFold(emailAddress, knownEmail) {
		return "", app.ErrCustomerNotFound
	}

	return knownCustomer, nil
}

type outbox struct {
	// otps has the OTP of each email, which the test reads as the customer reads the email.
	otps []string
	err  error
}

var emailOTP = regexp.MustCompile(`one-time password is ([0-9]{6})`)

func (o *outbox) Send(_ context.Context, _ string, email domain.Email) error {
	if o.err != nil {
		return o.err
	}

	o.otps = append(o.otps, emailOTP.FindStringSubmatch(email.Heading)[1])

	return nil
}

// browser is one consumer in one flow: it keeps the cookies and the state of the flow between requests.
type browser struct {
	t       *testing.T
	pages   consumerpages.Pages
	mail    *outbox
	now     time.Time
	cookies []*http.Cookie
	state   any
	// scopes and sharingDuration are those of the authorisation request.
	scopes          []string
	sharingDuration time.Duration
	customers       *customers
	registrations   *registrations
	arrangements    *arrangements
	subjects        *subjects
}

const bankName = "Coastal Bank"

func newBrowser(t *testing.T) *browser {
	t.Helper()

	return newBrowserOfBank(t, bankName)
}

// newBrowserOfBank gives a browser for the pages of a data holder with a different brand name.
func newBrowserOfBank(t *testing.T, bank string) *browser {
	t.Helper()

	b := &browser{
		t:             t,
		mail:          &outbox{},
		now:           time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC),
		scopes:        []string{"openid", "bank:accounts.basic:read"},
		customers:     &customers{accounts: ownAccounts()},
		registrations: &registrations{legalEntityName: "Budget Holdings Pty Ltd"},
		arrangements:  &arrangements{},
		subjects:      &subjects{},
	}
	now := func() time.Time { return b.now }
	// Each browser gets different OTPs, as different flows do.
	next := browsers.Add(1) * 1000
	b.pages = consumerpages.New(
		bank,
		b.registrations,
		app.Authenticator{
			Customers: b.customers,
			Mail:      b.mail,
			Random: func() uint64 {
				next++
				return next
			},
			Now: now,
		},
		app.Consents{
			Customers: b.customers,
			Subjects:  b.subjects,
			Store:     b.arrangements,
			NewID:     func() string { return "arrangement-1" },
			Now:       now,
		},
		now,
	)

	return b
}

// identify gives the OTP page for the email address or the customer number.
func (b *browser) identify(identifier string) page {
	b.t.Helper()

	token := formToken(b.t, b.get().body)

	return b.post(url.Values{"csrf_token": {token}, "action": {"continue"}, "identifier": {identifier}})
}

func (b *browser) enter(otp string) page {
	b.t.Helper()

	token := formToken(b.t, b.get().body)

	return b.post(url.Values{"csrf_token": {token}, "action": {"continue"}, "otp": {otp}})
}

func (b *browser) act(action string) page {
	b.t.Helper()

	token := formToken(b.t, b.get().body)

	return b.post(url.Values{"csrf_token": {token}, "action": {action}})
}

type page struct {
	status  int
	header  http.Header
	body    string
	outcome authserver.Outcome
}

func (b *browser) do(req *http.Request) page {
	b.t.Helper()

	for _, cookie := range b.cookies {
		req.AddCookie(cookie)
	}

	rec := httptest.NewRecorder()
	outcome := b.pages.Interact(rec, req, authserver.Interaction{
		ClientID:        clientID,
		Scopes:          b.scopes,
		SharingDuration: b.sharingDuration,
		Path:            flowPath,
		State:           b.state,
	})

	if cookies := rec.Result().Cookies(); len(cookies) > 0 {
		b.cookies = cookies
	}

	if next, ok := outcome.Continues(); ok {
		b.state = next
	}

	return page{status: rec.Code, header: rec.Header(), body: rec.Body.String(), outcome: outcome}
}

func (b *browser) get() page {
	b.t.Helper()
	return b.do(httptest.NewRequestWithContext(b.t.Context(), http.MethodGet, flowPath, nil))
}

func (b *browser) post(form url.Values) page {
	b.t.Helper()

	req := httptest.NewRequestWithContext(b.t.Context(), http.MethodPost, flowPath, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return b.do(req)
}

var hiddenToken = regexp.MustCompile(`<input type="hidden" name="csrf_token" value="([^"]+)">`)

func formToken(t *testing.T, body string) string {
	t.Helper()

	match := hiddenToken.FindStringSubmatch(body)
	require.Len(t, match, 2, "the page has a form token: %s", body)

	return match[1]
}

func TestTheFirstPageAsksOnlyForTheEmailAddressOrTheCustomerNumberInOneField(t *testing.T) {
	b := newBrowser(t)

	first := b.get()

	require.Equal(t, http.StatusOK, first.status)
	assert.Equal(t, "text/html; charset=utf-8", first.header.Get("Content-Type"))
	assert.Contains(t, first.body, `<form method="post" action="/authorize/flow-1">`)
	assert.Contains(t, first.body, `<label for="identifier">Email address or customer number</label>`)
	assert.Contains(t, first.body, `name="identifier"`)
	assert.NotContains(t, first.body, `inputmode="numeric"`, "an email address needs the full keyboard")
	assert.Equal(t, 1, strings.Count(first.body, `<input type="text"`), "the page asks for one value")
	assert.NotContains(t, first.body, `type="password"`)
	assert.NotContains(t, first.body, "<a ", "the page has no links")
	assert.NotContains(t, first.body, "<script")
	assert.NotContains(t, first.body, "http://")
	assert.NotContains(t, first.body, "https://")
}

func TestTheFirstPageNamesTheBrandOfTheDataRecipientAndSaysThatNoPasswordIsNecessary(t *testing.T) {
	first := newBrowser(t).get()

	assert.Contains(t, first.body, "Budget &lt;App&gt; &amp; Co", "the brand name is escaped")
	assert.Contains(t, first.body, "does not need your password")
	assert.Contains(t, first.body, "Do not give your password")
}

func TestEachPageHasTheHeadersOfAnAuthenticationPage(t *testing.T) {
	first := newBrowser(t).get()

	assert.Equal(
		t,
		"default-src 'none'; style-src 'self'; font-src 'self'; img-src 'self'; frame-ancestors 'none'; "+
			"base-uri 'none'",
		first.header.Get("Content-Security-Policy"),
	)
	assert.Equal(t, "no-referrer", first.header.Get("Referrer-Policy"))
	assert.Equal(t, "nosniff", first.header.Get("X-Content-Type-Options"))
	assert.Equal(t, "no-store", first.header.Get("Cache-Control"))
}

func TestTheFirstPageGivesTheBrowserACookieThatScriptsCannotRead(t *testing.T) {
	b := newBrowser(t)

	first := b.get()

	require.Len(t, b.cookies, 1)
	cookie := b.cookies[0]
	assert.Equal(t, "__Host-cdr-flow", cookie.Name)
	assert.Equal(t, formToken(t, first.body), cookie.Value)
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.GreaterOrEqual(t, len(cookie.Value), 32)
}

func TestTwoFlowsGetDifferentFormTokens(t *testing.T) {
	first, second := newBrowser(t).get(), newBrowser(t).get()

	assert.NotEqual(t, formToken(t, first.body), formToken(t, second.body))
}

func TestACancelDeniesTheFlowAndWritesNoPage(t *testing.T) {
	b := newBrowser(t)
	token := formToken(t, b.get().body)

	cancelled := b.post(url.Values{"csrf_token": {token}, "action": {"cancel"}})

	assert.Equal(t, authserver.Deny(authserver.AccessDenied), cancelled.outcome)
	assert.Empty(t, cancelled.body)
}

func TestAFormWithoutItsProtectionAgainstCrossSiteRequestForgeryIsRefused(t *testing.T) {
	tests := map[string]func(b *browser, token string) url.Values{
		"no form token": func(*browser, string) url.Values { return url.Values{"action": {"cancel"}} },
		"wrong form token": func(*browser, string) url.Values {
			return url.Values{"csrf_token": {"not-the-token"}, "action": {"cancel"}}
		},
		"no cookie": func(b *browser, token string) url.Values {
			b.cookies = nil
			return url.Values{"csrf_token": {token}, "action": {"cancel"}}
		},
		"cookie of a different browser": func(b *browser, token string) url.Values {
			other := newBrowser(b.t)
			other.get()
			b.cookies = other.cookies

			return url.Values{"csrf_token": {token}, "action": {"cancel"}}
		},
	}
	for name, form := range tests {
		t.Run(name, func(t *testing.T) {
			b := newBrowser(t)
			token := formToken(t, b.get().body)

			refused := b.post(form(b, token))

			assert.Equal(t, http.StatusForbidden, refused.status)
			_, continues := refused.outcome.Continues()
			assert.True(t, continues, "the flow does not end")
			assert.NotContains(t, refused.body, token)
		})
	}
}

func TestTheFlowContinuesAfterARefusedForm(t *testing.T) {
	b := newBrowser(t)
	token := formToken(t, b.get().body)
	b.post(url.Values{"action": {"cancel"}})

	cancelled := b.post(url.Values{"csrf_token": {token}, "action": {"cancel"}})

	assert.Equal(t, authserver.Deny(authserver.AccessDenied), cancelled.outcome)
}

func TestAReloadShowsTheSamePageWithTheSameFormToken(t *testing.T) {
	b := newBrowser(t)
	first := b.get()

	again := b.get()

	assert.Equal(t, http.StatusOK, again.status)
	assert.Equal(t, formToken(t, first.body), formToken(t, again.body))
}

func TestACustomerNumberGetsTheOTPPageAndAnEmailWithAnOTP(t *testing.T) {
	b := newBrowser(t)

	otpPage := b.identify(knownCustomer)

	require.Equal(t, http.StatusOK, otpPage.status)
	require.Len(t, b.mail.otps, 1)
	assert.Contains(t, otpPage.body, `name="otp"`)
	assert.Contains(t, otpPage.body, "valid for 5 minutes")
	assert.Contains(t, otpPage.body, "does not need your password")
	assert.Contains(t, otpPage.body, `value="resend"`)
	assert.Contains(t, otpPage.body, `value="cancel"`)
	assert.NotContains(t, otpPage.body, knownCustomer, "the page does not repeat the customer number")
	assert.NotContains(t, otpPage.body, b.mail.otps[0], "the page does not show the OTP")
	assert.NotContains(t, otpPage.body, knownEmail, "the page does not show the email address")
	assert.NotContains(t, otpPage.body, "<a ")
}

func TestTheEmailAddressInADifferentCaseGetsTheSameOTPPageAsTheCustomerNumber(t *testing.T) {
	byNumber, byEmail := newBrowser(t), newBrowser(t)

	numberPage := byNumber.identify(knownCustomer)
	emailPage := byEmail.identify(" Alex@Example.COM ")

	require.Len(t, byEmail.mail.otps, 1)
	assert.Equal(t, numberPage.status, emailPage.status)
	assert.Equal(
		t,
		hiddenToken.ReplaceAllString(numberPage.body, ""),
		hiddenToken.ReplaceAllString(emailPage.body, ""),
		"the two pages are the same, except for the form token",
	)
	assert.NotContains(t, emailPage.body, "Alex@Example.COM", "the page does not repeat the entry")
}

func TestAnEmailAddressThatNoCustomerHasGetsTheSameOTPPageAndNoEmail(t *testing.T) {
	known, unknown := newBrowser(t), newBrowser(t)

	knownPage := known.identify(knownEmail)
	unknownPage := unknown.identify("sam@example.com")

	assert.Empty(t, unknown.mail.otps)
	assert.Equal(t, knownPage.status, unknownPage.status)
	assert.Equal(
		t,
		hiddenToken.ReplaceAllString(knownPage.body, ""),
		hiddenToken.ReplaceAllString(unknownPage.body, ""),
		"the two pages are the same, except for the form token",
	)
}

func TestACustomerNumberThatDoesNotExistGetsTheSameOTPPageAndNoEmail(t *testing.T) {
	known, unknown := newBrowser(t), newBrowser(t)

	knownPage := known.identify(knownCustomer)
	unknownPage := unknown.identify("99999999")

	assert.Empty(t, unknown.mail.otps)
	assert.Equal(t, knownPage.status, unknownPage.status)
	assert.Equal(
		t,
		hiddenToken.ReplaceAllString(knownPage.body, ""),
		hiddenToken.ReplaceAllString(unknownPage.body, ""),
		"the two pages are the same, except for the form token",
	)
}

func TestAnEmptyIdentifierIsABadRequestAndSendsNothing(t *testing.T) {
	for name, identifier := range map[string]string{"empty": "", "spaces only": "   "} {
		t.Run(name, func(t *testing.T) {
			b := newBrowser(t)

			again := b.identify(identifier)

			assert.Equal(t, http.StatusBadRequest, again.status)
			assert.Contains(t, again.body, `name="identifier"`)
			assert.Empty(t, b.mail.otps)
		})
	}
}

func TestTheCorrectOTPGivesTheConsentPage(t *testing.T) {
	b := newBrowser(t)
	b.identify(knownCustomer)

	done := b.enter(b.mail.otps[0])

	require.Equal(t, http.StatusOK, done.status)
	assert.NotContains(t, done.body, `name="otp"`)
	assert.Contains(t, done.body, `value="confirm"`)
	assert.Contains(t, done.body, `value="cancel"`)
	_, continues := done.outcome.Continues()
	assert.True(t, continues)
}

func TestAnIncorrectOTPGivesTheOTPPageAgainWithANotice(t *testing.T) {
	b := newBrowser(t)
	b.identify(knownCustomer)

	again := b.enter("000000")

	require.Equal(t, http.StatusOK, again.status)
	assert.Contains(t, again.body, "not correct")
	assert.Contains(t, again.body, `name="otp"`)
	assert.NotContains(t, again.body, "000000", "the page does not repeat the entry")
}

func TestTheThirdIncorrectOTPDeniesTheFlow(t *testing.T) {
	b := newBrowser(t)
	b.identify(knownCustomer)
	b.enter("000001")
	b.enter("000002")

	third := b.enter("000003")

	assert.Equal(t, authserver.Deny(authserver.AccessDenied), third.outcome)
	assert.Empty(t, third.body)
}

func TestAnOTPIsRefusedAfter5MinutesAndAfterOneUse(t *testing.T) {
	t.Run("after 5 minutes", func(t *testing.T) {
		b := newBrowser(t)
		b.identify(knownCustomer)
		b.now = b.now.Add(5 * time.Minute)

		assert.Contains(t, b.enter(b.mail.otps[0]).body, "not correct")
	})

	t.Run("in a different flow", func(t *testing.T) {
		first, second := newBrowser(t), newBrowser(t)
		first.identify(knownCustomer)
		second.identify(knownCustomer)

		assert.Contains(t, second.enter(first.mail.otps[0]).body, "not correct")
	})
}

func TestANewOTPReplacesThePreviousOne(t *testing.T) {
	b := newBrowser(t)
	b.identify(knownCustomer)

	resent := b.act("resend")

	require.Equal(t, http.StatusOK, resent.status)
	require.Len(t, b.mail.otps, 2)
	assert.Contains(t, resent.body, "sent you a new one-time password")
	assert.Contains(t, b.enter(b.mail.otps[0]).body, "not correct")
	assert.Contains(t, b.enter(b.mail.otps[1]).body, `value="confirm"`)
}

func TestTheFlowSends3OTPsAtMost(t *testing.T) {
	b := newBrowser(t)
	b.identify(knownCustomer)
	b.act("resend")

	third := b.act("resend")
	fourth := b.act("resend")

	assert.Len(t, b.mail.otps, 3)
	assert.NotContains(t, third.body, `value="resend"`, "the page has no control for a fourth OTP")
	assert.Contains(t, third.body, "cannot send you a different one-time password")
	assert.Equal(t, http.StatusOK, fourth.status)
	assert.NotContains(t, fourth.body, "sent you a new one-time password")
	assert.Contains(t, b.enter(b.mail.otps[2]).body, `value="confirm"`, "the third OTP stays valid")
}

func TestAMailFailureEndsTheFlowAsTemporarilyUnavailable(t *testing.T) {
	b := newBrowser(t)
	b.mail.err = errors.New("the SMTP server is not available")

	failed := b.identify(knownCustomer)

	assert.Equal(t, authserver.Deny(authserver.TemporarilyUnavailable), failed.outcome)
	assert.Empty(t, failed.body)
}

func TestACancelOnTheOTPPageAndOnTheConsentPageDeniesTheFlow(t *testing.T) {
	t.Run("OTP page", func(t *testing.T) {
		b := newBrowser(t)
		b.identify(knownCustomer)

		assert.Equal(t, authserver.Deny(authserver.AccessDenied), b.act("cancel").outcome)
	})

	t.Run("consent page", func(t *testing.T) {
		b := newBrowser(t)
		b.identify(knownCustomer)
		b.enter(b.mail.otps[0])

		assert.Equal(t, authserver.Deny(authserver.AccessDenied), b.act("cancel").outcome)
	})
}

func TestAReloadOfTheOTPPageSendsNoEmail(t *testing.T) {
	b := newBrowser(t)
	b.identify(knownCustomer)

	again := b.get()

	assert.Contains(t, again.body, `name="otp"`)
	assert.Len(t, b.mail.otps, 1)
}

func TestAnActionThatThePageDoesNotHaveIsABadRequest(t *testing.T) {
	b := newBrowser(t)
	token := formToken(t, b.get().body)

	refused := b.post(url.Values{"csrf_token": {token}, "action": {"approve"}})

	assert.Equal(t, http.StatusBadRequest, refused.status)
	_, continues := refused.outcome.Continues()
	assert.True(t, continues)
}

func TestAFlowOfAClientWhoseRegistrationCannotBeReadIsTemporarilyUnavailable(t *testing.T) {
	pages := consumerpages.New(
		bankName, &registrations{err: errors.New("connection lost")}, app.Authenticator{}, app.Consents{}, time.Now,
	)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, flowPath, nil)

	outcome := pages.Interact(rec, req, authserver.Interaction{ClientID: clientID, Path: flowPath})

	assert.Equal(t, authserver.Deny(authserver.TemporarilyUnavailable), outcome)
	assert.Empty(t, rec.Body.String())
}

func TestTheRefusedPageSaysNothingAboutTheCause(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusInternalServerError} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodGet,
			"/authorize?request_uri=secret-reference",
			nil,
		)

		consumerpages.New(bankName, &registrations{}, app.Authenticator{}, app.Consents{}, time.Now).
			Refused(rec, req, status)

		assert.Equal(t, status, rec.Code)
		assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.Equal(t, "no-referrer", rec.Header().Get("Referrer-Policy"))
		assert.Contains(t, rec.Body.String(), "cannot continue")
		assert.NotContains(t, rec.Body.String(), "secret-reference")
		assert.NotContains(t, rec.Body.String(), "<a ")
	}
}
