//go:build conformance

package dataholder

import (
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These values of services/register/seed/example.json are those of the Registration Test software products.
const (
	// allScopesProduct has each data scope that the data holder supports.
	allScopesProduct         = "0b3d5f71-8e2a-4c4d-a6f8-6a7b8c9d0eb5"
	allScopesProductName     = "Registration Test All Scopes"
	recipientLegalEntityName = "Registration Test Pty Ltd"
)

const (
	oneOccasionText = " will share your data one time."
	dateFormat      = "2 January 2006"
)

// The data cluster language and the permission language of the CDS, word for word.
var (
	clusterCustomer = cluster{name: "Name and occupation", permissions: []string{"Name", "Occupation"}}
	clusterBasic    = cluster{
		name:        "Account name, type and balance",
		permissions: []string{"Name of account", "Type of account", "Account balance"},
	}
	clusterDetail = cluster{
		name: "Account numbers and features",
		permissions: []string{
			"Account number", "Interest rates", "Fees", "Discounts", "Account terms", "Account mail address",
		},
	}
	clusterDetailOnly = cluster{
		name: "Account balance and details",
		permissions: []string{
			"Name of account", "Type of account", "Account balance", "Account number", "Interest rates", "Fees",
			"Discounts", "Account terms", "Account mail address",
		},
	}
	clusterTransactions = cluster{
		name: "Transaction details",
		permissions: []string{
			"Incoming and outgoing transactions", "Amounts", "Dates", "Descriptions of transactions",
			"Who you have sent money to and received money from",
		},
	}
	clusterProfile = cluster{name: "Name", permissions: []string{"Full name and title(s)"}}
)

type cluster struct {
	name        string
	permissions []string
}

var (
	clusterBlock   = regexp.MustCompile(`(?s)<h3>([^<]+)</h3>\s*<ul>(.*?)</ul>`)
	permissionItem = regexp.MustCompile(`<li>([^<]+)</li>`)
	accountChoice  = regexp.MustCompile(
		`<input type="checkbox" id="[^"]+" name="account" value="([^"]+)"> <label for="[^"]+">([^<]+)</label>`,
	)
)

// clustersOf gives the data clusters of a consent page, in the sequence of the page.
func clustersOf(p webPage) []cluster {
	clusters := []cluster{}

	for _, block := range clusterBlock.FindAllStringSubmatch(p.body, -1) {
		found := cluster{name: html.UnescapeString(block[1]), permissions: []string{}}
		for _, item := range permissionItem.FindAllStringSubmatch(block[2], -1) {
			found.permissions = append(found.permissions, html.UnescapeString(item[1]))
		}

		clusters = append(clusters, found)
	}

	return clusters
}

// choicesOf gives the label of each account that a consent page lets the consumer select, by its form value.
func choicesOf(p webPage) map[string]string {
	choices := map[string]string{}
	for _, choice := range accountChoice.FindAllStringSubmatch(p.body, -1) {
		choices[choice[1]] = html.UnescapeString(choice[2])
	}

	return choices
}

func text(p webPage) string {
	return html.UnescapeString(p.body)
}

var consentHeading = regexp.MustCompile(`<h2>Data that [^<]+ will share</h2>`)

func isConsentPage(p webPage) bool {
	return consentHeading.MatchString(p.body)
}

func canConfirm(p webPage) bool {
	return strings.Contains(p.body, `value="confirm"`)
}

// claimsFor gives the claims of a request object with the scopes, and with the sharing duration when it is not nil.
func (c registeredClient) claimsFor(scope string, sharingDuration any) map[string]any {
	claims := c.requestObjectClaims()
	claims["scope"] = scope

	if sharingDuration != nil {
		claims["claims"] = map[string]any{"sharing_duration": sharingDuration}
	}

	return claims
}

// addAccount creates an account of the customer with the operator interface.
func (a account) addAccount(displayName, accountNumber string) operatorAccount {
	a.t.Helper()

	reply := operate(a.t, http.MethodPost, accountsPath(a.customer.CustomerNumber), operatorAccount{
		DisplayName:     displayName,
		ProductCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
		ProductName:     "Everyday Account",
		BSB:             "372100",
		AccountNumber:   accountNumber,
	})
	require.Equal(a.t, http.StatusCreated, reply.status, string(reply.body))

	var created operatorAccount
	reply.decode(a.t, &created)

	return created
}

// authenticate starts a flow, gives the customer number and the OTP, and gives the consent page.
func (a account) authenticate(client registeredClient, claims map[string]any) webPage {
	a.t.Helper()

	otpPage := a.identify(client, claims)
	otps := a.otps(1)

	consent := a.browser.submit(otpPage, "continue", url.Values{"otp": {otps[0]}})
	require.Equal(a.t, http.StatusOK, consent.status, consent.body)
	require.True(a.t, isConsentPage(consent), consent.body)

	return consent
}

func (a account) confirm(consent webPage, accountIDs ...string) webPage {
	a.t.Helper()

	return a.browser.submit(consent, "confirm", url.Values{"account": accountIDs})
}

// Verifies: DH-0801, DH-0802
func TestTheConsentPageNamesTheLegalEntityTheBrandAndTheSoftwareProduct(t *testing.T) {
	client := registered(t, allScopesProduct)

	consent := newAccount(t).authenticate(client, client.requestObjectClaims())

	assert.Contains(t, text(consent), recipientLegalEntityName+" asks "+bankName(t)+" to share your data.")
	assert.Contains(t, text(consent), "the brand name "+recipientBrandName+" and")
	assert.Contains(t, text(consent), "the application "+allScopesProductName+".")
}

// Verifies: DH-0803, DH-0804
func TestTheConsentPageHasTheDataLanguageOfTheCDSForTheScopesOfTheRequestOnly(t *testing.T) {
	client := registered(t, allScopesProduct)

	tests := map[string]struct {
		scope string
		want  []cluster
	}{
		"openid only":                {scope: "openid", want: []cluster{}},
		"common:customer.basic:read": {scope: "openid common:customer.basic:read", want: []cluster{clusterCustomer}},
		"bank:accounts.basic:read":   {scope: "openid bank:accounts.basic:read", want: []cluster{clusterBasic}},
		"bank:accounts.detail:read with bank:accounts.basic:read": {
			scope: "openid bank:accounts.detail:read bank:accounts.basic:read",
			want:  []cluster{clusterBasic, clusterDetail},
		},
		"bank:accounts.detail:read without bank:accounts.basic:read": {
			scope: "openid bank:accounts.detail:read",
			want:  []cluster{clusterDetailOnly},
		},
		"bank:transactions:read": {scope: "openid bank:transactions:read", want: []cluster{clusterTransactions}},
		"profile":                {scope: "openid profile", want: []cluster{clusterProfile}},
		"each scope": {
			scope: "openid profile common:customer.basic:read bank:accounts.basic:read bank:accounts.detail:read" +
				" bank:transactions:read",
			want: []cluster{clusterCustomer, clusterBasic, clusterDetail, clusterTransactions, clusterProfile},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			consent := newAccount(t).authenticate(client, client.claimsFor(tt.scope, nil))

			assert.Equal(t, tt.want, clustersOf(consent))
			assert.NotContains(t, consent.body, "openid")
		})
	}
}

