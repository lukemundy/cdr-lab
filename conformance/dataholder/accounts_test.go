//go:build conformance

package dataholder

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// All account data in this file is invented.

const (
	bankingAccountsPath = "/cds-au/v1/banking/accounts"
	accountScope        = "openid bank:accounts.basic:read bank:accounts.detail:read"
	// accountsProduct is only for the test that deletes its registration.
	accountsProduct = "3e5a7c94-1b2d-4f6e-8a9b-0c1d2e3f4a87"
)

type accountPage struct {
	Data struct {
		Accounts []map[string]any `json:"accounts"`
	} `json:"data"`
	Links map[string]string `json:"links"`
	Meta  map[string]int    `json:"meta"`
}

// bankingResource validates each response with a status that the pinned CDS banking OpenAPI specification documents.
func bankingResource(t *testing.T, cert tls.Certificate, path, version, token string) response {
	t.Helper()

	header := resourceHeader(token)
	header.Set("X-V", version)

	resp := secureRequest(t, cert, http.MethodGet, secureBaseURI(t)+path, header, nil)

	spec, err := LoadBankingSpec()
	require.NoError(t, err)

	resp.Body = io.NopCloser(bytes.NewReader(resp.body))
	require.NoError(t, spec.ValidateDocumentedResponse(resp.Response))

	return resp
}

func getAccounts(t *testing.T, client registeredClient, query, token string) accountPage {
	t.Helper()

	resp := bankingResource(t, client.cert, bankingAccountsPath+query, "3", token)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))

	var page accountPage
	require.NoError(t, json.Unmarshal(resp.body, &page))

	return page
}

func getAccountDetail(t *testing.T, client registeredClient, accountID, token string) response {
	t.Helper()

	return bankingResource(t, client.cert, bankingAccountsPath+"/"+accountID, "5", token)
}

func displayNames(page accountPage) []any {
	names := []any{}
	for _, account := range page.Data.Accounts {
		names = append(names, account["displayName"])
	}

	return names
}

// accountIDs is keyed by display name.
func accountIDs(t *testing.T, client registeredClient, tokens map[string]any) map[string]string {
	t.Helper()
	require.IsType(t, "", tokens["access_token"], tokens)

	ids := map[string]string{}
	for _, account := range getAccounts(t, client, "", tokens["access_token"].(string)).Data.Accounts {
		ids[account["displayName"].(string)] = requireResourceID(t, account["accountId"])
	}

	return ids
}

func registerNew(t *testing.T, productID string) registeredClient {
	t.Helper()

	r := newRegistrant(t, productID)
	reply := r.register(t, r.claims())
	require.Equal(t, http.StatusCreated, reply.status, reply.body)
	require.IsType(t, "", reply.body["client_id"])

	return registeredClient{registrant: r, clientID: reply.body["client_id"].(string), reply: reply}
}

func closeAccount(t *testing.T, customerNumber, accountID string) {
	t.Helper()

	reply := operate(t, http.MethodPost, closePath(customerNumber, accountID), nil)
	require.Equal(t, http.StatusOK, reply.status, string(reply.body))
}

