package consumerpages_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// consent gives the consent page of the known customer.
func (b *browser) consent() page {
	b.t.Helper()
	b.identify(knownCustomer)

	consent := b.enter(b.mail.otps[0])
	require.Equal(b.t, http.StatusOK, consent.status, consent.body)

	return consent
}

// confirm sends the consent form with the selected accounts.
func (b *browser) confirm(accountIDs ...string) page {
	b.t.Helper()

	token := formToken(b.t, b.get().body)

	return b.post(url.Values{"csrf_token": {token}, "action": {"confirm"}, "account": accountIDs})
}

// Verifies: DH-0801, DH-0802
func TestTheConsentPageNamesTheLegalEntityTheBrandAndTheSoftwareProduct(t *testing.T) {
	b := newBrowser(t)

	consent := b.consent()

	assert.Contains(t, consent.body, "Budget Holdings Pty Ltd asks Coastal Bank to share your data.")
	assert.Contains(t, consent.body, "the brand name Budget &lt;App&gt; &amp; Co")
	assert.Contains(t, consent.body, "the application Budget Tracker")
}

// Verifies: DH-0801
func TestTheConsentPageNamesTheBrandWhenTheRegistrationHasNoLegalEntityName(t *testing.T) {
	b := newBrowser(t)
	b.registrations.legalEntityName = ""

	consent := b.consent()

	assert.Contains(t, consent.body, "Budget &lt;App&gt; &amp; Co asks Coastal Bank to share your data.")
}

// Verifies: DH-0803, DH-0804
func TestTheConsentPageHasTheDataClustersOfTheScopesOnly(t *testing.T) {
	b := newBrowser(t)
	b.scopes = []string{"openid", "bank:accounts.detail:read", "profile"}

	consent := b.consent()

	assert.Contains(t, consent.body, "<h3>Account balance and details</h3>")
	assert.Contains(t, consent.body, "<li>Account mail address</li>")
	assert.Contains(t, consent.body, "<h3>Name</h3>")
	assert.Contains(t, consent.body, "<li>Full name and title(s)</li>")
	assert.NotContains(t, consent.body, "Transaction details")
	assert.NotContains(t, consent.body, "Account numbers and features")
	assert.NotContains(t, consent.body, "openid")
}

// Verifies: DH-0811
func TestTheConsentPageOfASharingDurationOf24HoursOrLessSaysOneTime(t *testing.T) {
	for name, sharingDuration := range map[string]time.Duration{"none": 0, "24 hours": 24 * time.Hour} {
		t.Run(name, func(t *testing.T) {
			b := newBrowser(t)
			b.sharingDuration = sharingDuration

			consent := b.consent()

			assert.Contains(t, consent.body, "Coastal Bank will share your data one time.")
			assert.NotContains(t, consent.body, "until")
		})
	}
}

// Verifies: DH-0812, DH-0813
func TestTheConsentPageOfALongerSharingDurationHasTheDurationAndTheEndDate(t *testing.T) {
	// The time of the browser is 10 October 2026, 09:00 UTC.
	tests := map[string]struct {
		sharingDuration time.Duration
		want            string
	}{
		"25 hours": {
			sharingDuration: 25 * time.Hour,
			want:            "for 1 day and 1 hour, until 11 October 2026 (UTC).",
		},
		"90 days": {sharingDuration: 90 * 24 * time.Hour, want: "for 90 days, until 8 January 2027 (UTC)."},
		"one year": {
			sharingDuration: 365 * 24 * time.Hour,
			want:            "for 1 year, until 10 October 2027 (UTC).",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := newBrowser(t)
			b.sharingDuration = tt.sharingDuration

			consent := b.consent()

			assert.Contains(t, consent.body, tt.want)
			assert.NotContains(t, consent.body, "one time")
		})
	}
}

// Verifies: DH-0814
func TestTheConsentPageShowsTheEndDateInUTC(t *testing.T) {
	b := newBrowser(t)
	b.now = time.Date(2026, 10, 10, 23, 30, 0, 0, time.FixedZone("AWST", 8*60*60))
	b.sharingDuration = 48 * time.Hour

	consent := b.consent()

	// 23:30 on 10 October in Perth is 15:30 UTC on the same day.
	assert.Contains(t, consent.body, "until 12 October 2026 (UTC).")
}

// Verifies: DH-0815, DH-0816
func TestTheConsentPageSaysThatTheDataIncludesEarlierDataAndHowToStop(t *testing.T) {
	b := newBrowser(t)

	consent := b.consent()

	assert.Contains(t, consent.body, "The data includes data from before today")
	assert.Contains(
		t,
		consent.body,
		"You can stop the sharing at any time. Sign in to Coastal Bank online banking and go to Data sharing, "+
			"or ask Coastal Bank.",
	)
}

