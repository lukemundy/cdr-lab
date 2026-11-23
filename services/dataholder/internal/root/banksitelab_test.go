package root

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/pkg/platform/jobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/arrangementstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/banksite"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/customerstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/disclosurestore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/eligibilityjobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/grantstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/httpapi"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/jtistore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/productstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/registrationstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/revokedtokenstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/signingkeystore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/transactionstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// All customer, account and arrangement data in this file is invented.

// siteLab has the online bank site, the APIs of a consent, the stores and the consent-token verification of the
// data holder, with a clock that moves only with advance.
type siteLab struct {
	site         http.Handler
	consentAPI   http.Handler
	session      []*http.Cookie
	customers    *customerstore.Store
	events       *eventstore.Store
	reader       arrangementstore.Queries
	grants       *grantstore.Store
	arrangements app.Arrangements
	verifier     *authserver.DataHolder
	signingKey   authserver.SigningKey
	clientID     string
	// now is the time of the start of the lab.
	now     time.Time
	advance func(time.Duration)
}

// noPages is never called: the test has no authorisation flow.
type noPages struct{}

func (noPages) Interact(http.ResponseWriter, *http.Request, authserver.Interaction) authserver.Outcome {
	return authserver.Outcome{}
}

func (noPages) Refused(http.ResponseWriter, *http.Request, int) {}

type activeStatuses struct{}

func (activeStatuses) Active(context.Context, string) (bool, error) { return true, nil }

func newSiteLab(t *testing.T) siteLab {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	inserter, err := jobs.NewInsertOnlyClient(pool)
	require.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Microsecond)
	current := now
	clock := func() time.Time { return current }
	events := newEventStore(pool, inserter)
	customers := customerstore.New(pool, eligibilityjobs.Hook(inserter), transactionstore.Insert)
	grants := grantstore.New(pool)
	reader := arrangementstore.NewQueries(pool)
	arrangements := app.Arrangements{Store: events, Grants: grants, Customers: customers, Reader: reader, Now: clock}

	registration, err := startedRegistrar(t, pool).Register(t.Context(), "request-1")
	require.NoError(t, err)

	signingKeys := signingkeystore.New(pool)
	require.NoError(t, signingKeys.EnsureSigningKey(t.Context()))
	keys, err := signingKeys.SigningKeys(t.Context())
	require.NoError(t, err)

	verifier, err := authserver.NewDataHolder(authserver.DataHolderConfig{
		Issuer:         testIssuer,
		SecureBaseURL:  "https://mtls.bank.example",
		Scopes:         domain.SupportedScopes(),
		Clients:        registrationstore.Clients{Reader: registrationstore.NewQueries(pool)},
		JTIs:           jtistore.AuthServer{Store: jtistore.New(pool)},
		SigningKeys:    signingKeys.SigningKeys,
		Pages:          noPages{},
		Grants:         grants,
		RevokedTokens:  revokedtokenstore.New(pool),
		Arrangements:   arrangementRevoker{arrangements},
		ClientStatuses: activeStatuses{},
	})
	require.NoError(t, err)

	require.NoError(t, customers.AddCustomer(t.Context(), domain.Customer{
		CustomerNumber: siteCustomer, LastName: "Quokka", EmailAddress: "quokka@example.test", LastUpdated: now,
	}))

	products := productstore.New(pool)
	require.NoError(t, products.ImportProducts(t.Context(), []domain.Product{{
		ProductID: "everyday", LastUpdated: now.Add(-time.Hour),
		ProductCategory: domain.CategoryTransAndSavingsAccounts, Name: "Everyday Account",
		Description: "An account.", Brand: "Coastal Bank",
	}}))

	consentData := app.ConsentData{
		Arrangements: events, Customers: customers, TransactionStore: transactionstore.New(pool), Now: clock,
	}
	consentAPI := http.NewServeMux()
	httpapi.RegisterSecure(consentAPI, httpapi.Secure{
		ResourceBaseURL: "https://mtls.bank.example", RegistrationPath: "/register", UserInfoPath: "/userinfo",
		ConsentTokens: verifier, ClientStatuses: activeClients{}, Arrangements: consentData,
		Customers: consentData, Accounts: consentData, Balances: consentData, Transactions: consentData,
		Disclosures: app.Disclosures{Store: disclosurestore.New(pool), Now: clock},
	})

	siteCustomers := app.Customers{
		Store: customers, Arrangements: arrangements, Products: products, BSB: openBSB,
		Random: func() uint64 { return 12_345_678 }, NewID: func() string { return "account-1" }, Now: clock,
	}
	customerData := app.CustomerData{
		Customers: customers, TransactionStore: transactionstore.New(pool), Arrangements: reader,
		Disclosures: disclosurestore.New(pool), Now: clock,
	}
	listeners, err := bankSiteListeners(testBankSite, nil, ui, banksite.Ports{
		BrandName:  "Coastal Bank",
		BrandID:    "d7385786-d8b1-4d4b-bb80-b21616df1a4d",
		Customers:  siteCustomers,
		Opener:     siteCustomers,
		Accounts:   customerData,
		Consents:   customerData,
		Withdrawer: arrangements,
		Now:        clock,
	})
	require.NoError(t, err)

	site := listeners[0].Handler
	session := siteRequest(t, site, http.MethodPost, "/api/session", `{"identifier": "`+siteCustomer+`"}`,
		http.StatusNoContent)
	// The customer opens the account on the site, as a consumer does.
	siteRequest(t, site, http.MethodPost, "/api/accounts", `{"productId": "everyday"}`, http.StatusCreated, session...)

	return siteLab{
		site:         site,
		consentAPI:   consentAPI,
		session:      session,
		customers:    customers,
		events:       events,
		reader:       reader,
		grants:       grants,
		arrangements: arrangements,
		verifier:     verifier,
		signingKey:   keys[0],
		clientID:     registration.ClientID,
		now:          now,
		advance:      func(d time.Duration) { current = current.Add(d) },
	}
}

