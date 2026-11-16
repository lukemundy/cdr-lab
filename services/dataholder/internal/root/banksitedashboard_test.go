package root

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// All customer, account and arrangement data in this file is invented.

type siteDataCluster struct {
	Name             string     `json:"name"`
	FirstDisclosedAt *time.Time `json:"firstDisclosedAt"`
	LastDisclosedAt  *time.Time `json:"lastDisclosedAt"`
}

type siteConsent struct {
	ArrangementID   string    `json:"arrangementId"`
	ClientName      string    `json:"clientName"`
	BrandName       string    `json:"brandName"`
	LegalEntityName string    `json:"legalEntityName"`
	Status          string    `json:"status"`
	SharingEndsAt   time.Time `json:"sharingEndsAt"`
	Accounts        []struct {
		AccountID string `json:"accountId"`
	} `json:"accounts"`
	DataClusters []siteDataCluster `json:"dataClusters"`
}

func (l siteLab) consents(t *testing.T) []siteConsent {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "https://example.com/api/consents", nil)
	for _, cookie := range l.session {
		req.AddCookie(cookie)
	}

	rec := httptest.NewRecorder()
	l.site.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body struct {
		Consents []siteConsent `json:"consents"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	return body.Consents
}

// getAccounts calls Get Accounts with the access token, on a connection with the client certificate.
func (l siteLab) getAccounts(t *testing.T, token string, cert *x509.Certificate) int {
	t.Helper()

	req := httptest.NewRequestWithContext(
		t.Context(), http.MethodGet, "https://mtls.bank.example/cds-au/v1/banking/accounts", nil,
	)
	req.TLS = &tls.ConnectionState{PeerCertificates: []*x509.Certificate{cert}}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("x-v", "3")
	req.Header.Set("x-fapi-auth-date", "Mon, 02 Nov 2026 08:00:00 GMT")

	rec := httptest.NewRecorder()
	l.consentAPI.ServeHTTP(rec, req)

	return rec.Code
}

// A refresh of a revoked grant gives invalid_grant: pkg/authserver,
// TestARevocationOfAnArrangementOfTheClientGivesNoContentAndEndsEachToken.
// Verifies: DH-1919, DH-1930
func TestTheDashboardShowsTheDisclosuresOfAConsentAndAWithdrawalOnTheDashboardStopsItsTokens(t *testing.T) {
	lab := newSiteLab(t)
	lab.consent(t, "arrangement-1", "account-1")

	before := lab.consents(t)
	require.Len(t, before, 1)
	require.Len(t, before[0].DataClusters, 1)
	assert.Nil(t, before[0].DataClusters[0].FirstDisclosedAt, "no disclosure before the call")

	cert := &x509.Certificate{Raw: []byte("client certificate of the software product")}
	token := lab.accessToken(t, "arrangement-1", cert)
	require.Equal(t, http.StatusOK, lab.getAccounts(t, token, cert))
	lab.advance(2 * time.Hour)
	require.Equal(t, http.StatusOK, lab.getAccounts(t, token, cert))
	require.Equal(t, http.StatusUnauthorized, lab.getAccounts(t, "not a token", cert))

	got := lab.consents(t)
	require.Len(t, got, 1)
	assert.Equal(t, "arrangement-1", got[0].ArrangementID)
	assert.Equal(t, "Recipient Web", got[0].ClientName)
	assert.Equal(t, "Recipient Brand", got[0].BrandName)
	assert.Equal(t, "Recipient Pty Ltd", got[0].LegalEntityName)
	assert.Equal(t, "CURRENT", got[0].Status)
	assert.True(t, lab.now.Add(-time.Minute).AddDate(1, 0, 0).Equal(got[0].SharingEndsAt), got[0].SharingEndsAt)
	require.Len(t, got[0].Accounts, 1)
	assert.Equal(t, "account-1", got[0].Accounts[0].AccountID)
	assertDisclosures(t, got[0], lab.now, lab.now.Add(2*time.Hour))

	lab.advance(time.Hour)
	siteRequest(t, lab.site, http.MethodPost, withdrawTarget("arrangement-1"), "", http.StatusNoContent,
		lab.session...)

	ended := lab.consents(t)
	require.Len(t, ended, 1)
	assert.Equal(t, "ENDED", ended[0].Status)
	assertDisclosures(t, ended[0], lab.now, lab.now.Add(2*time.Hour))

	state := lab.state(t, "arrangement-1")
	assert.Equal(t, domain.EndReasonWithdrawnByConsumer, state.EndReason)
	assert.True(t, lab.now.Add(3*time.Hour).Equal(state.EndedAt), "the end is at the time of the request: %s",
		state.EndedAt)
	assert.Equal(t, domain.WithdrawalMethodConsumerDashboard, state.WithdrawalMethod())
	assert.Equal(t, domain.NoticePending, state.NoticeStatus())
	assert.Equal(t, map[string]string{"arrangement-1": "CONSUMER_DASHBOARD"}, lab.operatorWithdrawalMethods(t))

	_, err := lab.verifier.VerifyConsentToken(t.Context(), token, cert)
	require.ErrorIs(t, err, authserver.ErrInvalidToken)
	require.ErrorContains(t, err, "its grant is revoked")
	assert.Equal(t, http.StatusUnauthorized, lab.getAccounts(t, token, cert))

	grant, err := lab.grants.GrantByRefreshToken(t.Context(), refreshToken("arrangement-1"))
	require.NoError(t, err)
	assert.False(t, grant.RevokedAt.IsZero(), "the grant of the refresh token is revoked")

	siteRequest(t, lab.site, http.MethodPost, withdrawTarget("arrangement-1"), "", http.StatusConflict,
		lab.session...)
}

func assertDisclosures(t *testing.T, consent siteConsent, first, last time.Time) {
	t.Helper()

	require.Len(t, consent.DataClusters, 1)
	cluster := consent.DataClusters[0]
	assert.Equal(t, "Account name, type and balance", cluster.Name)
	require.NotNil(t, cluster.FirstDisclosedAt)
	require.NotNil(t, cluster.LastDisclosedAt)
	assert.True(t, first.Equal(*cluster.FirstDisclosedAt), "first: %s", cluster.FirstDisclosedAt)
	assert.True(t, last.Equal(*cluster.LastDisclosedAt), "last: %s", cluster.LastDisclosedAt)
}
