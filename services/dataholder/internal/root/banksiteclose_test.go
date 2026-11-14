package root

import (
	"crypto/x509"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// All customer, account and arrangement data in this file is invented.

// DESIGN.md: Tests of the end of eligibility through the site.
func TestTheSiteCloseOfTheLastOpenAccountEndsEachConsentOfTheCustomerForTheEndOfEligibility(t *testing.T) {
	lab := newSiteLab(t)
	lab.consent(t, "arrangement-1", "account-1")
	lab.consent(t, "arrangement-2")

	cert := &x509.Certificate{Raw: []byte("client certificate of the software product")}
	tokens := map[string]string{
		"arrangement-1": lab.accessToken(t, "arrangement-1", cert),
		"arrangement-2": lab.accessToken(t, "arrangement-2", cert),
	}

	for arrangementID, token := range tokens {
		verified, err := lab.verifier.VerifyConsentToken(t.Context(), token, cert)
		require.NoError(t, err, "the token of %s is valid before the close", arrangementID)
		require.Equal(t, arrangementID, verified.ArrangementID)
	}

	body := lab.endsOfClose(t)
	assert.JSONEq(t, `{"ends": ["arrangement-1", "arrangement-2"]}`, body, "also the consent with no account")

	siteRequest(t, lab.site, http.MethodPost, "/api/accounts/account-1/close", body, http.StatusNoContent,
		lab.session...)

	accounts := lab.accounts(t)
	assert.Equal(t, "CLOSED", accounts[0].OpenStatus)
	assert.Empty(t, accounts[0].Actions)

	for arrangementID, token := range tokens {
		state := lab.state(t, arrangementID)
		assert.Equal(t, domain.EndReasonConsumerNotEligible, state.EndReason, arrangementID)
		assert.True(
			t,
			lab.now.Equal(state.EndedAt),
			"%s ends at the time of the close: %s",
			arrangementID,
			state.EndedAt,
		)
		assert.Equal(t, domain.NoticePending, state.NoticeStatus(), arrangementID)

		_, err := lab.verifier.VerifyConsentToken(t.Context(), token, cert)
		require.ErrorIs(t, err, authserver.ErrInvalidToken, arrangementID)
		assert.ErrorContains(t, err, "its grant is revoked", arrangementID)
	}
}

func TestASiteCloseAfterAConsentStartsAfterTheReadOfTheCloseActionChangesNothing(t *testing.T) {
	lab := newSiteLab(t)
	lab.consent(t, "arrangement-1", "account-1")

	body := lab.endsOfClose(t)
	lab.consent(t, "arrangement-2", "account-1")

	siteRequest(t, lab.site, http.MethodPost, "/api/accounts/account-1/close", body, http.StatusConflict,
		lab.session...)

	assert.Equal(t, "OPEN", lab.accounts(t)[0].OpenStatus)

	for _, arrangementID := range []string{"arrangement-1", "arrangement-2"} {
		assert.Empty(t, lab.state(t, arrangementID).EndReason, arrangementID)
	}
}

// A close that ignores the customer of the session fails this test.
func TestASiteCloseOfTheAccountOfADifferentCustomerIsNotFoundAndChangesNothing(t *testing.T) {
	const other = "00000002"

	lab := newSiteLab(t)
	require.NoError(t, lab.customers.AddCustomer(t.Context(), domain.Customer{
		CustomerNumber: other, LastName: "Bilby", EmailAddress: "bilby@example.test", LastUpdated: lab.now,
	}))
	require.NoError(t, lab.customers.AddAccount(t.Context(), other, domain.Account{
		AccountID: "account-of-other", DisplayName: "Saver", ProductCategory: domain.CategoryTransAndSavingsAccounts,
		ProductName: "Saver", BSB: openBSB, AccountNumber: "87654321", OpenStatus: domain.OpenStatusOpen,
		CreatedAt: lab.now,
	}))
	require.NoError(t, app.ArrangementType.Execute(t.Context(), lab.events, "arrangement-of-other",
		domain.CreateArrangement{Arrangement: domain.Arrangement{
			ArrangementID: "arrangement-of-other", CustomerNumber: other, ClientID: lab.clientID,
			SoftwareID: "software-1", ClientName: "Recipient Web", Scopes: []string{"bank:accounts.basic:read"},
			AccountIDs: []string{"account-of-other"}, ConfirmedAt: lab.now.Add(-time.Minute),
			SharingDurationSeconds: 3600,
		}}))

	siteRequest(t, lab.site, http.MethodPost, "/api/accounts/account-of-other/close",
		`{"ends": ["arrangement-of-other"]}`, http.StatusNotFound, lab.session...)

	record, err := lab.customers.Customer(t.Context(), other)
	require.NoError(t, err)
	require.Len(t, record.Accounts, 1)
	assert.Equal(t, domain.OpenStatusOpen, record.Accounts[0].OpenStatus)

	states, err := lab.reader.ArrangementsOf(t.Context(), other)
	require.NoError(t, err)
	require.Len(t, states, 1)
	assert.Empty(t, states[0].EndReason)
	assert.True(t, states[0].EndedAt.IsZero(), "the consent of the other customer did not end")
}
