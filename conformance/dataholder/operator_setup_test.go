//go:build conformance

package dataholder

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

// The suite calls the operator interface only to set up a state for a test of the CDR, and the data holder tests the
// interface itself. All customer and account data in this file is invented.

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

func operate(t *testing.T, method, path string, body any) operatorReply {
	t.Helper()

	var encoded []byte

	if body != nil {
		var err error

		encoded, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req, err := http.NewRequestWithContext(
		t.Context(),
		method,
		labtest.Env(t, "DATAHOLDER_INTERNAL_URL")+path,
		bytes.NewReader(encoded),
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+labtest.Env(t, "DATAHOLDER_OPERATOR_TOKEN"))
	resp, err := labtest.Client(t).Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	replyBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return operatorReply{status: resp.StatusCode, header: resp.Header, body: replyBody}
}

func newCustomerWithEmail(t *testing.T, emailAddress string) operatorCustomer {
	t.Helper()

	reply := operate(t, http.MethodPost, "/v1/customers", customerRequest{
		FirstName:    "Alex",
		LastName:     "Example",
		EmailAddress: emailAddress,
	})
	require.Equal(t, http.StatusCreated, reply.status, string(reply.body))

	var created operatorCustomer
	reply.decode(t, &created)

	return created
}

func accountsPath(customerNumber string) string {
	return "/v1/customers/" + customerNumber + "/accounts"
}

func closePath(customerNumber, accountID string) string {
	return accountsPath(customerNumber) + "/" + accountID + "/close"
}

func createAccount(t *testing.T, customerNumber string, account operatorAccount) operatorAccount {
	t.Helper()

	reply := operate(t, http.MethodPost, accountsPath(customerNumber), account)
	require.Equal(t, http.StatusCreated, reply.status, string(reply.body))

	var created operatorAccount
	reply.decode(t, &created)

	return created
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
	NoticeStatus  string     `json:"noticeStatus"`
}

func arrangementsPath(customerNumber string) string {
	return "/v1/customers/" + customerNumber + "/arrangements"
}

func withdrawPath(customerNumber, arrangementID string) string {
	return arrangementsPath(customerNumber) + "/" + arrangementID + "/withdraw"
}

func arrangementsOf(t *testing.T, customerNumber string) []operatorArrangement {
	t.Helper()

	reply := operate(t, http.MethodGet, arrangementsPath(customerNumber), nil)
	require.Equal(t, http.StatusOK, reply.status, string(reply.body))

	var arrangements []operatorArrangement
	reply.decode(t, &arrangements)

	return arrangements
}

func withdrawConsent(t *testing.T, customerNumber, arrangementID string) operatorReply {
	t.Helper()

	return operate(t, http.MethodPost, withdrawPath(customerNumber, arrangementID), nil)
}
