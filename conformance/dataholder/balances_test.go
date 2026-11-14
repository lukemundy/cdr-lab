//go:build conformance

package dataholder

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// All balance data in this file is invented.

const (
	balanceScope   = "openid bank:accounts.basic:read"
	everyDataScope = "openid common:customer.basic:read bank:accounts.basic:read bank:transactions:read"
)

type balancePage struct {
	Data struct {
		Balances []map[string]any `json:"balances"`
	} `json:"data"`
	Links map[string]string `json:"links"`
	Meta  map[string]int    `json:"meta"`
}

func (p balancePage) byAccount() map[string]map[string]any {
	balances := map[string]map[string]any{}
	for _, b := range p.Data.Balances {
		balances[b["accountId"].(string)] = b
	}

	return balances
}

func (p balancePage) accountIDs() []any {
	ids := []any{}
	for _, b := range p.Data.Balances {
		ids = append(ids, b["accountId"])
	}

	return ids
}

func getBalance(t *testing.T, client registeredClient, accountID, token string) response {
	t.Helper()

	return bankingResource(t, client.cert, bankingAccountsPath+"/"+accountID+"/balance", "1", token)
}

func getBulkBalances(t *testing.T, client registeredClient, query, token string) balancePage {
	t.Helper()

	resp := bankingResource(t, client.cert, bankingAccountsPath+"/balances"+query, "2", token)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))

	var page balancePage
	require.NoError(t, json.Unmarshal(resp.body, &page))

	return page
}

// postBalances validates a response of 200 against the pinned CDS banking OpenAPI specification.
func postBalances(t *testing.T, client registeredClient, query, contentType, body, token string) response {
	t.Helper()

	header := resourceHeader(token)
	header.Set("Content-Type", contentType)

	resp := secureRequest(t, client.cert, http.MethodPost,
		secureBaseURI(t)+bankingAccountsPath+"/balances"+query, header, strings.NewReader(body))

	if resp.StatusCode == http.StatusOK {
		spec, err := LoadBankingSpec()
		require.NoError(t, err)

		resp.Body = io.NopCloser(bytes.NewReader(resp.body))
		require.NoError(t, spec.ValidateResponse(resp.Response))
	}

	return resp
}

func getSpecificBalances(t *testing.T, client registeredClient, query, token string, ids []string) balancePage {
	t.Helper()

	body, err := json.Marshal(map[string]any{"data": map[string]any{"accountIds": ids}})
	require.NoError(t, err)

	resp := postBalances(t, client, query, "application/json", string(body), token)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))

	var page balancePage
	require.NoError(t, json.Unmarshal(resp.body, &page))

	return page
}

func balance(current, available, creditLimit string) map[string]any {
	return map[string]any{"currentBalance": current, "availableBalance": available, "creditLimit": creditLimit}
}

func amounts(b map[string]any) map[string]any {
	return map[string]any{
		"currentBalance":   b["currentBalance"],
		"availableBalance": b["availableBalance"],
		"creditLimit":      b["creditLimit"],
	}
}

