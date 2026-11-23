//go:build conformance

package dataholder

import (
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

// sinkEmail is an email that the mail sink of the lab keeps.
type sinkEmail struct {
	Subject string `json:"Subject"`
	Text    string `json:"Text"`
}

func mailSink(t *testing.T, path string, into any) {
	t.Helper()

	resp := get(t, labtest.Env(t, "MAIL_SINK_URL")+path, http.Header{})
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
	require.NoError(t, json.Unmarshal(resp.body, into), string(resp.body))
}

// emailsTo gives the emails to the address, the oldest first.
func emailsTo(t *testing.T, address string) []sinkEmail {
	t.Helper()

	var found struct {
		Messages []struct {
			ID string `json:"ID"`
		} `json:"messages"`
	}

	mailSink(t, "/api/v1/search?"+url.Values{"query": {`to:"` + address + `"`}}.Encode(), &found)

	emails := make([]sinkEmail, len(found.Messages))
	// The mail sink gives the newest message first.
	for i, message := range found.Messages {
		mailSink(t, "/api/v1/message/"+message.ID, &emails[len(emails)-1-i])
	}

	return emails
}

func emailCount(t *testing.T) int {
	t.Helper()

	var all struct {
		Total int `json:"total"`
	}

	mailSink(t, "/api/v1/messages?limit=1", &all)

	return all.Total
}

var emailOTP = regexp.MustCompile(`one-time password is ([0-9]{6})\.`)

func (e sinkEmail) otp(t *testing.T) string {
	t.Helper()

	match := emailOTP.FindStringSubmatch(e.Text)
	require.Len(t, match, 2, "the email has an OTP of 6 digits: %s", e.Text)

	return match[1]
}

// account is a customer of the bank with its own email address, in the browser of that customer.
type account struct {
	t        *testing.T
	customer operatorCustomer
	browser  *consumer
}

func newAccount(t *testing.T) account {
	t.Helper()

	address := strings.ToLower(rand.Text()) + "@mail.test"

	return account{t: t, customer: newCustomerWithEmail(t, address), browser: newConsumer(t)}
}

var accountNumbers atomic.Int64

// newAccountNumber gives a different account number at each call. The other tests use account numbers that do not
// start with 8.
func newAccountNumber() string {
	return strconv.FormatInt(80_000_000+accountNumbers.Add(1), 10)
}

// newEligibleAccount gives a customer with an open account, because a customer with no open account cannot give a
// consent.
func newEligibleAccount(t *testing.T) account {
	t.Helper()

	customer := newAccount(t)
	customer.addAccount("Everyday", newAccountNumber())

	return customer
}

// identify starts a flow of the software product and gives the customer number. It gives the OTP page.
func (a account) identify(client registeredClient, claims map[string]any) webPage {
	a.t.Helper()

	return a.identifyWith(client, claims, a.customer.CustomerNumber)
}

// identifyWith starts a flow of the software product and gives the identifier. It gives the OTP page.
func (a account) identifyWith(client registeredClient, claims map[string]any, identifier string) webPage {
	a.t.Helper()

	first := a.browser.open(client.authorisationURL(a.t, claims))
	require.Equal(a.t, http.StatusOK, first.status, first.body)

	otpPage := a.browser.submit(first, "continue", url.Values{"identifier": {identifier}})
	require.Equal(a.t, http.StatusOK, otpPage.status, otpPage.body)

	return otpPage
}

// otps gives the OTP of each email to the customer, the oldest first, when the customer has the number of emails.
func (a account) otps(want int) []string {
	a.t.Helper()

	var emails []sinkEmail

	require.Eventually(a.t, func() bool {
		emails = emailsTo(a.t, a.customer.EmailAddress)
		return len(emails) >= want
	}, labtest.PollLimit, 100*time.Millisecond, "the mail sink has %d emails to the customer", want)
	require.Len(a.t, emails, want)

	otps := make([]string, len(emails))
	for i, email := range emails {
		otps[i] = email.otp(a.t)
	}

	return otps
}

func isOTPPage(p webPage) bool {
	return strings.Contains(p.body, `name="otp"`)
}

func TestACustomerNumberGetsAnOTPByEmailAndTheOTPPage(t *testing.T) {
	client := registered(t, registrationTestProduct)
	customer := newAccount(t)

	otpPage := customer.identify(client, client.requestObjectClaims())

	assert.True(t, isOTPPage(otpPage), otpPage.body)
	assert.Contains(t, otpPage.body, "valid for 5 minutes", "the page gives the time for which the OTP is valid")
	assert.NotContains(t, otpPage.body, customer.customer.EmailAddress)
	assert.NotContains(t, otpPage.body, "<a ")

	emails := emailsTo(t, customer.customer.EmailAddress)
	require.Len(t, emails, 1)
	assert.Regexp(t, `^[0-9]{6}$`, emails[0].otp(t))
	assert.Contains(t, emails[0].Text, recipientBrandName)
	assert.Contains(t, emails[0].Text, "valid for 5 minutes")
	assert.Contains(t, emails[0].Text, "only for the sharing of your data")
	assert.Contains(t, emails[0].Text, recipientBrandName+" does not need your password",
		"the email says that the data recipient does not need the password of the consumer")
	assert.NotContains(t, otpPage.body, emails[0].otp(t), "the page does not show the OTP")
}

func TestTheCorrectOTPIsAcceptedOneTimeAndOnlyInItsFlow(t *testing.T) {
	client := registered(t, registrationTestProduct)

	t.Run("accepted", func(t *testing.T) {
		customer := newAccount(t)
		otpPage := customer.identify(client, client.requestObjectClaims())

		next := customer.browser.submit(otpPage, "continue", url.Values{"otp": {customer.otps(1)[0]}})

		require.Equal(t, http.StatusOK, next.status, next.body)
		assert.False(t, isOTPPage(next), "the response is not the OTP page")
		assert.Empty(t, next.header.Get("Location"), "the response is not a redirect")
	})

	t.Run("refused in a different flow", func(t *testing.T) {
		customer := newAccount(t)
		customer.identify(client, client.requestObjectClaims())
		first := customer.otps(1)[0]

		other := account{t: t, customer: customer.customer, browser: newConsumer(t)}
		otherPage := other.identify(client, client.requestObjectClaims())
		second := customer.otps(2)[1]
		require.NotEqual(t, first, second, "this test needs two different OTPs; run it again")

		refused := other.browser.submit(otherPage, "continue", url.Values{"otp": {first}})

		assert.True(t, isOTPPage(refused), refused.body)
		assert.Contains(t, refused.body, "not correct")
	})
}

func TestANewOTPMakesThePreviousOneNotValidAndTheFlowSends3AtMost(t *testing.T) {
	client := registered(t, registrationTestProduct)
	customer := newAccount(t)
	otpPage := customer.identify(client, client.requestObjectClaims())

	second := customer.browser.submit(otpPage, "resend", nil)
	require.True(t, isOTPPage(second), second.body)

	otps := customer.otps(2)
	require.NotEqual(t, otps[0], otps[1], "this test needs two different OTPs; run it again")

	old := customer.browser.submit(second, "continue", url.Values{"otp": {otps[0]}})
	assert.Contains(t, old.body, "not correct", "the first OTP is not valid after the second")

	third := customer.browser.submit(old, "resend", nil)
	customer.otps(3)
	assert.NotContains(t, third.body, `value="resend"`, "the page has no control for a fourth OTP")

	_, token := third.form(t)
	target, _ := third.form(t)
	fourth := customer.browser.post(target, url.Values{"csrf_token": {token}, "action": {"resend"}})
	require.True(t, isOTPPage(fourth), fourth.body)

	otps = customer.otps(3)
	accepted := customer.browser.submit(fourth, "continue", url.Values{"otp": {otps[2]}})
	assert.False(t, isOTPPage(accepted), "the third OTP is accepted: %s", accepted.body)
}

func TestTheEmailAddressInADifferentCaseGetsAnOTPByEmailAndAuthenticatesTheCustomer(t *testing.T) {
	client := registered(t, registrationTestProduct)
	customer := newEligibleAccount(t)

	otpPage := customer.identifyWith(
		client,
		client.requestObjectClaims(),
		strings.ToUpper(customer.customer.EmailAddress),
	)

	require.True(t, isOTPPage(otpPage), otpPage.body)
	assert.NotContains(t, strings.ToLower(otpPage.body), customer.customer.EmailAddress)

	consent := customer.browser.submit(otpPage, "continue", url.Values{"otp": {customer.otps(1)[0]}})

	require.Equal(t, http.StatusOK, consent.status, consent.body)
	assert.False(t, isOTPPage(consent), "the OTP authenticates the customer of the email address")
	assert.Contains(t, consent.body, "Everyday", "the consent page shows the account of the customer")
}

func TestAnIdentifierThatNoCustomerHasGetsTheSameOTPPageAndNoEmail(t *testing.T) {
	client := registered(t, registrationTestProduct)
	customer := newAccount(t)
	knownPage := customer.identify(client, client.requestObjectClaims())
	customer.otps(1)

	for name, identifier := range map[string]string{
		// The data holder creates only customer numbers of 8 digits.
		"customer number": "000000000",
		"email address":   "nobody." + strings.ToLower(rand.Text()) + "@mail.test",
	} {
		t.Run(name, func(t *testing.T) {
			before := emailCount(t)
			browser := newConsumer(t)
			first := browser.open(client.authorisationURL(t, client.requestObjectClaims()))
			unknownPage := browser.submit(first, "continue", url.Values{"identifier": {identifier}})

			require.Equal(t, http.StatusOK, unknownPage.status, unknownPage.body)
			assert.Equal(t, samePageText(knownPage.body), samePageText(unknownPage.body))
			assert.Equal(t, before, emailCount(t), "the data holder sent no email")

			refused := browser.submit(unknownPage, "continue", url.Values{"otp": {"000000"}})
			assert.True(t, isOTPPage(refused), refused.body)
		})
	}
}

// samePageText removes the two values of a page that are different for each flow.
func samePageText(body string) string {
	return formToken.ReplaceAllString(formAction.ReplaceAllString(body, ""), "")
}

func TestTheThirdIncorrectOTPGivesTheSoftwareProductAnAccessDeniedResponse(t *testing.T) {
	client := registered(t, registrationTestProduct)
	customer := newAccount(t)
	claims := client.requestObjectClaims()
	page := customer.identify(client, claims)
	// An entry of 5 digits is never the OTP.
	wrong := url.Values{"otp": {"00000"}}

	page = customer.browser.submit(page, "continue", wrong)
	require.True(t, isOTPPage(page), page.body)
	page = customer.browser.submit(page, "continue", wrong)
	require.True(t, isOTPPage(page), page.body)

	third := customer.browser.submit(page, "continue", wrong)

	client.requireAccessDenied(t, third, claims["state"])
}

func TestACancelAfterTheCustomerNumberGivesAnAccessDeniedResponse(t *testing.T) {
	client := registered(t, registrationTestProduct)

	t.Run("OTP page", func(t *testing.T) {
		customer := newAccount(t)
		claims := client.requestObjectClaims()
		otpPage := customer.identify(client, claims)

		client.requireAccessDenied(t, customer.browser.submit(otpPage, "cancel", nil), claims["state"])
	})

	t.Run("page after the correct OTP", func(t *testing.T) {
		customer := newAccount(t)
		claims := client.requestObjectClaims()
		otpPage := customer.identify(client, claims)
		next := customer.browser.submit(otpPage, "continue", url.Values{"otp": {customer.otps(1)[0]}})

		client.requireAccessDenied(t, customer.browser.submit(next, "cancel", nil), claims["state"])
	})
}
