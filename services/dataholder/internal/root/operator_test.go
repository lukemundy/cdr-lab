package root

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/pkg/platform/jobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/statusstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// All customer, account, transaction and arrangement data in this file is invented.

// The data holder creates only customer numbers of 8 digits.
const unknownCustomerNumber = "000000000"

type operatorLab struct {
	pool     *pgxpool.Pool
	handlers map[int]http.Handler
}

func newOperatorLab(t *testing.T) operatorLab {
	t.Helper()

	setOnly(t, apiMode)

	cfg, err := loadConfig(apiMode)
	require.NoError(t, err)

	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	proc, err := newProcess(t.Context(), pool, cfg, apiMode)
	require.NoError(t, err)

	handlers := map[int]http.Handler{}
	for _, listener := range proc.listeners {
		handlers[listener.Port] = listener.Handler
	}

	return operatorLab{pool: pool, handlers: handlers}
}

type operatorCustomer struct {
	CustomerNumber string            `json:"customerNumber"`
	FirstName      string            `json:"firstName,omitempty"`
	LastName       string            `json:"lastName"`
	EmailAddress   string            `json:"emailAddress"`
	Accounts       []operatorAccount `json:"accounts,omitempty"`
}

// customerRequest has no customer number, because the data holder creates it.
type customerRequest struct {
	FirstName    string `json:"firstName,omitempty"`
	LastName     string `json:"lastName"`
	EmailAddress string `json:"emailAddress"`
}

// operatorAccount has no AccountID in a request, because the data holder creates it.
type operatorAccount struct {
	AccountID       string `json:"accountId,omitempty"`
	DisplayName     string `json:"displayName"`
	ProductCategory string `json:"productCategory"`
	ProductName     string `json:"productName"`
	BSB             string `json:"bsb"`
	AccountNumber   string `json:"accountNumber"`
	CreditLimit     string `json:"creditLimit,omitempty"`
	// OpenStatus is only in a response, because the data holder sets it.
	OpenStatus string `json:"openStatus,omitempty"`
}

// operatorTransaction has no TransactionID in a request, because the data holder creates it.
type operatorTransaction struct {
	TransactionID   string `json:"transactionId,omitempty"`
	Type            string `json:"type"`
	Status          string `json:"status"`
	Description     string `json:"description"`
	Amount          string `json:"amount"`
	Reference       string `json:"reference"`
	PostingDateTime string `json:"postingDateTime,omitempty"`
}

type operatorArrangement struct {
	ArrangementID string     `json:"arrangementId"`
	SoftwareID    string     `json:"softwareId"`
	ClientName    string     `json:"clientName"`
	Scopes        []string   `json:"scopes"`
	AccountIDs    []string   `json:"accountIds"`
	ConfirmedAt   time.Time  `json:"confirmedAt"`
	SharingEndsAt time.Time  `json:"sharingEndsAt"`
	EndedAt       *time.Time `json:"endedAt"`
	EndReason     string     `json:"endReason"`
}

type operatorReply struct {
	status int
	header http.Header
	body   []byte
}

func (r operatorReply) decode(t *testing.T, into any) {
	t.Helper()
	assert.Equal(t, "application/json", r.header.Get("Content-Type"))
	require.NoError(t, json.Unmarshal(r.body, into), string(r.body))
}

// problemDetail checks the RFC 9457 format of an error response.
func (r operatorReply) problemDetail(t *testing.T, status int) string {
	t.Helper()
	require.Equal(t, status, r.status, string(r.body))
	assert.Equal(t, "application/problem+json", r.header.Get("Content-Type"))

	var problem struct {
		Type   string `json:"type"`
		Title  string `json:"title"`
		Status int    `json:"status"`
		Detail string `json:"detail"`
	}
	require.NoError(t, json.Unmarshal(r.body, &problem), string(r.body))
	assert.Equal(t, "about:blank", problem.Type)
	assert.Equal(t, http.StatusText(status), problem.Title)
	assert.Equal(t, status, problem.Status)

	return problem.Detail
}