func TestASoftwareProductReadsTheAccountsInAConsent(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newAccount(t)
	created := time.Now().UTC().Format(time.DateOnly)
	everyday := customer.addAccount("Everyday", "30000001")
	card := createAccount(t, customer.customer.CustomerNumber, operatorAccount{
		DisplayName:     "Card",
		ProductCategory: "CRED_AND_CHRG_CARDS",
		ProductName:     "Low Rate Card",
		BSB:             "372100",
		AccountNumber:   "3000000002",
		CreditLimit:     "2500.00",
	})
	savings := customer.addAccount("Savings", "30000003")
	createdAfter := time.Now().UTC().Format(time.DateOnly)

	tokens := customer.consents(client, accountScope, ninetyDays, card.AccountID, everyday.AccountID).tokens(t)
	token := tokens["access_token"].(string)

	page := getAccounts(t, client, "", token)
	require.Len(t, page.Data.Accounts, 2, "only the accounts of the consent, in the sequence of their creation")

	ids := map[string]string{}

	for i, want := range []struct{ name, masked, category, product string }{
		{"Everyday", "xxx-xxx xxxx0001", "TRANS_AND_SAVINGS_ACCOUNTS", "Everyday Account"},
		{"Card", "xxx-xxx xxxxxx0002", "CRED_AND_CHRG_CARDS", "Low Rate Card"},
	} {
		account := page.Data.Accounts[i]
		ids[want.name] = requireResourceID(t, account["accountId"])
		assert.Equal(t, want.name, account["displayName"])
		assert.Equal(t, want.masked, account["maskedNumber"])
		assert.Equal(t, want.category, account["productCategory"])
		assert.Equal(t, want.product, account["productName"])
		assert.Equal(t, "OPEN", account["openStatus"])
		assert.Equal(t, true, account["isOwned"])
		assert.Equal(t, "ONE_PARTY", account["accountOwnership"])
		assert.Contains(t, []string{created, createdAfter}, account["creationDate"])
	}

	assert.NotEqual(t, everyday.AccountID, ids["Everyday"], "not the account identifier of the operator interface")
	assert.NotEqual(t, card.AccountID, ids["Card"])
	assert.Equal(t, map[string]int{"totalRecords": 2, "totalPages": 1}, page.Meta)
	assert.Equal(t, secureBaseURI(t)+bankingAccountsPath+"?page=1&page-size=25", page.Links["self"])

	t.Run("Get Account Detail gives the BSB and the account number", func(t *testing.T) {
		resp := getAccountDetail(t, client, ids["Everyday"], token)
		require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))

		body := decodeJSON(t, resp)
		require.IsType(t, map[string]any{}, body["data"])
		detail := body["data"].(map[string]any)
		assert.Equal(t, ids["Everyday"], detail["accountId"])
		assert.Equal(t, "372100", detail["bsb"])
		assert.Equal(t, "30000001", detail["accountNumber"])
		assert.Equal(t, map[string]any{"self": secureBaseURI(t) + bankingAccountsPath + "/" + ids["Everyday"]},
			body["links"])
	})

	second := customer.consents(client, "openid bank:accounts.basic:read", ninetyDays,
		everyday.AccountID, savings.AccountID).tokens(t)
	secondIDs := accountIDs(t, client, second)

	t.Run("a second consent of the customer with the software product has the same accountId", func(t *testing.T) {
		assert.Equal(t, ids["Everyday"], secondIDs["Everyday"])
	})

	t.Run("Get Account Detail needs the scope bank:accounts.detail:read", func(t *testing.T) {
		resp := getAccountDetail(t, client, secondIDs["Everyday"], second["access_token"].(string))

		requireResourceError(t, resp, http.StatusForbidden, "urn:au-cds:error:cds-all:Authorisation/InvalidConsent",
			"The consent does not have the scope bank:accounts.detail:read.")
		assert.Equal(t, `Bearer error="insufficient_scope", scope="bank:accounts.detail:read"`,
			resp.Header.Get("WWW-Authenticate"))
	})

	other := newAccount(t)
	otherAccount := other.addAccount("Everyday", "30000004")
	otherIDs := accountIDs(t, client,
		other.consents(client, accountScope, ninetyDays, otherAccount.AccountID).tokens(t))

	t.Run("an accountId that the consent does not give is an invalid banking account", func(t *testing.T) {
		for name, accountID := range map[string]string{
			"does not exist":             uuid.NewString(),
			"identifier of the operator": everyday.AccountID,
			"not in the consent":         secondIDs["Savings"],
			"of a different customer":    otherIDs["Everyday"],
		} {
			t.Run(name, func(t *testing.T) {
				resp := getAccountDetail(t, client, accountID, token)

				requireResourceError(t, resp, http.StatusNotFound,
					"urn:au-cds:error:cds-banking:Authorisation/InvalidBankingAccount", accountID)
			})
		}
	})

	closeAccount(t, customer.customer.CustomerNumber, card.AccountID)

	t.Run("an account that the operator closes stays in the consent as closed", func(t *testing.T) {
		accounts := getAccounts(t, client, "", token).Data.Accounts
		require.Len(t, accounts, 2)
		assert.Equal(t, "CLOSED", accounts[1]["openStatus"])

		resp := getAccountDetail(t, client, ids["Card"], token)
		require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
		assert.Contains(t, string(resp.body), `"openStatus":"CLOSED"`)
	})

	t.Run("Get Accounts gives the accounts that match each filter", func(t *testing.T) {
		for query, want := range map[string][]any{
			"?open-status=OPEN":                     {"Everyday"},
			"?open-status=CLOSED":                   {"Card"},
			"?open-status=ALL":                      {"Everyday", "Card"},
			"?product-category=CRED_AND_CHRG_CARDS": {"Card"},
			"?is-owned=true":                        {"Everyday", "Card"},
			"?is-owned=false":                       {},
			"?open-status=OPEN&product-category=CRED_AND_CHRG_CARDS": {},
		} {
			assert.Equal(t, want, displayNames(getAccounts(t, client, query, token)), query)
		}
	})

	t.Run("Get Accounts is paginated", func(t *testing.T) {
		page := getAccounts(t, client, "?open-status=ALL&page=2&page-size=1", token)

		assert.Equal(t, []any{"Card"}, displayNames(page))
		assert.Equal(t, map[string]int{"totalRecords": 2, "totalPages": 2}, page.Meta)
		assert.Equal(t, secureBaseURI(t)+bankingAccountsPath+"?open-status=ALL&page=1&page-size=1", page.Links["prev"])

		resp := bankingResource(t, client.cert, bankingAccountsPath+"?page=3&page-size=1", "3", token)
		requireCDSError(t, resp, http.StatusUnprocessableEntity, "urn:au-cds:error:cds-all:Field/InvalidPage",
			"Invalid Page")
	})

	t.Run("Get Accounts refuses a filter that is not valid", func(t *testing.T) {
		for query, field := range map[string]string{
			"?open-status=DELETED":      "open-status",
			"?is-owned=yes":             "is-owned",
			"?product-category=SAVINGS": "product-category",
		} {
			resp := bankingResource(t, client.cert, bankingAccountsPath+query, "3", token)

			assert.Equal(t, field, requireCDSError(t, resp, http.StatusBadRequest,
				"urn:au-cds:error:cds-all:Field/Invalid", "Invalid Field"))
		}
	})
}