// Verifies: DH-0819, DH-0820
func TestTheConsentPageListsTheOpenAccountsOfTheCustomerWithMaskedNumbersAndNoneSelected(t *testing.T) {
	b := newBrowser(t)

	consent := b.consent()

	assert.Contains(t, consent.body, `name="account" value="account-1">`)
	assert.Contains(t, consent.body, "Everyday (xxx-xxx xxxx5678)")
	assert.Contains(t, consent.body, `name="account" value="account-2">`)
	assert.Contains(t, consent.body, "Holiday &lt;Saver&gt; (xxx-xxx xxxx4321)")
	assert.NotContains(t, consent.body, "12345678")
	assert.NotContains(t, consent.body, "checked")
	assert.NotContains(t, consent.body, "account-3", "a closed account")
}

// Verifies: DH-0822
func TestTheConsentPageOfARequestWithNoBankScopeHasNoAccountSelection(t *testing.T) {
	b := newBrowser(t)
	b.scopes = []string{"openid", "common:customer.basic:read"}

	consent := b.consent()

	assert.NotContains(t, consent.body, `name="account"`)
	assert.NotContains(t, consent.body, "Everyday")
	assert.Contains(t, consent.body, `value="confirm"`)
}

// Verifies: DH-0823
func TestACustomerWithNoOpenAccountCanOnlyCancel(t *testing.T) {
	for name, scopes := range map[string][]string{
		"a bank scope":    {"openid", "bank:accounts.basic:read"},
		"no bank scope":   {"openid", "profile"},
		"customer scopes": {"openid", "common:customer.basic:read"},
	} {
		t.Run(name, func(t *testing.T) {
			b := newBrowser(t)
			b.scopes = scopes
			b.customers.accounts = ownAccounts()[2:]

			consent := b.consent()

			assert.Contains(t, consent.body, "You have no open account at Coastal Bank, so you cannot share your data")
			assert.NotContains(t, consent.body, `value="confirm"`)
			assert.Contains(t, consent.body, `value="cancel"`)

			refused := b.confirm()
			assert.Equal(t, http.StatusBadRequest, refused.status)
			assert.NotContains(t, refused.body, `value="confirm"`)
			assert.Empty(t, b.arrangements.appends)
		})
	}
}

// Verifies: DH-0827
func TestAConfirmationCreatesAnArrangementAndApprovesTheFlow(t *testing.T) {
	b := newBrowser(t)
	b.sharingDuration = 90 * 24 * time.Hour
	b.consent()

	authenticatedAt := b.now
	b.now = b.now.Add(time.Minute)
	done := b.confirm("account-2")

	assert.Equal(t, authserver.Approve(authserver.Consent{
		ArrangementID:   "arrangement-1",
		Subject:         "subject-1",
		AuthenticatedAt: authenticatedAt,
		SharingEndsAt:   b.now.Add(90 * 24 * time.Hour),
	}), done.outcome)
	assert.Equal(t, [][2]string{{knownCustomer, "recipient.example"}}, b.subjects.calls)
	assert.Empty(t, done.body, "the authorisation server writes the redirect")
	require.Len(t, b.arrangements.appends, 1)

	created := b.arrangements.appends[0]
	assert.Equal(t, "arrangement", created.StreamType)
	assert.Equal(t, "arrangement-1", created.StreamID)
	require.Len(t, created.Events, 1)

	var payload domain.ArrangementCreated
	require.NoError(t, json.Unmarshal(created.Events[0].Payload, &payload))
	assert.Equal(t, domain.Arrangement{
		ArrangementID:          "arrangement-1",
		CustomerNumber:         knownCustomer,
		ClientID:               clientID,
		SoftwareID:             "software-1",
		ClientName:             "Budget Tracker",
		OrgName:                "Budget <App> & Co",
		LegalEntityName:        "Budget Holdings Pty Ltd",
		Scopes:                 []string{"openid", "bank:accounts.basic:read"},
		AccountIDs:             []string{"account-2"},
		ConfirmedAt:            b.now,
		SharingDurationSeconds: 7_776_000,
	}, payload.Arrangement)
}

func TestAConsumerWhoGaveTheEmailAddressConsentsAsTheCustomerOfTheAddress(t *testing.T) {
	b := newBrowser(t)
	b.identify("ALEX@example.com")
	consent := b.enter(b.mail.otps[0])
	require.Equal(t, http.StatusOK, consent.status, consent.body)
	assert.Contains(t, consent.body, "Everyday", "the page shows the accounts of the customer")

	b.confirm("account-2")

	assert.Equal(t, [][2]string{{knownCustomer, "recipient.example"}}, b.subjects.calls)
	require.Len(t, b.arrangements.appends, 1)

	var payload domain.ArrangementCreated
	require.NoError(t, json.Unmarshal(b.arrangements.appends[0].Events[0].Payload, &payload))
	assert.Equal(t, knownCustomer, payload.Arrangement.CustomerNumber)
}