// consent creates an arrangement of the registered client, and a grant for it.
func (l siteLab) consent(t *testing.T, arrangementID string, accountIDs ...string) {
	t.Helper()

	require.NoError(t, app.ArrangementType.Execute(t.Context(), l.events, arrangementID, domain.CreateArrangement{
		Arrangement: domain.Arrangement{
			ArrangementID: arrangementID, CustomerNumber: siteCustomer, ClientID: l.clientID,
			SoftwareID: "software-1", ClientName: "Recipient Web", OrgName: "Recipient Brand",
			LegalEntityName: "Recipient Pty Ltd", Scopes: []string{"bank:accounts.basic:read"},
			AccountIDs: append([]string{}, accountIDs...), ConfirmedAt: l.now.Add(-time.Minute),
			SharingDurationSeconds: int64(365 * 24 * time.Hour / time.Second),
		},
	}))
	require.NoError(t, l.grants.SaveGrant(t.Context(), authserver.Grant{
		ID: "grant-" + arrangementID, ClientID: l.clientID, Subject: "subject-1",
		Scopes: []string{"openid", "bank:accounts.basic:read"}, ArrangementID: arrangementID,
		SharingDuration: 365 * 24 * time.Hour, AuthenticatedAt: l.now, SharingEndsAt: l.now.AddDate(1, 0, 0),
		CreatedAt: l.now, Code: "code-" + arrangementID, CodeExpiresAt: l.now.Add(time.Minute),
		CodeConsumedAt: l.now, RedirectURI: "https://recipient.example/cb", RefreshToken: refreshToken(arrangementID),
	}))
}

func refreshToken(arrangementID string) string { return "refresh-" + arrangementID }

// accessToken signs a token of the grant of the arrangement with the key of the data holder, bound to cert.
func (l siteLab) accessToken(t *testing.T, arrangementID string, cert *x509.Certificate) string {
	t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.PS256, Key: l.signingKey.Key},
		(&jose.SignerOptions{}).WithHeader(jose.HeaderKey("kid"), l.signingKey.ID),
	)
	require.NoError(t, err)

	sum := sha256.Sum256(cert.Raw)
	token, err := jwt.Signed(signer).Claims(jwt.Claims{
		Issuer: testIssuer, ID: uuid.NewString(), Expiry: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
	}).Claims(map[string]any{
		"client_id": l.clientID, "scope": "bank:accounts.basic:read", "grant_id": "grant-" + arrangementID,
		"cnf": map[string]string{"x5t#S256": base64.RawURLEncoding.EncodeToString(sum[:])},
	}).Serialize()
	require.NoError(t, err)

	return token
}

type siteCloseAction struct {
	Action string `json:"action"`
	Ends   []struct {
		ArrangementID string `json:"arrangementId"`
		ClientName    string `json:"clientName"`
	} `json:"ends"`
}

type siteCloseAccount struct {
	AccountID  string            `json:"accountId"`
	OpenStatus string            `json:"openStatus"`
	Actions    []siteCloseAction `json:"actions"`
}

func (l siteLab) accounts(t *testing.T) []siteCloseAccount {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "https://example.com/api/accounts", nil)
	for _, cookie := range l.session {
		req.AddCookie(cookie)
	}

	rec := httptest.NewRecorder()
	l.site.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body struct {
		Accounts []siteCloseAccount `json:"accounts"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	return body.Accounts
}

// endsOfClose gives the body of a close with the arrangements of the close action of the only account.
func (l siteLab) endsOfClose(t *testing.T) string {
	t.Helper()

	accounts := l.accounts(t)
	require.Len(t, accounts, 1)
	require.Len(t, accounts[0].Actions, 1)
	require.Equal(t, "close", accounts[0].Actions[0].Action)

	ends := []string{}
	for _, e := range accounts[0].Actions[0].Ends {
		ends = append(ends, e.ArrangementID)
	}

	body, err := json.Marshal(map[string][]string{"ends": ends})
	require.NoError(t, err)

	return string(body)
}

func (l siteLab) state(t *testing.T, arrangementID string) domain.ArrangementState {
	t.Helper()

	states, err := l.reader.ArrangementsOf(t.Context(), siteCustomer)
	require.NoError(t, err)

	for _, state := range states {
		if state.Arrangement.ArrangementID == arrangementID {
			return state
		}
	}

	require.FailNow(t, "the customer has no arrangement "+arrangementID)

	return domain.ArrangementState{}
}
