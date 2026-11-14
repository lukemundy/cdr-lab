package console_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

type noGrants struct{}

func (noGrants) RevokeArrangement(context.Context, string) error { return nil }

type detailActions struct {
	Accounts []struct {
		AccountID  string `json:"accountId"`
		OpenStatus string `json:"openStatus"`
		Actions    []struct {
			Action string `json:"action"`
			Ends   []struct {
				ArrangementID string `json:"arrangementId"`
			} `json:"ends"`
		} `json:"actions"`
	} `json:"accounts"`
	Arrangements []struct {
		ArrangementID    string `json:"arrangementId"`
		Status           string `json:"status"`
		EndReason        string `json:"endReason"`
		WithdrawalMethod string `json:"withdrawalMethod"`
		NoticeStatus     string `json:"noticeStatus"`
		Actions          []struct {
			Action string `json:"action"`
			Ends   []struct {
				ArrangementID string `json:"arrangementId"`
				ClientName    string `json:"clientName"`
			} `json:"ends"`
		} `json:"actions"`
	} `json:"arrangements"`
}

// withCloses gives the ports the use cases of the operator interface, so that a close and a withdrawal have the same
// rules.
func (r records) withCloses() records {
	clock := func() time.Time { return now }
	arrangements := app.Arrangements{
		Store: r.events, Grants: noGrants{}, Customers: r.customers,
		Reader: r.arrangements, Now: clock,
	}
	r.ports.Accounts = app.Customers{Store: r.customers, Arrangements: arrangements, Now: clock}
	r.ports.Arrangements = arrangements

	return r
}

func detailOf(t *testing.T, r records) detailActions {
	t.Helper()

	resp := get(t, r.ports, "/api/customers/00000001")
	require.Equal(t, http.StatusOK, resp.status, resp.body)

	var body detailActions
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))

	return body
}

func twoAccountsAndAnArrangement(t *testing.T) records {
	t.Helper()

	r := newRecords(t).withCloses()
	addCustomer(t, r.customers, "00000001", "Alex", "Example")
	r.addAccount(t, "00000001", domain.Account{AccountID: "account-1", DisplayName: "Everyday", AccountNumber: "1"})
	r.addAccount(t, "00000001", domain.Account{AccountID: "account-2", DisplayName: "Saver", AccountNumber: "2"})
	r.addArrangement(t, domain.Arrangement{
		ArrangementID: "arrangement-current", Scopes: []string{"openid"}, AccountIDs: []string{},
		ConfirmedAt: now.Add(-time.Hour), SharingDurationSeconds: 86_400,
	})
	r.addArrangement(t, domain.Arrangement{
		ArrangementID: "arrangement-expired", Scopes: []string{"openid"}, AccountIDs: []string{},
		ConfirmedAt: now.Add(-time.Hour), SharingDurationSeconds: 60,
	})

	return r
}

const (
	closeEveryday = "/api/customers/00000001/accounts/account-1/close"
	closeSaver    = "/api/customers/00000001/accounts/account-2/close"
	endsNothing   = `{"ends": []}`
	endsCurrent   = `{"ends": ["arrangement-current"]}`
)

func TestEachOpenAccountHasACloseThatEndsNoArrangementWhileAnotherAccountIsOpenAndACreateOfATransaction(
	t *testing.T,
) {
	r := twoAccountsAndAnArrangement(t)

	detail := detailOf(t, r)

	require.Len(t, detail.Accounts, 2)

	for _, a := range detail.Accounts {
		require.Len(t, a.Actions, 2, a.AccountID)
		assert.Equal(t, "close", a.Actions[0].Action)
		assert.Equal(t, "create-transaction", a.Actions[1].Action)

		for _, action := range a.Actions {
			assert.Empty(t, action.Ends, a.AccountID)
			assert.NotNil(t, action.Ends, "ends is [] and not null")
		}
	}
}

