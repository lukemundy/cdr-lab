package root

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/arrangementstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/banksite"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/customerstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/disclosurestore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/httpapi"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/transactionstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	siteCustomer   = "00000001"
	consentAccess  = "access-token"
	sitePageSize   = 3
	accountsOfSite = 2
)

// arrangementToken gives the token of one arrangement for the value consentAccess.
type arrangementToken struct{ arrangementID string }

func (a arrangementToken) VerifyConsentToken(
	_ context.Context,
	token string,
	_ *x509.Certificate,
) (authserver.ConsentToken, error) {
	if token != consentAccess {
		return authserver.ConsentToken{}, authserver.ErrInvalidToken
	}

	return authserver.ConsentToken{
		ClientID:      "client-1",
		Scopes:        []string{"bank:accounts.basic:read", "bank:transactions:read"},
		ArrangementID: a.arrangementID,
	}, nil
}

type activeClients struct{}

func (activeClients) Status(context.Context, string) (domain.SoftwareProductStatus, error) {
	return domain.SoftwareProductActive, nil
}

// siteTransactions has each type and each status, the optional fields, extendedData, one transaction before the 90
// days and one before the 2 years.
func siteTransactions(now time.Time) []domain.Transaction {
	at := func(days int) *time.Time {
		t := now.AddDate(0, 0, -days)
		return &t
	}

	types := []domain.TransactionType{
		domain.TransactionDirectDebit, domain.TransactionFee, domain.TransactionInterestCharged,
		domain.TransactionInterestPaid, domain.TransactionOther, domain.TransactionPayment,
		domain.TransactionTransferIncoming, domain.TransactionTransferOutgoing,
	}

	transactions := make([]domain.Transaction, 0, len(types)+3)
	for i, kind := range types {
		transactions = append(transactions, domain.Transaction{
			TransactionID: "transaction-" + strconv.Itoa(i), Type: kind, Status: domain.TransactionPosted,
			Description: "Posted " + string(kind), Amount: domain.Amount(-100 * (i + 1)), Reference: "Ref",
			PostingDateTime: at(i), ExecutionDateTime: at(i), ValueDateTime: at(i),
		})
	}

	transactions[0].BillerCode, transactions[0].BillerName, transactions[0].CRN = "1234", "Power Co", "987654"
	transactions[5].MerchantName, transactions[5].MerchantCategoryCode = "Cafe", "5814"
	transactions[7].APCANumber = "123456"
	transactions[7].ExtendedData = &domain.ExtendedData{
		Payee: "Friend", ExtensionUType: domain.ExtensionNPPPayload,
		NPPPayload: domain.NPPPayload{
			Service: domain.NPPServiceX2P1, ServiceVersion: "03", ExtendedDescription: "Dinner",
		},
	}

	return append(transactions,
		domain.Transaction{
			TransactionID: "transaction-pending", Type: domain.TransactionPayment, Status: domain.TransactionPending,
			Description: "Pending", Amount: -2_500, ExecutionDateTime: at(0),
		},
		domain.Transaction{
			TransactionID: "transaction-old", Type: domain.TransactionPayment, Status: domain.TransactionPosted,
			Description: "Before the 90 days", Amount: -300, PostingDateTime: at(91),
		},
		domain.Transaction{
			TransactionID: "transaction-very-old", Type: domain.TransactionPayment, Status: domain.TransactionPosted,
			Description: "Before the 2 years", Amount: -300, PostingDateTime: at(800),
		},
	)
}

type accountData struct {
	consent, site http.Handler
	session       []*http.Cookie
}

