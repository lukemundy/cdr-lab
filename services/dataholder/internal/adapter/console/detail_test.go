package console_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/arrangementstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/customerstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/registrationstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/transactionstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

type records struct {
	customers    *customerstore.Store
	transactions *transactionstore.Store
	events       *eventstore.Store
	pool         *pgxpool.Pool
	arrangements arrangementstore.Queries
	ports        console.Ports
}

func newRecords(t *testing.T) records {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	r := records{
		customers:    customerstore.New(pool, nil, nil),
		transactions: transactionstore.New(pool),
		pool:         pool,
		events: eventstore.New(
			pool, append(arrangementstore.Projections(), registrationstore.Projections()...)...,
		),
		arrangements: arrangementstore.NewQueries(pool),
	}
	r.ports = console.Ports{
		Details: app.CustomerDetails{
			Customers: r.customers, Transactions: r.transactions, Arrangements: r.arrangements,
			Registrations: registrationstore.NewQueries(pool),
		},
		Registrations:      registrationstore.NewQueries(pool),
		ClientArrangements: r.arrangements,
	}

	return r
}

func (r records) addAccount(t *testing.T, number string, account domain.Account) {
	t.Helper()

	account.ProductCategory = domain.CategoryTransAndSavingsAccounts
	account.ProductName = "Mock Everyday Account"
	account.BSB = "123456"
	account.OpenStatus = domain.OpenStatusOpen
	account.CreatedAt = time.Date(2026, 10, 15, 23, 30, 0, 0, time.UTC)
	require.NoError(t, r.customers.AddAccount(t.Context(), number, account))
}

func (r records) addTransaction(t *testing.T, accountID string, transaction domain.Transaction) {
	t.Helper()

	require.NoError(t, r.transactions.AddTransaction(t.Context(), accountID, transaction))
}

func (r records) addArrangement(t *testing.T, arrangement domain.Arrangement) {
	t.Helper()

	arrangement.CustomerNumber, arrangement.ClientID, arrangement.SoftwareID = "00000001", "client-1", "software-1"
	arrangement.ClientName = "Recipient Web"
	require.NoError(t, app.ArrangementType.Execute(t.Context(), r.events, arrangement.ArrangementID,
		domain.CreateArrangement{Arrangement: arrangement}))
}

func posted(id, description string, cents domain.Amount, at time.Time) domain.Transaction {
	return domain.Transaction{
		TransactionID: id, Type: domain.TransactionPayment, Status: domain.TransactionPosted,
		Description: description, Amount: cents, Reference: "ref-" + id, PostingDateTime: &at,
	}
}

