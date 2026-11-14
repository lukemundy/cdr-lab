package root

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
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
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/productstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/transactionstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const openBSB = "376100"

// openScopes give Get Account Detail, the balances and the transactions.
var openScopes = []string{"bank:accounts.basic:read", "bank:accounts.detail:read", "bank:transactions:read"}

// detailToken gives the token of arrangement-1 with openScopes, for the value consentAccess.
type detailToken struct{}

func (detailToken) VerifyConsentToken(
	_ context.Context,
	token string,
	_ *x509.Certificate,
) (authserver.ConsentToken, error) {
	if token != consentAccess {
		return authserver.ConsentToken{}, authserver.ErrInvalidToken
	}

	return authserver.ConsentToken{
		ClientID: "client-1", Scopes: openScopes, ArrangementID: "arrangement-1",
	}, nil
}

// openedAccount is a customer of the site, signed in, with one account that the customer opened on the site, and an
// arrangement for that account: the next consent of the customer.
type openedAccount struct {
	accountData

	// opened is the account in the response of the open.
	opened record
}

func everydayProduct() domain.Product {
	return domain.Product{
		ProductID: "everyday", ProductCategory: domain.CategoryTransAndSavingsAccounts, Name: "Everyday Account",
	}
}

// newOpenedAccount opens an account of the product on the site with the body, as a consumer does.
func newOpenedAccount(t *testing.T, now time.Time, product domain.Product, body string) openedAccount {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	customers := customerstore.New(pool, nil, transactionstore.Insert)
	transactions := transactionstore.New(pool)
	products := productstore.New(pool)
	product.LastUpdated, product.Description, product.Brand = now.Add(-time.Hour), "A product.", "Coastal Bank"
	require.NoError(t, products.ImportProducts(t.Context(), []domain.Product{product}))

	clock := func() time.Time { return now }
	site := app.Customers{
		Store: customers, Products: products, BSB: openBSB, Random: func() uint64 { return 12_345_678 },
		NewID: func() string { return "account-opened" }, Now: clock,
	}

	siteMux := http.NewServeMux()
	require.NoError(t, banksite.Register(siteMux, ui, []string{"bank.test"}, banksite.Ports{
		BrandID: "d7385786-d8b1-4d4b-bb80-b21616df1a4d", Customers: site, Opener: site,
		Accounts: app.CustomerData{
			Customers: customers, TransactionStore: transactions, Arrangements: arrangementstore.NewQueries(pool),
			Now: clock,
		},
		Now: clock,
	}))

	created, err := site.Create(t.Context(), domain.Customer{LastName: "Quoll", EmailAddress: "quoll@example.test"})
	require.NoError(t, err)

	var data openedAccount

	data.site = siteMux
	data.session = signInOnSite(t, siteMux, created.CustomerNumber)

	rec := data.sitePost(t, "/accounts", body)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &data.opened))

	events := eventstore.New(pool, arrangementstore.Projections()...)
	require.NoError(t, app.ArrangementType.Execute(t.Context(), events, "arrangement-1", domain.CreateArrangement{
		Arrangement: domain.Arrangement{
			ArrangementID: "arrangement-1", CustomerNumber: created.CustomerNumber, ClientID: "client-1",
			SoftwareID: "software-1", ClientName: "Recipient",
			Scopes:     openScopes,
			AccountIDs: []string{"account-opened"}, ConfirmedAt: now, SharingDurationSeconds: 3600,
		},
	}))

	consentData := app.ConsentData{
		Arrangements: events, Customers: customers, TransactionStore: transactions, Now: clock,
	}
	consent := http.NewServeMux()
	httpapi.RegisterSecure(consent, httpapi.Secure{
		ResourceBaseURL: "https://mtls.bank.test", RegistrationPath: "/register", UserInfoPath: "/userinfo",
		ConsentTokens: detailToken{}, ClientStatuses: activeClients{},
		Customers: consentData, Accounts: consentData, Balances: consentData, Transactions: consentData,
		Disclosures: app.Disclosures{Store: disclosurestore.New(pool), Now: clock},
	})
	data.consent = consent

	return data
}

func signInOnSite(t *testing.T, site http.Handler, customerNumber string) []*http.Cookie {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "https://bank.test/api/session",
		strings.NewReader(`{"identifier": "`+customerNumber+`"}`))
	rec := httptest.NewRecorder()
	site.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

	return rec.Result().Cookies()
}

func (d accountData) sitePost(t *testing.T, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "https://bank.test/api"+path,
		strings.NewReader(body))
	for _, cookie := range d.session {
		req.AddCookie(cookie)
	}

	rec := httptest.NewRecorder()
	d.site.ServeHTTP(rec, req)

	return rec
}

