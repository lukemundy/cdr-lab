package banksite_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/banksite"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// consents gives the consents of the customer signedIn.
type consents struct {
	list []app.Consent
	err  error
}

func (c *consents) Consents(_ context.Context, customerNumber string) ([]app.Consent, error) {
	if customerNumber != signedIn {
		return nil, app.ErrCustomerNotFound
	}

	return c.list, c.err
}

func signedInConsents(t *testing.T, reader *consents) response {
	t.Helper()

	s := newSiteWith(t, banksite.Ports{
		Customers: newCustomers(domain.Customer{CustomerNumber: signedIn, LastName: "Quoll"}),
		Consents:  reader,
	})
	cookie := sessionCookie(t, s.signIn(signedIn))

	return s.do(request{
		method: http.MethodGet, target: "/api/consents",
		cookies: []*http.Cookie{requestCookie(cookieName, cookie.Value)},
	})
}

func budgetConsent() app.Consent {
	state := arrangement("arrangement-1", "Budget App")
	state.Arrangement.OrgName, state.Arrangement.LegalEntityName = "Budget Co", "Budget Holdings Pty Ltd"
	state.Arrangement.Scopes = []string{"openid", "bank:accounts.basic:read", "profile"}
	clusters := domain.DataClusters(state.Arrangement.Scopes)
	first, last := now.Add(-50*time.Minute), now.Add(-10*time.Minute)

	return app.Consent{
		ArrangementState: state,
		Accounts:         []domain.Account{everyday().Account},
		Clusters: []app.ClusterDisclosures{
			{DataCluster: clusters[0], Disclosed: &app.DisclosurePeriod{First: first, Last: last}},
			{DataCluster: clusters[1]},
		},
	}
}

// Verifies: DH-1766, DH-1767
func TestTheConsentsGiveEachArrangementWithItsDataRecipientAccountsAndDisclosures(t *testing.T) {
	resp := signedInConsents(t, &consents{list: []app.Consent{budgetConsent()}})

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{"consents": [{
		"arrangementId": "arrangement-1", "clientName": "Budget App", "brandName": "Budget Co",
		"legalEntityName": "Budget Holdings Pty Ltd", "status": "CURRENT",
		"confirmedAt": "2026-11-01T23:00:00Z", "sharingDurationSeconds": 86400,
		"sharingEndsAt": "2026-11-02T23:00:00Z",
		"accounts": [{
			"accountId": "account-1", "creationDate": "2026-10-01", "displayName": "Everyday",
			"maskedNumber": "xxx-xxx xxxx5678", "openStatus": "OPEN",
			"productCategory": "TRANS_AND_SAVINGS_ACCOUNTS", "productName": "Saver"
		}],
		"dataClusters": [
			{
				"name": "Account name, type and balance",
				"permissions": ["Name of account", "Type of account", "Account balance"],
				"firstDisclosedAt": "2026-11-01T23:10:00Z", "lastDisclosedAt": "2026-11-01T23:50:00Z"
			},
			{"name": "Name", "permissions": ["Full name and title(s)"]}
		]
	}]}`, resp.body)
}

// Verifies: DH-1766
func TestAnEndedConsentHasTheTimeOfItsEnd(t *testing.T) {
	withdrawn := budgetConsent()
	withdrawn.Arrangement.ArrangementID = "arrangement-2"
	withdrawn.EndedAt, withdrawn.EndReason = now.Add(-time.Minute), domain.EndReasonWithdrawnByConsumer
	expired := budgetConsent()
	expired.Arrangement.ArrangementID, expired.Arrangement.SharingDurationSeconds = "arrangement-3", 0
	expired.Arrangement.LegalEntityName = ""

	resp := signedInConsents(t, &consents{list: []app.Consent{withdrawn, expired}})

	require.Equal(t, http.StatusOK, resp.status)
	assert.Contains(t, resp.body, `"arrangementId":"arrangement-2","clientName":"Budget App","brandName":"Budget Co",`+
		`"legalEntityName":"Budget Holdings Pty Ltd","status":"ENDED"`)
	assert.Contains(t, resp.body, `"endedAt":"2026-11-01T23:59:00Z"`, "the time of the withdrawal")
	assert.Contains(t, resp.body, `"arrangementId":"arrangement-3","clientName":"Budget App","brandName":"Budget Co",`+
		`"status":"ENDED"`, "no legal entity name")
	assert.Contains(t, resp.body, `"endedAt":"2026-11-01T23:00:00Z"`, "the end of the sharing")
}

func TestACustomerWithNoConsentHasAnEmptyList(t *testing.T) {
	resp := signedInConsents(t, &consents{list: []app.Consent{}})

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{"consents": []}`, resp.body)
}

// Verifies: DH-1743
func TestTheConsentsNeedASession(t *testing.T) {
	s := newSiteWith(t, banksite.Ports{
		Customers: newCustomers(), Consents: &consents{list: []app.Consent{budgetConsent()}},
	})

	resp := s.do(request{method: http.MethodGet, target: "/api/consents"})

	assert.Equal(t, http.StatusForbidden, resp.status)
	assert.NotContains(t, resp.body, "arrangement-1")
}

func TestAFailureOfTheConsentsIsAServerError(t *testing.T) {
	resp := signedInConsents(t, &consents{err: errors.New("no database")})

	assert.Equal(t, http.StatusInternalServerError, resp.status)
	assert.NotContains(t, resp.body, "no database")
}

// Verifies: DH-1744
func TestTheConsentsOfACustomerThatTheDataHolderNoLongerHasEndTheSession(t *testing.T) {
	const gone = "00000002"

	s := newSiteWith(t, banksite.Ports{
		Customers: newCustomers(domain.Customer{CustomerNumber: gone, LastName: "Quoll"}),
		Consents:  &consents{},
	})
	cookie := requestCookie(cookieName, sessionCookie(t, s.signIn(gone)).Value)

	resp := s.do(request{method: http.MethodGet, target: "/api/consents", cookies: []*http.Cookie{cookie}})

	require.Equal(t, http.StatusForbidden, resp.status)
	assert.Contains(t, resp.body, `"type":"`+sessionEndedType+`"`)
	assertExpired(t, resp)
}
