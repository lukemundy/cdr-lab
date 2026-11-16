package banksite_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/banksite"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const signedIn = "00000001"

// accounts gives the accounts of the customer signedIn, and keeps the last page of a transaction request.
type accounts struct {
	list         []app.AccountDetail
	arrangements []domain.ArrangementState
	transactions []domain.Transaction
	total        int
	page         [2]int
	err          error
}

func (a *accounts) Accounts(_ context.Context, customerNumber string) (app.CustomerAccounts, error) {
	if customerNumber != signedIn {
		return app.CustomerAccounts{}, app.ErrCustomerNotFound
	}

	return app.CustomerAccounts{Accounts: a.list, Arrangements: a.arrangements}, a.err
}

func (a *accounts) account(customerNumber, accountID string) (app.AccountDetail, error) {
	if a.err != nil {
		return app.AccountDetail{}, a.err
	}

	if customerNumber != signedIn {
		return app.AccountDetail{}, app.ErrCustomerNotFound
	}

	for _, detail := range a.list {
		if detail.Account.AccountID == accountID {
			return detail, nil
		}
	}

	return app.AccountDetail{}, app.ErrAccountNotFound
}

func (a *accounts) Balance(_ context.Context, customerNumber, accountID string) (domain.Balance, error) {
	detail, err := a.account(customerNumber, accountID)

	return detail.Balance, err
}

func (a *accounts) Transactions(
	_ context.Context,
	customerNumber, accountID string,
	page, pageSize int,
) ([]domain.Transaction, int, error) {
	if _, err := a.account(customerNumber, accountID); err != nil {
		return nil, 0, err
	}

	a.page = [2]int{page, pageSize}

	return a.transactions, a.total, nil
}

func everyday() app.AccountDetail {
	return app.AccountDetail{
		Account: domain.Account{
			AccountID: "account-1", DisplayName: "Everyday", ProductCategory: domain.CategoryTransAndSavingsAccounts,
			ProductName: "Saver", BSB: "062000", AccountNumber: "12345678", CreditLimit: 10_000,
			OpenStatus: domain.OpenStatusOpen, CreatedAt: time.Date(2026, 10, 1, 23, 0, 0, 0, time.UTC),
		},
		Balance: domain.Balance{Current: -450, Available: 9_550, CreditLimit: 10_000},
	}
}

// signedInGet signs in the customer signedIn, and sends the request with the session cookie.
func signedInGet(t *testing.T, reader *accounts, target string) response {
	t.Helper()

	s := newSiteWith(t, banksite.Ports{
		Customers: newCustomers(domain.Customer{CustomerNumber: signedIn, LastName: "Quoll"}),
		Accounts:  reader,
	})
	cookie := sessionCookie(t, s.signIn(signedIn))

	return s.do(request{
		method: http.MethodGet, target: target, cookies: []*http.Cookie{requestCookie(cookieName, cookie.Value)},
	})
}

func TestTheAccountsAreThoseOfTheCustomerOfTheSessionWithTheFieldsOfTheCDSAndTheBalance(t *testing.T) {
	resp := signedInGet(t, &accounts{list: []app.AccountDetail{everyday()}}, "/api/accounts")

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{"accounts": [{
		"accountId": "account-1", "creationDate": "2026-10-01", "displayName": "Everyday",
		"maskedNumber": "xxx-xxx xxxx5678", "openStatus": "OPEN", "productCategory": "TRANS_AND_SAVINGS_ACCOUNTS",
		"productName": "Saver",
		"balance": {"currentBalance": "-4.50", "availableBalance": "95.50", "creditLimit": "100.00"},
		"actions": [{"action": "close", "ends": []}]
	}]}`, resp.body)
}

func arrangement(id, clientName string) domain.ArrangementState {
	return domain.ArrangementState{Created: true, Arrangement: domain.Arrangement{
		ArrangementID: id, CustomerNumber: signedIn, ClientID: "client-" + id, SoftwareID: "software-" + id,
		ClientName: clientName, ConfirmedAt: now.Add(-time.Hour), SharingDurationSeconds: 86_400,
	}}
}

// Verifies: DH-2243
func TestTheCloseOfTheLastOpenAccountGivesEachArrangementThatEndsWithItsClientName(t *testing.T) {
	ended := arrangement("arrangement-3", "Ended App")
	ended.EndedAt, ended.EndReason = now.Add(-time.Minute), domain.EndReasonRevokedByDataRecipient
	closed := everyday()
	closed.Account.AccountID, closed.Account.OpenStatus = "account-2", domain.OpenStatusClosed
	reader := &accounts{
		list: []app.AccountDetail{everyday(), closed},
		arrangements: []domain.ArrangementState{
			arrangement("arrangement-1", "Budget App"), ended, arrangement("arrangement-2", "Tax App"),
		},
	}

	resp := signedInGet(t, reader, "/api/accounts")

	require.Equal(t, http.StatusOK, resp.status)

	var body struct {
		Accounts []struct {
			Actions json.RawMessage `json:"actions"`
		} `json:"accounts"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))
	require.Len(t, body.Accounts, 2)
	assert.JSONEq(t, `[{"action": "close", "ends": [
		{"arrangementId": "arrangement-1", "clientName": "Budget App"},
		{"arrangementId": "arrangement-2", "clientName": "Tax App"}
	]}]`, string(body.Accounts[0].Actions), "an ended arrangement does not end again")
	assert.JSONEq(t, `[]`, string(body.Accounts[1].Actions), "a closed account has no action")
}