func TestASoftwareProductReadsTheBalancesOfTheAccountsInAConsent(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newAccount(t)
	number := customer.customer.CustomerNumber
	everyday := customer.addAccount("Everyday", "50000001")
	card := createAccount(t, number, operatorAccount{
		DisplayName: "Card", ProductCategory: "CRED_AND_CHRG_CARDS", ProductName: "Low Rate Card",
		BSB: "372100", AccountNumber: "5000000002", CreditLimit: "2500.00",
	})
	savings := customer.addAccount("Savings", "50000003")
	now := time.Now()
	pending := func(description, amount string) operatorTransaction {
		return operatorTransaction{
			Type: "PAYMENT", Status: "PENDING", Description: description, Amount: amount,
			ExecutionDateTime: now.UTC().Format(time.RFC3339),
		}
	}

	addTransaction(t, number, everyday.AccountID, posted("Salary", "1000.00", now.AddDate(-3, 0, 0)))
	addTransaction(t, number, everyday.AccountID, posted("Coffee", "-45.00", now.Add(-time.Hour)))
	addTransaction(t, number, everyday.AccountID, pending("Lunch", "-12.50"))
	addTransaction(t, number, everyday.AccountID, pending("Refund", "5.00"))
	addTransaction(t, number, card.AccountID, posted("Furniture", "-3000.00", now.Add(-time.Hour)))
	addTransaction(t, number, card.AccountID, pending("Fuel", "-10.00"))

	tokens := customer.consents(client, balanceScope, ninetyDays, everyday.AccountID, card.AccountID).tokens(t)
	token := tokens["access_token"].(string)
	ids := accountIDs(t, client, tokens)

	everydayBalance := balance("955.00", "942.50", "0.00")
	cardBalance := balance("-3000.00", "0.00", "2500.00")

	t.Run("Get Account Balance gives the balance from the transactions and the credit limit", func(t *testing.T) {
		for name, want := range map[string]map[string]any{"Everyday": everydayBalance, "Card": cardBalance} {
			resp := getBalance(t, client, ids[name], token)
			require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
			assert.Equal(t, "1", resp.Header.Get("x-v"))

			var body envelope[map[string]any]
			require.NoError(t, json.Unmarshal(resp.body, &body))
			assert.Equal(t, ids[name], body.Data["accountId"])
			assert.Equal(t, want, amounts(body.Data), name)
			assert.Equal(t, secureBaseURI(t)+bankingAccountsPath+"/"+ids[name]+"/balance", body.Links.Self)
		}
	})

	other := newAccount(t)
	otherAccount := other.addAccount("Everyday", "50000004")
	otherIDs := accountIDs(
		t,
		client,
		other.consents(client, balanceScope, ninetyDays, otherAccount.AccountID).tokens(t),
	)

	t.Run("Get Account Balance gives 404 for an account that the consent does not give", func(t *testing.T) {
		for name, accountID := range map[string]string{
			"does not exist":             uuid.NewString(),
			"identifier of the operator": everyday.AccountID,
			"of a different customer":    otherIDs["Everyday"],
		} {
			t.Run(name, func(t *testing.T) {
				resp := getBalance(t, client, accountID, token)

				requireResourceError(t, resp, http.StatusNotFound,
					"urn:au-cds:error:cds-banking:Authorisation/InvalidBankingAccount", accountID)
			})
		}
	})

	t.Run("Get Bulk Balances gives the balances with the filters of Get Accounts and pagination", func(t *testing.T) {
		all := getBulkBalances(t, client, "", token)
		assert.Equal(t, []any{ids["Everyday"], ids["Card"]}, all.accountIDs(), "in the sequence of creation")
		assert.Equal(t, cardBalance, amounts(all.byAccount()[ids["Card"]]))
		assert.Equal(t, map[string]int{"totalRecords": 2, "totalPages": 1}, all.Meta)

		cards := getBulkBalances(t, client, "?product-category=CRED_AND_CHRG_CARDS", token)
		assert.Equal(t, []any{ids["Card"]}, cards.accountIDs())

		second := getBulkBalances(t, client, "?page=2&page-size=1", token)
		assert.Equal(t, []any{ids["Card"]}, second.accountIDs())
		assert.Equal(t, map[string]int{"totalRecords": 2, "totalPages": 2}, second.Meta)
		assert.Equal(t, secureBaseURI(t)+bankingAccountsPath+"/balances?page=1&page-size=1", second.Links["prev"])
	})

	t.Run("Get Balances For Specific Accounts gives one balance for each requested account", func(t *testing.T) {
		one := getSpecificBalances(t, client, "", token, []string{ids["Card"]})
		assert.Equal(t, []any{ids["Card"]}, one.accountIDs())
		assert.Equal(t, cardBalance, amounts(one.Data.Balances[0]))

		repeated := getSpecificBalances(t, client, "", token, []string{ids["Card"], ids["Everyday"], ids["Card"]})
		assert.Equal(t, []any{ids["Everyday"], ids["Card"]}, repeated.accountIDs())
		assert.Equal(t, map[string]int{"totalRecords": 2, "totalPages": 1}, repeated.Meta)

		paged := getSpecificBalances(t, client, "?page=2&page-size=1", token, []string{ids["Everyday"], ids["Card"]})
		assert.Equal(t, []any{ids["Card"]}, paged.accountIDs())

		assert.Empty(t, getSpecificBalances(t, client, "", token, []string{}).Data.Balances)
	})

	t.Run(
		"Get Balances For Specific Accounts gives 422 and no data for accounts that it cannot give",
		func(t *testing.T) {
			unknown := uuid.NewString()
			body, err := json.Marshal(map[string]any{"data": map[string]any{
				"accountIds": []string{ids["Everyday"], unknown, savings.AccountID},
			}})
			require.NoError(t, err)

			resp := postBalances(t, client, "", "application/json", string(body), token)
			require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, string(resp.body))

			var errs struct {
				Data   any `json:"data"`
				Errors []struct {
					Code   string `json:"code"`
					Title  string `json:"title"`
					Detail string `json:"detail"`
				} `json:"errors"`
			}
			require.NoError(t, json.Unmarshal(resp.body, &errs))
			assert.Nil(t, errs.Data)
			require.Len(t, errs.Errors, 2)

			for i, detail := range []string{unknown, savings.AccountID} {
				assert.Equal(t, "urn:au-cds:error:cds-banking:Authorisation/InvalidBankingAccount", errs.Errors[i].Code)
				assert.Equal(t, "Invalid Banking Account", errs.Errors[i].Title)
				assert.Equal(t, detail, errs.Errors[i].Detail)
			}
		},
	)

	t.Run("Get Balances For Specific Accounts refuses a body that is not valid", func(t *testing.T) {
		const missing, invalid = "urn:au-cds:error:cds-all:Field/Missing", "urn:au-cds:error:cds-all:Field/Invalid"

		titles := map[string]string{missing: "Missing Required Field", invalid: "Invalid Field"}

		for body, want := range map[string]struct{ code, detail string }{
			`{}`:                            {missing, "data"},
			`{"data": {}}`:                  {missing, "data.accountIds"},
			`{"data": {"accountIds": "a"}}`: {invalid, "data.accountIds"},
			`not JSON`:                      {invalid, "data"},
		} {
			resp := postBalances(t, client, "", "application/json", body, token)

			detail := requireCDSError(t, resp, http.StatusBadRequest, want.code, titles[want.code])
			assert.Equal(t, want.detail, detail, body)
		}

		resp := postBalances(t, client, "", "text/plain", `{"data": {"accountIds": []}}`, token)
		assert.Equal(t, "The Content-Type must be application/json.", requireCDSError(
			t,
			resp,
			http.StatusUnsupportedMediaType,
			"urn:au-cds:error:cds-all:GeneralError/Expected",
			"Expected Error Encountered",
		))
	})

	closeAccount(t, number, card.AccountID)

	t.Run("a closed account in the consent has a balance", func(t *testing.T) {
		resp := getBalance(t, client, ids["Card"], token)
		require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))

		var body envelope[map[string]any]
		require.NoError(t, json.Unmarshal(resp.body, &body))
		assert.Equal(t, cardBalance, amounts(body.Data))
	})

	t.Run("the balance APIs need the scope bank:accounts.basic:read", func(t *testing.T) {
		noAccounts := customer.consents(client, noAccountScope, ninetyDays).tokens(t)["access_token"].(string)

		const detail = "The consent does not have the scope bank:accounts.basic:read."

		for name, resp := range map[string]response{
			"Get Account Balance": getBalance(t, client, ids["Everyday"], noAccounts),
			"Get Bulk Balances": bankingResource(
				t, client.cert, bankingAccountsPath+"/balances", "2", noAccounts,
			),
			"Get Balances For Specific Accounts": postBalances(
				t, client, "", "application/json", `{"data": {"accountIds": []}}`, noAccounts,
			),
		} {
			t.Run(name, func(t *testing.T) {
				requireResourceError(
					t, resp, http.StatusForbidden, "urn:au-cds:error:cds-all:Authorisation/InvalidConsent", detail,
				)
			})
		}
	})
}