// Verifies: DH-0811
func TestASharingDurationOf24HoursOrLessIsADisclosureOnOneOccasion(t *testing.T) {
	client := registered(t, allScopesProduct)

	for name, sharingDuration := range map[string]any{"no sharing_duration": nil, "0": 0, "24 hours": 86_400} {
		t.Run(name, func(t *testing.T) {
			consent := newAccount(t).authenticate(client, client.claimsFor("openid profile", sharingDuration))

			assert.Contains(t, text(consent), bankName(t)+oneOccasionText)
			assert.NotContains(t, consent.body, "until")
		})
	}
}

// Verifies: DH-0812
func TestALongerSharingDurationShowsTheDurationAndTheDateOnWhichTheSharingEnds(t *testing.T) {
	client := registered(t, allScopesProduct)

	const year = 365 * 24 * time.Hour

	tests := map[string]struct {
		sharingDuration int
		period          time.Duration
		duration        string
	}{
		"more than 24 hours": {sharingDuration: 90_000, period: 25 * time.Hour, duration: "1 day and 1 hour"},
		"90 days":            {sharingDuration: 7_776_000, period: 90 * 24 * time.Hour, duration: "90 days"},
		"one year":           {sharingDuration: 31_536_000, period: year, duration: "1 year"},
		"more than one year": {sharingDuration: 99_999_999, period: year, duration: "1 year"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			before := time.Now().UTC()

			consent := newAccount(t).authenticate(client, client.claimsFor("openid profile", tt.sharingDuration))

			assert.NotContains(t, text(consent), oneOccasionText)
			assert.Contains(t, consent.body, "for "+tt.duration+", until ")

			// The day can change between the start of the test and the page.
			ends := []string{
				"until " + before.Add(tt.period).Format(dateFormat) + " (UTC).",
				"until " + time.Now().UTC().Add(tt.period).Format(dateFormat) + " (UTC).",
			}
			assert.True(
				t,
				strings.Contains(consent.body, ends[0]) || strings.Contains(consent.body, ends[1]),
				"the page has the date on which the sharing ends, one of %q: %s", ends, consent.body,
			)
		})
	}
}