func TestAResourceIDIsTheSameForOneSoftwareProductAndDifferentForAnother(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newAccount(t)
	everyday := customer.addAccount("Everyday", "30000010")
	addTransaction(t, customer.customer.CustomerNumber, everyday.AccountID, posted("Coffee", "-4.50", time.Now()))

	// ids gives the accountId and the transactionId of the consent.
	ids := func(c registeredClient, tokens map[string]any) [2]string {
		t.Helper()

		accountID := accountIDs(t, c, tokens)["Everyday"]
		page := getTransactions(t, c, accountID, "", tokens["access_token"].(string))

		return [2]string{accountID, transactionIDs(page)["Coffee"]}
	}

	first := customer.consents(client, transactionScope, ninetyDays, everyday.AccountID).tokens(t)
	id := ids(client, first)

	accounts := registerNew(t, accountsProduct)
	other := customer.consents(accounts, transactionScope, ninetyDays, everyday.AccountID).tokens(t)
	otherID := ids(accounts, other)

	t.Run("a software product with the same sector identifier gets the same sub and different IDs",
		func(t *testing.T) {
			assert.Equal(t, client.subjectOf(t, first), accounts.subjectOf(t, other))
			assert.NotEqual(t, id[0], otherID[0])
			assert.NotEqual(t, id[1], otherID[1])
		})

	t.Run("the IDs stay the same after the software product registers again", func(t *testing.T) {
		deleted := accounts.manageRegistration(
			t, http.MethodDelete, accounts.cert, accounts.clientID, accounts.accessToken(t), "",
		)
		require.Equal(t, http.StatusNoContent, deleted.status, string(deleted.body))

		again := registerNew(t, accountsProduct)
		require.NotEqual(t, accounts.clientID, again.clientID)

		tokens := customer.consents(again, transactionScope, ninetyDays, everyday.AccountID).tokens(t)

		assert.Equal(t, otherID, ids(again, tokens))
	})
}

func TestTheConsentPageOffersOnlyOpenAccounts(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newAccount(t)
	open := customer.addAccount("Everyday", "30000020")
	closed := customer.addAccount("Old Savings", "30000021")
	closeAccount(t, customer.customer.CustomerNumber, closed.AccountID)

	consent := customer.authenticate(client, client.claimsFor(accountScope, ninetyDays))

	assert.Equal(t, []string{open.AccountID}, slices.Collect(maps.Keys(choicesOf(consent))))

	t.Run("a confirmation of a closed account is refused", func(t *testing.T) {
		refused := customer.confirm(consent, closed.AccountID)

		assert.Equal(t, http.StatusBadRequest, refused.status)
		assert.True(t, isConsentPage(refused), refused.body)
	})

	t.Run("a customer with only closed accounts can only cancel", func(t *testing.T) {
		onlyClosed := newAccount(t)
		account := onlyClosed.addAccount("Everyday", "30000022")
		closeAccount(t, onlyClosed.customer.CustomerNumber, account.AccountID)

		page := onlyClosed.authenticate(client, client.claimsFor(accountScope, ninetyDays))

		assert.Empty(t, choicesOf(page))
		assert.False(t, canConfirm(page))
		assert.Contains(t, page.body, `value="cancel"`)
	})
}