func TestAnArrangementOfARegistrationWithNoLegalEntityNameHasOnlyTheBrandName(t *testing.T) {
	b := newBrowser(t)
	b.registrations.legalEntityName = ""
	b.consent()

	b.confirm("account-2")

	require.Len(t, b.arrangements.appends, 1)

	var payload domain.ArrangementCreated
	require.NoError(t, json.Unmarshal(b.arrangements.appends[0].Events[0].Payload, &payload))
	assert.Equal(t, "Budget <App> & Co", payload.Arrangement.OrgName)
	assert.Empty(t, payload.Arrangement.LegalEntityName)
}

func TestAConfirmationOfARequestWithNoBankScopeNeedsNoAccount(t *testing.T) {
	b := newBrowser(t)
	b.scopes = []string{"openid", "profile"}
	b.consent()

	done := b.confirm()

	assert.Equal(t, b.approval(), done.outcome)
}

// approval is the outcome of a consent with no sharing duration at the time of the browser.
func (b *browser) approval() authserver.Outcome {
	return authserver.Approve(authserver.Consent{
		ArrangementID:   "arrangement-1",
		Subject:         "subject-1",
		AuthenticatedAt: b.now,
		SharingEndsAt:   b.now.Add(b.sharingDuration),
	})
}

// Verifies: DH-0826
func TestAConfirmationThatBreaksAnAccountRuleGivesTheConsentPageAgain(t *testing.T) {
	tests := map[string]struct {
		accountIDs []string
		notice     string
	}{
		"no account": {accountIDs: nil, notice: "Select one account or more to share your data."},
		"account of a different customer": {
			accountIDs: []string{"account-1", "account-9"},
			notice:     "Coastal Bank cannot share an account that you selected.",
		},
		"closed account": {
			accountIDs: []string{"account-1", "account-3"},
			notice:     "Coastal Bank cannot share an account that you selected.",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := newBrowser(t)
			b.consent()

			refused := b.confirm(tt.accountIDs...)

			assert.Equal(t, http.StatusBadRequest, refused.status)
			assert.Contains(t, refused.body, tt.notice)
			assert.Contains(t, refused.body, `value="confirm"`)
			assert.Empty(t, b.arrangements.appends)

			_, continues := refused.outcome.Continues()
			require.True(t, continues)
			assert.Equal(t, b.approval(), b.confirm("account-1").outcome)
		})
	}
}

func TestAnActionThatTheConsentPageDoesNotHaveIsABadRequest(t *testing.T) {
	b := newBrowser(t)
	b.consent()

	refused := b.act("resend")

	assert.Equal(t, http.StatusBadRequest, refused.status)
	assert.Empty(t, b.arrangements.appends)
	assert.Len(t, b.mail.otps, 1)
}

// Verifies: DH-0845
func TestAConsentFormWithoutItsProtectionAgainstCrossSiteRequestForgeryCreatesNoArrangement(t *testing.T) {
	b := newBrowser(t)
	b.consent()

	refused := b.post(url.Values{"csrf_token": {"a-different-token"}, "action": {"confirm"}, "account": {"account-1"}})

	assert.Equal(t, http.StatusForbidden, refused.status)
	assert.Empty(t, b.arrangements.appends)
}

func TestAReloadOfTheConsentPageShowsItAgain(t *testing.T) {
	b := newBrowser(t)
	first := b.consent()

	again := b.get()

	assert.Equal(t, http.StatusOK, again.status)
	assert.Equal(t, first.body, again.body)
}

func TestAConsentThatTheDataHolderCannotRecordIsTemporarilyUnavailable(t *testing.T) {
	t.Run("the accounts cannot be read", func(t *testing.T) {
		b := newBrowser(t)
		b.identify(knownCustomer)
		otp := b.mail.otps[0]
		b.customers.err = errors.New("connection lost")

		assert.Equal(t, authserver.Deny(authserver.TemporarilyUnavailable), b.enter(otp).outcome)
	})

	t.Run("the arrangement cannot be stored", func(t *testing.T) {
		b := newBrowser(t)
		b.consent()
		b.arrangements.err = errors.New("connection lost")

		assert.Equal(t, authserver.Deny(authserver.TemporarilyUnavailable), b.confirm("account-1").outcome)
	})
}