func TestADataRecipientReadsEachTypeOfDataOfAConsent(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newAccount(t)
	number := customer.customer.CustomerNumber
	everyday := customer.addAccount("Everyday", "60000001")
	addTransaction(t, number, everyday.AccountID, posted("Coffee", "-4.50", time.Now().Add(-time.Hour)))

	tokens := customer.consents(client, everyDataScope, ninetyDays, everyday.AccountID).tokens(t)
	token := tokens["access_token"].(string)

	customerResp := getCustomer(t, client.cert, resourceHeader(token))
	require.Equal(t, http.StatusOK, customerResp.StatusCode, string(customerResp.body))
	assert.Contains(t, string(customerResp.body), `"lastName":"`+customer.customer.LastName+`"`)

	accounts := getAccounts(t, client, "", token)
	assert.Equal(t, []any{"Everyday"}, displayNames(accounts))

	accountID := accounts.Data.Accounts[0]["accountId"].(string)
	assert.Equal(t, []any{"Coffee"}, descriptions(getTransactions(t, client, accountID, "", token)))

	resp := getBalance(t, client, accountID, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))

	var body envelope[map[string]any]
	require.NoError(t, json.Unmarshal(resp.body, &body))
	assert.Equal(t, balance("-4.50", "0.00", "0.00"), amounts(body.Data))
}