func (l operatorLab) request(t *testing.T, method, path string, header http.Header, body any) operatorReply {
	t.Helper()

	return l.requestHost(t, operatorHost, method, path, header, body)
}

func (l operatorLab) requestHost(
	t *testing.T,
	requestHost, method, path string,
	header http.Header,
	body any,
) operatorReply {
	t.Helper()

	var encoded []byte

	if body != nil {
		var err error

		encoded, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req := httptest.NewRequestWithContext(t.Context(), method, "http://"+requestHost+path, bytes.NewReader(encoded))
	req.Header = header
	rec := httptest.NewRecorder()
	l.handlers[internalPort].ServeHTTP(rec, req)

	return operatorReply{status: rec.Code, header: rec.Header(), body: rec.Body.Bytes()}
}

func (l operatorLab) operate(t *testing.T, method, path string, body any) operatorReply {
	t.Helper()

	return l.request(t, method, path, http.Header{"Authorization": {"Bearer test-operator-token"}}, body)
}

func (l operatorLab) newCustomer(t *testing.T) operatorCustomer {
	t.Helper()

	// The email address of each customer is unique.
	reply := l.operate(t, http.MethodPost, "/v1/customers", customerRequest{
		FirstName: "Alex", LastName: "Example", EmailAddress: strings.ToLower(rand.Text()) + "@example.com",
	})
	require.Equal(t, http.StatusCreated, reply.status, string(reply.body))

	var created operatorCustomer
	reply.decode(t, &created)

	return created
}

func (l operatorLab) customer(t *testing.T, customerNumber string) operatorReply {
	t.Helper()

	return l.operate(t, http.MethodGet, "/v1/customers/"+customerNumber, nil)
}

func (l operatorLab) createAccount(t *testing.T, customerNumber string, account operatorAccount) operatorAccount {
	t.Helper()

	reply := l.operate(t, http.MethodPost, accountsPath(customerNumber), account)
	require.Equal(t, http.StatusCreated, reply.status, string(reply.body))

	var created operatorAccount
	reply.decode(t, &created)

	return created
}

func everydayAccount(accountNumber string) operatorAccount {
	return operatorAccount{
		DisplayName:     "Everyday",
		ProductCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
		ProductName:     "Mock Everyday Account",
		BSB:             "123456",
		AccountNumber:   accountNumber,
	}
}

func accountsPath(customerNumber string) string {
	return "/v1/customers/" + customerNumber + "/accounts"
}

func closePath(customerNumber, accountID string) string {
	return accountsPath(customerNumber) + "/" + accountID + "/close"
}

func transactionsPath(customerNumber, accountID string) string {
	return accountsPath(customerNumber) + "/" + accountID + "/transactions"
}

func arrangementsPath(customerNumber string) string {
	return "/v1/customers/" + customerNumber + "/arrangements"
}

func withdrawPath(customerNumber, arrangementID string) string {
	return arrangementsPath(customerNumber) + "/" + arrangementID + "/withdraw"
}

func TestTheOperatorCreatesCustomersWithDifferentCustomerNumbersOf8Digits(t *testing.T) {
	l := newOperatorLab(t)

	first, second := l.newCustomer(t), l.newCustomer(t)

	assert.Regexp(t, `^[0-9]{8}$`, first.CustomerNumber)
	assert.Regexp(t, `^[0-9]{8}$`, second.CustomerNumber)
	assert.NotEqual(t, first.CustomerNumber, second.CustomerNumber)
	assert.Equal(t, operatorCustomer{
		CustomerNumber: first.CustomerNumber, FirstName: "Alex", LastName: "Example", EmailAddress: first.EmailAddress,
	}, first)
}

func TestTheOperatorCreatesAccountsAndReadsTheCustomerWithEachAccount(t *testing.T) {
	l := newOperatorLab(t)
	customer := l.newCustomer(t)
	saver := operatorAccount{
		DisplayName: "Holiday Saver", ProductCategory: "TRANS_AND_SAVINGS_ACCOUNTS", ProductName: "Mock Bonus Saver",
		BSB: "123456", AccountNumber: "10000002",
	}

	accounts := []operatorAccount{everydayAccount("10000001"), saver}
	for i, account := range accounts {
		created := l.createAccount(t, customer.CustomerNumber, account)
		_, err := uuid.Parse(created.AccountID)
		require.NoError(t, err, "the data holder creates the account identifier as a UUID")

		account.AccountID = created.AccountID
		account.CreditLimit, account.OpenStatus = "0.00", "OPEN"
		assert.Equal(t, account, created)

		accounts[i] = account
	}

	assert.NotEqual(t, accounts[0].AccountID, accounts[1].AccountID)

	reply := l.customer(t, customer.CustomerNumber)
	require.Equal(t, http.StatusOK, reply.status, string(reply.body))

	var read operatorCustomer
	reply.decode(t, &read)

	customer.Accounts = accounts
	assert.Equal(t, customer, read)
}

func TestACustomerWithNoAccountsHasAnEmptyAccountList(t *testing.T) {
	l := newOperatorLab(t)

	reply := l.customer(t, l.newCustomer(t).CustomerNumber)

	require.Equal(t, http.StatusOK, reply.status, string(reply.body))
	assert.Contains(t, string(reply.body), `"accounts":[]`)
}

func TestACustomerThatIsNotValidIsRefusedWithEachRule(t *testing.T) {
	l := newOperatorLab(t)

	reply := l.operate(t, http.MethodPost, "/v1/customers", customerRequest{EmailAddress: "not-an-address"})

	detail := reply.problemDetail(t, http.StatusBadRequest)
	assert.Contains(t, detail, "lastName")
	assert.Contains(t, detail, "emailAddress")
	assert.NotContains(t, detail, "not-an-address")
}

func TestACustomerBodyWithAFieldThatTheDataHolderDoesNotHoldOrACustomerNumberIsRefused(t *testing.T) {
	l := newOperatorLab(t)

	for field, value := range map[string]string{"dateOfBirth": "2000-01-01", "customerNumber": "12345678"} {
		t.Run(field, func(t *testing.T) {
			reply := l.operate(t, http.MethodPost, "/v1/customers", map[string]string{
				"lastName": "Example", "emailAddress": "alex@example.com", field: value,
			})

			assert.Contains(t, reply.problemDetail(t, http.StatusBadRequest), field)
		})
	}
}

func TestAnAccountThatIsNotValidIsRefusedWithEachRuleAndIsNotCreated(t *testing.T) {
	l := newOperatorLab(t)
	customer := l.newCustomer(t)
	account := everydayAccount("12345")
	account.BSB = "12-345"
	account.ProductCategory = "SAVINGS"
	account.DisplayName = ""

	reply := l.operate(t, http.MethodPost, accountsPath(customer.CustomerNumber), account)

	detail := reply.problemDetail(t, http.StatusBadRequest)
	for _, field := range []string{"displayName", "productCategory", "bsb", "accountNumber"} {
		assert.Contains(t, detail, field)
	}

	assert.Contains(t, string(l.customer(t, customer.CustomerNumber).body), `"accounts":[]`)
}

// Verifies: DH-0740
func TestASecondAccountWithTheSameBSBAndAccountNumberGivesConflict(t *testing.T) {
	l := newOperatorLab(t)
	first, second := l.newCustomer(t), l.newCustomer(t)
	account := everydayAccount("10000003")
	l.createAccount(t, first.CustomerNumber, account)

	l.operate(t, http.MethodPost, accountsPath(first.CustomerNumber), account).problemDetail(t, http.StatusConflict)
	l.operate(t, http.MethodPost, accountsPath(second.CustomerNumber), account).problemDetail(t, http.StatusConflict)
}

// Verifies: DH-0736
func TestACustomerWithTheEmailAddressOfADifferentCustomerInADifferentCaseGivesConflict(t *testing.T) {
	l := newOperatorLab(t)
	first := l.newCustomer(t)

	reply := l.operate(t, http.MethodPost, "/v1/customers", customerRequest{
		LastName: "Sample", EmailAddress: strings.ToUpper(first.EmailAddress),
	})

	assert.Equal(t, "A customer has this emailAddress.", reply.problemDetail(t, http.StatusConflict))
	assert.NotContains(t, string(reply.body), first.EmailAddress)
	assert.NotContains(t, string(reply.body), strings.ToUpper(first.EmailAddress))
}

// Verifies: DH-0732
func TestACustomerNumberThatDoesNotExistGivesNotFound(t *testing.T) {
	l := newOperatorLab(t)

	l.customer(t, unknownCustomerNumber).problemDetail(t, http.StatusNotFound)
	l.operate(t, http.MethodPost, accountsPath(unknownCustomerNumber), everydayAccount("10000004")).
		problemDetail(t, http.StatusNotFound)
	l.operate(t, http.MethodGet, arrangementsPath(unknownCustomerNumber), nil).problemDetail(t, http.StatusNotFound)
}

// Verifies: DH-0741
func TestTheOperatorCreatesAnAccountWithACreditLimit(t *testing.T) {
	l := newOperatorLab(t)
	card := everydayAccount("10000011")
	card.ProductCategory = "CRED_AND_CHRG_CARDS"
	card.CreditLimit = "2500.00"

	created := l.createAccount(t, l.newCustomer(t).CustomerNumber, card)

	assert.Equal(t, "2500.00", created.CreditLimit)
	assert.Equal(t, "OPEN", created.OpenStatus)
}

func TestACreditLimitThatIsNotAnAmountWithTwoDecimalDigitsIsRefusedAndNoAccountIsCreated(t *testing.T) {
	l := newOperatorLab(t)
	customer := l.newCustomer(t)

	for _, limit := range []string{"2500", "2500.5", "-1.00", "2,500.00"} {
		account := everydayAccount("10000020")
		account.CreditLimit = limit

		reply := l.operate(t, http.MethodPost, accountsPath(customer.CustomerNumber), account)

		assert.Contains(t, reply.problemDetail(t, http.StatusBadRequest), "creditLimit", limit)
	}

	assert.Contains(t, string(l.customer(t, customer.CustomerNumber).body), `"accounts":[]`)
}

// Verifies: DH-0723, DH-0747
func TestTheOperatorClosesAnOpenAccountOneTime(t *testing.T) {
	l := newOperatorLab(t)
	customer := l.newCustomer(t)
	created := l.createAccount(t, customer.CustomerNumber, everydayAccount("10000012"))

	closed := l.operate(t, http.MethodPost, closePath(customer.CustomerNumber, created.AccountID), nil)
	require.Equal(t, http.StatusOK, closed.status, string(closed.body))

	var account operatorAccount
	closed.decode(t, &account)

	created.OpenStatus = "CLOSED"
	assert.Equal(t, created, account)

	var read operatorCustomer
	l.customer(t, customer.CustomerNumber).decode(t, &read)
	assert.Equal(t, []operatorAccount{created}, read.Accounts)

	l.operate(t, http.MethodPost, closePath(customer.CustomerNumber, created.AccountID), nil).
		problemDetail(t, http.StatusConflict)
}

// Verifies: DH-0746
func TestTheOperatorCannotCloseAnAccountThatTheCustomerDoesNotHave(t *testing.T) {
	l := newOperatorLab(t)
	owner, other := l.newCustomer(t), l.newCustomer(t)
	created := l.createAccount(t, owner.CustomerNumber, everydayAccount("10000013"))

	for _, path := range []string{
		closePath(other.CustomerNumber, created.AccountID),
		closePath(owner.CustomerNumber, uuid.NewString()),
		closePath(unknownCustomerNumber, created.AccountID),
	} {
		l.operate(t, http.MethodPost, path, nil).problemDetail(t, http.StatusNotFound)
	}

	assert.Contains(t, string(l.customer(t, owner.CustomerNumber).body), `"openStatus":"OPEN"`)
}

func TestTheOperatorCreatesAndReadsTheTransactionsOfAnAccount(t *testing.T) {
	l := newOperatorLab(t)
	customer := l.newCustomer(t)
	account := l.createAccount(t, customer.CustomerNumber, everydayAccount("40000010"))
	path := transactionsPath(customer.CustomerNumber, account.AccountID)
	at := time.Date(2026, 10, 1, 9, 30, 0, 0, time.UTC)
	posted := func(description, amount string, at time.Time) operatorTransaction {
		return operatorTransaction{
			Type: "PAYMENT", Status: "POSTED", Description: description, Amount: amount,
			PostingDateTime: at.Format(time.RFC3339),
		}
	}
	add := func(transaction operatorTransaction) string {
		reply := l.operate(t, http.MethodPost, path, transaction)
		require.Equal(t, http.StatusCreated, reply.status, string(reply.body))

		var created operatorTransaction
		reply.decode(t, &created)

		return created.TransactionID
	}

	first := add(posted("Second by time", "-1.00", at))
	second := add(posted("First by time", "-2.00", at.Add(-time.Hour)))

	reply := l.operate(t, http.MethodGet, path, nil)
	require.Equal(t, http.StatusOK, reply.status, string(reply.body))

	var read []operatorTransaction
	reply.decode(t, &read)
	require.Len(t, read, 2)
	assert.Equal(t, []string{first, second}, []string{read[0].TransactionID, read[1].TransactionID},
		"in the sequence of creation")
	assert.Equal(t, "2026-10-01T09:30:00Z", read[0].PostingDateTime)

	t.Run("a transaction that breaks a rule is refused with the field", func(t *testing.T) {
		noPostingTime := posted("Coffee", "-1.00", at)
		noPostingTime.PostingDateTime = ""

		assert.Contains(t, l.operate(t, http.MethodPost, path, noPostingTime).problemDetail(t, http.StatusBadRequest),
			"postingDateTime")
		assert.Contains(t,
			l.operate(t, http.MethodPost, path, posted("Coffee", "-1", at)).problemDetail(t, http.StatusBadRequest),
			"amount")
	})

	t.Run("a transaction for an account that the customer does not have is refused", func(t *testing.T) {
		other := l.newCustomer(t)

		l.operate(t, http.MethodPost, transactionsPath(other.CustomerNumber, account.AccountID),
			posted("Coffee", "-1.00", at)).problemDetail(t, http.StatusNotFound)
	})

	t.Run("a transaction for a closed account gives conflict", func(t *testing.T) {
		closed := l.operate(t, http.MethodPost, closePath(customer.CustomerNumber, account.AccountID), nil)
		require.Equal(t, http.StatusOK, closed.status, string(closed.body))

		l.operate(t, http.MethodPost, path, posted("Coffee", "-1.00", at)).problemDetail(t, http.StatusConflict)
	})
}

// newArrangement makes the arrangement that a confirmed consent makes.
func (l operatorLab) newArrangement(
	t *testing.T,
	customerNumber string,
	sharingDurationSeconds int64,
) domain.Arrangement {
	t.Helper()

	inserter, err := jobs.NewInsertOnlyClient(l.pool)
	require.NoError(t, err)

	arrangement := domain.Arrangement{
		ArrangementID:          uuid.NewString(),
		CustomerNumber:         customerNumber,
		ClientID:               uuid.NewString(),
		SoftwareID:             "software-1",
		ClientName:             "Recipient Web",
		Scopes:                 []string{"openid", "bank:accounts.basic:read"},
		AccountIDs:             []string{uuid.NewString()},
		ConfirmedAt:            time.Now().UTC().Truncate(time.Second),
		SharingDurationSeconds: sharingDurationSeconds,
	}
	require.NoError(t, app.ArrangementType.Execute(
		t.Context(), newEventStore(l.pool, inserter), arrangement.ArrangementID,
		domain.CreateArrangement{Arrangement: arrangement},
	))

	return arrangement
}

func (l operatorLab) arrangementsOf(t *testing.T, customerNumber string) []operatorArrangement {
	t.Helper()

	reply := l.operate(t, http.MethodGet, arrangementsPath(customerNumber), nil)
	require.Equal(t, http.StatusOK, reply.status, string(reply.body))

	var arrangements []operatorArrangement
	reply.decode(t, &arrangements)

	return arrangements
}

// Verifies: DH-0727, DH-0764
func TestTheOperatorReadsAnArrangementAndWithdrawsItOneTime(t *testing.T) {
	const ninetyDays = 90 * 24 * 60 * 60

	l := newOperatorLab(t)
	number := l.newCustomer(t).CustomerNumber
	arrangement := l.newArrangement(t, number, ninetyDays)

	arrangements := l.arrangementsOf(t, number)
	require.Len(t, arrangements, 1)
	assert.Equal(t, operatorArrangement{
		ArrangementID: arrangement.ArrangementID,
		SoftwareID:    arrangement.SoftwareID,
		ClientName:    arrangement.ClientName,
		Scopes:        arrangement.Scopes,
		AccountIDs:    arrangement.AccountIDs,
		ConfirmedAt:   arrangement.ConfirmedAt,
		SharingEndsAt: arrangement.ConfirmedAt.Add(ninetyDays * time.Second),
	}, arrangements[0])

	requested := time.Now().Add(-time.Second)
	reply := l.operate(t, http.MethodPost, withdrawPath(number, arrangement.ArrangementID), nil)
	require.Equal(t, http.StatusOK, reply.status, string(reply.body))

	var withdrawn operatorArrangement
	reply.decode(t, &withdrawn)
	assert.Equal(t, arrangement.ArrangementID, withdrawn.ArrangementID)
	assert.Equal(t, "WITHDRAWN_BY_CONSUMER", withdrawn.EndReason)
	require.NotNil(t, withdrawn.EndedAt)
	assert.WithinRange(t, *withdrawn.EndedAt, requested, time.Now())
	assert.Equal(t, []operatorArrangement{withdrawn}, l.arrangementsOf(t, number))

	assert.Equal(t, "The arrangement ended.",
		l.operate(t, http.MethodPost, withdrawPath(number, arrangement.ArrangementID), nil).
			problemDetail(t, http.StatusConflict))
}

// Verifies: DH-0763
func TestAWithdrawalOfAnArrangementThatTheCustomerDoesNotHaveGivesNotFoundAndEndsNothing(t *testing.T) {
	l := newOperatorLab(t)
	number, other := l.newCustomer(t).CustomerNumber, l.newCustomer(t).CustomerNumber
	arrangementID := l.newArrangement(t, number, 3600).ArrangementID

	tests := map[string]struct{ customerNumber, arrangementID, detail string }{
		"arrangement of a different customer": {
			other, arrangementID, "The customer has no arrangement with this arrangementId.",
		},
		"arrangement that does not exist": {
			number, uuid.NewString(), "The customer has no arrangement with this arrangementId.",
		},
		"customer that does not exist": {
			"99999999", arrangementID, "The data holder has no customer with this customer number.",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			reply := l.operate(t, http.MethodPost, withdrawPath(tc.customerNumber, tc.arrangementID), nil)

			assert.Equal(t, tc.detail, reply.problemDetail(t, http.StatusNotFound))
		})
	}

	assert.Nil(t, l.arrangementsOf(t, number)[0].EndedAt, "no refused request ends the arrangement")
}