func TestTheDetailOfACustomerGivesTheCustomerEachAccountWithItsBalancesAndEachArrangement(t *testing.T) {
	r := newRecords(t)
	addCustomer(t, r.customers, "00000001", "Alex", "Example")
	r.addAccount(t, "00000001", domain.Account{
		AccountID: "account-1", DisplayName: "Everyday", AccountNumber: "12345678", CreditLimit: 10_000,
	})
	r.addAccount(t, "00000001", domain.Account{AccountID: "account-2", DisplayName: "Saver", AccountNumber: "87654321"})
	r.addTransaction(t, "account-1", posted("transaction-1", "Salary", 250_000, now.Add(-time.Hour)))
	pending := posted("transaction-2", "Coffee", -450, now)
	pending.Status, pending.PostingDateTime, pending.ExecutionDateTime = domain.TransactionPending, nil, &now
	r.addTransaction(t, "account-1", pending)

	r.addArrangement(t, domain.Arrangement{
		ArrangementID: "arrangement-1", Scopes: []string{"openid", "bank:accounts.basic:read"},
		AccountIDs: []string{"account-1"}, ConfirmedAt: now.Add(-2 * time.Hour), SharingDurationSeconds: 86_400,
	})
	r.addArrangement(t, domain.Arrangement{
		ArrangementID: "arrangement-2", Scopes: []string{"openid"}, AccountIDs: []string{},
		ConfirmedAt: now.Add(-time.Hour), SharingDurationSeconds: 86_400,
	})
	// A withdrawal before the data holder recorded the method.
	require.NoError(t, app.ArrangementType.Execute(t.Context(), r.events, "arrangement-2", domain.WithdrawArrangement{
		WithdrawnAt: now.Add(-time.Minute), Method: domain.WithdrawalMethodAlternative,
	}))

	resp := get(t, r.ports, "/api/customers/00000001")

	require.Equal(t, http.StatusOK, resp.status, resp.body)
	assert.JSONEq(t, `{
		"customerNumber": "00000001", "firstName": "Alex", "lastName": "Example",
		"emailAddress": "00000001@mail.test", "lastUpdated": "2026-10-15T13:30:00.123456Z",
		"accounts": [{
			"accountId": "account-1", "displayName": "Everyday", "productCategory": "TRANS_AND_SAVINGS_ACCOUNTS",
			"productName": "Mock Everyday Account", "bsb": "123456", "accountNumber": "12345678",
			"creditLimit": "100.00", "currentBalance": "2500.00", "availableBalance": "2595.50",
			"openStatus": "OPEN", "creationDate": "2026-10-15",
			"actions": [{"action": "close", "ends": []}, {"action": "create-transaction", "ends": []}]
		}, {
			"accountId": "account-2", "displayName": "Saver", "productCategory": "TRANS_AND_SAVINGS_ACCOUNTS",
			"productName": "Mock Everyday Account", "bsb": "123456", "accountNumber": "87654321",
			"creditLimit": "0.00", "currentBalance": "0.00", "availableBalance": "0.00",
			"openStatus": "OPEN", "creationDate": "2026-10-15",
			"actions": [{"action": "close", "ends": []}, {"action": "create-transaction", "ends": []}]
		}],
		"arrangements": [{
			"arrangementId": "arrangement-1", "clientId": "client-1", "clientRegistered": false,
			"softwareId": "software-1", "clientName": "Recipient Web",
			"scopes": ["openid", "bank:accounts.basic:read"], "accountIds": ["account-1"],
			"confirmedAt": "2026-10-19T22:00:00Z", "sharingEndsAt": "2026-10-20T22:00:00Z", "status": "CURRENT",
			"actions": [{"action": "withdraw", "ends": [
				{"arrangementId": "arrangement-1", "softwareId": "software-1", "clientName": "Recipient Web"}
			]}]
		}, {
			"arrangementId": "arrangement-2", "clientId": "client-1", "clientRegistered": false,
			"softwareId": "software-1", "clientName": "Recipient Web",
			"scopes": ["openid"], "accountIds": [],
			"confirmedAt": "2026-10-19T23:00:00Z", "sharingEndsAt": "2026-10-20T23:00:00Z", "status": "ENDED",
			"endedAt": "2026-10-19T23:59:00Z", "endReason": "WITHDRAWN_BY_CONSUMER",
			"withdrawalMethod": "ALTERNATIVE_METHOD", "noticeStatus": "PENDING", "noticeAttempts": [], "actions": []
		}]
	}`, resp.body)
}