func TestTheCloseOfAnAccountThatIsNotTheLastOpenAccountEndsNoArrangement(t *testing.T) {
	second := everyday()
	second.Account.AccountID = "account-2"
	reader := &accounts{
		list:         []app.AccountDetail{everyday(), second},
		arrangements: []domain.ArrangementState{arrangement("arrangement-1", "Budget App")},
	}

	resp := signedInGet(t, reader, "/api/accounts")

	require.Equal(t, http.StatusOK, resp.status)
	assert.Equal(t, 2, strings.Count(resp.body, `"actions":[{"action":"close","ends":[]}]`), resp.body)
}

func TestACustomerWithNoAccountsHasAnEmptyList(t *testing.T) {
	resp := signedInGet(t, &accounts{}, "/api/accounts")

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{"accounts": []}`, resp.body)
}

func TestTheBalanceHasTheFieldsOfBankingBalance(t *testing.T) {
	resp := signedInGet(t, &accounts{list: []app.AccountDetail{everyday()}}, "/api/accounts/account-1/balance")

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{"currentBalance": "-4.50", "availableBalance": "95.50", "creditLimit": "100.00"}`, resp.body)
}

func TestTheTransactionsHaveTheFieldsOfBankingTransactionV2WithTheIdentifiersOfTheDataHolder(t *testing.T) {
	posted := time.Date(2026, 10, 31, 1, 2, 3, 0, time.FixedZone("AWST", 8*60*60))
	reader := &accounts{
		list: []app.AccountDetail{everyday()},
		transactions: []domain.Transaction{
			{
				TransactionID: "transaction-1", Type: domain.TransactionPayment, Status: domain.TransactionPosted,
				Description: "Coffee", Amount: -450, Reference: "Ref 1", PostingDateTime: &posted,
				ExecutionDateTime: &posted, MerchantName: "Cafe", MerchantCategoryCode: "5814",
				ExtendedData: &domain.ExtendedData{Payee: "Cafe", ExtensionUType: domain.ExtensionNPPPayload},
			},
			{
				TransactionID: "transaction-2", Type: domain.TransactionDirectDebit,
				Status: domain.TransactionPending, Description: "Power", Amount: -12_000, ExecutionDateTime: &posted,
				BillerCode: "1234", BillerName: "Power Co", CRN: "987654", APCANumber: "123456",
			},
		},
		total: 27,
	}

	resp := signedInGet(t, reader, "/api/accounts/account-1/transactions")

	require.Equal(t, http.StatusOK, resp.status)
	assert.Equal(t, [2]int{1, 25}, reader.page, "the defaults of the CDS")
	assert.JSONEq(t, `{"totalRecords": 27, "totalPages": 2, "transactions": [
		{
			"accountId": "account-1", "transactionId": "transaction-1", "isDetailAvailable": true,
			"type": "PAYMENT", "status": "POSTED", "description": "Coffee", "amount": "-4.50", "reference": "Ref 1",
			"postingDateTime": "2026-10-30T17:02:03Z", "executionDateTime": "2026-10-30T17:02:03Z",
			"merchantName": "Cafe", "merchantCategoryCode": "5814"
		},
		{
			"accountId": "account-1", "transactionId": "transaction-2", "isDetailAvailable": false,
			"type": "DIRECT_DEBIT", "status": "PENDING", "description": "Power", "amount": "-120.00",
			"reference": "", "executionDateTime": "2026-10-30T17:02:03Z", "billerCode": "1234",
			"billerName": "Power Co", "crn": "987654", "apcaNumber": "123456"
		}
	]}`, resp.body)
}