// Verifies: DH-0815, DH-0816
func TestTheConsentPageSaysThatTheDataIncludesEarlierDataAndHowToWithdraw(t *testing.T) {
	client := registered(t, allScopesProduct)

	consent := newAccount(t).authenticate(client, client.requestObjectClaims())

	assert.Contains(t, consent.body, "The data includes data from before today")
	assert.Contains(t, text(consent), "You can stop the sharing at any time. Sign in to "+bankName(t)+
		" online banking and go to Data sharing, or ask "+bankName(t)+".")
	assert.NotContains(t, consent.body, "<a ", "the page refers to no other document")
	assert.NotContains(t, consent.body, "<script")
}

// Verifies: DH-0819, DH-0820
func TestTheConsentPageListsOnlyTheAccountsOfTheCustomerWithMaskedNumbersAndNoneSelected(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer, other := newAccount(t), newAccount(t)
	everyday := customer.addAccount("Everyday", "20000001")
	saver := customer.addAccount("Holiday Saver", "2000000002")
	otherAccount := other.addAccount("Other Everyday", "20000003")

	consent := customer.authenticate(client, client.requestObjectClaims())

	assert.Equal(t, map[string]string{
		everyday.AccountID: "Everyday (xxx-xxx xxxx0001)",
		saver.AccountID:    "Holiday Saver (xxx-xxx xxxxxx0002)",
	}, choicesOf(consent))
	assert.NotContains(t, consent.body, otherAccount.AccountID)
	assert.NotContains(t, consent.body, "20000001", "the page does not show a full account number")
	assert.NotContains(t, consent.body, "checked", "no account is selected at the start")
	assert.True(t, canConfirm(consent))
}

// Verifies: DH-0826
func TestAConfirmationThatBreaksAnAccountRuleIsRefusedAndTheFlowContinues(t *testing.T) {
	client := registered(t, allScopesProduct)
	other := newAccount(t)
	otherAccount := other.addAccount("Other Everyday", "20000004")

	// selected has the accounts of the form, in addition to the account of the customer when withOwn is true.
	tests := map[string]struct {
		withOwn  bool
		selected []string
	}{
		"no account":                      {},
		"account of a different customer": {withOwn: true, selected: []string{otherAccount.AccountID}},
		"account that does not exist":     {selected: []string{uuid.NewString()}},
	}

	number := 20000010

	for name, tt := range tests {
		number++

		t.Run(name, func(t *testing.T) {
			customer := newAccount(t)
			own := customer.addAccount("Everyday", strconv.Itoa(number))
			claims := client.requestObjectClaims()
			consent := customer.authenticate(client, claims)

			selected := tt.selected
			if tt.withOwn {
				selected = append(selected, own.AccountID)
			}

			refused := customer.confirm(consent, selected...)

			assert.Equal(t, http.StatusBadRequest, refused.status, refused.body)
			assert.Empty(t, refused.header.Get("Location"), "the data holder does not redirect the browser")
			require.True(t, isConsentPage(refused), refused.body)
			assert.Contains(t, refused.body, `role="alert"`)

			confirmed := customer.confirm(refused, own.AccountID)
			assert.NotEmpty(t, client.responseClaims(t, confirmed)["code"])
		})
	}
}

// Verifies: DH-0822
func TestARequestWithNoBankScopeHasNoAccountSelection(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newAccount(t)
	customer.addAccount("Everyday", "20000020")

	claims := client.claimsFor("openid profile common:customer.basic:read", nil)

	consent := customer.authenticate(client, claims)

	assert.Empty(t, choicesOf(consent))
	assert.NotContains(t, consent.body, `name="account"`)
	assert.NotContains(t, consent.body, "Everyday")

	confirmed := customer.confirm(consent)
	assert.NotEmpty(t, client.responseClaims(t, confirmed)["code"])
}