func TestAnArrangementWithANoticeGivesEachAttemptOfTheNotice(t *testing.T) {
	r := newRecords(t)
	addCustomer(t, r.customers, "00000001", "Alex", "Example")

	for _, id := range []string{"arrangement-pending", "arrangement-failed", "arrangement-revoked"} {
		r.addArrangement(t, domain.Arrangement{
			ArrangementID: id, Scopes: []string{"openid"}, AccountIDs: []string{},
			ConfirmedAt: now.Add(-time.Hour), SharingDurationSeconds: 86_400,
		})
	}

	execute := func(id string, command domain.ArrangementCommand) {
		t.Helper()
		require.NoError(t, app.ArrangementType.Execute(t.Context(), r.events, id, command))
	}
	at := func(seconds int) time.Time {
		return now.Add(-time.Hour + time.Duration(seconds)*time.Second).In(time.FixedZone("UTC+8", 8*3600))
	}
	withdraw := domain.WithdrawArrangement{WithdrawnAt: now.Add(-time.Hour), Method: domain.WithdrawalMethodAlternative}

	execute("arrangement-pending", withdraw)
	execute("arrangement-pending", domain.RecordNoticeAttempt{
		Attempt: domain.NoticeAttempt{Attempt: 1, At: at(1), Result: "status 503"},
	})
	execute("arrangement-failed", withdraw)
	execute("arrangement-failed", domain.RecordNoticeAttempt{
		Attempt: domain.NoticeAttempt{Attempt: 1, At: at(1), Result: "timeout"},
	})
	execute("arrangement-failed", domain.RecordNoticeAttempt{
		Attempt: domain.NoticeAttempt{Attempt: 2, At: at(11), Result: "no connection"},
	})
	execute("arrangement-failed", domain.RecordNotice{
		RecordedAt: at(71), Status: domain.NoticeFailed,
		Attempt: domain.NoticeAttempt{Attempt: 3, At: at(71), Result: "tls error"},
	})
	execute("arrangement-revoked", domain.RevokeArrangement{RevokedAt: now.Add(-time.Hour)})

	resp := get(t, r.ports, "/api/customers/00000001")

	require.Equal(t, http.StatusOK, resp.status, resp.body)

	var detail struct {
		Arrangements []struct {
			ArrangementID  string          `json:"arrangementId"`
			NoticeAttempts json.RawMessage `json:"noticeAttempts"`
		} `json:"arrangements"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.body), &detail))
	require.Len(t, detail.Arrangements, 3)

	attemptsOf := map[string]string{}
	for _, a := range detail.Arrangements {
		attemptsOf[a.ArrangementID] = string(a.NoticeAttempts)
	}

	assert.JSONEq(t, `[{"attempt": 1, "at": "2026-10-19T23:00:01Z", "result": "status 503"}]`,
		attemptsOf["arrangement-pending"])
	assert.JSONEq(t, `[
		{"attempt": 1, "at": "2026-10-19T23:00:01Z", "result": "timeout"},
		{"attempt": 2, "at": "2026-10-19T23:00:11Z", "result": "no connection"},
		{"attempt": 3, "at": "2026-10-19T23:01:11Z", "result": "tls error"}
	]`, attemptsOf["arrangement-failed"])
	assert.Empty(t, attemptsOf["arrangement-revoked"], "an end with no notice has no attempts")
}

// Verifies: DH-1639
func TestAnArrangementLinksToTheRegistrationOfItsClientOnlyWhileTheRegistrationExists(t *testing.T) {
	r := newRecords(t)
	addCustomer(t, r.customers, "00000001", "", "Example")
	r.addArrangement(t, domain.Arrangement{
		ArrangementID: "arrangement-1", Scopes: []string{"openid"}, AccountIDs: []string{},
		ConfirmedAt: now.Add(-time.Hour), SharingDurationSeconds: 86_400,
	})
	r.register(t, registration("client-1", "software-1"))

	registered := func() bool {
		resp := get(t, r.ports, "/api/customers/00000001")
		require.Equal(t, http.StatusOK, resp.status, resp.body)

		var detail struct {
			Arrangements []struct {
				ClientID         string `json:"clientId"`
				ClientRegistered bool   `json:"clientRegistered"`
			} `json:"arrangements"`
		}
		require.NoError(t, json.Unmarshal([]byte(resp.body), &detail))
		require.Len(t, detail.Arrangements, 1)
		assert.Equal(t, "client-1", detail.Arrangements[0].ClientID)

		return detail.Arrangements[0].ClientRegistered
	}

	assert.True(t, registered())

	require.NoError(t, app.ClientRegistrationType.Execute(
		t.Context(), r.events, "software-1", domain.DeleteClient{ClientID: "client-1"},
	))
	r.register(t, registration("client-2", "software-1"))

	assert.False(t, registered(), "the new registration of the software product has a different client ID")
}

// Verifies: DH-1640
func TestAnArrangementAtTheEndOfItsSharingHasEnded(t *testing.T) {
	r := newRecords(t)
	addCustomer(t, r.customers, "00000001", "", "Example")
	r.addArrangement(t, domain.Arrangement{
		ArrangementID: "arrangement-1", Scopes: []string{"openid"}, AccountIDs: []string{},
		ConfirmedAt: now.Add(-time.Hour), SharingDurationSeconds: 3600,
	})

	resp := get(t, r.ports, "/api/customers/00000001")

	require.Equal(t, http.StatusOK, resp.status, resp.body)

	var detail struct {
		Accounts     []any `json:"accounts"`
		Arrangements []struct {
			Status  string `json:"status"`
			EndedAt string `json:"endedAt"`
		} `json:"arrangements"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.body), &detail))
	assert.Empty(t, detail.Accounts)
	require.Len(t, detail.Arrangements, 1)
	assert.Equal(t, "ENDED", detail.Arrangements[0].Status)
	assert.Empty(t, detail.Arrangements[0].EndedAt, "only an event before the end of the sharing has a time")
}