// DESIGN.md, Opening of an account: the account of an open has the category and the name of its product, and the BSB
// of the brand directory. The site gives no BSB, so Get Account Detail shows it.
func TestAnAccountThatTheConsumerOpensOnTheSiteShowsOnTheSiteAndInTheAPIsOfTheNextConsent(t *testing.T) {
	data := newOpenedAccount(t, time.Now().UTC().Truncate(time.Second), everydayProduct(), `{"productId": "everyday"}`)

	var site struct {
		Accounts []record `json:"accounts"`
	}

	data.siteGet(t, "/accounts", &site)
	require.Len(t, site.Accounts, 1)
	// The consent after the open changes only the close action of the account.
	assert.Equal(t, withoutActions(data.opened), withoutActions(site.Accounts[0]),
		"the response of the open is the account of the list")

	var accounts struct {
		Data struct {
			Accounts []record `json:"accounts"`
		} `json:"data"`
	}

	data.consentGet(t, "/accounts", 3, &accounts)
	require.Len(t, accounts.Data.Accounts, 1)

	consentAccount := accounts.Data.Accounts[0]
	assert.Equal(t, "TRANS_AND_SAVINGS_ACCOUNTS", consentAccount["productCategory"])
	assert.Equal(t, "Everyday Account", consentAccount["productName"])
	assert.Equal(t, "Everyday Account", consentAccount["displayName"])
	assert.Equal(t, "OPEN", consentAccount["openStatus"])

	for name, value := range withoutIDs(site.Accounts[0]) {
		if name != "balance" && name != "actions" {
			assert.Equal(t, consentAccount[name], value, name)
		}
	}

	var detail struct {
		Data record `json:"data"`
	}

	data.consentGet(t, "/accounts/"+consentAccount["accountId"].(string), 5, &detail)
	assert.Equal(t, openBSB, detail.Data["bsb"])
	assert.Equal(t, "12345678", detail.Data["accountNumber"])
}

func withoutActions(r record) record {
	clone := maps.Clone(r)
	delete(clone, "actions")

	return clone
}

// DESIGN.md, Opening of an account: revolving credit gets the credit limit, and a loan gets a drawdown of the loan
// amount. The values are those of an account that the operator creates with the same credit limit and transaction.
func TestALendingAccountThatTheConsumerOpensHasTheValueOfTheConsumerInTheAPIsOfTheNextConsent(t *testing.T) {
	at := time.Now().UTC().Truncate(time.Second)

	tests := map[string]struct {
		product      domain.Product
		body         string
		balance      record
		transactions []record
	}{
		"card": {
			product: domain.Product{
				ProductID: "card", ProductCategory: domain.CategoryCredAndChrgCards, Name: "Glide Card",
			},
			body:    `{"productId": "card", "creditLimit": "5000.00"}`,
			balance: record{"currentBalance": "0.00", "availableBalance": "5000.00", "creditLimit": "5000.00"},
		},
		"loan": {
			product: domain.Product{
				ProductID: "loan", ProductCategory: domain.CategoryResidentialMortgages, Name: "Nest First Home Loan",
			},
			body:    `{"productId": "loan", "loanAmount": "650000.50"}`,
			balance: record{"currentBalance": "-650000.50", "availableBalance": "0.00", "creditLimit": "0.00"},
			transactions: []record{{
				"type": "OTHER", "status": "POSTED", "description": "Loan drawdown", "amount": "-650000.50",
				"reference": "", "isDetailAvailable": false,
				"postingDateTime": at.Format(time.RFC3339), "executionDateTime": at.Format(time.RFC3339),
			}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			data := newOpenedAccount(t, at, tt.product, tt.body)

			var accounts struct {
				Data struct {
					Accounts []record `json:"accounts"`
				} `json:"data"`
			}

			data.consentGet(t, "/accounts", 3, &accounts)
			require.Len(t, accounts.Data.Accounts, 1)
			assert.Equal(t, string(tt.product.ProductCategory), accounts.Data.Accounts[0]["productCategory"])

			accountID := accounts.Data.Accounts[0]["accountId"].(string)

			var balance struct {
				Data record `json:"data"`
			}

			data.consentGet(t, "/accounts/"+accountID+"/balance", 1, &balance)
			assert.Equal(t, tt.balance, withoutIDs(balance.Data))
			assert.Equal(t, tt.balance, data.opened["balance"], "the site gives the balance of the consent")
			assert.Equal(t, tt.transactions, data.consentTransactions(t, accountID))
		})
	}
}
