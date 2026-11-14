package root

import (
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/operatorhttp"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/operatorapi"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// All customer, account and arrangement data in this file is invented.

func withdrawTarget(arrangementID string) string {
	return "/api/consents/" + arrangementID + "/withdraw"
}

func (l siteLab) operatorWithdrawalMethods(t *testing.T) map[string]string {
	t.Helper()

	const token = "operator-token"

	mux := http.NewServeMux()
	operatorapi.Register(mux, app.Customers{}, l.arrangements,
		operatorhttp.Config{Token: token, AllowedHosts: []string{"example.com"}})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet,
		"http://example.com/v1/customers/"+siteCustomer+"/arrangements", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var arrangements []struct {
		ArrangementID    string `json:"arrangementId"`
		WithdrawalMethod string `json:"withdrawalMethod"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &arrangements))

	methods := map[string]string{}
	for _, a := range arrangements {
		methods[a.ArrangementID] = a.WithdrawalMethod
	}

	return methods
}

func TestASiteWithdrawalEndsOnlyTheConsentOfItsPath(t *testing.T) {
	lab := newSiteLab(t)
	lab.consent(t, "arrangement-1", "account-1")
	lab.consent(t, "arrangement-2", "account-1")

	cert := &x509.Certificate{Raw: []byte("client certificate of the software product")}
	otherToken := lab.accessToken(t, "arrangement-2", cert)

	siteRequest(t, lab.site, http.MethodPost, withdrawTarget("arrangement-1"), "", http.StatusNoContent,
		lab.session...)

	assert.Equal(t, domain.EndReasonWithdrawnByConsumer, lab.state(t, "arrangement-1").EndReason)
	assert.Empty(t, lab.state(t, "arrangement-2").EndReason)
	assert.Equal(t, map[string]string{"arrangement-1": "CONSUMER_DASHBOARD", "arrangement-2": ""},
		lab.operatorWithdrawalMethods(t))

	_, err := lab.verifier.VerifyConsentToken(t.Context(), otherToken, cert)
	require.NoError(t, err, "the token of the other consent stays valid")
}

// A withdrawal that ignores the customer of the session fails this test.
func TestASiteWithdrawalOfTheConsentOfADifferentCustomerIsNotFoundAndChangesNothing(t *testing.T) {
	const other = "00000002"

	lab := newSiteLab(t)
	require.NoError(t, lab.customers.AddCustomer(t.Context(), domain.Customer{
		CustomerNumber: other, LastName: "Bilby", EmailAddress: "bilby@example.test", LastUpdated: lab.now,
	}))
	require.NoError(t, app.ArrangementType.Execute(t.Context(), lab.events, "arrangement-of-other",
		domain.CreateArrangement{Arrangement: domain.Arrangement{
			ArrangementID: "arrangement-of-other", CustomerNumber: other, ClientID: lab.clientID,
			SoftwareID: "software-1", ClientName: "Recipient Web", Scopes: []string{"bank:accounts.basic:read"},
			AccountIDs: []string{}, ConfirmedAt: lab.now, SharingDurationSeconds: 3600,
		}}))

	for _, arrangementID := range []string{"arrangement-of-other", "arrangement-unknown"} {
		siteRequest(t, lab.site, http.MethodPost, withdrawTarget(arrangementID), "", http.StatusNotFound,
			lab.session...)
	}

	states, err := lab.reader.ArrangementsOf(t.Context(), other)
	require.NoError(t, err)
	require.Len(t, states, 1)
	assert.Empty(t, states[0].EndReason)
	assert.True(t, states[0].EndedAt.IsZero(), "the consent of the other customer did not end")
}
