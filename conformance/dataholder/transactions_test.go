//go:build conformance

package dataholder

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// All transaction data in this file is invented.

const transactionScope = "openid bank:accounts.basic:read bank:transactions:read"

// operatorTransaction has no TransactionID in a request, because the data holder creates it.
type operatorTransaction struct {
	TransactionID     string         `json:"transactionId,omitempty"`
	Type              string         `json:"type"`
	Status            string         `json:"status"`
	Description       string         `json:"description"`
	Amount            string         `json:"amount"`
	Reference         string         `json:"reference"`
	PostingDateTime   string         `json:"postingDateTime,omitempty"`
	ExecutionDateTime string         `json:"executionDateTime,omitempty"`
	ExtendedData      map[string]any `json:"extendedData,omitempty"`
}

type transactionPage struct {
	Data struct {
		Transactions []map[string]any `json:"transactions"`
	} `json:"data"`
	Links map[string]string `json:"links"`
	Meta  map[string]any    `json:"meta"`
}

func transactionsPath(customerNumber, accountID string) string {
	return accountsPath(customerNumber) + "/" + accountID + "/transactions"
}

func posted(description, amount string, at time.Time) operatorTransaction {
	return operatorTransaction{
		Type: "PAYMENT", Status: "POSTED", Description: description, Amount: amount,
		PostingDateTime: at.UTC().Format(time.RFC3339),
	}
}

func nppDetail(transaction operatorTransaction, party, name string) operatorTransaction {
	transaction.ExtendedData = map[string]any{
		party: name, "extensionUType": "nppPayload",
		"nppPayload": map[string]any{
			"service": "X2P1", "serviceVersion": "03", "extendedDescription": transaction.Description,
		},
	}

	return transaction
}

func addTransaction(t *testing.T, customerNumber, accountID string, transaction operatorTransaction) string {
	t.Helper()

	reply := operate(t, http.MethodPost, transactionsPath(customerNumber, accountID), transaction)
	require.Equal(t, http.StatusCreated, reply.status, string(reply.body))

	var created operatorTransaction
	reply.decode(t, &created)

	return created.TransactionID
}

func getTransactions(t *testing.T, client registeredClient, accountID, query, token string) transactionPage {
	t.Helper()

	resp := bankingResource(t, client.cert, bankingAccountsPath+"/"+accountID+"/transactions"+query, "2", token)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))

	var page transactionPage
	require.NoError(t, json.Unmarshal(resp.body, &page))

	return page
}

func getTransactionDetail(t *testing.T, client registeredClient, accountID, transactionID, token string) response {
	t.Helper()

	return bankingResource(
		t, client.cert, bankingAccountsPath+"/"+accountID+"/transactions/"+transactionID, "3", token,
	)
}

func descriptions(page transactionPage) []any {
	found := []any{}
	for _, transaction := range page.Data.Transactions {
		found = append(found, transaction["description"])
	}

	return found
}

// transactionIDs is keyed by description.
func transactionIDs(page transactionPage) map[string]string {
	ids := map[string]string{}
	for _, transaction := range page.Data.Transactions {
		ids[transaction["description"].(string)] = transaction["transactionId"].(string)
	}

	return ids
}

func timeQuery(name string, at time.Time) string {
	return name + "=" + url.QueryEscape(at.UTC().Format(time.RFC3339))
}