// Verifies: DH-0823
func TestACustomerWithNoOpenAccountCanOnlyCancel(t *testing.T) {
	client := registered(t, allScopesProduct)

	for name, scope := range map[string]string{"a bank scope": accountScope, "no bank scope": noAccountScope} {
		t.Run(name, func(t *testing.T) {
			customer := newAccount(t)
			claims := client.claimsFor(scope, ninetyDays)

			consent := customer.authenticate(client, claims)

			assert.Contains(
				t,
				text(consent),
				"You have no open account at "+bankName(t)+", so you cannot share your data",
			)
			assert.False(t, canConfirm(consent), consent.body)

			refused := customer.confirm(consent)
			assert.Equal(t, http.StatusBadRequest, refused.status, refused.body)
			assert.Empty(t, refused.header.Get("Location"))

			client.requireAccessDenied(t, customer.browser.submit(consent, "cancel", nil), claims["state"])
		})
	}
}

// Verifies: DH-0825
func TestACancelOnTheConsentPageGivesTheSoftwareProductASignedAccessDeniedResponse(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newAccount(t)
	customer.addAccount("Everyday", "20000030")

	claims := client.requestObjectClaims()
	consent := customer.authenticate(client, claims)

	cancelled := customer.browser.submit(consent, "cancel", nil)

	client.requireAccessDenied(t, cancelled, claims["state"])
}

func TestAConfirmationGivesTheSoftwareProductASignedResponseWithACode(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newAccount(t)
	own := customer.addAccount("Everyday", "20000040")
	claims := client.claimsFor("openid bank:accounts.basic:read bank:transactions:read", 7_776_000)
	consent := customer.authenticate(client, claims)

	confirmed := customer.confirm(consent, own.AccountID)

	// responseClaims checks the redirect URI, the one query parameter, PS256 and the signature.
	response := client.responseClaims(t, confirmed)
	assert.Equal(t, client.issuer, response["iss"])
	assert.Equal(t, client.clientID, response["aud"])
	assert.Equal(t, claims["state"], response["state"])
	assert.NotContains(t, response, "error")
	assert.NotContains(t, response, "cdr_arrangement_id")

	require.IsType(t, "", response["code"])
	assert.NotEmpty(t, response["code"])

	require.IsType(t, float64(0), response["exp"])
	expires := time.Unix(int64(response["exp"].(float64)), 0)
	assert.WithinDuration(t, time.Now().Add(10*time.Minute), expires, time.Minute)

	t.Run("the flow does not continue", func(t *testing.T) {
		again := customer.confirm(consent, own.AccountID)

		requireRefusedPage(t, again, http.StatusBadRequest)
	})
}

func TestTwoConfirmationsOfOneCustomerGiveDifferentCodes(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newAccount(t)
	own := customer.addAccount("Everyday", "20000050")

	first := customer.confirm(customer.authenticate(client, client.requestObjectClaims()), own.AccountID)

	second := account{t: t, customer: customer.customer, browser: newConsumer(t)}
	otpPage := second.identify(client, client.requestObjectClaims())
	consent := second.browser.submit(otpPage, "continue", url.Values{"otp": {customer.otps(2)[1]}})
	require.True(t, isConsentPage(consent), consent.body)

	firstCode := client.responseClaims(t, first)["code"]
	secondCode := client.responseClaims(t, second.confirm(consent, own.AccountID))["code"]

	assert.NotEmpty(t, firstCode)
	assert.NotEqual(t, firstCode, secondCode)
}

// Verifies: DH-0845
func TestAConsentFormWithoutItsProtectionAgainstCrossSiteRequestForgeryIsRefused(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newAccount(t)
	own := customer.addAccount("Everyday", "20000060")
	consent := customer.authenticate(client, client.requestObjectClaims())
	target, token := consent.form(t)

	t.Run("no form token", func(t *testing.T) {
		refused := customer.browser.post(target, url.Values{"action": {"confirm"}, "account": {own.AccountID}})

		assert.Equal(t, http.StatusForbidden, refused.status, refused.body)
		assert.Empty(t, refused.header.Get("Location"))
	})

	t.Run("form from a different browser", func(t *testing.T) {
		form := url.Values{"csrf_token": {token}, "action": {"confirm"}, "account": {own.AccountID}}

		refused := newConsumer(t).post(target, form)

		assert.Equal(t, http.StatusForbidden, refused.status, refused.body)
		assert.Empty(t, refused.header.Get("Location"))
	})

	t.Run("the flow continues for the browser that started it", func(t *testing.T) {
		confirmed := customer.confirm(consent, own.AccountID)

		assert.NotEmpty(t, client.responseClaims(t, confirmed)["code"])
	})
}