// newAccountData gives a customer with an open account that has a credit limit and transactions, and a closed
// account, a second customer with an account, and an arrangement for the two accounts of the first customer.
func newAccountData(t *testing.T, now time.Time) accountData {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	customers := customerstore.New(pool, nil, nil)
	transactions := transactionstore.New(pool)

	for i, number := range []string{siteCustomer, "00000002"} {
		require.NoError(t, customers.AddCustomer(t.Context(), domain.Customer{
			CustomerNumber: number,
			LastName:       "Quagga",
			EmailAddress:   "quagga." + number + "@example.test",
			LastUpdated:    now,
		}))
		require.NoError(t, customers.AddAccount(t.Context(), number, domain.Account{
			AccountID: "account-" + number, DisplayName: "Everyday", ProductCategory: domain.CategoryCredAndChrgCards,
			ProductName: "Card", BSB: "062000", AccountNumber: "1234567" + strconv.Itoa(i), CreditLimit: 500_000,
			OpenStatus: domain.OpenStatusOpen, CreatedAt: now.AddDate(0, -1, 0),
		}))
	}

	require.NoError(t, customers.AddAccount(t.Context(), siteCustomer, domain.Account{
		AccountID: "account-closed", DisplayName: "Old saver", ProductCategory: domain.CategoryTransAndSavingsAccounts,
		ProductName: "Saver", BSB: "062000", AccountNumber: "7654321", OpenStatus: domain.OpenStatusClosed,
		CreatedAt: now.AddDate(-1, 0, 0),
	}))

	for _, transaction := range siteTransactions(now) {
		require.NoError(t, transactions.AddTransaction(t.Context(), "account-"+siteCustomer, transaction))
	}

	events := eventstore.New(pool, arrangementstore.Projections()...)
	require.NoError(t, app.ArrangementType.Execute(t.Context(), events, "arrangement-1", domain.CreateArrangement{
		Arrangement: domain.Arrangement{
			ArrangementID: "arrangement-1", CustomerNumber: siteCustomer, ClientID: "client-1",
			SoftwareID: "software-1", ClientName: "Recipient", Scopes: []string{"bank:accounts.basic:read"},
			AccountIDs: []string{"account-" + siteCustomer, "account-closed"}, ConfirmedAt: now,
			SharingDurationSeconds: 3600,
		},
	}))

	clock := func() time.Time { return now }
	consentData := app.ConsentData{
		Arrangements: events, Customers: customers, TransactionStore: transactions, Now: clock,
	}

	consent := http.NewServeMux()
	httpapi.RegisterSecure(consent, httpapi.Secure{
		ResourceBaseURL: "https://mtls.bank.test", RegistrationPath: "/register", UserInfoPath: "/userinfo",
		ConsentTokens: arrangementToken{arrangementID: "arrangement-1"}, ClientStatuses: activeClients{},
		Customers: consentData, Accounts: consentData, Balances: consentData, Transactions: consentData,
		Disclosures: app.Disclosures{Store: disclosurestore.New(pool), Now: clock},
	})

	site := http.NewServeMux()
	require.NoError(t, banksite.Register(site, ui, []string{"bank.test"}, banksite.Ports{
		BrandID:   "d7385786-d8b1-4d4b-bb80-b21616df1a4d",
		Customers: app.Customers{Store: customers, Now: clock},
		Accounts: app.CustomerData{
			Customers: customers, TransactionStore: transactions, Arrangements: arrangementstore.NewQueries(pool),
			Now: clock,
		},
		Now: clock,
	}))

	signIn := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "https://bank.test/api/session",
		strings.NewReader(`{"identifier": "`+siteCustomer+`"}`))
	rec := httptest.NewRecorder()
	site.ServeHTTP(rec, signIn)
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

	return accountData{consent: consent, site: site, session: rec.Result().Cookies()}
}

func getJSON(t *testing.T, handler http.Handler, target string, header http.Header, into any) {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	maps.Copy(req.Header, header)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, "%s: %s", target, rec.Body.String())
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), into))
}

func (d accountData) consentGet(t *testing.T, path string, version int, into any) {
	t.Helper()

	getJSON(t, d.consent, "https://mtls.bank.test/cds-au/v1/banking"+path, http.Header{
		"Authorization":    {"Bearer " + consentAccess},
		"X-V":              {strconv.Itoa(version)},
		"X-Fapi-Auth-Date": {"Thu, 15 Oct 2026 08:00:00 GMT"},
	}, into)
}

func (d accountData) siteGet(t *testing.T, path string, into any) {
	t.Helper()

	rec := d.siteRequest(t, path)
	require.Equal(t, http.StatusOK, rec.Code, "%s: %s", path, rec.Body.String())
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), into))
}

func (d accountData) siteRequest(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "https://bank.test/api"+path, nil)
	for _, cookie := range d.session {
		req.AddCookie(cookie)
	}

	rec := httptest.NewRecorder()
	d.site.ServeHTTP(rec, req)

	return rec
}

