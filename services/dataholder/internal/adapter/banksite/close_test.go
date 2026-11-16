package banksite_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/banksite"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// signedInClose signs in the customer signedIn, and closes the account of the path with the body.
func signedInClose(t *testing.T, store *customers, accountID, body string) response {
	t.Helper()

	s := newSiteWith(t, banksite.Ports{Customers: store, Accounts: &accounts{}})
	cookie := requestCookie(cookieName, sessionCookie(t, s.signIn(signedIn)).Value)

	return s.do(request{
		method: http.MethodPost, target: "/api/accounts/" + accountID + "/close", body: body,
		cookies: []*http.Cookie{cookie},
	})
}

func customerOfSession() *customers {
	return newCustomers(domain.Customer{CustomerNumber: signedIn, LastName: "Quoll"})
}

// Verifies: DH-2244
func TestACloseClosesTheAccountOfTheCustomerOfTheSessionWithTheArrangementsOfTheConfirmation(t *testing.T) {
	store := customerOfSession()

	resp := signedInClose(t, store, "account-1", `{"ends": ["arrangement-1", "arrangement-2"]}`)

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)
	assert.Empty(t, resp.body)
	assert.Equal(t, []closeCall{{
		customerNumber: signedIn, accountID: "account-1", ends: []string{"arrangement-1", "arrangement-2"},
	}}, store.closes)
}

func TestACloseThatEndsNoArrangementSendsAnEmptyList(t *testing.T) {
	store := customerOfSession()

	resp := signedInClose(t, store, "account-1", `{"ends": []}`)

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)
	require.Len(t, store.closes, 1)
	assert.Equal(t, []string{}, store.closes[0].ends)
}

func TestACloseWithABodyThatIsNotValidGives400AndClosesNothing(t *testing.T) {
	for name, body := range map[string]string{
		"no body":       "",
		"no ends":       `{}`,
		"null":          `null`,
		"ends null":     `{"ends": null}`,
		"other field":   `{"ends": [], "accountId": "account-1"}`,
		"not a list":    `{"ends": "arrangement-1"}`,
		"not an object": `["arrangement-1"]`,
	} {
		t.Run(name, func(t *testing.T) {
			store := customerOfSession()

			resp := signedInClose(t, store, "account-1", body)

			assert.Equal(t, http.StatusBadRequest, resp.status)
			assert.Empty(t, store.closes)
		})
	}
}

func TestARefusedCloseGivesTheStatusOfTheRefusalAndNoIdentifier(t *testing.T) {
	tests := map[string]struct {
		err    error
		status int
	}{
		"the close ends different arrangements": {err: app.ErrEndsChanged, status: http.StatusConflict},
		"the account is closed":                 {err: domain.ErrAccountClosed, status: http.StatusConflict},
		"an account of a different customer":    {err: app.ErrAccountNotFound, status: http.StatusNotFound},
		"a failure": {
			err:    errors.New("no database"),
			status: http.StatusInternalServerError,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			store := customerOfSession()
			store.closeErr = tc.err

			resp := signedInClose(t, store, "account-9", `{"ends": ["arrangement-1"]}`)

			assert.Equal(t, tc.status, resp.status)
			assert.NotContains(t, resp.body, "account-9")
			assert.NotContains(t, resp.body, "arrangement-1")
			assert.NotContains(t, resp.body, "no database")
		})
	}
}

// Verifies: DH-2245
func TestACloseThatEndsDifferentArrangementsSaysThatTheAccountsChanged(t *testing.T) {
	store := customerOfSession()
	store.closeErr = app.ErrEndsChanged

	resp := signedInClose(t, store, "account-1", `{"ends": []}`)

	require.Equal(t, http.StatusConflict, resp.status)
	assert.Contains(t, resp.body, "changed")
	assert.Contains(t, resp.body, "We closed nothing")
}

func TestACloseNeedsASession(t *testing.T) {
	store := customerOfSession()
	s := newSiteWith(t, banksite.Ports{Customers: store, Accounts: &accounts{}})

	resp := s.do(request{method: http.MethodPost, target: "/api/accounts/account-1/close", body: `{"ends": []}`})

	assert.Equal(t, http.StatusForbidden, resp.status)
	assert.Empty(t, store.closes)
}

func TestACloseForACustomerThatTheDataHolderNoLongerHasEndsTheSession(t *testing.T) {
	store := customerOfSession()
	store.closeErr = app.ErrCustomerNotFound
	s := newSiteWith(t, banksite.Ports{Customers: store, Accounts: &accounts{}})
	cookie := requestCookie(cookieName, sessionCookie(t, s.signIn(signedIn)).Value)

	resp := s.do(request{
		method: http.MethodPost, target: "/api/accounts/account-1/close", body: `{"ends": []}`,
		cookies: []*http.Cookie{cookie},
	})

	require.Equal(t, http.StatusForbidden, resp.status)
	assert.Contains(t, resp.body, `"type":"`+sessionEndedType+`"`)
	assertExpired(t, resp)
}