// Verifies: DH-1622
func TestAnUnknownCustomerOrAccountIsNotFoundWithNoIdentifierInTheDetail(t *testing.T) {
	r := newRecords(t)
	addCustomer(t, r.customers, "00000001", "Alex", "Example")

	for _, target := range []string{
		"/api/customers/00047291",
		"/api/customers/00047291/accounts/account-zebracorn/transactions",
		"/api/customers/00000001/accounts/account-zebracorn/transactions",
	} {
		resp := get(t, r.ports, target)

		require.Equal(t, http.StatusNotFound, resp.status, target)
		assert.NotContains(t, resp.body, "00047291", target)
		assert.NotContains(t, resp.body, "zebracorn", target)
	}
}

// Verifies: DH-1625, DH-1627
func TestTheTransactionsOfAnAccountAreNewestFirstInPages(t *testing.T) {
	r := newRecords(t)
	addCustomer(t, r.customers, "00000001", "Alex", "Example")
	r.addAccount(t, "00000001", domain.Account{AccountID: "account-1", DisplayName: "Everyday", AccountNumber: "1"})

	for i := range 3 {
		r.addTransaction(t, "account-1",
			posted(fmt.Sprintf("transaction-%d", i), "Coffee", -450, now.Add(time.Duration(i)*time.Hour)))
	}

	detailed := posted("transaction-9", "Rent", -120_000, now.AddDate(10, 0, 0))
	detailed.MerchantName, detailed.ValueDateTime = "Example Agents", &now
	detailed.ExtendedData = &domain.ExtendedData{
		Payee: "Example Agents", ExtensionUType: domain.ExtensionNPPPayload,
		NPPPayload: domain.NPPPayload{Service: domain.NPPServiceX2P1, ServiceVersion: "03"},
	}
	r.addTransaction(t, "account-1", detailed)

	resp := get(t, r.ports, "/api/customers/00000001/accounts/account-1/transactions?pageSize=2")

	require.Equal(t, http.StatusOK, resp.status, resp.body)
	assert.JSONEq(t, `{"data": [{
		"transactionId": "transaction-9", "type": "PAYMENT", "status": "POSTED", "description": "Rent",
		"amount": "-1200.00", "reference": "ref-transaction-9", "postingDateTime": "2036-10-20T00:00:00Z",
		"valueDateTime": "2026-10-20T00:00:00Z", "merchantName": "Example Agents",
		"extendedData": {"payee": "Example Agents", "extensionUType": "nppPayload",
			"nppPayload": {"extendedDescription": "", "service": "X2P1", "serviceVersion": "03"}}
	}, {
		"transactionId": "transaction-2", "type": "PAYMENT", "status": "POSTED", "description": "Coffee",
		"amount": "-4.50", "reference": "ref-transaction-2", "postingDateTime": "2026-10-20T02:00:00Z"
	}], "totalRecords": 4}`, resp.body)

	resp = get(t, r.ports, "/api/customers/00000001/accounts/account-1/transactions?page=3&pageSize=2")

	require.Equal(t, http.StatusOK, resp.status, resp.body)
	assert.JSONEq(t, `{"data": [], "totalRecords": 4}`, resp.body)
}

func TestAPageOfTransactionsThatIsNotValidGivesBadRequestWithNoValue(t *testing.T) {
	r := newRecords(t)

	for _, target := range []string{"page=zebracorn", "page=0", "pageSize=101"} {
		resp := get(t, r.ports, "/api/customers/00000001/accounts/account-1/transactions?"+target)

		require.Equal(t, http.StatusBadRequest, resp.status, target)
		assert.NotContains(t, resp.body, "zebracorn", target)
	}
}