func TestACloseClosesTheAccountAndThenTheLastOpenAccountGivesEachArrangementThatItEnds(t *testing.T) {
	r := twoAccountsAndAnArrangement(t)

	resp := post(t, r.ports, closeEveryday, endsNothing)

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)
	assert.Empty(t, resp.body)

	detail := detailOf(t, r)
	assert.Equal(t, "CLOSED", detail.Accounts[0].OpenStatus)
	assert.Empty(t, detail.Accounts[0].Actions, "a closed account has no close and no create of a transaction")
	require.Len(t, detail.Accounts[1].Actions, 2)
	require.Len(t, detail.Accounts[1].Actions[0].Ends, 1, "an arrangement that ended does not end again")
	assert.Equal(t, "arrangement-current", detail.Accounts[1].Actions[0].Ends[0].ArrangementID)
}

func TestTheCloseOfTheLastOpenAccountEndsEachArrangementThatItsConfirmationNamed(t *testing.T) {
	r := twoAccountsAndAnArrangement(t)
	require.Equal(t, http.StatusNoContent, post(t, r.ports, closeEveryday, endsNothing).status)

	resp := post(t, r.ports, closeSaver, endsCurrent)

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)

	detail := detailOf(t, r)
	assert.Equal(t, "ENDED", detail.Arrangements[0].Status)
	assert.Equal(t, "CONSUMER_NOT_ELIGIBLE", detail.Arrangements[0].EndReason)
}

// Two tabs show the detail with two open accounts. Each close ends no arrangement in the confirmation.
func TestACloseThatADifferentCloseMadeTheLastIsAConflictAndChangesNothing(t *testing.T) {
	r := twoAccountsAndAnArrangement(t)
	require.Equal(t, http.StatusNoContent, post(t, r.ports, closeSaver, endsNothing).status)
	before := get(t, r.ports, "/api/customers/00000001").body

	resp := post(t, r.ports, closeEveryday, endsNothing)

	require.Equal(t, http.StatusConflict, resp.status, resp.body)
	assert.JSONEq(t, `{
		"type": "about:blank", "title": "Conflict", "status": 409,
		"detail": "The customer changed after the confirmation, and the close now ends different arrangements. `+
		`Look at the new state of the customer. The data holder changed nothing."
	}`, resp.body)
	assert.JSONEq(t, before, get(t, r.ports, "/api/customers/00000001").body)
}

func TestACloseThatIsNotTheLastWithArrangementsInTheRequestIsAConflict(t *testing.T) {
	r := twoAccountsAndAnArrangement(t)

	resp := post(t, r.ports, closeEveryday, endsCurrent)

	require.Equal(t, http.StatusConflict, resp.status, resp.body)
	assert.Equal(t, "OPEN", detailOf(t, r).Accounts[0].OpenStatus)
}

func TestTheCloseOfAClosedAccountIsAConflictAndChangesNothing(t *testing.T) {
	r := twoAccountsAndAnArrangement(t)
	require.Equal(t, http.StatusNoContent, post(t, r.ports, closeEveryday, endsNothing).status)
	before := get(t, r.ports, "/api/customers/00000001").body

	resp := post(t, r.ports, closeEveryday, endsNothing)

	require.Equal(t, http.StatusConflict, resp.status, resp.body)
	assert.JSONEq(t, `{
		"type": "about:blank", "title": "Conflict", "status": 409,
		"detail": "The account is closed. The data holder changed nothing."
	}`, resp.body)
	assert.JSONEq(t, before, get(t, r.ports, "/api/customers/00000001").body)
}

func TestACloseWithABodyThatIsNotValidIsABadRequestAndChangesNothing(t *testing.T) {
	r := twoAccountsAndAnArrangement(t)

	for _, body := range []string{
		"", "zebracorn", "{}", `{"ends": null}`, `{"ends": "arrangement-current"}`, `{"ends": [1]}`,
		`{"ends": [], "version": 1}`, `[]`,
	} {
		resp := post(t, r.ports, closeEveryday, body)

		require.Equal(t, http.StatusBadRequest, resp.status, body)
		assert.Contains(t, resp.body, "The request body must be a JSON object with only ends", body)
		assert.NotContains(t, resp.body, "zebracorn", body)
	}

	assert.Equal(t, "OPEN", detailOf(t, r).Accounts[0].OpenStatus)
}