func TestASoftwareProductReadsTheTransactionsOfAnAccountInAConsent(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newAccount(t)
	number := customer.customer.CustomerNumber
	everyday := customer.addAccount("Everyday", "40000001")
	savings := customer.addAccount("Savings", "40000002")
	now := time.Now()

	pending := operatorTransaction{
		Type: "PAYMENT", Status: "PENDING", Description: "Lunch", Amount: "-12.50",
		ExecutionDateTime: now.Add(-30 * time.Minute).UTC().Format(time.RFC3339),
	}
	salary := nppDetail(posted("Salary", "2500.00", now.AddDate(0, 0, -10)), "payer", "Employer")
	salary.Reference = "OCTOBER PAY"
	coffeeID := addTransaction(
		t,
		number,
		everyday.AccountID,
		posted("Coffee at the cafe", "-45.00", now.Add(-time.Hour)),
	)
	addTransaction(t, number, everyday.AccountID, salary)
	addTransaction(t, number, everyday.AccountID, pending)
	addTransaction(t, number, everyday.AccountID, posted("Old groceries", "-80.00", now.AddDate(0, 0, -100)))
	addTransaction(t, number, everyday.AccountID, posted("Older than two years", "-10.00", now.AddDate(-3, 0, 0)))
	addTransaction(t, number, savings.AccountID, posted("Interest", "1.25", now.Add(-time.Hour)))

	tokens := customer.consents(client, transactionScope, ninetyDays, everyday.AccountID).tokens(t)
	token := tokens["access_token"].(string)
	accountID := accountIDs(t, client, tokens)["Everyday"]

	page := getTransactions(t, client, accountID, "", token)

	assert.Equal(t, []any{"Lunch", "Coffee at the cafe", "Salary"}, descriptions(page),
		"the last 90 days, in descending order of the effective time")
	assert.Equal(t, map[string]any{"totalRecords": 3.0, "totalPages": 1.0, "isQueryParamUnsupported": false},
		page.Meta)

	ids := transactionIDs(page)

	for _, transaction := range page.Data.Transactions {
		requireResourceID(t, transaction["transactionId"])
		assert.Equal(t, accountID, transaction["accountId"])
		assert.Equal(t, transaction["description"] == "Salary", transaction["isDetailAvailable"])
	}

	assert.NotEqual(t, coffeeID, ids["Coffee at the cafe"], "not the identifier of the operator interface")

	t.Run("each filter gives the transactions that match it", func(t *testing.T) {
		withOld := []any{"Lunch", "Coffee at the cafe", "Salary", "Old groceries"}

		for name, tt := range map[string]struct {
			query string
			want  []any
		}{
			"oldest-time":                  {timeQuery("oldest-time", now.AddDate(0, 0, -200)), withOld},
			"oldest-time before two years": {timeQuery("oldest-time", now.AddDate(-5, 0, 0)), withOld},
			"newest-time":                  {timeQuery("newest-time", now.Add(-2*time.Hour)), []any{"Salary"}},
			"amounts": {
				"min-amount=-45.00&max-amount=-12.50", []any{"Lunch", "Coffee at the cafe"},
			},
			"amount with 3 decimal digits": {"min-amount=-45.005&max-amount=-44.995", []any{"Coffee at the cafe"}},
			"text in the description":      {"text=COFFEE", []any{"Coffee at the cafe"}},
			"text in the reference":        {"text=october", []any{"Salary"}},
		} {
			t.Run(name, func(t *testing.T) {
				assert.Equal(t, tt.want, descriptions(getTransactions(t, client, accountID, "?"+tt.query, token)))
			})
		}
	})

	t.Run("the list is paginated", func(t *testing.T) {
		second := getTransactions(t, client, accountID, "?page=2&page-size=1&text=a", token)

		assert.Equal(t, []any{"Salary"}, descriptions(second))
		assert.EqualValues(t, 2, second.Meta["totalPages"])
		assert.Contains(t, second.Links["prev"], "text=a")
	})

	t.Run("a time that is not an RFC 3339 date-time with an offset is refused", func(t *testing.T) {
		resp := bankingResource(t, client.cert,
			bankingAccountsPath+"/"+accountID+"/transactions?oldest-time=2026-10-01T00:00:00", "2", token)

		assert.Equal(t, "oldest-time", requireCDSError(t, resp, http.StatusBadRequest,
			"urn:au-cds:error:cds-all:Field/InvalidDateTime", "Invalid Date"))
	})

	t.Run("Get Transaction Detail gives the extendedData", func(t *testing.T) {
		resp := getTransactionDetail(t, client, accountID, ids["Salary"], token)
		require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))

		detail := decodeJSON(t, resp)["data"].(map[string]any)
		assert.Equal(t, ids["Salary"], detail["transactionId"])
		assert.Equal(t, salary.ExtendedData["payer"], detail["extendedData"].(map[string]any)["payer"])
	})

	t.Run("a transaction with no detail or not of the account is an invalid resource", func(t *testing.T) {
		for name, id := range map[string]string{
			"no detail":                  ids["Coffee at the cafe"],
			"identifier of the operator": coffeeID,
			"does not exist":             uuid.NewString(),
		} {
			t.Run(name, func(t *testing.T) {
				requireResourceError(t, getTransactionDetail(t, client, accountID, id, token),
					http.StatusNotFound, "urn:au-cds:error:cds-all:Resource/Invalid", id)
			})
		}
	})

	second := customer.consents(client, transactionScope, ninetyDays, everyday.AccountID, savings.AccountID).tokens(t)
	secondAccounts := accountIDs(t, client, second)

	t.Run("a second consent has the same transactionId", func(t *testing.T) {
		assert.Equal(
			t,
			ids,
			transactionIDs(getTransactions(t, client, secondAccounts["Everyday"], "", second["access_token"].(string))),
		)
	})

	t.Run("an account that the consent does not give is an invalid banking account", func(t *testing.T) {
		for _, id := range []string{secondAccounts["Savings"], everyday.AccountID} {
			resp := bankingResource(t, client.cert, bankingAccountsPath+"/"+id+"/transactions", "2", token)

			requireResourceError(t, resp, http.StatusNotFound,
				"urn:au-cds:error:cds-banking:Authorisation/InvalidBankingAccount", id)
		}
	})

	t.Run("the transaction APIs need the scope bank:transactions:read", func(t *testing.T) {
		basic := customer.consents(client, "openid bank:accounts.basic:read", ninetyDays, everyday.AccountID).tokens(t)

		resp := bankingResource(t, client.cert, bankingAccountsPath+"/"+accountID+"/transactions", "2",
			basic["access_token"].(string))

		requireResourceError(t, resp, http.StatusForbidden, "urn:au-cds:error:cds-all:Authorisation/InvalidConsent",
			"The consent does not have the scope bank:transactions:read.")
	})
}