func TestTheTransactionsHaveThePageOfTheRequest(t *testing.T) {
	reader := &accounts{list: []app.AccountDetail{everyday()}}

	resp := signedInGet(t, reader, "/api/accounts/account-1/transactions?page=3&page-size=1000")

	require.Equal(t, http.StatusOK, resp.status)
	assert.Equal(t, [2]int{3, 1000}, reader.page)
	assert.JSONEq(t, `{"totalRecords": 0, "totalPages": 0, "transactions": []}`, resp.body)
}

func TestAPageThatIsNotValidIsABadRequest(t *testing.T) {
	for _, query := range []string{"page=0", "page=zebracorn", "page-size=0", "page-size=1001"} {
		t.Run(query, func(t *testing.T) {
			reader := &accounts{list: []app.AccountDetail{everyday()}}

			resp := signedInGet(t, reader, "/api/accounts/account-1/transactions?"+query)

			assert.Equal(t, http.StatusBadRequest, resp.status)
			assert.Equal(t, [2]int{}, reader.page)
		})
	}
}

func TestTheAccountDataNeedsASession(t *testing.T) {
	for _, target := range []string{
		"/api/accounts", "/api/accounts/account-1/balance", "/api/accounts/account-1/transactions",
	} {
		t.Run(target, func(t *testing.T) {
			s := newSiteWith(t, banksite.Ports{
				Customers: newCustomers(), Accounts: &accounts{list: []app.AccountDetail{everyday()}},
			})

			resp := s.do(request{method: http.MethodGet, target: target})

			assert.Equal(t, http.StatusForbidden, resp.status)
			assert.NotContains(t, resp.body, "account-1")
		})
	}
}

// Verifies: DH-2209
func TestAnAccountOfADifferentCustomerIsNotFound(t *testing.T) {
	for _, target := range []string{"/api/accounts/account-9/balance", "/api/accounts/account-9/transactions"} {
		t.Run(target, func(t *testing.T) {
			resp := signedInGet(t, &accounts{list: []app.AccountDetail{everyday()}}, target)

			assert.Equal(t, http.StatusNotFound, resp.status)
			assert.NotContains(t, resp.body, "account-9")
		})
	}
}

func TestAFailureOfTheAccountDataIsAServerError(t *testing.T) {
	for _, target := range []string{
		"/api/accounts", "/api/accounts/account-1/balance", "/api/accounts/account-1/transactions",
	} {
		t.Run(target, func(t *testing.T) {
			resp := signedInGet(t, &accounts{err: errors.New("no database")}, target)

			assert.Equal(t, http.StatusInternalServerError, resp.status)
			assert.NotContains(t, resp.body, "no database")
		})
	}
}

// The operator interface has no removal of a customer, but a session must not outlive its customer in each case.
func TestASessionOfACustomerThatTheDataHolderNoLongerHasEnds(t *testing.T) {
	for _, target := range []string{
		"/api/accounts", "/api/accounts/account-1/balance", "/api/accounts/account-1/transactions",
	} {
		t.Run(target, func(t *testing.T) {
			const gone = "00000002"

			s := newSiteWith(t, banksite.Ports{
				Customers: newCustomers(domain.Customer{CustomerNumber: gone, LastName: "Quoll"}),
				Accounts:  &accounts{list: []app.AccountDetail{everyday()}},
			})
			cookie := requestCookie(cookieName, sessionCookie(t, s.signIn(gone)).Value)

			resp := s.do(request{method: http.MethodGet, target: target, cookies: []*http.Cookie{cookie}})

			require.Equal(t, http.StatusForbidden, resp.status)
			assert.Contains(t, resp.body, `"type":"`+sessionEndedType+`"`)
			assertExpired(t, resp)
			assert.Equal(t, http.StatusForbidden, s.readSession(cookie).status, "the session ended")
		})
	}
}
