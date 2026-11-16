package banksite_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/banksite"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

type withdrawCall struct {
	customerNumber, arrangementID string
	method                        domain.WithdrawalMethod
}

type withdrawer struct {
	calls []withdrawCall
	err   error
}

func (w *withdrawer) WithdrawByConsumer(
	_ context.Context,
	customerNumber, arrangementID string,
	method domain.WithdrawalMethod,
) (domain.ArrangementState, error) {
	w.calls = append(w.calls, withdrawCall{customerNumber, arrangementID, method})

	return domain.ArrangementState{}, w.err
}

func withdrawSite(t *testing.T, consents *withdrawer) *site {
	t.Helper()

	return newSiteWith(t, banksite.Ports{Customers: customerOfSession(), Accounts: &accounts{}, Withdrawer: consents})
}

func signedInWithdraw(t *testing.T, consents *withdrawer, arrangementID string) response {
	t.Helper()

	s := withdrawSite(t, consents)
	cookie := requestCookie(cookieName, sessionCookie(t, s.signIn(signedIn)).Value)

	return s.do(request{
		method: http.MethodPost, target: "/api/consents/" + arrangementID + "/withdraw",
		cookies: []*http.Cookie{cookie},
	})
}

// Verifies: DH-1769
func TestAWithdrawalOnTheSiteWithdrawsTheConsentOfTheCustomerOfTheSessionWithTheDashboardMethod(t *testing.T) {
	consents := &withdrawer{}

	resp := signedInWithdraw(t, consents, "arrangement-1")

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)
	assert.Empty(t, resp.body)
	assert.Equal(t, []withdrawCall{{
		customerNumber: signedIn, arrangementID: "arrangement-1", method: domain.WithdrawalMethodConsumerDashboard,
	}}, consents.calls)
}

// Verifies: DH-1771
func TestARefusedWithdrawalOnTheSiteGivesTheStatusOfTheRefusalAndNoIdentifier(t *testing.T) {
	tests := map[string]struct {
		err    error
		status int
	}{
		"the consent ended":                 {err: domain.ErrArrangementEnded, status: http.StatusConflict},
		"a consent of a different customer": {err: domain.ErrArrangementNotFound, status: http.StatusNotFound},
		"a failure":                         {err: errors.New("no database"), status: http.StatusInternalServerError},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			resp := signedInWithdraw(t, &withdrawer{err: tc.err}, "arrangement-9")

			assert.Equal(t, tc.status, resp.status)
			assert.NotContains(t, resp.body, "arrangement-9")
			assert.NotContains(t, resp.body, "no database")
		})
	}
}

// Verifies: DH-1770
func TestAWithdrawalOfAConsentThatEndedSaysThatTheConsentChanged(t *testing.T) {
	resp := signedInWithdraw(t, &withdrawer{err: domain.ErrArrangementEnded}, "arrangement-1")

	require.Equal(t, http.StatusConflict, resp.status)
	assert.Contains(t, resp.body, "This consent changed after you opened this confirmation: it already ended.")
	assert.Contains(t, resp.body, "We changed nothing.")
}

// Verifies: DH-1743
func TestAWithdrawalNeedsASession(t *testing.T) {
	consents := &withdrawer{}
	s := withdrawSite(t, consents)

	resp := s.do(request{method: http.MethodPost, target: "/api/consents/arrangement-1/withdraw"})

	assert.Equal(t, http.StatusForbidden, resp.status)
	assert.Empty(t, consents.calls)
}

// Verifies: DH-1744
func TestAWithdrawalForACustomerThatTheDataHolderNoLongerHasEndsTheSession(t *testing.T) {
	resp := signedInWithdraw(t, &withdrawer{err: app.ErrCustomerNotFound}, "arrangement-1")

	require.Equal(t, http.StatusForbidden, resp.status)
	assert.Contains(t, resp.body, `"type":"`+sessionEndedType+`"`)
	assertExpired(t, resp)
}