// A reader that ignores the customer of the session fails this test.
// Verifies: DH-2209
func TestTheBankSiteGivesNoAccountOfADifferentCustomer(t *testing.T) {
	const otherAccount = "account-00000002"

	data := newAccountData(t, time.Now().UTC().Truncate(time.Second))

	var list struct {
		Accounts []record `json:"accounts"`
	}

	data.siteGet(t, "/accounts", &list)

	for _, account := range list.Accounts {
		assert.NotEqual(t, otherAccount, account["accountId"])
	}

	for _, path := range []string{
		"/accounts/" + otherAccount + "/balance", "/accounts/" + otherAccount + "/transactions",
	} {
		t.Run(path, func(t *testing.T) {
			rec := data.siteRequest(t, path)

			var problem record

			assert.Equal(t, http.StatusNotFound, rec.Code)
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &problem))
			assert.ElementsMatch(t, []string{"type", "title", "status", "detail"}, slices.Collect(maps.Keys(problem)),
				"a problem with no account data")
			assert.NotContains(t, rec.Body.String(), otherAccount)
		})
	}
}

type record = map[string]any

// withoutIDs removes the identifiers that the site and the banking APIs give differently by design.
func withoutIDs(r record) record {
	clone := record{}

	for name, value := range r {
		if name != "accountId" && name != "transactionId" {
			clone[name] = value
		}
	}

	return clone
}

// Verifies: DH-2208
func TestTheBankSiteGivesTheAccountsBalancesAndTransactionsOfTheAPIsOfAConsent(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	data := newAccountData(t, now)

	var consentAccounts struct {
		Data struct {
			Accounts []record `json:"accounts"`
		} `json:"data"`
	}

	data.consentGet(t, "/accounts", 3, &consentAccounts)

	var siteAccounts struct {
		Accounts []record `json:"accounts"`
	}

	data.siteGet(t, "/accounts", &siteAccounts)

	require.Len(t, consentAccounts.Data.Accounts, accountsOfSite)
	require.Len(t, siteAccounts.Accounts, accountsOfSite, "no account of a different customer")

	for i, siteAccount := range siteAccounts.Accounts {
		consentAccount := consentAccounts.Data.Accounts[i]
		consentID, siteID := consentAccount["accountId"].(string), siteAccount["accountId"].(string)

		t.Run(siteAccount["displayName"].(string), func(t *testing.T) {
			// Get Accounts has no balance and no actions.
			for name, value := range withoutIDs(siteAccount) {
				if name != "balance" && name != "actions" {
					assert.Equal(t, consentAccount[name], value, name)
				}
			}

			var consentBalance struct {
				Data record `json:"data"`
			}

			data.consentGet(t, "/accounts/"+consentID+"/balance", 1, &consentBalance)

			var siteBalance record

			data.siteGet(t, "/accounts/"+siteID+"/balance", &siteBalance)
			assert.Equal(t, withoutIDs(consentBalance.Data), siteBalance)
			assert.Equal(t, siteBalance, siteAccount["balance"], "the balance in the list")

			consentTransactions := data.consentTransactions(t, consentID)
			assert.Equal(t, consentTransactions, data.siteTransactions(t, siteID))

			if i == 0 {
				assert.Len(t, consentTransactions, 9, "the 90 days, with no older transaction")
			}
		})
	}
}

func (d accountData) consentTransactions(t *testing.T, accountID string) []record {
	t.Helper()

	var all []record

	for page := 1; ; page++ {
		var body struct {
			Data struct {
				Transactions []record `json:"transactions"`
			} `json:"data"`
			Meta struct {
				TotalPages int `json:"totalPages"`
			} `json:"meta"`
		}

		d.consentGet(t, "/accounts/"+accountID+"/transactions?page-size=2&page="+strconv.Itoa(page), 2, &body)

		for _, transaction := range body.Data.Transactions {
			all = append(all, withoutIDs(transaction))
		}

		if page >= body.Meta.TotalPages {
			return all
		}
	}
}

// siteTransactions gives each page of the site, with a different page size, so that the pages do not hide a
// difference of the order.
func (d accountData) siteTransactions(t *testing.T, accountID string) []record {
	t.Helper()

	var all []record

	for page := 1; ; page++ {
		var body struct {
			Transactions []record `json:"transactions"`
			TotalPages   int      `json:"totalPages"`
		}

		d.siteGet(t, "/accounts/"+accountID+"/transactions?page-size="+strconv.Itoa(sitePageSize)+
			"&page="+strconv.Itoa(page), &body)

		for _, transaction := range body.Transactions {
			assert.Equal(t, accountID, transaction["accountId"], "the account identifier of the data holder")
			all = append(all, withoutIDs(transaction))
		}

		if page >= body.TotalPages {
			return all
		}
	}
}