func TestAnArrangementWithTheSharingDuration0CannotBeWithdrawn(t *testing.T) {
	l := newOperatorLab(t)
	number := l.newCustomer(t).CustomerNumber
	arrangementID := l.newArrangement(t, number, 0).ArrangementID

	l.operate(t, http.MethodPost, withdrawPath(number, arrangementID), nil).problemDetail(t, http.StatusConflict)
}

func TestACustomerWithNoArrangementsHasAnEmptyArrangementList(t *testing.T) {
	l := newOperatorLab(t)

	reply := l.operate(t, http.MethodGet, arrangementsPath(l.newCustomer(t).CustomerNumber), nil)

	require.Equal(t, http.StatusOK, reply.status, string(reply.body))
	assert.JSONEq(t, "[]", string(reply.body))
}

func TestEachOperatorRequestNeedsAnAllowedHostAndTheOperatorTokenAndARefusedRequestChangesNothing(t *testing.T) {
	l := newOperatorLab(t)
	customer := l.newCustomer(t)

	requests := map[string]struct {
		method, path string
		body         any
	}{
		"create a customer": {method: http.MethodPost, path: "/v1/customers", body: customerRequest{
			LastName: "Example", EmailAddress: "alex@example.com",
		}},
		"read a customer": {method: http.MethodGet, path: "/v1/customers/" + customer.CustomerNumber},
		"create an account": {
			method: http.MethodPost, path: accountsPath(customer.CustomerNumber), body: everydayAccount("10000005"),
		},
		"close an account":      {method: http.MethodPost, path: closePath(customer.CustomerNumber, uuid.NewString())},
		"read the arrangements": {method: http.MethodGet, path: arrangementsPath(customer.CustomerNumber)},
		"withdraw a consent": {
			method: http.MethodPost, path: withdrawPath(customer.CustomerNumber, uuid.NewString()),
		},
		"unknown path": {method: http.MethodGet, path: "/v1/unknown"},
	}
	for name, req := range requests {
		t.Run(name+" with no token", func(t *testing.T) {
			reply := l.request(t, req.method, req.path, http.Header{}, req.body)

			reply.problemDetail(t, http.StatusUnauthorized)
			assert.Equal(t, "Bearer", reply.header.Get("WWW-Authenticate"))
		})
		t.Run(name+" for a host that is not allowed", func(t *testing.T) {
			header := http.Header{"Authorization": {"Bearer test-operator-token"}}

			reply := l.requestHost(t, "attacker.example:8080", req.method, req.path, header, req.body)

			reply.problemDetail(t, http.StatusForbidden)
		})
		t.Run(name+" with a different token", func(t *testing.T) {
			header := http.Header{"Authorization": {"Bearer not-the-operator-token"}}

			reply := l.request(t, req.method, req.path, header, req.body)

			reply.problemDetail(t, http.StatusUnauthorized)
			assert.Equal(t, `Bearer error="invalid_token"`, reply.header.Get("WWW-Authenticate"))
		})
	}

	assert.Contains(t, string(l.customer(t, customer.CustomerNumber).body), `"accounts":[]`)
}