func TestTheCloseOfAnUnknownCustomerOrAccountIsNotFoundWithNoIdentifierInTheDetail(t *testing.T) {
	r := twoAccountsAndAnArrangement(t)

	for _, target := range []string{
		"/api/customers/00047291/accounts/account-1/close",
		"/api/customers/00000001/accounts/account-zebracorn/close",
	} {
		resp := post(t, r.ports, target, endsNothing)

		require.Equal(t, http.StatusNotFound, resp.status, target)
		assert.NotContains(t, resp.body, "00047291", target)
		assert.NotContains(t, resp.body, "zebracorn", target)
	}
}

const withdrawCurrent = "/api/customers/00000001/arrangements/arrangement-current/withdraw"

func TestEachArrangementThatDidNotEndHasAWithdrawThatEndsOnlyThatArrangement(t *testing.T) {
	detail := detailOf(t, twoAccountsAndAnArrangement(t))

	require.Len(t, detail.Arrangements, 2)
	current, expired := detail.Arrangements[0], detail.Arrangements[1]
	require.Len(t, current.Actions, 1)
	assert.Equal(t, "withdraw", current.Actions[0].Action)
	require.Len(t, current.Actions[0].Ends, 1)
	assert.Equal(t, "arrangement-current", current.Actions[0].Ends[0].ArrangementID)
	assert.Equal(t, "Recipient Web", current.Actions[0].Ends[0].ClientName)
	assert.Equal(t, "ENDED", expired.Status)
	assert.Empty(t, expired.Actions, "an arrangement that ended has no withdraw")
	assert.NotNil(t, expired.Actions, "actions is [] and not null")
}

func TestAWithdrawalEndsTheArrangementWithTheAlternativeMethodAndANotice(t *testing.T) {
	r := twoAccountsAndAnArrangement(t)

	resp := post(t, r.ports, withdrawCurrent, "")

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)
	assert.Empty(t, resp.body)

	withdrawn := detailOf(t, r).Arrangements[0]
	assert.Equal(t, "ENDED", withdrawn.Status)
	assert.Equal(t, "WITHDRAWN_BY_CONSUMER", withdrawn.EndReason)
	assert.Equal(t, "ALTERNATIVE_METHOD", withdrawn.WithdrawalMethod)
	assert.Equal(t, "PENDING", withdrawn.NoticeStatus)
	assert.Empty(t, withdrawn.Actions)
}

// Two tabs show the same arrangement as current.
func TestTheWithdrawalOfAnArrangementThatEndedIsAConflictAndChangesNothing(t *testing.T) {
	r := twoAccountsAndAnArrangement(t)
	require.Equal(t, http.StatusNoContent, post(t, r.ports, withdrawCurrent, "").status)
	before := get(t, r.ports, "/api/customers/00000001").body

	for _, target := range []string{
		withdrawCurrent, "/api/customers/00000001/arrangements/arrangement-expired/withdraw",
	} {
		resp := post(t, r.ports, target, "")

		require.Equal(t, http.StatusConflict, resp.status, resp.body)
		assert.JSONEq(t, `{
			"type": "about:blank", "title": "Conflict", "status": 409,
			"detail": "The arrangement ended. The data holder changed nothing."
		}`, resp.body)
	}

	assert.JSONEq(t, before, get(t, r.ports, "/api/customers/00000001").body)
}

func TestTheWithdrawalOfAnUnknownCustomerOrArrangementIsNotFoundWithNoIdentifierInTheDetail(t *testing.T) {
	r := twoAccountsAndAnArrangement(t)
	addCustomer(t, r.customers, "00000002", "Sam", "Example")

	for _, target := range []string{
		"/api/customers/00047291/arrangements/arrangement-current/withdraw",
		"/api/customers/00000001/arrangements/arrangement-zebracorn/withdraw",
		"/api/customers/00000002/arrangements/arrangement-current/withdraw",
	} {
		resp := post(t, r.ports, target, "")

		require.Equal(t, http.StatusNotFound, resp.status, target)
		assert.NotContains(t, resp.body, "00047291", target)
		assert.NotContains(t, resp.body, "zebracorn", target)
		assert.NotContains(t, resp.body, "arrangement-current", target)
	}

	assert.Equal(t, "CURRENT", detailOf(t, r).Arrangements[0].Status)
}