func TestAnUnknownOperatorPathGivesNotFound(t *testing.T) {
	l := newOperatorLab(t)
	customer := l.newCustomer(t)

	for _, req := range []struct{ method, path string }{
		{method: http.MethodGet, path: "/v1/unknown"},
		{method: http.MethodGet, path: "/v1/customers"},
		{method: http.MethodDelete, path: "/v1/customers/" + customer.CustomerNumber},
		{method: http.MethodPut, path: "/v1/customers/" + customer.CustomerNumber},
	} {
		t.Run(req.method+" "+req.path, func(t *testing.T) {
			l.operate(t, req.method, req.path, nil).problemDetail(t, http.StatusNotFound)
		})
	}
}

func TestTheOperatorInterfaceDoesNotPermitCrossOriginRequests(t *testing.T) {
	l := newOperatorLab(t)
	header := http.Header{"Authorization": {"Bearer test-operator-token"}, "Origin": {"https://app.example"}}

	reply := l.request(t, http.MethodGet, "/v1/unknown", header, nil)

	assert.Empty(t, reply.header.Get("Access-Control-Allow-Origin"))
}

func TestHealthProbesReturnOK(t *testing.T) {
	l := newOperatorLab(t)
	require.NoError(
		t,
		statusstore.New(l.pool, nil).Save(t.Context(), map[string]domain.SoftwareProductStatus{}, `"v1"`),
	)

	for _, path := range []string{"/livez", "/readyz"} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			l.handlers[internalPort].ServeHTTP(
				rec,
				httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil),
			)

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}
